import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/dashboard/Dashboard.vue')
  },
  {
    path: '/dramas',
    name: 'DramaList',
    component: () => import('../views/drama/DramaList.vue')
  },
  {
    path: '/dramas/create',
    name: 'DramaCreate',
    component: () => import('../views/drama/DramaCreate.vue')
  },
  // Episode workflow — must be registered BEFORE /dramas/:id nested routes
  {
    path: '/dramas/:id/episode/:episodeNumber',
    name: 'EpisodeWorkflowNew',
    component: () => import('../views/drama/EpisodeWorkflow.vue')
  },
  {
    path: '/dramas/:id/settings',
    name: 'DramaSettings',
    component: () => import('../views/workflow/DramaSettings.vue')
  },
  // ================================================================
  // REMOVED (absorbed into DramaManagement tabs):
  //   /dramas/:id/characters  → CharacterExtraction
  //   /dramas/:id/images/characters → CharacterImages
  // CharactersTab & ScenesTab now include CRUD + extract inline.
  // ================================================================
  {
    path: '/dramas/:id',
    component: () => import('../views/drama/DramaManagement.vue'),
    children: [
      {
        path: '',
        redirect: { name: 'DramaManagement-Overview' }
      },
      {
        path: 'overview',
        name: 'DramaManagement-Overview',
        component: () => import('../views/drama/management/OverviewTab.vue')
      },
      {
        path: 'episodes',
        name: 'DramaManagement-Episodes',
        component: () => import('../views/drama/management/EpisodesTab.vue')
      },
      {
        path: 'characters',
        name: 'DramaManagement-Characters',
        component: () => import('../views/drama/management/CharactersTab.vue')
      },
      {
        path: 'scenes',
        name: 'DramaManagement-Scenes',
        component: () => import('../views/drama/management/ScenesTab.vue')
      },
      {
        path: 'props',
        name: 'DramaManagement-Props',
        component: () => import('../views/drama/management/PropsTab.vue')
      }
    ]
  },
  {
    path: '/settings',
    redirect: '/settings/ai-config'
  },
  {
    path: '/settings/ai-config',
    name: 'AIConfig',
    component: () => import('../views/settings/AIConfig.vue')
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 开源版本 - 无需认证

export default router
