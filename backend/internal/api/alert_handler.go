package api

import (
	"emailAlert/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AlertHandler 告警历史处理器
type AlertHandler struct {
	alertService service.AlertService
}

// NewAlertHandler 创建告警历史处理器
func NewAlertHandler(alertService service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// GetAlerts 获取告警历史列表
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	// 获取筛选参数
	filters := make(map[string]interface{})

	if subject := c.Query("subject"); subject != "" {
		filters["subject"] = subject
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if mailboxID := c.Query("mailbox_id"); mailboxID != "" {
		if id, err := strconv.ParseUint(mailboxID, 10, 32); err == nil {
			filters["mailbox_id"] = uint(id)
		}
	}
	if ruleID := c.Query("rule_id"); ruleID != "" {
		if id, err := strconv.ParseUint(ruleID, 10, 32); err == nil {
			filters["rule_id"] = uint(id)
		}
	}
	if startDate := c.Query("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}
	if sender := c.Query("sender"); sender != "" {
		filters["sender"] = sender
	}

	// 获取排序参数
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	filters["sort_by"] = sortBy
	filters["sort_order"] = sortOrder

	alerts, total, err := h.alertService.GetAlerts(page, size, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取告警历史失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取告警历史成功",
		"data": gin.H{
			"alerts": alerts,
			"total":  total,
			"page":   page,
			"size":   size,
		},
	})
}

// GetAlert 获取告警详情
func (h *AlertHandler) GetAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的告警ID",
			"data":    nil,
		})
		return
	}

	alert, err := h.alertService.GetAlertByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "告警记录不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取告警详情成功",
		"data":    alert,
	})
}

// UpdateAlertStatus 更新告警状态
func (h *AlertHandler) UpdateAlertStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的告警ID",
			"data":    nil,
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 验证状态值
	validStatuses := []string{"pending", "sent", "failed", "canceled"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的状态值，支持的状态: pending, sent, failed, canceled",
			"data":    nil,
		})
		return
	}

	err = h.alertService.UpdateAlertStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新告警状态失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新告警状态成功",
		"data":    nil,
	})
}

// RetryAlert 重试告警发送
func (h *AlertHandler) RetryAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的告警ID",
			"data":    nil,
		})
		return
	}

	err = h.alertService.RetryAlert(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "重试告警发送失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "重试告警发送成功",
		"data":    nil,
	})
}

// GetAlertStats 获取告警统计信息
func (h *AlertHandler) GetAlertStats(c *gin.Context) {
	// 获取时间范围参数
	period := c.DefaultQuery("period", "7d") // 默认7天

	var startDate, endDate string
	now := time.Now()

	switch period {
	case "1d":
		startDate = now.AddDate(0, 0, -1).Format("2006-01-02 00:00:00")
		endDate = now.Format("2006-01-02 23:59:59")
	case "7d":
		startDate = now.AddDate(0, 0, -7).Format("2006-01-02 00:00:00")
		endDate = now.Format("2006-01-02 23:59:59")
	case "30d":
		startDate = now.AddDate(0, 0, -30).Format("2006-01-02 00:00:00")
		endDate = now.Format("2006-01-02 23:59:59")
	default:
		// 自定义时间范围
		startDate = c.Query("start_date")
		endDate = c.Query("end_date")
		if startDate == "" || endDate == "" {
			startDate = now.AddDate(0, 0, -7).Format("2006-01-02 00:00:00")
			endDate = now.Format("2006-01-02 23:59:59")
		}
	}

	stats, err := h.alertService.GetAlertStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取告警统计失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取告警统计成功",
		"data":    stats,
	})
}

// GetAlertTrends 获取告警趋势数据
func (h *AlertHandler) GetAlertTrends(c *gin.Context) {
	// 获取时间范围参数
	period := c.DefaultQuery("period", "7d") // 默认7天

	trends, err := h.alertService.GetAlertTrends(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取告警趋势失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取告警趋势成功",
		"data":    trends,
	})
}

// DeleteAlert 删除告警记录
func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的告警ID",
			"data":    nil,
		})
		return
	}

	err = h.alertService.DeleteAlert(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除告警记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除告警记录成功",
		"data":    nil,
	})
}

// BatchUpdateAlerts 批量更新告警状态
func (h *AlertHandler) BatchUpdateAlerts(c *gin.Context) {
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请选择要更新的告警记录",
			"data":    nil,
		})
		return
	}

	// 验证状态值
	validStatuses := []string{"pending", "sent", "failed", "canceled"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的状态值",
			"data":    nil,
		})
		return
	}

	err := h.alertService.BatchUpdateAlerts(req.IDs, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量更新告警状态失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量更新告警状态成功",
		"data":    nil,
	})
}
