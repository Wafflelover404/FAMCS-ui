package solver

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

const (
	MinN     = 2
	MaxN     = 10
	MaxColor = 9
	MaxMoves = 100
	MaxStock = 10
)

type Color uint8

const Background Color = 0

type Colors []Color

func (c Colors) MarshalJSON() ([]byte, error) {
	ints := make([]int, len(c))
	for i, v := range c {
		ints[i] = int(v)
	}
	return json.Marshal(ints)
}

func (c *Colors) UnmarshalJSON(data []byte) error {
	var ints []int
	if err := json.Unmarshal(data, &ints); err != nil {
		return err
	}
	out := make(Colors, len(ints))
	for i, v := range ints {
		out[i] = Color(v)
	}
	*c = out
	return nil
}

type Orient uint8

const (
	Horizontal Orient = iota
	Vertical
)

func (o Orient) String() string {
	if o == Horizontal {
		return "horizontal"
	}
	return "vertical"
}

func (o Orient) MarshalJSON() ([]byte, error) {
	if o == Horizontal {
		return []byte(`"h"`), nil
	}
	return []byte(`"v"`), nil
}

func (o *Orient) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	switch s {
	case "h", "horizontal":
		*o = Horizontal
	case "v", "vertical":
		*o = Vertical
	default:
		return fmt.Errorf("invalid orientation %q", s)
	}
	return nil
}

type Move struct {
	Step   int    `json:"step"`
	R      int    `json:"r"`
	C      int    `json:"c"`
	Orient Orient `json:"orient"`
	Color  Color  `json:"color"`
}

func (m Move) Cells(n int) (int, int) {
	u := m.R*n + m.C
	if m.Orient == Horizontal {
		return u, u + 1
	}
	return u, u + n
}

func (m Move) String() string {
	return fmt.Sprintf("%02d color=%d (%d,%d) %s", m.Step, m.Color, m.R, m.C, m.Orient)
}

type Board struct {
	N     int    `json:"n"`
	Cells Colors `json:"cells"`
}

func NewBoard(n int) *Board {
	return &Board{N: n, Cells: make([]Color, n*n)}
}

func (b *Board) Clone() *Board {
	c := &Board{N: b.N, Cells: make([]Color, len(b.Cells))}
	copy(c.Cells, b.Cells)
	return c
}

func (b *Board) At(r, c int) Color { return b.Cells[r*b.N+c] }

func (b *Board) Index(r, c int) int { return r*b.N + c }

func (b *Board) RowCol(i int) (int, int) { return i / b.N, i % b.N }

func (b *Board) InBounds(r, c int) bool { return r >= 0 && r < b.N && c >= 0 && c < b.N }

func (b *Board) IsBackground(i int) bool { return b.Cells[i] == Background }

func (b *Board) Parity(i int) int {
	r, c := b.RowCol(i)
	return (r + c) & 1
}

func (b *Board) Colors() []Color {
	var seen [MaxColor + 1]bool
	for _, c := range b.Cells {
		if c != Background {
			seen[c] = true
		}
	}
	out := make([]Color, 0, MaxColor)
	for c := Color(1); c <= MaxColor; c++ {
		if seen[c] {
			out = append(out, c)
		}
	}
	return out
}

func (b *Board) CellsOf(c Color) []int {
	var out []int
	for i, v := range b.Cells {
		if v == c {
			out = append(out, i)
		}
	}
	return out
}

func (b *Board) HasAnyColored() bool {
	for _, c := range b.Cells {
		if c != Background {
			return true
		}
	}
	return false
}

func (b *Board) Neighbors4(i int, dst []int) []int {
	dst = dst[:0]
	r, c := b.RowCol(i)
	if c+1 < b.N {
		dst = append(dst, i+1)
	}
	if c > 0 {
		dst = append(dst, i-1)
	}
	if r+1 < b.N {
		dst = append(dst, i+b.N)
	}
	if r > 0 {
		dst = append(dst, i-b.N)
	}
	return dst
}

type Pair struct {
	A, B int
	Move Move
}

func (b *Board) Pairs() []Pair {
	var out []Pair
	for r := 0; r < b.N; r++ {
		for c := 0; c < b.N; c++ {
			u := b.Index(r, c)
			if b.IsBackground(u) {
				continue
			}
			if c+1 < b.N && !b.IsBackground(u+1) {
				out = append(out, Pair{A: u, B: u + 1, Move: Move{R: r, C: c, Orient: Horizontal}})
			}
			if r+1 < b.N && !b.IsBackground(u+b.N) {
				out = append(out, Pair{A: u, B: u + b.N, Move: Move{R: r, C: c, Orient: Vertical}})
			}
		}
	}
	return out
}

func (b *Board) AllPlacements() []Move {
	out := make([]Move, 0, 2*b.N*(b.N-1))
	for r := 0; r < b.N; r++ {
		for c := 0; c < b.N; c++ {
			if c+1 < b.N {
				out = append(out, Move{R: r, C: c, Orient: Horizontal})
			}
			if r+1 < b.N {
				out = append(out, Move{R: r, C: c, Orient: Vertical})
			}
		}
	}
	return out
}

func (b *Board) Hash() uint64 {
	h := fnv.New64a()
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(b.N))
	_, _ = h.Write(buf[:])
	for _, c := range b.Cells {
		_, _ = h.Write([]byte{byte(c)})
	}
	return h.Sum64()
}

func (b *Board) Equal(o *Board) bool {
	if b == nil || o == nil || b.N != o.N || len(b.Cells) != len(o.Cells) {
		return false
	}
	for i := range b.Cells {
		if b.Cells[i] != o.Cells[i] {
			return false
		}
	}
	return true
}

func (b *Board) Validate(k int) error {
	if b.N < MinN || b.N > MaxN {
		return fmt.Errorf("n=%d out of range [%d,%d]", b.N, MinN, MaxN)
	}
	if len(b.Cells) != b.N*b.N {
		return fmt.Errorf("board has %d cells, expected %d", len(b.Cells), b.N*b.N)
	}
	if k < 0 || k > MaxColor {
		return fmt.Errorf("k=%d out of range [0,%d]", k, MaxColor)
	}
	for i, c := range b.Cells {
		if int(c) > k {
			r, col := b.RowCol(i)
			return fmt.Errorf("cell (%d,%d) has colour %d > k=%d", r, col, c, k)
		}
	}
	return nil
}

func ParseGrid(n int, s string) (*Board, error) {
	if n < MinN || n > MaxN {
		return nil, fmt.Errorf("n=%d out of range [%d,%d]", n, MinN, MaxN)
	}
	fields := strings.Fields(s)
	if len(fields) != n*n {
		return nil, fmt.Errorf("expected %d values for a %d×%d grid, got %d", n*n, n, n, len(fields))
	}
	b := NewBoard(n)
	for i, f := range fields {
		v, err := strconv.Atoi(f)
		if err != nil {
			return nil, fmt.Errorf("cell %d: %q is not an integer", i, f)
		}
		if v < 0 || v > MaxColor {
			return nil, fmt.Errorf("cell %d: colour %d out of range [0,%d]", i, v, MaxColor)
		}
		b.Cells[i] = Color(v)
	}
	return b, nil
}

func (b *Board) Format() string {
	var sb strings.Builder
	for r := 0; r < b.N; r++ {
		for c := 0; c < b.N; c++ {
			if c > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(strconv.Itoa(int(b.At(r, c))))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
