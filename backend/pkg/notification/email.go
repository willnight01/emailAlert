package notification

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	// 可以添加连接池等高级功能
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier() *EmailNotifier {
	return &EmailNotifier{}
}

// EmailConfig 邮件配置结构
type EmailConfig struct {
	Host     string   `json:"host"`                // SMTP服务器地址
	Port     int      `json:"port"`                // SMTP端口
	Username string   `json:"username"`            // 用户名
	Password string   `json:"password"`            // 密码或应用专用密码
	SSL      bool     `json:"ssl"`                 // 是否启用SSL/TLS
	From     string   `json:"from"`                // 发件人地址
	FromName string   `json:"from_name,omitempty"` // 发件人姓名
	To       []string `json:"to"`                  // 收件人列表
	CC       []string `json:"cc,omitempty"`        // 抄送列表
	BCC      []string `json:"bcc,omitempty"`       // 密送列表
	Subject  string   `json:"subject,omitempty"`   // 邮件主题模板
	Template string   `json:"template,omitempty"`  // 邮件内容模板
	Format   string   `json:"format,omitempty"`    // 邮件格式：text/html/mixed
	Timeout  int      `json:"timeout,omitempty"`   // 连接超时时间（秒）
	ReplyTo  string   `json:"reply_to,omitempty"`  // 回复地址
	Priority int      `json:"priority,omitempty"`  // 优先级：1(高) 2(普通) 3(低)
}

// EmailMessage 邮件消息结构
type EmailMessage struct {
	From        string            `json:"from"`
	FromName    string            `json:"from_name,omitempty"`
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	TextContent string            `json:"text_content,omitempty"`
	HTMLContent string            `json:"html_content,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	Priority    int               `json:"priority,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
}

// EmailAttachment 邮件附件结构
type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content"`
}

