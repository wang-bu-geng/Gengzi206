import { ref, type ComputedRef, type Ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage } from "element-plus";
import { imageAPI } from "@/api/image";
import { characterLibraryAPI } from "@/api/character-library";
import { dramaAPI } from "@/api/drama";

interface UseImageGenerationOptions {
  dramaId: string;
  currentEpisode: ComputedRef<any>;
  selectedImageModel: Ref<string>;
  loadDramaData: () => Promise<void>;
  /** 容器组件是否已卸载（共享的单一事实源） */
  isUnmounted: () => boolean;
  /** 容器统一管理的可清理定时器 */
  safeTimeout: (callback: () => void, delay: number) => number;
}

/**
 * 阶段 1：角色/场景图片生成（单个、批量）与页面重进时的任务恢复轮询
 * 从 EpisodeWorkflow.vue 抽离，逻辑保持不变。
 */
export function useImageGeneration({
  dramaId,
  currentEpisode,
  selectedImageModel,
  loadDramaData,
  isUnmounted,
  safeTimeout,
}: UseImageGenerationOptions) {
  const { t: $t } = useI18n();

  const batchGeneratingCharacters = ref(false);
  const batchGeneratingScenes = ref(false);
  const generatingCharacterImages = ref<Record<number, boolean>>({});
  const generatingSceneImages = ref<Record<string, boolean>>({});

  // 选择状态
  const selectedCharacterIds = ref<number[]>([]);
  const selectedSceneIds = ref<number[]>([]);
  const selectAllCharacters = ref(false);
  const selectAllScenes = ref(false);

  // 检查并启动轮询
  const checkAndStartPolling = async () => {
    if (!currentEpisode.value) return;

    // 检查角色的生成状态
    for (const char of currentEpisode.value.characters || []) {
      if (
        char.image_generation_status === "pending" ||
        char.image_generation_status === "processing"
      ) {
        // 查找对应的image_generation记录
        try {
          const imageGenList = await imageAPI.listImages({
            drama_id: dramaId,
            status: char.image_generation_status as any,
          });

          // 找到这个角色的image_generation记录
          const charImageGen = imageGenList.items.find(
            (img) =>
              img.character_id === char.id &&
              (img.status === "pending" || img.status === "processing"),
          );

          if (charImageGen) {
            // 启动轮询
            generatingCharacterImages.value[char.id] = true;
            pollImageStatus(charImageGen.id, async () => {
              await loadDramaData();
              ElMessage.success(`${char.name}的图片生成完成！`);
            }).finally(() => {
              generatingCharacterImages.value[char.id] = false;
            });
          }
        } catch (error) {
          console.error("[轮询] 查询角色图片生成记录失败:", error);
        }
      }
    }

    // 检查场景的生成状态
    for (const scene of currentEpisode.value.scenes || []) {
      if (
        scene.image_generation_status === "pending" ||
        scene.image_generation_status === "processing"
      ) {
        // 查找对应的image_generation记录
        try {
          const imageGenList = await imageAPI.listImages({
            drama_id: dramaId,
            status: scene.image_generation_status as any,
          });

          // 找到这个场景的image_generation记录
          const sceneImageGen = imageGenList.items.find(
            (img) =>
              img.scene_id === scene.id &&
              (img.status === "pending" || img.status === "processing"),
          );

          if (sceneImageGen) {
            // 启动轮询
            generatingSceneImages.value[scene.id] = true;
            pollImageStatus(sceneImageGen.id, async () => {
              await loadDramaData();
              ElMessage.success(`${scene.location}的图片生成完成！`);
            }).finally(() => {
              generatingSceneImages.value[scene.id] = false;
            });
          }
        } catch (error) {
          console.error("[轮询] 查询场景图片生成记录失败:", error);
        }
      }
    }
  };

  // 轮询检查图片生成状态
  const pollImageStatus = async (
    imageGenId: number | string,
    onComplete: () => Promise<void>,
  ) => {
    const maxAttempts = 100;
    const pollInterval = 6000;

    for (let i = 0; i < maxAttempts; i++) {
      if (isUnmounted()) return;
      try {
        await new Promise<void>((resolve) => {
          safeTimeout(() => resolve(), pollInterval);
        });

        if (isUnmounted()) return;

        const imageGen = await imageAPI.getImage(imageGenId);

        if (imageGen.status === "completed") {
          await onComplete();
          return;
        } else if (imageGen.status === "failed") {
          ElMessage.error(`图片生成失败: ${imageGen.error_msg || "未知错误"}`);
          return;
        }
      } catch (error: any) {
        console.error("[轮询] 检查图片状态失败:", error);
      }
    }

    if (!isUnmounted()) {
      ElMessage.warning("图片生成超时，请稍后刷新页面查看结果");
    }
  };

  const generateCharacterImage = async (characterId: number) => {
    generatingCharacterImages.value[characterId] = true;

    try {
      // 获取用户选择的图片生成模型
      const model = selectedImageModel.value || undefined;
      const response = await characterLibraryAPI.generateCharacterImage(
        characterId.toString(),
        model,
      );
      const imageGenId = response.image_generation?.id;

      if (imageGenId) {
        ElMessage.info("角色图片生成中，请稍候...");
        // 轮询检查生成状态
        await pollImageStatus(imageGenId, async () => {
          await loadDramaData();
          ElMessage.success("角色图片生成完成！");
        });
      } else {
        ElMessage.success("角色图片生成已启动");
        await loadDramaData();
      }
    } catch (error: any) {
      ElMessage.error(error.message || "生成失败");
    } finally {
      generatingCharacterImages.value[characterId] = false;
    }
  };

  const toggleSelectAllCharacters = () => {
    if (selectAllCharacters.value) {
      selectedCharacterIds.value =
        currentEpisode.value?.characters?.map((char: any) => char.id) || [];
    } else {
      selectedCharacterIds.value = [];
    }
  };

  const toggleSelectAllScenes = () => {
    if (selectAllScenes.value) {
      selectedSceneIds.value =
        currentEpisode.value?.scenes?.map((scene: any) => scene.id) || [];
    } else {
      selectedSceneIds.value = [];
    }
  };

  const batchGenerateCharacterImages = async () => {
    if (selectedCharacterIds.value.length === 0) {
      ElMessage.warning("请先选择要生成的角色");
      return;
    }

    batchGeneratingCharacters.value = true;
    try {
      // 获取用户选择的图片生成模型
      const model = selectedImageModel.value || undefined;

      // 使用批量生成API
      await characterLibraryAPI.batchGenerateCharacterImages(
        selectedCharacterIds.value.map((id) => id.toString()),
        model,
      );

      ElMessage.success($t("workflow.batchTaskSubmitted"));
      await loadDramaData();
    } catch (error: any) {
      ElMessage.error(error.message || $t("workflow.batchGenerateFailed"));
    } finally {
      batchGeneratingCharacters.value = false;
    }
  };

  const generateSceneImage = async (sceneId: string) => {
    generatingSceneImages.value[sceneId] = true;

    try {
      // 获取用户选择的图片生成模型
      const model = selectedImageModel.value || undefined;
      const response = await dramaAPI.generateSceneImage({
        scene_id: parseInt(sceneId),
        model,
      });
      const imageGenId = response.image_generation?.id;

      if (imageGenId) {
        ElMessage.info($t("workflow.sceneImageGenerating"));
        // 轮询检查生成状态
        await pollImageStatus(imageGenId, async () => {
          await loadDramaData();
          ElMessage.success($t("workflow.sceneImageComplete"));
        });
      } else {
        ElMessage.success($t("workflow.sceneImageStarted"));
        await loadDramaData();
      }
    } catch (error: any) {
      ElMessage.error(error.message || "生成失败");
    } finally {
      generatingSceneImages.value[sceneId] = false;
    }
  };

  const batchGenerateSceneImages = async () => {
    if (selectedSceneIds.value.length === 0) {
      ElMessage.warning("请先选择要生成的场景");
      return;
    }

    batchGeneratingScenes.value = true;
    try {
      const promises = selectedSceneIds.value.map((sceneId) =>
        generateSceneImage(sceneId.toString()),
      );
      const results = await Promise.allSettled(promises);

      const successCount = results.filter((r) => r.status === "fulfilled").length;
      const failCount = results.filter((r) => r.status === "rejected").length;

      if (failCount === 0) {
        ElMessage.success(
          $t("workflow.batchCompleteSuccess", { count: successCount }),
        );
      } else {
        ElMessage.warning(
          $t("workflow.batchCompletePartial", {
            success: successCount,
            fail: failCount,
          }),
        );
      }
    } catch (error: any) {
      ElMessage.error(error.message || $t("workflow.batchGenerateFailed"));
    } finally {
      batchGeneratingScenes.value = false;
    }
  };

  return {
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
  };
}
