package solver

import "time"

const DefaultNodeCap int64 = 20_000_000

type Options struct {
	NodeCap int64

	Timeout time.Duration

	Tracer *Tracer
}

func (o Options) nodeCap() int64 {
	if o.NodeCap <= 0 {
		return DefaultNodeCap
	}
	return o.NodeCap
}

type Result struct {
	Verdict Verdict `json:"verdict"`

	Moves []Move `json:"moves,omitempty"`

	Reason *Reason `json:"reason,omitempty"`

	Colors []ColorStat `json:"colors,omitempty"`

	Stats StatsSnap `json:"stats"`

	Verified bool `json:"verified"`
}

func feasible(target *Board, moves []Move, cols []ColorStat, t *Tracer) Result {
	if !ValidMoves(target.N, moves) || !Verify(target, moves) {
		return Result{
			Verdict: VerdictUnproven,
			Reason: &Reason{
				Kind:    ReasonSearchExhausted,
				Message: "internal error: constructed sequence does not replay to the target",
			},
			Colors: cols,
			Stats:  t.Snapshot(),
		}
	}
	t.Solved(len(moves))
	return Result{
		Verdict:  VerdictFeasible,
		Moves:    moves,
		Colors:   cols,
		Stats:    t.Snapshot(),
		Verified: true,
	}
}

func rejected(r *Reason, cols []ColorStat, t *Tracer) Result {
	return Result{Verdict: r.Verdict(), Reason: r, Colors: cols, Stats: t.Snapshot()}
}
