package service

import (
	"bytes"
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// TemplateService 模版服务层
type TemplateService struct {
	templateRepo *repository.TemplateRepository
}

// NewTemplateService 创建模版服务实例
func NewTemplateService(templateRepo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{templateRepo: templateRepo}
}

// CreateTemplateRequest 创建模版请求
type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=email dingtalk wechat markdown"`
	Subject     string `json:"subject"`
	Content     string `json:"content" binding:"required"`
	Variables   string `json:"variables"`
	IsDefault   bool   `json:"is_default"`
	Description string `json:"description"`
}

// UpdateTemplateRequest 更新模版请求
type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type" binding:"omitempty,oneof=email dingtalk wechat markdown"`
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	Variables   string `json:"variables"`
	IsDefault   bool   `json:"is_default"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
	Description string `json:"description"`
}

// TemplateListResponse 模版列表响应
type TemplateListResponse struct {
	List  []model.Template `json:"list"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

// TemplatePreviewRequest 模版预览请求
type TemplatePreviewRequest struct {
	Content    string                    `json:"content" binding:"required"`
	Subject    string                    `json:"subject"`
	RenderData *model.TemplateRenderData `json:"render_data"`
}

// TemplatePreviewResponse 模版预览响应
type TemplatePreviewResponse struct {
	Subject       string                   `json:"subject"`
	Content       string                   `json:"content"`
	UsedVars      []string                 `json:"used_vars"`
	AvailableVars []model.TemplateVariable `json:"available_vars"`
}

// Create 创建模版
func (s *TemplateService) Create(req *CreateTemplateRequest) (*model.Template, error) {
	// 检查模版名称是否已存在
	exists, err := s.templateRepo.NameExists(req.Name, 0)
	if err != nil {
		return nil, fmt.Errorf("检查模版名称失败: %v", err)
	}
	if exists {
		return nil, errors.New("模版名称已存在")
	}

	// 验证模版内容语法
	if err := s.validateTemplate(req.Content); err != nil {
		return nil, fmt.Errorf("模版语法错误: %v", err)
	}

	// 如果有主题模版，也需要验证
	if req.Subject != "" {
		if err := s.validateTemplate(req.Subject); err != nil {
			return nil, fmt.Errorf("主题模版语法错误: %v", err)
		}
	}

	// 创建模版模型
	template := &model.Template{
		Name:        req.Name,
		Type:        req.Type,
		Subject:     req.Subject,
		Content:     req.Content,
		Variables:   req.Variables,
		IsDefault:   req.IsDefault,
		Status:      "active",
		Description: req.Description,
	}

	// 如果设置为默认模版，需要先取消同类型的其他默认模版
	if req.IsDefault {
		if err := s.templateRepo.SetDefault(0, req.Type); err != nil {
			return nil, fmt.Errorf("更新默认模版状态失败: %v", err)
		}
	}

	// 保存到数据库
	err = s.templateRepo.Create(template)
	if err != nil {
		return nil, fmt.Errorf("创建模版失败: %v", err)
	}

	return template, nil
}

// GetByID 根据ID获取模版
func (s *TemplateService) GetByID(id uint) (*model.Template, error) {
	return s.templateRepo.GetByID(id)
}

