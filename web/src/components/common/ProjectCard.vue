<template>
  <article
    class="project-card"
    @click="$emit('click')"
    tabindex="0"
    @keydown.enter="$emit('click')"
  >
    <!-- 封面：取景框占位（暂无真实封面图，统一终端网格 + 四角框） -->
    <div class="card-cover">
      <div class="cover-grid" />
      <span class="cover-corner cover-corner--tl" />
      <span class="cover-corner cover-corner--tr" />
      <span class="cover-corner cover-corner--bl" />
      <span class="cover-corner cover-corner--br" />
      <span class="cover-scan" />

      <div class="cover-content">
        <el-icon class="cover-icon"><VideoCamera /></el-icon>
      </div>

      <span class="cover-style">STYLE: {{ styleCode }}</span>
      <div class="cover-badges">
        <span class="badge episode-badge">
          <el-icon><Film /></el-icon>
          EPISODE {{ episodeCount }}
        </span>
      </div>
      <div class="cover-overlay">
        <slot name="actions"></slot>
      </div>
    </div>

    <!-- 内容区 -->
    <div class="card-body">
      <h3 class="card-title">{{ title }}</h3>
      <p v-if="description" class="card-description">{{ description }}</p>

      <div class="card-footer">
        <span class="meta-style">{{ styleLabel }}</span>
        <span class="meta-time">
          <el-icon><Clock /></el-icon>
          {{ formattedDate }}
        </span>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Film, VideoCamera, Clock } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  title: string
  description?: string
  updatedAt: string
  episodeCount?: number
  style?: string
}>(), {
  description: '',
  episodeCount: 0,
  style: ''
})

defineEmits<{
  click: []
}>()

// 风格只作标识标签，不再映射彩色渐变封面
const styleCodes: Record<string, string> = {
  ghibli: 'GHIBLI',
  guoman: 'GUOMAN',
  wasteland: 'WASTELAND',
  nostalgia: 'NOSTALGIA',
  pixel: 'PIXEL_8BIT',
  voxel: 'VOXEL_3D',
  urban: 'URBAN',
  guoman3d: 'GUOMAN_3D',
  chibi3d: 'CHIBI_3D',
  custom: 'CUSTOM'
}

const styleLabels: Record<string, string> = {
  ghibli: '吉卜力',
  guoman: '国漫',
  wasteland: '废土',
  nostalgia: '复古',
  pixel: '像素',
  voxel: '体素',
  urban: '都市',
  guoman3d: '3D 国漫',
  chibi3d: 'Q 版 3D',
  custom: '自定义'
}

const styleCode = computed(() => styleCodes[props.style] || 'UNSPECIFIED')
const styleLabel = computed(() => styleLabels[props.style] || '未指定画风')

const formattedDate = computed(() => {
  const date = new Date(props.updatedAt)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days} 天前`
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`
})
</script>

<style scoped>
.project-card {
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  transition:
    border-color var(--transition-fast),
    box-shadow var(--transition-fast);
}

.project-card:hover {
  border-color: var(--accent);
  box-shadow: var(--shadow-card-hover);
}

.project-card:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

/* ── 封面 ───────────────────────────────── */
.card-cover {
  position: relative;
  height: 168px;
  overflow: hidden;
  background: var(--bg-secondary);
}

/* 终端网格 */
.cover-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(var(--border-primary) 1px, transparent 1px),
    linear-gradient(90deg, var(--border-primary) 1px, transparent 1px);
  background-size: 22px 22px;
  opacity: 0.55;
}

/* 四角取景框 */
.cover-corner {
  position: absolute;
  width: 14px;
  height: 14px;
  border-color: color-mix(in srgb, var(--accent) 60%, transparent);
  border-style: solid;
  border-width: 0;
  transition: border-color var(--transition-fast);
}

.cover-corner--tl { top: 8px; left: 8px; border-top-width: 1px; border-left-width: 1px; }
.cover-corner--tr { top: 8px; right: 8px; border-top-width: 1px; border-right-width: 1px; }
.cover-corner--bl { bottom: 8px; left: 8px; border-bottom-width: 1px; border-left-width: 1px; }
.cover-corner--br { bottom: 8px; right: 8px; border-bottom-width: 1px; border-right-width: 1px; }

.project-card:hover .cover-corner {
  border-color: var(--accent);
}

/* hover 红色扫描线 */
.cover-scan {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 2px;
  background: var(--accent);
  opacity: 0;
  box-shadow: 0 0 10px var(--accent);
  transform: translateY(0);
}

.project-card:hover .cover-scan {
  opacity: 0.9;
  animation: cover-scan-y 1.6s linear infinite;
}

@keyframes cover-scan-y {
  0% { transform: translateY(0); }
  100% { transform: translateY(168px); }
}

.cover-content {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-icon {
  font-size: 44px;
  color: color-mix(in srgb, var(--text-muted) 80%, transparent);
  transition: color var(--transition-fast);
}

.project-card:hover .cover-icon {
  color: var(--accent);
}

.cover-style {
  position: absolute;
  top: 12px;
  left: 12px;
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}

.cover-badges {
  position: absolute;
  bottom: 10px;
  left: 10px;
  display: flex;
  gap: 6px;
  z-index: 2;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  background: color-mix(in srgb, var(--bg-primary) 72%, transparent);
  color: var(--text-secondary);
  border: 1px solid var(--border-secondary);
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.08em;
  border-radius: var(--radius-sm);
}

.badge .el-icon {
  font-size: 0.75rem;
}

.cover-overlay {
  position: absolute;
  top: 0;
  right: 0;
  padding: 8px;
  display: flex;
  gap: 6px;
  opacity: 0;
  transition: opacity var(--transition-fast);
  z-index: 3;
}

.project-card:hover .cover-overlay {
  opacity: 1;
}

/* ── 内容区 ─────────────────────────────── */
.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 14px 16px 16px;
  gap: 8px;
}

.card-title {
  margin: 0;
  font-size: 15px;
  font-weight: 800;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}

.card-description {
  margin: 0;
  font-size: 12.5px;
  color: var(--text-secondary);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.6;
  min-height: 40px;
}

.card-footer {
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px solid var(--border-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.meta-style {
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  padding: 2px 7px;
  color: var(--accent);
  border: 1px solid color-mix(in srgb, var(--accent) 45%, transparent);
  background: var(--accent-light);
  border-radius: var(--radius-sm);
  white-space: nowrap;
}

.meta-time {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11.5px;
  color: var(--text-muted);
  white-space: nowrap;
}

.meta-time .el-icon {
  font-size: 0.8rem;
}

/* 悬停操作钮：深色描边方钮 */
:deep(.action-button) {
  width: 28px !important;
  height: 28px !important;
  padding: 0 !important;
  background: color-mix(in srgb, var(--bg-primary) 82%, transparent) !important;
  border: 1px solid var(--border-secondary) !important;
  border-radius: var(--radius-sm) !important;
  color: var(--text-secondary) !important;
}

:deep(.action-button:hover) {
  border-color: var(--accent) !important;
  color: var(--accent) !important;
  background: var(--accent-light) !important;
}

@media (prefers-reduced-motion: reduce) {
  .project-card:hover .cover-scan {
    animation: none;
  }
}
</style>
