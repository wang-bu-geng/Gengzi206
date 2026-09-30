<template>
  <!-- Minimalist action button with icon and optional tooltip -->
  <!-- 简约操作按钮，带图标和可选提示 -->
  <el-tooltip
    v-if="tooltip"
    :content="tooltip"
    placement="top"
    :show-after="500"
  >
    <button
      :class="['action-button', variant, { disabled }]"
      :disabled="disabled"
      @click="$emit('click')"
    >
      <el-icon :size="size">
        <component :is="icon" />
      </el-icon>
    </button>
  </el-tooltip>
  <button
    v-else
    :class="['action-button', variant, { disabled }]"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <el-icon :size="size">
      <component :is="icon" />
    </el-icon>
  </button>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

/**
 * ActionButton - Minimalist icon button for actions
 * 操作按钮 - 简约图标按钮用于各种操作
 */
withDefaults(defineProps<{
  icon: Component
  tooltip?: string
  variant?: 'default' | 'primary' | 'danger'
  size?: number
  disabled?: boolean
}>(), {
  variant: 'default',
  size: 16,
  disabled: false
})

defineEmits<{
  click: []
}>()
</script>

<style scoped>
.action-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-button:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}

.action-button:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}

.action-button.primary:hover {
  background: var(--accent-light);
  color: var(--accent);
}

.action-button.danger:hover {
  background: #fef2f2;
  color: #ef4444;
}

.dark .action-button.danger:hover {
  background: rgba(239, 68, 68, 0.15);
}

.action-button.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-button.disabled:hover {
  background: transparent;
  color: var(--text-muted);
}
</style>
