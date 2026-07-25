package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"famcs-ui/backend/internal/cache"
	"famcs-ui/backend/internal/queue"
	"famcs-ui/backend/internal/solver"
	"famcs-ui/backend/internal/store"
)

const (
	defaultSolveNodeCap = 3_000_000
	maxSolveNodeCap     = 15_000_000
	defaultCountGuard   = 300_000
	maxCountGuard       = 3_000_000
	traceDetailCap      = 20_000
	requestTimeout      = 20 * time.Second
	countTTL            = time.Hour
)

var allowedOrigins = map[string]bool{
	"http://localhost:5173": true,
	"http://127.0.0.1:5173": true,
	"http://localhost:4173": true,
	"http://127.0.0.1:4173": true,
}

type Deps struct {
	Store *store.Store
	Cache *cache.Cache
	Queue queue.Queue
}

type Server struct {
	Deps
}

func NewRouter(d Deps) http.Handler {
	s := &Server{Deps: d}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /api/v1/presets", handlePresets)
	mux.HandleFunc("POST /api/v1/analyze", s.handleAnalyze)
	mux.HandleFunc("POST /api/v1/solve", s.handleSolve)
	mux.HandleFunc("POST /api/v1/count", s.handleCount)
	mux.HandleFunc("POST /api/v1/simulate", handleSimulate)
	mux.HandleFunc("POST /api/v1/random", handleRandom)
	mux.HandleFunc("POST /api/v1/verify", s.handleVerify)
	mux.HandleFunc("GET /api/v1/jobs/{id}", s.handleJobStatus)
	mux.HandleFunc("GET /api/v1/ops/queue", s.handleQueueDepth)
	mux.HandleFunc("GET /api/v1/benchmarks", s.handleBenchmarks)
	return withCORS(withLogging(mux))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "elapsed", time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handlePresets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, builtinPresets())
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	b, err := req.toBoard()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	key := ""
	if s.Cache != nil {
		raw, _ := json.Marshal(req)
		key = cache.Key("analyze", string(raw))
		if hit, ok := cache.Get[AnalyzeResponse](ctx, s.Cache, key); ok {
			writeJSON(w, http.StatusOK, hit)
			return
		}
	}

	resp := AnalyzeResponse{Parity: make([]int, len(b.Cells))}
	for i := range b.Cells {
		resp.Parity[i] = b.Parity(i)
	}

	reachable, residue := solver.Reachable(b)
	resp.Reachable = reachable
	resp.Residue = residue

	if len(req.Stock) > 0 {
		stock := stockFrom(req.K, req.Stock)
		cols, reason := solver.StockNeed(b, stock)
		resp.Colors = cols
		resp.Reason = reason
	} else {
		for _, c := range b.Colors() {
			resp.Colors = append(resp.Colors, solver.MatchColor(b, c))
		}
		if !reachable {
			resp.Reason = solver.Prescreen(b)
		}
	}

	if s.Cache != nil && key != "" {
		_ = cache.Set(ctx, s.Cache, key, resp, cache.AnalyzeTTL)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSolve(w http.ResponseWriter, r *http.Request) {
	var req SolveRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	b, err := req.toBoard()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	nodeCap := int64(defaultSolveNodeCap)
	if req.NodeCap > 0 {
		nodeCap = req.NodeCap
	}
	if nodeCap > maxSolveNodeCap {
		nodeCap = maxSolveNodeCap
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	cacheable := s.Cache != nil && !req.Trace
	key := ""
	if cacheable {
		raw, _ := json.Marshal(req)
		key = cache.Key("solve", string(raw))
		if hit, ok := cache.Get[SolveResponse](ctx, s.Cache, key); ok {
			writeJSON(w, http.StatusOK, hit)
			return
		}
	}

	detailCap := 1
	if req.Trace {
		detailCap = traceDetailCap
	}
	tracer := solver.NewTracer(detailCap, nil)
	opts := solver.Options{NodeCap: nodeCap, Tracer: tracer}

	var res solver.Result
	switch req.Mode {
	case "stock":
		stock := stockFrom(req.K, req.Stock)
		res = solver.Solve1(ctx, b, stock, opts)
	case "sequence":
		seq := sequenceFrom(req.Sequence)
		res = solver.Solve2(ctx, b, seq, opts)
	default:
		writeError(w, http.StatusBadRequest, "mode must be \"stock\" or \"sequence\"")
		return
	}

	resp := SolveResponse{
		Verdict:  res.Verdict,
		Moves:    res.Moves,
		Reason:   res.Reason,
		Colors:   res.Colors,
		Stats:    res.Stats,
		Verified: res.Verified,
	}
	if req.Trace {
		resp.Trace = tracer.Events()
		resp.Truncated = tracer.Truncated()
	}

	if cacheable && key != "" {
		_ = cache.Set(ctx, s.Cache, key, resp, cache.SolveTTL)
	}
	if s.Store != nil {
		go func() {
			bgCtx := context.Background()
			_ = s.Store.RecordBenchmark(bgCtx, "solve:"+req.Mode, req.N, req.K, res.Stats.Nodes, float64(res.Stats.ElapsedNS)/1e6, string(res.Verdict))
		}()
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCount(w http.ResponseWriter, r *http.Request) {
	var req CountRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.N < solver.MinN || req.N > solver.MaxN {
		writeError(w, http.StatusBadRequest, "n out of range")
		return
	}

	guard := int64(defaultCountGuard)
	if req.Guard > 0 {
		guard = req.Guard
	}
	if guard > maxCountGuard {
		guard = maxCountGuard
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	key := ""
	if s.Cache != nil {
		raw, _ := json.Marshal(req)
		key = cache.Key("count", string(raw))
		if hit, ok := cache.Get[solver.CountResult](ctx, s.Cache, key); ok {
			writeJSON(w, http.StatusOK, hit)
			return
		}
	}

	var res solver.CountResult
	switch req.Mode {
	case "stock":
		stock := stockFrom(req.K, req.Stock)
		res = solver.CountStock(ctx, req.N, stock, guard)
	case "sequence":
		seq := make([]solver.Color, len(req.Sequence))
		for i, v := range req.Sequence {
			seq[i] = solver.Color(v)
		}
		res = solver.CountSequence(ctx, req.N, seq, guard)
	default:
		writeError(w, http.StatusBadRequest, "mode must be \"stock\" or \"sequence\"")
		return
	}

	if s.Cache != nil && key != "" {
		_ = cache.Set(ctx, s.Cache, key, res, countTTL)
	}
	writeJSON(w, http.StatusOK, res)
}

func handleSimulate(w http.ResponseWriter, r *http.Request) {
	var req SimulateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.N < solver.MinN || req.N > solver.MaxN {
		writeError(w, http.StatusBadRequest, "n out of range")
		return
	}
	if !solver.ValidMoves(req.N, req.Moves) {
		writeError(w, http.StatusBadRequest, "invalid move list")
		return
	}
	writeJSON(w, http.StatusOK, SimulateResponse{
		Frames:   solver.Frames(req.N, req.Moves),
		Coverage: solver.Coverage(req.N, req.Moves),
		Visible:  solver.Visible(req.N, req.Moves),
	})
}

func handleRandom(w http.ResponseWriter, r *http.Request) {
	var req RandomRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.N < solver.MinN || req.N > solver.MaxN {
		writeError(w, http.StatusBadRequest, "n out of range")
		return
	}
	if req.K < 1 || req.K > solver.MaxColor {
		writeError(w, http.StatusBadRequest, "k out of range")
		return
	}
	if req.M < 1 || req.M > solver.MaxMoves {
		writeError(w, http.StatusBadRequest, "m out of range")
		return
	}

	seed := req.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))
	seq := solver.RandomSequence(rng, req.M, req.K)
	b, moves := solver.GenerateRandom(rng, req.N, seq)
	stock := solver.StockOf(seq, req.K)

	cells := make([]int, len(b.Cells))
	for i, c := range b.Cells {
		cells[i] = int(c)
	}
	seqOut := make([]int, len(seq))
	for i, c := range seq {
		seqOut[i] = int(c)
	}

	writeJSON(w, http.StatusOK, RandomResponse{
		N:        req.N,
		K:        req.K,
		Cells:    cells,
		Sequence: seqOut,
		Stock:    stock[1:],
		Moves:    moves,
	})
}

type VerifyRequest struct {
	Suite  string `json:"suite"`
	Trials int    `json:"trials,omitempty"`
	Seed   int64  `json:"seed,omitempty"`
	MaxN   int    `json:"maxN,omitempty"`
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if s.Queue == nil || s.Store == nil {
		writeError(w, http.StatusServiceUnavailable, "queue/store not configured")
		return
	}
	var req VerifyRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Suite != "stress" && req.Suite != "matching" {
		writeError(w, http.StatusBadRequest, "suite must be \"stress\" or \"matching\"")
		return
	}

	ctx := r.Context()
	id, err := s.Store.CreateJob(ctx, "verify", req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	params, _ := json.Marshal(req)
	if err := s.Queue.Enqueue(ctx, queue.Job{ID: id, Kind: "verify", Params: params}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": id})
}

func (s *Server) handleJobStatus(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		writeError(w, http.StatusServiceUnavailable, "store not configured")
		return
	}
	id := r.PathValue("id")
	job, err := s.Store.GetJob(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleQueueDepth(w http.ResponseWriter, r *http.Request) {
	if s.Queue == nil {
		writeJSON(w, http.StatusOK, map[string]any{"configured": false})
		return
	}
	depth, err := s.Queue.Depth(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"configured": true, "depth": depth})
}

func (s *Server) handleBenchmarks(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	mode := r.URL.Query().Get("mode")
	rows, err := s.Store.ListBenchmarks(r.Context(), mode, 500)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
