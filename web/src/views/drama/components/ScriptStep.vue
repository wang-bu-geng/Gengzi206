<template>
  <!-- 阶段 0: 章节内容 + 提取角色场景 -->
  <el-card
    shadow="never"
    class="stage-card stage-card-fullscreen"
  >
    <div class="stage-body stage-body-fullscreen">
      <!-- 未保存时显示输入框 -->
      <div v-if="!hasScript" class="generation-form">
        <el-input
          v-model="scriptContent"
          type="textarea"
          :placeholder="$t('workflow.scriptPlaceholder')"
          class="script-textarea script-textarea-fullscreen"
        />

        <div class="action-buttons-inline">
          <el-button
            type="primary"
            size="default"
            @click="$emit('save')"
            :disabled="!scriptContent.trim() || generatingScript"
          >
            <el-icon><Check /></el-icon>
            <span>{{ $t("workflow.saveChapter") }}</span>
          </el-button>
        </div>
      </div>

      <!-- 已保存时显示内容 -->
      <div v-if="hasScript" class="overview-section">
        <div class="episode-info">
          <h3>
            {{ $t("workflow.chapterContent", { number: episodeNumber }) }}
          </h3>
          <el-tag type="success" size="large">{{
            $t("workflow.saved")
          }}</el-tag>
        </div>
        <div class="overview-content">
          <el-input
            v-model="currentEpisode.script_content"
            type="textarea"
            :rows="15"
            readonly
            class="script-display"
          />
        </div>

        <el-divider />

        <!-- 显示已提取的角色和场景 -->
        <div v-if="hasExtractedData" class="extracted-info">
          <el-alert
            type="success"
            :closable="false"
            style="margin-bottom: 16px"
          >
            <template #title>
              <div style="display: flex; align-items: center; gap: 16px">
                <span>✅ {{ $t("workflow.extractedData") }}</span>
                <el-tag v-if="hasCharacters" type="success"
                  >{{ $t("workflow.characters") }}:
                  {{ charactersCount }}</el-tag
                >
                <el-tag v-if="currentEpisode?.scenes" type="success"
                  >{{ $t("workflow.scenes") }}:
                  {{ currentEpisode.scenes.length }}</el-tag
                >
              </div>
            </template>
          </el-alert>

          <!-- 角色列表 -->
          <div v-if="hasCharacters" style="margin-bottom: 16px">
            <h4 class="extracted-title">
              {{ $t("workflow.extractedCharacters") }}：
            </h4>
            <div style="display: flex; flex-wrap: wrap; gap: 8px">
              <el-tag
                v-for="char in currentEpisode?.characters"
                :key="char.id"
                type="info"
              >
                {{ char.name }}
                <span v-if="char.role" class="secondary-text"
                  >({{ char.role }})</span
                >
              </el-tag>
            </div>
          </div>

          <!-- 场景列表 -->
          <div
            v-if="
              currentEpisode?.scenes && currentEpisode.scenes.length > 0
            "
          >
            <h4 class="extracted-title">
              {{ $t("workflow.extractedScenes") }}：
            </h4>
            <div style="display: flex; flex-wrap: wrap; gap: 8px">
              <el-tag
                v-for="scene in currentEpisode.scenes"
                :key="scene.id"
                type="warning"
              >
                {{ scene.location }}
                <span class="secondary-text">· {{ scene.time }}</span>
              </el-tag>
            </div>
          </div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { Check } from "@element-plus/icons-vue";

defineProps<{
  episodeNumber: number;
  hasScript: boolean;
  generatingScript: boolean;
  currentEpisode: any;
  hasExtractedData: boolean;
  hasCharacters: boolean;
  charactersCount: number;
}>();

defineEmits<{
  (e: "save"): void;
}>();

// 剧本输入内容（与容器 scriptContent 双向同步）
const scriptContent = defineModel<string>("scriptContent", { required: true });
</script>

<style scoped lang="scss">
@use "./step-common.scss";

.stage-card {
  &.stage-card-fullscreen {
    .stage-body-fullscreen {
      min-height: calc(100vh - 200px);
    }
  }
}

.action-buttons-inline {
  display: flex;
  gap: 12px;
}

.script-textarea {
  margin: 16px 0;

  &.script-textarea-fullscreen {
    :deep(textarea) {
      min-height: 500px;
      font-size: 14px;
      line-height: 1.8;
    }
  }
}

.extracted-title {
  margin-bottom: 8px;
  color: var(--text-secondary);
}

.secondary-text {
  color: var(--text-muted);
  margin-left: 4px;
}
</style>
