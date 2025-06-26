package api

import (
	"emailAlert/internal/model"
	"emailAlert/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TemplateHandler 模版管理处理器
type TemplateHandler struct {
	templateService *service.TemplateService
}

// NewTemplateHandler 创建模版管理处理器
func NewTemplateHandler(templateService *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

// GetTemplates 获取模版列表
// @Summary 获取模版列表
// @Description 分页获取模版列表，支持类型和状态过滤
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param type query string false "模版类型" Enums(email, dingtalk, wechat, markdown)
// @Param status query string false "状态过滤" Enums(active, inactive)
// @Success 200 {object} APIResponse{data=service.TemplateListResponse}
// @Router /api/v1/templates [get]
func (h *TemplateHandler) GetTemplates(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	templateType := c.Query("type")
	status := c.Query("status")

	// 调用服务层
	result, err := h.templateService.List(page, size, templateType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "获取模版列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取模版列表成功",
		Data:    result,
	})
}

// CreateTemplate 创建模版
// @Summary 创建模版
// @Description 创建新的消息模版
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param template body service.CreateTemplateRequest true "模版信息"
// @Success 201 {object} APIResponse{data=model.Template}
// @Router /api/v1/templates [post]
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req service.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	template, err := h.templateService.Create(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "模版名称已存在" {
			statusCode = http.StatusConflict
		} else if err.Error() == "模版语法错误" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "创建模版失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Code:    201,
		Message: "创建模版成功",
		Data:    template,
	})
}

// GetTemplate 获取模版详情
// @Summary 获取模版详情
// @Description 根据ID获取模版详情
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param id path int true "模版ID"
// @Success 200 {object} APIResponse{data=model.Template}
// @Router /api/v1/templates/{id} [get]
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的模版ID",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	template, err := h.templateService.GetByID(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "模版不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "获取模版详情失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取模版详情成功",
		Data:    template,
	})
}

// UpdateTemplate 更新模版
// @Summary 更新模版
// @Description 更新模版信息
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param id path int true "模版ID"
// @Param template body service.UpdateTemplateRequest true "模版信息"
// @Success 200 {object} APIResponse{data=model.Template}
// @Router /api/v1/templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的模版ID",
			Error:   err.Error(),
		})
		return
	}

	var req service.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	template, err := h.templateService.Update(uint(id), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "模版不存在" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "模版名称已被其他模版使用" {
			statusCode = http.StatusConflict
		} else if err.Error() == "模版语法错误" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "更新模版失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "更新模版成功",
		Data:    template,
	})
}

// DeleteTemplate 删除模版
// @Summary 删除模版
// @Description 删除指定的模版
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param id path int true "模版ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的模版ID",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	err = h.templateService.Delete(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "模版不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "删除模版失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "删除模版成功",
	})
}

// SetDefaultTemplate 设置默认模版
// @Summary 设置默认模版
// @Description 设置指定模版为默认模版
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param id path int true "模版ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/templates/{id}/default [put]
func (h *TemplateHandler) SetDefaultTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的模版ID",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	err = h.templateService.SetDefault(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "模版不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "设置默认模版失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "设置默认模版成功",
	})
}

// PreviewTemplate 预览模版
// @Summary 预览模版
// @Description 预览模版渲染效果
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param preview body service.TemplatePreviewRequest true "预览请求"
// @Success 200 {object} APIResponse{data=service.TemplatePreviewResponse}
// @Router /api/v1/templates/preview [post]
func (h *TemplateHandler) PreviewTemplate(c *gin.Context) {
	var req service.TemplatePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "请求参数无效",
			Error:   err.Error(),
		})
		return
	}

	// 调用服务层
	result, err := h.templateService.Preview(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "模版预览失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "模版预览成功",
		Data:    result,
	})
}

// RenderTemplate 渲染模版
// @Summary 渲染模版
// @Description 使用指定数据渲染模版
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param id path int true "模版ID"
// @Param data body model.TemplateRenderData true "渲染数据"
// @Success 200 {object} APIResponse{data=service.TemplatePreviewResponse}
// @Router /api/v1/templates/{id}/render [post]
func (h *TemplateHandler) RenderTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "无效的模版ID",
			Error:   err.Error(),
		})
		return
	}

	var renderData *model.TemplateRenderData

	// 检查请求体是否为空
	if c.Request.Body != nil {
		var tempData model.TemplateRenderData
		if err := c.ShouldBindJSON(&tempData); err != nil {
			// 如果解析失败，使用nil让服务层提供默认数据
			renderData = nil
		} else {
			// 检查是否为空数据（所有字段都是默认值）
			if tempData.Email == nil && tempData.Alert == nil && tempData.Rule == nil && tempData.Mailbox == nil {
				renderData = nil
			} else {
				renderData = &tempData
			}
		}
	}

	// 调用服务层
	result, err := h.templateService.Render(uint(id), renderData)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "模版不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "渲染模版失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "渲染模版成功",
		Data:    result,
	})
}

// GetTemplatesByType 根据类型获取模版列表
// @Summary 根据类型获取模版列表
// @Description 获取指定类型的所有活跃模版
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param type path string true "模版类型" Enums(email, dingtalk, wechat, markdown)
// @Success 200 {object} APIResponse{data=[]model.Template}
// @Router /api/v1/templates/type/{type} [get]
func (h *TemplateHandler) GetTemplatesByType(c *gin.Context) {
	templateType := c.Param("type")
	if templateType == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "模版类型不能为空",
		})
		return
	}

	// 调用服务层
	templates, err := h.templateService.GetByType(templateType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    500,
			Message: "获取模版列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取模版列表成功",
		Data:    templates,
	})
}

// GetDefaultTemplate 获取默认模版
// @Summary 获取默认模版
// @Description 获取指定类型的默认模版
// @Tags 模版管理
// @Accept json
// @Produce json
// @Param type path string true "模版类型" Enums(email, dingtalk, wechat, markdown)
// @Success 200 {object} APIResponse{data=model.Template}
// @Router /api/v1/templates/default/{type} [get]
func (h *TemplateHandler) GetDefaultTemplate(c *gin.Context) {
	templateType := c.Param("type")
	if templateType == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    400,
			Message: "模版类型不能为空",
		})
		return
	}

	// 调用服务层
	template, err := h.templateService.GetDefaultByType(templateType)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "默认模版不存在" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Code:    statusCode,
			Message: "获取默认模版失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取默认模版成功",
		Data:    template,
	})
}

// GetAvailableVariables 获取可用变量列表
// @Summary 获取可用变量列表
// @Description 获取模版中可用的变量列表
// @Tags 模版管理
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]model.TemplateVariable}
// @Router /api/v1/templates/variables [get]
func (h *TemplateHandler) GetAvailableVariables(c *gin.Context) {
	// 调用服务层
	variables := h.templateService.GetAvailableVariables()

	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "获取可用变量成功",
		Data:    variables,
	})
}
