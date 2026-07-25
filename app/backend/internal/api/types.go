package api 

import (
	"errors"

	"famcs-ui/backend/internal/solver"
)

type BoardInput struct {
	N     int   `json:"n"`
	K     int   `json:"k"`
	Cells []int `json:"cells"`
}

func (bi BoardInput) toBoard() (*solver.Board, error) {
	if bi.N < solver.MinN || bi.N > solver.MaxN {
		return nil, errors.New("n out of range")
	}
	if bi.K < 0 || bi.K > solver.MaxColor {
		return nil, errors.New("k out of range")
	}
	if len(bi.Cells) != bi.N*bi.N {
		return nil, errors.New("cells length must equal n*n")
	}
	b := solver.NewBoard(bi.N)
	for i, v := range bi.Cells {
		if v < 0 || v > bi.K {
			return nil, errors.New("cell colour out of range for k")
		}
		b.Cells[i] = solver.Color(v)
	}
	return b, nil
}

func stockFrom(k int, stock []int) []int {
	out := make([]int, k+1)
	for i := 0; i < k && i < len(stock); i++ {
		out[i+1] = stock[i]
	}
	return out
}

func sequenceFrom(seq []int) []solver.Color {
	out := make([]solver.Color, len(seq)+1)
	for i, v := range seq {
		out[i+1] = solver.Color(v)
	}
	return out
}

type AnalyzeRequest struct {
	BoardInput
	Stock []int `json:"stock,omitempty"`
}

type AnalyzeResponse struct {
	Reachable bool               `json:"reachable"`
	Residue   []int              `json:"residue,omitempty"`
	Reason    *solver.Reason     `json:"reason,omitempty"`
	Colors    []solver.ColorStat `json:"colors"`
	Parity    []int              `json:"parity"`
}

type SolveRequest struct {
	BoardInput
	Mode     string `json:"mode"`
	Stock    []int  `json:"stock,omitempty"`
	Sequence []int  `json:"sequence,omitempty"`
	NodeCap  int64  `json:"nodeCap,omitempty"`
	Trace    bool   `json:"trace,omitempty"`
}

type SolveResponse struct {
	Verdict   solver.Verdict     `json:"verdict"`
	Moves     []solver.Move      `json:"moves,omitempty"`
	Reason    *solver.Reason     `json:"reason,omitempty"`
	Colors    []solver.ColorStat `json:"colors,omitempty"`
	Stats     solver.StatsSnap   `json:"stats"`
	Verified  bool               `json:"verified"`
	Trace     []solver.Event     `json:"trace,omitempty"`
	Truncated bool               `json:"truncated,omitempty"`
}

type CountRequest struct {
	N        int    `json:"n"`
	K        int    `json:"k"`
	Mode     string `json:"mode"`
	Stock    []int  `json:"stock,omitempty"`
	Sequence []int  `json:"sequence,omitempty"`
	Guard    int64  `json:"guard,omitempty"`
}

type SimulateRequest struct {
	N     int           `json:"n"`
	Moves []solver.Move `json:"moves"`
}

type SimulateResponse struct {
	Frames   []solver.Frame       `json:"frames"`
	Coverage []solver.CellHistory `json:"coverage"`
	Visible  []bool               `json:"visible"`
}

type RandomRequest struct {
	N    int   `json:"n"`
	M    int   `json:"m"`
	K    int   `json:"k"`
	Seed int64 `json:"seed,omitempty"`
}

type RandomResponse struct {
	N        int           `json:"n"`
	K        int           `json:"k"`
	Cells    []int         `json:"cells"`
	Sequence []int         `json:"sequence"`
	Stock    []int         `json:"stock"`
	Moves    []solver.Move `json:"moves"`
}

type Preset struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	N     int    `json:"n"`
	K     int    `json:"k"`
	Cells []int  `json:"cells"`
	Stock []int  `json:"stock"`
}
