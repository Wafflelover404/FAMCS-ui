package solver 

import (
	"context"
	"math/rand"
	"testing"
)

func mustGrid(t *testing.T, n int, s string) *Board {
	t.Helper()
	b, err := ParseGrid(n, s)
	if err != nil {
		t.Fatalf("ParseGrid: %v", err)
	}
	return b
}

func TestSolve1_StatementFigure(t *testing.T) {
	b := mustGrid(t, 7, `
		0 0 0 3 0 0 0
		0 0 6 6 1 0 0
		0 0 0 1 5 5 0
		0 0 7 0 0 0 0
		0 5 7 0 0 4 4
		0 0 0 0 0 0 0
		0 0 0 0 0 0 0
	`)
	stock := []int{0, 2, 0, 1, 1, 2, 1, 1}
	res := Solve1(context.Background(), b, stock, Options{})
	if res.Verdict != VerdictFeasible {
		t.Fatalf("verdict = %s, want feasible; reason=%+v", res.Verdict, res.Reason)
	}
	if !Verify(b, res.Moves) {
		t.Fatalf("moves do not replay to target")
	}
}

func TestSolve1_RunExamplesLines(t *testing.T) {
	b := mustGrid(t, 2, "1 1\n1 1\n")

	if res := Solve1(context.Background(), b, []int{0, 2}, Options{}); res.Verdict != VerdictFeasible {
		t.Errorf("stock=2: verdict=%s want feasible (reason=%+v)", res.Verdict, res.Reason)
	}
	if res := Solve1(context.Background(), b, []int{0, 3}, Options{}); res.Verdict != VerdictFeasible {
		t.Errorf("stock=3: verdict=%s want feasible (surplus buried first)", res.Verdict)
	}
	if res := Solve1(context.Background(), b, []int{0, 1}, Options{}); res.Verdict == VerdictFeasible {
		t.Errorf("stock=1: want infeasible (need_c=2 for a solid 1x2 block), got feasible")
	}
}

func TestSolve2_RunExamplesLines(t *testing.T) {
	b := mustGrid(t, 3, "1 2 2\n0 0 0\n0 0 0\n")

	seqOK := []Color{0, 1, 2}
	res := Solve2(context.Background(), b, seqOK, Options{})
	if res.Verdict != VerdictFeasible {
		t.Fatalf("seq=[1,2]: verdict=%s want feasible (reason=%+v)", res.Verdict, res.Reason)
	}
	if !Verify(b, res.Moves) {
		t.Fatalf("seq=[1,2]: moves do not replay to target")
	}

	seqBad := []Color{0, 2, 1}
	if res := Solve2(context.Background(), b, seqBad, Options{}); res.Verdict == VerdictFeasible {
		t.Fatalf("seq=[2,1]: want infeasible, got feasible with moves=%v", res.Moves)
	}
}

func TestCountStock_2x2(t *testing.T) {
	res := CountStock(context.Background(), 2, []int{0, 1, 1}, 0)
	if !res.Exact {
		t.Fatalf("count not exact (guard hit)")
	}
	if res.Total != 28 {
		t.Fatalf("count1 2x2 {1:1,2:1} = %d, want 28", res.Total)
	}
}

func TestCountSequence_2x2(t *testing.T) {
	res := CountSequence(context.Background(), 2, []Color{1, 2}, 0)
	if !res.Exact {
		t.Fatalf("count not exact (guard hit)")
	}
	if res.Total != 16 {
		t.Fatalf("count2 2x2 [1,2] = %d, want 16", res.Total)
	}
}

func TestReachable_NoAdjacentSameColor(t *testing.T) {

	b := mustGrid(t, 2, "1 2\n2 1\n")
	if ok, _ := Reachable(b); ok {
		t.Fatalf("checkerboard board reported reachable")
	}
	r := Prescreen(b)
	if r == nil || r.Kind != ReasonNoAdjacentSameColor {
		t.Fatalf("expected ReasonNoAdjacentSameColor, got %+v", r)
	}
}

func TestReachable_IsolatedCell(t *testing.T) {
	b := mustGrid(t, 3, "1 0 0\n0 0 0\n0 0 0\n")
	r := Prescreen(b)
	if r == nil || r.Kind != ReasonIsolatedCell {
		t.Fatalf("expected ReasonIsolatedCell, got %+v", r)
	}
}

func TestMatchColor_KnownMatching(t *testing.T) {

	b := mustGrid(t, 2, "1 1\n1 1\n")
	st := MatchColor(b, 1)
	if st.CellCount != 4 {
		t.Fatalf("cellCount = %d, want 4", st.CellCount)
	}
	if st.Matching != 2 {
		t.Fatalf("matching = %d, want 2", st.Matching)
	}
	if st.Need != 2 {
		t.Fatalf("need = %d, want 2", st.Need)
	}
}

func TestGenerateRandom_AlwaysSolvable(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))
	const trials = 500
	for i := 0; i < trials; i++ {
		n := 2 + rng.Intn(3)
		m := 1 + rng.Intn(6)
		k := 1 + rng.Intn(4)
		seq := RandomSequence(rng, m, k)
		b, _ := GenerateRandom(rng, n, seq)

		seq1 := make([]Color, m+1)
		copy(seq1[1:], seq)
		res2 := Solve2(context.Background(), b, seq1, Options{})
		if res2.Verdict != VerdictFeasible {
			t.Fatalf("trial %d: mode2 failed on a constructively-reachable board (n=%d,m=%d,seq=%v)\nboard:\n%s",
				i, n, m, seq, b.Format())
		}

		stock := StockOf(seq, k)
		res1 := Solve1(context.Background(), b, stock, Options{})
		if res1.Verdict != VerdictFeasible {
			t.Fatalf("trial %d: mode1 failed on a constructively-reachable board (n=%d,stock=%v)\nboard:\n%s",
				i, n, stock, b.Format())
		}
	}
}

func TestSolve1_EmptyBoardEdgeCases(t *testing.T) {
	b := mustGrid(t, 2, "0 0\n0 0\n")
	if res := Solve1(context.Background(), b, []int{0, 0}, Options{}); res.Verdict != VerdictFeasible {
		t.Errorf("empty board, stock=0: verdict=%s want feasible", res.Verdict)
	}
	if res := Solve1(context.Background(), b, []int{0, 1}, Options{}); res.Verdict != VerdictInfeasible {
		t.Errorf("empty board, stock=1: verdict=%s want infeasible", res.Verdict)
	}
}

func TestSolve2_EmptySequenceOnEmptyBoard(t *testing.T) {
	b := mustGrid(t, 2, "0 0\n0 0\n")
	seq := []Color{0}
	if res := Solve2(context.Background(), b, seq, Options{}); res.Verdict != VerdictFeasible {
		t.Errorf("empty board, m=0: verdict=%s want feasible", res.Verdict)
	}
}

func TestSimulateVisible(t *testing.T) {
	moves := []Move{
		{Step: 1, R: 0, C: 0, Orient: Horizontal, Color: 1},
		{Step: 2, R: 0, C: 0, Orient: Horizontal, Color: 2},
	}
	vis := Visible(2, moves)
	if vis[0] {
		t.Errorf("move 1 should be fully buried")
	}
	if !vis[1] {
		t.Errorf("move 2 should be visible")
	}
}