// EmailResponse 邮件发送响应
type EmailResponse struct {
	Success    bool          `json:"success"`
	MessageID  string        `json:"message_id,omitempty"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	Recipients int           `json:"recipients"`
}

// SendMessage 发送邮件消息
func (e *EmailNotifier) SendMessage(config *EmailConfig, title, content string) (*EmailResponse, error) {
	startTime := time.Now()

	// 验证配置
	if err := e.ValidateConfig(config); err != nil {
		return &EmailResponse{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}, err
	}

	// 构建邮件消息
	message, err := e.buildMessage(config, title, content)
	if err != nil {
		return &EmailResponse{
			Success:  false,
			Error:    fmt.Sprintf("构建邮件失败: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	// 发送邮件
	err = e.sendSMTP(config, message)
	if err != nil {
		return &EmailResponse{
			Success:    false,
			Error:      fmt.Sprintf("发送邮件失败: %v", err),
			Duration:   time.Since(startTime),
			Recipients: len(message.To) + len(message.CC) + len(message.BCC),
		}, err
	}

	return &EmailResponse{
		Success:    true,
		Duration:   time.Since(startTime),
		Recipients: len(message.To) + len(message.CC) + len(message.BCC),
		MessageID:  fmt.Sprintf("<%d@%s>", time.Now().Unix(), "emailalert"),
	}, nil
}

// buildMessage 构建邮件消息
func (e *EmailNotifier) buildMessage(config *EmailConfig, title, content string) (*EmailMessage, error) {
	message := &EmailMessage{
		From:     config.From,
		FromName: config.FromName,
		To:       config.To,
		CC:       config.CC,
		BCC:      config.BCC,
		ReplyTo:  config.ReplyTo,
		Priority: config.Priority,
		Headers:  make(map[string]string),
	}

	// 处理主题
	if config.Subject != "" {
		message.Subject = e.renderTemplate(config.Subject, title, content)
	} else {
		message.Subject = title
	}

	// 处理内容
	if config.Template != "" {
		renderedContent := e.renderTemplate(config.Template, title, content)
		// 根据格式设置内容
		switch strings.ToLower(config.Format) {
		case "html":
			message.HTMLContent = renderedContent
		case "text":
			message.TextContent = renderedContent
		case "mixed", "":
			// 检查内容是否包含HTML标签
			if e.containsHTML(renderedContent) {
				message.HTMLContent = renderedContent
				message.TextContent = e.stripHTML(renderedContent)
			} else {
				message.TextContent = renderedContent
			}
		}
	} else {
		// 使用默认格式
		if e.containsHTML(content) {
			message.HTMLContent = e.buildDefaultHTMLContent(title, content)
			message.TextContent = fmt.Sprintf("%s\n\n%s", title, e.stripHTML(content))
		} else {
			message.TextContent = fmt.Sprintf("%s\n\n%s", title, content)
		}
	}

	return message, nil
}

// renderTemplate 渲染邮件模板
func (e *EmailNotifier) renderTemplate(template, title, content string) string {
	result := template
	result = strings.ReplaceAll(result, "{{title}}", title)
	result = strings.ReplaceAll(result, "{{content}}", content)
	result = strings.ReplaceAll(result, "{{timestamp}}", time.Now().Format("2006-01-02 15:04:05"))
	result = strings.ReplaceAll(result, "{{date}}", time.Now().Format("2006-01-02"))
	result = strings.ReplaceAll(result, "{{time}}", time.Now().Format("15:04:05"))
	return result
}

// containsHTML 检查内容是否包含HTML标签
func (e *EmailNotifier) containsHTML(content string) bool {
	htmlTags := []string{"<html>", "<body>", "<div>", "<p>", "<br>", "<h1>", "<h2>", "<h3>", "<strong>", "<em>", "<ul>", "<ol>", "<li>", "<a>", "<img>"}
	lowerContent := strings.ToLower(content)
	for _, tag := range htmlTags {
		if strings.Contains(lowerContent, tag) {
			return true
		}
	}
	return false
}

// stripHTML 简单的HTML标签移除
func (e *EmailNotifier) stripHTML(content string) string {
	// 简单的HTML标签移除（实际项目中可以使用专门的HTML解析库）
	result := content
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br />", "\n")
	result = strings.ReplaceAll(result, "</p>", "\n\n")
	result = strings.ReplaceAll(result, "</div>", "\n")
	result = strings.ReplaceAll(result, "</h1>", "\n")
	result = strings.ReplaceAll(result, "</h2>", "\n")
	result = strings.ReplaceAll(result, "</h3>", "\n")

	// 移除其他HTML标签
	for strings.Contains(result, "<") && strings.Contains(result, ">") {
		start := strings.Index(result, "<")
		end := strings.Index(result[start:], ">")
		if end != -1 {
			result = result[:start] + result[start+end+1:]
		} else {
			break
		}
	}

	return strings.TrimSpace(result)
}

// buildDefaultHTMLContent 构建默认HTML内容
func (e *EmailNotifier) buildDefaultHTMLContent(title, content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .header { background-color: #f4f4f4; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .content { padding: 20px; }
        .footer { margin-top: 20px; padding: 10px; font-size: 12px; color: #666; border-top: 1px solid #eee; }
    </style>
</head>
<body>
    <div class="header">
        <h2>%s</h2>
    </div>
    <div class="content">
        %s
    </div>
    <div class="footer">
        <p>此邮件由邮件告警平台自动发送，请勿回复。</p>
        <p>发送时间: %s</p>
    </div>
</body>
</html>`, title, title, content, time.Now().Format("2006-01-02 15:04:05"))
}

// sendSMTP 通过SMTP发送邮件
func (e *EmailNotifier) sendSMTP(config *EmailConfig, message *EmailMessage) error {
	// 构建SMTP地址
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// 设置超时 - 增加到60秒
	timeout := 60 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	// 建立SMTP连接（createSMTPClient已经包含了身份验证）
	client, err := e.createSMTPClient(config, addr, timeout)
	if err != nil {
		return err
	}
	defer client.Close()

	// 设置发件人
	if err = client.Mail(message.From); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	allRecipients := make([]string, 0)
	allRecipients = append(allRecipients, message.To...)
	allRecipients = append(allRecipients, message.CC...)
	allRecipients = append(allRecipients, message.BCC...)

	for _, recipient := range allRecipients {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("设置收件人失败 (%s): %v", recipient, err)
		}
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始发送邮件数据失败: %v", err)
	}
	defer writer.Close()

	// 构建邮件头和内容
	emailContent := e.buildEmailContent(message)
	if _, err = writer.Write([]byte(emailContent)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}

	return nil
}

