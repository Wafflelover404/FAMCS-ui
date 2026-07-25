package solver

import "fmt"

type Verdict string

const (
	VerdictFeasible Verdict = "feasible"

	VerdictInfeasible Verdict = "infeasible"

	VerdictUnproven Verdict = "unproven"
)

type ReasonKind string

const (
	ReasonNoAdjacentSameColor ReasonKind = "no_adjacent_same_color"

	ReasonGreedyPeelStuck ReasonKind = "greedy_peel_stuck"

	ReasonInsufficientStock ReasonKind = "insufficient_stock"

	ReasonIsolatedCell ReasonKind = "isolated_cell"

	ReasonColorNotInStock ReasonKind = "color_not_in_stock"

	ReasonEmptyBoardNonEmptyStock ReasonKind = "empty_board_nonempty_stock"

	ReasonNoPlacements ReasonKind = "no_placements"

	ReasonSearchExhausted ReasonKind = "search_exhausted"

	ReasonNodeCapExceeded ReasonKind = "node_cap_exceeded"

	ReasonCanceled ReasonKind = "canceled"
)

type Reason struct {
	Kind    ReasonKind `json:"kind"`
	Cells   []int      `json:"cells,omitempty"`
	Color   Color      `json:"color,omitempty"`
	Need    int        `json:"need,omitempty"`
	Have    int        `json:"have,omitempty"`
	Message string     `json:"message"`
}

func (r *Reason) Verdict() Verdict {
	if r == nil {
		return VerdictFeasible
	}
	switch r.Kind {
	case ReasonNodeCapExceeded, ReasonCanceled:
		return VerdictUnproven
	default:
		return VerdictInfeasible
	}
}

func (r *Reason) Error() string {
	if r == nil {
		return ""
	}
	return r.Message
}

func reasonNoAdjacentSameColor() *Reason {
	return &Reason{
		Kind:    ReasonNoAdjacentSameColor,
		Message: "no two adjacent cells share a colour, so no domino could have been placed last",
	}
}

func reasonGreedyPeelStuck(cells []int) *Reason {
	return &Reason{
		Kind:    ReasonGreedyPeelStuck,
		Cells:   cells,
		Message: fmt.Sprintf("greedy reverse peel stalled with %d cell(s) still unexplained", len(cells)),
	}
}

func reasonInsufficientStock(c Color, need, have int, cells []int) *Reason {
	return &Reason{
		Kind:  ReasonInsufficientStock,
		Cells: cells,
		Color: c,
		Need:  need,
		Have:  have,
		Message: fmt.Sprintf("colour %d needs at least %d domino(es) but only %d are available",
			c, need, have),
	}
}

func reasonIsolatedCell(cell int) *Reason {
	return &Reason{
		Kind:    ReasonIsolatedCell,
		Cells:   []int{cell},
		Message: "a coloured cell has no non-background neighbour, so no domino could cover it",
	}
}

func reasonColorNotInStock(c Color, cells []int) *Reason {
	return &Reason{
		Kind:    ReasonColorNotInStock,
		Cells:   cells,
		Color:   c,
		Message: fmt.Sprintf("the picture shows colour %d, which is not in the stock", c),
	}
}

func reasonEmptyBoardNonEmptyStock(remaining int) *Reason {
	return &Reason{
		Kind: ReasonEmptyBoardNonEmptyStock,
		Need: remaining,
		Message: fmt.Sprintf("the board is empty but %d domino(es) must still be placed; "+
			"any placement would remain visible", remaining),
	}
}

func reasonNoPlacements() *Reason {
	return &Reason{
		Kind:    ReasonNoPlacements,
		Message: "no two adjacent non-background cells exist, so no domino fits on the board",
	}
}

func reasonSearchExhausted() *Reason {
	return &Reason{
		Kind:    ReasonSearchExhausted,
		Message: "the search space was explored exhaustively and contains no solution",
	}
}

func reasonNodeCapExceeded(cap int64) *Reason {
	return &Reason{
		Kind:    ReasonNodeCapExceeded,
		Need:    int(cap),
		Message: fmt.Sprintf("search budget of %d nodes exhausted; no conclusion reached", cap),
	}
}

func reasonCanceled() *Reason {
	return &Reason{
		Kind:    ReasonCanceled,
		Message: "search was cancelled before reaching a conclusion",
	}
}
