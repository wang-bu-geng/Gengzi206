<template>
  <!-- Stat Card Component - Display statistics with modern design -->
  <!-- 统计卡片组件 - 现代设计的统计数据展示 -->
  <div :class="['stat-card', `variant-${variant}`]">
    <div class="stat-icon" :style="{ background: iconBg }">
      <el-icon :size="24" :color="iconColor">
        <component :is="icon" />
      </el-icon>
    </div>
    <div class="stat-content">
      <span class="stat-label">{{ label }}</span>
      <div class="stat-value-row">
        <span class="stat-value" :style="{ color: valueColor }">{{ formattedValue }}</span>
        <span v-if="suffix" class="stat-suffix">{{ suffix }}</span>
      </div>
      <span v-if="description" class="stat-description">{{ description }}</span>
    </div>
    <div v-if="trend !== undefined" :class="['stat-trend', trendDirection]">
      <el-icon :size="14">
        <component :is="trendIcon" />
      </el-icon>
      <span>{{ Math.abs(trend) }}%</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import { TrendCharts, ArrowUp, ArrowDown } from '@element-plus/icons-vue'

/**
 * StatCard - Display statistics with trend indicator
 * 统计卡片 - 带趋势指示器的统计数据展示
 */
const props = withDefaults(defineProps<{
  label: string
  value: number | string
  icon: Component
  iconColor?: string
  iconBg?: string
  valueColor?: string
  suffix?: string
  description?: string
  trend?: number
  variant?: 'default' | 'compact'
}>(), {
  iconColor: 'var(--accent)',
  iconBg: 'var(--accent-light)',
  valueColor: 'var(--accent)',
  variant: 'default'
})

// Format large numbers / 格式化大数字
const formattedValue = computed(() => {
  if (typeof props.value === 'string') return props.value
  if (props.value >= 1000000) {
    return (props.value / 1000000).toFixed(1) + 'M'
  }
  if (props.value >= 1000) {
    return (props.value / 1000).toFixed(1) + 'K'
  }
  return props.value.toString()
})

// Trend direction / 趋势方向
const trendDirection = computed(() => {
  if (props.trend === undefined) return ''
  return props.trend >= 0 ? 'up' : 'down'
})

// Trend icon / 趋势图标
const trendIcon = computed(() => {
  if (props.trend === undefined) return TrendCharts
  return props.trend >= 0 ? ArrowUp : ArrowDown
})
</script>

<style scoped>
.stat-card {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3);
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  transition: all var(--transition-normal);
}

.stat-card:hover {
  border-color: var(--border-secondary);
  box-shadow: var(--shadow-card-hover);
}

/* Compact variant / 紧凑变体 */
.variant-compact {
  padding: var(--space-2);
  gap: var(--space-2);
}

.variant-compact .stat-icon {
  width: 2.5rem;
  height: 2.5rem;
}

.variant-compact .stat-value {
  font-size: 1.5rem;
}

/* Icon / 图标 */
.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  border-radius: var(--radius-lg);
  flex-shrink: 0;
}

/* Content / 内容 */
.stat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.stat-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-muted);
}

.stat-value-row {
  display: flex;
  align-items: baseline;
  gap: var(--space-1);
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.stat-suffix {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-muted);
}

.stat-description {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: var(--space-1);
}

/* Trend / 趋势 */
.stat-trend {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-md);
  font-size: 0.75rem;
  font-weight: 600;
}

.stat-trend.up {
  background: var(--success-light);
  color: var(--success);
}

.stat-trend.down {
  background: var(--error-light);
  color: var(--error);
}
</style>