// buildEmailContent 构建邮件内容
func (e *EmailNotifier) buildEmailContent(message *EmailMessage) string {
	var content strings.Builder

	// 邮件头
	content.WriteString(fmt.Sprintf("From: %s\r\n", e.formatEmailAddress(message.From, message.FromName)))
	content.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(message.To, ", ")))

	if len(message.CC) > 0 {
		content.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(message.CC, ", ")))
	}

	if message.ReplyTo != "" {
		content.WriteString(fmt.Sprintf("Reply-To: %s\r\n", message.ReplyTo))
	}

	// 主题（编码处理中文）
	content.WriteString(fmt.Sprintf("Subject: %s\r\n", e.encodeSubject(message.Subject)))

	// 其他头部
	content.WriteString("MIME-Version: 1.0\r\n")
	content.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	content.WriteString("X-Mailer: EmailAlert v1.0\r\n")

	// 优先级
	if message.Priority > 0 {
		priorityMap := map[int]string{1: "High", 2: "Normal", 3: "Low"}
		if priority, ok := priorityMap[message.Priority]; ok {
			content.WriteString(fmt.Sprintf("X-Priority: %d\r\n", message.Priority))
			content.WriteString(fmt.Sprintf("X-MSMail-Priority: %s\r\n", priority))
		}
	}

	// 自定义头部
	for key, value := range message.Headers {
		content.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// 内容类型和边界
	boundary := fmt.Sprintf("boundary_%d", time.Now().Unix())

	if message.HTMLContent != "" && message.TextContent != "" {
		// 多部分邮件（文本+HTML）
		content.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		content.WriteString("\r\n")

		// 文本部分
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: base64\r\n")
		content.WriteString("\r\n")
		content.WriteString(base64.StdEncoding.EncodeToString([]byte(message.TextContent)))
		content.WriteString("\r\n\r\n")

		// HTML部分
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: base64\r\n")
		content.WriteString("\r\n")
		content.WriteString(base64.StdEncoding.EncodeToString([]byte(message.HTMLContent)))
		content.WriteString("\r\n\r\n")

		content.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if message.HTMLContent != "" {
		// 仅HTML
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: base64\r\n")
		content.WriteString("\r\n")
		content.WriteString(base64.StdEncoding.EncodeToString([]byte(message.HTMLContent)))
		content.WriteString("\r\n")
	} else {
		// 仅文本
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: base64\r\n")
		content.WriteString("\r\n")
		content.WriteString(base64.StdEncoding.EncodeToString([]byte(message.TextContent)))
		content.WriteString("\r\n")
	}

	return content.String()
}

// formatEmailAddress 格式化邮件地址
func (e *EmailNotifier) formatEmailAddress(email, name string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("\"%s\" <%s>", name, email)
}

// encodeSubject 编码邮件主题（支持中文）
func (e *EmailNotifier) encodeSubject(subject string) string {
	return mime.QEncoding.Encode("UTF-8", subject)
}

// TestConnection 测试邮件连接
func (e *EmailNotifier) TestConnection(config *EmailConfig) (*EmailResponse, error) {
	startTime := time.Now()

	// 先验证配置
	if err := e.ValidateConfig(config); err != nil {
		return &EmailResponse{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}, err
	}

	// 仅测试SMTP连接，不发送邮件
	err := e.testSMTPConnection(config)
	if err != nil {
		return &EmailResponse{
			Success:  false,
			Error:    fmt.Sprintf("SMTP连接测试失败: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	return &EmailResponse{
		Success:   true,
		Duration:  time.Since(startTime),
		MessageID: "connection-test-success",
	}, nil
}

// testSMTPConnection 仅测试SMTP连接（不发送邮件）
func (e *EmailNotifier) testSMTPConnection(config *EmailConfig) error {
	// 构建SMTP地址
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// 设置超时 - 增加到60秒
	timeout := 60 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	// 仅测试连接和TLS，不进行身份验证（避免状态污染）
	client, err := e.createSMTPClientForTest(config, addr, timeout)
	if err != nil {
		return err
	}
	defer client.Close()

	return nil
}

// createSMTPClient 创建SMTP客户端（统一的连接逻辑）
func (e *EmailNotifier) createSMTPClient(config *EmailConfig, addr string, timeout time.Duration) (*smtp.Client, error) {
	if config.Port == 465 {
		// 465端口 - 直接使用SSL/TLS连接 (SMTPS)
		return e.createDirectTLSClient(config, addr, timeout)
	} else {
		// 其他端口 - 使用普通连接，可选STARTTLS
		return e.createSTARTTLSClient(config, addr, timeout)
	}
}

// createDirectTLSClient 创建直接TLS连接的SMTP客户端（端口465）
func (e *EmailNotifier) createDirectTLSClient(config *EmailConfig, addr string, timeout time.Duration) (*smtp.Client, error) {
	// 创建TLS配置，增加兼容性设置
	tlsConfig := &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12, // 最低TLS 1.2
		MaxVersion:         0,                // 允许所有版本
	}

	// 直接建立TLS连接
	tlsConn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, tlsConfig)
	if err != nil {
		// 如果TLS连接失败，尝试降级到TLS 1.0
		tlsConfig.MinVersion = tls.VersionTLS10
		tlsConn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("建立TLS连接失败 (端口%d): %v", config.Port, err)
		}
	}

	// 创建SMTP客户端
	client, err := smtp.NewClient(tlsConn, config.Host)
	if err != nil {
		tlsConn.Close()
		return nil, fmt.Errorf("创建SMTP客户端失败: %v", err)
	}

	// 身份验证
	if config.Username != "" && config.Password != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err = client.Auth(auth); err != nil {
			client.Close()
			return nil, fmt.Errorf("SMTP身份验证失败: %v", err)
		}
	}

	return client, nil
}

