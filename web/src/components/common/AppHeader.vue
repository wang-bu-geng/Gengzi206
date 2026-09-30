<template>
  <header class="cyber-nav" :class="{ scrolled: isScrolled }">
    <div class="nav-content">
      <!-- Wordmark：红方块 + 更子 GENGZI -->
      <router-link to="/" class="nav-logo">
        <span class="logo-mark"></span>
        <span class="logo-cn">更子</span>
        <span class="logo-en">GENGZI</span>
      </router-link>

      <!-- 导航：中文 + 全大写英文混排 -->
      <nav class="nav-links">
        <router-link to="/">
          <span class="nav-zh">首页</span>
          <span class="nav-en">INDEX</span>
        </router-link>
        <router-link to="/dramas">
          <span class="nav-zh">我的作品</span>
          <span class="nav-en">WORKS</span>
        </router-link>
        <router-link to="/settings">
          <span class="nav-zh">设置</span>
          <span class="nav-en">SETTINGS</span>
        </router-link>
      </nav>

      <!-- 右侧：主题翻转 + 红色主行动 -->
      <div class="nav-actions">
        <ThemeToggle />
        <router-link to="/dramas/create" class="nav-create">
          <span>创建作品</span>
          <span class="nav-create-en">CREATE_01</span>
        </router-link>
      </div>
    </div>
    <!-- 红色扫描底线 -->
    <div class="nav-scanline" />
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import ThemeToggle from './ThemeToggle.vue'

const emit = defineEmits<{
  (e: 'open-ai-config'): void
  (e: 'config-updated'): void
}>()

const isScrolled = ref(false)

const handleScroll = () => {
  isScrolled.value = window.scrollY > 0
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})

defineExpose({
  openAIConfig: () => {
    emit('open-ai-config')
  }
})
</script>

<style scoped>
.cyber-nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  z-index: 9999;
  background: rgba(7, 7, 9, 0.72);
  backdrop-filter: blur(14px) saturate(140%);
  -webkit-backdrop-filter: blur(14px) saturate(140%);
  border-bottom: 1px solid rgba(245, 245, 247, 0.06);
  transition: background 0.3s ease, border-color 0.3s ease;
}

.cyber-nav.scrolled {
  background: rgba(5, 5, 7, 0.88);
  border-bottom-color: rgba(245, 245, 247, 0.1);
}

/* 红色扫描底线 */
.nav-scanline {
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 45, 45, 0.35) 50%,
    transparent 100%
  );
  opacity: 0.5;
  transition: opacity 0.3s ease;
}

.cyber-nav.scrolled .nav-scanline {
  opacity: 1;
}

.nav-content {
  max-width: 1440px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 28px;
}

/* ── Wordmark ─────────────────────────── */
.nav-logo {
  display: flex;
  align-items: baseline;
  gap: 9px;
  text-decoration: none;
  flex-shrink: 0;
}

.logo-mark {
  align-self: center;
  width: 11px;
  height: 11px;
  background: var(--accent);
  box-shadow: 0 0 10px rgba(255, 45, 45, 0.65);
  transition: transform var(--transition-fast);
}

.nav-logo:hover .logo-mark {
  transform: scale(1.18);
}

.logo-cn {
  font-size: 19px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: 0.02em;
  line-height: 1;
}

.logo-en {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.24em;
  line-height: 1;
}

/* ── 导航 ─────────────────────────────── */
.nav-links {
  display: flex;
  align-items: center;
  gap: 40px;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
}

.nav-links a {
  position: relative;
  display: flex;
  align-items: baseline;
  gap: 6px;
  text-decoration: none;
  padding: 4px 0;
  color: var(--text-secondary);
  transition: color var(--transition-fast);
}

.nav-zh {
  font-size: 13px;
  font-weight: 500;
  line-height: 1;
}

.nav-en {
  font-family: var(--font-mono);
  font-size: 9.5px;
  font-weight: 500;
  letter-spacing: 0.16em;
  line-height: 1;
  color: var(--text-muted);
  transition: color var(--transition-fast);
}

.nav-links a:hover {
  color: var(--text-primary);
}

.nav-links a.router-link-exact-active {
  color: var(--accent);
}

.nav-links a.router-link-exact-active .nav-en {
  color: var(--accent);
}

.nav-links a.router-link-exact-active::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -19px;
  height: 2px;
  background: var(--accent);
  box-shadow: 0 0 8px rgba(255, 45, 45, 0.7);
}

/* ── 右侧操作 ─────────────────────────── */
.nav-actions {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-shrink: 0;
}

.nav-create {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding: 0 16px;
  background: var(--accent);
  color: #ffffff;
  font-size: 12.5px;
  font-weight: 600;
  text-decoration: none;
  letter-spacing: 0.02em;
  border: 1px solid var(--accent);
  transition: background var(--transition-fast), box-shadow var(--transition-fast);
}

.nav-create-en {
  font-family: var(--font-mono);
  font-size: 9.5px;
  font-weight: 700;
  letter-spacing: 0.14em;
  opacity: 0.85;
}

.nav-create:hover {
  background: var(--accent-hover);
  border-color: var(--accent-hover);
  color: #ffffff;
  box-shadow: 0 0 18px -2px rgba(255, 45, 45, 0.6);
}

/* ── 浅色翻转 ─────────────────────────── */
:root:not(.dark) .cyber-nav {
  background: rgba(255, 255, 255, 0.78);
  border-bottom-color: rgba(11, 11, 12, 0.08);
}

:root:not(.dark) .cyber-nav.scrolled {
  background: rgba(255, 255, 255, 0.92);
  border-bottom-color: rgba(11, 11, 12, 0.12);
}

:root:not(.dark) .nav-links a.router-link-exact-active::after {
  box-shadow: 0 0 8px rgba(212, 0, 0, 0.45);
}

/* ── 响应式 ───────────────────────────── */
@media (max-width: 768px) {
  .cyber-nav {
    height: 52px;
  }

  .nav-content {
    padding: 0 16px;
  }

  .logo-en {
    display: none;
  }

  .nav-links {
    gap: 22px;
  }

  .nav-en {
    display: none;
  }

  .nav-links a.router-link-exact-active::after {
    bottom: -16px;
  }

  .nav-create {
    height: 30px;
    padding: 0 12px;
  }

  .nav-create-en {
    display: none;
  }
}

@media (max-width: 420px) {
  .nav-links {
    gap: 16px;
  }
}
</style>
