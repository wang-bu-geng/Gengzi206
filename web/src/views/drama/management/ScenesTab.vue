<template>
  <div>
    <div class="tab-header">
      <h2>{{ $t('drama.management.sceneList') }}</h2>
      <div class="tab-header-actions">
        <template v-if="selectedIds.length > 0">
          <el-button type="danger" @click="batchDelete">批量删除 ({{ selectedIds.length }})</el-button>
          <el-button type="success" @click="batchGenerateImages">批量生成图片 ({{ selectedIds.length }})</el-button>
        </template>
        <el-button :icon="Document" @click="extractPanelVisible = true">{{ $t('prop.extract') }}</el-button>
        <el-button type="primary" :icon="Plus" @click="openAdd">{{ $t('common.add') }}</el-button>
      </div>
    </div>

    <div class="select-mode-bar" v-if="scenesList.length > 0">
      <el-checkbox v-model="selectMode" @change="onSelectModeChange">多选模式</el-checkbox>
    </div>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="6" v-for="scene in scenesList" :key="scene.id">
        <el-card shadow="hover" class="scene-card" :class="{ selected: selectedIds.includes(scene.id) }">
          <div class="scene-preview" @click="selectMode ? toggleSelect(scene.id) : editScene(scene)">
            <div v-if="selectMode" class="select-overlay">
              <el-checkbox :model-value="selectedIds.includes(scene.id)" @click.stop />
            </div>
            <ImagePreview
              :image-url="getImageUrl(scene)"
              :alt="scene.location + ' - ' + scene.time"
              :size="120"
              :show-placeholder-text="false"
            />
          </div>

          <div class="scene-info">
            <h4>{{ scene.name || scene.location }}</h4>
            <p class="desc">{{ scene.description }}</p>
          </div>

          <div class="scene-actions" v-if="!selectMode">
            <el-button size="small" @click="editScene(scene)">{{ $t('common.edit') }}</el-button>
            <el-button size="small" @click="generateSceneImage(scene)">{{ $t('prop.generateImage') }}</el-button>
            <el-button size="small" type="danger" @click="deleteScene(scene)">{{ $t('common.delete') }}</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="scenesList.length === 0" :description="$t('drama.management.noScenes')" />

    <!-- CRUD 侧边栏 -->
    <SlidePanel v-model="panelVisible" :title="panelTitle" @save="saveNow">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" @keydown.ctrl.enter="saveNow">
        <el-form-item :label="$t('common.image')">
          <el-upload
            class="avatar-uploader"
            action="/api/v1/upload/image"
            :show-file-list="false"
            :on-success="handleImageSuccess"
            :before-upload="beforeUpload"
          >
            <img
              v-if="hasImage(form)"
              :src="getImageUrl(form)"
              class="avatar-preview scene-preview-img"
            />
            <el-icon v-else class="avatar-uploader-placeholder scene-placeholder"><Plus /></el-icon>
          </el-upload>
        </el-form-item>
        <el-form-item :label="$t('common.name')" prop="location">
          <el-input v-model="form.location" :placeholder="$t('common.name')" />
        </el-form-item>
        <el-form-item :label="$t('common.description')">
          <el-input v-model="form.prompt" type="textarea" :rows="4" :placeholder="$t('common.description')" />
        </el-form-item>
      </el-form>
    </SlidePanel>

    <!-- 提取侧边栏 -->
    <SlidePanel v-model="extractPanelVisible" :title="$t('prop.extractTitle')">
      <el-form label-width="100px">
        <el-form-item :label="$t('prop.selectEpisode')">
          <el-select v-model="extractEpisodeId" :placeholder="$t('common.pleaseSelect')" style="width: 100%">
            <el-option v-for="ep in sortedEpisodes" :key="ep.id" :label="ep.title" :value="ep.id" />
          </el-select>
        </el-form-item>
        <el-alert :title="$t('prop.extractTip')" type="info" :closable="false" show-icon />
      </el-form>
      <template #footer>
        <el-button @click="extractPanelVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :disabled="!extractEpisodeId" @click="handleExtract">
          {{ $t('prop.startExtract') }}
        </el-button>
      </template>
    </SlidePanel>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Document, Plus } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import { SlidePanel, ImagePreview } from '@/components/common'
import { useDramaContext } from '@/composables/useDramaContext'
import { useAutoSave } from '@/composables/useAutoSave'
import { useImageTaskTerminal, stopAllImageTaskTimers } from '@/composables/useImageTaskTerminal'
import { getImageUrl, hasImage } from '@/utils/image'

