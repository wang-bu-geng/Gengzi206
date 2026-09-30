<template>
  <div>
    <div class="tab-header">
      <h2>{{ $t('drama.management.characterList') }}</h2>
      <div class="tab-header-actions">
        <template v-if="selectedIds.length > 0">
          <el-button type="danger" @click="batchDelete">批量删除 ({{ selectedIds.length }})</el-button>
          <el-button type="success" @click="batchGenerateImages">批量生成图片 ({{ selectedIds.length }})</el-button>
        </template>
        <el-button :icon="Document" @click="extractPanelVisible = true">{{ $t('prop.extract') }}</el-button>
        <el-button type="primary" :icon="Plus" @click="openAdd">{{ $t('character.add') }}</el-button>
      </div>
    </div>

    <!-- 多选模式开关 -->
    <div class="select-mode-bar" v-if="characters.length > 0">
      <el-checkbox v-model="selectMode" @change="onSelectModeChange">
        多选模式
      </el-checkbox>
    </div>

    <!-- 卡片网格 -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="6" v-for="character in characters" :key="character.id">
        <el-card shadow="hover" class="character-card" :class="{ selected: selectedIds.includes(character.id) }">
          <div class="character-preview" @click="selectMode ? toggleSelect(character.id) : editCharacter(character)">
            <div v-if="selectMode" class="select-overlay">
              <el-checkbox :model-value="selectedIds.includes(character.id)" @click.stop />
            </div>
            <ImagePreview
              v-if="character.local_path || character.image_url"
              :image-url="getImageUrl(character)"
              :alt="character.name"
              :size="120"
            />
            <el-avatar v-else :size="120">{{ character.name[0] }}</el-avatar>
          </div>

          <div class="character-info">
            <div class="character-name">
              <h4>{{ character.name }}</h4>
              <el-tag :type="character.role === 'main' ? 'danger' : 'info'" size="small">
                {{ roleLabel(character.role) }}
              </el-tag>
            </div>

            <div v-if="turnaroundUrl(character)" class="turnaround-row">
              <span class="turnaround-tag">三视图</span>
              <el-image
                :src="turnaroundUrl(character)"
                :preview-src-list="[turnaroundUrl(character)]"
                :preview-teleported="true"
                fit="cover"
                class="turnaround-thumb"
              />
            </div>
            <p class="desc">{{ character.appearance || character.description }}</p>
          </div>

          <div class="character-actions" v-if="!selectMode">
            <el-button size="small" @click="editCharacter(character)">{{ $t('common.edit') }}</el-button>
            <el-button size="small" @click="generateCharacterImage(character)">{{ $t('prop.generateImage') }}</el-button>
            <el-button size="small" :loading="turnaroundPending.has(character.id)" @click="generateTurnaround(character)">三视图</el-button>
            <el-button size="small" type="danger" @click="deleteCharacter(character)">{{ $t('common.delete') }}</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty
      v-if="characters.length === 0"
      :description="$t('drama.management.noCharacters')"
    />

    <!-- CRUD 侧边栏 -->
    <SlidePanel v-model="panelVisible" :title="panelTitle" @save="saveNow">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" @keydown.ctrl.enter="saveNow">
        <el-form-item :label="$t('character.image')">
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
        <el-form-item :label="$t('character.name')" prop="name">
          <el-input v-model="form.name" :placeholder="$t('character.name')" />
        </el-form-item>
        <el-form-item :label="$t('character.role')">
          <el-select v-model="form.role" :placeholder="$t('common.pleaseSelect')">
            <el-option label="Main" value="main" />
            <el-option label="Supporting" value="supporting" />
            <el-option label="Minor" value="minor" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('character.appearance')">
          <el-input v-model="form.appearance" type="textarea" :rows="3" :placeholder="$t('character.appearance')" />
        </el-form-item>
        <el-form-item :label="$t('character.personality')">
          <el-input v-model="form.personality" type="textarea" :rows="3" :placeholder="$t('character.personality')" />
        </el-form-item>
        <el-form-item :label="$t('character.description')">
          <el-input v-model="form.description" type="textarea" :rows="3" :placeholder="$t('common.description')" />
        </el-form-item>
      </el-form>
    </SlidePanel>

    <!-- 提取侧边栏 -->
    <SlidePanel v-model="extractPanelVisible" :title="$t('prop.extractTitle')">
      <el-form label-width="100px">
        <el-form-item :label="$t('prop.selectEpisode')">
          <el-select v-model="extractEpisodeId" :placeholder="$t('common.pleaseSelect')" style="width: 100%">
            <el-option
              v-for="ep in sortedEpisodes"
              :key="ep.id"
              :label="ep.title"
              :value="ep.id"
            />
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
import { characterLibraryAPI } from '@/api/character-library'
import { SlidePanel, ImagePreview } from '@/components/common'
import { useDramaContext } from '@/composables/useDramaContext'
import { useAutoSave } from '@/composables/useAutoSave'
import { useImageTaskTerminal, stopAllImageTaskTimers } from '@/composables/useImageTaskTerminal'
import { getImageUrl, hasImage } from '@/utils/image'

