package solver

import (
	"context"
	"strconv"
	"strings"
)

const DefaultEnumerateGuard = 20_000_000

type CountResult struct {
	Total int64 `json:"total"`

	Exact bool `json:"exact"`

	StatesVisited int64 `json:"statesVisited"`
}

type enumState struct {
	g   []Color
	rem []int
	idx int
}

func gridKey(g []Color) string {
	buf := make([]byte, len(g))
	for i, c := range g {
		buf[i] = byte(c)
	}
	return string(buf)
}

func CountStock(ctx context.Context, n int, stock []int, guard int64) CountResult {
	return countConfigs(ctx, n, stock, nil, guard)
}

func CountSequence(ctx context.Context, n int, seq []Color, guard int64) CountResult {
	return countConfigs(ctx, n, nil, seq, guard)
}

func countConfigs(ctx context.Context, n int, stock []int, seq []Color, guard int64) CountResult {
	if guard <= 0 {
		guard = DefaultEnumerateGuard
	}
	ordered := seq != nil
	b := NewBoard(n)
	allPlace := b.AllPlacements()

	finals := make(map[string]struct{})
	vis := make(map[string]struct{})

	stack := make([]enumState, 0, 64)
	s0 := enumState{g: make([]Color, n*n)}
	if ordered {
		s0.idx = 0
	} else {
		s0.rem = append([]int(nil), stock...)
	}
	stack = append(stack, s0)

	exact := true
	var visited int64
	for len(stack) > 0 {
		visited++
		if visited > guard {
			exact = false
			break
		}
		if visited&4095 == 0 && ctx.Err() != nil {
			exact = false
			break
		}

		s := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		var remaining int
		if ordered {
			remaining = len(seq) - s.idx
		} else {
			for _, r := range s.rem {
				remaining += r
			}
		}
		if remaining == 0 {
			finals[gridKey(s.g)] = struct{}{}
			continue
		}

		var vk string
		if ordered {
			vk = gridKey(s.g) + "#" + strconv.Itoa(s.idx)
		} else {
			var sb strings.Builder
			sb.WriteString(gridKey(s.g))
			sb.WriteByte('#')
			for _, r := range s.rem {
				sb.WriteString(strconv.Itoa(r))
				sb.WriteByte(',')
			}
			vk = sb.String()
		}
		if _, seen := vis[vk]; seen {
			continue
		}
		vis[vk] = struct{}{}

		var colours []Color
		if ordered {
			colours = []Color{seq[s.idx]}
		} else {
			for c := 1; c < len(s.rem); c++ {
				if s.rem[c] > 0 {
					colours = append(colours, Color(c))
				}
			}
		}
		for _, col := range colours {
			for _, mv := range allPlace {
				u, v := mv.Cells(n)
				t := enumState{g: append([]Color(nil), s.g...)}
				t.g[u], t.g[v] = col, col
				if ordered {
					t.idx = s.idx + 1
				} else {
					t.rem = append([]int(nil), s.rem...)
					t.rem[col]--
				}
				stack = append(stack, t)
			}
		}
	}

	return CountResult{Total: int64(len(finals)), Exact: exact, StatesVisited: visited}
}
