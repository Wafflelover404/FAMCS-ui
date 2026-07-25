package solver 

func Simulate(n int, moves []Move) *Board {
	b := NewBoard(n)
	for _, m := range moves {
		u, v := m.Cells(n)
		b.Cells[u] = m.Color
		b.Cells[v] = m.Color
	}
	return b
}

func Verify(target *Board, moves []Move) bool {
	return Simulate(target.N, moves).Equal(target)
}

func ValidMoves(n int, moves []Move) bool {
	for _, m := range moves {
		if m.R < 0 || m.C < 0 || m.R >= n || m.C >= n {
			return false
		}
		if m.Orient == Horizontal && m.C+1 >= n {
			return false
		}
		if m.Orient == Vertical && m.R+1 >= n {
			return false
		}
		if m.Color == Background || m.Color > MaxColor {
			return false
		}
	}
	return true
}

type Frame struct {
	Step  int    `json:"step"`
	Cells Colors `json:"cells"`
}

func Frames(n int, moves []Move) []Frame {
	out := make([]Frame, 0, len(moves)+1)
	cur := make([]Color, n*n)
	snap := func(step int) {
		cp := make([]Color, len(cur))
		copy(cp, cur)
		out = append(out, Frame{Step: step, Cells: cp})
	}
	snap(0)
	for i, m := range moves {
		u, v := m.Cells(n)
		cur[u], cur[v] = m.Color, m.Color
		snap(i + 1)
	}
	return out
}

type CellHistory struct {
	Cell int `json:"cell"`

	Steps []int `json:"steps"`

	Winner int `json:"winner"`

	Color Color `json:"color"`
}

func Coverage(n int, moves []Move) []CellHistory {
	out := make([]CellHistory, n*n)
	for i := range out {
		out[i].Cell = i
	}
	for i, m := range moves {
		step := i + 1
		u, v := m.Cells(n)
		for _, cell := range [2]int{u, v} {
			out[cell].Steps = append(out[cell].Steps, step)
			out[cell].Winner = step
			out[cell].Color = m.Color
		}
	}
	return out
}

func Visible(n int, moves []Move) []bool {
	final := make([]int, n*n)
	for i, m := range moves {
		u, v := m.Cells(n)
		final[u], final[v] = i+1, i+1
	}
	out := make([]bool, len(moves))
	for i, m := range moves {
		u, v := m.Cells(n)
		out[i] = final[u] == i+1 || final[v] == i+1
	}
	return out
}
