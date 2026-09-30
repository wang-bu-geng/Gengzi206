import { computed, ref, type ComputedRef } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage } from "element-plus";
import { videoAPI } from "@/api/video";
import { getVideoUrl } from "@/utils/image";

interface UseVideoGenerationOptions {
  dramaId: string;
  currentEpisode: ComputedRef<any>;
  loadDramaData: () => Promise<void>;
  /** 容器组件是否已卸载（共享的单一事实源） */
  isUnmounted: () => boolean;
  /** 容器统一管理的可清理定时器 */
  safeTimeout: (callback: () => void, delay: number) => number;
}

/**
 * 阶段 2：分镜视频生成（单个 / 一键批量）+ 真实进度轮询
 * 从 EpisodeWorkflow.vue 抽离，逻辑保持不变。
 */
export function useVideoGeneration({
  dramaId,
  currentEpisode,
  loadDramaData,
  isUnmounted,
  safeTimeout,
}: UseVideoGenerationOptions) {
  const { t: $t } = useI18n();

  // 每个分镜的视频生成状态
  const generatingVideo = ref<Record<string, boolean>>({});

  // 每个分镜视频生成的真实进度（百分比，按分镜 id 索引）
  const videoProgress = ref<Record<number, number>>({});

  // 批量视频生成状态
  const batchGeneratingVideo = ref(false);
  const batchVideoTotal = ref(0);
  const batchVideoDone = ref(0);

  // 失败分镜记录（纯观察出口，供生成终端展示真实错误；不影响执行流程）
  const failedVideoShots = ref<Record<number, string>>({});

  // 整体进度 = 当前所有生成中视频进度的平均值
  const overallVideoProgress = computed(() => {
    const values = Object.values(videoProgress.value).filter((v) => v > 0);
    if (values.length === 0) return 0;
    return Math.round(values.reduce((a, b) => a + b, 0) / values.length);
  });

  // 轮询一组视频生成任务，更新对应分镜的真实进度
  const pollVideoGenerationProgress = (
    videoIds: number[],
    mapping: Map<number, number>,
  ): Promise<void> => {
    return new Promise<void>((resolve) => {
      const check = async () => {
        if (isUnmounted()) {
          resolve();
          return;
        }
        try {
          const results = await Promise.all(
            videoIds.map((id) => videoAPI.getVideoGeneration(id)),
          );
          let done = 0;
          for (const v of results) {
            const shotId = mapping.get(v.id);
            if (shotId == null) continue;
            const p =
              v.status === "completed" || v.status === "failed"
                ? 100
                : v.progress ?? 0;
            videoProgress.value = { ...videoProgress.value, [shotId]: p };
            if (v.status === "failed") {
              failedVideoShots.value = {
                ...failedVideoShots.value,
                [shotId]: v.error_msg || "渲染失败",
              };
            }
            if (v.status === "completed" || v.status === "failed") done++;
          }
          if (done >= results.length) {
            resolve();
            return;
          }
          safeTimeout(check, 3000);
        } catch {
          safeTimeout(check, 3000);
        }
      };
      check();
    });
  };

  // 为单个分镜生成视频
  const generateVideoForShot = async (shot: any) => {
    const prompt = shot.video_prompt || shot.action || shot.description || "";
    if (!prompt || prompt.trim().length < 5) {
      ElMessage.warning($t("workflow.videoPromptRequired"));
      return;
    }

    try {
      generatingVideo.value = { ...generatingVideo.value, [shot.id]: true };
      failedVideoShots.value = { ...failedVideoShots.value, [Number(shot.id)]: "" };
      const videoGen = await videoAPI.generateVideo({
        drama_id: dramaId,
        storyboard_id: Number(shot.id),
        prompt: prompt.trim(),
      });
      ElMessage.success($t("workflow.videoTaskSubmitted"));
      // 轮询单分镜真实进度
      await pollVideoGenerationProgress(
        [videoGen.id],
        new Map([[videoGen.id, Number(shot.id)]]),
      );
      await loadDramaData();
    } catch (error: any) {
      ElMessage.error(error.message || $t("workflow.videoGenerateFailed"));
    } finally {
      generatingVideo.value = { ...generatingVideo.value, [shot.id]: false };
    }
  };

  // 一键生成全部未生成视频的分镜
  const batchGenerateVideos = async () => {
    const shots = currentEpisode.value?.storyboards || [];
    const pendingShots = shots.filter((s) => !s.video_url);
    if (pendingShots.length === 0) {
      ElMessage.info($t("workflow.noShotsNeedVideo"));
      return;
    }

    batchGeneratingVideo.value = true;
    videoProgress.value = {};
    failedVideoShots.value = {};
    batchVideoTotal.value = 0;
    batchVideoDone.value = 0;

    try {
      const created = await videoAPI.batchGenerateForEpisode(
        Number(currentEpisode.value!.id),
      );
      if (!created || created.length === 0) {
        ElMessage.warning($t("workflow.noShotsNeedVideo"));
        return;
      }

      batchVideoTotal.value = created.length;
      const mapping = new Map<number, number>();
      for (const v of created) {
        if (v.storyboard_id != null)
          mapping.set(v.id, Number(v.storyboard_id));
      }

      ElMessage.success($t("workflow.videoTaskSubmitted"));
      await pollVideoGenerationProgress(
        created.map((v) => v.id),
        mapping,
      );
      batchVideoDone.value = created.length;
      await loadDramaData();
    } catch (error: any) {
      ElMessage.error(error.message || $t("workflow.videoGenerateFailed"));
    } finally {
      batchGeneratingVideo.value = false;
    }
  };

  // 下载单个分镜视频
  const downloadShotVideo = (shot: any) => {
    const url = getVideoUrl(shot) || shot.video_url;
    if (!url) {
      ElMessage.warning($t("workflow.noVideoYet"));
      return;
    }
    const a = document.createElement("a");
    a.href = url;
    a.download = `shot_${shot.id || Date.now()}.mp4`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  };

  return {
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
  };
}
