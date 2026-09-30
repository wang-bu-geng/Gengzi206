<template>
  <div class="cyber-page">
    <!-- ══════════ HERO / 生成终端首屏 ══════════ -->
    <section class="hero">
      <!-- 左侧真实数据 HUD（全部来自 stats 接口 / 会话计时 / 本地时钟） -->
      <aside class="hud hud--left">
        <p class="hud-line hud-ok">
          <span class="hud-dot" />STATUS: ACTIVE
        </p>
        <p class="hud-line">WORKS_TOTAL: <b>{{ String(stats.projects).padStart(3, '0') }}</b></p>
        <p class="hud-line">ASSETS_IMG: <b>{{ String(stats.images).padStart(3, '0') }}</b></p>
        <p class="hud-line">ASSETS_VID: <b>{{ String(stats.videos).padStart(3, '0') }}</b></p>
        <p class="hud-line hud-gap">UPTIME: <b>{{ uptime }}</b></p>
        <p class="hud-line hud-dim">NODE: LOCAL_SSR</p>
      </aside>

      <!-- 右侧 HUD -->
      <aside class="hud hud--right">
        <p class="hud-line hud-dim">LAST_SYNC: {{ clock }}</p>
        <p class="hud-line hud-dim">VERSION 1.0.0</p>
        <p class="hud-line hud-gap"><span class="hud-rec">REC</span></p>
      </aside>

      <div class="hero-inner">
        <div class="hero-pill">
          <span class="pill-cn">AI 驱动 · 一键生成</span>
          <span class="pill-en">GENERATIVE_CORE</span>
        </div>

        <h1 class="hero-title">
          <span class="title-line">打造你的</span>
          <span class="title-line">专属短剧</span>
        </h1>

        <p class="hero-subtitle">
          从剧本到成片，全程 AI 辅助。只需一个创意，<br class="subtitle-br" />
          即可生成完整的微短剧作品。突破创作边界，重新定义叙事。
        </p>

        <div class="hero-cta">
          <button class="btn-terminal btn-terminal--solid" @click="handleCreate">
            <span>立即开始创作</span>
            <span class="btn-en">START_GEN</span>
          </button>
          <button class="btn-terminal btn-terminal--ghost" @click="goToDramas">
            <span>浏览我的作品</span>
            <span class="btn-en">BROWSE</span>
          </button>
        </div>
      </div>

      <div class="scroll-hint">
        <span class="scroll-line" />
        <span class="scroll-text">SCROLL TO DISCOVER</span>
      </div>
    </section>

    <!-- ══════════ CORE_MODULES / 核心模块 ══════════ -->
    <section class="panel-section">
      <div class="section-inner">
        <div class="section-head">
          <h2 class="section-title">
            <span class="title-bar" />核心模块
            <span class="title-en">/ CORE_MODULES</span>
          </h2>
          <p class="section-sub">从剧本创作到视频生成，全流程智能化</p>
        </div>

        <div class="module-grid">
          <div class="module-card" v-for="(feature, idx) in features" :key="feature.title">
            <p class="module-index">MODULE_{{ String(idx + 1).padStart(2, '0') }}</p>
            <div class="module-icon">
              <el-icon :size="20">
                <component :is="feature.icon" />
              </el-icon>
            </div>
            <h3 class="module-name">{{ feature.title }}</h3>
            <p class="module-desc">{{ feature.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- ══════════ WORKFLOW / 创作链路 ══════════ -->
    <section class="panel-section panel-section--tight">
      <div class="section-inner">
        <div class="section-head">
          <h2 class="section-title">
            <span class="title-bar" />创作链路
            <span class="title-en">/ GENERATION_PIPELINE</span>
          </h2>
          <p class="section-sub">四步完成，人人都能做导演</p>
        </div>

        <div class="pipeline">
          <template v-for="(step, index) in workflowSteps" :key="step.title">
            <div class="pipeline-step">
              <p class="step-code">STEP_{{ String(index + 1).padStart(2, '0') }}</p>
              <h3 class="step-name">{{ step.title }}</h3>
              <p class="step-desc">{{ step.desc }}</p>
            </div>
            <span v-if="index < workflowSteps.length - 1" class="pipeline-link">→</span>
          </template>
        </div>
      </div>
    </section>

    <!-- ══════════ WORKS_ARCHIVE / 最近存档 ══════════ -->
    <section v-if="recentDramas.length > 0" class="panel-section panel-section--tight">
      <div class="section-inner">
        <div class="section-head section-head--row">
          <h2 class="section-title">
            <span class="title-bar" />作品存档
            <span class="title-en">/ WORKS_ARCHIVE</span>
          </h2>
          <a class="archive-more" @click="goToDramas">
            查看全部 <span class="font-mono">VIEW_ALL →</span>
          </a>
        </div>

        <div class="archive-grid">
          <div
            v-for="(drama, idx) in recentDramas"
            :key="drama.id"
            class="archive-card"
            :class="{ 'archive-card--active': idx === 0 }"
            @click="viewDrama(drama.id)"
          >
            <div class="archive-cover">
              <!-- 无线框占位：暂无封面数据时的取景框样式 -->
              <span class="cover-frame cover-frame--a" />
              <span class="cover-frame cover-frame--b" />
              <span class="cover-id">#{{ String(drama.id).padStart(3, '0') }}</span>
              <span class="cover-ep">{{ drama.total_episodes || 0 }} EPS</span>
            </div>
            <div class="archive-info">
              <h3 class="archive-name">{{ drama.title }}</h3>
              <p v-if="drama.description" class="archive-desc">{{ drama.description }}</p>
              <div class="archive-meta">
                <span class="meta-chip">{{ getStyleLabel(drama.style) }}</span>
                <span class="meta-time">{{ formatDate(drama.updated_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ══════════ CTA / 终端引导 ══════════ -->
    <section class="cta-terminal">
      <p class="cta-prompt"><span class="prompt-arrow">&gt;</span> 准备好开始创作了吗？</p>
      <h2 class="cta-title">让 AI 成为你的创作伙伴</h2>
      <button class="btn-terminal btn-terminal--solid cta-btn" @click="handleCreate">
        <span>立即开始</span>
        <span class="btn-en">INIT_SESSION</span>
      </button>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import {
  EditPen,
  Picture,
  VideoPlay,
  Film
} from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'

const router = useRouter()

const stats = reactive({
  projects: 0,
  images: 0,
  videos: 0,
  tasks: 0
})

const recentDramas = ref<any[]>([])

// 会话存活时间（真实数据，从页面挂载开始计时）
const uptime = ref('00.00.00')
// 本地时钟（真实数据，用于 LAST_SYNC 展示）
const clock = ref('')
let tickTimer = 0
const mountedAt = Date.now()

const pad2 = (n: number) => String(n).padStart(2, '0')

const tick = () => {
  const elapsed = Math.floor((Date.now() - mountedAt) / 1000)
  uptime.value = `${pad2(Math.floor(elapsed / 3600))}.${pad2(Math.floor((elapsed % 3600) / 60))}.${pad2(elapsed % 60)}`
  const d = new Date()
  clock.value =
    `${d.getFullYear()}.${pad2(d.getMonth() + 1)}.${pad2(d.getDate())} ` +
    `${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`
}

const features = [
  {
    icon: EditPen,
    title: '智能剧本生成',
    desc: '输入创意关键词，AI 自动生成完整剧本，支持多风格多主题'
  },
  {
    icon: Picture,
    title: 'AI 图片生成',
    desc: '基于角色和场景描述，一键生成高质量分镜插图'
  },
  {
    icon: VideoPlay,
    title: '视频一键合成',
    desc: '智能串联图片与音频，自动生成完整短剧视频'
  },
  {
    icon: Film,
    title: '角色资产管理',
    desc: '统一管理角色、场景、道具，保持作品风格一致性'
  }
]

const workflowSteps = [
  {
    title: '创建项目',
    desc: '输入剧名和简介，选择你喜欢的画风'
  },
  {
    title: '生成剧本',
    desc: 'AI 根据创意自动生成多集完整剧本'
  },
  {
    title: '制作分镜',
    desc: '角色设定 + 场景描述，AI 生成精美插图'
  },
  {
    title: '合成视频',
    desc: '一键合成带配音的完整短剧视频'
  }
]

const styleMap: Record<string, string> = {
  ghibli: '吉卜力',
  guoman: '国漫',
  wasteland: '废土',
  nostalgia: '复古',
  pixel: '像素',
  voxel: '体素',
  urban: '都市',
  guoman3d: '3D国漫',
  chibi3d: 'Q版3D'
}

const getStyleLabel = (style: string) => styleMap[style] || style

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days}天前`
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const handleCreate = () => {
  router.push('/dramas/create')
}

const goToDramas = () => {
  router.push('/dramas')
}

const viewDrama = (id: string) => {
  router.push(`/dramas/${id}`)
}

onMounted(async () => {
  tick()
  tickTimer = window.setInterval(tick, 1000)

  try {
    const res = await dramaAPI.getStats()
    stats.projects = res.projects ?? 0
    stats.images = res.images ?? 0
    stats.videos = res.videos ?? 0
    stats.tasks = res.tasks ?? 0
  } catch {
    // 静默失败
  }

  try {
    const listRes = await dramaAPI.list({ page: 1, page_size: 4 })
    recentDramas.value = listRes.items || []
  } catch {
    // 静默失败
  }
})

onBeforeUnmount(() => {
  window.clearInterval(tickTimer)
})
</script>

<style scoped>
/* ════════════════════════════════════════════
   GENGZI Cyber Terminal — Dashboard
   ════════════════════════════════════════════ */
.cyber-page {
  overflow-x: hidden;
}

/* ── HERO ──────────────────────────────── */
.hero {
  position: relative;
  min-height: calc(100vh - 64px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 24px 72px;
}

.hero-inner {
  position: relative;
  max-width: 920px;
  text-align: center;
}

.hero-pill {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  border: 1px solid var(--accent);
  padding: 7px 18px;
  margin-bottom: 34px;
  color: var(--accent);
  background: var(--accent-light);
}

.pill-cn {
  font-size: 12.5px;
  font-weight: 600;
  letter-spacing: 0.32em;
  text-indent: 0.32em;
}

.pill-en {
  font-family: var(--font-mono);
  font-size: 9.5px;
  font-weight: 700;
  letter-spacing: 0.18em;
  opacity: 0.75;
}

.hero-title {
  margin: 0 0 26px;
  font-weight: 900;
  line-height: 1.04;
  letter-spacing: 0.01em;
  color: var(--text-primary);
}

.title-line {
  display: block;
  font-size: clamp(48px, 9.2vw, 128px);
}

.hero-subtitle {
  font-size: 16px;
  line-height: 1.85;
  color: var(--text-secondary);
  margin: 0 auto 42px;
  max-width: 560px;
}

/* ── 终端按钮 ──────────────────────────── */
.hero-cta {
  display: flex;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
}

.btn-terminal {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-width: 218px;
  height: 52px;
  padding: 0 26px;
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.04em;
  cursor: pointer;
  font-family: inherit;
  transition: all var(--transition-fast);
}

.btn-en {
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.16em;
  opacity: 0.75;
}

.btn-terminal--solid {
  background: var(--text-primary);
  color: var(--bg-primary);
  border: 1px solid var(--text-primary);
}

.btn-terminal--solid:hover {
  box-shadow: 0 0 26px -4px rgba(255, 45, 45, 0.55);
  border-color: var(--accent);
}

.btn-terminal--ghost {
  background: transparent;
  color: var(--text-primary);
  border: 1px solid var(--border-secondary);
}

.btn-terminal--ghost:hover {
  border-color: var(--accent);
  color: var(--accent);
}

/* ── 两侧 HUD ──────────────────────────── */
.hud {
  position: absolute;
  top: 50%;
  transform: translateY(-58%);
  font-family: var(--font-mono);
  font-size: 11px;
  line-height: 2;
  letter-spacing: 0.08em;
  color: var(--text-muted);
  text-align: left;
}

.hud--left {
  left: 36px;
}

.hud--right {
  right: 36px;
  text-align: right;
}

.hud-line b {
  color: var(--text-secondary);
  font-weight: 500;
}

.hud-ok {
  color: var(--success);
}

.hud-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  margin-right: 6px;
  background: var(--success);
  border-radius: 50%;
  box-shadow: 0 0 8px var(--success);
  animation: hud-pulse 2s steps(2, end) infinite;
}

@keyframes hud-pulse {
  50% { opacity: 0.35; }
}

.hud-gap {
  margin-top: 14px;
}

.hud-dim {
  opacity: 0.75;
}

.hud-rec {
  color: var(--accent);
  font-weight: 700;
  letter-spacing: 0.2em;
}

.hud-rec::after {
  content: ' ●';
  animation: hud-pulse 1.2s steps(2, end) infinite;
}

/* ── SCROLL 提示 ───────────────────────── */
.scroll-hint {
  position: absolute;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.scroll-line {
  width: 1px;
  height: 34px;
  background: var(--accent);
  opacity: 0.7;
  animation: scroll-drop 1.8s ease-in-out infinite;
  transform-origin: top;
}

@keyframes scroll-drop {
  0% { transform: scaleY(0); transform-origin: top; }
  45% { transform: scaleY(1); transform-origin: top; }
  55% { transform: scaleY(1); transform-origin: bottom; }
  100% { transform: scaleY(0); transform-origin: bottom; }
}

.scroll-text {
  font-family: var(--font-mono);
  font-size: 9px;
  letter-spacing: 0.32em;
  text-indent: 0.32em;
  color: var(--text-muted);
}

/* ── 后续区块：半透明面板保证可读 ──────── */
.panel-section {
  position: relative;
  padding: 96px 24px;
  background: linear-gradient(
    180deg,
    transparent 0%,
    color-mix(in srgb, var(--bg-primary) 88%, transparent) 6%,
    color-mix(in srgb, var(--bg-primary) 94%, transparent) 100%
  );
  border-top: 1px solid var(--border-primary);
}

.panel-section--tight {
  padding-top: 24px;
}

.section-inner {
  max-width: 1200px;
  margin: 0 auto;
}

.section-head {
  margin-bottom: 48px;
}

.section-head--row {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 14px;
  margin: 0 0 12px;
  font-size: 30px;
  font-weight: 800;
  color: var(--text-primary);
}

.title-bar {
  display: inline-block;
  width: 4px;
  height: 30px;
  background: var(--accent);
  box-shadow: 0 0 12px rgba(255, 45, 45, 0.5);
}

.title-en {
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}

.section-sub {
  margin: 0;
  padding-left: 18px;
  font-size: 14.5px;
  color: var(--text-secondary);
}

/* ── 模块卡 ────────────────────────────── */
.module-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.module-card {
  position: relative;
  padding: 24px 22px 26px;
  background: color-mix(in srgb, var(--bg-card) 82%, transparent);
  border: 1px solid var(--border-primary);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast), transform var(--transition-fast);
}

.module-card:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-card-hover);
  transform: translateY(-2px);
}

.module-index {
  margin: 0 0 20px;
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.18em;
  color: var(--accent);
}

.module-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  margin-bottom: 18px;
  color: var(--accent);
  background: var(--accent-light);
  border: 1px solid color-mix(in srgb, var(--accent) 40%, transparent);
}

.module-name {
  margin: 0 0 10px;
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
}

.module-desc {
  margin: 0;
  font-size: 13px;
  line-height: 1.75;
  color: var(--text-secondary);
}

/* ── 创作链路 ──────────────────────────── */
.pipeline {
  display: flex;
  align-items: stretch;
  gap: 12px;
}

.pipeline-step {
  flex: 1;
  position: relative;
  padding: 26px 22px;
  background: color-mix(in srgb, var(--bg-card) 82%, transparent);
  border: 1px solid var(--border-primary);
}

.pipeline-step::before {
  content: '';
  position: absolute;
  top: -1px;
  left: -1px;
  width: 14px;
  height: 14px;
  border-top: 2px solid var(--accent);
  border-left: 2px solid var(--accent);
}

.step-code {
  margin: 0 0 14px;
  font-family: var(--font-mono);
  font-size: 10.5px;
  letter-spacing: 0.16em;
  color: var(--accent);
}

.step-name {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.step-desc {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-secondary);
}

.pipeline-link {
  align-self: center;
  color: var(--accent);
  font-family: var(--font-mono);
  font-size: 15px;
  flex-shrink: 0;
}

/* ── 作品存档 ──────────────────────────── */
.archive-more {
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: color var(--transition-fast);
}

.archive-more:hover {
  color: var(--accent);
}

.archive-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.archive-card {
  border: 1px solid var(--border-primary);
  background: color-mix(in srgb, var(--bg-card) 82%, transparent);
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.archive-card:hover,
.archive-card--active {
  border-color: var(--accent);
  box-shadow: var(--shadow-card-hover);
}

.archive-cover {
  position: relative;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  border-bottom: 1px solid var(--border-primary);
  background:
    linear-gradient(var(--border-primary) 1px, transparent 1px) 0 0 / 100% 22px,
    linear-gradient(90deg, var(--border-primary) 1px, transparent 1px) 0 0 / 22px 100%,
    var(--bg-card);
  opacity: 0.9;
}

/* 红色取景框 X 占位 */
.cover-frame {
  position: absolute;
  inset: 16px;
  border: 1px solid color-mix(in srgb, var(--accent) 45%, transparent);
}

.cover-frame--a {
  transform: perspective(380px) rotateX(28deg) rotateY(-18deg);
}

.cover-frame--b {
  inset: 26px 40px;
  transform: perspective(380px) rotateX(-18deg) rotateY(24deg);
  opacity: 0.55;
}

.cover-id {
  position: absolute;
  top: 10px;
  left: 12px;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.14em;
  color: var(--text-muted);
}

.archive-card--active .cover-id {
  color: var(--accent);
}

.cover-ep {
  position: absolute;
  bottom: 10px;
  right: 12px;
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  padding: 2px 7px;
  color: var(--text-secondary);
  border: 1px solid var(--border-secondary);
}

.archive-info {
  padding: 16px 16px 18px;
}

.archive-name {
  margin: 0 0 8px;
  font-size: 15.5px;
  font-weight: 700;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.archive-desc {
  margin: 0 0 14px;
  font-size: 12.5px;
  line-height: 1.6;
  color: var(--text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 40px;
}

.archive-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.meta-chip {
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.1em;
  padding: 3px 8px;
  color: var(--accent);
  border: 1px solid color-mix(in srgb, var(--accent) 45%, transparent);
  background: var(--accent-light);
}

.meta-time {
  font-size: 11.5px;
  color: var(--text-muted);
}

/* ── CTA ───────────────────────────────── */
.cta-terminal {
  text-align: center;
  padding: 110px 24px 120px;
  border-top: 1px solid var(--border-primary);
}

.cta-prompt {
  margin: 0 0 14px;
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.14em;
  color: var(--success);
}

.prompt-arrow {
  color: var(--accent);
}

.cta-title {
  margin: 0 0 36px;
  font-size: clamp(28px, 4vw, 44px);
  font-weight: 800;
  color: var(--text-primary);
}

.cta-btn {
  min-width: 260px;
}

/* ── 响应式 ────────────────────────────── */
@media (max-width: 1024px) {
  .module-grid,
  .archive-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .hud {
    display: none;
  }
}

@media (max-width: 768px) {
  .hero {
    min-height: calc(100vh - 56px);
    padding: 40px 20px 64px;
  }

  .subtitle-br {
    display: none;
  }

  .hero-subtitle {
    font-size: 14.5px;
  }

  .hero-cta {
    flex-direction: column;
    align-items: stretch;
    max-width: 320px;
    margin: 0 auto;
  }

  .btn-terminal {
    min-width: 0;
    width: 100%;
  }

  .panel-section {
    padding: 64px 18px;
  }

  .section-title {
    font-size: 23px;
  }

  .title-en {
    display: none;
  }

  .pipeline {
    flex-direction: column;
  }

  .pipeline-link {
    transform: rotate(90deg);
    margin: -4px 0;
  }

  .module-grid,
  .archive-grid {
    grid-template-columns: 1fr;
  }
}
</style>
