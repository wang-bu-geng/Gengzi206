<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <!-- Page Sub Header -->
      <div class="page-sub-header">
        <el-button text @click="$router.back()" class="back-btn">
          <el-icon><ArrowLeft /></el-icon>
          <span>{{ $t('common.back') }}</span>
        </el-button>
        <div class="page-title">
          <h1>{{ drama?.title || '' }}</h1>
          <span class="subtitle">{{
            drama?.description || $t('drama.management.overview')
          }}</span>
        </div>
      </div>

      <!-- Tab Navigation -->
      <div class="tabs-wrapper">
        <div class="tab-nav">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="tab-btn"
            :class="{ active: currentTab === tab.key }"
            @click="switchTab(tab.key)"
          >
            {{ tab.label }}
          </button>
        </div>

        <!-- Tab Content -->
        <div class="tab-content">
          <router-view v-slot="{ Component }">
            <keep-alive>
              <component :is="Component" />
            </keep-alive>
          </router-view>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, provide, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { dramaAPI } from '@/api/drama'
import type { Drama } from '@/types/drama'
import { DRAMA_CONTEXT_KEY } from '@/composables/useDramaContext'

const router = useRouter()
const route = useRoute()

const drama = ref<Drama>()
const scenes = ref<any[]>([])

const tabs = [
  { key: 'overview', label: '概览' },
  { key: 'episodes', label: '章节管理' },
  { key: 'characters', label: '角色管理' },
  { key: 'scenes', label: '场景库' },
  { key: 'props', label: '道具管理' },
]

const LAST_TAB_KEY = 'drama_last_tab'

const currentTab = computed(() => {
  const path = route.path
  for (const tab of tabs) {
    if (path.endsWith(`/${tab.key}`)) return tab.key
  }
  return 'overview'
})

function switchTab(key: string) {
  if (currentTab.value === key) return
  localStorage.setItem(LAST_TAB_KEY, key)
  router.push({
    name: `DramaManagement-${key.charAt(0).toUpperCase() + key.slice(1)}`,
    params: { id: route.params.id },
  })
}

async function loadDramaData() {
  try {
    const data = await dramaAPI.get(route.params.id as string)
    drama.value = data
    loadScenes()
  } catch (error: any) {
    ElMessage.error(error.message || '加载项目数据失败')
  }
}

function loadScenes() {
  if (drama.value?.scenes) {
    scenes.value = drama.value.scenes
  } else {
    scenes.value = []
  }
}

provide(DRAMA_CONTEXT_KEY, {
  dramaId: route.params.id as string,
  dramaData: drama as any,
  loadDramaData,
  scenes: scenes as any,
  loadScenes,
})

onMounted(async () => {
  await loadDramaData()
  const lastTab = localStorage.getItem(LAST_TAB_KEY)
  const validTabs = tabs.map((t) => t.key)
  const path = route.path
  const onTab = validTabs.some((t) => path.endsWith(`/${t}`))
  if (!onTab && lastTab && validTabs.includes(lastTab)) {
    router.replace({
      name: `DramaManagement-${lastTab.charAt(0).toUpperCase() + lastTab.slice(1)}`,
      params: { id: route.params.id },
    })
  }
})
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: var(--bg-primary);
  transition: background var(--transition-normal);
}

.content-wrapper {
  margin: 0 auto;
  width: 100%;
}

.tabs-wrapper {
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  overflow: hidden;
}

.tab-nav {
  display: flex;
  border-bottom: 2px solid var(--border-primary);
  background: var(--bg-secondary);
  padding: 0 var(--space-4);
  gap: 0;
}

.tab-btn {
  position: relative;
  padding: 14px 20px;
  border: none;
  background: transparent;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: color var(--transition-normal);
  white-space: nowrap;
}

.tab-btn:hover {
  color: var(--text-primary);
}

.tab-btn.active {
  color: var(--accent);
  font-weight: 600;
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--accent);
}

.tab-content {
  padding: var(--space-4);
}

@media (min-width: 768px) {
  .tab-content {
    padding: var(--space-5);
  }
}

.page-sub-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4) 0;
}

.back-btn {
  margin-right: var(--space-2);
}

.page-title {
  display: flex;
  flex-direction: column;
}

.page-title h1 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.3;
}

.subtitle {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin-top: 2px;
}
</style>
