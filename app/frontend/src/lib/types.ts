export type Orient = "h" | "v" 

export interface Move {
  step: number
  r: number
  c: number
  orient: Orient
  color: number
}

export type Verdict = "feasible" | "infeasible" | "unproven"

export type ReasonKind =
  | "no_adjacent_same_color"
  | "greedy_peel_stuck"
  | "insufficient_stock"
  | "isolated_cell"
  | "color_not_in_stock"
  | "empty_board_nonempty_stock"
  | "no_placements"
  | "search_exhausted"
  | "node_cap_exceeded"
  | "canceled"

export interface Reason {
  kind: ReasonKind
  cells?: number[]
  color?: number
  need?: number
  have?: number
  message: string
}

export interface Edge {
  a: number
  b: number
}

export interface ColorStat {
  color: number
  cells: number[]
  cellCount: number
  matching: number
  need: number
  have: number
  deficit: number
  edges: Edge[] | null
  singles: number[] | null
}

export interface StatsSnap {
  nodes: number
  peakDepth: number
  prunes: Record<string, number>
  memoHits: number
  memoStores: number
  memoSize: number
  augments: number
  elapsedNs: number
  nodesPerSec: number
}

export interface TraceEvent {
  seq: number
  kind: "peel" | "unpeel" | "prune" | "augment" | "solved" | "exhausted" | "stats"
  depth?: number
  step?: number
  move?: Move
  cells?: number[]
  prune?: string
  nodes: number
  elapsedNs: number
  stats?: StatsSnap
}

export interface AnalyzeResponse {
  reachable: boolean
  residue?: number[]
  reason?: Reason
  colors: ColorStat[]
  parity: number[]
}

export interface SolveResponse {
  verdict: Verdict
  moves?: Move[]
  reason?: Reason
  colors?: ColorStat[]
  stats: StatsSnap
  verified: boolean
  trace?: TraceEvent[]
  truncated?: boolean
}

export interface CountResponse {
  total: number
  exact: boolean
  statesVisited: number
}

export interface Frame {
  step: number
  cells: number[]
}

export interface CellHistory {
  cell: number
  steps: number[] | null
  winner: number
  color: number
}

export interface SimulateResponse {
  frames: Frame[]
  coverage: CellHistory[]
  visible: boolean[]
}

export interface RandomResponse {
  n: number
  k: number
  cells: number[]
  sequence: number[]
  stock: number[]
  moves: Move[]
}

export interface Preset {
  slug: string
  title: string
  n: number
  k: number
  cells: number[]
  stock: number[]
}

export type Mode = "stock" | "sequence"
