<template>
  <!-- Drama Create Page / 创建短剧页面 -->
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <!-- Page Sub Header / 页面子头部 -->
      <div class="page-sub-header">
        <el-button text @click="goBack" class="back-btn">
          <el-icon><ArrowLeft /></el-icon>
          <span>返回</span>
        </el-button>
        <div class="page-title">
          <h1>创建新项目</h1>
          <span class="subtitle">填写基本信息来创建你的短剧项目</span>
        </div>
      </div>

      <!-- Form Card / 表单卡片 -->
      <div class="form-card">

        <el-form 
          ref="formRef" 
          :model="form" 
          :rules="rules" 
          label-position="top"
          class="create-form"
          @submit.prevent="handleSubmit"
        >
          <el-form-item label="项目标题" prop="title" required>
            <el-input 
              v-model="form.title" 
              placeholder="给你的短剧起个名字"
              size="large"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="项目描述" prop="description">
            <el-input 
              v-model="form.description" 
              type="textarea" 
              :rows="4"
              placeholder="简要描述你的短剧内容、风格或创意（可选）"
              maxlength="500"
              show-word-limit
              resize="none"
            />
          </el-form-item>

          <el-form-item label="创作风格" prop="style">
            <div class="style-grid">
              <div
                v-for="style in stylePresets"
                :key="style.id"
                class="style-card"
                :class="{ active: selectedStyle === style.id }"
                @click="selectedStyle = style.id"
              >
                <div class="style-icon">
                  <StyleIcon :name="style.id" />
                </div>
                <div class="style-name">{{ style.name }}</div>
                <div class="style-desc">{{ style.description }}</div>
                <div v-if="selectedStyle === style.id" class="style-check">
                  <el-icon><Check /></el-icon>
                </div>
              </div>
            </div>
            <el-input
              v-if="selectedStyle === CUSTOM_STYLE_ID"
              v-model="form.custom_style"
              type="textarea"
              :rows="3"
              maxlength="500"
              show-word-limit
              resize="none"
              placeholder="描述你想要的视觉风格，例如：水墨丹青风格，黑白灰为主色调，留白意境，毛笔笔触，淡雅国风（图片和视频生成都会遵循此描述）"
              class="custom-style-input"
            />
          </el-form-item>

          <div class="form-actions">
            <el-button size="large" @click="goBack">取消</el-button>
            <el-button 
              type="primary" 
              size="large"
              :loading="loading"
              @click="handleSubmit"
            >
              <el-icon v-if="!loading"><Plus /></el-icon>
              创建项目
            </el-button>
          </div>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { ArrowLeft, Plus, Check } from '@element-plus/icons-vue'
import { dramaAPI } from '@/api/drama'
import type { CreateDramaRequest } from '@/types/drama'
import { stylePresets, CUSTOM_STYLE_ID } from '@/config/stylePresets'
import StyleIcon from '@/components/common/StyleIcon.vue'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const selectedStyle = ref('ghibli')

const form = reactive<CreateDramaRequest>({
  title: '',
  description: '',
  style: 'ghibli',
  custom_style: '',
})

const rules: FormRules = {
  title: [
    { required: true, message: '请输入项目标题', trigger: 'blur' },
    { min: 1, max: 100, message: '标题长度在 1 到 100 个字符', trigger: 'blur' }
  ]
}

// Submit form / 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      if (selectedStyle.value === CUSTOM_STYLE_ID && !form.custom_style?.trim()) {
        ElMessage.warning('请填写自定义风格描述')
        return
      }
      loading.value = true
      try {
        form.style = selectedStyle.value
        const drama = await dramaAPI.create(form)
        ElMessage.success('创建成功')
        router.push(`/dramas/${drama.id}`)
      } catch (error: any) {
        ElMessage.error(error.message || '创建失败')
      } finally {
        loading.value = false
      }
    }
  })
}

// Go back / 返回上一页
const goBack = () => {
  router.back()
}
</script>

<style scoped>
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100vh;
  background-color: var(--bg-primary);
  padding: var(--space-2) var(--space-3);
  transition: background-color var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    padding: var(--space-3) var(--space-4);
  }
}

.content-wrapper {
  max-width: 640px;
  margin: 0 auto;
}

/* Page Sub Header */
.page-sub-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) 0;
}

.page-sub-header .back-btn {
  color: var(--text-secondary);
  font-weight: 500;
}

.page-sub-header .page-title h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.2;
}

.page-sub-header .page-title .subtitle {
  display: block;
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-top: 2px;
}

/* ========================================
   Form Card / 表单卡片
   ======================================== */
.form-card {
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-card);
}

/* ========================================
   Form Styles / 表单样式 - 紧凑内边距
   ======================================== */
.create-form {
  padding: var(--space-4);
}

.create-form :deep(.el-form-item) {
  margin-bottom: var(--space-4);
}

/* ========================================
   Form Actions / 表单操作区
   ======================================== */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-primary);
  margin-top: var(--space-2);
}

.form-actions .el-button {
  min-width: 100px;
}

/* Style Grid */
.style-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  width: 100%;
}

@media (min-width: 640px) {
  .style-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.custom-style-input {
  margin-top: 12px;
}

.style-card {
  position: relative;
  padding: 16px 12px;
  border: 2px solid var(--border-primary);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: var(--bg-card);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 8px;
}

.style-card:hover {
  border-color: var(--accent);
  transform: translateY(-2px);
  box-shadow: var(--shadow-card);
}

.style-card.active {
  border-color: var(--accent);
  background: var(--accent-light);
}

.style-icon {
  font-size: 30px;
  line-height: 1;
  color: var(--text-secondary);
  display: flex;
  transition: color 0.2s ease;
}

.style-card:hover .style-icon {
  color: var(--accent);
}

.style-card.active .style-icon {
  color: var(--accent);
}

.style-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.style-desc {
  font-size: 11px;
  color: var(--text-secondary);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.style-check {
  position: absolute;
  top: 8px;
  right: 8px;
  color: var(--accent);
  font-size: 16px;
}
</style>
