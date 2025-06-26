package service

import (
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"log"
)

// TemplateInitService 模版初始化服务
type TemplateInitService struct {
	templateRepo *repository.TemplateRepository
}

// NewTemplateInitService 创建模版初始化服务
func NewTemplateInitService(templateRepo *repository.TemplateRepository) *TemplateInitService {
	return &TemplateInitService{templateRepo: templateRepo}
}

// InitDefaultTemplates 初始化默认模版
func (s *TemplateInitService) InitDefaultTemplates() error {
	// 检查是否已经有默认模版
	templates, err := s.templateRepo.GetActiveTemplates()
	if err != nil {
		return err
	}

	// 如果已经有模版，跳过初始化
	if len(templates) > 0 {
		log.Println("默认模版已存在，跳过初始化")
		return nil
	}

	// 创建默认模版
	defaultTemplates := s.getDefaultTemplates()

	for _, template := range defaultTemplates {
		// 检查同名模版是否存在
		exists, err := s.templateRepo.NameExists(template.Name, 0)
		if err != nil {
			log.Printf("检查模版 %s 是否存在失败: %v", template.Name, err)
			continue
		}

		if exists {
			log.Printf("模版 %s 已存在，跳过创建", template.Name)
			continue
		}

		// 创建模版
		if err := s.templateRepo.Create(&template); err != nil {
			log.Printf("创建默认模版 %s 失败: %v", template.Name, err)
			continue
		}

		log.Printf("成功创建默认模版: %s", template.Name)
	}

	return nil
}

// getDefaultTemplates 获取默认模版列表
func (s *TemplateInitService) getDefaultTemplates() []model.Template {
	return []model.Template{
		// 邮件默认模版
		{
			Name:        "邮件告警默认模版",
			Type:        "email",
			Subject:     "【{{.Alert.Status}}】{{.System.AppName}} - {{.Email.Subject}}",
			Content:     s.getEmailDefaultContent(),
			Variables:   s.getEmailVariables(),
			IsDefault:   true,
			Status:      "active",
			Description: "邮件告警的默认模版，包含完整的告警信息展示",
		},
		// 钉钉默认模版
		{
			Name:        "钉钉告警默认模版",
			Type:        "dingtalk",
			Content:     s.getDingTalkDefaultContent(),
			Variables:   s.getDingTalkVariables(),
			IsDefault:   true,
			Status:      "active",
			Description: "钉钉机器人消息的默认模版，支持Markdown格式",
		},
		// 企业微信默认模版
		{
			Name:        "企业微信告警默认模版",
			Type:        "wechat",
			Content:     s.getWeChatDefaultContent(),
			Variables:   s.getWeChatVariables(),
			IsDefault:   true,
			Status:      "active",
			Description: "企业微信消息的默认模版，支持文本和Markdown格式",
		},
		// Markdown格式模版
		{
			Name:        "Markdown告警模版",
			Type:        "markdown",
			Content:     s.getMarkdownDefaultContent(),
			Variables:   s.getMarkdownVariables(),
			IsDefault:   true,
			Status:      "active",
			Description: "通用Markdown格式告警模版，适用于多种通知渠道",
		},
		// 简单文本模版
		{
			Name:        "简单邮件模版",
			Type:        "email",
			Subject:     "告警通知 - {{.Email.Subject}}",
			Content:     s.getSimpleEmailContent(),
			Variables:   s.getSimpleVariables(),
			IsDefault:   false,
			Status:      "active",
			Description: "简化的邮件告警模版，只包含核心信息",
		},
	}
}

// getEmailDefaultContent 获取邮件默认模版内容
func (s *TemplateInitService) getEmailDefaultContent() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>告警通知</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #d73502; border-bottom: 2px solid #d73502; padding-bottom: 10px;">
            🚨 系统告警通知
        </h2>
        
        <div style="background-color: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0; color: #495057;">告警详情</h3>
            <p><strong>邮件主题：</strong>{{.Email.Subject}}</p>
            <p><strong>发件人：</strong>{{.Email.Sender}}</p>
            <p><strong>接收时间：</strong>{{.Email.ReceivedAt}}</p>
            <p><strong>匹配规则：</strong>{{.Rule.Name}}</p>
        </div>
        
        <div style="background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0; color: #856404;">邮件内容</h3>
            <div style="background-color: white; padding: 10px; border-radius: 3px; white-space: pre-wrap;">{{.Email.Content}}</div>
        </div>
        
        <div style="background-color: #e9ecef; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0; color: #495057;">系统信息</h3>
            <p><strong>系统：</strong>{{.System.AppName}} v{{.System.AppVersion}}</p>
            <p><strong>服务器：</strong>{{.System.ServerName}}</p>
            <p><strong>环境：</strong>{{.System.Environment}}</p>
            <p><strong>通知时间：</strong>{{.Time.NowFormat}}</p>
        </div>
        
        <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #dee2e6; color: #6c757d; font-size: 14px;">
            此邮件由 {{.System.AppName}} 自动发送，请勿回复
        </div>
    </div>