// createSTARTTLSClient 创建STARTTLS连接的SMTP客户端（端口25、587等）
func (e *EmailNotifier) createSTARTTLSClient(config *EmailConfig, addr string, timeout time.Duration) (*smtp.Client, error) {
	// 建立普通TCP连接
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("连接SMTP服务器失败: %v", err)
	}

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		conn.Close()
		// 对于163等邮箱，如果直接创建客户端失败，提供特殊建议
		if strings.Contains(config.Host, "163.com") || strings.Contains(config.Host, "126.com") {
			return nil, fmt.Errorf("创建SMTP客户端失败(163/126邮箱): %v - 建议使用端口994/465或检查邮箱SMTP服务设置", err)
		}
		return nil, fmt.Errorf("创建SMTP客户端失败: %v", err)
	}

	// 如果启用SSL或者端口是587，尝试STARTTLS升级
	if config.SSL || config.Port == 587 {
		// 检查服务器是否支持STARTTLS
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName:         config.Host,
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12, // 最低TLS 1.2
				MaxVersion:         0,                // 允许所有版本
			}
			if err = client.StartTLS(tlsConfig); err != nil {
				// 如果STARTTLS失败，尝试降级到TLS 1.0
				tlsConfig.MinVersion = tls.VersionTLS10
				if err = client.StartTLS(tlsConfig); err != nil {
					client.Close()
					return nil, fmt.Errorf("STARTTLS升级失败: %v", err)
				}
			}
		} else if config.SSL {
			client.Close()
			return nil, fmt.Errorf("服务器不支持STARTTLS，但配置要求SSL")
		}
	}

	// 身份验证
	if config.Username != "" && config.Password != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err = client.Auth(auth); err != nil {
			client.Close()
			return nil, fmt.Errorf("SMTP身份验证失败: %v", err)
		}
	}

	return client, nil
}

// createSMTPClientForTest 创建用于测试的SMTP客户端（不进行身份验证）
func (e *EmailNotifier) createSMTPClientForTest(config *EmailConfig, addr string, timeout time.Duration) (*smtp.Client, error) {
	if config.Port == 465 {
		// 465端口 - 直接使用SSL/TLS连接 (SMTPS)
		return e.createDirectTLSClientForTest(config, addr, timeout)
	} else {
		// 其他端口 - 使用普通连接，可选STARTTLS
		return e.createSTARTTLSClientForTest(config, addr, timeout)
	}
}

