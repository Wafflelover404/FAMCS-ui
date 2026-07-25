package solver 

var (
	dRow = [4]int{0, 0, 1, -1}
	dCol = [4]int{1, -1, 0, 0}
)

type Edge struct {
	A int `json:"a"`
	B int `json:"b"`
}

type ColorStat struct {
	Color Color `json:"color"`

	Cells []int `json:"cells"`

	CellCount int `json:"cellCount"`

	Matching int `json:"matching"`

	Need int `json:"need"`

	Have int `json:"have"`

	Deficit int `json:"deficit"`

	Edges []Edge `json:"edges"`

	Singles []int `json:"singles"`
}

func MatchColor(b *Board, c Color) ColorStat {
	return matchColor(b, c, nil)
}

func matchColor(b *Board, c Color, t *Tracer) ColorStat {
	st := ColorStat{Color: c, Cells: b.CellsOf(c)}
	st.CellCount = len(st.Cells)
	if st.CellCount == 0 {
		return st
	}

	k := &kuhn{b: b, color: c, matchTo: make([]int, len(b.Cells)), tracer: t}
	for i := range k.matchTo {
		k.matchTo[i] = -1
	}

	for _, u := range st.Cells {
		if b.Parity(u) != 0 {
			continue
		}
		k.used = make([]bool, len(b.Cells))
		if k.try(u) {
			st.Matching++
		}
	}

	st.Need = st.CellCount - st.Matching
	matched := make([]bool, len(b.Cells))
	for v, u := range k.matchTo {
		if u >= 0 {
			st.Edges = append(st.Edges, Edge{A: u, B: v})
			matched[u], matched[v] = true, true
		}
	}
	for _, u := range st.Cells {
		if !matched[u] {
			st.Singles = append(st.Singles, u)
		}
	}
	return st
}

type kuhn struct {
	b       *Board
	color   Color
	matchTo []int
	used    []bool
	tracer  *Tracer
}

func (k *kuhn) try(u int) bool {
	n := k.b.N
	r, c := u/n, u%n
	for d := 0; d < 4; d++ {
		rr, cc := r+dRow[d], c+dCol[d]
		if !k.b.InBounds(rr, cc) {
			continue
		}
		v := rr*n + cc
		if k.b.Cells[v] != k.color || k.used[v] {
			continue
		}
		k.used[v] = true
		if k.matchTo[v] == -1 || k.try(k.matchTo[v]) {
			k.matchTo[v] = u
			k.tracer.Augment(u, v)
			return true
		}
	}
	return false
}

func StockNeed(b *Board, stock []int) ([]ColorStat, *Reason) {
	var out []ColorStat
	for _, c := range b.Colors() {
		st := MatchColor(b, c)
		if int(c) < len(stock) {
			st.Have = stock[c]
		}
		if st.Need > st.Have {
			st.Deficit = st.Need - st.Have
		}
		out = append(out, st)
	}
	for _, st := range out {
		if st.Have == 0 && st.CellCount > 0 {
			return out, reasonColorNotInStock(st.Color, st.Cells)
		}
	}
	for _, st := range out {
		if st.Deficit > 0 {
			return out, reasonInsufficientStock(st.Color, st.Need, st.Have, st.Singles)
		}
	}
	return out, nil
}
