<template>
  <!-- 阶段 1: 生成图片 -->
  <el-card class="workflow-card">
    <div class="stage-body">
      <!-- 角色图片生成 -->
      <div class="image-gen-section">
        <div class="section-header">
          <div class="section-title">
            <h3>
              <el-icon><User /></el-icon>
              {{ $t("workflow.characterImages") }}
            </h3>
            <el-alert type="info" :closable="false" style="margin: 0">
              {{
                $t("workflow.characterCount", { count: charactersCount })
              }}
            </el-alert>
          </div>
          <div class="section-actions">
            <el-checkbox
              v-model="selectAllCharacters"
              @change="$emit('toggle-select-all-characters')"
              style="margin-right: 12px"
            >
              {{ $t("workflow.selectAll") }}
            </el-checkbox>
            <el-button
              type="primary"
              @click="$emit('batch-generate-character-images')"
              :loading="batchGeneratingCharacters"
              :disabled="selectedCharacterIds.length === 0"
              size="default"
            >
              {{ $t("workflow.batchGenerate") }} ({{
                selectedCharacterIds.length
              }})
            </el-button>
          </div>
        </div>

        <div class="character-image-list">
          <div
            v-for="char in currentEpisode?.characters"
            :key="char.id"
            class="character-item"
          >
            <el-card shadow="hover" class="fixed-card">
              <div class="card-header">
                <el-checkbox
                  v-model="selectedCharacterIds"
                  :value="char.id"
                  style="margin-right: 8px"
                />
                <div class="header-left">
                  <h4>{{ char.name }}</h4>
                  <el-tag size="small">{{ char.role }}</el-tag>
                </div>
                <el-button
                  type="danger"
                  size="small"
                  :icon="Delete"
                  circle
                  @click="$emit('delete-character', char.id)"
                  :title="$t('workflow.deleteCharacter')"
                />
              </div>

              <div class="card-image-container">
                <div v-if="hasImage(char)" class="char-image">
                  <el-image :src="getImageUrl(char)" fit="cover" />
                </div>
                <div
                  v-else-if="
                    char.image_generation_status === 'pending' ||
                    char.image_generation_status === 'processing' ||
                    generatingCharacterImages[char.id]
                  "
                  class="char-placeholder generating"
                >
                  <el-icon :size="64" class="rotating"
                    ><Loading
                  /></el-icon>
                  <span>{{ $t("common.generating") }}</span>
                  <el-tag
                    type="warning"
                    size="small"
                    style="margin-top: 8px"
                    >{{
                      char.image_generation_status === "pending"
                        ? $t("common.queuing")
                        : $t("common.processing")
                    }}</el-tag
                  >
                </div>
                <div
                  v-else-if="char.image_generation_status === 'failed'"
                  class="char-placeholder failed"
                >
                  <el-icon :size="64"><WarningFilled /></el-icon>
                  <span>{{ $t("common.generateFailed") }}</span>
                  <el-tag
                    type="danger"
                    size="small"
                    style="margin-top: 8px"
                    >{{ $t("common.clickToRegenerate") }}</el-tag
                  >
                </div>
                <div v-else class="char-placeholder">
                  <el-icon :size="64"><User /></el-icon>
                  <span>{{ $t("common.notGenerated") }}</span>
                </div>
              </div>

              <div class="card-actions">
                <el-tooltip
                  :content="$t('tooltip.editPrompt')"
                  placement="top"
                >
                  <el-button
                    size="small"
                    @click="$emit('open-prompt-dialog', char, 'character')"
                    :icon="Edit"
                    circle
                  />
                </el-tooltip>
                <el-tooltip
                  :content="$t('tooltip.aiGenerate')"
                  placement="top"
                >
                  <el-button
                    type="primary"
                    size="small"
                    @click="$emit('generate-character-image', char.id)"
                    :loading="generatingCharacterImages[char.id]"
                    :icon="MagicStick"
                    circle
                  />
                </el-tooltip>
                <el-tooltip
                  :content="$t('tooltip.uploadImage')"
                  placement="top"
                >
                  <el-button
                    size="small"
                    @click="$emit('upload-character-image', char.id)"
                    :icon="Upload"
                    circle
                  />
                </el-tooltip>
                <el-tooltip
                  :content="$t('tooltip.selectFromLibrary')"
                  placement="top"
                >
                  <el-button
                    size="small"
                    @click="$emit('select-from-library', char.id)"
                    :icon="Picture"
                    circle
                  />
                </el-tooltip>
                <el-tooltip
                  :content="$t('workflow.addToLibrary')"
                  placement="top"
                >
                  <el-button
                    size="small"
                    @click="$emit('add-to-character-library', char)"
                    :icon="FolderAdd"
                    :disabled="!char.image_url"
                    circle
                  />
                </el-tooltip>
              </div>
            </el-card>
          </div>
        </div>
      </div>

      <el-divider />

      <!-- 场景图片生成 -->
      <div class="image-gen-section">
        <div class="section-header">
          <div class="section-title">
            <h3>
              <el-icon><Place /></el-icon>
              {{ $t("workflow.sceneImages") }}
            </h3>
            <el-alert type="info" :closable="false" style="margin: 0">
              {{
                $t("workflow.sceneCount", {
                  count: currentEpisode?.scenes?.length || 0,
                })
              }}
            </el-alert>
          </div>
          <div class="section-actions">
            <!-- <el-button
                  :icon="Document"
                  @click="openExtractSceneDialog"
                  size="default"
                >
                  {{ $t("workflow.extractFromScript") }}
                </el-button> -->
            <el-checkbox
              v-model="selectAllScenes"
              @change="$emit('toggle-select-all-scenes')"
              style="margin-left: 12px; margin-right: 12px"
            >
              {{ $t("workflow.selectAll") }}
            </el-checkbox>
            <el-button
              type="primary"
              @click="$emit('batch-generate-scene-images')"
              :loading="batchGeneratingScenes"
              :disabled="selectedSceneIds.length === 0"
              size="default"
            >
              {{ $t("workflow.batchGenerateSelected") }} ({{
                selectedSceneIds.length
              }})
            </el-button>

            <el-button
              :icon="Plus"
              @click="$emit('open-add-scene-dialog')"
              size="default"
            >
              {{ $t("workflow.addScene") }}
            </el-button>
          </div>
        </div>

        <div class="scene-image-list">
          <div
            v-for="scene in currentEpisode?.scenes"
            :key="scene.id"
            class="scene-item"
          >
            <el-card shadow="hover" class="fixed-card">
              <div class="card-header">
                <el-checkbox
                  v-model="selectedSceneIds"
                  :value="scene.id"
                  style="margin-right: 8px"
                />
                <div class="header-left">
                  <h4>{{ scene.location }}</h4>
                  <el-tag size="small">{{ scene.time }}</el-tag>
                </div>
              </div>

              <div class="card-image-container">
                <div v-if="hasImage(scene)" class="scene-image">
                  <el-image :src="getImageUrl(scene)" fit="cover" />
                </div>
                <div
                  v-else-if="
                    scene.image_generation_status === 'pending' ||
                    scene.image_generation_status === 'processing' ||
                    generatingSceneImages[scene.id]
                  "
                  class="scene-placeholder generating"
                >
                  <el-icon :size="64" class="rotating"
                    ><Loading
                  /></el-icon>
                  <span>{{ $t("common.generating") }}</span>
                  <el-tag
                    type="warning"
                    size="small"
                    style="margin-top: 8px"
                    >{{
                      scene.image_generation_status === "pending"
                        ? $t("common.queuing")
                        : $t("common.processing")
                    }}</el-tag
                  >
                </div>
                <div
                  v-else-if="scene.image_generation_status === 'failed'"
                  class="scene-placeholder failed"
                  @click="$emit('generate-scene-image', scene.id)"
                  style="cursor: pointer"
                >
                  <el-icon :size="64"><WarningFilled /></el-icon>
                  <span>{{ $t("common.generateFailed") }}</span>
                  <el-tag
                    type="danger"
                    size="small"
                    style="margin-top: 8px"
                    >{{ $t("common.clickToRegenerate") }}</el-tag
                  >
                </div>
                <div v-else class="scene-placeholder">
                  <el-icon :size="64"><Place /></el-icon>
                  <span>{{ $t("common.notGenerated") }}</span>
                </div>
              </div>

              <div class="card-actions">
                <el-tooltip
                  :content="$t('tooltip.editPrompt')"
                  placement="top"
                >
                  <el-button
                    size="small"
                    @click="$emit('open-prompt-dialog', scene, 'scene')"
                    :icon="Edit"
                    circle
                  />
                </el-tooltip>
                <el-tooltip
                  :content="$t('tooltip.aiGenerate')"
                  placement="top"
                >
                  <el-button
                    type="primary"
                    size="small"
                    @click="$emit('generate-scene-image', scene.id)"
                    :loading="generatingSceneImages[scene.id]"
                    :icon="MagicStick"
                    circle
                  />
                </el-tooltip>
                <el-tooltip
                  :content="$t('tooltip.uploadImage')"
                  placement="top"
                >
                  <el-button
                    size="small"
                    @click="$emit('upload-scene-image', scene.id)"
                    :icon="Upload"
                    circle
                  />
                </el-tooltip>
              </div>
            </el-card>
          </div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import {
  User,
  Delete,
  Loading,
  WarningFilled,
  Edit,
  MagicStick,
  Upload,
  Picture,
  FolderAdd,
  Place,
  Plus,
} from "@element-plus/icons-vue";
import { getImageUrl, hasImage } from "@/utils/image";

