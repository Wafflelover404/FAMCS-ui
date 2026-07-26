import { useEffect, useRef } from "react"
import { useBoardStore } from "../store/useBoardStore"
import { useSolveStore } from "../store/useSolveStore"
import * as api from "../lib/api"
import { dominoColor, reasonText } from "../lib/domino"

export function AnalysisPanel() {
  const n = useBoardStore((s) => s.n)
  const k = useBoardStore((s) => s.k)
  const cells = useBoardStore((s) => s.cells)
  const mode = useBoardStore((s) => s.mode)
  const stock = useBoardStore((s) => s.stock)

  const analysis = useSolveStore((s) => s.analysis)
  const analyzing = useSolveStore((s) => s.analyzing)
  const setAnalyzing = useSolveStore((s) => s.setAnalyzing)
  const setAnalysis = useSolveStore((s) => s.setAnalysis)
  const solveResult = useSolveStore((s) => s.solveResult)

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    setAnalyzing(true)
    debounceRef.current = setTimeout(async () => {
      try {
        const res = await api.analyze(n, k, cells, mode, stock)
        setAnalysis(res)
      } catch (e) {
        setAnalysis(null, e instanceof Error ? e.message : "ошибка")
      }
    }, 180)
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [n, k, cells.join(","), mode, stock.join(",")])

  const verdict = solveResult?.verdict

  return (
    <div className="flex h-full flex-col gap-4 overflow-y-auto border-l border-(--color-rule) bg-(--color-surface) p-4 text-sm">
      <h2 className="font-mono text-xs uppercase tracking-wide text-(--color-muted)">Разбор</h2>

      <div
        className={`rounded border px-3 py-2 font-display text-xs font-bold uppercase tracking-wide ${
          verdict === "feasible"
            ? "border-(--color-ok) text-(--color-ok)"
            : verdict === "infeasible"
              ? "border-(--color-fail) text-(--color-fail)"
              : verdict === "unproven"
                ? "border-(--color-warn) text-(--color-warn)"
                : analysis?.reachable
                  ? "border-(--color-rule) text-(--color-muted)"
                  : "border-(--color-fail) text-(--color-fail)"
        }`}
      >
        {verdict === "feasible" && "достижима"}
        {verdict === "infeasible" && "невозможно"}
        {verdict === "unproven" && "не доказано"}
        {!verdict && (analyzing ? "проверка…" : analysis?.reachable ? "достижима (предварительно)" : "недостижима")}
      </div>

      {(solveResult?.reason || (!solveResult && analysis?.reason)) && (
        <div className="rounded border border-(--color-rule) bg-(--color-raised) p-2.5 text-xs leading-relaxed">
          {reasonText[(solveResult?.reason ?? analysis?.reason)!.kind] ?? (solveResult?.reason ?? analysis?.reason)!.message}
        </div>
      )}

      <section>
        <table className="w-full font-mono text-xs">
          <thead>
            <tr className="text-left text-(--color-muted)">
              <th className="pb-1 font-normal">цвет</th>
              <th className="pb-1 font-normal">|S|</th>
              <th className="pb-1 font-normal">ν</th>
              <th className="pb-1 font-normal">need</th>
              {mode === "stock" && <th className="pb-1 font-normal">have</th>}
            </tr>
          </thead>
          <tbody>
            {(solveResult?.colors ?? analysis?.colors ?? []).map((cs) => (
              <tr key={cs.color} className={cs.deficit > 0 ? "text-(--color-fail)" : ""}>
                <td className="py-0.5">
                  <span
                    className="inline-block h-3 w-3 rounded-sm align-middle"
                    style={{ background: dominoColor(cs.color) }}
                  />{" "}
                  {cs.color}
                </td>
                <td className="tabular">{cs.cellCount}</td>
                <td className="tabular">{cs.matching}</td>
                <td className="tabular">{cs.need}</td>
                {mode === "stock" && <td className="tabular">{cs.have}</td>}
              </tr>
            ))}
          </tbody>
        </table>
        <p className="mt-2 text-[11px] text-(--color-muted)">ν — паросочетание (алгоритм Куна, O(V·E))</p>
      </section>

      {solveResult && (
        <section className="mt-auto">
          <h3 className="mb-1 font-mono text-xs uppercase tracking-wide text-(--color-muted)">Поиск</h3>
          <dl className="grid grid-cols-2 gap-x-2 gap-y-0.5 font-mono text-xs tabular">
            <dt className="text-(--color-muted)">узлов</dt>
            <dd className="text-right">{solveResult.stats.nodes.toLocaleString("ru-RU")}</dd>
            <dt className="text-(--color-muted)">глубина</dt>
            <dd className="text-right">{solveResult.stats.peakDepth}</dd>
            <dt className="text-(--color-muted)">memo hit</dt>
            <dd className="text-right">
              {solveResult.stats.memoHits > 0
                ? `${solveResult.stats.memoHits.toLocaleString("ru-RU")}`
                : "0"}
            </dd>
            <dt className="text-(--color-muted)">время</dt>
            <dd className="text-right">{(solveResult.stats.elapsedNs / 1e6).toFixed(1)} мс</dd>
          </dl>
        </section>
      )}
    </div>
  )
}
