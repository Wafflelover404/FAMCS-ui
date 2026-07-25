package solver

import "math/rand"

func GenerateRandom(rng *rand.Rand, n int, colors []Color) (*Board, []Move) {
	b := NewBoard(n)
	if n < 2 {
		return b, nil
	}
	moves := make([]Move, 0, len(colors))
	for i, col := range colors {
		var mv Move
		if rng.Float64() < 0.5 {
			r := rng.Intn(n)
			c := rng.Intn(n - 1)
			mv = Move{Step: i + 1, R: r, C: c, Orient: Horizontal, Color: col}
		} else {
			r := rng.Intn(n - 1)
			c := rng.Intn(n)
			mv = Move{Step: i + 1, R: r, C: c, Orient: Vertical, Color: col}
		}
		u, v := mv.Cells(n)
		b.Cells[u], b.Cells[v] = col, col
		moves = append(moves, mv)
	}
	return b, moves
}

func RandomSequence(rng *rand.Rand, m, k int) []Color {
	if k < 1 {
		k = 1
	}
	out := make([]Color, m)
	for i := range out {
		out[i] = Color(rng.Intn(k) + 1)
	}
	return out
}

func StockOf(colors []Color, k int) []int {
	out := make([]int, k+1)
	for _, c := range colors {
		if int(c) >= len(out) {
			grown := make([]int, int(c)+1)
			copy(grown, out)
			out = grown
		}
		out[c]++
	}
	return out
}
