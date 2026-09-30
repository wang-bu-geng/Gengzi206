<template>
  <div class="app-layout">
    <!-- 全局持久 3D 背景与 CRT 层，路由切换不销毁 -->
    <GlobalBackground />
    <CrtOverlay />
    <!-- 全局生成终端：视频/分镜/图片所有生成流程共用 -->
    <GenerationTerminal />
    <AppHeader :fixed="true" />
    <main class="app-main">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { AppHeader } from '@/components/common'
import CrtOverlay from '@/components/common/CrtOverlay.vue'
import GenerationTerminal from '@/views/drama/components/GenerationTerminal.vue'

// 3D 背景异步分包加载，不阻塞首屏外壳渲染
const GlobalBackground = defineAsyncComponent(
  () => import('@/components/common/GlobalBackground.vue')
)
</script>

<style>
#app {
  width: 100%;
  min-height: 100vh;
}

.app-layout {
  position: relative;
  min-height: 100vh;
  /* 透明，露出全局 Three.js 背景；底色兜底在 body */
  background: transparent;
  transition: background var(--transition-normal);
}

.app-main {
  position: relative;
  z-index: 1;
  padding-top: 64px;
  min-height: calc(100vh - 64px);
}

@media (max-width: 768px) {
  .app-main {
    padding-top: 56px;
    min-height: calc(100vh - 56px);
  }
}
</style>
