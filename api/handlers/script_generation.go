package handlers

import (
	"github.com/wangbugeng/wangbugeng-drama/application/services"
	"github.com/wangbugeng/wangbugeng-drama/pkg/config"
	"github.com/wangbugeng/wangbugeng-drama/pkg/logger"
	"github.com/wangbugeng/wangbugeng-drama/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ScriptGenerationHandler struct {
	scriptService *services.ScriptGenerationService
	taskService   *services.TaskService
	log           *logger.Logger
}

func NewScriptGenerationHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *ScriptGenerationHandler {
	return &ScriptGenerationHandler{
		scriptService: services.NewScriptGenerationService(db, cfg, log),
		taskService:   services.NewTaskService(db, log),
		log:           log,
	}
}

func (h *ScriptGenerationHandler) GenerateCharacters(c *gin.Context) {
	var req services.GenerateCharactersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 直接调用服务层的异步方法，该方法会创建任务并返回任务ID
	taskID, err := h.scriptService.GenerateCharacters(&req)
	if err != nil {
		h.log.Errorw("Failed to generate characters", "error", err, "drama_id", req.DramaID)
		response.InternalError(c, err.Error())
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "角色生成任务已创建，正在后台处理...",
	})
}
