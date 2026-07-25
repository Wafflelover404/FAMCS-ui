import type { 
  AnalyzeResponse,
  CountResponse,
  Mode,
  Preset,
  RandomResponse,
  SimulateResponse,
  SolveResponse,
} from "./types"

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }
  return res.json() as Promise<T>
}

export function analyze(
  n: number,
  k: number,
  cells: number[],
  mode: Mode,
  stock?: number[],
): Promise<AnalyzeResponse> {
  return post("/api/v1/analyze", {
    n,
    k,
    cells,
    stock: mode === "stock" ? stock : undefined,
  })
}

export function solve(
  n: number,
  k: number,
  cells: number[],
  mode: Mode,
  stock: number[] | undefined,
  sequence: number[] | undefined,
  trace = true,
): Promise<SolveResponse> {
  return post("/api/v1/solve", { n, k, cells, mode, stock, sequence, trace })
}

export function count(
  n: number,
  k: number,
  mode: Mode,
  stock: number[] | undefined,
  sequence: number[] | undefined,
  guard?: number,
): Promise<CountResponse> {
  return post("/api/v1/count", { n, k, mode, stock, sequence, guard })
}

export function simulate(n: number, moves: unknown[]): Promise<SimulateResponse> {
  return post("/api/v1/simulate", { n, moves })
}

export function randomBoard(n: number, m: number, k: number, seed?: number): Promise<RandomResponse> {
  return post("/api/v1/random", { n, m, k, seed })
}

export async function presets(): Promise<Preset[]> {
  const res = await fetch("/api/v1/presets")
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}
