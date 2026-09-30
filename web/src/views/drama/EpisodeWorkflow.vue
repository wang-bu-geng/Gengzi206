<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <!-- Page Sub Header -->
      <div class="page-sub-header">
        <div class="sub-header-left">
          <el-button text @click="$router.back()" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
            <span>{{ $t("workflow.backToProject") }}</span>
          </el-button>
          <h1 class="header-title">
            {{ $t("workflow.episodeProduction", { number: episodeNumber }) }}
          </h1>
        </div>
        <div class="sub-header-center">
          <div class="custom-steps">
            <div
              class="step-item"
              :class="{ active: currentStep >= 0, current: currentStep === 0 }"
            >
              <div class="step-circle">1</div>
              <span class="step-text">{{ $t("workflow.steps.content") }}</span>
            </div>
            <el-icon class="step-arrow"><ArrowRight /></el-icon>
            <div
              class="step-item"
              :class="{ active: currentStep >= 1, current: currentStep === 1 }"
            >
              <div class="step-circle">2</div>
              <span class="step-text">{{
                $t("workflow.steps.generateImages")
              }}</span>
            </div>
            <el-icon class="step-arrow"><ArrowRight /></el-icon>
            <div
              class="step-item"
              :class="{ active: currentStep >= 2, current: currentStep === 2 }"
            >
              <div class="step-circle">3</div>
              <span class="step-text">{{
                $t("workflow.steps.splitStoryboard")
              }}</span>
            </div>
          </div>
        </div>
        <div class="sub-header-right">
          <el-button
            :icon="Setting"
            @click="showModelConfigDialog"
            :title="$t('workflow.modelConfig')"
          >
            图文配置
          </el-button>
        </div>
      </div>

      <div class="content-container">
        <!-- 阶段 0: 章节内容 + 提取角色场景 -->
        <ScriptStep
          v-show="currentStep === 0"
          v-model:script-content="scriptContent"
          :episode-number="episodeNumber"
          :has-script="hasScript"
          :generating-script="generatingScript"
          :current-episode="currentEpisode"
          :has-extracted-data="hasExtractedData"
          :has-characters="hasCharacters"
          :characters-count="charactersCount"
          @save="saveChapterScript"
        />

        <!-- 阶段 1: 生成图片 -->
        <ImageStep
          v-show="currentStep === 1"
          v-model:select-all-characters="selectAllCharacters"
          v-model:selected-character-ids="selectedCharacterIds"
          v-model:select-all-scenes="selectAllScenes"
          v-model:selected-scene-ids="selectedSceneIds"
          :current-episode="currentEpisode"
          :characters-count="charactersCount"
          :batch-generating-characters="batchGeneratingCharacters"
          :generating-character-images="generatingCharacterImages"
          :batch-generating-scenes="batchGeneratingScenes"
          :generating-scene-images="generatingSceneImages"
          @toggle-select-all-characters="toggleSelectAllCharacters"
          @batch-generate-character-images="batchGenerateCharacterImages"
          @delete-character="deleteCharacter"
          @open-prompt-dialog="openPromptDialog"
          @generate-character-image="generateCharacterImage"
          @upload-character-image="uploadCharacterImage"
          @select-from-library="selectFromLibrary"
          @add-to-character-library="addToCharacterLibrary"
          @toggle-select-all-scenes="toggleSelectAllScenes"
          @batch-generate-scene-images="batchGenerateSceneImages"
          @open-add-scene-dialog="openAddSceneDialog"
          @generate-scene-image="generateSceneImage"
          @upload-scene-image="uploadSceneImage"
        />

        <!-- 阶段 2: 拆分分镜 -->
        <StoryboardStep
          v-show="currentStep === 2"
          :current-episode="currentEpisode"
          :batch-generating-video="batchGeneratingVideo"
          :overall-video-progress="overallVideoProgress"
          :batch-video-done="batchVideoDone"
          :batch-video-total="batchVideoTotal"
          :generating-video="generatingVideo"
          :video-progress="videoProgress"
          :generating-shots="generatingShots"
          :task-progress="taskProgress"
          :task-message="taskMessage"
          @batch-generate-videos="batchGenerateVideos"
          @edit-shot="editShot"
          @generate-video-for-shot="generateVideoForShot"
          @download-shot-video="downloadShotVideo"
          @generate-shots="generateShots"
        />
      </div>

      <div class="actions-container">
        <div class="action-buttons" v-show="currentStep === 0">
          <el-button
            type="primary"
            size="large"
            @click="handleExtractCharactersAndBackgrounds"
            :loading="extractingCharactersAndBackgrounds"
            :disabled="!hasScript"
          >
            <el-icon><MagicStick /></el-icon>
            {{
              hasExtractedData
                ? $t("workflow.reExtract")
                : $t("workflow.extractCharactersAndScenes")
            }}
          </el-button>
          <el-button
            type="success"
            size="large"
            @click="nextStep"
            :disabled="!hasExtractedData"
          >
            {{ $t("workflow.nextStepGenerateImages") }}
            <el-icon><ArrowRight /></el-icon>
          </el-button>
          <div v-if="!hasExtractedData" style="margin-top: 8px">
            <el-alert
              type="warning"
              :closable="false"
              style="display: inline-block"
            >
              <template #title>
                <span style="font-size: 12px">
                  {{ $t("workflow.extractWarning") }}
                </span>
              </template>
            </el-alert>
          </div>
        </div>

        <div class="action-buttons" v-show="currentStep === 1">
          <el-button size="large" @click="prevStep">
            <el-icon><ArrowLeft /></el-icon>
            {{ $t("workflow.prevStep") }}
          </el-button>
          <el-button
            type="success"
            size="large"
            @click="nextStep"
            :disabled="!allImagesGenerated"
          >
            {{ $t("workflow.nextStepSplitShots") }}
            <el-icon><ArrowRight /></el-icon>
          </el-button>
          <div v-if="!allImagesGenerated" style="margin-top: 8px">
            <el-alert
              type="warning"
              :closable="false"
              style="display: inline-block"
            >
              <template #title>
                <span style="font-size: 12px">
                  {{ $t("workflow.generateAllImagesFirst") }}
                </span>
              </template>
            </el-alert>
          </div>
        </div>

        <div class="action-buttons" v-show="currentStep === 2">
          <el-button size="large" @click="prevStep">
            <el-icon><ArrowLeft /></el-icon>
            {{ $t("workflow.prevStep") }}
          </el-button>
          <el-button size="large" @click="regenerateShots" :icon="MagicStick">
            {{ $t("workflow.reSplitShots") }}
          </el-button>
          <el-button type="success" size="large" @click="goBack">
            <el-icon><Check /></el-icon>
            {{ $t("workflow.done") }}
          </el-button>
        </div>
      </div>
    </div>

    <div class="components-box">
      <!-- 镜头编辑对话框 -->
      <el-dialog
        v-model="shotEditDialogVisible"
        :title="$t('workflow.editShot')"
        width="800px"
        :close-on-click-modal="false"
      >
        <el-form v-if="editingShot" label-width="100px" size="default">
          <el-form-item :label="$t('workflow.shotTitle')">
            <el-input
              v-model="editingShot.title"
              :placeholder="$t('workflow.shotTitlePlaceholder')"
            />
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="8">
              <el-form-item :label="$t('workflow.shotType')">
                <el-select
                  v-model="editingShot.shot_type"
                  :placeholder="$t('workflow.selectShotType')"
                >
                  <el-option :label="$t('workflow.longShot')" value="远景" />
                  <el-option :label="$t('workflow.fullShot')" value="全景" />
                  <el-option :label="$t('workflow.mediumShot')" value="中景" />
                  <el-option :label="$t('workflow.closeUp')" value="近景" />
                  <el-option
                    :label="$t('workflow.extremeCloseUp')"
                    value="特写"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('workflow.cameraAngle')">
                <el-select
                  v-model="editingShot.angle"
                  :placeholder="$t('workflow.selectAngle')"
                >
                  <el-option :label="$t('workflow.eyeLevel')" value="平视" />
                  <el-option :label="$t('workflow.lowAngle')" value="仰视" />
                  <el-option :label="$t('workflow.highAngle')" value="俯视" />
                  <el-option :label="$t('workflow.sideView')" value="侧面" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('workflow.cameraMovement')">
                <el-select
                  v-model="editingShot.movement"
                  :placeholder="$t('workflow.selectMovement')"
                >
                  <el-option
                    :label="$t('workflow.staticShot')"
                    value="固定镜头"
                  />
                  <el-option :label="$t('workflow.pushIn')" value="推镜" />
                  <el-option :label="$t('workflow.pullOut')" value="拉镜" />
                  <el-option :label="$t('workflow.followShot')" value="跟镜" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item :label="$t('workflow.location')">
                <el-input
                  v-model="editingShot.location"
                  :placeholder="$t('workflow.locationPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('workflow.time')">
                <el-input
                  v-model="editingShot.time"
                  :placeholder="$t('workflow.timeSetting')"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item :label="$t('workflow.shotDescription')">
            <el-input
              v-model="editingShot.description"
              type="textarea"
              :rows="2"
              :placeholder="$t('workflow.shotDescriptionPlaceholder')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.actionDescription')">
            <el-input
              v-model="editingShot.action"
              type="textarea"
              :rows="3"
              :placeholder="$t('workflow.detailedAction')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.dialogue')">
            <el-input
              v-model="editingShot.dialogue"
              type="textarea"
              :rows="2"
              :placeholder="$t('workflow.characterDialogue')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.result')">
            <el-input
              v-model="editingShot.result"
              type="textarea"
              :rows="2"
              :placeholder="$t('workflow.actionResult')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.atmosphere')">
            <el-input
              v-model="editingShot.atmosphere"
              type="textarea"
              :rows="2"
              :placeholder="$t('workflow.atmosphereDescription')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.imagePrompt')">
            <el-input
              v-model="editingShot.image_prompt"
              type="textarea"
              :rows="3"
              :placeholder="$t('workflow.imagePromptPlaceholder')"
            />
          </el-form-item>

          <el-form-item :label="$t('workflow.videoPrompt')">
            <el-input
              v-model="editingShot.video_prompt"
              type="textarea"
              :rows="3"
              :placeholder="$t('workflow.videoPromptPlaceholder')"
            />
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item :label="$t('workflow.bgmHint')">
                <el-input
                  v-model="editingShot.bgm_prompt"
                  :placeholder="$t('workflow.bgmAtmosphere')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('workflow.soundEffect')">
                <el-input
                  v-model="editingShot.sound_effect"
                  :placeholder="$t('workflow.soundEffectDescription')"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item :label="$t('workflow.durationSeconds')">
            <el-input-number
              v-model="editingShot.duration"
              :min="1"
              :max="60"
            />
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="shotEditDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button
            type="primary"
            @click="saveShotEdit"
            :loading="savingShot"
            >{{ $t("common.save") }}</el-button
          >
        </template>
      </el-dialog>

      <!-- 提示词编辑对话框 -->
      <el-dialog
        v-model="promptDialogVisible"
        :title="$t('workflow.editPrompt')"
        width="600px"
      >
        <el-form label-width="80px">
          <el-form-item :label="$t('common.name')">
            <el-input v-model="currentEditItem.name" disabled />
          </el-form-item>
          <el-form-item
            v-if="currentEditType === 'scene'"
            :label="$t('workflow.time')"
          >
            <el-input
              v-model="currentEditItem.time"
              :placeholder="$t('workflow.timePlaceholder')"
            />
          </el-form-item>
          <el-form-item :label="$t('workflow.imagePrompt')">
            <el-input
              v-model="editPrompt"
              type="textarea"
              :rows="6"
              :placeholder="$t('workflow.imagePromptPlaceholder')"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="promptDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="savePrompt">{{
            $t("common.save")
          }}</el-button>
        </template>
      </el-dialog>

      <!-- 角色库选择对话框 -->
      <el-dialog
        v-model="libraryDialogVisible"
        :title="$t('workflow.selectFromLibrary')"
        width="800px"
      >
        <div class="library-grid">
          <div
            v-for="item in libraryItems"
            :key="item.id"
            class="library-item"
            @click="selectLibraryItem(item)"
          >
            <el-image :src="getImageUrl(item)" fit="cover" />
            <div class="library-item-name">{{ item.name }}</div>
          </div>
        </div>
        <div v-if="libraryItems.length === 0" class="empty-library">
          <el-empty :description="$t('workflow.emptyLibrary')" />
        </div>
      </el-dialog>

      <!-- AI模型配置对话框 -->
      <el-dialog
        v-model="modelConfigDialogVisible"
        :title="$t('workflow.aiModelConfig')"
        width="600px"
        :close-on-click-modal="false"
      >
        <el-form label-width="120px">
          <el-form-item :label="$t('workflow.textGenModel')">
            <el-select
              v-model="selectedTextModel"
              :placeholder="$t('workflow.selectTextModel')"
              style="width: 100%"
            >
              <el-option
                v-for="model in textModels"
                :key="model.modelName"
                :label="model.modelName"
                :value="model.modelName"
              />
            </el-select>
            <div class="model-tip">
              {{ $t("workflow.textModelTip") }}
            </div>
          </el-form-item>

          <el-form-item :label="$t('workflow.imageGenModel')">
            <el-select
              v-model="selectedImageModel"
              :placeholder="$t('workflow.selectImageModel')"
              style="width: 100%"
            >
              <el-option
                v-for="model in imageModels"
                :key="model.modelName"
                :label="model.modelName"
                :value="model.modelName"
              />
            </el-select>
            <div class="model-tip">
              {{ $t("workflow.modelConfigTip") }}
            </div>
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="modelConfigDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="saveModelConfig">{{
            $t("common.saveConfig")
          }}</el-button>
        </template>
      </el-dialog>

      <!-- 图片上传对话框 -->
      <el-dialog
        v-model="uploadDialogVisible"
        :title="$t('tooltip.uploadImage')"
        width="500px"
      >
        <el-upload
          class="upload-area"
          drag
          :action="uploadAction"
          :headers="uploadHeaders"
          :on-success="handleUploadSuccess"
          :on-error="handleUploadError"
          :show-file-list="false"
          accept="image/jpeg,image/png,image/jpg"
        >
          <el-icon class="el-icon--upload"><Upload /></el-icon>
          <div class="el-upload__text">
            {{ $t("workflow.dragFilesHere")
            }}<em>{{ $t("workflow.clickToUpload") }}</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              {{ $t("workflow.uploadFormatTip") }}
            </div>
          </template>
        </el-upload>
      </el-dialog>

      <!-- 添加场景对话框 -->
      <el-dialog
        v-model="addSceneDialogVisible"
        :title="$t('workflow.addScene')"
        width="600px"
      >
        <el-form :model="newScene" label-width="100px">
          <el-form-item :label="$t('workflow.sceneImage')">
            <el-upload
              class="avatar-uploader"
              :action="`/api/v1/upload/image`"
              :show-file-list="false"
              :on-success="handleSceneImageSuccess"
              :before-upload="beforeAvatarUpload"
            >
              <img
                v-if="hasImage(newScene)"
                :src="getImageUrl(newScene)"
                class="avatar"
                style="width: 160px; height: 90px; object-fit: cover"
              />
              <el-icon
                v-else
                class="avatar-uploader-icon"
                style="
                  border: 1px dashed #d9d9d9;
                  border-radius: 6px;
                  cursor: pointer;
                  position: relative;
                  overflow: hidden;
                  width: 160px;
                  height: 90px;
                  font-size: 28px;
                  color: #8c939d;
                  text-align: center;
                  line-height: 90px;
                "
                ><Plus
              /></el-icon>
            </el-upload>
          </el-form-item>
          <el-form-item :label="$t('workflow.sceneName')">
            <el-input
              v-model="newScene.location"
              :placeholder="$t('workflow.sceneNamePlaceholder')"
            />
          </el-form-item>
          <el-form-item :label="$t('workflow.time')">
            <el-input
              v-model="newScene.time"
              :placeholder="$t('workflow.timePlaceholder')"
            />
          </el-form-item>
          <el-form-item :label="$t('workflow.sceneDescription')">
            <el-input
              v-model="newScene.prompt"
              type="textarea"
              :rows="4"
              :placeholder="$t('workflow.sceneDescriptionPlaceholder')"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="addSceneDialogVisible = false">{{
            $t("common.cancel")
          }}</el-button>
          <el-button type="primary" @click="saveScene">{{
            $t("common.confirm")
          }}</el-button>
        </template>
      </el-dialog>

      <!-- 从剧本提取场景对话框 -->
      <el-dialog
        v-model="extractScenesDialogVisible"
        :title="$t('workflow.extractSceneDialogTitle')"
        width="500px"
      >
        <el-alert type="info" :closable="false" style="margin-bottom: 16px">
          {{ $t("workflow.extractSceneDialogTip") }}
        </el-alert>
        <template #footer>
          <el-button @click="extractScenesDialogVisible = false">
            {{ $t("common.cancel") }}
          </el-button>
          <el-button
            type="primary"
            @click="handleExtractScenes"
            :loading="extractingScenes"
          >
            {{ $t("workflow.startExtract") }}
          </el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  MagicStick,
  ArrowRight,
  ArrowLeft,
  Upload,
  Setting,
  Plus,
  Check,
} from "@element-plus/icons-vue";
import { dramaAPI } from "@/api/drama";
import { characterLibraryAPI } from "@/api/character-library";
import type { Drama } from "@/types/drama";
import { getImageUrl, hasImage } from "@/utils/image";
import { useModelConfig } from "./composables/useModelConfig";
import { useStoryboardSplit } from "./composables/useStoryboardSplit";
import { useVideoGeneration } from "./composables/useVideoGeneration";
import { useCharacterExtraction } from "./composables/useCharacterExtraction";
import { useImageGeneration } from "./composables/useImageGeneration";
import { useTerminalBridge } from "./composables/useTerminalBridge";
import ScriptStep from "./components/ScriptStep.vue";
import ImageStep from "./components/ImageStep.vue";
import StoryboardStep from "./components/StoryboardStep.vue";

