import { create } from "zustand"

export type ViewMode = "2d" | "3d"
export type CameraPreset = "top" | "south" | "east" | "iso"

interface UiState {
  theme: "dark" | "light"
  toggleTheme: () => void

  viewMode: ViewMode
  setViewMode: (v: ViewMode) => void

  explode: number
  setExplode: (v: number) => void

  cameraPreset: CameraPreset
  setCameraPreset: (p: CameraPreset) => void

  ghost: boolean
  toggleGhost: () => void

  hoveredStep: number | null
  setHoveredStep: (s: number | null) => void

  selectedCell: number | null
  setSelectedCell: (c: number | null) => void

  playbackStep: number
  setPlaybackStep: (s: number) => void
  playing: boolean
  setPlaying: (p: boolean) => void
  speed: number
  setSpeed: (s: number) => void
}

export const useUiStore = create<UiState>((set) => ({
  theme: "dark",
  toggleTheme: () =>
    set((s) => {
      const next = s.theme === "dark" ? "light" : "dark"
      document.documentElement.setAttribute("data-theme", next)
      return { theme: next }
    }),

  viewMode: "3d",
  setViewMode: (viewMode) => set({ viewMode }),

  explode: 0.6,
  setExplode: (explode) => set({ explode }),

  cameraPreset: "iso",
  setCameraPreset: (cameraPreset) => set({ cameraPreset }),

  ghost: false,
  toggleGhost: () => set((s) => ({ ghost: !s.ghost })),

  hoveredStep: null,
  setHoveredStep: (hoveredStep) => set({ hoveredStep }),

  selectedCell: null,
  setSelectedCell: (selectedCell) => set({ selectedCell }),

  playbackStep: 0,
  setPlaybackStep: (playbackStep) => set({ playbackStep }),
  playing: false,
  setPlaying: (playing) => set({ playing }),
  speed: 1,
  setSpeed: (speed) => set({ speed }),
}))
