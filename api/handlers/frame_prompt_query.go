package handlers

import (
	"github.com/wangbugeng/wangbugeng-drama/domain/models"
	"github.com/wangbugeng/wangbugeng-drama/pkg/logger"
	"github.com/wangbugeng/wangbugeng-drama/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetStoryboardFramePrompts 查询镜头的所有帧提示词
// GET /api/v1/storyboards/:id/frame-prompts
func GetStoryboardFramePrompts(db *gorm.DB, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		storyboardID := c.Param("id")

		var framePrompts []models.FramePrompt
		if err := db.Where("storyboard_id = ?", storyboardID).
			Order("created_at DESC").
			Find(&framePrompts).Error; err != nil {
			log.Errorw("Failed to query frame prompts", "error", err)
			response.InternalError(c, err.Error())
			return
		}

		response.Success(c, gin.H{
			"frame_prompts": framePrompts,
		})
	}
}
