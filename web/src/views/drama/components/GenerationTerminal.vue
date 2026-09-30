<template>
  <!-- 生成终端：全屏沉浸式任务面板（左日志流 / 右大字 + 红线进度）
       只呈现真实轮询派生数据；固定深色终端表面，不随浅色主题翻转 -->
  <Teleport to="body">
    <!-- 最小化胶囊 -->
    <button
      v-if="t.active && t.minimized"
      class="term-pill"
      :class="{ 'term-pill--error': t.result === 'error' }"
      @click="restore()"
    >
      <span class="pill-dot" :class="{ 'pill-dot--idle': !t.running }" />
      <span class="pill-label">{{ pillLabel }}</span>
      <span v-if="t.running && t.progress != null" class="pill-pct">{{ t.progress }}%</span>
      <span v-else-if="t.result === 'success'" class="pill-ok">完成</span>
      <span v-else-if="t.result === 'error'" class="pill-err">异常</span>
    </button>

    <!-- 全屏终端 -->
    <transition name="term-fade">
      <div v-if="t.active && !t.minimized" class="term-root" role="dialog" aria-label="生成终端">
        <!-- 氛围层：扫描线 / 暗角 / 边缘红光 -->
        <div class="term-fx" aria-hidden="true">
          <div class="fx-scanlines" />
          <div class="fx-vignette" />
          <div class="fx-edge-glow" />
        </div>

        <div class="term-head">
          <p class="term-head-left">
            <span class="term-head-tag">日志://</span>{{ sessionCode }}
          </p>
          <p class="term-head-right" :class="systemLevelClass">
            <span class="sys-dot" />{{ systemText }}
          </p>
        </div>

        <div class="term-body">
          <!-- 左：日志流 -->
          <section class="term-logpanel">
            <div ref="logBoxRef" class="term-logs">
              <p
                v-for="line in t.logs"
                :key="line.id"
                class="term-log"
                :class="`term-log--${line.level}`"
              >
                <span class="log-prefix">&gt;</span>
                <span class="log-time">[{{ line.time }}]</span>
                {{ line.text }}
              </p>
              <p v-if="t.logs.length === 0" class="term-log term-log--muted">
                <span class="log-prefix">&gt;</span> WAITING_FOR_OUTPUT…
              </p>
              <span v-if="t.running" class="log-cursor" aria-hidden="true">▌</span>
            </div>
          </section>

          <!-- 右：大字状态 + 红线进度 -->
          <section class="term-stage">
            <div class="stage-chip">
              <span class="chip-dot" />{{ t.chip }}
            </div>
            <h2 class="stage-title" :key="t.kind + '|' + t.subject">
              <span class="stage-line stage-line--1">{{ t.title }}</span>
              <span class="stage-line stage-line--accent stage-line--2">{{ t.subject }}</span>
            </h2>

            <!-- 结束结论 -->
            <div v-if="t.result" class="stage-result" :class="`stage-result--${t.result}`">
              <span class="result-mark">{{ t.result === "success" ? "✓" : "✕" }}</span>
              {{ t.resultText }}
            </div>

            <!-- 进度区 -->
            <div class="stage-progress">
              <div class="progress-head">
                <span class="progress-label">总体进度 OVERALL</span>
                <span v-if="t.progress != null" class="progress-pct">{{ t.progress.toFixed(1) }}%</span>
                <span v-else class="progress-pct progress-pct--muted">--.-%</span>
              </div>
              <div class="progress-track">
                <div
                  v-if="t.progress != null"
                  class="progress-fill"
                  :style="{ width: t.progress + '%' }"
                />
                <div v-else class="progress-indeterminate" />
              </div>
              <div class="progress-meta">
                <span>{{ t.meta || "—" }}</span>
                <span class="meta-elapsed">ELAPSED: {{ elapsedText }}</span>
              </div>
            </div>

            <!-- 操作 -->
            <div class="stage-actions">
              <button
                v-if="t.running"
                class="term-btn term-btn--ghost"
                @click="minimize()"
              >
                最小化 <span class="btn-mono">MINIMIZE</span>
              </button>
              <button
                v-else
                class="term-btn term-btn--solid"
                @click="close()"
              >
                关闭终端 <span class="btn-mono">CLOSE</span>
              </button>
            </div>
          </section>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch, onBeforeUnmount } from 'vue'
