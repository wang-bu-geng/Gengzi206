<template>
  <teleport to="body">
    <transition name="slide-panel-fade">
      <div
        v-if="modelValue"
        class="slide-panel-overlay"
        @click.self="close"
      >
        <transition name="slide-panel-slide">
          <div
            v-if="modelValue"
            class="slide-panel"
            :style="{ width: computedWidth }"
            @keydown="handleKeydown"
            tabindex="0"
            ref="panelRef"
          >
            <div class="slide-panel-header">
              <h3 class="slide-panel-title">{{ title }}</h3>
              <el-button text class="slide-panel-close-btn" @click="close">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
            <div class="slide-panel-body">
              <slot />
            </div>
            <div class="slide-panel-footer">
              <slot name="footer">
                <el-button @click="close">{{ $t('common.close') }}</el-button>
              </slot>
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import { Close } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title?: string
  width?: number
}>(), {
  width: 480
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void
  (e: 'close'): void
  (e: 'save'): void
}>()

const panelRef = ref<HTMLElement | null>(null)

const computedWidth = computed(() => {
  if (typeof props.width === 'number') {
    return `${props.width}px`
  }
  return '480px'
})

function close() {
  emit('update:modelValue', false)
  emit('close')
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    close()
  }
  if (e.ctrlKey && e.key === 'Enter') {
    emit('save')
  }
}

// Auto-focus panel on open for keyboard events
watch(() => props.modelValue, async (val) => {
  if (val) {
    await nextTick()
    panelRef.value?.focus()
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

onUnmounted(() => {
  document.body.style.overflow = ''
})
</script>

<style scoped>
.slide-panel-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  justify-content: flex-end;
}

.slide-panel {
  height: 100%;
  background: var(--bg-card, #fff);
  display: flex;
  flex-direction: column;
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.12);
  outline: none;
  overflow: hidden;
}

.slide-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-primary, #e5e7eb);
  flex-shrink: 0;
}

.slide-panel-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.slide-panel-close-btn {
  font-size: 18px;
}

.slide-panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.slide-panel-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 20px;
  border-top: 1px solid var(--border-primary, #e5e7eb);
  flex-shrink: 0;
}

/* Transitions */
.slide-panel-fade-enter-active,
.slide-panel-fade-leave-active {
  transition: opacity 0.25s ease;
}
.slide-panel-fade-enter-from,
.slide-panel-fade-leave-to {
  opacity: 0;
}

.slide-panel-slide-enter-active {
  transition: transform 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
}
.slide-panel-slide-leave-active {
  transition: transform 0.2s ease-in;
}
.slide-panel-slide-enter-from,
.slide-panel-slide-leave-to {
  transform: translateX(100%);
}
</style>
