<template>
  <!-- Create Drama Dialog / 创建短剧弹窗 -->
  <el-dialog
    v-model="visible"
    :title="$t('drama.createNew')"
    width="520px"
    :close-on-click-modal="false"
    class="create-dialog"
    @closed="handleClosed"
  >
    <div class="dialog-desc">{{ $t("drama.createDesc") }}</div>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      class="create-form"
      @submit.prevent="handleSubmit"
    >
      <el-form-item :label="$t('drama.projectName')" prop="title" required>
        <el-input
          v-model="form.title"
          :placeholder="$t('drama.projectNamePlaceholder')"
          size="large"
          maxlength="100"
          show-word-limit
        />
      </el-form-item>

      <el-form-item :label="$t('drama.projectDesc')" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="4"
          :placeholder="$t('drama.projectDescPlaceholder')"
          maxlength="500"
          show-word-limit
          resize="none"
        />
      </el-form-item>

      <el-form-item :label="$t('drama.style')" prop="style" required>
        <el-select
          v-model="form.style"
          :placeholder="$t('drama.stylePlaceholder')"
          size="large"
          style="width: 100%"
        >
          <el-option :label="$t('drama.styles.ghibli')" value="ghibli" />
          <el-option :label="$t('drama.styles.guoman')" value="guoman" />
          <el-option :label="$t('drama.styles.wasteland')" value="wasteland" />
          <el-option :label="$t('drama.styles.nostalgia')" value="nostalgia" />
          <el-option :label="$t('drama.styles.pixel')" value="pixel" />
          <el-option :label="$t('drama.styles.voxel')" value="voxel" />
          <el-option :label="$t('drama.styles.urban')" value="urban" />
          <el-option :label="$t('drama.styles.guoman3d')" value="guoman3d" />
          <el-option :label="$t('drama.styles.chibi3d')" value="chibi3d" />
          <el-option label="自定义风格" value="custom" />
        </el-select>
        <el-input
          v-if="form.style === 'custom'"
          v-model="form.custom_style"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          resize="none"
          placeholder="描述你想要的视觉风格，例如：水墨丹青、黑白灰色调、毛笔笔触、留白意境（图片和视频都会遵循）"
          style="margin-top: 0.75rem"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button size="large" @click="handleClose">
          {{ $t("common.cancel") }}
        </el-button>
        <el-button
          type="primary"
          size="large"
          :loading="loading"
          @click="handleSubmit"
        >
          <el-icon v-if="!loading"><Plus /></el-icon>
          {{ $t("drama.createNew") }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance, type FormRules } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import { dramaAPI } from "@/api/drama";
import type { CreateDramaRequest } from "@/types/drama";

/**
 * CreateDramaDialog - Reusable dialog for creating new drama projects
 * 创建短剧弹窗 - 可复用的创建短剧项目弹窗
 */
const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
  created: [id: string];
}>();

const router = useRouter();
const formRef = ref<FormInstance>();
const loading = ref(false);

// v-model binding / 双向绑定
const visible = ref(props.modelValue);
watch(
  () => props.modelValue,
  (val) => {
    visible.value = val;
  },
);
watch(visible, (val) => {
  emit("update:modelValue", val);
});

// Form data / 表单数据
const form = reactive<CreateDramaRequest>({
  title: "",
  description: "",
  style: "ghibli",
  custom_style: "",
});

// Validation rules / 验证规则
const rules: FormRules = {
  title: [
    { required: true, message: "请输入项目标题", trigger: "blur" },
    {
      min: 1,
      max: 100,
      message: "标题长度在 1 到 100 个字符",
      trigger: "blur",
    },
  ],
  style: [{ required: true, message: "请选择风格", trigger: "change" }],
};

// Reset form when dialog closes / 关闭时重置表单
const handleClosed = () => {
  form.title = "";
  form.description = "";
  form.custom_style = "";
  formRef.value?.resetFields();
};

// Close dialog / 关闭弹窗
const handleClose = () => {
  visible.value = false;
};

// Submit form / 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (valid) {
      if (form.style === "custom" && !form.custom_style?.trim()) {
        ElMessage.warning("请填写自定义风格描述");
        return;
      }
      loading.value = true;
      try {
        const drama = await dramaAPI.create(form);
        ElMessage.success("创建成功");
        visible.value = false;
        emit("created", drama.id);
        // Navigate to drama detail page / 跳转到短剧详情页
        router.push(`/dramas/${drama.id}`);
      } catch (error: any) {
        ElMessage.error(error.message || "创建失败");
      } finally {
        loading.value = false;
      }
    }
  });
};
</script>

<style scoped>
/* ========================================
   Dialog Styles / 弹窗样式
   ======================================== */
.create-dialog :deep(.el-dialog) {
  border-radius: var(--radius-xl);
}

.create-dialog :deep(.el-dialog__header) {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-primary);
  margin-right: 0;
}

.create-dialog :deep(.el-dialog__title) {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.create-dialog :deep(.el-dialog__body) {
  padding: 1.5rem;
}

.dialog-desc {
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

/* ========================================
   Form Styles / 表单样式
   ======================================== */
.create-form :deep(.el-form-item) {
  margin-bottom: 1.25rem;
}

.create-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.create-form :deep(.el-input__wrapper),
.create-form :deep(.el-textarea__inner) {
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
  transition: all var(--transition-fast);
}

.create-form :deep(.el-input__wrapper:hover),
.create-form :deep(.el-textarea__inner:hover) {
  box-shadow: 0 0 0 1px var(--border-secondary) inset;
}

.create-form :deep(.el-input__wrapper.is-focus),
.create-form :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 2px var(--accent) inset;
}

.create-form :deep(.el-input__inner),
.create-form :deep(.el-textarea__inner) {
  color: var(--text-primary);
}

.create-form :deep(.el-input__inner::placeholder),
.create-form :deep(.el-textarea__inner::placeholder) {
  color: var(--text-muted);
}

.create-form :deep(.el-input__count) {
  color: var(--text-muted);
  background: transparent;
}

/* ========================================
   Footer Styles / 底部样式
   ======================================== */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.dialog-footer .el-button {
  min-width: 100px;
}
</style>
