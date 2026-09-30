<template>
  <!-- Page header component with title, subtitle and action buttons -->
  <!-- 页面头部组件，包含标题、副标题和操作按钮 -->
  <header :class="['page-header', { 'with-back': showBack, 'with-border': showBorder }]">
    <div class="header-content">
      <!-- Back button section / 返回按钮区域 -->
      <div v-if="showBack" class="header-nav">
        <button class="back-btn" @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          <span>{{ backText }}</span>
        </button>
        <div class="nav-divider"></div>
      </div>

      <!-- Title section / 标题区域 -->
      <div class="header-text">
        <div class="title-row">
          <div v-if="$slots.icon" class="title-icon">
            <slot name="icon"></slot>
          </div>
          <h1 class="header-title">{{ title }}</h1>
          <slot name="badge"></slot>
        </div>
        <p v-if="subtitle" class="header-subtitle">{{ subtitle }}</p>
      </div>

      <!-- Actions section / 操作区域 -->
      <div class="header-actions">
        <slot name="actions"></slot>
      </div>
    </div>

    <!-- Extra content / 额外内容 -->
    <div v-if="$slots.extra" class="header-extra">
      <slot name="extra"></slot>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'

/**
 * PageHeader - Reusable page header component
 * 页面头部组件 - 可复用的页面头部
 */
const props = withDefaults(defineProps<{
  title: string
  subtitle?: string
  showBack?: boolean
  backText?: string
  showBorder?: boolean
}>(), {
  showBack: false,
  backText: '返回',
  showBorder: true
})

const emit = defineEmits<{
  back: []
}>()

const router = useRouter()

// Handle back navigation / 处理返回导航
const handleBack = () => {
  emit('back')
  router.back()
}
</script>

<style scoped>
.page-header {
  margin-bottom: var(--space-3);
}

.page-header.with-border {
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--border-primary);
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

@media (min-width: 768px) {
  .header-content {
    flex-direction: row;
    align-items: center;
  }
  
  .page-header.with-back .header-content {
    flex-wrap: nowrap;
  }
}

/* Navigation / 导航 */
.header-nav {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.875rem;
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.back-btn:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
  border-color: var(--border-secondary);
}

.back-btn:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.nav-divider {
  width: 1px;
  height: 2rem;
  background: var(--border-primary);
}

/* Title / 标题 */
.header-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.title-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  background: linear-gradient(135deg, var(--accent) 0%, #06b6d4 100%);
  border-radius: var(--radius-lg);
  color: white;
  flex-shrink: 0;
}

.header-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.025em;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: baseline;
  gap: 10px;
}

@media (min-width: 768px) {
  .header-title {
    font-size: 1.75rem;
  }
}

.header-subtitle {
  margin: 0;
  font-size: 0.875rem;
  color: var(--text-muted);
  font-weight: 500;
  max-width: 480px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Actions / 操作 */
.header-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  align-items: center;
  flex-shrink: 0;
}

@media (min-width: 768px) {
  .header-actions {
    margin-left: auto;
  }
}

/* Extra / 额外内容 */
.header-extra {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-primary);
}
</style>
