<template>
  <!-- 阶段 2: 拆分分镜 —— 赛博分镜卡墙 -->
  <el-card shadow="never" class="stage-card">
    <div class="stage-body">
      <!-- 分镜卡墙 -->
      <div
        v-if="
          currentEpisode?.storyboards &&
          currentEpisode.storyboards.length > 0
        "
        class="shots-wall"
      >
        <div class="shots-header">
          <h3 class="wall-title">
            <span class="title-bar" />分镜序列
            <span class="title-en">/ SHOT_SEQUENCE</span>
            <span class="wall-count">{{ currentEpisode.storyboards.length }}</span>
          </h3>
          <div class="shots-header-actions">
            <el-button
              type="primary"
              size="small"
              :loading="batchGeneratingVideo"
              @click="$emit('batch-generate-videos')"
            >
              {{ $t("workflow.batchGenerateVideos") }}
              <span class="btn-en">BATCH_RENDER</span>
            </el-button>
          </div>
        </div>

        <!-- 批量视频生成整体进度（真实轮询百分比） -->
        <div v-if="batchGeneratingVideo" class="line-progress">
          <div class="lp-head">
            <span class="lp-label">批量渲染 BATCH_PROGRESS</span>
            <span class="lp-pct">{{ overallVideoProgress }}%</span>
          </div>
          <div class="lp-track">
            <div class="lp-fill" :style="{ width: overallVideoProgress + '%' }" />
          </div>
          <div class="lp-meta">{{ batchVideoDone }}/{{ batchVideoTotal }} 个分镜已完成</div>
        </div>

        <div class="shot-grid">
          <article
            v-for="(row, index) in currentEpisode.storyboards"
            :key="row.id"
            class="shot-card"
            :class="`shot-card--${shotState(row)}`"
          >
            <header class="shot-head">
              <span class="shot-no">
                SHOT_{{ pad(Number(row.storyboard_number ?? (Number(index) + 1))) }}
              </span>
              <span class="shot-status">
                <span class="status-dot" />
                {{ statusLabel(row) }}
              </span>
            </header>

            <h4 class="shot-title">{{ row.title || "未命名分镜" }}</h4>

            <div class="shot-meta">
              <span v-if="row.shot_type" class="meta-chip">
                景别 <em>{{ row.shot_type }}</em>
              </span>
              <span v-if="row.movement" class="meta-chip">
                运镜 <em>{{ row.movement }}</em>
              </span>
              <span v-if="row.duration" class="meta-chip">
                时长 <em>{{ row.duration }}s</em>
              </span>
            </div>

            <p v-if="row.location" class="shot-field">
              <span class="field-label">地点 LOCATION</span>{{ row.location }}
            </p>
            <p v-if="characterNames(row)" class="shot-field">
              <span class="field-label">角色 CAST</span>{{ characterNames(row) }}
            </p>
            <p v-if="row.action || row.description" class="shot-action">
              {{ row.action || row.description }}
            </p>

            <!-- 单镜真实进度 -->
            <div
              v-if="shotState(row) === 'rendering'"
              class="line-progress line-progress--shot"
            >
              <div class="lp-head">
                <span class="lp-label">RENDERING</span>
                <span class="lp-pct">{{ videoProgress[Number(row.id)] ?? 0 }}%</span>
              </div>
              <div class="lp-track">
                <div
                  class="lp-fill"
                  :style="{ width: (videoProgress[Number(row.id)] ?? 0) + '%' }"
                />
              </div>
            </div>

            <footer class="shot-actions">
              <el-button size="small" @click="$emit('edit-shot', row, Number(index))">
                {{ $t("common.edit") }}
              </el-button>
              <el-button
                type="primary"
                size="small"
                :loading="generatingVideo[row.id]"
                @click="$emit('generate-video-for-shot', row)"
              >
                {{ $t("workflow.generateVideo") }}
              </el-button>
              <el-button
                size="small"
                class="dl-btn"
                :disabled="!row.video_url"
                @click="$emit('download-shot-video', row)"
              >
                {{ $t("workflow.downloadVideo") }}
              </el-button>
            </footer>
          </article>
        </div>
      </div>

      <!-- 未拆分时显示 -->
      <div v-else class="empty-shots">
        <div class="empty-frame">
          <span class="frame-corner frame-corner--tl" />
          <span class="frame-corner frame-corner--tr" />
          <span class="frame-corner frame-corner--bl" />
          <span class="frame-corner frame-corner--br" />
          <el-icon :size="34"><MagicStick /></el-icon>
        </div>
        <h3 class="empty-title">尚未拆分分镜</h3>
        <p class="empty-en">NO_SHOTBOARD_DATA</p>
        <p class="empty-tip">AI 将根据当前剧本自动拆解镜头序列、景别与运镜</p>
        <el-button
          type="primary"
          :loading="generatingShots"
          :icon="MagicStick"
          class="split-btn"
          @click="$emit('generate-shots')"
        >
          {{ generatingShots ? $t("workflow.aiSplitting") : $t("workflow.aiAutoSplit") }}
        </el-button>

        <!-- 任务进度显示（真实轮询） -->
        <div v-if="generatingShots" class="line-progress empty-progress">
          <div class="lp-head">
            <span class="lp-label">结构解析 STRUCTURE_PARSE</span>
            <span class="lp-pct">{{ taskProgress }}%</span>
          </div>
          <div class="lp-track">
            <div class="lp-fill" :style="{ width: taskProgress + '%' }" />
          </div>
          <div class="lp-meta">{{ taskMessage || "状态轮询中…" }}</div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { MagicStick } from "@element-plus/icons-vue";

