import { useState } from "react"
import * as api from "../lib/api"
import type { Mode } from "../lib/types"
import { dominoColor } from "../lib/domino"

export function CensusPage() {
  const [n, setN] = useState(2)
  const [mode, setMode] = useState<Mode>("stock")
  const [stock, setStock] = useState<number[]>([1, 1])
  const [sequence, setSequence] = useState<number[]>([1, 2])
  const [result, setResult] = useState<{ total: number; exact: boolean; statesVisited: number } | null>(null)
  const [loading, setLoading] = useState(false)

  const k = mode === "stock" ? stock.length : Math.max(1, ...sequence, 0)

  async function run() {
    setLoading(true)
    try {
      const res = await api.count(n, k, mode, mode === "stock" ? stock : undefined, mode === "sequence" ? sequence : undefined)
      setResult(res)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="mx-auto max-w-2xl p-8">
      <h2 className="font-display text-lg font-bold uppercase tracking-wide">Перепись конфигураций</h2>
      <p className="mt-2 text-sm text-(--color-muted)">
        Обобщения 3–4: сколько различных итоговых картин достижимо из данного запаса или последовательности.
      </p>

      <div className="mt-6 flex flex-wrap items-end gap-4 text-sm">
        <label className="flex flex-col gap-1">
          <span className="font-mono text-xs text-(--color-muted)">n</span>
          <input
            type="number"
            min={2}
            max={5}
            value={n}
            onChange={(e) => setN(parseInt(e.target.value || "2", 10))}
            className="w-16 rounded border border-(--color-rule) bg-(--color-raised) px-2 py-1"
          />
        </label>

        <div className="flex gap-3 font-mono text-xs">
          <label className="flex items-center gap-1.5">
            <input type="radio" checked={mode === "stock"} onChange={() => setMode("stock")} />
            запас
          </label>
          <label className="flex items-center gap-1.5">
            <input type="radio" checked={mode === "sequence"} onChange={() => setMode("sequence")} />
            последовательность
          </label>
        </div>

        {mode === "stock" ? (
          <label className="flex flex-col gap-1">
            <span className="font-mono text-xs text-(--color-muted)">запас (через запятую)</span>
            <input
              value={stock.join(",")}
              onChange={(e) => setStock(e.target.value.split(",").map((v) => parseInt(v.trim(), 10) || 0))}
              className="w-40 rounded border border-(--color-rule) bg-(--color-raised) px-2 py-1 font-mono"
            />
          </label>
        ) : (
          <label className="flex flex-col gap-1">
            <span className="font-mono text-xs text-(--color-muted)">последовательность</span>
            <input
              value={sequence.join(",")}
              onChange={(e) => setSequence(e.target.value.split(",").map((v) => parseInt(v.trim(), 10) || 0))}
              className="w-40 rounded border border-(--color-rule) bg-(--color-raised) px-2 py-1 font-mono"
            />
          </label>
        )}

        <button
          onClick={run}
          disabled={loading}
          className="rounded bg-(--color-ink) px-4 py-1.5 text-xs font-bold uppercase text-(--color-bg) disabled:opacity-50"
        >
          {loading ? "считаем…" : "посчитать"}
        </button>
      </div>

      {result && (
        <div className="mt-8 rounded border border-(--color-rule) bg-(--color-surface) p-6">
          <div className="font-display text-4xl font-extrabold tabular">{result.total.toLocaleString("ru-RU")}</div>
          <div className="mt-1 text-sm text-(--color-muted)">
            {result.exact ? "точное значение" : "нижняя оценка (лимит перебора достигнут)"} · состояний просмотрено{" "}
            {result.statesVisited.toLocaleString("ru-RU")}
          </div>
        </div>
      )}

      <div className="mt-10 text-xs text-(--color-muted)">
        Из §7 работы: поле 2×2, запас {"{1:1, 2:1}"} →{" "}
        <span className="font-mono text-(--color-ink)">28</span>; та же пара в порядке «1,2» →{" "}
        <span className="font-mono text-(--color-ink)">16</span>.
        <div className="mt-2 flex gap-1">
          <span className="h-3 w-3 rounded-sm" style={{ background: dominoColor(1) }} />
          <span className="h-3 w-3 rounded-sm" style={{ background: dominoColor(2) }} />
        </div>
      </div>
    </div>
  )
}
