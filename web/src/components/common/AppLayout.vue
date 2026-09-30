<template>
  <div class="app-layout">
    <AppHeader />
    <main class="app-main">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { AppHeader } from '@/components/common'
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  transition: background var(--transition-normal);
}

.app-main {
  flex: 1;
  padding-top: 44px;
}

@media (max-width: 768px) {
  .app-main {
    padding-top: 44px;
  }
}
</style>
