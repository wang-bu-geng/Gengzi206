<template>
  <div>
    <div class="tab-header">
      <h2>{{ $t('drama.management.propList') }}</h2>
      <div class="tab-header-actions">
        <template v-if="selectedIds.length > 0">
          <el-button type="danger" @click="batchDelete">批量删除 ({{ selectedIds.length }})</el-button>
          <el-button type="success" @click="batchGenerateImages">批量生成图片 ({{ selectedIds.length }})</el-button>
        </template>
        <el-button :icon="Document" @click="extractPanelVisible = true">{{ $t('prop.extract') }}</el-button>
        <el-button type="primary" :icon="Plus" @click="openAdd">{{ $t('common.add') }}</el-button>
      </div>
    </div>

    <div class="select-mode-bar" v-if="propsList.length > 0">
      <el-checkbox v-model="selectMode" @change="onSelectModeChange">多选模式</el-checkbox>
    </div>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="6" v-for="prop in propsList" :key="prop.id">
        <el-card shadow="hover" class="prop-card" :class="{ selected: selectedIds.includes(prop.id) }">
          <div class="prop-preview" @click="selectMode ? toggleSelect(prop.id) : editProp(prop)">
            <div v-if="selectMode" class="select-overlay">
              <el-checkbox :model-value="selectedIds.includes(prop.id)" @click.stop />
            </div>
            <ImagePreview
              :image-url="getImageUrl(prop)"
              :alt="prop.name"
              :size="120"
              :show-placeholder-text="false"
            />
          </div>

          <div class="prop-info">
            <h4>{{ prop.name }}</h4>
            <el-tag size="small" v-if="prop.type">{{ prop.type }}</el-tag>
            <p class="desc">{{ prop.description || prop.prompt }}</p>
          </div>

          <div class="prop-actions" v-if="!selectMode">
            <el-button size="small" @click="editProp(prop)">{{ $t('common.edit') }}</el-button>
            <el-button size="small" @click="generatePropImage(prop)" :disabled="!prop.prompt">{{ $t('prop.generateImage') }}</el-button>
            <el-button size="small" type="danger" @click="deleteProp(prop)">{{ $t('common.delete') }}</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="propsList.length === 0" :description="$t('drama.management.noProps')" />

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
              class="avatar-preview"
            />
            <el-icon v-else class="avatar-uploader-placeholder"><Plus /></el-icon>
          </el-upload>
        </el-form-item>
        <el-form-item :label="$t('prop.name')" prop="name">
          <el-input v-model="form.name" :placeholder="$t('prop.name')" />
        </el-form-item>
        <el-form-item :label="$t('prop.type')">
          <el-input v-model="form.type" :placeholder="$t('prop.typePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('prop.description')">
          <el-input v-model="form.description" type="textarea" :rows="3" :placeholder="$t('prop.description')" />
        </el-form-item>
        <el-form-item :label="$t('prop.prompt')">
          <el-input v-model="form.prompt" type="textarea" :rows="3" :placeholder="$t('prop.promptPlaceholder')" />
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
import { propAPI } from '@/api/prop'
import { SlidePanel, ImagePreview } from '@/components/common'
import { useDramaContext } from '@/composables/useDramaContext'
import { useAutoSave } from '@/composables/useAutoSave'
import { getImageUrl, hasImage } from '@/utils/image'

const { dramaData, dramaId, loadDramaData } = useDramaContext()

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

const propsList = computed(() => dramaData.value?.props || [])

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
  name: '',
  type: '',
  description: '',
  prompt: '',
  image_url: '',
  local_path: '',
})

const formRef = ref<FormInstance>()

const rules: FormRules = {
  name: [{ required: true, message: '请输入道具名称', trigger: 'blur' }],
}

const panelTitle = computed(() => (editingId.value ? '编辑道具' : '添加道具'))

const autoSaveEnabled = computed(() => panelVisible.value && form.value.name.trim().length > 0)
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
    const propData = {
      drama_id: dramaData.value!.id,
      name: data.name,
      description: data.description,
      prompt: data.prompt,
      type: data.type,
      image_url: data.image_url,
      local_path: data.local_path,
    }
    if (editingId.value) {
      await propAPI.update(editingId.value, propData)
    } else {
      await propAPI.create(propData as any)
    }
    await loadDramaData()
    return true
  },
})

