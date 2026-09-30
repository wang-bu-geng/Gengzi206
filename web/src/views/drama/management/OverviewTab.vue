<template>
  <div>
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <StatCard
        :label="$t('drama.management.episodeStats')"
        :value="episodesCount"
        :icon="Document"
        icon-color="var(--accent)"
        icon-bg="var(--accent-light)"
        value-color="var(--accent)"
        :description="$t('drama.management.episodesCreated')"
      />
      <StatCard
        :label="$t('drama.management.characterStats')"
        :value="charactersCount"
        :icon="User"
        icon-color="var(--success)"
        icon-bg="var(--success-light)"
        value-color="var(--success)"
        :description="$t('drama.management.charactersCreated')"
      />
      <StatCard
        :label="$t('drama.management.sceneStats')"
        :value="scenesCount"
        :icon="Picture"
        icon-color="var(--warning)"
        icon-bg="var(--warning-light)"
        value-color="var(--warning)"
        :description="$t('drama.management.sceneLibraryCount')"
      />
      <StatCard
        :label="$t('drama.management.propStats')"
        :value="propsCount"
        :icon="Box"
        icon-color="var(--primary)"
        icon-bg="var(--primary-light)"
        value-color="var(--primary)"
        :description="$t('drama.management.propsCreated')"
      />
    </div>

    <!-- 引导卡片：无章节时显示 -->
    <el-alert
      v-if="episodesCount === 0"
      :title="$t('drama.management.startFirstEpisode')"
      type="info"
      :closable="false"
      style="margin-top: 20px"
    >
      <template #default>
        <p style="margin: 8px 0">
          {{ $t('drama.management.noEpisodesYet') }}
        </p>
        <el-button
          type="primary"
          :icon="Plus"
          @click="createNewEpisode"
          style="margin-top: 8px"
        >
          {{ $t('drama.management.createFirstEpisode') }}
        </el-button>
      </template>
    </el-alert>

    <!-- 项目信息 -->
    <el-card shadow="never" class="project-info-card">
      <template #header>
        <div class="card-header">
          <h3 class="card-title">
            {{ $t('drama.management.projectInfo') }}
          </h3>
          <div class="card-header-tags">
            <el-tag size="small" type="warning">专业模式</el-tag>
            <el-tag :type="getStatusType(drama?.status)" size="small">{{
              getStatusText(drama?.status)
            }}</el-tag>
          </div>
        </div>
      </template>
      <el-descriptions :column="2" border class="project-descriptions">
        <el-descriptions-item :label="$t('drama.management.projectName')">
          <span class="info-value">{{ drama?.title }}</span>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('common.createdAt')">
          <span class="info-value">{{ formatDate(drama?.created_at) }}</span>
        </el-descriptions-item>
        <el-descriptions-item
          :label="$t('drama.management.projectDesc')"
          :span="2"
        >
          <span class="info-desc">{{
            drama?.description || $t('drama.management.noDescription')
          }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Document, User, Picture, Plus, Box } from '@element-plus/icons-vue'
import { StatCard } from '@/components/common'
import { useDramaContext } from '@/composables/useDramaContext'

const router = useRouter()
const { dramaData, dramaId, scenes } = useDramaContext()

const drama = computed(() => dramaData.value)
const episodesCount = computed(() => drama.value?.episodes?.length || 0)
const charactersCount = computed(() => drama.value?.characters?.length || 0)
const scenesCount = computed(() => scenes.value.length)
const propsCount = computed(() => drama.value?.props?.length || 0)

function getStatusType(status?: string) {
  const map: Record<string, any> = {
    draft: 'info',
    in_progress: 'warning',
    completed: 'success',
  }
  return map[status || 'draft'] || 'info'
}

function getStatusText(status?: string) {
  const map: Record<string, string> = {
    draft: '草稿',
    in_progress: '制作中',
    completed: '已完成',
  }
  return map[status || 'draft'] || '草稿'
}

function formatDate(date?: string) {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

function createNewEpisode() {
  const nextEpisodeNumber = episodesCount.value + 1
  router.push({
    name: 'EpisodeWorkflowNew',
    params: { id: dramaId, episodeNumber: nextEpisodeNumber },
  })
}
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: var(--space-2);
  margin-bottom: var(--space-3);
}

@media (min-width: 640px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-3);
  }
}

@media (min-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.project-info-card {
  margin-top: var(--space-5);
  border-radius: var(--radius-lg);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header-tags {
  display: flex;
  gap: 8px;
}

.card-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.project-descriptions {
  width: 100%;
}

.info-value {
  font-weight: 500;
  color: var(--text-primary);
}

.info-desc {
  color: var(--text-secondary);
  line-height: 1.6;
}
</style>
