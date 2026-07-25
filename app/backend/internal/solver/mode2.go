package solver

import (
	"context"
	"sort"
)

type mode2Search struct {
	ctx      context.Context
	b        *Board
	st       []Color
	k        int
	seq      []Color
	pos      []Move
	pairs    []Pair
	memo     map[string]bool
	tracer   *Tracer
	nodeCap  int64
	nodes    int64
	capHit   bool
	canceled bool

	needBuf []int
	haveBuf []int
}

func (s *mode2Search) prune2(j int) bool {
	for i := range s.needBuf {
		s.needBuf[i] = 0
	}
	for i, c := range s.st {
		if s.b.IsBackground(i) || c == Background {
			continue
		}
		s.needBuf[c]++
	}
	for i := range s.haveBuf {
		s.haveBuf[i] = 0
	}
	for t := 1; t <= j; t++ {
		s.haveBuf[s.seq[t]]++
	}
	for c := 1; c <= s.k; c++ {
		if s.needBuf[c] > 2*s.haveBuf[c] {
			return true
		}
	}

	nbrs := make([]int, 0, 4)
	for i := range s.st {
		if s.b.IsBackground(i) {
			continue
		}
		nbrs = s.b.Neighbors4(i, nbrs)
		ok := false
		for _, v := range nbrs {
			if !s.b.IsBackground(v) {
				ok = true
				break
			}
		}
		if !ok {
			return true
		}
	}
	return false
}

func (s *mode2Search) encode(j int) string {
	buf := make([]byte, len(s.st)+2)
	for i, c := range s.st {
		if s.b.IsBackground(i) {
			buf[i] = 255
		} else {
			buf[i] = byte(c)
		}
	}
	buf[len(s.st)] = byte(j & 0xff)
	buf[len(s.st)+1] = byte((j >> 8) & 0xff)
	return string(buf)
}

func (s *mode2Search) dfs2(j, depth int) bool {
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

	if j == 0 {
		for i, c := range s.st {
			if !s.b.IsBackground(i) && c != Background {
				return false
			}
		}
		return true
	}
	if s.prune2(j) {
		if s.tracer != nil {
			s.tracer.Prune(depth, PruneBalance, nil)
		}
		return false
	}
	key := s.encode(j)
	if s.memo[key] {
		if s.tracer != nil {
			s.tracer.Prune(depth, PruneMemo, nil)
		}
		return false
	}

	c := s.seq[j]
	var order []Pair
	for _, p := range s.pairs {
		sa, sb := s.st[p.A], s.st[p.B]
		if (sa == Background || sa == c) && (sb == Background || sb == c) {
			order = append(order, p)
		}
	}
	progress := func(p Pair) bool {
		return s.st[p.A] != Background || s.st[p.B] != Background
	}
	sort.SliceStable(order, func(i, j2 int) bool {
		return progress(order[i]) && !progress(order[j2])
	})

	for _, p := range order {
		oa, ob := s.st[p.A], s.st[p.B]
		s.st[p.A], s.st[p.B] = Background, Background
		mv := p.Move
		mv.Color = c
		s.pos[j] = mv
		if s.tracer != nil {
			s.tracer.Peel(depth, j, mv, []int{p.A, p.B})
		}
		if s.dfs2(j-1, depth+1) {
			return true
		}
		if s.tracer != nil {
			s.tracer.Unpeel(depth, j, mv)
		}
		s.st[p.A], s.st[p.B] = oa, ob
	}

	if len(s.memo) < 4_000_000 {
		s.memo[key] = true
		if s.tracer != nil {
			s.tracer.MemoStore(len(s.memo))
		}
	}
	return false
}

func Solve2(ctx context.Context, b *Board, seq []Color, opts Options) Result {
	tracer := opts.Tracer
	m := len(seq) - 1

	if r := Prescreen(b); r != nil {
		return rejected(r, nil, tracer)
	}
	if !b.HasAnyColored() {
		if m == 0 {
			return feasible(b, nil, nil, tracer)
		}
		return rejected(reasonEmptyBoardNonEmptyStock(m), nil, tracer)
	}

	k := 0
	for _, c := range seq[1:] {
		if int(c) > k {
			k = int(c)
		}
	}
	var cols []ColorStat
	for _, c := range b.Colors() {
		cols = append(cols, MatchColor(b, c))
	}

	s := &mode2Search{
		ctx:     ctx,
		b:       b,
		st:      append([]Color(nil), b.Cells...),
		k:       k,
		seq:     seq,
		pos:     make([]Move, m+1),
		pairs:   b.Pairs(),
		memo:    make(map[string]bool),
		tracer:  tracer,
		nodeCap: opts.nodeCap(),
		needBuf: make([]int, k+1),
		haveBuf: make([]int, k+1),
	}
	ok := s.dfs2(m, 0)
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

	moves := make([]Move, m)
	for t := 1; t <= m; t++ {
		moves[t-1] = s.pos[t]
		moves[t-1].Step = t
	}
	return feasible(b, moves, cols, tracer)
}