const { dramaData, dramaId, loadDramaData } = useDramaContext()
const { track: trackImageTasks, failImmediately: failImageTask } = useImageTaskTerminal()

const characterImageIdentity = (c: any) => `${c?.local_path || ''}|${c?.image_url || ''}`
const findCharacter = (id: number | string) =>
  (dramaData.value?.characters || []).find((c) => String(c.id) === String(id))

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

const characters = computed(() => dramaData.value?.characters || [])

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
  if (idx >= 0) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(id)
  }
}

// CRUD 面板
const panelVisible = ref(false)
const editingId = ref<number | null>(null)

const form = ref({
  name: '',
  role: 'supporting',
  appearance: '',
  personality: '',
  description: '',
  image_url: '',
  local_path: '',
})

const formRef = ref<FormInstance>()

const rules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

const panelTitle = computed(() => (editingId.value ? '编辑角色' : '添加角色'))

// 自动保存
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
    } catch {
      return false
    }
  },
  saveFn: async (data) => {
    if (editingId.value) {
      await characterLibraryAPI.updateCharacter(editingId.value, {
        name: data.name,
        role: data.role,
        appearance: data.appearance,
        personality: data.personality,
        description: data.description,
        image_url: data.image_url,
        local_path: data.local_path,
      })
    } else {
      const allCharacters = [
        ...(dramaData.value?.characters || []).map((c) => ({
          name: c.name,
          role: c.role,
          appearance: c.appearance,
          personality: c.personality,
          description: c.description,
          image_url: c.image_url,
          local_path: c.local_path,
        })),
        data,
      ]
      await dramaAPI.saveCharacters(dramaData.value!.id, allCharacters)
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
  stopAllImageTaskTimers()
  autoSave.destroy()
})

function openAdd() {
  editingId.value = null
  form.value = {
    name: '',
    role: 'supporting',
    appearance: '',
    personality: '',
    description: '',
    image_url: '',
    local_path: '',
  }
  panelVisible.value = true
}

function editCharacter(character: any) {
  if (selectMode.value) return
  editingId.value = character.id
  form.value = {
    name: character.name,
    role: character.role || 'supporting',
    appearance: character.appearance || '',
    personality: character.personality || '',
    description: character.description || '',
    image_url: character.image_url || '',
    local_path: character.local_path || '',
  }
  panelVisible.value = true
}

async function deleteCharacter(character: any) {
  if (!character.id) return
  try {
    await ElMessageBox.confirm(
      `确定要删除角色"${character.name}"吗？此操作不可恢复。`,
      '删除确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await characterLibraryAPI.deleteCharacter(character.id)
    ElMessage.success('角色已删除')
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '删除失败')
  }
}

async function generateCharacterImage(character: any) {
  const subject = `角色「${character.name}」`
  const chip = '角色立绘 CHARACTER_PORTRAIT'
  const baseline = characterImageIdentity(character)
  try {
    await characterLibraryAPI.generateCharacterImage(character.id)
    trackImageTasks({
      subject,
      chip,
      targets: [
        {
          id: character.id,
          tag: `CHAR_${character.id}「${character.name}」`,
          hasResult: (c: any) =>
            !!c && characterImageIdentity(c) !== baseline && !!(c.local_path || c.image_url),
        },
      ],
      refresh: loadDramaData,
      findEntity: findCharacter,
      maxAttempts: 36,
    })
  } catch (error: any) {
    failImageTask({ subject, chip, message: error.message || '生成请求失败' })
  }
}

// 三视图（turnaround）
const turnaroundPending = ref<Set<number>>(new Set())

function normalizeRefs(raw: any): string[] {
  if (!raw) return []
  const filterStrings = (list: any[]): string[] =>
    list.filter((x): x is string => typeof x === 'string' && !!x)
  if (Array.isArray(raw)) return filterStrings(raw)
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? filterStrings(parsed) : []
  } catch {
    return []
  }
}

