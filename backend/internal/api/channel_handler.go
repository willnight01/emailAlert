package api

import (
	"emailAlert/internal/model"
	"emailAlert/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ChannelHandler 通知渠道处理器
type ChannelHandler struct {
	channelService service.ChannelService
}

// NewChannelHandler 创建通知渠道处理器
func NewChannelHandler(channelService service.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

// CreateChannelRequest 创建渠道请求
type CreateChannelRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Config      string `json:"config" binding:"required"`
	Status      string `json:"status"`
	Description string `json:"description"`
	TemplateID  *uint  `json:"template_id"` // 关联的模版ID
}

// UpdateChannelRequest 更新渠道请求
type UpdateChannelRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Config      string `json:"config" binding:"required"`
	Status      string `json:"status"`
	Description string `json:"description"`
	TemplateID  *uint  `json:"template_id"` // 关联的模版ID
}

// TestChannelConfigRequest 测试渠道配置请求
type TestChannelConfigRequest struct {
	Type   string `json:"type" binding:"required"`
	Config string `json:"config" binding:"required"`
}

// UpdateChannelStatusRequest 更新渠道状态请求
type UpdateChannelStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// GetChannels 获取通知渠道列表
func (h *ChannelHandler) GetChannels(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	channelType := c.Query("type")
	status := c.Query("status")

	// 参数验证
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	channels, total, err := h.channelService.GetChannelList(page, size, channelType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取通知渠道列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取通知渠道列表成功",
		"data": gin.H{
			"channels": channels,
			"total":    total,
			"page":     page,
			"size":     size,
		},
	})
}

// CreateChannel 创建通知渠道
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数验证失败",
			"error":   err.Error(),
		})
		return
	}

	// 设置默认状态
	if req.Status == "" {
		req.Status = "active"
	}

	channel := &model.Channel{
		Name:        req.Name,
		Type:        req.Type,
		Config:      req.Config,
		Status:      req.Status,
		Description: req.Description,
		TemplateID:  req.TemplateID,
	}

	if err := h.channelService.CreateChannel(channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建通知渠道失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建通知渠道成功",
		"data":    channel,
	})
}

// GetChannel 获取通知渠道详情
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道ID",
			"error":   err.Error(),
		})
		return
	}

	channel, err := h.channelService.GetChannel(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "通知渠道不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取通知渠道详情成功",
		"data":    channel,
	})
}

// UpdateChannel 更新通知渠道
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道ID",
			"error":   err.Error(),
		})
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数验证失败",
			"error":   err.Error(),
		})
		return
	}

	// 获取现有渠道
	channel, err := h.channelService.GetChannel(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "通知渠道不存在",
			"error":   err.Error(),
		})
		return
	}

	// 更新字段
	channel.Name = req.Name
	channel.Type = req.Type
	channel.Config = req.Config
	channel.Status = req.Status
	channel.Description = req.Description
	channel.TemplateID = req.TemplateID

	if err := h.channelService.UpdateChannel(channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "更新通知渠道失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新通知渠道成功",
		"data":    channel,
	})
}

// DeleteChannel 删除通知渠道
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道ID",
			"error":   err.Error(),
		})
		return
	}

	if err := h.channelService.DeleteChannel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除通知渠道失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除通知渠道成功",
	})
}

// TestChannel 测试通知渠道
func (h *ChannelHandler) TestChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道ID",
			"error":   err.Error(),
		})
		return
	}

	if err := h.channelService.TestChannel(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "测试通知渠道失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "测试通知渠道成功",
	})
}

// TestChannelConfig 测试渠道配置
func (h *ChannelHandler) TestChannelConfig(c *gin.Context) {
	var req TestChannelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数验证失败",
			"error":   err.Error(),
		})
		return
	}

	// 验证必填字段
	if req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "渠道类型不能为空",
			"error":   "type field is required",
		})
		return
	}

	if req.Config == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "渠道配置不能为空",
			"error":   "config field is required",
		})
		return
	}

	if err := h.channelService.TestChannelConfig(req.Type, req.Config); err != nil {
		// 记录详细错误日志
		log.Printf("渠道配置测试失败 - Type: %s, Error: %v", req.Type, err)

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "测试渠道配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "测试渠道配置成功",
	})
}

// UpdateChannelStatus 更新渠道状态
func (h *ChannelHandler) UpdateChannelStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道ID",
			"error":   err.Error(),
		})
		return
	}

	var req UpdateChannelStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数验证失败",
			"error":   err.Error(),
		})
		return
	}

	if err := h.channelService.UpdateChannelStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新渠道状态失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新渠道状态成功",
	})
}

// GetChannelTypes 获取支持的渠道类型
func (h *ChannelHandler) GetChannelTypes(c *gin.Context) {
	types := h.channelService.GetChannelTypes()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取渠道类型成功",
		"data":    types,
	})
}

// SendNotification 发送通知
func (h *ChannelHandler) SendNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的渠道ID",
			"error":   err.Error(),
		})
		return
	}

	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数验证失败",
			"error":   err.Error(),
		})
		return
	}

	if err := h.channelService.SendNotification(uint(id), req.Title, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "发送通知失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "发送通知成功",
	})
}
