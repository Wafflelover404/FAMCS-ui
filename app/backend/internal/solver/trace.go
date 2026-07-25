package solver

import (
	"sync"
	"time"
)

type EventKind string

const (
	EventPeel EventKind = "peel"

	EventUnpeel EventKind = "unpeel"

	EventPrune EventKind = "prune"

	EventAugment EventKind = "augment"

	EventSolved EventKind = "solved"

	EventExhausted EventKind = "exhausted"

	EventStats EventKind = "stats"
)

type PruneKind string

const (
	PruneBalance PruneKind = "balance"

	PruneIsolated PruneKind = "isolated"

	PruneMemo PruneKind = "memo"

	PruneNoStock PruneKind = "no_stock"

	PruneDeadCell PruneKind = "dead_cell"
)

type Event struct {
	Seq       int64      `json:"seq"`
	Kind      EventKind  `json:"kind"`
	Depth     int        `json:"depth,omitempty"`
	Step      int        `json:"step,omitempty"`
	Move      *Move      `json:"move,omitempty"`
	Cells     []int      `json:"cells,omitempty"`
	Prune     PruneKind  `json:"prune,omitempty"`
	Nodes     int64      `json:"nodes"`
	ElapsedNS int64      `json:"elapsedNs"`
	Stats     *StatsSnap `json:"stats,omitempty"`
}

type Stats struct {
	Nodes      int64
	PeakDepth  int
	Prunes     map[PruneKind]int64
	MemoHits   int64
	MemoStores int64
	MemoSize   int
	Augments   int64
}

type StatsSnap struct {
	Nodes      int64               `json:"nodes"`
	PeakDepth  int                 `json:"peakDepth"`
	Prunes     map[PruneKind]int64 `json:"prunes"`
	MemoHits   int64               `json:"memoHits"`
	MemoStores int64               `json:"memoStores"`
	MemoSize   int                 `json:"memoSize"`
	Augments   int64               `json:"augments"`
	ElapsedNS  int64               `json:"elapsedNs"`
	NodesPerS  float64             `json:"nodesPerSec"`
}

const DefaultDetailCap = 50_000

const DefaultStatsInterval = time.Second / 30

type Tracer struct {
	mu        sync.Mutex
	events    []Event
	seq       int64
	stats     Stats
	start     time.Time
	detailCap int
	sink      func(Event)
	statsIvl  time.Duration
	lastStats time.Time
}

func NewTracer(detailCap int, sink func(Event)) *Tracer {
	if detailCap <= 0 {
		detailCap = DefaultDetailCap
	}
	now := time.Now()
	return &Tracer{
		start:     now,
		lastStats: now,
		detailCap: detailCap,
		sink:      sink,
		statsIvl:  DefaultStatsInterval,
		stats:     Stats{Prunes: make(map[PruneKind]int64, 5)},
	}
}

func (t *Tracer) Node(depth int) int64 {
	if t == nil {
		return 0
	}
	t.stats.Nodes++
	if depth > t.stats.PeakDepth {
		t.stats.PeakDepth = depth
	}
	if t.stats.Nodes > int64(t.detailCap) {
		t.maybeStats()
	}
	return t.stats.Nodes
}

func (t *Tracer) Nodes() int64 {
	if t == nil {
		return 0
	}
	return t.stats.Nodes
}

func (t *Tracer) Peel(depth, step int, m Move, cells []int) {
	t.emit(Event{Kind: EventPeel, Depth: depth, Step: step, Move: &m, Cells: cells})
}

func (t *Tracer) Unpeel(depth, step int, m Move) {
	t.emit(Event{Kind: EventUnpeel, Depth: depth, Step: step, Move: &m})
}

func (t *Tracer) Prune(depth int, kind PruneKind, cells []int) {
	if t == nil {
		return
	}
	t.stats.Prunes[kind]++
	switch kind {
	case PruneMemo:
		t.stats.MemoHits++
	}
	t.emit(Event{Kind: EventPrune, Depth: depth, Prune: kind, Cells: cells})
}

func (t *Tracer) MemoStore(size int) {
	if t == nil {
		return
	}
	t.stats.MemoStores++
	t.stats.MemoSize = size
}

func (t *Tracer) Augment(left, right int) {
	if t == nil {
		return
	}
	t.stats.Augments++
	t.emit(Event{Kind: EventAugment, Cells: []int{left, right}})
}

func (t *Tracer) Solved(moves int) {
	t.emit(Event{Kind: EventSolved, Step: moves})
}

func (t *Tracer) Exhausted() {
	t.emit(Event{Kind: EventExhausted})
}

func (t *Tracer) emit(e Event) {
	if t == nil {
		return
	}
	t.mu.Lock()
	if len(t.events) >= t.detailCap {
		t.mu.Unlock()
		t.maybeStats()
		return
	}
	t.seq++
	e.Seq = t.seq
	e.Nodes = t.stats.Nodes
	e.ElapsedNS = time.Since(t.start).Nanoseconds()
	t.events = append(t.events, e)
	t.mu.Unlock()
	if t.sink != nil {
		t.sink(e)
	}
}

func (t *Tracer) maybeStats() {
	if t == nil || t.sink == nil {
		return
	}
	now := time.Now()
	if now.Sub(t.lastStats) < t.statsIvl {
		return
	}
	t.lastStats = now
	snap := t.Snapshot()
	t.mu.Lock()
	t.seq++
	seq := t.seq
	t.mu.Unlock()
	t.sink(Event{
		Seq:       seq,
		Kind:      EventStats,
		Nodes:     snap.Nodes,
		ElapsedNS: snap.ElapsedNS,
		Stats:     &snap,
	})
}

func (t *Tracer) Snapshot() StatsSnap {
	if t == nil {
		return StatsSnap{Prunes: map[PruneKind]int64{}}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	prunes := make(map[PruneKind]int64, len(t.stats.Prunes))
	for k, v := range t.stats.Prunes {
		prunes[k] = v
	}
	elapsed := time.Since(t.start)
	var nps float64
	if s := elapsed.Seconds(); s > 0 {
		nps = float64(t.stats.Nodes) / s
	}
	return StatsSnap{
		Nodes:      t.stats.Nodes,
		PeakDepth:  t.stats.PeakDepth,
		Prunes:     prunes,
		MemoHits:   t.stats.MemoHits,
		MemoStores: t.stats.MemoStores,
		MemoSize:   t.stats.MemoSize,
		Augments:   t.stats.Augments,
		ElapsedNS:  elapsed.Nanoseconds(),
		NodesPerS:  nps,
	}
}

func (t *Tracer) Events() []Event {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Event, len(t.events))
	copy(out, t.events)
	return out
}

func (t *Tracer) Truncated() bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.events) >= t.detailCap
}