function turnaroundUrl(character: any): string {
  const refs = normalizeRefs(character.reference_images)
  const last = refs[refs.length - 1]
  if (!last) return ''
  if (last.startsWith('http') || last.startsWith('data:')) return last
  return `/static/${last}`
}

async function generateTurnaround(character: any) {
  if (!character.id || turnaroundPending.value.has(character.id)) return
  turnaroundPending.value = new Set(turnaroundPending.value).add(character.id)
  const clearPending = () => {
    const next = new Set(turnaroundPending.value)
    next.delete(character.id)
    turnaroundPending.value = next
  }
  const subject = `三视图「${character.name}」`
  const chip = '角色三视图 CHARACTER_TURNAROUND'
  try {
    await characterLibraryAPI.generateTurnaround(character.id)
    trackImageTasks({
      subject,
      chip,
      targets: [
        {
          id: character.id,
          tag: `TURNAROUND_${character.id}「${character.name}」`,
          hasResult: (c: any) => !!c && normalizeRefs(c.reference_images).length > 0,
        },
      ],
      refresh: loadDramaData,
      findEntity: findCharacter,
      maxAttempts: 36,
      onSettled: (summary) => {
        clearPending()
        if (summary.success > 0) ElMessage.success(`「${character.name}」三视图已生成`)
      },
    })
  } catch (error: any) {
    clearPending()
    failImageTask({ subject, chip, message: error.message || '三视图生成请求失败' })
  }
}

// 批量操作
async function batchDelete() {
  if (selectedIds.value.length === 0) return
  const names = characters.value
    .filter((c) => selectedIds.value.includes(c.id))
    .map((c) => c.name)
    .join('、')
  try {
    await ElMessageBox.confirm(
      `确定要批量删除以下角色吗？\n${names}\n\n此操作不可恢复。`,
      '批量删除确认',
      { confirmButtonText: '确定删除', cancelButtonText: '取消', type: 'warning' }
    )
    for (const id of selectedIds.value) {
      await characterLibraryAPI.deleteCharacter(id)
    }
    ElMessage.success(`已删除 ${selectedIds.value.length} 个角色`)
    selectedIds.value = []
    selectMode.value = false
    await loadDramaData()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '批量删除失败')
  }
}

async function batchGenerateImages() {
  if (selectedIds.value.length === 0) return
  const ids = [...selectedIds.value]
  const chip = '批量立绘 BATCH_CHARACTER_PORTRAIT'
  const subject = `批量立绘 ×${ids.length}`
  const baselines = new Map(ids.map((id) => [String(id), characterImageIdentity(findCharacter(id))]))
  try {
    await characterLibraryAPI.batchGenerateCharacterImages(ids)
    trackImageTasks({
      subject,
      chip,
      targets: ids.map((id) => {
        const c = findCharacter(id)
        return {
          id,
          tag: `CHAR_${id}「${c?.name || ''}」`,
          hasResult: (entity: any) =>
            !!entity &&
            characterImageIdentity(entity) !== baselines.get(String(id)) &&
            !!(entity.local_path || entity.image_url),
        }
      }),
      refresh: loadDramaData,
      findEntity: findCharacter,
      maxAttempts: 48,
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
    extractEpisodeId.value = String(sortedEpisodes.value[0].id)
  }
})

async function handleExtract() {
  if (!extractEpisodeId.value) return
  try {
    await characterLibraryAPI.extractFromEpisode(String(extractEpisodeId.value))
    extractPanelVisible.value = false
    ElMessage.success('角色提取任务已提交')
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

function roleLabel(role?: string) {
  if (role === 'main') return 'Main'
  if (role === 'supporting') return 'Supporting'
  return 'Minor'
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

.character-card {
  margin-bottom: var(--space-4);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  overflow: hidden;
  transition: all var(--transition-normal);
}

.character-card:hover {
  border-color: var(--border-secondary);
  box-shadow: var(--shadow-card-hover);
}

.character-card.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent);
}

.character-card :deep(.el-card__body) {
  padding: 0;
}

.character-preview {
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

.character-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-normal);
}

.character-card:hover .character-preview img {
  transform: scale(1.05);
}

.character-info {
  text-align: center;
  padding: var(--space-4);
}

.character-name {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-2);
}

.character-info h4 {
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

.character-actions {
  display: flex;
  gap: var(--space-2);
  justify-content: center;
  padding: 0 var(--space-4) var(--space-4);
}

.turnaround-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  margin: var(--space-2) 0;
}

.turnaround-tag {
  font-size: 11px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.turnaround-thumb {
  width: 72px;
  height: 40px;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  cursor: zoom-in;
  background: rgba(255, 255, 255, 0.04);
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