// createDirectTLSClientForTest 创建直接TLS连接的SMTP客户端（测试用，无身份验证）
func (e *EmailNotifier) createDirectTLSClientForTest(config *EmailConfig, addr string, timeout time.Duration) (*smtp.Client, error) {
	// 创建TLS配置，增加兼容性设置
	tlsConfig := &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12, // 最低TLS 1.2
		MaxVersion:         0,                // 允许所有版本
	}

	// 直接建立TLS连接
	tlsConn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, tlsConfig)
	if err != nil {
		// 如果TLS连接失败，尝试降级到TLS 1.0
		tlsConfig.MinVersion = tls.VersionTLS10
		tlsConn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("建立TLS连接失败 (端口%d): %v", config.Port, err)
		}
	}

	// 创建SMTP客户端
	client, err := smtp.NewClient(tlsConn, config.Host)
	if err != nil {
		tlsConn.Close()
		return nil, fmt.Errorf("创建SMTP客户端失败: %v", err)
	}

	// 测试用，不进行身份验证
	return client, nil
}

// createSTARTTLSClientForTest 创建STARTTLS连接的SMTP客户端（测试用，无身份验证）
func (e *EmailNotifier) createSTARTTLSClientForTest(config *EmailConfig, addr string, timeout time.Duration) (*smtp.Client, error) {
	// 建立普通TCP连接
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("连接SMTP服务器失败: %v", err)
	}

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		conn.Close()
		// 对于163等邮箱，如果直接创建客户端失败，提供特殊建议
		if strings.Contains(config.Host, "163.com") || strings.Contains(config.Host, "126.com") {
			return nil, fmt.Errorf("创建SMTP客户端失败(163/126邮箱): %v - 建议使用端口994/465或检查邮箱SMTP服务设置", err)
		}
		return nil, fmt.Errorf("创建SMTP客户端失败: %v", err)
	}

	// 如果启用SSL或者端口是587，尝试STARTTLS升级
	if config.SSL || config.Port == 587 {
		// 检查服务器是否支持STARTTLS
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName:         config.Host,
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12, // 最低TLS 1.2
				MaxVersion:         0,                // 允许所有版本
			}
			if err = client.StartTLS(tlsConfig); err != nil {
				// 如果STARTTLS失败，尝试降级到TLS 1.0
				tlsConfig.MinVersion = tls.VersionTLS10
				if err = client.StartTLS(tlsConfig); err != nil {
					client.Close()
					return nil, fmt.Errorf("STARTTLS升级失败: %v", err)
				}
			}
		} else if config.SSL {
			client.Close()
			return nil, fmt.Errorf("服务器不支持STARTTLS，但配置要求SSL")
		}
	}

	// 测试用，不进行身份验证
	return client, nil
}

// ValidateConfig 验证邮件配置
func (e *EmailNotifier) ValidateConfig(config *EmailConfig) error {
	if config.Host == "" {
		return fmt.Errorf("SMTP服务器地址不能为空")
	}

	if config.Port == 0 {
		return fmt.Errorf("SMTP端口不能为空")
	}

	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("SMTP端口必须在1-65535之间")
	}

	if config.From == "" {
		return fmt.Errorf("发件人地址不能为空")
	}

	if !e.isValidEmail(config.From) {
		debugInfo := e.debugEmailValidation(config.From)
		return fmt.Errorf("发件人地址格式错误: '%s' - %s", config.From, debugInfo)
	}

	if len(config.To) == 0 {
		return fmt.Errorf("收件人不能为空")
	}

	// 验证所有邮件地址格式
	allEmails := make([]string, 0)
	allEmails = append(allEmails, config.To...)
	allEmails = append(allEmails, config.CC...)
	allEmails = append(allEmails, config.BCC...)

	if config.ReplyTo != "" {
		allEmails = append(allEmails, config.ReplyTo)
	}

	for _, email := range allEmails {
		if !e.isValidEmail(email) {
			return fmt.Errorf("邮件地址格式错误: %s", email)
		}
	}

	// 验证优先级
	if config.Priority != 0 && (config.Priority < 1 || config.Priority > 3) {
		return fmt.Errorf("邮件优先级必须在1-3之间")
	}

	// 验证格式
	if config.Format != "" {
		validFormats := []string{"text", "html", "mixed"}
		found := false
		for _, format := range validFormats {
			if strings.EqualFold(config.Format, format) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("不支持的邮件格式: %s", config.Format)
		}
	}

	return nil
}

