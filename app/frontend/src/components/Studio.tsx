import { ParamsPanel } from "./ParamsPanel" 
import { Board2D } from "./Board2D"
import { Board3D } from "./Board3D"
import { AnalysisPanel } from "./AnalysisPanel"
import { MoveList } from "./MoveList"
import { Playback } from "./Playback"
import { useUiStore } from "../store/useUiStore"

export function Studio() {
  const viewMode = useUiStore((s) => s.viewMode)
  const setViewMode = useUiStore((s) => s.setViewMode)

  return (
    <div className="grid min-h-0 flex-1 grid-cols-[280px_1fr_300px]">
      <ParamsPanel />

      <div className="flex min-h-0 flex-col">
        <div className="flex min-h-0 flex-1 flex-col gap-3 p-4">
          <div className="flex items-center gap-2">
            <div className="flex overflow-hidden rounded border border-(--color-rule) text-xs font-mono">
              <button
                onClick={() => setViewMode("2d")}
                className={`px-3 py-1 ${viewMode === "2d" ? "bg-(--color-raised) text-(--color-ink)" : "text-(--color-muted)"}`}
              >
                2D
              </button>
              <button
                onClick={() => setViewMode("3d")}
                className={`px-3 py-1 ${viewMode === "3d" ? "bg-(--color-raised) text-(--color-ink)" : "text-(--color-muted)"}`}
              >
                3D
              </button>
            </div>
          </div>
          <div className="min-h-0 flex-1">
            {viewMode === "2d" ? (
              <div className="mx-auto h-full max-w-xl">
                <Board2D />
              </div>
            ) : (
              <Board3D />
            )}
          </div>
        </div>
        <Playback />
        <div className="h-56 border-t border-(--color-rule)">
          <MoveList />
        </div>
      </div>

      <AnalysisPanel />
    </div>
  )
}
