package api

import (
	"emailAlert/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SystemStatusHandler 系统状态处理器
type SystemStatusHandler struct {
	systemStatusService *service.SystemStatusService
}

// NewSystemStatusHandler 创建系统状态处理器
func NewSystemStatusHandler(systemStatusService *service.SystemStatusService) *SystemStatusHandler {
	return &SystemStatusHandler{
		systemStatusService: systemStatusService,
	}
}

// GetSystemHealth 获取系统健康状态
// @Summary 获取系统健康状态
// @Description 获取系统各组件的健康状态，包括数据库、邮件监控、通知服务等
// @Tags 系统状态
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/system/health [get]
func (h *SystemStatusHandler) GetSystemHealth(c *gin.Context) {
	health, err := h.systemStatusService.GetSystemHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "获取系统健康状态失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取系统健康状态成功",
		Data:    health,
	})
}

// GetSystemStats 获取系统统计信息
// @Summary 获取系统统计信息
// @Description 获取系统运行时统计、业务统计、性能统计、连接统计等信息
// @Tags 系统状态
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/system/stats [get]
func (h *SystemStatusHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.systemStatusService.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "获取系统统计信息失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取系统统计信息成功",
		Data:    stats,
	})
}

// GetSystemStatus 获取系统状态（兼容现有接口）
// @Summary 获取系统状态
// @Description 获取系统基本状态信息（兼容性接口）
// @Tags 系统状态
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/system/status [get]
func (h *SystemStatusHandler) GetSystemStatus(c *gin.Context) {
	health, err := h.systemStatusService.GetSystemHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "获取系统状态失败",
			Error:   err.Error(),
		})
		return
	}

	stats, err := h.systemStatusService.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "获取系统状态失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回兼容格式的数据
	compatData := map[string]interface{}{
		"status":     health.Status,
		"uptime":     health.Uptime,
		"version":    health.Version,
		"mailboxes":  stats.Business.TotalMailboxes,
		"rules":      stats.Business.TotalRules,
		"alerts":     stats.Business.TotalAlerts,
		"services":   len(health.Services),
		"health":     health,
		"statistics": stats,
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取系统状态成功",
		Data:    compatData,
	})
}

// CleanupHistoryData 清理历史数据
// @Summary 清理历史数据
// @Description 根据时间范围清理告警历史和通知历史数据
// @Tags 系统状态
// @Accept json
// @Produce json
// @Param body body CleanupRequest true "清理参数"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/system/cleanup [post]
func (h *SystemStatusHandler) CleanupHistoryData(c *gin.Context) {
	var req struct {
		DataType  string `json:"data_type" binding:"required"`  // alerts, notifications, both
		TimeRange string `json:"time_range" binding:"required"` // all, 1month, 3months, 6months, 1year, 2years
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证参数
	validDataTypes := map[string]bool{
		"alerts":        true,
		"notifications": true,
		"both":          true,
	}
	validTimeRanges := map[string]bool{
		"all":     true,
		"1month":  true,
		"3months": true,
		"6months": true,
		"1year":   true,
		"2years":  true,
	}

	if !validDataTypes[req.DataType] {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的数据类型",
			Error:   "支持的类型: alerts, notifications, both",
		})
		return
	}

	if !validTimeRanges[req.TimeRange] {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的时间范围",
			Error:   "支持的范围: all, 1month, 3months, 6months, 1year, 2years",
		})
		return
	}

	// 执行清理
	result, err := h.systemStatusService.CleanupHistoryData(req.DataType, req.TimeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "清理失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "清理完成",
		Data:    result,
	})
}
