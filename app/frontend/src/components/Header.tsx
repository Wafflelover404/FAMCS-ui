import { Link, useLocation } from "react-router-dom"
import { useEffect, useState } from "react"
import { useUiStore } from "../store/useUiStore"

export function Header() {
  const theme = useUiStore((s) => s.theme)
  const toggleTheme = useUiStore((s) => s.toggleTheme)
  const location = useLocation()
  const [online, setOnline] = useState<boolean | null>(null)

  useEffect(() => {
    let alive = true
    fetch("/healthz")
      .then((r) => alive && setOnline(r.ok))
      .catch(() => alive && setOnline(false))
    return () => {
      alive = false
    }
  }, [])

  return (
    <header className="flex items-center justify-between border-b border-(--color-rule) bg-(--color-surface) px-5 py-3">
      <div className="flex items-baseline gap-3">
        <h1 className="font-display text-sm font-extrabold uppercase tracking-wide text-(--color-ink)">
          Перекрывающиеся прямоугольники
        </h1>
        <span className="font-mono text-xs text-(--color-muted)">Задача 2</span>
      </div>
      <nav className="flex items-center gap-4 text-sm">
        <Link
          to="/"
          className={location.pathname === "/" ? "text-(--color-ink)" : "text-(--color-muted) hover:text-(--color-ink)"}
        >
          Студия
        </Link>
        <Link
          to="/census"
          className={
            location.pathname === "/census" ? "text-(--color-ink)" : "text-(--color-muted) hover:text-(--color-ink)"
          }
        >
          Перепись
        </Link>
        <button
          onClick={toggleTheme}
          className="rounded border border-(--color-rule) px-2 py-1 text-xs text-(--color-muted) hover:text-(--color-ink)"
          aria-label="Переключить тему"
        >
          {theme === "dark" ? "◐ тёмная" : "◑ светлая"}
        </button>
        <span className="flex items-center gap-1.5 font-mono text-xs text-(--color-muted)">
          <span
            className={`h-1.5 w-1.5 rounded-full ${
              online === null ? "bg-(--color-muted)" : online ? "bg-(--color-ok)" : "bg-(--color-fail)"
            }`}
          />
          {online === null ? "проверка" : online ? "подключено" : "нет связи"}
        </span>
      </nav>
    </header>
  )
}
