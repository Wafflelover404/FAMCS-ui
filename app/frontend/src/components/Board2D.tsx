import { useRef, useState } from "react"
import { useBoardStore } from "../store/useBoardStore"
import { useSolveStore } from "../store/useSolveStore"
import { useUiStore } from "../store/useUiStore"
import { dominoColor } from "../lib/domino"
import { coverageHistory, frameAt } from "../lib/geometry"

export function Board2D() {
  const n = useBoardStore((s) => s.n)
  const storeCells = useBoardStore((s) => s.cells)
  const paintColor = useBoardStore((s) => s.paintColor)
  const paintCells = useBoardStore((s) => s.paintCells)
  const lastMoves = useBoardStore((s) => s.lastMoves)
  const playbackStep = useUiStore((s) => s.playbackStep)
  const cells =
    lastMoves && lastMoves.length > 0 && playbackStep < lastMoves.length
      ? frameAt(lastMoves, n, playbackStep)
      : storeCells

  const analysis = useSolveStore((s) => s.analysis)
  const residue = new Set(analysis?.residue ?? [])
  const parity = analysis?.parity

  const hoveredStep = useUiStore((s) => s.hoveredStep)
  const setHoveredStep = useUiStore((s) => s.setHoveredStep)
  const selectedCell = useUiStore((s) => s.selectedCell)
  const setSelectedCell = useUiStore((s) => s.setSelectedCell)
  const [showParity, setShowParity] = useState(false)

  const dragRef = useRef<{ color: number; touched: Set<number> } | null>(null)
  const gridRef = useRef<HTMLDivElement>(null)

  const cellStepMap = lastMoves ? coverageHistory(lastMoves, n).winnerStep : null
  const hoverCells = new Set<number>()
  if (hoveredStep != null && lastMoves) {
    const mv = lastMoves.find((m) => m.step === hoveredStep)
    if (mv) {
      const u = mv.r * n + mv.c
      hoverCells.add(u)
      hoverCells.add(mv.orient === "h" ? u + 1 : u + n)
    }
  }

  function commit() {
    if (!dragRef.current) return
    paintCells([...dragRef.current.touched], dragRef.current.color)
    dragRef.current = null
  }

  function paintAt(i: number, color: number) {
    if (!dragRef.current) dragRef.current = { color, touched: new Set() }
    dragRef.current.touched.add(i)
    const grid = gridRef.current
    if (grid) {
      const el = grid.children[i] as HTMLElement | undefined
      if (el) el.style.background = color === 0 ? "" : dominoColor(color)
    }
  }

  function focusCell(i: number) {
    const el = gridRef.current?.children[i] as HTMLElement | undefined
    el?.focus()
  }

  function onKeyDown(e: React.KeyboardEvent, i: number) {
    const r = Math.floor(i / n)
    const c = i % n
    if (e.key === "ArrowRight" && c + 1 < n) focusCell(i + 1)
    else if (e.key === "ArrowLeft" && c > 0) focusCell(i - 1)
    else if (e.key === "ArrowDown" && r + 1 < n) focusCell(i + n)
    else if (e.key === "ArrowUp" && r > 0) focusCell(i - n)
    else if (/^[0-9]$/.test(e.key)) {
      const v = parseInt(e.key, 10)
      paintCells([i], v)
    }
  }

  return (
    <div className="flex h-full min-h-0 flex-col gap-2">
      <div className="flex items-center justify-between text-xs text-(--color-muted)">
        <span className="font-mono">
          {n}×{n}
        </span>
        <label className="flex items-center gap-1.5">
          <input type="checkbox" checked={showParity} onChange={(e) => setShowParity(e.target.checked)} />
          шахматная раскраска
        </label>
      </div>
      <div className="flex min-h-0 flex-1 items-center justify-center overflow-hidden">
        <div
          ref={gridRef}
          role="grid"
          aria-label="Игровое поле"
          className="grid touch-none select-none gap-[2px] rounded bg-(--color-rule) p-[2px]"
          style={{ gridTemplateColumns: `repeat(${n}, minmax(0,1fr))`, aspectRatio: "1/1", height: "100%", width: "auto", maxWidth: "100%" }}
          onPointerUp={commit}
          onPointerLeave={commit}
        >
          {cells.map((color, i) => {
            const isResidue = residue.has(i)
            const isHovered = hoverCells.has(i)
            const isSelected = selectedCell === i
            const step = cellStepMap ? cellStepMap[i] : 0
            return (
              <button
                key={i}
                tabIndex={0}
                role="gridcell"
                aria-label={`Клетка ${Math.floor(i / n) + 1},${(i % n) + 1}, цвет ${color}`}
                className={`relative flex items-center justify-center font-mono text-[10px] font-medium outline-none transition-[filter] duration-150 ${
                  isSelected ? "ring-2 ring-inset ring-(--color-ink)" : ""
                } ${isHovered ? "ring-2 ring-inset ring-white" : ""} ${isResidue ? "animate-pulse" : ""}`}
                style={{
                  background: color === 0 ? "var(--color-surface)" : dominoColor(color),
                  color: color === 0 ? "var(--color-muted)" : "rgba(0,0,0,0.72)",
                  backgroundImage:
                    showParity && parity
                      ? `linear-gradient(${parity[i] === 0 ? "rgba(255,255,255,0.08)" : "rgba(0,0,0,0.08)"} 0 0)`
                      : undefined,
                  outline: isResidue ? "2px solid var(--color-fail)" : undefined,
                }}
                onPointerDown={(e) => {
                  e.preventDefault()
                  const erase = e.button === 2 || e.shiftKey
                  paintAt(i, erase ? 0 : paintColor)
                }}
                onPointerEnter={() => {
                  if (dragRef.current) paintAt(i, dragRef.current.color)
                }}
                onContextMenu={(e) => e.preventDefault()}
                onClick={() => setSelectedCell(i)}
                onMouseEnter={() => {
                  if (step) setHoveredStep(step)
                }}
                onMouseLeave={() => {
                  if (step) setHoveredStep(null)
                }}
                onKeyDown={(e) => onKeyDown(e, i)}
              >
                {color !== 0 ? color : ""}
              </button>
            )
          })}
        </div>
      </div>
      {selectedCell != null && (
        <div className="font-mono text-[11px] text-(--color-muted)">
          клетка ({Math.floor(selectedCell / n)},{selectedCell % n}){" "}
          {lastMoves
            ? (() => {
                const hist = coverageHistory(lastMoves, n)
                const steps = hist.steps[selectedCell]
                if (steps.length === 0) return "— не накрыта ни разу"
                return `— накрыта ${steps.length} раз: шаги ${steps.join(", ")} · цвет от шага ${hist.winnerStep[selectedCell]}`
              })()
            : `— цвет ${cells[selectedCell]}`}
        </div>
      )}
    </div>
  )
}