defineProps<{
  currentEpisode: any;
  charactersCount: number;
  batchGeneratingCharacters: boolean;
  generatingCharacterImages: Record<number, boolean>;
  batchGeneratingScenes: boolean;
  generatingSceneImages: Record<string, boolean>;
}>();

defineEmits<{
  (e: "toggle-select-all-characters"): void;
  (e: "batch-generate-character-images"): void;
  (e: "delete-character", id: number): void;
  (e: "open-prompt-dialog", item: any, type: "character" | "scene"): void;
  (e: "generate-character-image", id: number): void;
  (e: "upload-character-image", id: number): void;
  (e: "select-from-library", id: number): void;
  (e: "add-to-character-library", character: any): void;
  (e: "toggle-select-all-scenes"): void;
  (e: "batch-generate-scene-images"): void;
  (e: "open-add-scene-dialog"): void;
  (e: "generate-scene-image", id: string): void;
  (e: "upload-scene-image", id: string): void;
}>();

// 选择状态（与容器内同名单向数据源双向同步）
const selectAllCharacters = defineModel<boolean>("selectAllCharacters", {
  required: true,
});
const selectedCharacterIds = defineModel<number[]>("selectedCharacterIds", {
  required: true,
});
const selectAllScenes = defineModel<boolean>("selectAllScenes", {
  required: true,
});
const selectedSceneIds = defineModel<number[]>("selectedSceneIds", {
  required: true,
});
</script>

<style scoped lang="scss">
@use "./step-common.scss";

.workflow-card {
  height: calc(100% - 24px);
  margin: 12px;
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-primary);

  :deep(.el-card__body) {
    padding: 0;
  }
}

.image-gen-section {
  margin-bottom: 32px;

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    padding: 16px;
    background: var(--bg-secondary);
    // border-radius: 8px;
    // border: 1px solid var(--border-primary);

    .section-title {
      display: flex;
      align-items: center;
      gap: 16px;

      h3 {
        display: flex;
        align-items: center;
        gap: 8px;
        margin: 0;
        font-size: 16px;
        font-weight: 600;
        color: var(--text-primary);

        .el-icon {
          color: var(--accent);
          font-size: 18px;
        }
      }

      .el-alert {
        border-radius: 4px;
      }
    }

    .section-actions {
      display: flex;
      align-items: center;
    }
  }
}

.character-image-list,
.scene-image-list {
  padding: 5px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
  margin-top: 16px;

  .character-item,
  .scene-item {
    min-height: 360px;
  }
}
</style>
