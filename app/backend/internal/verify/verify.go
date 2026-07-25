package verify 

import (
	"context"
	"fmt"
	"math/rand"

	"famcs-ui/backend/internal/solver"
)

type Failure struct {
	Trial int    `json:"trial"`
	Mode  string `json:"mode"`
	N     int    `json:"n"`
	Board string `json:"board"`
	Info  string `json:"info"`
}

type Result struct {
	Trials   int       `json:"trials"`
	Passed   int       `json:"passed"`
	Failed   int       `json:"failed"`
	Failures []Failure `json:"failures,omitempty"`
}

func RunStressSuite(ctx context.Context, trials int, seed int64) Result {
	rng := rand.New(rand.NewSource(seed))
	res := Result{Trials: trials}

	for t := 0; t < trials; t++ {
		if ctx.Err() != nil {
			break
		}
		n := 2 + rng.Intn(3)
		m := 1 + rng.Intn(6)
		k := 1 + rng.Intn(4)
		seq := solver.RandomSequence(rng, m, k)
		b, _ := solver.GenerateRandom(rng, n, seq)

		seq1 := make([]solver.Color, m+1)
		copy(seq1[1:], seq)
		r2 := solver.Solve2(ctx, b, seq1, solver.Options{})
		if r2.Verdict != solver.VerdictFeasible {
			res.Failed++
			res.Failures = append(res.Failures, Failure{
				Trial: t, Mode: "mode2", N: n, Board: b.Format(),
				Info: fmt.Sprintf("verdict=%s reason=%v", r2.Verdict, r2.Reason),
			})
			continue
		}

		stock := solver.StockOf(seq, k)
		r1 := solver.Solve1(ctx, b, stock, solver.Options{})
		if r1.Verdict != solver.VerdictFeasible {
			res.Failed++
			res.Failures = append(res.Failures, Failure{
				Trial: t, Mode: "mode1", N: n, Board: b.Format(),
				Info: fmt.Sprintf("verdict=%s reason=%v", r1.Verdict, r1.Reason),
			})
			continue
		}
		res.Passed++
	}
	return res
}

func RunMatchingCrossCheck(maxN int) Result {
	res := Result{}
	for n := 2; n <= maxN; n++ {
		if n*n > 12 {
			continue
		}
		for mask := 0; mask < (1 << (n * n)); mask++ {
			res.Trials++
			cells := make([]solver.Color, n*n)
			for i := range cells {
				if mask&(1<<i) != 0 {
					cells[i] = 1
				}
			}
			b := &solver.Board{N: n, Cells: cells}
			exact := bruteForceMaxMatching(b, 1)
			got := solver.MatchColor(b, 1).Matching
			if exact != got {
				res.Failed++
				res.Failures = append(res.Failures, Failure{
					N: n, Board: b.Format(),
					Info: fmt.Sprintf("kuhn=%d brute=%d", got, exact),
				})
				continue
			}
			res.Passed++
		}
	}
	return res
}

func bruteForceMaxMatching(b *solver.Board, color solver.Color) int {
	cells := b.CellsOf(color)
	pairs := [][2]int{}
	for _, u := range cells {
		r, c := b.RowCol(u)
		if c+1 < b.N && b.At(r, c+1) == color {
			pairs = append(pairs, [2]int{u, u + 1})
		}
		if r+1 < b.N && b.At(r+1, c) == color {
			pairs = append(pairs, [2]int{u, u + b.N})
		}
	}
	best := 0
	var rec func(idx int, used map[int]bool, count int)
	rec = func(idx int, used map[int]bool, count int) {
		if count > best {
			best = count
		}
		for i := idx; i < len(pairs); i++ {
			p := pairs[i]
			if used[p[0]] || used[p[1]] {
				continue
			}
			used[p[0]], used[p[1]] = true, true
			rec(i+1, used, count+1)
			delete(used, p[0])
			delete(used, p[1])
		}
	}
	rec(0, map[int]bool{}, 0)
	return best
}
