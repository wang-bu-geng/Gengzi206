<template>
  <!-- 全局 Three.js 背景：信号红线框多面体 + 粒子星空
       挂在 App 外壳中，路由切换不销毁；移动端 / 降级环境不启动 WebGL -->
  <div ref="containerRef" class="global-bg" aria-hidden="true" />
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import * as THREE from 'three'

const containerRef = ref<HTMLDivElement | null>(null)

let renderer: THREE.WebGLRenderer | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null
let frameId = 0
let disposed = false

// 可动对象集合，便于主题切换与卸载
let coreGroup: THREE.Group | null = null
let outerWire: THREE.LineSegments | null = null
let innerWire: THREE.LineSegments | null = null
let vertexPoints: THREE.Points | null = null
let starField: THREE.Points | null = null
let starGeometry: THREE.BufferGeometry | null = null
let starKinds: Uint8Array | null = null // 0=冷白尘 1=红点 2=亮白锚点
let outerMaterial: THREE.LineBasicMaterial | null = null
let innerMaterial: THREE.LineBasicMaterial | null = null
let vertexMaterial: THREE.PointsMaterial | null = null
let starMaterial: THREE.PointsMaterial | null = null

const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)')
const isMobile = window.matchMedia('(max-width: 768px)')

// 鼠标视差（-1 ~ 1）
const pointer = { x: 0, y: 0 }
const smooth = { x: 0, y: 0 }

const isDark = () => document.documentElement.classList.contains('dark')

/** 依据主题刷新线框与粒子配色 */
function applyThemeColors() {
  const dark = isDark()
  if (outerMaterial) {
    outerMaterial.color.set(dark ? 0xff2d2d : 0xd40000)
    outerMaterial.opacity = dark ? 0.22 : 0.1
  }
  if (innerMaterial) {
    innerMaterial.color.set(dark ? 0xff5252 : 0xe04545)
    innerMaterial.opacity = dark ? 0.14 : 0.07
  }
  if (vertexMaterial) {
    vertexMaterial.color.set(dark ? 0xff3b3b : 0xd40000)
    vertexMaterial.opacity = dark ? 0.85 : 0.55
  }
  if (starMaterial) {
    starMaterial.opacity = dark ? 0.9 : 0.4
  }
  if (starGeometry && starKinds) {
    paintStars(dark)
  }
}

/** 按星点类型重写颜色 attribute */
function paintStars(dark: boolean) {
  if (!starGeometry || !starKinds) return
  const count = starKinds.length
  const colors = new Float32Array(count * 3)
  const c = new THREE.Color()
  for (let i = 0; i < count; i++) {
    const kind = starKinds[i]
    if (kind === 1) {
      // 红点：稀疏的信号红粒子
      c.set(dark ? 0xff3b3b : 0xd40000)
      c.multiplyScalar(dark ? 0.9 : 0.7)
    } else if (kind === 2) {
      // 亮白锚点
      c.set(dark ? 0xf5f5f7 : 0x333338)
    } else {
      // 冷白尘：不同灰阶
      const v = dark ? 0.35 + Math.random() * 0.4 : 0.15 + Math.random() * 0.25
      c.setRGB(v, v, v + (dark ? 0.04 : 0))
    }
    colors[i * 3] = c.r
    colors[i * 3 + 1] = c.g
    colors[i * 3 + 2] = c.b
  }
  starGeometry.setAttribute('color', new THREE.BufferAttribute(colors, 3))
}

function handlePointerMove(e: PointerEvent) {
  pointer.x = (e.clientX / window.innerWidth) * 2 - 1
  pointer.y = (e.clientY / window.innerHeight) * 2 - 1
}

function handleResize() {
  if (!renderer || !camera) return
  const w = window.innerWidth
  const h = window.innerHeight
  camera.aspect = w / h
  camera.updateProjectionMatrix()
  renderer.setSize(w, h)
}

function handleVisibility() {
  if (document.hidden) {
    stopLoop()
  } else if (!reduceMotion.matches) {
    startLoop()
  }
}

let themeObserver: MutationObserver | null = null

