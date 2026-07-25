import type { Move } from "./types" 

export function cellsOfMove(m: Move, n: number): [number, number] {
  const u = m.r * n + m.c
  return m.orient === "h" ? [u, u + 1] : [u, u + n]
}

export function moveVisibility(moves: Move[], n: number): boolean[] {
  const winner = new Array(n * n).fill(0)
  for (const m of moves) {
    const [a, b] = cellsOfMove(m, n)
    winner[a] = m.step
    winner[b] = m.step
  }
  return moves.map((m) => {
    const [a, b] = cellsOfMove(m, n)
    return winner[a] === m.step || winner[b] === m.step
  })
}

export function frameAt(moves: Move[], n: number, step: number): number[] {
  const cells = new Array(n * n).fill(0)
  for (const m of moves) {
    if (m.step > step) break
    const [a, b] = cellsOfMove(m, n)
    cells[a] = m.color
    cells[b] = m.color
  }
  return cells
}

export function coverageHistory(moves: Move[], n: number) {
  const steps: number[][] = Array.from({ length: n * n }, () => [])
  const winnerStep = new Array(n * n).fill(0)
  const winnerColor = new Array(n * n).fill(0)
  for (const m of moves) {
    const [a, b] = cellsOfMove(m, n)
    for (const cell of [a, b]) {
      steps[cell].push(m.step)
      winnerStep[cell] = m.step
      winnerColor[cell] = m.color
    }
  }
  return { steps, winnerStep, winnerColor }
}
