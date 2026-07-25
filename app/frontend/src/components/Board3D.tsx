import { useMemo, useRef } from "react" 
import { Canvas, useThree } from "@react-three/fiber"
import { OrbitControls, OrthographicCamera } from "@react-three/drei"
import * as THREE from "three"
import { useBoardStore } from "../store/useBoardStore"
import { useUiStore, type CameraPreset } from "../store/useUiStore"
import { cellsOfMove, moveVisibility } from "../lib/geometry"

const DOMINO_HEX: Record<number, string> = {
  1: "#E0455A",
  2: "#E2762F",
  3: "#D9A518",
  4: "#35A85E",
  5: "#14A99B",
  6: "#2E9BD1",
  7: "#4A6BE8",
  8: "#9B4DDB",
  9: "#E0409B",
}

function cameraFor(preset: CameraPreset, n: number, height: number): [number, number, number] {
  const d = n * 2.2 + height
  switch (preset) {
    case "top":
      return [n / 2, d, n / 2 + 0.001]
    case "south":
      return [n / 2, height / 2 + 1, d]
    case "east":
      return [d, height / 2 + 1, n / 2]
    default:
      return [n * 0.9, d * 0.55, n * 1.3]
  }
}

interface SlabProps {
  x: number
  y: number
  z: number
  w: number
  d: number
  h: number
  color: string
  dim: boolean
  highlighted: boolean
  onHover: (v: boolean) => void
  onClick: () => void
}

function Slab({ x, y, z, w, d, h, color, dim, highlighted, onHover, onClick }: SlabProps) {
  return (
    <mesh
      position={[x, y, z]}
      onPointerOver={(e) => {
        e.stopPropagation()
        onHover(true)
      }}
      onPointerOut={(e) => {
        e.stopPropagation()
        onHover(false)
      }}
      onClick={(e) => {
        e.stopPropagation()
        onClick()
      }}
    >
      <boxGeometry args={[w * 0.92, h, d * 0.92]} />
      <meshStandardMaterial
        color={color}
        transparent={dim}
        opacity={dim ? 0.25 : 1}
        emissive={highlighted ? new THREE.Color(color) : undefined}
        emissiveIntensity={highlighted ? 0.4 : 0}
      />
    </mesh>
  )
}

