<template>
  <div class="dramas-page">
    <!-- Page Header / 页面头部 -->
    <section class="page-header">
      <div class="header-content">
        <div class="header-text">
          <h1 class="page-title">
            <span class="title-bar" />作品存档
            <span class="title-en">/ WORKS_ARCHIVE</span>
          </h1>
          <p class="page-subtitle">
            ARCHIVE_COUNT // 共 <span class="total-count">{{ total }}</span> 个作品
          </p>
        </div>
        <div class="header-actions">
          <el-input
            v-model="searchQuery"
            placeholder="搜索作品..."
            class="search-input"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
      </div>
    </section>

    <!-- Projects Grid / 作品网格 -->
    <section class="projects-section">
      <div
        v-loading="loading"
        class="projects-grid"
        :class="{ 'is-empty': !loading && dramas.length === 0 }"
      >
        <!-- Empty state / 空状态 -->
        <div v-if="!loading && dramas.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="44"><Film /></el-icon>
          </div>
          <h3 class="empty-title">暂无存档</h3>
          <p class="empty-en">NO_ARCHIVE_FOUND</p>
          <p class="empty-desc">创建你的第一个短剧项目，开始创作之旅</p>
          <el-button type="primary" size="large" @click="handleCreate" class="empty-btn">
            <el-icon><Plus /></el-icon>
            <span>创建作品</span>
            <span class="btn-en">INIT_SESSION</span>
          </el-button>
        </div>

        <!-- Project Cards / 项目卡片列表 -->
        <ProjectCard
          v-for="drama in dramas"
          :key="drama.id"
          :title="drama.title"
          :description="drama.description"
          :updated-at="drama.updated_at"
          :episode-count="drama.total_episodes || 0"
          :style="drama.style"
          @click="viewDrama(drama.id)"
        >
          <template #actions>
            <ActionButton
              :icon="Edit"
              tooltip="编辑"
              @click="editDrama(drama.id)"
            />
            <el-popconfirm
              title="确定要删除这个作品吗？"
              confirm-button-text="删除"
              cancel-button-text="取消"
              @confirm="deleteDrama(drama.id)"
            >
              <template #reference>
                <el-button :icon="Delete" class="action-button danger" link />
              </template>
            </el-popconfirm>
          </template>
        </ProjectCard>

        <!-- Create New Card / 创建新作品卡片 -->
        <div v-if="!loading && dramas.length > 0" class="create-card" @click="handleCreate">
          <div class="create-inner">
            <div class="create-icon">
              <el-icon :size="32"><Plus /></el-icon>
            </div>
            <h3 class="create-title">创建新作品</h3>
            <p class="create-desc">NEW_ARCHIVE // 开启新的创作灵感</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Pagination / 分页 -->
    <div v-if="total > pageSize" class="pagination-section">
      <el-pagination
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.page_size"
        :total="total"
        :page-sizes="[12, 24, 36, 48]"
        :pager-count="5"
        layout="prev, pager, next, total"
        background
        @size-change="loadDramas"
        @current-change="loadDramas"
      />
    </div>

    <!-- Edit Dialog / 编辑对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑作品"
      width="520px"
      :close-on-click-modal="false"
      class="edit-dialog"
    >
      <el-form
        :model="editForm"
        label-position="top"
        v-loading="editLoading"
        class="edit-form"
      >
        <el-form-item label="作品名称" required>
          <el-input
            v-model="editForm.title"
            placeholder="给你的作品起个名字"
            size="large"
          />
        </el-form-item>
        <el-form-item label="作品简介">
          <el-input
            v-model="editForm.description"
            type="textarea"
            :rows="4"
            placeholder="简单描述一下你的作品..."
            resize="none"
          />
        </el-form-item>
        <el-form-item label="风格类型" required>
          <el-select
            v-model="editForm.style"
            placeholder="选择一个画风"
            size="large"
            style="width: 100%"
          >
            <el-option label="吉卜力风格" value="ghibli" />
            <el-option label="国漫画风" value="guoman" />
            <el-option label="废土风格" value="wasteland" />
            <el-option label="复古风格" value="nostalgia" />
            <el-option label="像素风格" value="pixel" />
            <el-option label="体素风格" value="voxel" />
            <el-option label="都市风格" value="urban" />
            <el-option label="3D国漫" value="guoman3d" />
            <el-option label="Q版3D" value="chibi3d" />
            <el-option label="自定义风格" value="custom" />
          </el-select>
          <el-input
            v-if="editForm.style === 'custom'"
            v-model="editForm.custom_style"
            type="textarea"
            :rows="3"
            maxlength="500"
            show-word-limit
            resize="none"
            placeholder="描述你想要的视觉风格（图片和视频都会遵循）"
            style="margin-top: 12px"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button size="large" @click="editDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            size="large"
            :loading="editLoading"
            @click="saveEdit"
          >
            保存
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Create Drama Dialog / 创建短剧弹窗 -->
    <CreateDramaDialog v-model="createDialogVisible" @created="onCreated" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { useRouter } from "vue-router"
