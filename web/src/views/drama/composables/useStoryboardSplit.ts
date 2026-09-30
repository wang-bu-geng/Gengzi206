import { onUnmounted, ref, type ComputedRef, type Ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import { generationAPI } from "@/api/generation";

interface UseStoryboardSplitOptions {
  drama: Ref<any>;
  currentEpisode: ComputedRef<any>;
  episodeNumber: number;
  selectedTextModel: Ref<string>;
  loadDramaData: () => Promise<void>;
  /** 容器组件是否已卸载（共享的单一事实源） */
  isUnmounted: () => boolean;
}

/**
 * 阶段 2：AI 拆分分镜（异步任务 + 2s 轮询进度）
 * 从 EpisodeWorkflow.vue 抽离，逻辑保持不变。
 */
export function useStoryboardSplit({
  drama,
  currentEpisode,
  episodeNumber,
  selectedTextModel,
  loadDramaData,
  isUnmounted,
}: UseStoryboardSplitOptions) {
  const { t: $t } = useI18n();

  const generatingShots = ref(false);
  const taskProgress = ref(0);
  const taskMessage = ref("");
  let pollTimer: any = null;

  const generateShots = async () => {
    if (!currentEpisode.value?.id) {
      ElMessage.error("章节信息不存在");
      return;
    }

    generatingShots.value = true;
    taskProgress.value = 0;
    taskMessage.value = "初始化任务...";

    try {
      const episodeId = currentEpisode.value.id.toString();

      // 【调试日志】输出当前操作的集数信息
      console.log("=== 开始生成分镜 ===");
      console.log("当前 episodeNumber (路由参数):", episodeNumber);
      console.log("当前 episodeId (从 currentEpisode 获取):", episodeId);
      console.log("currentEpisode 完整信息:", {
        id: currentEpisode.value?.id,
        episode_number: currentEpisode.value?.episode_number,
        title: currentEpisode.value?.title,
      });
      console.log(
        "所有剧集列表:",
        drama.value?.episodes?.map((ep: any) => ({
          id: ep.id,
          episode_number: ep.episode_number,
          title: ep.title,
        })),
      );

      // 创建异步任务
      const response = await generationAPI.generateStoryboard(
        episodeId,
        selectedTextModel.value,
      );

      taskMessage.value = response.message || "任务已创建";

      // 开始轮询任务状态
      await pollTaskStatus(response.task_id);
    } catch (error: any) {
      ElMessage.error(error.message || "拆分失败");
      generatingShots.value = false;
    }
  };

  const pollTaskStatus = async (taskId: string) => {
    const checkStatus = async () => {
      if (isUnmounted()) return;
      try {
        const task = await generationAPI.getTaskStatus(taskId);

        if (isUnmounted()) return;

        taskProgress.value = task.progress;
        taskMessage.value = task.message || `处理中... ${task.progress}%`;

        if (task.status === "completed") {
          if (pollTimer) {
            clearInterval(pollTimer);
            pollTimer = null;
          }
          generatingShots.value = false;

          ElMessage.success($t("workflow.splitSuccess"));

          // 拆分完成后刷新分镜列表，无需跳转剪辑页
          await loadDramaData();
        } else if (task.status === "failed") {
          if (pollTimer) {
            clearInterval(pollTimer);
            pollTimer = null;
          }
          generatingShots.value = false;
          ElMessage.error(task.error || task.message || "分镜拆分失败");
        }
      } catch (error: any) {
        if (isUnmounted()) return;
        if (pollTimer) {
          clearInterval(pollTimer);
          pollTimer = null;
        }
        generatingShots.value = false;
        ElMessage.error("查询任务状态失败: " + error.message);
      }
    };

    await checkStatus();

    pollTimer = window.setInterval(checkStatus, 2000);
  };

  const regenerateShots = async () => {
    await ElMessageBox.confirm($t("workflow.reSplitConfirm"), $t("common.tip"), {
      type: "warning",
    });

    await generateShots();
  };

  onUnmounted(() => {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  });

  return {
    generatingShots,
    taskProgress,
    taskMessage,
    generateShots,
    regenerateShots,
  };
}