const { dramaData, dramaId: _dramaId, scenes, loadScenes } = useDramaContext()
const { track: trackImageTasks, failImmediately: failImageTask } = useImageTaskTerminal()

const sceneImageIdentity = (sc: any) => `${sc?.local_path || ''}|${sc?.image_url || ''}`
const findScene = (id: number | string) =>
  scenesList.value.find((sc) => String(sc.id) === String(id))

let isUnmounted = false
const activeTimers: number[] = []

const safeInterval = (callback: () => void, delay: number, maxAttempts = 10) => {
  let count = 0
  const timer = window.setInterval(() => {
    if (isUnmounted || ++count >= maxAttempts) {
      clearInterval(timer)
      const idx = activeTimers.indexOf(timer)
      if (idx > -1) activeTimers.splice(idx, 1)
      return
    }
    callback()
  }, delay)
  activeTimers.push(timer)
  return timer
}

const clearAllTimers = () => {
  activeTimers.forEach((t) => clearInterval(t))
  activeTimers.length = 0
}

const scenesList = computed(() => scenes.value)

const sortedEpisodes = computed(() => {
  if (!dramaData.value?.episodes) return []
  return [...dramaData.value.episodes].sort((a, b) => a.episode_number - b.episode_number)
})

// 多选
const selectMode = ref(false)
const selectedIds = ref<number[]>([])

function onSelectModeChange(val: boolean) {
  if (!val) selectedIds.value = []
}

function toggleSelect(id: number) {
  if (!selectMode.value) return
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

// CRUD 面板
const panelVisible = ref(false)
const editingId = ref<number | null>(null)

const form = ref({
  location: '',
  prompt: '',
  image_url: '',
  local_path: '',
})

const formRef = ref<FormInstance>()

const rules: FormRules = {
  location: [{ required: true, message: '请输入场景名称', trigger: 'blur' }],
}

const panelTitle = computed(() => (editingId.value ? '编辑场景' : '添加场景'))

const autoSaveEnabled = computed(() => panelVisible.value && form.value.location.trim().length > 0)
const autoSave = useAutoSave({
  source: form,
  enabled: autoSaveEnabled,
  delay: 300,
  validate: async () => {
    if (!formRef.value) return true
    try {
      await formRef.value.validate()
      return true
    } catch { return false }
  },
  saveFn: async (data) => {
    if (editingId.value) {
      await dramaAPI.updateScene(String(editingId.value), {
        location: data.location,
        description: data.prompt,
        image_url: data.image_url,
        local_path: data.local_path,
      })
    } else {
      await dramaAPI.createScene({
        drama_id: Number(dramaData.value!.id),
        location: data.location,
        prompt: data.prompt,
        description: data.prompt,
        image_url: data.image_url,
        local_path: data.local_path,
      })
    }
    await loadScenes()
    return true
  },
})

async function saveNow() {
  await autoSave.saveNow()
}

onUnmounted(() => {
  isUnmounted = true
  clearAllTimers()
  stopAllImageTaskTimers()
  autoSave.destroy()
})

function openAdd() {
  editingId.value = null
  form.value = { location: '', prompt: '', image_url: '', local_path: '' }
  panelVisible.value = true
}

function editScene(scene: any) {
  if (selectMode.value) return
  editingId.value = scene.id
  form.value = {
    location: scene.location || scene.name || '',
    prompt: scene.prompt || scene.description || '',
    image_url: scene.image_url || '',
    local_path: scene.local_path || '',
  }
  panelVisible.value = true
}

async function deleteScene(scene: any) {
  if (!scene.id) return
  try {
    await ElMessageBox.confirm(
      `确定要删除场景"${scene.name || scene.location}"吗？此操作不可恢复。`,
      '删除确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await dramaAPI.deleteScene(String(scene.id))
    ElMessage.success('场景已删除')
    await loadScenes()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '删除失败')
  }
}

async function generateSceneImage(scene: any) {
  const subject = `场景「${scene.location || scene.name || scene.id}」`
  const chip = '场景图 SCENE_SYNTH'
  const baseline = sceneImageIdentity(scene)
  try {
    await dramaAPI.generateSceneImage({ scene_id: scene.id })
    trackImageTasks({
      subject,
      chip,
      targets: [
        {
          id: scene.id,
          tag: `SCENE_${scene.id}「${scene.location || scene.name || ''}」`,
          hasResult: (sc: any) =>
            !!sc && sceneImageIdentity(sc) !== baseline && !!(sc.local_path || sc.image_url),
        },
      ],
      refresh: loadScenes,
      findEntity: findScene,
      maxAttempts: 24,
    })
  } catch (error: any) {
    failImageTask({ subject, chip, message: error.message || '生成请求失败' })
  }
}

// 批量
async function batchDelete() {
  if (selectedIds.value.length === 0) return
  const names = scenesList.value
    .filter((s) => selectedIds.value.includes(s.id))
    .map((s) => s.name || s.location)
    .join('、')
  try {
    await ElMessageBox.confirm(
      `确定要批量删除以下场景吗？\n${names}\n\n此操作不可恢复。`,
      '批量删除确认',
      { confirmButtonText: '确定删除', cancelButtonText: '取消', type: 'warning' }
    )
    for (const id of selectedIds.value) {
      await dramaAPI.deleteScene(String(id))
    }
    ElMessage.success(`已删除 ${selectedIds.value.length} 个场景`)
    selectedIds.value = []
    selectMode.value = false
    await loadScenes()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '批量删除失败')
  }
}