const route = useRoute();
const router = useRouter();
const { t: $t } = useI18n();
const dramaId = route.params.id as string;
const episodeNumber = parseInt(route.params.episodeNumber as string);

const drama = ref<Drama>();

let isUnmounted = false;
const activeTimers: number[] = [];
const safeTimeout = (callback: () => void, delay: number) => {
  const timer = window.setTimeout(() => {
    const idx = activeTimers.indexOf(timer);
    if (idx > -1) activeTimers.splice(idx, 1);
    if (!isUnmounted) callback();
  }, delay);
  activeTimers.push(timer);
  return timer;
};
const clearAllTimers = () => {
  activeTimers.forEach((t) => clearTimeout(t));
  activeTimers.length = 0;
};

// 生成 localStorage key
const getStepStorageKey = () =>
  `episode_workflow_step_${dramaId}_${episodeNumber}`;

// 从 localStorage 恢复步骤，如果没有则默认为 0
const savedStep = localStorage.getItem(getStepStorageKey());
const currentStep = ref(savedStep ? parseInt(savedStep) : 0);
const scriptContent = ref("");
const generatingScript = ref(false);

// 对话框状态
const promptDialogVisible = ref(false);
const libraryDialogVisible = ref(false);
const uploadDialogVisible = ref(false);
const addSceneDialogVisible = ref(false);
const extractScenesDialogVisible = ref(false);
const currentEditItem = ref<any>({ name: "" });
const currentEditType = ref<"character" | "scene">("character");
const editPrompt = ref("");
const libraryItems = ref<any[]>([]);
const currentUploadTarget = ref<any>(null);

