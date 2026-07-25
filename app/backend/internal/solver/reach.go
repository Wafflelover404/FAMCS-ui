package solver 

type cellState int8

const (
	stBackground cellState = -1
	stFree       cellState = 0
)

func initialState(b *Board) []cellState {
	s := make([]cellState, len(b.Cells))
	for i, c := range b.Cells {
		if c == Background {
			s[i] = stBackground
		} else {
			s[i] = cellState(c)
		}
	}
	return s
}

func peelable(sa, sb cellState) (Color, bool) {
	if sa == stBackground || sb == stBackground {
		return 0, false
	}
	if sa == stFree && sb == stFree {

		return 0, false
	}
	col := sa
	if col == stFree {
		col = sb
	}
	if (sa == stFree || sa == col) && (sb == stFree || sb == col) {
		return Color(col), true
	}
	return 0, false
}

func Reachable(b *Board) (bool, []int) {
	return reachableWith(b, nil)
}

func ReachableTraced(b *Board, t *Tracer) (bool, []int) {
	return reachableWith(b, t)
}

func reachableWith(b *Board, t *Tracer) (bool, []int) {
	s := initialState(b)
	pairs := b.Pairs()

	for changed := true; changed; {
		changed = false
		for _, p := range pairs {
			sa, sb := s[p.A], s[p.B]
			if sa == stFree && sb == stFree {
				continue
			}
			col, ok := peelable(sa, sb)
			if !ok {
				continue
			}
			s[p.A], s[p.B] = stFree, stFree
			changed = true
			if t != nil {
				m := p.Move
				m.Color = col
				t.Peel(0, 0, m, []int{p.A, p.B})
			}
		}
	}

	var residue []int
	for i, v := range s {
		if v > stFree {
			residue = append(residue, i)
		}
	}
	return len(residue) == 0, residue
}

func HasAdjacentSameColor(b *Board) bool {
	for _, p := range b.Pairs() {
		if b.Cells[p.A] == b.Cells[p.B] {
			return true
		}
	}
	return false
}

func IsolatedColoredCell(b *Board) int {
	nbrs := make([]int, 0, 4)
	for i := range b.Cells {
		if b.IsBackground(i) {
			continue
		}
		nbrs = b.Neighbors4(i, nbrs)
		ok := false
		for _, v := range nbrs {
			if !b.IsBackground(v) {
				ok = true
				break
			}
		}
		if !ok {
			return i
		}
	}
	return -1
}

func Prescreen(b *Board) *Reason {
	if !b.HasAnyColored() {
		return nil
	}
	if i := IsolatedColoredCell(b); i >= 0 {
		return reasonIsolatedCell(i)
	}
	if len(b.Pairs()) == 0 {
		return reasonNoPlacements()
	}
	if !HasAdjacentSameColor(b) {
		return reasonNoAdjacentSameColor()
	}
	if ok, residue := Reachable(b); !ok {
		return reasonGreedyPeelStuck(residue)
	}
	return nil
}
