<template>
  <!-- 全局 CRT 氛围层：扫描线 + 暗角 + 四角取景框 + 十字准星
       纯装饰、pointer-events:none；移动端隐藏取景框与准星 -->
  <div class="crt-overlay" aria-hidden="true">
    <div class="crt-scanlines" />
    <div class="crt-vignette" />
    <div class="crt-flicker" />
    <span class="crt-corner crt-corner--tl" />
    <span class="crt-corner crt-corner--tr" />
    <span class="crt-corner crt-corner--bl" />
    <span class="crt-corner crt-corner--br" />
    <span v-if="crosshairEnabled" ref="crossRef" class="crt-cross" :class="{ visible: crossVisible }" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'

const crossRef = ref<HTMLSpanElement | null>(null)
const crossVisible = ref(false)

const coarsePointer = window.matchMedia('(pointer: coarse)')
const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)')
const crosshairEnabled = !coarsePointer.matches

let rafId = 0
const target = { x: -100, y: -100 }
const pos = { x: -100, y: -100 }

function onPointerMove(e: PointerEvent) {
  target.x = e.clientX
  target.y = e.clientY
  crossVisible.value = true
}

function onPointerLeave() {
  crossVisible.value = false
}

/** 准星轻微滞后，制造"取景追踪"感 */
function follow() {
  rafId = requestAnimationFrame(follow)
  pos.x += (target.x - pos.x) * 0.22
  pos.y += (target.y - pos.y) * 0.22
  if (crossRef.value) {
    crossRef.value.style.transform = `translate(${pos.x}px, ${pos.y}px) translate(-50%, -50%)`
  }
}

onMounted(() => {
  if (crosshairEnabled && !reduceMotion.matches) {
    window.addEventListener('pointermove', onPointerMove, { passive: true })
    document.documentElement.addEventListener('mouseleave', onPointerLeave)
    rafId = requestAnimationFrame(follow)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', onPointerMove)
  document.documentElement.removeEventListener('mouseleave', onPointerLeave)
  cancelAnimationFrame(rafId)
})
</script>

<style scoped>
.crt-overlay {
  position: fixed;
  inset: 0;
  z-index: 9990;
  pointer-events: none;
  overflow: hidden;
}

/* 扫描线 */
.crt-scanlines {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    to bottom,
    rgba(255, 255, 255, var(--crt-scanline-opacity)) 0,
    rgba(255, 255, 255, var(--crt-scanline-opacity)) 1px,
    transparent 1px,
    transparent 3px
  );
}

/* 暗角 */
.crt-vignette {
  position: absolute;
  inset: 0;
  background: radial-gradient(
    ellipse at center,
    transparent 52%,
    rgba(0, 0, 0, var(--crt-vignette-opacity)) 100%
  );
}

/* 极轻微的亮度闪烁（浅色主题关闭） */
.crt-flicker {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.012);
  animation: crt-flicker var(--crt-flicker-duration) steps(2, end) infinite;
}

@keyframes crt-flicker {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}

/* 四角取景框 */
.crt-corner {
  position: absolute;
  width: 18px;
  height: 18px;
  border: 0;
  border-color: var(--crt-bracket-color);
  border-style: solid;
}

.crt-corner--tl {
  top: 62px;
  left: 18px;
  border-top-width: 1px;
  border-left-width: 1px;
}

.crt-corner--tr {
  top: 62px;
  right: 18px;
  border-top-width: 1px;
  border-right-width: 1px;
}

.crt-corner--bl {
  bottom: 18px;
  left: 18px;
  border-bottom-width: 1px;
  border-left-width: 1px;
}

.crt-corner--br {
  bottom: 18px;
  right: 18px;
  border-bottom-width: 1px;
  border-right-width: 1px;
}

/* 十字准星 */
.crt-cross {
  position: absolute;
  top: 0;
  left: 0;
  width: 22px;
  height: 22px;
  opacity: 0;
  transition: opacity 0.25s ease;
  will-change: transform;
}

.crt-cross.visible {
  opacity: 0.55;
}

.crt-cross::before,
.crt-cross::after {
  content: '';
  position: absolute;
  background: var(--accent);
}

.crt-cross::before {
  left: 50%;
  top: 0;
  width: 1px;
  height: 100%;
  transform: translateX(-50%);
}

.crt-cross::after {
  top: 50%;
  left: 0;
  height: 1px;
  width: 100%;
  transform: translateY(-50%);
}

@media (max-width: 768px) {
  .crt-corner {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .crt-flicker {
    animation: none;
  }
}
</style>
