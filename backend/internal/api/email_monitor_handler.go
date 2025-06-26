package api

import (
	"emailAlert/internal/service"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// EmailMonitorHandler 邮件监控API处理器
type EmailMonitorHandler struct {
	emailMonitorService *service.EmailMonitorService
}

// NewEmailMonitorHandler 创建邮件监控API处理器实例
func NewEmailMonitorHandler(emailMonitorService *service.EmailMonitorService) *EmailMonitorHandler {
	return &EmailMonitorHandler{
		emailMonitorService: emailMonitorService,
	}
}

// StartMonitor 启动邮件监控
// @Summary 启动邮件监控
// @Description 启动邮件监控服务，开始监控所有活跃的邮箱
// @Tags 邮件监控
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/monitor/start [post]
func (h *EmailMonitorHandler) StartMonitor(c *gin.Context) {
	err := h.emailMonitorService.Start()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "启动邮件监控失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "邮件监控已启动",
	})
}

// StopMonitor 停止邮件监控
// @Summary 停止邮件监控
// @Description 停止邮件监控服务
// @Tags 邮件监控
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/monitor/stop [post]
func (h *EmailMonitorHandler) StopMonitor(c *gin.Context) {
	// 检查监控是否正在运行
	if !h.emailMonitorService.IsRunning() {
		// 如果监控未运行，也返回成功，因为目标状态已达成
		c.JSON(http.StatusOK, APIResponse{
			Code:    200,
			Message: "邮件监控已停止",
		})
		return
	}

	// 停止监控服务
	err := h.emailMonitorService.Stop()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "停止邮件监控失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "邮件监控已停止",
	})
}

// GetMonitorStatus 获取监控状态
// @Summary 获取监控状态
// @Description 获取邮件监控服务的当前状态信息
// @Tags 邮件监控
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Router /api/v1/monitor/status [get]
func (h *EmailMonitorHandler) GetMonitorStatus(c *gin.Context) {
	status := h.emailMonitorService.GetStatus()
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取监控状态成功",
		Data:    status,
	})
}

// RefreshMailboxes 刷新邮箱配置
// @Summary 刷新邮箱配置
// @Description 重新加载邮箱配置，用于邮箱配置变更后的热更新
// @Tags 邮件监控
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/monitor/refresh [post]
func (h *EmailMonitorHandler) RefreshMailboxes(c *gin.Context) {
	err := h.emailMonitorService.RefreshMailboxes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "刷新邮箱配置失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "邮箱配置已刷新",
	})
}

// GetEmailStats 获取邮件统计信息
// @Summary 获取邮件统计信息
// @Description 获取邮件处理的统计信息
// @Tags 邮件监控
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/monitor/stats [get]
func (h *EmailMonitorHandler) GetEmailStats(c *gin.Context) {
	stats, err := h.emailMonitorService.GetEmailStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "获取统计信息失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取统计信息成功",
		Data:    stats,
	})
}

// GetMonitorLogs 获取实时监控日志 (Server-Sent Events)
func (h *EmailMonitorHandler) GetMonitorLogs(c *gin.Context) {
	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 注册日志客户端
	logClient := h.emailMonitorService.AddLogClient()
	defer h.emailMonitorService.RemoveLogClient(logClient)

	// 发送心跳以保持连接
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case logEntry, ok := <-logClient:
			if !ok {
				return
			}

			// 发送日志数据
			data, _ := json.Marshal(logEntry)
			c.SSEvent("log", string(data))
			c.Writer.Flush()

		case <-ticker.C:
			// 发送心跳
			c.SSEvent("heartbeat", "ping")
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			// 客户端断开连接
			return
		}
	}
}

// RegisterRoutes 注册路由
func (h *EmailMonitorHandler) RegisterRoutes(router *gin.RouterGroup) {
	monitor := router.Group("/monitor")
	{
		monitor.POST("/start", h.StartMonitor)
		monitor.POST("/stop", h.StopMonitor)
		monitor.GET("/status", h.GetMonitorStatus)
		monitor.POST("/refresh", h.RefreshMailboxes)
		monitor.GET("/stats", h.GetEmailStats)
		monitor.GET("/logs", h.GetMonitorLogs)
	}
}
