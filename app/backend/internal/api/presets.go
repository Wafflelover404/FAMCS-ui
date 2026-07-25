package api 

func builtinPresets() []Preset {
	return []Preset{
		{
			Slug:  "statement",
			Title: "Из условия",
			N:     6,
			K:     7,
			Cells: []int{
				0, 0, 3, 0, 0, 0,
				0, 6, 6, 1, 0, 0,
				0, 0, 1, 5, 5, 0,
				0, 7, 0, 0, 0, 0,
				5, 7, 0, 0, 4, 4,
				0, 0, 0, 0, 0, 0,
			},
			Stock: []int{2, 0, 1, 1, 2, 1, 1},
		},
		{
			Slug:  "solid-block",
			Title: "Сплошной блок 2×2",
			N:     2,
			K:     1,
			Cells: []int{1, 1, 1, 1},
			Stock: []int{2},
		},
		{
			Slug:  "l-shape",
			Title: "Уголок 3×3",
			N:     3,
			K:     2,
			Cells: []int{
				1, 2, 2,
				0, 0, 0,
				0, 0, 0,
			},
			Stock: []int{1, 2},
		},
		{
			Slug:  "impossible-checkerboard",
			Title: "Недостижимо: шахматка",
			N:     3,
			K:     2,
			Cells: []int{
				1, 2, 1,
				2, 1, 2,
				1, 2, 1,
			},
			Stock: []int{5, 4},
		},
	}
}