async function batchGenerateImages() {
  if (selectedIds.value.length === 0) return
  const ids = [...selectedIds.value]
  const chip = '批量场景图 BATCH_SCENE'
  const subject = `批量场景图 ×${ids.length}`
  const baselines = new Map(
    ids.map((id) => [id, sceneImageIdentity(findScene(id))]),
  )
  try {
    for (const id of ids) {
      await dramaAPI.generateSceneImage({ scene_id: id })
    }
    trackImageTasks({
      subject,
      chip,
      targets: ids.map((id) => ({
        id,
        tag: `SCENE_${id}「${findScene(id)?.location || findScene(id)?.name || ''}」`,
        hasResult: (sc: any) =>
          !!sc && sceneImageIdentity(sc) !== baselines.get(id) && !!(sc.local_path || sc.image_url),
      })),
      refresh: loadScenes,
      findEntity: findScene,
      maxAttempts: 36,
    })
  } catch (error: any) {
    failImageTask({ subject, chip, message: error.message || '批量生成请求失败' })
  }
}

// 提取
const extractPanelVisible = ref(false)
const extractEpisodeId = ref<string | null>(null)

watch(extractPanelVisible, (val) => {
  if (val && sortedEpisodes.value.length > 0 && !extractEpisodeId.value) {
    extractEpisodeId.value = sortedEpisodes.value[0].id
  }
})

async function handleExtract() {
  if (!extractEpisodeId.value) return
  try {
    await dramaAPI.extractBackgrounds(String(extractEpisodeId.value))
    extractPanelVisible.value = false
    ElMessage.success('场景提取任务已提交')
    safeInterval(async () => {
      await loadScenes()
    }, 5000, 10)
  } catch (error: any) {
    ElMessage.error(error.message || '提取失败')
  }
}

// 图片上传
function handleImageSuccess(response: any) {
  if (response.data?.url) {
    form.value.image_url = response.data.url
    form.value.local_path = response.data.local_path || ''
  }
}

function beforeUpload(file: any) {
  const isImage = file.type.startsWith('image/')
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isImage) ElMessage.error('只能上传图片文件!')
  if (!isLt10M) ElMessage.error('图片大小不能超过 10MB!')
  return isImage && isLt10M
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
  flex-wrap: wrap;
}

.select-mode-bar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.scene-card {
  margin-bottom: var(--space-4);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  overflow: hidden;
  transition: all var(--transition-normal);
}

.scene-card:hover {
  border-color: var(--border-secondary);
  box-shadow: var(--shadow-card-hover);
}

.scene-card.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent);
}

.scene-card :deep(.el-card__body) {
  padding: 0;
}

.scene-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 160px;
  background: linear-gradient(135deg, var(--accent) 0%, #06b6d4 100%);
  overflow: hidden;
  cursor: pointer;
  position: relative;
}

.select-overlay {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 2;
}

.scene-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-normal);
}

.scene-card:hover .scene-preview img {
  transform: scale(1.05);
}

.scene-info {
  text-align: center;
  padding: var(--space-4);
}

.scene-info h4 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.desc {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: var(--space-2) 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

.scene-actions {
  display: flex;
  gap: var(--space-2);
  justify-content: center;
  padding: 0 var(--space-4) var(--space-4);
}

.avatar-preview {
  object-fit: cover;
  border-radius: 6px;
}

.scene-preview-img {
  width: 160px;
  height: 90px;
}

.scene-placeholder {
  width: 160px;
  height: 90px;
}

.avatar-uploader-placeholder {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  font-size: 28px;
  color: #8c939d;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