const props = defineProps<{
  currentEpisode: any;
  batchGeneratingVideo: boolean;
  overallVideoProgress: number;
  batchVideoDone: number;
  batchVideoTotal: number;
  generatingVideo: Record<string, boolean>;
  videoProgress: Record<number, number>;
  generatingShots: boolean;
  taskProgress: number;
  taskMessage: string;
}>();

defineEmits<{
  (e: "batch-generate-videos"): void;
  (e: "edit-shot", shot: any, index: number): void;
  (e: "generate-video-for-shot", shot: any): void;
  (e: "download-shot-video", shot: any): void;
  (e: "generate-shots"): void;
}>();

const pad = (n: number) => String(n).padStart(2, "0");

/** 三态：已成片 / 渲染中 / 待渲染（失败回落到待渲染，可重试） */
const shotState = (row: any): "done" | "rendering" | "pending" => {
  if (props.generatingVideo[row.id]) return "rendering";
  // 轮询进度存在且未到 100 也视为渲染中，覆盖按钮 loading 间隙
  const p = props.videoProgress[row.id as number];
  if (p != null && p > 0 && p < 100) return "rendering";
  if (row.video_url) return "done";
  return "pending";
};

const statusLabel = (row: any) => {
  const map = {
    done: "已成片 RENDERED",
    rendering: "渲染中 RENDERING",
    pending: "待渲染 PENDING",
  } as const;
  return map[shotState(row)];
};

const characterNames = (row: any) => {
  if (!row.characters || row.characters.length === 0) return "";
  return row.characters.map((c: any) => c.name || c).join("、");
};
</script>

<style scoped lang="scss">
@use "./step-common.scss";

/* ── 墙头部 ─────────────────────────────── */
.shots-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}

.wall-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-size: 18px;
  font-weight: 800;
  color: var(--text-primary);
}

.title-bar {
  display: inline-block;
  width: 4px;
  height: 18px;
  background: var(--accent);
  box-shadow: 0 0 10px rgba(255, 45, 45, 0.5);
}

.title-en {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}

.wall-count {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  padding: 2px 8px;
  color: var(--accent);
  border: 1px solid color-mix(in srgb, var(--accent) 45%, transparent);
  background: var(--accent-light);
  border-radius: var(--radius-sm);
}

.btn-en {
  font-family: var(--font-mono);
  font-size: 9px;
  letter-spacing: 0.1em;
  opacity: 0.8;
  margin-left: 2px;
}

.shots-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* ── 细红线进度 ─────────────────────────── */
.line-progress {
  margin: 4px 0 18px;
  padding: 12px 14px;
  border: 1px solid var(--border-primary);
  background: color-mix(in srgb, var(--bg-card) 70%, transparent);
}

.line-progress--shot {
  margin: 10px 0 0;
  padding: 10px 12px;
}

.lp-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 8px;
}

.lp-label {
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  color: var(--text-secondary);
}

.lp-pct {
  font-family: var(--font-mono);
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
}

.lp-track {
  height: 3px;
  background: color-mix(in srgb, var(--text-primary) 12%, transparent);
  overflow: hidden;
}

.lp-fill {
  height: 100%;
  background: var(--accent);
  box-shadow: 0 0 10px rgba(255, 45, 45, 0.75);
  transition: width 0.6s cubic-bezier(0.16, 1, 0.3, 1);
}

