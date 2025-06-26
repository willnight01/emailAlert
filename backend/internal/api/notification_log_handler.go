package api

import (
	"emailAlert/internal/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type NotificationLogHandler struct {
	notificationLogRepo repository.NotificationLogRepository
	alertRepo           *repository.AlertRepository
	channelRepo         repository.ChannelRepository
}

func NewNotificationLogHandler(
	notificationLogRepo repository.NotificationLogRepository,
	alertRepo *repository.AlertRepository,
	channelRepo repository.ChannelRepository,
) *NotificationLogHandler {
	return &NotificationLogHandler{
		notificationLogRepo: notificationLogRepo,
		alertRepo:           alertRepo,
		channelRepo:         channelRepo,
	}
}

// GetNotificationLogs 获取通知日志列表
func (h *NotificationLogHandler) GetNotificationLogs(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	channelID := c.Query("channel_id")
	status := c.Query("status")
	content := c.Query("content")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 计算偏移量
	offset := (page - 1) * size

	// 构建查询条件
	conditions := make(map[string]interface{})
	if channelID != "" {
		if id, err := strconv.ParseUint(channelID, 10, 32); err == nil {
			conditions["channel_id"] = uint(id)
		}
	}
	if status != "" {
		conditions["status"] = status
	}
	if content != "" {
		conditions["content_like"] = content
	}

	// 处理日期范围
	var startTime, endTime *time.Time
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			startTime = &t
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// 结束日期设置为当天的23:59:59
			endTimeDay := t.Add(24*time.Hour - time.Second)
			endTime = &endTimeDay
		}
	}

	// 获取通知日志列表
	logs, total, err := h.notificationLogRepo.GetNotificationLogsWithDetails(
		conditions, offset, size, startTime, endTime,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取通知日志失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取通知日志成功",
		"data": gin.H{
			"logs":  logs,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// GetNotificationLog 获取单个通知日志详情
func (h *NotificationLogHandler) GetNotificationLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的日志ID",
		})
		return
	}

	log, err := h.notificationLogRepo.GetByIDWithDetails(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "通知日志不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取通知日志详情成功",
		"data":    log,
	})
}

// DeleteNotificationLog 删除通知日志
func (h *NotificationLogHandler) DeleteNotificationLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的日志ID",
		})
		return
	}

	err = h.notificationLogRepo.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除通知日志失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除通知日志成功",
	})
}

// RetryNotificationLog 重试发送通知
func (h *NotificationLogHandler) RetryNotificationLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的日志ID",
		})
		return
	}

	// 获取通知日志
	log, err := h.notificationLogRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "通知日志不存在",
		})
		return
	}

	// 检查是否可以重试
	if log.Status == "success" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "已发送成功的通知无需重试",
		})
		return
	}

	if log.RetryCount >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "重试次数已达上限",
		})
		return
	}

	// 更新重试次数和状态
	log.RetryCount++
	log.Status = "pending"
	log.ErrorMsg = ""

	err = h.notificationLogRepo.Update(log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新通知日志失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "通知已加入重试队列",
	})
}

// GetNotificationLogStats 获取通知日志统计
func (h *NotificationLogHandler) GetNotificationLogStats(c *gin.Context) {
	// 获取时间范围参数
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	stats, err := h.notificationLogRepo.GetStatistics(startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取统计数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取统计数据成功",
		"data":    stats,
	})
}