import { ElMessage } from "element-plus"
import {
  Plus,
  Film,
  Edit,
  Delete,
  Search,
} from "@element-plus/icons-vue"
import { dramaAPI } from "@/api/drama"
import type { Drama, DramaListQuery } from "@/types/drama"
import {
  ProjectCard,
  ActionButton,
  CreateDramaDialog,
} from "@/components/common"

const router = useRouter()
const loading = ref(false)
const dramas = ref<Drama[]>([])
const total = ref(0)
const pageSize = ref(12)
const searchQuery = ref('')

const queryParams = ref<DramaListQuery>({
  page: 1,
  page_size: 12,
})

const createDialogVisible = ref(false)

const editDialogVisible = ref(false)
const editLoading = ref(false)
const editForm = ref({
  id: "",
  title: "",
  description: "",
  style: "ghibli",
  custom_style: "",
})

const loadDramas = async () => {
  loading.value = true
  try {
    const res = await dramaAPI.list(queryParams.value)
    dramas.value = res.items || []
    total.value = res.pagination?.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || "加载失败")
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  createDialogVisible.value = true
}

const onCreated = () => {
  loadDramas()
}

const viewDrama = (id: string) => {
  router.push(`/dramas/${id}`)
}

const editDrama = async (id: string) => {
  editLoading.value = true
  editDialogVisible.value = true
  try {
    const drama = await dramaAPI.get(id)
    editForm.value = {
      id: drama.id,
      title: drama.title,
      description: drama.description || "",
      style: drama.style || "ghibli",
      custom_style: drama.custom_style || "",
    }
  } catch (error: any) {
    ElMessage.error(error.message || "加载失败")
    editDialogVisible.value = false
  } finally {
    editLoading.value = false
  }
}

const saveEdit = async () => {
  if (!editForm.value.title) {
    ElMessage.warning("请输入作品名称")
    return
  }
  if (editForm.value.style === 'custom' && !editForm.value.custom_style.trim()) {
    ElMessage.warning("请填写自定义风格描述")
    return
  }

  editLoading.value = true
  try {
    await dramaAPI.update(editForm.value.id, {
      title: editForm.value.title,
      description: editForm.value.description,
      style: editForm.value.style,
      custom_style: editForm.value.custom_style,
    })
    ElMessage.success("保存成功")
    editDialogVisible.value = false
    loadDramas()
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败")
  } finally {
    editLoading.value = false
  }
}

const deleteDrama = async (id: string) => {
  try {
    await dramaAPI.delete(id)
    ElMessage.success("删除成功")
    loadDramas()
  } catch (error: any) {
    ElMessage.error(error.message || "删除失败")
  }
}

onMounted(() => {
  loadDramas()
})
</script>

<style scoped>
.dramas-page {
  min-height: 100vh;
  /* 背景透明，露出全局 Three.js 终端背景 */
}

/* Page Header */
.page-header {
  padding: var(--space-8) var(--space-6) var(--space-6);
  max-width: 1400px;
  margin: 0 auto;
}

.header-content {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-6);
  flex-wrap: wrap;
}

.header-text {
  flex: 1;
  min-width: 200px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 1.75rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0 0 var(--space-2);
}

.title-bar {
  display: inline-block;
  width: 4px;
  height: 26px;
  background: var(--accent);
  box-shadow: 0 0 10px rgba(255, 45, 45, 0.5);
}

.title-en {
  font-family: var(--font-mono);
  font-size: 0.8rem;
  font-weight: 500;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}

.page-subtitle {
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
  margin: 0;
  padding-left: 16px;
}

