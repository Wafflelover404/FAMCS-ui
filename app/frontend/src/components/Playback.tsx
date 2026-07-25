import { useEffect, useRef } from "react"
import { useBoardStore } from "../store/useBoardStore"
import { useUiStore } from "../store/useUiStore"

export function Playback() {
  const moves = useBoardStore((s) => s.lastMoves)
  const step = useUiStore((s) => s.playbackStep)
  const setStep = useUiStore((s) => s.setPlaybackStep)
  const playing = useUiStore((s) => s.playing)
  const setPlaying = useUiStore((s) => s.setPlaying)
  const speed = useUiStore((s) => s.speed)
  const setSpeed = useUiStore((s) => s.setSpeed)

  const total = moves?.length ?? 0
  const rafRef = useRef<number | null>(null)
  const lastRef = useRef(0)

  useEffect(() => {
    if (!playing || total === 0) return
    lastRef.current = performance.now()
    function tick(t: number) {
      const dt = t - lastRef.current
      const stepMs = 500 / speed
      if (dt >= stepMs) {
        lastRef.current = t
        setStep(Math.min(total, useUiStore.getState().playbackStep + 1))
        if (useUiStore.getState().playbackStep >= total) {
          setPlaying(false)
          return
        }
      }
      rafRef.current = requestAnimationFrame(tick)
    }
    rafRef.current = requestAnimationFrame(tick)
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current)
    }
  }, [playing, speed, total, setStep, setPlaying])

  if (!moves || moves.length === 0) return null

  return (
    <div className="flex items-center gap-3 border-t border-(--color-rule) bg-(--color-surface) px-4 py-2">
      <div className="flex items-center gap-1">
        <button onClick={() => setStep(0)} className="text-(--color-muted) hover:text-(--color-ink)" aria-label="В начало">
          ◀◀
        </button>
        <button
          onClick={() => setStep(Math.max(0, step - 1))}
          className="text-(--color-muted) hover:text-(--color-ink)"
          aria-label="Шаг назад"
        >
          ◀
        </button>
        <button
          onClick={() => setPlaying(!playing)}
          className="w-6 text-(--color-ink)"
          aria-label={playing ? "Пауза" : "Играть"}
        >
          {playing ? "❚❚" : "▶"}
        </button>
        <button
          onClick={() => setStep(Math.min(total, step + 1))}
          className="text-(--color-muted) hover:text-(--color-ink)"
          aria-label="Шаг вперёд"
        >
          ▶
        </button>
        <button onClick={() => setStep(total)} className="text-(--color-muted) hover:text-(--color-ink)" aria-label="В конец">
          ▶▶
        </button>
      </div>
      <input
        type="range"
        min={0}
        max={total}
        value={step}
        onChange={(e) => setStep(parseInt(e.target.value, 10))}
        className="flex-1"
      />
      <span className="font-mono text-xs tabular text-(--color-muted)">
        шаг {String(step).padStart(2, "0")} / {total}
      </span>
      <select
        value={speed}
        onChange={(e) => setSpeed(parseFloat(e.target.value))}
        className="rounded border border-(--color-rule) bg-(--color-raised) px-1 py-0.5 font-mono text-xs text-(--color-ink)"
      >
        <option value={0.5}>0.5×</option>
        <option value={1}>1×</option>
        <option value={2}>2×</option>
        <option value={4}>4×</option>
      </select>
    </div>
  )
}
