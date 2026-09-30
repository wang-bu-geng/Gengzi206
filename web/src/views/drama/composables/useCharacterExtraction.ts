import { ref, type ComputedRef, type Ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import { generationAPI } from "@/api/generation";
import { dramaAPI } from "@/api/drama";

interface UseCharacterExtractionOptions {
  dramaId: string;
  currentEpisode: ComputedRef<any>;
  selectedTextModel: Ref<string>;
  hasExtractedData: ComputedRef<boolean>;
  loadDramaData: () => Promise<void>;
  /** 容器组件是否已卸载（共享的单一事实源） */
  isUnmounted: () => boolean;
  /** 容器统一管理的可清理定时器 */
  safeTimeout: (callback: () => void, delay: number) => number;
}

/**
 * 阶段 0：从剧本提取角色与场景（并行异步任务 + 轮询）
 * 从 EpisodeWorkflow.vue 抽离，逻辑保持不变。
 */
export function useCharacterExtraction({
  dramaId,
  currentEpisode,
  selectedTextModel,
  hasExtractedData,
  loadDramaData,
  isUnmounted,
  safeTimeout,
}: UseCharacterExtractionOptions) {
  const { t: $t } = useI18n();

  const extractingCharactersAndBackgrounds = ref(false);

  const handleExtractCharactersAndBackgrounds = async () => {
    // 如果已经提取过，显示确认对话框
    if (hasExtractedData.value) {
      try {
        await ElMessageBox.confirm(
          $t("workflow.reExtractConfirmMessage"),
          $t("workflow.reExtractConfirmTitle"),
          {
            confirmButtonText: $t("common.confirm"),
            cancelButtonText: $t("common.cancel"),
            type: "warning",
            distinguishCancelAndClose: true,
          },
        );
      } catch {
        ElMessage.info($t("workflow.extractCancelled"));
        return;
      }
    }

    // 显示即将开始的提示
    if (hasExtractedData.value) {
      ElMessage.info($t("workflow.startReExtracting"));
    }

    await extractCharactersAndBackgrounds();
  };

  const extractCharactersAndBackgrounds = async () => {
    if (!currentEpisode.value?.id) {
      ElMessage.error("章节信息不存在");
      return;
    }

    extractingCharactersAndBackgrounds.value = true;

    try {
      const episodeId = currentEpisode.value.id;

      // 并行创建异步任务
      const [characterTask, backgroundTask] = await Promise.all([
        generationAPI.generateCharacters({
          drama_id: dramaId.toString(),
          episode_id: episodeId,
          outline: currentEpisode.value.script_content || "",
          count: 0,
          model: selectedTextModel.value, // 传递用户选择的文本模型
        }),
        dramaAPI.extractBackgrounds(
          episodeId.toString(),
          selectedTextModel.value,
        ), // 传递用户选择的文本模型
      ]);

      ElMessage.success("任务已创建，正在后台处理...");

      // 并行轮询两个任务
      await Promise.all([
        pollExtractTask(characterTask.task_id, "character"),
        pollExtractTask(backgroundTask.task_id, "background"),
      ]);

      ElMessage.success($t("workflow.charactersAndScenesExtractSuccess"));
      await loadDramaData();
    } catch (error: any) {
      console.error($t("workflow.charactersAndScenesExtractFailed") + ":", error);

      const errorData = error.response?.data?.error;
      const errorMsg = errorData?.message || error.message || "提取失败";

      if (
        errorMsg.includes("no config found") ||
        errorMsg.includes("AI client") ||
        errorMsg.includes("failed to get AI client")
      ) {
        ElMessage({
          type: "warning",
          message: '未配置AI服务，请前往"设置 > AI服务配置"添加文本生成服务',
          duration: 5000,
          showClose: true,
        });
      } else {
        ElMessage.error(errorMsg);
      }
    } finally {
      extractingCharactersAndBackgrounds.value = false;
    }
  };

  // 轮询提取任务状态
  const pollExtractTask = async (
    taskId: string,
    type: "character" | "background",
  ) => {
    const maxAttempts = 60;
    const interval = 2000;

    for (let i = 0; i < maxAttempts; i++) {
      if (isUnmounted()) return;

      await new Promise<void>((resolve) => {
        safeTimeout(() => resolve(), interval);
      });

      if (isUnmounted()) return;

      try {
        const task = await generationAPI.getTaskStatus(taskId);

        if (task.status === "completed") {
          if (type === "character" && task.result) {
            const result =
              typeof task.result === "string"
                ? JSON.parse(task.result)
                : task.result;
            if (result.characters && result.characters.length > 0) {
              await dramaAPI.saveCharacters(
                dramaId,
                result.characters,
                currentEpisode.value?.id,
              );
            }
          }
          return;
        } else if (task.status === "failed") {
          throw new Error(
            task.error ||
              task.message ||
              (type === "character"
                ? $t("workflow.characterGenerationFailed")
                : $t("workflow.sceneExtractionFailed")),
          );
        }
      } catch (error: any) {
        console.error(`轮询${type}任务状态失败:`, error);
        throw error;
      }
    }

    throw new Error(
      type === "character"
        ? $t("workflow.characterGenerationTimeout")
        : $t("workflow.sceneExtractionTimeout"),
    );
  };

  return {
    extractingCharactersAndBackgrounds,
    handleExtractCharactersAndBackgrounds,
  };
}
