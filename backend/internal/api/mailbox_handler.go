package api

import (
	"emailAlert/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MailboxHandler 邮箱管理处理器
type MailboxHandler struct {
	mailboxService *service.MailboxService
}

// NewMailboxHandler 创建邮箱管理处理器
func NewMailboxHandler(mailboxService *service.MailboxService) *MailboxHandler {
	return &MailboxHandler{mailboxService: mailboxService}
}

// GetMailboxes 获取邮箱列表
// @Summary 获取邮箱列表
// @Description 分页获取邮箱配置列表，支持状态过滤
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param status query string false "状态过滤" Enums(active, inactive)
// @Success 200 {object} APIResponse{data=service.MailboxListResponse}
// @Router /api/v1/mailboxes [get]
func (h *MailboxHandler) GetMailboxes(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.Query("status")

	// 调用服务层
	result, err := h.mailboxService.List(page, size, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "获取邮箱列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取邮箱列表成功",
		Data:    result,
	})
}

// CreateMailbox 创建邮箱配置
// @Summary 创建邮箱配置
// @Description 创建新的邮箱配置
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param mailbox body service.CreateMailboxRequest true "邮箱配置信息"
// @Success 201 {object} APIResponse{data=model.Mailbox}
// @Router /api/v1/mailboxes [post]
func (h *MailboxHandler) CreateMailbox(c *gin.Context) {
	var req service.CreateMailboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	mailbox, err := h.mailboxService.Create(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "邮箱地址已存在" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "创建邮箱配置失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Code:    201,
		Message: "创建邮箱配置成功",
		Data:    mailbox,
	})
}

// GetMailbox 获取邮箱详情
// @Summary 获取邮箱详情
// @Description 根据ID获取邮箱配置详情
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param id path int true "邮箱ID"
// @Success 200 {object} APIResponse{data=model.Mailbox}
// @Router /api/v1/mailboxes/{id} [get]
func (h *MailboxHandler) GetMailbox(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的邮箱ID",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	mailbox, err := h.mailboxService.GetByID(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "邮箱配置不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "获取邮箱详情失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取邮箱详情成功",
		Data:    mailbox,
	})
}

// GetMailboxWithPassword 获取邮箱详情（包含密码，用于编辑）
// @Summary 获取邮箱详情（包含密码）
// @Description 根据ID获取邮箱配置详情，包含密码字段，用于编辑
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param id path int true "邮箱ID"
// @Success 200 {object} APIResponse{data=model.MailboxWithPassword}
// @Router /api/v1/mailboxes/{id}/edit [get]
func (h *MailboxHandler) GetMailboxWithPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的邮箱ID",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层获取包含密码的邮箱信息
	mailbox, err := h.mailboxService.GetByIDWithPassword(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "邮箱配置不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "获取邮箱详情失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取邮箱详情成功",
		Data:    mailbox,
	})
}

// UpdateMailbox 更新邮箱配置
// @Summary 更新邮箱配置
// @Description 更新邮箱配置信息
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param id path int true "邮箱ID"
// @Param mailbox body service.UpdateMailboxRequest true "邮箱配置信息"
// @Success 200 {object} APIResponse{data=model.Mailbox}
// @Router /api/v1/mailboxes/{id} [put]
func (h *MailboxHandler) UpdateMailbox(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的邮箱ID",
			Error:   err.Error(),
		})
		return
	}

	var req service.UpdateMailboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	mailbox, err := h.mailboxService.Update(uint(id), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "邮箱配置不存在" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "邮箱地址已被其他配置使用" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "更新邮箱配置失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "更新邮箱配置成功",
		Data:    mailbox,
	})
}

// DeleteMailbox 删除邮箱配置
// @Summary 删除邮箱配置
// @Description 删除指定的邮箱配置
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param id path int true "邮箱ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/mailboxes/{id} [delete]
func (h *MailboxHandler) DeleteMailbox(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的邮箱ID",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	err = h.mailboxService.Delete(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "邮箱配置不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "删除邮箱配置失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "删除邮箱配置成功",
	})
}

