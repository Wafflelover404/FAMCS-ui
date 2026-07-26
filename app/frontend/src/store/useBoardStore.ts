import { create } from "zustand"
import type { Mode, Move, Preset } from "../lib/types"

interface HistoryEntry {
  n: number
  k: number
  cells: number[]
}

interface BoardState {
  n: number
  k: number
  cells: number[]
  mode: Mode
  stock: number[]
  sequence: number[]
  paintColor: number
  presetSlug: string | null
  lastMoves: Move[] | null

  past: HistoryEntry[]
  future: HistoryEntry[]

  setN: (n: number) => void
  setK: (k: number) => void
  setMode: (m: Mode) => void
  setPaintColor: (c: number) => void
  paintCells: (indices: number[], color: number) => void
  setStockAt: (color: number, value: number) => void
  setSequence: (seq: number[]) => void
  addSequenceStep: (color: number) => void
  removeSequenceStep: (i: number) => void
  moveSequenceStep: (from: number, to: number) => void
  loadPreset: (p: Preset) => void
  loadCells: (n: number, k: number, cells: number[]) => void
  clearBoard: () => void
  undo: () => void
  redo: () => void
  setLastMoves: (m: Move[] | null) => void
}

function snapshot(s: BoardState): HistoryEntry {
  return { n: s.n, k: s.k, cells: [...s.cells] }
}

function clampStock(stock: number[], k: number): number[] {
  const out = stock.slice(0, k)
  while (out.length < k) out.push(0)
  return out
}

export const useBoardStore = create<BoardState>((set, get) => ({
  n: 7,
  k: 7,
  cells: new Array(49).fill(0),
  mode: "stock",
  stock: [2, 0, 1, 1, 2, 1, 1],
  sequence: [1, 2, 1, 2],
  paintColor: 1,
  presetSlug: "statement",
  lastMoves: null,
  past: [],
  future: [],

  setN: (n) =>
    set((s) => {
      const cells = new Array(n * n).fill(0)
      return {
        past: [...s.past, snapshot(s)],
        future: [],
        n,
        cells,
        presetSlug: null,
        lastMoves: null,
      }
    }),

  setK: (k) =>
    set((s) => {
      const cells = s.cells.map((c) => (c > k ? 0 : c))
      const paintColor = s.paintColor > k ? Math.max(1, k) : s.paintColor
      return {
        past: [...s.past, snapshot(s)],
        future: [],
        k,
        cells,
        paintColor,
        stock: clampStock(s.stock, k),
        presetSlug: null,
        lastMoves: null,
      }
    }),

  setMode: (mode) => set({ mode }),

  setPaintColor: (c) => set({ paintColor: c }),

  paintCells: (indices, color) =>
    set((s) => {
      if (indices.length === 0) return s
      const cells = [...s.cells]
      for (const i of indices) cells[i] = color
      return { past: [...s.past, snapshot(s)], future: [], cells, presetSlug: null, lastMoves: null }
    }),

  setStockAt: (color, value) =>
    set((s) => {
      const stock = [...s.stock]
      stock[color - 1] = Math.max(0, value)
      return { stock }
    }),

  setSequence: (sequence) => set({ sequence }),

  addSequenceStep: (color) => set((s) => ({ sequence: [...s.sequence, color] })),

  removeSequenceStep: (i) =>
    set((s) => ({ sequence: s.sequence.filter((_, idx) => idx !== i) })),

  moveSequenceStep: (from, to) =>
    set((s) => {
      const seq = [...s.sequence]
      const [item] = seq.splice(from, 1)
      seq.splice(to, 0, item)
      return { sequence: seq }
    }),

  loadPreset: (p) =>
    set((s) => ({
      past: [...s.past, snapshot(s)],
      future: [],
      n: p.n,
      k: p.k,
      cells: [...p.cells],
      stock: clampStock(p.stock, p.k),
      presetSlug: p.slug,
      lastMoves: null,
    })),

  loadCells: (n, k, cells) =>
    set((s) => ({
      past: [...s.past, snapshot(s)],
      future: [],
      n,
      k,
      cells: [...cells],
      stock: clampStock(s.stock, k),
      presetSlug: null,
      lastMoves: null,
    })),

  clearBoard: () =>
    set((s) => ({
      past: [...s.past, snapshot(s)],
      future: [],
      cells: new Array(s.n * s.n).fill(0),
      presetSlug: null,
      lastMoves: null,
    })),

  undo: () => {
    const s = get()
    if (s.past.length === 0) return
    const prev = s.past[s.past.length - 1]
    set({
      past: s.past.slice(0, -1),
      future: [snapshot(s), ...s.future],
      n: prev.n,
      k: prev.k,
      cells: prev.cells,
      presetSlug: null,
    })
  },

  redo: () => {
    const s = get()
    if (s.future.length === 0) return
    const next = s.future[0]
    set({
      past: [...s.past, snapshot(s)],
      future: s.future.slice(1),
      n: next.n,
      k: next.k,
      cells: next.cells,
      presetSlug: null,
    })
  },

  setLastMoves: (m) => set({ lastMoves: m }),
}))
