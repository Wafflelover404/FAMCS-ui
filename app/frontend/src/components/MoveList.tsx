import { useBoardStore } from "../store/useBoardStore" 
import { useUiStore } from "../store/useUiStore"
import { dominoColor } from "../lib/domino"
import { moveVisibility } from "../lib/geometry"
import type { Move } from "../lib/types"

function exportJSON(moves: Move[]): string {
  return JSON.stringify(moves, null, 2)
}

function exportPlain(moves: Move[]): string {
  return moves.map((m) => `${String(m.step).padStart(2, "0")} ${m.color} (${m.r},${m.c}) ${m.orient}`).join("\n")
}

function exportTikz(moves: Move[], n: number): string {
  const lines = [`\\begin{tikzpicture}[x=0.6cm,y=-0.6cm]`, `\\draw[gray!40] (0,0) grid (${n},${n});`]
  for (const m of moves) {
    const x0 = m.c
    const y0 = m.r
    const x1 = m.orient === "h" ? m.c + 2 : m.c + 1
    const y1 = m.orient === "h" ? m.r + 1 : m.r + 2
    lines.push(`\\fill[domino${m.color}] (${x0},${y0}) rectangle (${x1},${y1});`)
    lines.push(`\\node at (${(x0 + x1) / 2},${(y0 + y1) / 2}) {\\small ${m.step}};`)
  }
  lines.push(`\\end{tikzpicture}`)
  return lines.join("\n")
}

function download(filename: string, content: string) {
  const blob = new Blob([content], { type: "text/plain;charset=utf-8" })
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function MoveList() {
  const moves = useBoardStore((s) => s.lastMoves)
  const n = useBoardStore((s) => s.n)
  const hoveredStep = useUiStore((s) => s.hoveredStep)
  const setHoveredStep = useUiStore((s) => s.setHoveredStep)
  const playbackStep = useUiStore((s) => s.playbackStep)

  if (!moves || moves.length === 0) {
    return (
      <div className="p-3 text-xs text-(--color-muted)">
        Решите головоломку, чтобы увидеть последовательность ходов.
      </div>
    )
  }

  const visible = moveVisibility(moves, n)

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between border-b border-(--color-rule) px-3 py-2">
        <h3 className="font-mono text-xs uppercase tracking-wide text-(--color-muted)">Ходы</h3>
        <div className="flex gap-2 text-xs">
          <button onClick={() => download("moves.json", exportJSON(moves))} className="text-(--color-muted) hover:text-(--color-ink)">
            JSON
          </button>
          <button onClick={() => download("moves.txt", exportPlain(moves))} className="text-(--color-muted) hover:text-(--color-ink)">
            TXT
          </button>
          <button onClick={() => download("moves.tex", exportTikz(moves, n))} className="text-(--color-muted) hover:text-(--color-ink)">
            TikZ
          </button>
        </div>
      </div>
      <ol className="flex-1 overflow-y-auto font-mono text-xs">
        {moves.map((m, i) => (
          <li
            key={m.step}
            onMouseEnter={() => setHoveredStep(m.step)}
            onMouseLeave={() => setHoveredStep(null)}
            className={`flex items-center gap-2 border-b border-(--color-rule)/40 px-3 py-1.5 ${
              hoveredStep === m.step ? "bg-(--color-raised)" : ""
            } ${m.step === playbackStep ? "border-l-2 border-l-(--color-ink)" : "border-l-2 border-l-transparent"}`}
          >
            <span className="w-6 text-(--color-muted)">{String(m.step).padStart(2, "0")}</span>
            <span className="h-3 w-3 flex-none rounded-sm" style={{ background: dominoColor(m.color) }} />
            <span>
              ({m.r},{m.c}) {m.orient === "h" ? "гор." : "верт."}
            </span>
            <span className={`ml-auto text-[10px] ${visible[i] ? "text-(--color-ok)" : "text-(--color-muted)"}`}>
              {visible[i] ? "● виден" : "○ скрыт"}
            </span>
          </li>
        ))}
      </ol>
    </div>
  )
}
