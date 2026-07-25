import { create } from "zustand"
import type { AnalyzeResponse, CountResponse, SolveResponse } from "../lib/types"

interface SolveState {
  analyzing: boolean
  analysis: AnalyzeResponse | null
  analyzeError: string | null

  solving: boolean
  solveResult: SolveResponse | null
  solveError: string | null

  counting: boolean
  countResult: CountResponse | null
  countError: string | null

  setAnalyzing: (v: boolean) => void
  setAnalysis: (a: AnalyzeResponse | null, err?: string | null) => void
  setSolving: (v: boolean) => void
  setSolveResult: (r: SolveResponse | null, err?: string | null) => void
  setCounting: (v: boolean) => void
  setCountResult: (r: CountResponse | null, err?: string | null) => void
}

export const useSolveStore = create<SolveState>((set) => ({
  analyzing: false,
  analysis: null,
  analyzeError: null,
  solving: false,
  solveResult: null,
  solveError: null,
  counting: false,
  countResult: null,
  countError: null,

  setAnalyzing: (v) => set({ analyzing: v }),
  setAnalysis: (a, err = null) => set({ analysis: a, analyzeError: err, analyzing: false }),
  setSolving: (v) => set({ solving: v }),
  setSolveResult: (r, err = null) => set({ solveResult: r, solveError: err, solving: false }),
  setCounting: (v) => set({ counting: v }),
  setCountResult: (r, err = null) => set({ countResult: r, countError: err, counting: false }),
}))