import { useGenerationTerminal } from '@/composables/useGenerationTerminal'

const {
  terminal: t,
  formatElapsed,
  minimize,
  restore,
  close,
} = useGenerationTerminal()

const logBoxRef = ref<HTMLElement | null>(null)

const sessionCode = computed(() => {
  const map: Record<string, string> = {
    'video-batch': 'VIDEO_BATCH_01',
    'video-single': 'VIDEO_SHOT_01',
    storyboard: 'STRUCTURE_PARSE_01',
    image: 'IMAGE_SYNTH_01',
  }
  return map[t.kind] || 'TERMINAL_01'
})

const pillLabel = computed(() => {
  const map: Record<string, string> = {
    'video-batch': '批量渲染',
    'video-single': '单镜渲染',
    storyboard: '分镜拆分',
    image: '图像合成',
  }
  return `${map[t.kind] || '任务'} // ${sessionCode.value}`
})

const systemText = computed(() => `系统: ${t.systemStatus || 'STANDBY'}`)
const systemLevelClass = computed(() => {
  if (t.result === 'error') return 'sys--error'
  if (t.result === 'success') return 'sys--ok'
  return 'sys--run'
})

const elapsedText = computed(() => formatElapsed(t.elapsed))

// 日志自动滚底
watch(
  () => t.logs.length,
  async () => {
    await nextTick()
    const box = logBoxRef.value
    if (box) box.scrollTop = box.scrollHeight
  },
)

// ESC 最小化（运行中不直接关闭，避免误丢过程）
const onKey = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && t.active && !t.minimized && t.running) minimize()
}
window.addEventListener('keydown', onKey)
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
</script>

<style scoped>
/* 终端为固定深色表面：两套主题下都保持 CRT 黑 */
.term-root {
  position: fixed;
  z-index: 9995;
  top: 56px;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(5, 5, 7, 0.97);
  color: #f5f5f7;
  display: flex;
  flex-direction: column;
  font-family: var(--font-sans);
  overflow: hidden;
  animation: term-boot 0.5s cubic-bezier(0.16, 1, 0.3, 1) both;
}

@keyframes term-boot {
  from { clip-path: inset(0 0 100% 0); }
  to { clip-path: inset(0 0 0 0); }
}

/* ── 氛围层：扫描线 / 暗角 / 边缘红光 ─────── */
.term-fx {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.term-head,
.term-body {
  position: relative;
  z-index: 1;
}

.fx-scanlines {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    to bottom,
    rgba(255, 255, 255, 0.025) 0px,
    rgba(255, 255, 255, 0.025) 1px,
    transparent 1px,
    transparent 4px
  );
  animation: fx-scan-drift 8s linear infinite;
}

@keyframes fx-scan-drift {
  from { background-position-y: 0; }
  to { background-position-y: 48px; }
}

.fx-vignette {
  position: absolute;
  inset: 0;
  background: radial-gradient(
    ellipse at 50% 42%,
    transparent 55%,
    rgba(0, 0, 0, 0.55) 100%
  );
}

.fx-edge-glow {
  position: absolute;
  inset: 0;
  box-shadow: inset 0 0 120px -40px rgba(255, 45, 45, 0.35);
  animation: fx-glow-breathe 3.6s ease-in-out infinite;
}

@keyframes fx-glow-breathe {
  0%, 100% { opacity: 0.55; }
  50% { opacity: 1; }
}

/* 系统状态点 */
.sys-dot,
.chip-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  margin-right: 8px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 8px currentColor;
  animation: sys-blink 1.1s steps(2, end) infinite;
}

.chip-dot {
  width: 5px;
  height: 5px;
  margin-right: 9px;
  animation: sys-blink 1.6s steps(2, end) infinite;
}

.sys--ok .sys-dot,
.sys--error .sys-dot {
  animation: none;
}