async function saveNow() {
  await autoSave.saveNow()
}

onUnmounted(() => {
  isUnmounted = true
  clearAllTimers()
  autoSave.destroy()
})

function openAdd() {
  editingId.value = null
  form.value = { name: '', type: '', description: '', prompt: '', image_url: '', local_path: '' }
  panelVisible.value = true
}

function editProp(prop: any) {
  if (selectMode.value) return
  editingId.value = prop.id
  form.value = {
    name: prop.name,
    type: prop.type || '',
    description: prop.description || '',
    prompt: prop.prompt || '',
    image_url: prop.image_url || '',
    local_path: prop.local_path || '',
  }
  panelVisible.value = true
}

async function deleteProp(prop: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除道具"${prop.name}"吗？此操作不可恢复。`,
      '删除确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await propAPI.delete(prop.id)
    ElMessage.success('道具已删除')
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '删除失败')
  }
}

async function generatePropImage(prop: any) {
  if (!prop.prompt) {
    ElMessage.warning('请先设置道具的图片提示词')
    editProp(prop)
    return
  }
  try {
    await propAPI.generateImage(prop.id)
    ElMessage.success('图片生成任务已提交')
    safeInterval(async () => {
      await loadDramaData()
    }, 5000, 10)
  } catch (error: any) {
    ElMessage.error(error.message || '生成失败')
  }
}

// 批量
async function batchDelete() {
  if (selectedIds.value.length === 0) return
  const names = propsList.value
    .filter((p) => selectedIds.value.includes(p.id))
    .map((p) => p.name)
    .join('、')
  try {
    await ElMessageBox.confirm(
      `确定要批量删除以下道具吗？\n${names}\n\n此操作不可恢复。`,
      '批量删除确认',
      { confirmButtonText: '确定删除', cancelButtonText: '取消', type: 'warning' }
    )
    for (const id of selectedIds.value) {
      await propAPI.delete(id)
    }
    ElMessage.success(`已删除 ${selectedIds.value.length} 个道具`)
    selectedIds.value = []
    selectMode.value = false
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '批量删除失败')
  }
}

async function batchGenerateImages() {
  if (selectedIds.value.length === 0) return
  const validProps = propsList.value.filter(
    (p) => selectedIds.value.includes(p.id) && p.prompt
  )
  const noPrompt = selectedIds.value.filter(
    (id) => !propsList.value.find((p) => p.id === id)?.prompt
  )
  if (noPrompt.length > 0) {
    ElMessage.warning(`${noPrompt.length} 个道具缺少图片提示词，已跳过`)
  }
  if (validProps.length === 0) return
  try {
    for (const prop of validProps) {
      await propAPI.generateImage(prop.id)
    }
    ElMessage.success(`已提交 ${validProps.length} 个道具的图片生成任务`)
    safeInterval(async () => {
      await loadDramaData()
    }, 5000, 10)
  } catch (error: any) {
    ElMessage.error(error.message || '批量生成失败')
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
    await propAPI.extractFromScript(extractEpisodeId.value)
    extractPanelVisible.value = false
    ElMessage.success('道具提取任务已提交')
    safeInterval(async () => {
      await loadDramaData()
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

.prop-card {
  margin-bottom: var(--space-4);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  overflow: hidden;
  transition: all var(--transition-normal);
}

.prop-card:hover {
  border-color: var(--border-secondary);
  box-shadow: var(--shadow-card-hover);
}

.prop-card.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent);
}

.prop-card :deep(.el-card__body) {
  padding: 0;
}

.prop-preview {
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

.prop-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-normal);
}

.prop-card:hover .prop-preview img {
  transform: scale(1.05);
}

.prop-info {
  text-align: center;
  padding: var(--space-4);
}

.prop-info h4 {
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

.prop-actions {
  display: flex;
  gap: var(--space-2);
  justify-content: center;
  padding: 0 var(--space-4) var(--space-4);
}

.avatar-preview {
  width: 100px;
  height: 100px;
  object-fit: cover;
  border-radius: 6px;
}

.avatar-uploader-placeholder {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  width: 100px;
  height: 100px;
  font-size: 28px;
  color: #8c939d;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
