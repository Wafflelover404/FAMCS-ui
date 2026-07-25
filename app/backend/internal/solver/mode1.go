package solver

import (
	"context"
	"sort"
)

type mode1Search struct {
	ctx      context.Context
	b        *Board
	st       []Color
	rem      []int
	revCover []Move
	tracer   *Tracer
	nodeCap  int64
	nodes    int64
	capHit   bool
	canceled bool
}

func moveFor(r, c, d int, cu Color) Move {
	switch d {
	case 0:
		return Move{R: r, C: c, Orient: Horizontal, Color: cu}
	case 1:
		return Move{R: r, C: c - 1, Orient: Horizontal, Color: cu}
	case 2:
		return Move{R: r, C: c, Orient: Vertical, Color: cu}
	default:
		return Move{R: r - 1, C: c, Orient: Vertical, Color: cu}
	}
}

func (s *mode1Search) optionCount(u int) int {
	cu := s.st[u]
	r, c := u/s.b.N, u%s.b.N
	cnt := 0
	for d := 0; d < 4; d++ {
		rr, cc := r+dRow[d], c+dCol[d]
		if !s.b.InBounds(rr, cc) {
			continue
		}
		v := s.b.Index(rr, cc)
		if s.b.IsBackground(v) {
			continue
		}
		if s.st[v] == Background || s.st[v] == cu {
			cnt++
		}
	}
	return cnt
}

func (s *mode1Search) pickMRV() int {
	const inf = 1 << 30
	best, bestCnt := -2, inf
	for i := range s.st {
		if s.b.IsBackground(i) || s.st[i] == Background {
			continue
		}
		if best == -2 {
			best = -3
		}
		cnt := s.optionCount(i)
		if cnt > 0 && cnt < bestCnt {
			bestCnt, best = cnt, i
		}
	}
	if best == -3 {
		return -2
	}
	if best == -2 {
		return -1
	}
	return best
}

func (s *mode1Search) dfs1(depth int) bool {
	s.nodes++
	if s.tracer != nil {
		s.tracer.Node(depth)
	}
	if s.nodes > s.nodeCap {
		s.capHit = true
		return false
	}
	if s.nodes&2047 == 0 && s.ctx.Err() != nil {
		s.canceled = true
		return false
	}

	u := s.pickMRV()
	if u == -1 {
		return true
	}
	if u == -2 {
		if s.tracer != nil {
			s.tracer.Prune(depth, PruneDeadCell, nil)
		}
		return false
	}

	r, c := u/s.b.N, u%s.b.N
	cu := s.st[u]
	if s.rem[cu] == 0 {
		if s.tracer != nil {
			s.tracer.Prune(depth, PruneNoStock, []int{u})
		}
		return false
	}

	type candidate struct{ d, v int }
	var cands []candidate
	for d := 0; d < 4; d++ {
		rr, cc := r+dRow[d], c+dCol[d]
		if !s.b.InBounds(rr, cc) {
			continue
		}
		v := s.b.Index(rr, cc)
		if s.b.IsBackground(v) {
			continue
		}
		if s.st[v] == Background || s.st[v] == cu {
			cands = append(cands, candidate{d, v})
		}
	}

	sort.SliceStable(cands, func(i, j int) bool {
		return s.st[cands[i].v] == cu && s.st[cands[j].v] != cu
	})

	for _, cd := range cands {
		v := cd.v
		ov := s.st[v]
		s.st[u], s.st[v] = Background, Background
		s.rem[cu]--
		mv := moveFor(r, c, cd.d, cu)
		s.revCover = append(s.revCover, mv)
		if s.tracer != nil {
			s.tracer.Peel(depth, len(s.revCover), mv, []int{u, v})
		}
		if s.dfs1(depth + 1) {
			return true
		}
		s.revCover = s.revCover[:len(s.revCover)-1]
		if s.tracer != nil {
			s.tracer.Unpeel(depth, len(s.revCover)+1, mv)
		}
		s.st[u], s.st[v] = cu, ov
		s.rem[cu]++
	}
	return false
}

func (s *mode1Search) assemble() ([]Move, bool) {
	var out []Move
	surplus := 0
	for c := 1; c < len(s.rem); c++ {
		surplus += s.rem[c]
	}
	if surplus > 0 {
		pairs := s.b.Pairs()
		if len(pairs) == 0 {
			return nil, false
		}
		slot := pairs[0].Move
		for c := 1; c < len(s.rem); c++ {
			for t := 0; t < s.rem[c]; t++ {
				mm := slot
				mm.Color = Color(c)
				out = append(out, mm)
			}
		}
	}
	for i := len(s.revCover) - 1; i >= 0; i-- {
		out = append(out, s.revCover[i])
	}
	for i := range out {
		out[i].Step = i + 1
	}
	return out, true
}

func Solve1(ctx context.Context, b *Board, stock []int, opts Options) Result {
	tracer := opts.Tracer

	if r := Prescreen(b); r != nil {
		return rejected(r, nil, tracer)
	}
	if !b.HasAnyColored() {
		total := 0
		for _, c := range stock {
			total += c
		}
		if total == 0 {
			return feasible(b, nil, nil, tracer)
		}
		return rejected(reasonEmptyBoardNonEmptyStock(total), nil, tracer)
	}

	cols, reason := StockNeed(b, stock)
	if reason != nil {
		return rejected(reason, cols, tracer)
	}

	rem := make([]int, MaxColor+1)
	for c := 1; c <= MaxColor && c < len(stock); c++ {
		rem[c] = stock[c]
	}

	s := &mode1Search{
		ctx:     ctx,
		b:       b,
		st:      append([]Color(nil), b.Cells...),
		rem:     rem,
		tracer:  tracer,
		nodeCap: opts.nodeCap(),
	}
	ok := s.dfs1(0)
	if !ok {
		if s.canceled {
			return rejected(reasonCanceled(), cols, tracer)
		}
		if s.capHit {
			return rejected(reasonNodeCapExceeded(s.nodeCap), cols, tracer)
		}
		if tracer != nil {
			tracer.Exhausted()
		}
		return rejected(reasonSearchExhausted(), cols, tracer)
	}

	moves, assembled := s.assemble()
	if !assembled {
		return rejected(reasonSearchExhausted(), cols, tracer)
	}
	return feasible(b, moves, cols, tracer)
}