function buildScene() {
  const container = containerRef.value
  if (!container) return

  // 移动端不启动 WebGL（仅保留纯黑/底色与 CRT 层）
  if (isMobile.matches) return

  try {
    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'low-power' })
  } catch {
    renderer = null
    return
  }
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  renderer.setSize(window.innerWidth, window.innerHeight)
  renderer.setClearColor(0x000000, 0)
  container.appendChild(renderer.domElement)

  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(55, window.innerWidth / window.innerHeight, 0.1, 100)
  camera.position.z = 7.2

  coreGroup = new THREE.Group()
  scene.add(coreGroup)
  const dark = isDark()

  // 主线框：二十面体（设计稿中的巨大多面体）
  const outerGeo = new THREE.IcosahedronGeometry(2.5, 1)
  outerMaterial = new THREE.LineBasicMaterial({
    color: dark ? 0xff2d2d : 0xd40000,
    transparent: true,
    opacity: dark ? 0.22 : 0.1
  })
  outerWire = new THREE.LineSegments(new THREE.EdgesGeometry(outerGeo), outerMaterial)
  coreGroup.add(outerWire)

  // 内层：八面体反向旋转
  const innerGeo = new THREE.OctahedronGeometry(1.55, 0)
  innerMaterial = new THREE.LineBasicMaterial({
    color: dark ? 0xff5252 : 0xe04545,
    transparent: true,
    opacity: dark ? 0.14 : 0.07
  })
  innerWire = new THREE.LineSegments(new THREE.EdgesGeometry(innerGeo), innerMaterial)
  coreGroup.add(innerWire)

  // 多面体顶点：红色节点
  vertexMaterial = new THREE.PointsMaterial({
    color: dark ? 0xff3b3b : 0xd40000,
    size: 0.045,
    sizeAttenuation: true,
    transparent: true,
    opacity: dark ? 0.85 : 0.55
  })
  vertexPoints = new THREE.Points(new THREE.IcosahedronGeometry(2.5, 1), vertexMaterial)
  coreGroup.add(vertexPoints)

  // 粒子星空：球壳内随机分布
  const STAR_COUNT = 1400
  starGeometry = new THREE.BufferGeometry()
  const positions = new Float32Array(STAR_COUNT * 3)
  starKinds = new Uint8Array(STAR_COUNT)
  for (let i = 0; i < STAR_COUNT; i++) {
    // 均匀球面分布，半径 6 ~ 22
    const r = 6 + Math.pow(Math.random(), 0.7) * 16
    const theta = Math.random() * Math.PI * 2
    const phi = Math.acos(2 * Math.random() - 1)
    positions[i * 3] = r * Math.sin(phi) * Math.cos(theta)
    positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta)
    positions[i * 3 + 2] = r * Math.cos(phi) - 4
    const roll = Math.random()
    starKinds[i] = roll < 0.06 ? 1 : roll < 0.12 ? 2 : 0
  }
  starGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3))
  paintStars(dark)
  starMaterial = new THREE.PointsMaterial({
    size: 0.035,
    sizeAttenuation: true,
    vertexColors: true,
    transparent: true,
    opacity: dark ? 0.9 : 0.4,
    depthWrite: false
  })
  starField = new THREE.Points(starGeometry, starMaterial)
  scene.add(starField)

  // 初始角度稍微偏转，避免正对一个面
  coreGroup.rotation.x = 0.18
  coreGroup.rotation.y = -0.35

  // 静态渲染一帧（降级 / reduced-motion 也保留画面）
  renderer.render(scene, camera)

  // 主题切换跟随
  themeObserver = new MutationObserver(applyThemeColors)
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

  window.addEventListener('resize', handleResize)
  window.addEventListener('pointermove', handlePointerMove, { passive: true })
  document.addEventListener('visibilitychange', handleVisibility)

  if (!reduceMotion.matches) startLoop()
}

let lastFrameAt = 0

function tick(now = 0) {
  if (disposed || !renderer || !scene || !camera) return
  frameId = requestAnimationFrame(tick)

  const dt = lastFrameAt ? Math.min((now - lastFrameAt) / 1000, 0.05) : 0.016
  lastFrameAt = now

  // 视差缓动
  smooth.x += (pointer.x - smooth.x) * 0.04
  smooth.y += (pointer.y - smooth.y) * 0.04

  if (coreGroup) {
    coreGroup.rotation.y += dt * 0.08
    coreGroup.rotation.x = 0.18 + smooth.y * 0.12
  }
  if (outerWire) outerWire.rotation.z += dt * 0.02
  if (innerWire) {
    innerWire.rotation.y -= dt * 0.16
    innerWire.rotation.x += dt * 0.05
  }
  if (vertexPoints) {
    vertexPoints.rotation.y += dt * 0.08
    vertexPoints.rotation.x = coreGroup?.rotation.x ?? 0
  }
  if (starField) starField.rotation.y += dt * 0.008

  camera.position.x = smooth.x * 0.35
  camera.position.y = -smooth.y * 0.22
  camera.lookAt(0, 0, 0)

  renderer.render(scene, camera)
}

function startLoop() {
  if (frameId) cancelAnimationFrame(frameId)
  lastFrameAt = 0
  frameId = requestAnimationFrame(tick)
}

function stopLoop() {
  if (frameId) cancelAnimationFrame(frameId)
  frameId = 0
}

onMounted(() => {
  buildScene()
})

onBeforeUnmount(() => {
  disposed = true
  stopLoop()
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('pointermove', handlePointerMove)
  document.removeEventListener('visibilitychange', handleVisibility)
  themeObserver?.disconnect()

  // 释放 GPU 资源，防止泄漏
  outerWire?.geometry.dispose()
  innerWire?.geometry.dispose()
  vertexPoints?.geometry.dispose()
  starGeometry?.dispose()
  outerMaterial?.dispose()
  innerMaterial?.dispose()
  vertexMaterial?.dispose()
  starMaterial?.dispose()
  renderer?.dispose()
  if (renderer?.domElement.parentElement === containerRef.value && renderer.domElement) {
    containerRef.value?.removeChild(renderer.domElement)
  }
  renderer = null
  scene = null
  camera = null
})
</script>

<style scoped>
.global-bg {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
}

.global-bg :deep(canvas) {
  display: block;
  width: 100%;
  height: 100%;
}

/* 低端降级：纯径向红雾，保持氛围 */
.global-bg::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 50% 42%, rgba(255, 45, 45, 0.06) 0%, transparent 55%),
    radial-gradient(circle at 82% 78%, rgba(255, 45, 45, 0.03) 0%, transparent 45%);
}

@media (max-width: 768px) {
  .global-bg::before {
    background:
      radial-gradient(circle at 50% 30%, rgba(255, 45, 45, 0.07) 0%, transparent 60%);
  }
}
</style>