</body>
</html>`
}

// getDingTalkDefaultContent 获取钉钉默认模版内容
func (s *TemplateInitService) getDingTalkDefaultContent() string {
	return `## 🚨 系统告警通知

**告警详情**
- **邮件主题：** {{.Email.Subject}}
- **发件人：** {{.Email.Sender}}
- **接收时间：** {{.Email.ReceivedAt}}
- **匹配规则：** {{.Rule.Name}}

**邮件内容**
> {{.Email.Content}}

**系统信息**
- **系统：** {{.System.AppName}} v{{.System.AppVersion}}
- **服务器：** {{.System.ServerName}}
- **环境：** {{.System.Environment}}
- **通知时间：** {{.Time.NowFormat}}

---
*此消息由 {{.System.AppName}} 自动发送*`
}

// getWeChatDefaultContent 获取企业微信默认模版内容
func (s *TemplateInitService) getWeChatDefaultContent() string {
	return `【系统告警通知】

告警详情：
邮件主题：{{.Email.Subject}}
发件人：{{.Email.Sender}}
接收时间：{{.Email.ReceivedAt}}
匹配规则：{{.Rule.Name}}

邮件内容：
{{.Email.Content}}

系统信息：
系统：{{.System.AppName}} v{{.System.AppVersion}}
服务器：{{.System.ServerName}}
环境：{{.System.Environment}}
通知时间：{{.Time.NowFormat}}

此消息由 {{.System.AppName}} 自动发送`
}

// getMarkdownDefaultContent 获取Markdown默认模版内容
func (s *TemplateInitService) getMarkdownDefaultContent() string {
	return `# 🚨 系统告警通知

## 告警详情

| 项目 | 内容 |
|------|------|
| 邮件主题 | {{.Email.Subject}} |
| 发件人 | {{.Email.Sender}} |
| 接收时间 | {{.Email.ReceivedAt}} |
| 匹配规则 | {{.Rule.Name}} |

## 邮件内容

` + "```" + `
{{.Email.Content}}
` + "```" + `

## 系统信息

- **系统：** {{.System.AppName}} v{{.System.AppVersion}}
- **服务器：** {{.System.ServerName}}
- **环境：** {{.System.Environment}}
- **通知时间：** {{.Time.NowFormat}}

---

*此消息由 {{.System.AppName}} 自动发送*`
}

// getSimpleEmailContent 获取简单邮件模版内容
func (s *TemplateInitService) getSimpleEmailContent() string {
	return `告警通知

邮件主题：{{.Email.Subject}}
发件人：{{.Email.Sender}}
接收时间：{{.Email.ReceivedAt}}

邮件内容：
{{.Email.Content}}

通知时间：{{.Time.NowFormat}}
系统：{{.System.AppName}}`
}

// getEmailVariables 获取邮件模版变量说明
func (s *TemplateInitService) getEmailVariables() string {
	// 这里简化处理，实际项目中可以使用JSON序列化存储详细的变量信息
	return "邮件告警模版变量说明，支持邮件、告警、规则、系统、时间等分类变量"
}

// getDingTalkVariables 获取钉钉模版变量说明
func (s *TemplateInitService) getDingTalkVariables() string {
	return "钉钉Markdown模版变量说明，支持所有系统变量和Markdown格式"
}

// getWeChatVariables 获取企业微信模版变量说明
func (s *TemplateInitService) getWeChatVariables() string {
	return "企业微信文本模版变量说明，支持纯文本格式的所有系统变量"
}

// getMarkdownVariables 获取Markdown模版变量说明
func (s *TemplateInitService) getMarkdownVariables() string {
	return "通用Markdown模版变量说明，支持表格、代码块等Markdown语法"
}

// getSimpleVariables 获取简单模版变量说明
func (s *TemplateInitService) getSimpleVariables() string {
	return "简化模版变量说明，只包含核心的邮件和系统信息变量"
}
