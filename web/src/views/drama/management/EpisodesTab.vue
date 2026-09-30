<template>
  <div>
    <div class="tab-header">
      <h2>{{ $t('drama.management.episodeList') }}</h2>
      <div class="tab-header-actions">
        <el-button
          v-if="selectedEpisodes.length > 0"
          type="danger"
          @click="batchDeleteEpisodes"
        >
          批量删除 ({{ selectedEpisodes.length }})
        </el-button>
        <el-button type="primary" :icon="Plus" @click="createNewEpisode">
          {{ $t('drama.management.createNewEpisode') }}
        </el-button>
      </div>
    </div>

    <!-- 空状态 -->
    <el-empty
      v-if="episodesCount === 0"
      :description="$t('drama.management.noEpisodes')"
      style="margin-top: 40px"
    >
      <template #image>
        <el-icon :size="80" class="empty-icon"><Document /></el-icon>
      </template>
      <el-button type="primary" :icon="Plus" @click="createNewEpisode">
        {{ $t('drama.management.createFirstEpisode') }}
      </el-button>
    </el-empty>

    <!-- 章节表格 -->
    <el-table
      v-else
      ref="tableRef"
      :data="sortedEpisodes"
      border
      stripe
      style="margin-top: 16px"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column
        type="index"
        :label="$t('storyboard.table.number')"
        width="80"
      />
      <el-table-column
        prop="title"
        :label="$t('drama.management.episodeList')"
        min-width="200"
      />
      <el-table-column :label="$t('common.status')" width="120">
        <template #default="{ row }">
          <el-tag :type="getEpisodeStatusType(row)">{{
            getEpisodeStatusText(row)
          }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Shots" width="100">
        <template #default="{ row }">
          {{ row.shots?.length || 0 }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('common.createdAt')" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column
        :label="$t('storyboard.table.operations')"
        width="220"
        fixed="right"
      >
        <template #default="{ row }">
          <el-button
            size="small"
            type="primary"
            @click="enterEpisodeWorkflow(row)"
          >
            {{ $t('drama.management.goToEdit') }}
          </el-button>
          <el-button
            size="small"
            type="danger"
            @click="deleteEpisode(row)"
          >
            {{ $t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Plus } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { useDramaContext } from '@/composables/useDramaContext'

const router = useRouter()
const { dramaData, dramaId, loadDramaData } = useDramaContext()

const tableRef = ref()
const selectedEpisodes = ref<any[]>([])

const episodesCount = computed(() => dramaData.value?.episodes?.length || 0)

const sortedEpisodes = computed(() => {
  if (!dramaData.value?.episodes) return []
  return [...dramaData.value.episodes].sort(
    (a, b) => a.episode_number - b.episode_number
  )
})

function handleSelectionChange(rows: any[]) {
  selectedEpisodes.value = rows
}

function getEpisodeStatusType(episode: any) {
  if (episode.shots && episode.shots.length > 0) return 'success'
  if (episode.script_content) return 'warning'
  return 'info'
}

function getEpisodeStatusText(episode: any) {
  if (episode.shots && episode.shots.length > 0) return '已拆分'
  if (episode.script_content) return '已创建'
  return '草稿'
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

function enterEpisodeWorkflow(episode: any) {
  router.push({
    name: 'EpisodeWorkflowNew',
    params: { id: dramaId, episodeNumber: episode.episode_number },
  })
}

async function deleteEpisode(episode: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除第${episode.episode_number}章吗？此操作将同时删除该章节的所有相关数据（角色、场景、分镜等）。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const existingEpisodes = dramaData.value?.episodes || []
    const updatedEpisodes = existingEpisodes
      .filter((ep) => ep.episode_number !== episode.episode_number)
      .map((ep) => ({
        episode_number: ep.episode_number,
        title: ep.title,
        script_content: ep.script_content,
        description: ep.description,
        duration: ep.duration,
        status: ep.status,
      }))

    await dramaAPI.saveEpisodes(dramaData.value!.id, updatedEpisodes)
    ElMessage.success(`第${episode.episode_number}章删除成功`)
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function batchDeleteEpisodes() {
  if (selectedEpisodes.value.length === 0) return

  const titles = selectedEpisodes.value.map((e) => `第${e.episode_number}章`).join('、')
  try {
    await ElMessageBox.confirm(
      `确定要批量删除以下章节吗？\n${titles}\n\n此操作将同时删除这些章节的所有相关数据。`,
      '批量删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const existingEpisodes = dramaData.value?.episodes || []
    const deleteNumbers = new Set(selectedEpisodes.value.map((e) => e.episode_number))
    const updatedEpisodes = existingEpisodes
      .filter((ep) => !deleteNumbers.has(ep.episode_number))
      .map((ep) => ({
        episode_number: ep.episode_number,
        title: ep.title,
        script_content: ep.script_content,
        description: ep.description,
        duration: ep.duration,
        status: ep.status,
      }))

    await dramaAPI.saveEpisodes(dramaData.value!.id, updatedEpisodes)
    ElMessage.success(`已删除 ${selectedEpisodes.value.length} 个章节`)
    selectedEpisodes.value = []
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '批量删除失败')
    }
  }
}
</script>

<style scoped>
.tab-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

@media (min-width: 640px) {
  .tab-header {
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
}

.tab-header h2 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.tab-header-actions {
  display: flex;
  gap: 10px;
}

.empty-icon {
  color: var(--accent);
}
</style>