// List 获取模版列表
func (s *TemplateService) List(page, pageSize int, templateType, status string) (*TemplateListResponse, error) {
	// 参数验证
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	templates, total, err := s.templateRepo.List(page, pageSize, templateType, status)
	if err != nil {
		return nil, fmt.Errorf("获取模版列表失败: %v", err)
	}

	return &TemplateListResponse{
		List:  templates,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

// Update 更新模版
func (s *TemplateService) Update(id uint, req *UpdateTemplateRequest) (*model.Template, error) {
	// 获取原有模版
	existingTemplate, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 检查模版名称是否被其他模版使用
	if req.Name != "" && req.Name != existingTemplate.Name {
		exists, err := s.templateRepo.NameExists(req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("检查模版名称失败: %v", err)
		}
		if exists {
			return nil, errors.New("模版名称已被其他模版使用")
		}
	}

	// 验证模版内容语法
	if req.Content != "" {
		if err := s.validateTemplate(req.Content); err != nil {
			return nil, fmt.Errorf("模版语法错误: %v", err)
		}
	}

	// 如果有主题模版，也需要验证
	if req.Subject != "" {
		if err := s.validateTemplate(req.Subject); err != nil {
			return nil, fmt.Errorf("主题模版语法错误: %v", err)
		}
	}

	// 准备更新数据
	updateData := &model.Template{}

	if req.Name != "" {
		updateData.Name = req.Name
	}
	if req.Type != "" {
		updateData.Type = req.Type
	}
	if req.Subject != "" {
		updateData.Subject = req.Subject
	}
	if req.Content != "" {
		updateData.Content = req.Content
	}
	if req.Variables != "" {
		updateData.Variables = req.Variables
	}
	updateData.IsDefault = req.IsDefault
	if req.Status != "" {
		updateData.Status = req.Status
	}
	updateData.Description = req.Description

	// 如果设置为默认模版，需要先取消同类型的其他默认模版
	templateType := req.Type
	if templateType == "" {
		templateType = existingTemplate.Type
	}

	if req.IsDefault {
		if err := s.templateRepo.SetDefault(id, templateType); err != nil {
			return nil, fmt.Errorf("更新默认模版状态失败: %v", err)
		}
	}

	// 更新数据库
	err = s.templateRepo.Update(id, updateData)
	if err != nil {
		return nil, fmt.Errorf("更新模版失败: %v", err)
	}

	// 返回更新后的模版
	return s.GetByID(id)
}

// Delete 删除模版
func (s *TemplateService) Delete(id uint) error {
	// 检查模版是否存在
	_, err := s.templateRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 执行删除
	return s.templateRepo.Delete(id)
}

// SetDefault 设置默认模版
func (s *TemplateService) SetDefault(id uint) error {
	// 获取模版信息
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 设置为默认模版
	return s.templateRepo.SetDefault(id, template.Type)
}

// GetDefaultByType 获取指定类型的默认模版
func (s *TemplateService) GetDefaultByType(templateType string) (*model.Template, error) {
	return s.templateRepo.GetDefaultByType(templateType)
}

// GetByType 根据类型获取模版列表
func (s *TemplateService) GetByType(templateType string) ([]model.Template, error) {
	return s.templateRepo.GetByType(templateType)
}

// Preview 预览模版
func (s *TemplateService) Preview(req *TemplatePreviewRequest) (*TemplatePreviewResponse, error) {
	// 如果没有提供渲染数据，使用默认数据
	if req.RenderData == nil {
		req.RenderData = s.getDefaultRenderData()
	}

	// 渲染内容
	content, usedVars, err := s.renderTemplate(req.Content, req.RenderData)
	if err != nil {
		return nil, fmt.Errorf("渲染内容失败: %v", err)
	}

	// 渲染主题（如果有）
	subject := ""
	if req.Subject != "" {
		subject, _, err = s.renderTemplate(req.Subject, req.RenderData)
		if err != nil {
			return nil, fmt.Errorf("渲染主题失败: %v", err)
		}
	}

	// 获取可用变量
	availableVars := s.getAvailableVariables()

	return &TemplatePreviewResponse{
		Subject:       subject,
		Content:       content,
		UsedVars:      usedVars,
		AvailableVars: availableVars,
	}, nil
}

// Render 渲染模版
func (s *TemplateService) Render(templateID uint, renderData *model.TemplateRenderData) (*TemplatePreviewResponse, error) {
	// 获取模版
	tmpl, err := s.templateRepo.GetByID(templateID)
	if err != nil {
		return nil, err
	}

	// 如果没有提供渲染数据，使用默认数据
	if renderData == nil {
		renderData = s.getDefaultRenderData()
	}

	// 渲染模版
	return s.Preview(&TemplatePreviewRequest{
		Content:    tmpl.Content,
		Subject:    tmpl.Subject,
		RenderData: renderData,
	})
}

// RenderByType 使用默认模版渲染
func (s *TemplateService) RenderByType(templateType string, renderData *model.TemplateRenderData) (*TemplatePreviewResponse, error) {
	// 获取默认模版
	tmpl, err := s.templateRepo.GetDefaultByType(templateType)
	if err != nil {
		return nil, err
	}

	// 渲染模版
	return s.Preview(&TemplatePreviewRequest{
		Content:    tmpl.Content,
		Subject:    tmpl.Subject,
		RenderData: renderData,
	})
}

// GetAvailableVariables 获取可用变量列表
func (s *TemplateService) GetAvailableVariables() []model.TemplateVariable {
	return s.getAvailableVariables()
}

// validateTemplate 验证模版语法
func (s *TemplateService) validateTemplate(content string) error {
	_, err := template.New("test").Parse(content)
	return err
}

// renderTemplate 渲染模版
func (s *TemplateService) renderTemplate(content string, data *model.TemplateRenderData) (string, []string, error) {
	tmpl, err := template.New("template").Parse(content)
	if err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", nil, err
	}

	// 提取使用的变量
	usedVars := s.extractUsedVariables(content)

	return buf.String(), usedVars, nil
}

// extractUsedVariables 提取模版中使用的变量
func (s *TemplateService) extractUsedVariables(content string) []string {
	var vars []string

	// 简单的变量提取逻辑，查找 {{ .xxx }} 格式的变量
	parts := strings.Split(content, "{{")
	for _, part := range parts {
		if strings.Contains(part, "}}") {
			varPart := strings.Split(part, "}}")[0]
			varPart = strings.TrimSpace(varPart)
			if strings.HasPrefix(varPart, ".") {
				vars = append(vars, varPart)
			}
		}
	}

	return vars
}

// getDefaultRenderData 获取默认渲染数据
func (s *TemplateService) getDefaultRenderData() *model.TemplateRenderData {
	now := time.Now()
	return &model.TemplateRenderData{
		Email: &model.EmailData{
			Subject:     "系统告警邮件",
			Sender:      "system@example.com",
			Content:     "这是一封示例邮件内容，用于模版预览。",
			HTMLContent: "这是一封示例邮件内容，用于模版预览。",
			ReceivedAt:  now,
			MessageID:   "example-message-id",
			Size:        1024,
		},
		Alert: &model.Alert{
			Subject:    "系统告警邮件",
			Sender:     "system@example.com",
			Content:    "这是一封示例邮件内容，用于模版预览。",
			ReceivedAt: now,
			Status:     "active",
		},
		Rule: &model.AlertRule{
			Name:        "示例告警规则",
			MatchType:   "keyword",
			MatchValue:  "错误,异常",
			FieldType:   "subject",
			Priority:    5,
			Description: "示例规则描述",
		},
		Mailbox: &model.Mailbox{
			Name:  "示例邮箱",
			Email: "monitor@example.com",
			Host:  "imap.example.com",
			Port:  993,
		},
		System: model.SystemInfo{
			AppName:     "邮件告警平台",
			AppVersion:  "1.0.0",
			ServerName:  "localhost",
			Environment: "development",
		},
		Time: model.TimeInfo{
			Now:       now,
			NowFormat: now.Format("2006-01-02 15:04:05"),
			NowUnix:   now.Unix(),
			Today:     now.Format("2006-01-02"),
			Yesterday: now.AddDate(0, 0, -1).Format("2006-01-02"),
		},
	}
}

// getAvailableVariables 获取可用变量列表
func (s *TemplateService) getAvailableVariables() []model.TemplateVariable {
	vars := []model.TemplateVariable{
		// 邮件变量
		{Name: ".Email.Subject", Description: "邮件主题", Example: "系统告警邮件", Category: "email"},
		{Name: ".Email.Sender", Description: "发件人", Example: "admin@example.com", Category: "email"},
		{Name: ".Email.Content", Description: "邮件内容", Example: "邮件正文内容", Category: "email"},
		{Name: ".Email.ReceivedAt", Description: "接收时间", Example: "2024-01-01 12:00:00", Category: "email"},
		{Name: ".Email.MessageID", Description: "邮件ID", Example: "message-id-123", Category: "email"},

		// 告警变量
		{Name: ".Alert.Subject", Description: "告警主题", Example: "系统告警", Category: "alert"},
		{Name: ".Alert.Sender", Description: "告警发送者", Example: "system@example.com", Category: "alert"},
		{Name: ".Alert.Content", Description: "告警内容", Example: "告警详细信息", Category: "alert"},
		{Name: ".Alert.Status", Description: "告警状态", Example: "pending", Category: "alert"},
		{Name: ".Alert.ReceivedAt", Description: "告警时间", Example: "2024-01-01 12:00:00", Category: "alert"},

		// 规则变量
		{Name: ".Rule.Name", Description: "规则名称", Example: "生产环境告警", Category: "rule"},
		{Name: ".Rule.MatchType", Description: "匹配类型", Example: "keyword", Category: "rule"},
		{Name: ".Rule.MatchValue", Description: "匹配值", Example: "错误,异常", Category: "rule"},
		{Name: ".Rule.Priority", Description: "规则优先级", Example: "9", Category: "rule"},

		// 邮箱变量
		{Name: ".Mailbox.Name", Description: "邮箱名称", Example: "监控邮箱", Category: "mailbox"},
		{Name: ".Mailbox.Email", Description: "邮箱地址", Example: "monitor@example.com", Category: "mailbox"},

		// 系统变量
		{Name: ".System.AppName", Description: "应用名称", Example: "邮件告警平台", Category: "system"},
		{Name: ".System.AppVersion", Description: "应用版本", Example: "1.0.0", Category: "system"},
		{Name: ".System.ServerName", Description: "服务器名称", Example: "localhost", Category: "system"},
		{Name: ".System.Environment", Description: "运行环境", Example: "production", Category: "system"},

		// 时间变量
		{Name: ".Time.Now", Description: "当前时间", Example: "2024-01-01 12:00:00", Category: "time"},
		{Name: ".Time.NowFormat", Description: "格式化当前时间", Example: "2024-01-01 12:00:00", Category: "time"},
		{Name: ".Time.Today", Description: "今天日期", Example: "2024-01-01", Category: "time"},
		{Name: ".Time.Yesterday", Description: "昨天日期", Example: "2023-12-31", Category: "time"},
	}

	return vars
}