// isValidEmail 邮件地址验证 - 采用更宽松的验证策略
func (e *EmailNotifier) isValidEmail(email string) bool {
	// 去除前后空格
	email = strings.TrimSpace(email)

	// 基本长度检查
	if len(email) < 3 || len(email) > 320 {
		return false
	}

	// 必须包含@符号
	if !strings.Contains(email, "@") {
		return false
	}

	// 分割邮箱地址
	parts := strings.Split(email, "@")
	if len(parts) < 2 {
		return false
	}

	// 如果有多个@，取最后一个作为分割点（支持特殊格式）
	if len(parts) > 2 {
		local := strings.Join(parts[:len(parts)-1], "@")
		domain := parts[len(parts)-1]
		parts = []string{local, domain}
	}

	local := parts[0]
	domain := parts[1]

	// 检查本地部分不能为空
	if len(local) == 0 {
		return false
	}

	// 检查域名部分不能为空且必须包含点
	if len(domain) == 0 || !strings.Contains(domain, ".") {
		return false
	}

	// 域名不能以点开头或结尾
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	// 域名不能包含连续的点
	if strings.Contains(domain, "..") {
		return false
	}

	return true
}

// debugEmailValidation 调试邮件地址验证（用于开发调试）
func (e *EmailNotifier) debugEmailValidation(email string) string {
	original := email
	email = strings.TrimSpace(email)

	if len(email) < 3 || len(email) > 320 {
		return fmt.Sprintf("长度不符合要求: 原始='%s', 处理后='%s', 长度=%d", original, email, len(email))
	}

	if !strings.Contains(email, "@") {
		return fmt.Sprintf("不包含@符号: '%s'", email)
	}

	parts := strings.Split(email, "@")
	if len(parts) < 2 {
		return fmt.Sprintf("@分割后部分不足: '%s', 部分数=%d", email, len(parts))
	}

	if len(parts) > 2 {
		local := strings.Join(parts[:len(parts)-1], "@")
		domain := parts[len(parts)-1]
		return fmt.Sprintf("多个@符号处理: 本地部分='%s', 域名='%s'", local, domain)
	}

	local := parts[0]
	domain := parts[1]

	if len(local) == 0 {
		return fmt.Sprintf("本地部分为空: 本地='%s', 域名='%s'", local, domain)
	}

	if len(domain) == 0 || !strings.Contains(domain, ".") {
		return fmt.Sprintf("域名问题: 域名='%s', 包含点=%v", domain, strings.Contains(domain, "."))
	}

	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Sprintf("域名以点开头或结尾: '%s'", domain)
	}

	if strings.Contains(domain, "..") {
		return fmt.Sprintf("域名包含连续点: '%s'", domain)
	}

	return fmt.Sprintf("验证通过: 本地='%s', 域名='%s'", local, domain)
}

// GetSupportedFormats 获取支持的邮件格式
func (e *EmailNotifier) GetSupportedFormats() []string {
	return []string{"text", "html", "mixed"}
}

// GetCommonSMTPConfigs 获取常见SMTP配置
func (e *EmailNotifier) GetCommonSMTPConfigs() map[string]EmailConfig {
	return map[string]EmailConfig{
		"gmail": {
			Host: "smtp.gmail.com",
			Port: 587,
			SSL:  true,
		},
		"outlook": {
			Host: "smtp-mail.outlook.com",
			Port: 587,
			SSL:  true,
		},
		"qq": {
			Host: "smtp.qq.com",
			Port: 587,
			SSL:  true,
		},
		"163": {
			Host: "smtp.163.com",
			Port: 994,
			SSL:  true,
		},
		"126": {
			Host: "smtp.126.com",
			Port: 994,
			SSL:  true,
		},
	}
}
