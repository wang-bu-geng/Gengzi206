package handlers

import (
	"strconv"

	"github.com/wangbugeng/wangbugeng-drama/application/services"
	"github.com/wangbugeng/wangbugeng-drama/pkg/config"
	"github.com/wangbugeng/wangbugeng-drama/pkg/logger"
	"github.com/wangbugeng/wangbugeng-drama/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoryboardHandler struct {
	storyboardService *services.StoryboardService
	taskService       *services.TaskService
	log               *logger.Logger
}

func NewStoryboardHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *StoryboardHandler {
	return &StoryboardHandler{
		storyboardService: services.NewStoryboardService(db, cfg, log),
		taskService:       services.NewTaskService(db, log),
		log:               log,
	}
}

// GenerateStoryboard 生成分镜头（异步）
func (h *StoryboardHandler) GenerateStoryboard(c *gin.Context) {
	episodeID := c.Param("episode_id")

	// 接收可选的 model 参数
	var req struct {
		Model string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有提供body或者解析失败，使用空字符串（使用默认模型）
		req.Model = ""
	}

	// 调用生成服务，该服务已经是异步的，会返回任务ID
	taskID, err := h.storyboardService.GenerateStoryboard(episodeID, req.Model)
	if err != nil {
		h.log.Errorw("Failed to generate storyboard", "error", err, "episode_id", episodeID)
		response.InternalError(c, err.Error())
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "分镜头生成任务已创建，正在后台处理...",
	})
}

// UpdateStoryboard 更新分镜
func (h *StoryboardHandler) UpdateStoryboard(c *gin.Context) {
	storyboardID := c.Param("id")

	var req map[string]interface{}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	err := h.storyboardService.UpdateStoryboard(storyboardID, req)
	if err != nil {
		h.log.Errorw("Failed to update storyboard", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Storyboard updated successfully"})
}

// CreateStoryboard 创建分镜
func (h *StoryboardHandler) CreateStoryboard(c *gin.Context) {
	var req services.CreateStoryboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	sb, err := h.storyboardService.CreateStoryboard(&req)
	if err != nil {
		h.log.Errorw("Failed to create storyboard", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, sb)
}

// DeleteStoryboard 删除分镜
func (h *StoryboardHandler) DeleteStoryboard(c *gin.Context) {
	storyboardIDStr := c.Param("id")
	storyboardID, err := strconv.ParseUint(storyboardIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := h.storyboardService.DeleteStoryboard(uint(storyboardID)); err != nil {
		h.log.Errorw("Failed to delete storyboard", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