// TestMailbox 测试邮箱连接
// @Summary 测试邮箱连接
// @Description 测试指定邮箱的连接状态
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param id path int true "邮箱ID"
// @Success 200 {object} APIResponse{data=email.ConnectionInfo}
// @Router /api/v1/mailboxes/{id}/test [post]
func (h *MailboxHandler) TestMailbox(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的邮箱ID",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层测试连接
	connectionInfo, err := h.mailboxService.TestConnection(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "邮箱配置不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "测试邮箱连接失败",
			Error:   err.Error(),
		})
		return
	}

	// 根据连接状态确定响应状态码
	statusCode := http.StatusOK
	message := "邮箱连接测试完成"

	if connectionInfo.Status == "error" {
		message = "邮箱连接测试失败"
	}

	c.JSON(statusCode, APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    connectionInfo,
	})
}

// TestMailboxConfig 测试邮箱配置
// @Summary 测试邮箱配置
// @Description 使用提供的配置参数测试邮箱连接
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param config body service.CreateMailboxRequest true "邮箱配置参数"
// @Success 200 {object} APIResponse{data=email.ConnectionInfo}
// @Router /api/v1/mailboxes/test-config [post]
func (h *MailboxHandler) TestMailboxConfig(c *gin.Context) {
	var req service.CreateMailboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层测试连接
	connectionInfo, err := h.mailboxService.TestConnectionWithConfig(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "测试邮箱连接失败",
			Error:   err.Error(),
		})
		return
	}

	// 根据连接状态确定响应消息
	message := "邮箱连接测试完成"
	if connectionInfo.Status == "error" {
		message = "邮箱连接测试失败"
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: message,
		Data:    connectionInfo,
	})
}

// UpdateMailboxStatus 更新邮箱状态
// @Summary 更新邮箱状态
// @Description 更新邮箱的激活状态
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param id path int true "邮箱ID"
// @Param status body map[string]string true "状态信息"
// @Success 200 {object} APIResponse
// @Router /api/v1/mailboxes/{id}/status [put]
func (h *MailboxHandler) UpdateMailboxStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的邮箱ID",
			Error:   err.Error(),
		})
		return
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	status, exists := req["status"]
	if !exists {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "缺少status参数",
		})
		return
	}

	// 调用服务层
	err = h.mailboxService.UpdateStatus(uint(id), status)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "邮箱配置不存在" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "无效的状态值" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "更新邮箱状态失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "更新邮箱状态成功",
	})
}

// DiagnoseMailbox 诊断邮箱连接问题
// @Summary 诊断邮箱连接问题
// @Description 详细诊断邮箱连接问题，提供解决建议
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param id path int true "邮箱ID"
// @Success 200 {object} APIResponse{data=email.EmailDiagnosis}
// @Router /api/v1/mailboxes/{id}/diagnose [post]
func (h *MailboxHandler) DiagnoseMailbox(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的邮箱ID",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层诊断邮箱
	diagnosis, err := h.mailboxService.DiagnoseMailbox(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "获取邮箱配置失败: 邮箱配置不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "诊断邮箱失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "邮箱诊断完成",
		Data:    diagnosis,
	})
}

// DiagnoseMailboxConfig 诊断邮箱配置
// @Summary 诊断邮箱配置
// @Description 使用提供的配置参数诊断邮箱连接问题，提供解决建议
// @Tags 邮箱管理
// @Accept json
// @Produce json
// @Param config body service.CreateMailboxRequest true "邮箱配置参数"
// @Success 200 {object} APIResponse{data=email.EmailDiagnosis}
// @Router /api/v1/mailboxes/diagnose-config [post]
func (h *MailboxHandler) DiagnoseMailboxConfig(c *gin.Context) {
	var req service.CreateMailboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层诊断邮箱配置
	diagnosis := h.mailboxService.DiagnoseMailboxWithConfig(&req)

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "邮箱配置诊断完成",
		Data:    diagnosis,
	})
}
