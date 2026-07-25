package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"famcs-ui/backend/internal/solver"

	"context"
)

func fail() {
	os.Exit(1)
}

func readGrid(r *bufio.Reader, n, k int) (*solver.Board, bool) {
	b := solver.NewBoard(n)
	for i := 0; i < n*n; i++ {
		var v int
		if _, err := fmt.Fscan(r, &v); err != nil {
			return nil, false
		}
		if v < 0 || v > k {
			return nil, false
		}
		b.Cells[i] = solver.Color(v)
	}
	return b, true
}

func atoiOrExit(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		fail()
	}
	return v
}

func main() {
	explain := flagBool("--explain")
	asJSON := flagBool("--json")
	nodeCap := flagInt64("--node-cap", 0)
	args := stripFlags()

	if len(args) < 2 {
		fail()
	}
	mode := args[1]

	switch mode {
	case "mode1", "count1":
		runMode1(args, mode == "count1", explain, asJSON, nodeCap)
	case "mode2", "count2":
		runMode2(args, mode == "count2", explain, asJSON, nodeCap)
	default:
		fail()
	}
}

func runMode1(args []string, counting, explain, asJSON bool, nodeCap int64) {
	if len(args) < 4 {
		fail()
	}
	n := atoiOrExit(args[2])
	k := atoiOrExit(args[3])
	if len(args) < 4+k {
		fail()
	}
	stock := make([]int, k+1)
	for c := 1; c <= k; c++ {
		stock[c] = atoiOrExit(args[3+c])
	}

	if counting {
		res := solver.CountStock(context.Background(), n, stock, nodeCap)
		printCount(res, explain, asJSON)
		os.Exit(1)
	}

	r := bufio.NewReader(os.Stdin)
	b, ok := readGrid(r, n, k)
	if !ok {
		fail()
	}
	opts := solver.Options{Tracer: solver.NewTracer(1, nil)}
	if nodeCap > 0 {
		opts.NodeCap = nodeCap
	}
	res := solver.Solve1(context.Background(), b, stock, opts)
	printSolve(res, explain, asJSON)
	if res.Verdict != solver.VerdictFeasible {
		os.Exit(1)
	}
}

func runMode2(args []string, counting, explain, asJSON bool, nodeCap int64) {
	if len(args) < 4 {
		fail()
	}
	n := atoiOrExit(args[2])
	m := atoiOrExit(args[3])
	if len(args) < 4+m {
		fail()
	}
	seq := make([]solver.Color, m+1)
	k := 0
	for t := 1; t <= m; t++ {
		v := atoiOrExit(args[3+t])
		seq[t] = solver.Color(v)
		if v > k {
			k = v
		}
	}

	if counting {
		flat := seq[1:]
		res := solver.CountSequence(context.Background(), n, flat, nodeCap)
		printCount(res, explain, asJSON)
		os.Exit(1)
	}

	r := bufio.NewReader(os.Stdin)
	b, ok := readGrid(r, n, k)
	if !ok {
		fail()
	}
	opts := solver.Options{Tracer: solver.NewTracer(1, nil)}
	if nodeCap > 0 {
		opts.NodeCap = nodeCap
	}
	res := solver.Solve2(context.Background(), b, seq, opts)
	printSolve(res, explain, asJSON)
	if res.Verdict != solver.VerdictFeasible {
		os.Exit(1)
	}
}

func printSolve(res solver.Result, explain, asJSON bool) {
	if asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(res)
		return
	}
	if !explain {
		return
	}
	switch res.Verdict {
	case solver.VerdictFeasible:
		for _, m := range res.Moves {
			fmt.Printf("%02d %d (%d,%d) %s\n", m.Step, m.Color, m.R, m.C, m.Orient)
		}
	case solver.VerdictInfeasible:
		fmt.Printf("невозможно: %s\n", res.Reason.Message)
	case solver.VerdictUnproven:
		fmt.Printf("не доказано: %s\n", res.Reason.Message)
	}
}

func printCount(res solver.CountResult, explain, asJSON bool) {
	if asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(res)
		return
	}
	fmt.Println(res.Total)
	if explain && !res.Exact {
		fmt.Fprintln(os.Stderr, "warning: guard exhausted, this is a lower bound")
	}
}

var rawArgs = os.Args

func flagBool(name string) bool {
	for _, a := range rawArgs {
		if a == name {
			return true
		}
	}
	return false
}

func flagInt64(name string, def int64) int64 {
	for i, a := range rawArgs {
		if a == name && i+1 < len(rawArgs) {
			if v, err := strconv.ParseInt(rawArgs[i+1], 10, 64); err == nil {
				return v
			}
		}
	}
	return def
}

func stripFlags() []string {
	out := make([]string, 0, len(rawArgs))
	skip := false
	for _, a := range rawArgs {
		if skip {
			skip = false
			continue
		}
		if a == "--explain" || a == "--json" {
			continue
		}
		if a == "--node-cap" {
			skip = true
			continue
		}
		out = append(out, a)
	}
	return out
}
