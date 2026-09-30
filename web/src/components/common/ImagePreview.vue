<template>
  <div class="image-preview-wrapper">
    <!-- 缩略图 -->
    <div
      class="thumbnail-container"
      @click="handlePreview"
      :class="{ 'has-image': hasImage }"
    >
      <img v-if="hasImage" :src="imageUrl" :alt="alt" class="thumbnail-image" />
      <div v-else class="no-image-placeholder">
        <el-icon :size="size / 2"><Picture /></el-icon>
        <span v-if="showPlaceholderText">{{ placeholderText }}</span>
      </div>
    </div>

    <!-- 放大预览对话框 -->
    <el-dialog
      v-model="previewVisible"
      :width="dialogWidth"
      align-center
      :show-close="true"
      class="image-preview-dialog"
      @close="handleClose"
    >
      <template #header>
        <div class="preview-header">
          <span class="preview-title">{{ alt || "图片预览" }}</span>
        </div>
      </template>

      <div class="preview-content">
        <img :src="imageUrl" :alt="alt" class="preview-image" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { Picture } from "@element-plus/icons-vue";

interface Props {
  imageUrl?: string;
  alt?: string;
  size?: number;
  placeholderText?: string;
  showPlaceholderText?: boolean;
  dialogWidth?: string;
}

const props = withDefaults(defineProps<Props>(), {
  imageUrl: "",
  alt: "",
  size: 120,
  placeholderText: "暂无图片",
  showPlaceholderText: true,
  dialogWidth: "800px",
});

const previewVisible = ref(false);

const hasImage = computed(() => {
  return props.imageUrl && props.imageUrl.trim() !== "";
});

const handlePreview = () => {
  if (hasImage.value) {
    previewVisible.value = true;
  }
};

const handleClose = () => {
  previewVisible.value = false;
};
</script>

<style scoped>
.image-preview-wrapper {
  display: inline-block;
  width: 100%;
  height: 100%;
}

.thumbnail-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  transition: all var(--transition-fast);
}

.thumbnail-container.has-image {
  cursor: pointer;
}

.thumbnail-container.has-image:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.no-image-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-muted);
  padding: 16px;
  text-align: center;
}

.no-image-placeholder span {
  font-size: 12px;
}

/* 预览对话框样式 */
.image-preview-dialog :deep(.el-dialog) {
  border-radius: var(--radius-xl);
  background: var(--bg-primary);
}

.image-preview-dialog :deep(.el-dialog__header) {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-primary);
  margin-right: 0;
}

.image-preview-dialog :deep(.el-dialog__body) {
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #000;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.preview-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.preview-content {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  max-height: 80vh;
}

.preview-image {
  max-width: 100%;
  max-height: 80vh;
  object-fit: contain;
  display: block;
}
</style>