@keyframes sys-blink {
  50% { opacity: 0.25; }
}

.term-fade-enter-active,
.term-fade-leave-active {
  transition: opacity 0.22s ease;
}
.term-fade-enter-from,
.term-fade-leave-to {
  opacity: 0;
}

/* ── 顶部行 ─────────────────────────────── */
.term-head {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 32px;
  border-bottom: 1px solid rgba(245, 245, 247, 0.08);
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.08em;
}

.term-head-tag {
  color: #ff5252;
}

.term-head-left {
  margin: 0;
  color: #8a8a92;
}

.term-head-right {
  margin: 0;
  font-weight: 700;
}

.sys--run { color: #ff5252; }
.sys--ok { color: #3ddc84; }
.sys--error { color: #ffb020; }

/* ── 主体两栏 ───────────────────────────── */
.term-body {
  flex: 1;
  display: flex;
  min-height: 0;
}

.term-logpanel {
  width: 38%;
  min-width: 340px;
  max-width: 560px;
  padding: 20px 0 20px 32px;
  display: flex;
  position: relative;
}

/* 四角取景框（外框两个角） */
.term-logpanel::before,
.term-logpanel::after {
  content: "";
  position: absolute;
  width: 14px;
  height: 14px;
  pointer-events: none;
  z-index: 2;
}

.term-logpanel::before {
  top: 12px;
  left: 24px;
  border-top: 1px solid rgba(255, 45, 45, 0.7);
  border-left: 1px solid rgba(255, 45, 45, 0.7);
}

.term-logpanel::after {
  bottom: 12px;
  right: 16px;
  border-bottom: 1px solid rgba(255, 45, 45, 0.7);
  border-right: 1px solid rgba(255, 45, 45, 0.7);
}

.term-logs {
  flex: 1;
  overflow-y: auto;
  border: 1px solid rgba(245, 245, 247, 0.1);
  padding: 18px 20px;
  margin-right: 24px;
  background: rgba(255, 255, 255, 0.015);
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.9;
  scrollbar-width: thin;
  scrollbar-color: #323238 transparent;
  position: relative;
}

/* 内框另外两个角 */
.term-logs::before,
.term-logs::after {
  content: "";
  position: absolute;
  width: 10px;
  height: 10px;
  pointer-events: none;
}

.term-logs::before {
  top: -1px;
  right: -1px;
  border-top: 1px solid rgba(245, 245, 247, 0.35);
  border-right: 1px solid rgba(245, 245, 247, 0.35);
}

.term-logs::after {
  bottom: -1px;
  left: -1px;
  border-bottom: 1px solid rgba(245, 245, 247, 0.35);
  border-left: 1px solid rgba(245, 245, 247, 0.35);
}

.term-log {
  margin: 0;
  color: #b9b9c0;
  word-break: break-all;
  animation: log-in 0.28s cubic-bezier(0.16, 1, 0.3, 1) both;
}

@keyframes log-in {
  from {
    opacity: 0;
    transform: translateX(-8px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.log-prefix {
  color: #ff5252;
  margin-right: 8px;
  font-weight: 700;
}

.log-time {
  color: #55555c;
  margin-right: 8px;
  font-size: 10.5px;
}

.term-log--success { color: #3ddc84; }
.term-log--success .log-prefix { color: #3ddc84; }
.term-log--warning { color: #ffb020; }
.term-log--warning .log-prefix { color: #ffb020; }
.term-log--error { color: #ff5252; }
.term-log--muted { color: #55555c; }

/* 闪烁块光标 */
.log-cursor {
  display: inline-block;
  margin-top: 4px;
  color: #ff5252;
  font-size: 12px;
  line-height: 1;
  animation: cursor-blink 0.9s steps(2, end) infinite;
}

@keyframes cursor-blink {
  50% { opacity: 0; }
}

/* ── 右侧舞台 ───────────────────────────── */
.term-stage {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 40px 64px 60px;
  min-width: 0;
}

.stage-chip {
  align-self: flex-start;
  border: 1px solid rgba(255, 45, 45, 0.55);
  color: #ff5252;
  background: rgba(255, 45, 45, 0.08);
  padding: 7px 16px;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.12em;
  margin-bottom: 30px;
}

.stage-title {
  margin: 0 0 8px;
  font-weight: 900;
  line-height: 1.06;
}

.stage-line {
  display: block;
  font-size: clamp(44px, 6vw, 84px);
  color: #f5f5f7;
}

.stage-line--1 {
  animation: stage-rise 0.55s cubic-bezier(0.16, 1, 0.3, 1) both;
}

.stage-line--2 {
  animation: stage-rise 0.55s cubic-bezier(0.16, 1, 0.3, 1) 0.12s both,
    stage-flicker 4.2s linear 0.8s infinite;
}

@keyframes stage-rise {
  from {
    opacity: 0;
    transform: translateY(0.35em);
    clip-path: inset(0 0 100% 0);
  }
  to {
    opacity: 1;
    transform: translateY(0);
    clip-path: inset(0 0 0 0);
  }
}

/* CRT 信号扰动：极轻微的闪烁+红影错位，不影响阅读 */
@keyframes stage-flicker {
  0%, 92%, 100% {
    opacity: 1;
    text-shadow: 0 0 34px rgba(255, 45, 45, 0.35);
    transform: translateX(0);
  }
  93% {
    opacity: 0.75;
    text-shadow: -2px 0 rgba(0, 200, 255, 0.35), 2px 0 rgba(255, 45, 45, 0.6);
    transform: translateX(1px);
  }
  94.5% { opacity: 1; transform: translateX(-1px); }
  96% { opacity: 0.85; transform: translateX(0); }
}

.stage-line--accent {
  color: #ff2d2d;
  text-shadow: 0 0 34px rgba(255, 45, 45, 0.35);
}

/* ── 结论 ───────────────────────────────── */
.stage-result {
  margin-top: 22px;
  font-family: var(--font-mono);
  font-size: 14px;
  letter-spacing: 0.06em;
}

.stage-result--success { color: #3ddc84; }
.stage-result--error { color: #ffb020; }

.result-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  margin-right: 10px;
  border: 1px solid currentColor;
  font-size: 12px;
}

/* ── 进度 ───────────────────────────────── */
.stage-progress {
  margin-top: 42px;
  max-width: 640px;
  width: 100%;
}

.progress-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 12px;
}

.progress-label {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.18em;
  color: #8a8a92;
}

.progress-pct {
  font-family: var(--font-mono);
  font-size: 30px;
  font-weight: 700;
  color: #f5f5f7;
  letter-spacing: 0.02em;
}

.progress-pct--muted {
  color: #55555c;
}

.progress-track {
  position: relative;
  height: 3px;
  background: rgba(245, 245, 247, 0.12);
  overflow: hidden;
}

.progress-fill {
  position: relative;
  height: 100%;
  background: #ff2d2d;
  box-shadow: 0 0 12px rgba(255, 45, 45, 0.8);
  transition: width 0.6s cubic-bezier(0.16, 1, 0.3, 1);
  overflow: hidden;
}

/* 进度条流光 */
.progress-fill::after {
  content: "";
  position: absolute;
  inset: 0;
  background: linear-gradient(
    100deg,
    transparent 20%,
    rgba(255, 255, 255, 0.55) 50%,
    transparent 80%
  );
  transform: translateX(-100%);
  animation: progress-shine 1.8s ease-in-out infinite;
}

@keyframes progress-shine {
  60%, 100% { transform: translateX(100%); }
}

.progress-indeterminate {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 32%;
  background: #ff2d2d;
  box-shadow: 0 0 12px rgba(255, 45, 45, 0.8);
  animation: term-slide 1.4s cubic-bezier(0.45, 0, 0.55, 1) infinite;
}

@keyframes term-slide {
  0% { left: -34%; }
  100% { left: 102%; }
}

.progress-meta {
  margin-top: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.05em;
  color: #8a8a92;
}

.meta-elapsed {
  color: #55555c;
  white-space: nowrap;
}

/* ── 按钮 ───────────────────────────────── */
.stage-actions {
  margin-top: 44px;
  display: flex;
  gap: 14px;
}

.term-btn {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  height: 44px;
  padding: 0 24px;
  font-size: 14px;
  font-weight: 700;
  font-family: inherit;
  cursor: pointer;
  transition: all var(--transition-fast, 120ms ease);
}

.btn-mono {
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.14em;
  opacity: 0.8;
}

.term-btn--ghost {
  background: transparent;
  color: #f5f5f7;
  border: 1px solid rgba(245, 245, 247, 0.25);
}

.term-btn--ghost:hover {
  border-color: #ff2d2d;
  color: #ff5252;
}

.term-btn--solid {
  background: #f5f5f7;
  color: #050506;
  border: 1px solid #f5f5f7;
}

.term-btn--solid:hover {
  box-shadow: 0 0 22px -4px rgba(255, 45, 45, 0.6);
  border-color: #ff2d2d;
}

/* ── 最小化胶囊 ─────────────────────────── */
.term-pill {
  position: fixed;
  z-index: 9996;
  right: 24px;
  bottom: 24px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  height: 38px;
  padding: 0 16px;
  background: rgba(10, 10, 12, 0.92);
  border: 1px solid rgba(255, 45, 45, 0.6);
  color: #f5f5f7;
  font-size: 12px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  box-shadow: 0 0 22px -6px rgba(255, 45, 45, 0.5);
  transition: box-shadow 0.2s ease;
}

.term-pill:hover {
  box-shadow: 0 0 30px -4px rgba(255, 45, 45, 0.75);
}

.term-pill--error {
  border-color: #ffb020;
  box-shadow: 0 0 22px -6px rgba(255, 176, 32, 0.55);
}

.pill-dot {
  width: 7px;
  height: 7px;
  background: #ff2d2d;
  border-radius: 50%;
  box-shadow: 0 0 8px #ff2d2d;
  animation: pill-pulse 1.2s steps(2, end) infinite;
}

.pill-dot--idle {
  animation: none;
  background: #3ddc84;
  box-shadow: 0 0 8px #3ddc84;
}

.term-pill--error .pill-dot--idle {
  background: #ffb020;
  box-shadow: 0 0 8px #ffb020;
}

@keyframes pill-pulse {
  50% { opacity: 0.3; }
}

.pill-label {
  font-family: var(--font-mono);
  font-size: 10.5px;
  letter-spacing: 0.06em;
}

.pill-pct {
  font-family: var(--font-mono);
  font-size: 11px;
  color: #ff5252;
  font-weight: 700;
}

.pill-ok {
  font-family: var(--font-mono);
  font-size: 10px;
  color: #3ddc84;
  letter-spacing: 0.1em;
}

.pill-err {
  font-family: var(--font-mono);
  font-size: 10px;
  color: #ffb020;
  letter-spacing: 0.1em;
}

/* ── 响应式：上下堆叠 ───────────────────── */
@media (max-width: 768px) {
  .term-root {
    top: 52px;
  }

  .term-head {
    padding: 12px 16px;
  }

  .term-body {
    flex-direction: column-reverse;
  }

  .term-stage {
    padding: 24px 20px 12px;
    justify-content: flex-start;
  }

  .stage-chip {
    margin-bottom: 18px;
  }

  .stage-progress {
    margin-top: 26px;
  }

  .stage-actions {
    margin-top: 26px;
  }

  .term-logpanel {
    width: 100%;
    min-width: 0;
    max-width: none;
    flex: 1;
    padding: 8px 16px 16px;
  }

  .term-logs {
    margin-right: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .progress-indeterminate,
  .fx-scanlines,
  .fx-edge-glow,
  .sys-dot,
  .chip-dot,
  .log-cursor,
  .stage-line--2,
  .progress-fill::after,
  .term-log {
    animation: none !important;
  }

  .progress-indeterminate {
    width: 100%;
    opacity: 0.6;
  }

  .term-root {
    animation: none;
  }

  .term-fade-enter-active,
  .term-fade-leave-active {
    transition: none;
  }
}
</style>