// 添加场景相关
const newScene = ref<any>({
  location: "",
  time: "",
  prompt: "",
  image_url: "",
  local_path: "",
});
const extractingScenes = ref(false);
const uploadAction = computed(() => "/api/v1/upload/image");
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem("token")}`,
}));

const hasScript = computed(() => {
  const currentEp = currentEpisode.value;
  return (
    currentEp && currentEp.script_content && currentEp.script_content.length > 0
  );
});

const currentEpisode = computed(() => {
  if (!drama.value?.episodes) return null;
  return drama.value.episodes.find((ep) => ep.episode_number === episodeNumber);
});

const hasCharacters = computed(() => {
  return (
    currentEpisode.value?.characters &&
    currentEpisode.value.characters.length > 0
  );
});

const charactersCount = computed(() => {
  return currentEpisode.value?.characters?.length || 0;
});

const hasExtractedData = computed(() => {
  const hasScenes =
    currentEpisode.value?.scenes && currentEpisode.value.scenes.length > 0;
  // 只要有角色或场景，就认为已经提取过数据
  return hasCharacters.value || hasScenes;
});

const allImagesGenerated = computed(() => {
  // 如果没有提取任何数据，允许跳过（可能是空章节或用户想直接进入拆解分镜）
  if (!hasExtractedData.value) return true;

  const characters = currentEpisode.value?.characters || [];
  const scenes = currentEpisode.value?.scenes || [];

  // 如果角色和场景都为空，允许跳过
  if (characters.length === 0 && scenes.length === 0) return true;

  // 检查所有有数据的项是否都已生成图片
  const allCharsHaveImages =
    characters.length === 0 || characters.every((char) => char.image_url);
  const allScenesHaveImages =
    scenes.length === 0 || scenes.every((scene) => scene.image_url);

  return allCharsHaveImages && allScenesHaveImages;
});

const goBack = () => {
  // 使用 replace 避免在历史记录中留下当前页面
  router.replace(`/dramas/${dramaId}`);
};

const nextStep = () => {
  if (currentStep.value < 3) {
    currentStep.value++;
  }
};

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--;
  }
};

const loadDramaData = async () => {
  try {
    const data = await dramaAPI.get(dramaId);
    drama.value = data;

    if (!hasScript.value) {
      scriptContent.value = "";
      // 如果没有剧本内容，重置到第一步
      currentStep.value = 0;
    }

    // 检查是否有生成中的角色或场景，自动启动轮询
    await checkAndStartPolling();
  } catch (error: any) {
    ElMessage.error(error.message || "加载项目数据失败");
  }
};

// ===== 抽离的组合式逻辑 =====
// AI 模型配置（文本/图片模型选择）
const {
  textModels,
  imageModels,
  selectedTextModel,
  selectedImageModel,
  modelConfigDialogVisible,
  loadAIConfigs,
  showModelConfigDialog,
  saveModelConfig,
  loadSavedModelConfig,
} = useModelConfig(dramaId);

// 阶段 2：AI 拆分分镜（异步任务 + 轮询）
const {
  generatingShots,
  taskProgress,
  taskMessage,
  generateShots,
  regenerateShots,
} = useStoryboardSplit({
  drama,
  currentEpisode,
  episodeNumber,
  selectedTextModel,
  loadDramaData,
  isUnmounted: () => isUnmounted,
});

// 阶段 2：分镜视频生成（单个/批量）+ 真实进度
const {
  generatingVideo,
  videoProgress,
  failedVideoShots,
  batchGeneratingVideo,
  batchVideoTotal,
  batchVideoDone,
  overallVideoProgress,
  generateVideoForShot,
  batchGenerateVideos,
  downloadShotVideo,
} = useVideoGeneration({
  dramaId,
  currentEpisode,
  loadDramaData,
  isUnmounted: () => isUnmounted,
  safeTimeout,
});

// 阶段 0：从剧本提取角色与场景（并行任务 + 轮询）
const {
  extractingCharactersAndBackgrounds,
  handleExtractCharactersAndBackgrounds,
} = useCharacterExtraction({
  dramaId,
  currentEpisode,
  selectedTextModel,
  hasExtractedData,
  loadDramaData,
  isUnmounted: () => isUnmounted,
  safeTimeout,
});

// 阶段 1：角色/场景图片生成（单个/批量）+ 重进页面恢复轮询
const {
  batchGeneratingCharacters,
  batchGeneratingScenes,
  generatingCharacterImages,
  generatingSceneImages,
  selectedCharacterIds,
  selectedSceneIds,
  selectAllCharacters,
  selectAllScenes,
  checkAndStartPolling,
  generateCharacterImage,
  toggleSelectAllCharacters,
  toggleSelectAllScenes,
  batchGenerateCharacterImages,
  generateSceneImage,
  batchGenerateSceneImages,
} = useImageGeneration({
  dramaId,
  currentEpisode,
  selectedImageModel,
  loadDramaData,
  isUnmounted: () => isUnmounted,
  safeTimeout,
});

// 生成终端：只观察上述真实任务状态，派生日志流与进度，不介入执行逻辑
useTerminalBridge({
  episodeNumber,
  currentEpisode,
  batchGeneratingVideo,
  batchVideoTotal,
  batchVideoDone,
  overallVideoProgress,
  videoProgress,
  generatingVideo,
  failedVideoShots,
  generatingShots,
  taskProgress,
  taskMessage,
  batchGeneratingCharacters,
  batchGeneratingScenes,
  generatingCharacterImages,
  generatingSceneImages,
  selectedCharacterIds,
  selectedSceneIds,
});

const saveChapterScript = async () => {
  try {
    const existingEpisodes = drama.value?.episodes || [];

    // 查找当前章节
    const episodeIndex = existingEpisodes.findIndex(
      (ep) => ep.episode_number === episodeNumber,
    );

    let updatedEpisodes;
    if (episodeIndex >= 0) {
      // 更新已有章节
      updatedEpisodes = [...existingEpisodes];
      updatedEpisodes[episodeIndex] = {
        ...updatedEpisodes[episodeIndex],
        script_content: scriptContent.value,
      };
    } else {
      // 创建新章节
      const newEpisode = {
        episode_number: episodeNumber,
        title: `第${episodeNumber}集`,
        script_content: scriptContent.value,
      };
      updatedEpisodes = [...existingEpisodes, newEpisode];
    }

    await dramaAPI.saveEpisodes(dramaId, updatedEpisodes);
    ElMessage.success("章节保存成功！");
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
  }
};

const editCurrentEpisodeScript = () => {
  scriptContent.value = currentEpisode.value?.script_content || "";
};

const shotEditDialogVisible = ref(false);
const editingShot = ref<any>(null);
const editingShotIndex = ref<number>(-1);
const savingShot = ref(false);

const editShot = (shot: any, index: number) => {
  editingShot.value = { ...shot };
  editingShotIndex.value = index;
  shotEditDialogVisible.value = true;
};

const saveShotEdit = async () => {
  if (!editingShot.value) return;

  try {
    savingShot.value = true;

    // 调用API更新镜头
    await dramaAPI.updateStoryboard(
      editingShot.value.id.toString(),
      editingShot.value,
    );

    // 更新本地数据
    if (currentEpisode.value?.storyboards) {
      currentEpisode.value.storyboards[editingShotIndex.value] = {
        ...editingShot.value,
      };
    }

    ElMessage.success("镜头修改成功");
    shotEditDialogVisible.value = false;
  } catch (error: any) {
    ElMessage.error("保存失败: " + (error.message || "未知错误"));
  } finally {
    savingShot.value = false;
  }
};

// 对话框相关方法
const openPromptDialog = (item: any, type: "character" | "scene") => {
  currentEditItem.value = item;
  currentEditItem.value.name = item.name || item.location;
  currentEditType.value = type;
  editPrompt.value = item.prompt || item.appearance || item.description || "";
  promptDialogVisible.value = true;
};

const savePrompt = async () => {
  try {
    if (currentEditType.value === "character") {
      await characterLibraryAPI.updateCharacter(currentEditItem.value.id, {
        appearance: editPrompt.value,
      });
      await generateCharacterImage(currentEditItem.value.id);
    } else {
      // 保存场景提示词和时间（合并到一个 API 调用）
      await dramaAPI.updateScene(currentEditItem.value.id.toString(), {
        prompt: editPrompt.value,
        time: currentEditItem.value.time || "",
      });

      ElMessage.success("保存成功");
      await loadDramaData();
    }
    promptDialogVisible.value = false;
  } catch (error: any) {
    ElMessage.error(error.message || "保存失败");
  }
};

const uploadCharacterImage = (characterId: number) => {
  currentUploadTarget.value = { id: characterId, type: "character" };
  uploadDialogVisible.value = true;
};

const uploadSceneImage = (sceneId: string) => {
  currentUploadTarget.value = { id: sceneId, type: "scene" };
  uploadDialogVisible.value = true;
};

const selectFromLibrary = async (characterId: number) => {
  try {
    const result = await characterLibraryAPI.list({ page_size: 50 });
    libraryItems.value = result.items || [];
    currentUploadTarget.value = characterId;
    libraryDialogVisible.value = true;
  } catch (error: any) {
    ElMessage.error(error.message || $t("workflow.loadLibraryFailed"));
  }
};

const addToCharacterLibrary = async (character: any) => {
  if (!character.image_url) {
    ElMessage.warning($t("workflow.generateImageFirst"));
    return;
  }

  try {
    await ElMessageBox.confirm(
      $t("workflow.addToLibraryConfirm", { name: character.name }),
      $t("workflow.addToLibrary"),
      {
        confirmButtonText: $t("common.confirm"),
        cancelButtonText: $t("common.cancel"),
        type: "info",
      },
    );

    await characterLibraryAPI.addCharacterToLibrary(character.id.toString());
    ElMessage.success($t("workflow.addedToLibrary"));
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || $t("workflow.addFailed"));
    }
  }
};

const selectLibraryItem = async (item: any) => {
  try {
    if (currentUploadTarget.value?.type === "character") {
      await characterLibraryAPI.applyFromLibrary(
        currentUploadTarget.value.id.toString(),
        item.id,
      );
      ElMessage.success("应用角色形象成功！");
      await loadDramaData();
      libraryDialogVisible.value = false;
    }
  } catch (error: any) {
    ElMessage.error(error.message || "应用失败");
  }
};

const handleUploadSuccess = async (response: any) => {
  try {
    const imageUrl = response.url || response.data?.url;
    const localPath = response.local_path || response.data?.local_path;

    if (!imageUrl && !localPath) {
      ElMessage.error("上传失败：未获取到图片地址");
      return;
    }

    if (currentUploadTarget.value?.type === "character") {
      await characterLibraryAPI.updateCharacter(
        currentUploadTarget.value.id.toString(),
        {
          image_url: imageUrl,
          local_path: localPath,
        },
      );
      ElMessage.success("上传成功！");
    } else if (currentUploadTarget.value?.type === "scene") {
      // 更新场景图片
      await dramaAPI.updateScene(currentUploadTarget.value.id.toString(), {
        image_url: imageUrl,
        local_path: localPath,
      });
      ElMessage.success($t("workflow.sceneImageUploadSuccess"));
    }

    await loadDramaData();
    uploadDialogVisible.value = false;
  } catch (error: any) {
    ElMessage.error(error.message || "上传失败");
  }
};

const handleUploadError = () => {
  ElMessage.error("上传失败，请重试");
};

const deleteCharacter = async (characterId: number) => {
  try {
    await ElMessageBox.confirm(
      $t("workflow.deleteCharacterConfirm"),
      $t("workflow.deleteConfirmTitle"),
      {
        type: "warning",
        confirmButtonText: $t("workflow.confirmButtonText"),
        cancelButtonText: $t("workflow.cancelButtonText"),
      },
    );

    await characterLibraryAPI.deleteCharacter(characterId);
    ElMessage.success("角色已删除");
    await loadDramaData();
  } catch (error: any) {
    if (error !== "cancel") {
      ElMessage.error(error.message || "删除失败");
    }
  }
};

const goToCompose = () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error("章节信息不存在");
    return;
  }

  router.push({
    name: "SceneComposition",
    params: {
      id: dramaId,
      episodeId: currentEpisode.value.id,
    },
  });
};

// 打开添加场景对话框
const openAddSceneDialog = () => {
  newScene.value = {
    location: "",
    time: "",
    prompt: "",
    image_url: "",
    local_path: "",
  };
  addSceneDialogVisible.value = true;
};

// 保存场景
const saveScene = async () => {
  if (!newScene.value.location) {
    ElMessage.warning($t("workflow.pleaseEnterSceneName"));
    return;
  }

  if (!currentEpisode.value?.id) {
    ElMessage.error($t("workflow.chapterInfoNotExist"));
    return;
  }

  try {
    // 创建场景，关联到当前章节
    await dramaAPI.createScene({
      drama_id: parseInt(dramaId),
      episode_id: parseInt(currentEpisode.value.id),
      location: newScene.value.location,
      time: newScene.value.time || "",
      prompt: newScene.value.prompt,
      image_url: newScene.value.image_url,
      local_path: newScene.value.local_path,
    });

    ElMessage.success($t("workflow.sceneAddSuccess"));
    addSceneDialogVisible.value = false;

    // 重新加载数据以更新场景列表
    await loadDramaData();
  } catch (error: any) {
    ElMessage.error(error.message || $t("workflow.sceneAddFailed"));
  }
};

// 处理场景图片上传成功
const handleSceneImageSuccess = (response: any) => {
  console.log("场景图片上传响应:", response);

  // 处理不同的响应结构
  const imageUrl = response.url || response.data?.url;
  const localPath = response.local_path || response.data?.local_path;

  if (imageUrl) {
    newScene.value.image_url = imageUrl;
  }
  if (localPath) {
    newScene.value.local_path = localPath;
  }

  console.log("更新后的 newScene:", newScene.value);

  if (imageUrl || localPath) {
    ElMessage.success($t("workflow.imageUploadSuccess"));
  } else {
    ElMessage.warning($t("workflow.imageUploadSuccessNoUrl"));
  }
};

// 图片上传前的校验
const beforeAvatarUpload = (file: File) => {
  const isImage = file.type.startsWith("image/");
  const isLt10M = file.size / 1024 / 1024 < 10;

  if (!isImage) {
    ElMessage.error("只能上传图片文件!");
    return false;
  }
  if (!isLt10M) {
    ElMessage.error("图片大小不能超过 10MB!");
    return false;
  }
  return true;
};

// 打开从剧本提取场景对话框
const openExtractSceneDialog = () => {
  extractScenesDialogVisible.value = true;
};

// 从剧本提取场景
const handleExtractScenes = async () => {
  if (!currentEpisode.value?.id) {
    ElMessage.error($t("workflow.chapterInfoNotExist"));
    return;
  }

  try {
    extractingScenes.value = true;
    await dramaAPI.extractBackgrounds(currentEpisode.value.id.toString());

    ElMessage.success($t("workflow.sceneExtractSubmitted"));
    extractScenesDialogVisible.value = false;

    // 自动刷新几次
    let checkCount = 0;
    const maxChecks = 5;
    const checkInterval = window.setInterval(async () => {
      if (isUnmounted) {
        clearInterval(checkInterval);
        return;
      }
      checkCount++;
      await loadDramaData();

      if (checkCount >= maxChecks) {
        clearInterval(checkInterval);
      }
    }, 3000);
  } catch (error: any) {
    ElMessage.error(error.message || $t("workflow.sceneExtractFailed"));
  } finally {
    extractingScenes.value = false;
  }
};

// 监听步骤变化，保存到 localStorage
watch(currentStep, (newStep) => {
  localStorage.setItem(getStepStorageKey(), newStep.toString());
});

onMounted(() => {
  loadDramaData();
  loadSavedModelConfig();
  loadAIConfigs();
});

onUnmounted(() => {
  isUnmounted = true;
  clearAllTimers();
});
</script>

<style scoped lang="scss">
/* ========================================
   Page Layout / 页面布局 - 紧凑边距
   ======================================== */
.page-container {
  min-height: 100vh;
  background: var(--bg-primary);
  // padding: var(--space-2) var(--space-3);
  transition: background var(--transition-normal);
}

@media (min-width: 768px) {
  .page-container {
    // padding: var(--space-3) var(--space-4);
  }
}

@media (min-width: 1024px) {
  .page-container {
    // padding: var(--space-4) var(--space-5);
  }
}

.content-wrapper {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0 auto;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

.content-container {
  height: calc(100% - 134px);
  overflow-y: auto;
}

.actions-container {
  height: 70px;
  background: var(--bg-card);
  overflow: hidden;
}

/* Header styles matching PageHeader component */
.page-header {
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--border-primary);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
}

.page-sub-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) 0;
  flex-wrap: wrap;
}

.sub-header-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-shrink: 0;
}

.sub-header-center {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-width: 0;
}

.sub-header-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.875rem;
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;

  &:hover {
    background: var(--bg-card-hover);
    color: var(--text-primary);
    border-color: var(--border-secondary);
  }
}

.nav-divider {
  width: 1px;
  height: 2rem;
  background: var(--border-primary);
}

.header-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.025em;
  line-height: 1.2;
  white-space: nowrap;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.header-right {
  flex-shrink: 0;
}

.custom-steps {
  display: flex;
  align-items: center;
  gap: 12px;

  .step-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    border-radius: 20px;
    background: var(--bg-card-hover);
    transition: all 0.3s;

    &.active {
      background: var(--accent-light);

      .step-circle {
        background: var(--accent);
        color: var(--text-inverse);
      }
    }

    &.current {
      background: var(--accent);
      color: var(--text-inverse);

      .step-circle {
        background: var(--bg-card);
        color: var(--accent);
      }

      .step-text {
        color: var(--text-inverse);
      }
    }

    .step-circle {
      width: 28px;
      height: 28px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--border-secondary);
      color: var(--text-secondary);
      font-weight: 600;
      transition: all 0.3s;
    }

    .step-text {
      font-size: 14px;
      font-weight: 500;
      white-space: nowrap;
    }
  }

  .step-arrow {
    color: var(--border-secondary);
  }
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;

    .header-info {
      h2 {
        margin: 0 0 4px 0;
        font-size: 20px;
      }

      p {
        margin: 0;
        color: var(--text-muted);
        font-size: 14px;
      }
    }
  }
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin: 12px 0;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
}

.model-tip {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

// 角色库选择对话框
.library-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  max-height: 500px;
  overflow-y: auto;
  padding: 8px;

  .library-item {
    cursor: pointer;
    border: 2px solid transparent;
    border-radius: 8px;
    overflow: hidden;
    transition: all 0.3s;

    &:hover {
      border-color: var(--accent);
      transform: translateY(-2px);
      box-shadow: var(--shadow-lg);
    }

    .el-image {
      width: 100%;
      height: 150px;
    }

    .library-item-name {
      padding: 8px;
      text-align: center;
      font-size: 12px;
      background: var(--bg-secondary);
      color: var(--text-primary);
    }
  }
}

.empty-library {
  padding: 40px 0;
}

// 上传区域
.upload-area {
  :deep(.el-upload-dragger) {
    width: 100%;
    height: 200px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
}

/* ========================================
   Dark Mode / 深色模式
   ======================================== */
:deep(.el-card) {
  background: var(--bg-card);
  border-color: var(--border-primary);
}

:deep(.el-card__header) {
  background: var(--bg-secondary);
  border-color: var(--border-primary);
}

:deep(.el-table) {
  --el-table-bg-color: var(--bg-card);
  --el-table-header-bg-color: var(--bg-secondary);
  --el-table-tr-bg-color: var(--bg-card);
  --el-table-row-hover-bg-color: var(--bg-card-hover);
  --el-table-border-color: var(--border-primary);
  --el-table-text-color: var(--text-primary);
  background: var(--bg-card);
}

:deep(.el-table th.el-table__cell),
:deep(.el-table td.el-table__cell) {
  background: var(--bg-card);
  border-color: var(--border-primary);
}

:deep(
  .el-table--striped .el-table__body tr.el-table__row--striped td.el-table__cell
) {
  background: var(--bg-secondary);
}

:deep(.el-table__header-wrapper th) {
  background: var(--bg-secondary) !important;
  color: var(--text-secondary);
}

:deep(.el-dialog) {
  background: var(--bg-card);
}

:deep(.el-dialog__header) {
  background: var(--bg-card);
}

:deep(.el-form-item__label) {
  color: var(--text-primary);
}

:deep(.el-input__wrapper) {
  background: var(--bg-secondary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

:deep(.el-input__inner) {
  color: var(--text-primary);
}

:deep(.el-textarea__inner) {
  background: var(--bg-secondary);
  color: var(--text-primary);
  box-shadow: 0 0 0 1px var(--border-primary) inset;
}

:deep(.el-select-dropdown) {
  background: var(--bg-elevated);
  border-color: var(--border-primary);
}

:deep(.el-upload-dragger) {
  background: var(--bg-secondary);
  border-color: var(--border-primary);
}
</style>