.total-count {
  font-weight: 600;
  color: var(--accent);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.search-input {
  width: 280px;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: var(--radius-xl);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
  transition: all var(--transition-fast);
}

.search-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--border-secondary) inset;
}

.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px var(--accent) inset;
}

/* Projects Section */
.projects-section {
  padding: 0 var(--space-6) var(--space-12);
  max-width: 1400px;
  margin: 0 auto;
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: var(--space-5);
  min-height: 300px;
}

@media (min-width: 640px) {
  .projects-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 900px) {
  .projects-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1200px) {
  .projects-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

.projects-grid.is-empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: var(--space-16) var(--space-6);
}

.empty-icon {
  width: 88px;
  height: 88px;
  margin: 0 auto var(--space-5);
  background: transparent;
  border: 1px solid color-mix(in srgb, var(--accent) 55%, transparent);
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  position: relative;
}

.empty-icon::before,
.empty-icon::after {
  content: "";
  position: absolute;
  background: var(--accent);
}

.empty-icon::before {
  width: 12px;
  height: 1px;
  top: -1px;
  left: 8px;
  box-shadow: 56px 0 0 var(--accent);
}

.empty-icon::after {
  width: 1px;
  height: 12px;
  left: -1px;
  top: 8px;
  box-shadow: 0 56px 0 var(--accent);
}

.empty-title {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0 0 4px;
}

.empty-en {
  margin: 0 0 var(--space-4);
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.18em;
  color: var(--text-muted);
}

.empty-desc {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0 0 var(--space-6);
}

.empty-btn {
  height: 46px;
  padding: 0 26px;
  border-radius: var(--radius-md);
  font-weight: 700;
  background: var(--accent-gradient);
  border: none;
  box-shadow: var(--shadow-glow);
}

.empty-btn .btn-en {
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  opacity: 0.85;
  margin-left: 4px;
}

.empty-btn:hover {
  box-shadow: var(--shadow-glow-purple);
}

/* Create Card */
.create-card {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 280px;
  background: color-mix(in srgb, var(--bg-card) 55%, transparent);
  border: 1px dashed var(--border-secondary);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition:
    border-color var(--transition-fast),
    background var(--transition-fast);
}

.create-card:hover {
  border-color: var(--accent);
  background: var(--accent-light);
}

.create-inner {
  text-align: center;
  padding: var(--space-6);
}

.create-icon {
  width: 52px;
  height: 52px;
  margin: 0 auto var(--space-4);
  background: transparent;
  border: 1px solid var(--border-secondary);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.create-card:hover .create-icon {
  border-color: var(--accent);
  background: var(--accent);
  color: #fff;
}

.create-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 var(--space-1);
}

.create-desc {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0;
}

/* Pagination */
.pagination-section {
  padding: 0 var(--space-6) var(--space-12);
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  justify-content: center;
}

.pagination-section :deep(.el-pagination) {
  --el-pagination-button-bg-color: var(--bg-card);
  --el-pagination-button-hover-bg-color: var(--bg-card-hover);
  --el-pagination-button-active-bg-color: var(--accent);
}

.pagination-section :deep(.el-pagination.is-background .el-pager li) {
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  margin: 0 2px;
}

.pagination-section :deep(.el-pagination.is-background .btn-prev),
.pagination-section :deep(.el-pagination.is-background .btn-next) {
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
}

/* Edit Dialog */
.edit-dialog :deep(.el-dialog) {
  border-radius: var(--radius-2xl);
}

.edit-dialog :deep(.el-dialog__header) {
  padding: var(--space-5) var(--space-6);
  border-bottom: 1px solid var(--border-primary);
}

.edit-dialog :deep(.el-dialog__title) {
  font-size: 1.125rem;
  font-weight: 600;
}

.edit-dialog :deep(.el-dialog__body) {
  padding: var(--space-6);
}

.edit-dialog :deep(.el-dialog__footer) {
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid var(--border-primary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
}

/* Responsive */
@media (max-width: 768px) {
  .page-header {
    padding: var(--space-6) var(--space-4) var(--space-4);
  }

  .page-title {
    font-size: 1.5rem;
  }

  .search-input {
    width: 100%;
  }

  .header-content {
    flex-direction: column;
    align-items: stretch;
  }

  .projects-section {
    padding: 0 var(--space-4) var(--space-8);
  }

  .pagination-section {
    padding: 0 var(--space-4) var(--space-8);
  }
}
</style>