function CameraRig({ n, height }: { n: number; height: number }) {
  const preset = useUiStore((s) => s.cameraPreset)
  const { camera } = useThree()
  const pos = cameraFor(preset, n, height)
  useMemo(() => {
    camera.position.set(...pos)
    camera.lookAt(n / 2, height / 2, n / 2)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [preset, n, height])
  return null
}

function Scene() {
  const n = useBoardStore((s) => s.n)
  const cells = useBoardStore((s) => s.cells)
  const moves = useBoardStore((s) => s.lastMoves)
  const explode = useUiStore((s) => s.explode)
  const ghost = useUiStore((s) => s.ghost)
  const hoveredStep = useUiStore((s) => s.hoveredStep)
  const setHoveredStep = useUiStore((s) => s.setHoveredStep)
  const playbackStep = useUiStore((s) => s.playbackStep)
  const setPlaybackStep = useUiStore((s) => s.setPlaybackStep)

  const gap = 0.55
  const visible = moves ? moveVisibility(moves, n) : []
  const stepCount = moves?.length ?? 0
  const height = stepCount * gap * explode

  const slabs = useMemo(() => {
    if (moves && moves.length > 0) {
      const visibleByStep = new Map(moves.map((m, i) => [m.step, visible[i]]))
      return moves
        .filter((m) => m.step <= playbackStep)
        .map((m) => {
          const [a] = cellsOfMove(m, n)
          const r0 = Math.floor(a / n)
          const c0 = a % n
          const w = m.orient === "h" ? 2 : 1
          const d = m.orient === "h" ? 1 : 2
          const x = c0 + w / 2
          const z = r0 + d / 2
          const y = (m.step - 0.5) * gap * explode
          return {
            key: `m${m.step}`,
            x,
            y,
            z,
            w,
            d,
            h: gap * 0.85,
            color: DOMINO_HEX[m.color] ?? "#888",
            step: m.step,
            dim: ghost ? false : !visibleByStep.get(m.step),
          }
        })
    }
    return cells
      .map((c, i) => ({ c, i }))
      .filter((v) => v.c !== 0)
      .map(({ c, i }) => {
        const r = Math.floor(i / n)
        const col = i % n
        return {
          key: `c${i}`,
          x: col + 0.5,
          y: gap * 0.4,
          z: r + 0.5,
          w: 1,
          d: 1,
          h: gap * 0.7,
          color: DOMINO_HEX[c] ?? "#888",
          step: 0,
          dim: false,
        }
      })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [moves, cells, n, explode, ghost, playbackStep])

  return (
    <group>
      <ambientLight intensity={0.65} />
      <directionalLight position={[n, n * 2, n]} intensity={0.9} />
      <mesh position={[n / 2, -0.02, n / 2]} rotation={[-Math.PI / 2, 0, 0]}>
        <planeGeometry args={[n + 1, n + 1]} />
        <meshBasicMaterial color="#1a2129" />
      </mesh>
      <gridHelper args={[n, n, "#2A3540", "#2A3540"]} position={[n / 2, 0, n / 2]} />
      {slabs.map((s) => (
        <Slab
          key={s.key}
          x={s.x}
          y={s.y}
          z={s.z}
          w={s.w}
          d={s.d}
          h={s.h}
          color={s.color}
          dim={s.dim}
          highlighted={hoveredStep === s.step && s.step !== 0}
          onHover={(v) => setHoveredStep(v ? s.step : null)}
          onClick={() => s.step > 0 && setPlaybackStep(s.step)}
        />
      ))}
      <CameraRig n={n} height={height} />
    </group>
  )
}

export function Board3D() {
  const n = useBoardStore((s) => s.n)
  const explode = useUiStore((s) => s.explode)
  const setExplode = useUiStore((s) => s.setExplode)
  const cameraPreset = useUiStore((s) => s.cameraPreset)
  const setCameraPreset = useUiStore((s) => s.setCameraPreset)
  const ghost = useUiStore((s) => s.ghost)
  const toggleGhost = useUiStore((s) => s.toggleGhost)
  const moves = useBoardStore((s) => s.lastMoves)
  const controlsRef = useRef(null)

  return (
    <div className="flex h-full flex-col gap-2">
      <div className="flex flex-wrap items-center gap-2 text-xs">
        {(["top", "south", "east", "iso"] as CameraPreset[]).map((p) => (
          <button
            key={p}
            onClick={() => setCameraPreset(p)}
            className={`rounded border px-2 py-0.5 font-mono uppercase ${
              cameraPreset === p ? "border-(--color-ink) text-(--color-ink)" : "border-(--color-rule) text-(--color-muted)"
            }`}
          >
            {{ top: "топ", south: "юг", east: "восток", iso: "изо" }[p]}
          </button>
        ))}
        <button
          onClick={toggleGhost}
          className={`rounded border px-2 py-0.5 font-mono ${
            ghost ? "border-(--color-ink) text-(--color-ink)" : "border-(--color-rule) text-(--color-muted)"
          }`}
        >
          призрак
        </button>
        <div className="ml-auto flex items-center gap-2">
          <span className="font-mono text-(--color-muted)">разнести</span>
          <input
            type="range"
            min={0}
            max={1}
            step={0.01}
            value={explode}
            disabled={!moves}
            onChange={(e) => setExplode(parseFloat(e.target.value))}
            className="w-24"
          />
        </div>
      </div>
      <div className="min-h-0 flex-1 overflow-hidden rounded border border-(--color-rule) bg-(--color-bg)">
        <Canvas shadows={false}>
          <OrthographicCamera makeDefault zoom={38} position={[n, n, n]} />
          <Scene />
          <OrbitControls ref={controlsRef} target={[n / 2, 0, n / 2]} enableDamping={false} />
        </Canvas>
      </div>
    </div>
  )
}
