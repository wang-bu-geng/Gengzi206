package handlers

import (
	"github.com/wangbugeng/wangbugeng-drama/application/services"
	"github.com/wangbugeng/wangbugeng-drama/pkg/logger"
	"github.com/wangbugeng/wangbugeng-drama/pkg/response"
	"github.com/gin-gonic/gin"
)

// FramePromptHandler 处理帧提示词生成请求
type FramePromptHandler struct {
	framePromptService *services.FramePromptService
	log                *logger.Logger
}

// NewFramePromptHandler 创建帧提示词处理器
func NewFramePromptHandler(framePromptService *services.FramePromptService, log *logger.Logger) *FramePromptHandler {
	return &FramePromptHandler{
		framePromptService: framePromptService,
		log:                log,
	}
}

// GenerateFramePrompt 生成指定类型的帧提示词
// POST /api/v1/storyboards/:id/frame-prompt
func (h *FramePromptHandler) GenerateFramePrompt(c *gin.Context) {
	storyboardID := c.Param("id")

	var req struct {
		FrameType  string `json:"frame_type"`
		PanelCount int    `json:"panel_count"`
		Model      string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	serviceReq := services.GenerateFramePromptRequest{
		StoryboardID: storyboardID,
		FrameType:    services.FrameType(req.FrameType),
		PanelCount:   req.PanelCount,
	}

	// 直接调用服务层的异步方法，该方法会创建任务并返回任务ID
	taskID, err := h.framePromptService.GenerateFramePrompt(serviceReq, req.Model)
	if err != nil {
		h.log.Errorw("Failed to generate frame prompt", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "帧提示词生成任务已创建，正在后台处理...",
	})
}
