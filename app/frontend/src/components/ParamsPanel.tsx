import { useEffect, useState } from "react"
import { useBoardStore } from "../store/useBoardStore"
import { useSolveStore } from "../store/useSolveStore"
import * as api from "../lib/api"
import { dominoColor, dominoNames, MAX_COLOR } from "../lib/domino"
import { useUiStore } from "../store/useUiStore"
import type { Preset } from "../lib/types"

function formatGrid(n: number, cells: number[]): string {
  const rows: string[] = []
  for (let r = 0; r < n; r++) rows.push(cells.slice(r * n, r * n + n).join(" "))
  return rows.join("\n")
}

function parseGrid(text: string): number[] | null {
  const nums = text
    .trim()
    .split(/\s+/)
    .filter((t) => t.length > 0)
    .map((t) => parseInt(t, 10))
  if (nums.some((v) => Number.isNaN(v))) return null
  return nums
}

export function ParamsPanel() {
  const n = useBoardStore((s) => s.n)
  const k = useBoardStore((s) => s.k)
  const cells = useBoardStore((s) => s.cells)
  const mode = useBoardStore((s) => s.mode)
  const stock = useBoardStore((s) => s.stock)
  const sequence = useBoardStore((s) => s.sequence)
  const paintColor = useBoardStore((s) => s.paintColor)
  const setN = useBoardStore((s) => s.setN)
  const setK = useBoardStore((s) => s.setK)
  const setMode = useBoardStore((s) => s.setMode)
  const setPaintColor = useBoardStore((s) => s.setPaintColor)
  const setStockAt = useBoardStore((s) => s.setStockAt)
  const addSequenceStep = useBoardStore((s) => s.addSequenceStep)
  const removeSequenceStep = useBoardStore((s) => s.removeSequenceStep)
  const moveSequenceStep = useBoardStore((s) => s.moveSequenceStep)
  const loadPreset = useBoardStore((s) => s.loadPreset)
  const loadCells = useBoardStore((s) => s.loadCells)
  const clearBoard = useBoardStore((s) => s.clearBoard)
  const undo = useBoardStore((s) => s.undo)
  const redo = useBoardStore((s) => s.redo)
  const setLastMoves = useBoardStore((s) => s.setLastMoves)

  const solving = useSolveStore((s) => s.solving)
  const setSolving = useSolveStore((s) => s.setSolving)
  const setSolveResult = useSolveStore((s) => s.setSolveResult)

  const [presetList, setPresetList] = useState<Preset[]>([])
  const [importText, setImportText] = useState("")
  const [showImport, setShowImport] = useState(false)

  useEffect(() => {
    api.presets().then((list) => {
      setPresetList(list)
      const initial = list.find((p) => p.slug === "statement") ?? list[0]
      if (initial) loadPreset(initial)
    }).catch(() => {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function handleRandom() {
    const m = mode === "sequence" ? Math.max(2, sequence.length) : Math.max(2, stock.reduce((a, b) => a + b, 0)) || 6
    const res = await api.randomBoard(n, m, k)
    loadCells(res.n, res.k, res.cells)
    setLastMoves(null)
  }

  function handleImport() {
    const nums = parseGrid(importText)
    if (!nums || nums.length !== n * n) return
    loadCells(n, k, nums)
    setShowImport(false)
  }

  async function handleSolve() {
    setSolving(true)
    try {
      const res = await api.solve(n, k, cells, mode, mode === "stock" ? stock : undefined, mode === "sequence" ? sequence : undefined)
      setSolveResult(res)
      setLastMoves(res.moves ?? null)
      useUiStore.getState().setPlaybackStep(res.moves?.length ?? 0)
    } catch (e) {
      setSolveResult(null, e instanceof Error ? e.message : "ошибка")
    }
  }

  const colors = Array.from({ length: k }, (_, i) => i + 1)

  return (
    <div className="flex h-full flex-col gap-5 overflow-y-auto border-r border-(--color-rule) bg-(--color-surface) p-4 text-sm">
      <section>
        <h2 className="mb-2 font-mono text-xs uppercase tracking-wide text-(--color-muted)">Параметры</h2>
        <label className="mb-2 flex items-center justify-between gap-2">
          <span>n</span>
          <input
            type="range"
            min={2}
            max={10}
            value={n}
            onChange={(e) => setN(parseInt(e.target.value, 10))}
            className="flex-1"
          />
          <span className="w-6 text-right font-mono tabular">{n}</span>
        </label>
        <label className="mb-2 flex items-center justify-between gap-2">
          <span>k</span>
          <input
            type="range"
            min={1}
            max={MAX_COLOR}
            value={k}
            onChange={(e) => setK(parseInt(e.target.value, 10))}
            className="flex-1"
          />
          <span className="w-6 text-right font-mono tabular">{k}</span>
        </label>

        <div className="mt-3 flex gap-3 font-mono text-xs">
          <label className="flex items-center gap-1.5">
            <input type="radio" checked={mode === "stock"} onChange={() => setMode("stock")} />
            Режим 1 (запас)
          </label>
          <label className="flex items-center gap-1.5">
            <input type="radio" checked={mode === "sequence"} onChange={() => setMode("sequence")} />
            Режим 2 (порядок)
          </label>
        </div>
      </section>

      <section>
        <h2 className="mb-2 font-mono text-xs uppercase tracking-wide text-(--color-muted)">Палитра</h2>
        <div className="flex flex-wrap gap-1.5">
          {colors.map((c) => (
            <button
              key={c}
              onClick={() => setPaintColor(c)}
              aria-label={`${dominoNames[c]} (${c})`}
              className={`flex h-7 w-7 items-center justify-center rounded font-mono text-[11px] font-semibold text-black/70 ${
                paintColor === c ? "ring-2 ring-(--color-ink)" : ""
              }`}
              style={{ background: dominoColor(c) }}
            >
              {c}
            </button>
          ))}
          <button
            onClick={() => setPaintColor(0)}
            className={`flex h-7 w-7 items-center justify-center rounded border border-(--color-rule) font-mono text-[11px] ${
              paintColor === 0 ? "ring-2 ring-(--color-ink)" : ""
            }`}
            aria-label="Стереть"
          >
            ×
          </button>
        </div>
      </section>

      {mode === "stock" ? (
        <section>
          <h2 className="mb-2 font-mono text-xs uppercase tracking-wide text-(--color-muted)">Запас</h2>
          <div className="grid grid-cols-2 gap-1.5">
            {colors.map((c) => (
              <label key={c} className="flex items-center gap-1.5 font-mono text-xs">
                <span className="h-3 w-3 rounded-sm" style={{ background: dominoColor(c) }} />
                <input
                  type="number"
                  min={0}
                  max={10}
                  value={stock[c - 1] ?? 0}
                  onChange={(e) => setStockAt(c, parseInt(e.target.value || "0", 10))}
                  className="w-12 rounded border border-(--color-rule) bg-(--color-raised) px-1 py-0.5 text-(--color-ink)"
                />
              </label>
            ))}
          </div>
        </section>
      ) : (
        <section>
          <h2 className="mb-2 font-mono text-xs uppercase tracking-wide text-(--color-muted)">Последовательность</h2>
          <div className="flex flex-wrap gap-1.5">
            {sequence.map((c, i) => (
              <span
                key={i}
                className="group relative flex h-7 w-7 items-center justify-center rounded font-mono text-[11px] font-semibold text-black/70"
                style={{ background: dominoColor(c) }}
              >
                {c}
                <button
                  onClick={() => removeSequenceStep(i)}
                  className="absolute -right-1 -top-1 hidden h-3.5 w-3.5 rounded-full bg-(--color-fail) text-[9px] leading-none text-white group-hover:block"
                  aria-label="Удалить шаг"
                >
                  ×
                </button>
                <span className="absolute -bottom-3 left-0 right-0 flex justify-center gap-0.5">
                  {i > 0 && (
                    <button onClick={() => moveSequenceStep(i, i - 1)} className="text-[8px] text-(--color-muted)">
                      ◀
                    </button>
                  )}
                  {i < sequence.length - 1 && (
                    <button onClick={() => moveSequenceStep(i, i + 1)} className="text-[8px] text-(--color-muted)">
                      ▶
                    </button>
                  )}
                </span>
              </span>
            ))}
            <button
              onClick={() => addSequenceStep(paintColor || 1)}
              className="flex h-7 w-7 items-center justify-center rounded border border-dashed border-(--color-rule) text-(--color-muted)"
              aria-label="Добавить шаг"
            >
              +
            </button>
          </div>
        </section>
      )}

      <section>
        <h2 className="mb-2 font-mono text-xs uppercase tracking-wide text-(--color-muted)">Примеры</h2>
        <div className="flex flex-col gap-1">
          {presetList.map((p) => (
            <button
              key={p.slug}
              onClick={() => loadPreset(p)}
              className="rounded border border-(--color-rule) px-2 py-1 text-left text-xs hover:bg-(--color-raised)"
            >
              {p.title}
            </button>
          ))}
          <button
            onClick={handleRandom}
            className="rounded border border-(--color-rule) px-2 py-1 text-left text-xs hover:bg-(--color-raised)"
          >
            Случайная партия
          </button>
          <button
            onClick={() => setShowImport((v) => !v)}
            className="rounded border border-(--color-rule) px-2 py-1 text-left text-xs hover:bg-(--color-raised)"
          >
            Импорт / экспорт
          </button>
          {showImport && (
            <div className="flex flex-col gap-1.5">
              <textarea
                value={importText || formatGrid(n, cells)}
                onChange={(e) => setImportText(e.target.value)}
                rows={n}
                className="w-full rounded border border-(--color-rule) bg-(--color-raised) p-1.5 font-mono text-[11px] text-(--color-ink)"
              />
              <button onClick={handleImport} className="rounded bg-(--color-ink) px-2 py-1 text-xs text-(--color-bg)">
                загрузить
              </button>
            </div>
          )}
          <button onClick={clearBoard} className="rounded border border-(--color-rule) px-2 py-1 text-left text-xs hover:bg-(--color-raised)">
            очистить поле
          </button>
          <div className="flex gap-1.5">
            <button onClick={undo} className="flex-1 rounded border border-(--color-rule) px-2 py-1 text-xs hover:bg-(--color-raised)">
              ↶ отменить
            </button>
            <button onClick={redo} className="flex-1 rounded border border-(--color-rule) px-2 py-1 text-xs hover:bg-(--color-raised)">
              ↷ вернуть
            </button>
          </div>
        </div>
      </section>

      <button
        onClick={handleSolve}
        disabled={solving}
        className="mt-auto rounded bg-(--color-ink) py-2.5 font-display text-xs font-bold uppercase tracking-wide text-(--color-bg) disabled:opacity-50"
      >
        {solving ? "решаем…" : "решить"}
      </button>
    </div>
  )
}