.lp-meta {
  margin-top: 8px;
  font-family: var(--font-mono);
  font-size: 10.5px;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}

/* ── 卡片网格 ───────────────────────────── */
.shot-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 14px;
}

.shot-card {
  display: flex;
  flex-direction: column;
  padding: 14px 16px 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-left: 2px solid var(--border-secondary);
  border-radius: var(--radius-md);
  transition:
    border-color var(--transition-fast),
    box-shadow var(--transition-fast);
}

.shot-card:hover {
  border-color: color-mix(in srgb, var(--accent) 55%, var(--border-primary));
  border-left-color: var(--accent);
  box-shadow: var(--shadow-card-hover);
}

.shot-card--done {
  border-left-color: var(--success);
}

.shot-card--rendering {
  border-color: color-mix(in srgb, var(--accent) 60%, var(--border-primary));
  border-left-color: var(--accent);
  box-shadow: 0 0 22px -8px rgba(255, 45, 45, 0.45);
}

.shot-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.shot-no {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
}

.shot-card--done .shot-no,
.shot-card--rendering .shot-no {
  color: var(--text-primary);
}

.shot-status {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-family: var(--font-mono);
  font-size: 9.5px;
  letter-spacing: 0.08em;
  color: var(--text-muted);
  white-space: nowrap;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.shot-card--done .shot-status { color: var(--success); }
.shot-card--rendering .shot-status { color: var(--accent); }
.shot-card--pending .shot-status { color: var(--text-muted); }

.shot-card--rendering .status-dot {
  animation: shot-blink 1s steps(2, end) infinite;
}

@keyframes shot-blink {
  50% { opacity: 0.25; }
}

.shot-title {
  margin: 0 0 10px;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.4;
  color: var(--text-primary);
}

.shot-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}

.meta-chip {
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 2px 7px;
  color: var(--text-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: var(--radius-sm);
  white-space: nowrap;
}

.meta-chip em {
  font-style: normal;
  color: var(--text-primary);
  margin-left: 4px;
}

.shot-field {
  margin: 0 0 6px;
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.6;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field-label {
  font-family: var(--font-mono);
  font-size: 9.5px;
  letter-spacing: 0.08em;
  color: var(--text-muted);
  margin-right: 8px;
}

/* 长文区：常规无衬线 + 正常行高，不用等宽字体 */
.shot-action {
  margin: 2px 0 12px;
  font-size: 13px;
  line-height: 1.75;
  color: var(--text-primary);
  display: -webkit-box;
  -webkit-line-clamp: 4;
  line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.shot-actions {
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px solid var(--border-primary);
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.shot-actions :deep(.el-button) {
  margin: 0;
}

.dl-btn {
  margin-left: auto;
}

/* ── 空态 ───────────────────────────────── */
.empty-shots {
  padding: 70px 20px;
  text-align: center;
}

.empty-frame {
  position: relative;
  width: 92px;
  height: 92px;
  margin: 0 auto 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--accent) 40%, transparent);
  color: var(--accent);
}

.frame-corner {
  position: absolute;
  width: 14px;
  height: 14px;
  border: 0 solid var(--accent);
}

.frame-corner--tl { top: -1px; left: -1px; border-top-width: 2px; border-left-width: 2px; }
.frame-corner--tr { top: -1px; right: -1px; border-top-width: 2px; border-right-width: 2px; }
.frame-corner--bl { bottom: -1px; left: -1px; border-bottom-width: 2px; border-left-width: 2px; }
.frame-corner--br { bottom: -1px; right: -1px; border-bottom-width: 2px; border-right-width: 2px; }

.empty-title {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 800;
  color: var(--text-primary);
}

.empty-en {
  margin: 0 0 12px;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.18em;
  color: var(--text-muted);
}

.empty-tip {
  margin: 0 0 24px;
  font-size: 13px;
  color: var(--text-secondary);
}

.split-btn {
  height: 42px;
  padding: 0 24px;
  border-radius: var(--radius-md);
  font-weight: 700;
}

.empty-progress {
  max-width: 420px;
  margin: 26px auto 0;
  text-align: left;
}

@media (max-width: 768px) {
  .shot-grid {
    grid-template-columns: 1fr;
  }

  .dl-btn {
    margin-left: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .shot-card--rendering .status-dot {
    animation: none;
  }

  .lp-fill {
    transition: none;
  }
}
</style>
