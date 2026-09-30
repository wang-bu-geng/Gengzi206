import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage } from "element-plus";
import { aiAPI } from "@/api/ai";

export interface ModelOption {
  modelName: string;
  configName: string;
  configId: number;
  priority: number;
}

/**
 * AI 模型配置：加载可用的文本/图片模型、记忆用户选择（localStorage）
 * 从 EpisodeWorkflow.vue 抽离，逻辑保持不变。
 */
export function useModelConfig(dramaId: string) {
  const { t: $t } = useI18n();

  const textModels = ref<ModelOption[]>([]);
  const imageModels = ref<ModelOption[]>([]);
  const selectedTextModel = ref<string>("");
  const selectedImageModel = ref<string>("");
  const modelConfigDialogVisible = ref(false);

  // 加载AI模型配置
  const loadAIConfigs = async () => {
    try {
      const [textList, imageList] = await Promise.all([
        aiAPI.list("text"),
        aiAPI.list("image"),
      ]);

      // 只使用激活的配置
      const activeTextList = textList.filter((c) => c.is_active);
      const activeImageList = imageList.filter((c) => c.is_active);

      // 展开模型列表并去重（保留优先级最高的）
      const allTextModels = activeTextList
        .flatMap((config) => {
          const models = Array.isArray(config.model)
            ? config.model
            : [config.model];
          return models.map((modelName) => ({
            modelName,
            configName: config.name,
            configId: config.id,
            priority: config.priority || 0,
          }));
        })
        .sort((a, b) => b.priority - a.priority);

      // 按模型名称去重，保留优先级最高的（已排序，第一个就是优先级最高的）
      const textModelMap = new Map<string, ModelOption>();
      allTextModels.forEach((model) => {
        if (!textModelMap.has(model.modelName)) {
          textModelMap.set(model.modelName, model);
        }
      });
      textModels.value = Array.from(textModelMap.values());

      const allImageModels = activeImageList
        .flatMap((config) => {
          const models = Array.isArray(config.model)
            ? config.model
            : [config.model];
          return models.map((modelName) => ({
            modelName,
            configName: config.name,
            configId: config.id,
            priority: config.priority || 0,
          }));
        })
        .sort((a, b) => b.priority - a.priority);

      // 按模型名称去重，保留优先级最高的
      const imageModelMap = new Map<string, ModelOption>();
      allImageModels.forEach((model) => {
        if (!imageModelMap.has(model.modelName)) {
          imageModelMap.set(model.modelName, model);
        }
      });
      imageModels.value = Array.from(imageModelMap.values());

      // 设置默认选择（优先级最高的）
      if (textModels.value.length > 0 && !selectedTextModel.value) {
        selectedTextModel.value = textModels.value[0].modelName;
      }
      if (imageModels.value.length > 0 && !selectedImageModel.value) {
        // 默认选择优先级最高配置下的首个模型（列表已按 priority 降序排序）
        selectedImageModel.value = imageModels.value[0].modelName;
      }

      // 验证已选择的模型是否还在可用列表中，如果不在则重置为默认值
      const availableTextModelNames = textModels.value.map((m) => m.modelName);
      const availableImageModelNames = imageModels.value.map(
        (m) => m.modelName,
      );

      if (
        selectedTextModel.value &&
        !availableTextModelNames.includes(selectedTextModel.value)
      ) {
        console.warn(
          `已选择的文本模型 ${selectedTextModel.value} 不在可用列表中，重置为默认值`,
        );
        selectedTextModel.value =
          textModels.value.length > 0 ? textModels.value[0].modelName : "";
        // 更新 localStorage
        if (selectedTextModel.value) {
          localStorage.setItem(
            `ai_text_model_${dramaId}`,
            selectedTextModel.value,
          );
        }
      }

      if (
        selectedImageModel.value &&
        !availableImageModelNames.includes(selectedImageModel.value)
      ) {
        console.warn(
          `已选择的图片模型 ${selectedImageModel.value} 不在可用列表中，重置为默认值`,
        );
        // 重置为优先级最高配置下的首个模型
        selectedImageModel.value =
          imageModels.value.length > 0 ? imageModels.value[0].modelName : "";
        // 更新 localStorage
        if (selectedImageModel.value) {
          localStorage.setItem(
            `ai_image_model_${dramaId}`,
            selectedImageModel.value,
          );
        }
      }
    } catch (error: any) {
      console.error("加载AI配置失败:", error);
    }
  };

  // 显示模型配置对话框
  const showModelConfigDialog = () => {
    modelConfigDialogVisible.value = true;
    loadAIConfigs();
  };

  // 保存模型配置
  const saveModelConfig = () => {
    if (!selectedTextModel.value || !selectedImageModel.value) {
      ElMessage.warning($t("workflow.pleaseSelectModels"));
      return;
    }

    // 保存模型名称到localStorage
    localStorage.setItem(
      `ai_text_model_${dramaId}`,
      selectedTextModel.value,
    );
    localStorage.setItem(
      `ai_image_model_${dramaId}`,
      selectedImageModel.value,
    );

    ElMessage.success($t("workflow.modelConfigSaved"));
    modelConfigDialogVisible.value = false;
  };

  // 从localStorage加载已保存的模型配置
  const loadSavedModelConfig = () => {
    const savedTextModel = localStorage.getItem(`ai_text_model_${dramaId}`);
    const savedImageModel = localStorage.getItem(`ai_image_model_${dramaId}`);

    if (savedTextModel) {
      selectedTextModel.value = savedTextModel;
    }
    if (savedImageModel) {
      selectedImageModel.value = savedImageModel;
    }
  };

  return {
    textModels,
    imageModels,
    selectedTextModel,
    selectedImageModel,
    modelConfigDialogVisible,
    loadAIConfigs,
    showModelConfigDialog,
    saveModelConfig,
    loadSavedModelConfig,
  };
}
