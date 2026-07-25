package main 

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"

	"famcs-ui/backend/internal/solver"
)

func gridString(n int, cells []solver.Color) string {
	var buf bytes.Buffer
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			if c > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(strconv.Itoa(int(cells[r*n+c])))
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

func runOracle(bin string, args []string, stdin string) int {
	cmd := exec.Command(bin, args...)
	cmd.Stdin = bytes.NewBufferString(stdin)
	err := cmd.Run()
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	panic(err)
}

func main() {
	bin := flag.String("bin", "./overlap", "path to a compiled overlap.cpp binary")
	trials := flag.Int("trials", 10000, "number of random boards to check")
	seed := flag.Int64("seed", 999, "PRNG seed")
	flag.Parse()

	if _, err := os.Stat(*bin); err != nil {
		fmt.Fprintf(os.Stderr, "oracle binary not found at %q: %v\n", *bin, err)
		fmt.Fprintf(os.Stderr, "build it with: g++ -O2 -std=c++17 -o overlap overlap.cpp\n")
		os.Exit(2)
	}

	rng := rand.New(rand.NewSource(*seed))
	mismatches := 0
	for i := 0; i < *trials; i++ {
		n := 2 + rng.Intn(5)
		m := 1 + rng.Intn(10)
		k := 1 + rng.Intn(6)
		seq := solver.RandomSequence(rng, m, k)
		b, _ := solver.GenerateRandom(rng, n, seq)
		gs := gridString(n, b.Cells)

		args2 := []string{"mode2", strconv.Itoa(n), strconv.Itoa(m)}
		for _, c := range seq {
			args2 = append(args2, strconv.Itoa(int(c)))
		}
		wantRC2 := runOracle(*bin, args2, gs)

		seq1 := make([]solver.Color, m+1)
		copy(seq1[1:], seq)
		res2 := solver.Solve2(context.Background(), b, seq1, solver.Options{})
		gotFeasible2 := res2.Verdict == solver.VerdictFeasible
		if gotFeasible2 != (wantRC2 == 0) {
			mismatches++
			fmt.Printf("[MODE2 MISMATCH t=%d] n=%d m=%d seq=%v want_rc=%d got=%s\nboard:\n%s\n", i, n, m, seq, wantRC2, res2.Verdict, gs)
		} else if gotFeasible2 && !solver.Verify(b, res2.Moves) {
			mismatches++
			fmt.Printf("[MODE2 UNSOUND t=%d] moves do not replay\n", i)
		}

		stock := solver.StockOf(seq, k)
		args1 := []string{"mode1", strconv.Itoa(n), strconv.Itoa(len(stock) - 1)}
		for c := 1; c < len(stock); c++ {
			args1 = append(args1, strconv.Itoa(stock[c]))
		}
		wantRC1 := runOracle(*bin, args1, gs)
		res1 := solver.Solve1(context.Background(), b, stock, solver.Options{})
		gotFeasible1 := res1.Verdict == solver.VerdictFeasible
		if gotFeasible1 != (wantRC1 == 0) {
			mismatches++
			fmt.Printf("[MODE1 MISMATCH t=%d] n=%d stock=%v want_rc=%d got=%s reason=%+v\nboard:\n%s\n", i, n, stock, wantRC1, res1.Verdict, res1.Reason, gs)
		} else if gotFeasible1 && !solver.Verify(b, res1.Moves) {
			mismatches++
			fmt.Printf("[MODE1 UNSOUND t=%d] moves do not replay\n", i)
		}

		if (i+1)%1000 == 0 {
			fmt.Printf("... %d/%d (mismatches so far: %d)\n", i+1, *trials, mismatches)
		}
	}
	fmt.Printf("trials=%d mismatches=%d\n", *trials, mismatches)
	if mismatches > 0 {
		os.Exit(1)
	}
}
