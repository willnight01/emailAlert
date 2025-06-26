package notification

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebhookNotifier 自定义Webhook通知器
type WebhookNotifier struct {
	client *http.Client
}

// NewWebhookNotifier 创建Webhook通知器
func NewWebhookNotifier() *WebhookNotifier {
	return &WebhookNotifier{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false, // 默认验证SSL证书
				},
			},
		},
	}
}

// WebhookConfig Webhook配置结构
type WebhookConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers,omitempty"`
	ContentType string            `json:"content_type"`
	AuthType    string            `json:"auth_type,omitempty"` // none, basic, bearer, apikey
	Username    string            `json:"username,omitempty"`  // Basic Auth用户名
	Password    string            `json:"password,omitempty"`  // Basic Auth密码
	Token       string            `json:"token,omitempty"`     // Bearer Token或API Key
	Timeout     int               `json:"timeout,omitempty"`   // 超时时间（秒）
	Retries     int               `json:"retries,omitempty"`   // 重试次数
	SkipSSL     bool              `json:"skip_ssl,omitempty"`  // 是否跳过SSL验证
	Template    string            `json:"template,omitempty"`  // 消息模板
}

// WebhookMessage Webhook消息结构
type WebhookMessage struct {
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level,omitempty"`
	Source    string                 `json:"source,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// WebhookResponse Webhook响应结构
type WebhookResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	Success    bool                `json:"success"`
	Error      string              `json:"error,omitempty"`
	Duration   time.Duration       `json:"duration"`
}

// SendMessage 发送Webhook消息
func (w *WebhookNotifier) SendMessage(config *WebhookConfig, title, content string) (*WebhookResponse, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("Webhook URL不能为空")
	}

	// 设置默认值
	if config.Method == "" {
		config.Method = "POST"
	}
	if config.ContentType == "" {
		config.ContentType = "application/json"
	}
	if config.Timeout > 0 {
		w.client.Timeout = time.Duration(config.Timeout) * time.Second
	}
	if config.Retries == 0 {
		config.Retries = 1
	}

	// 设置SSL验证
	if transport, ok := w.client.Transport.(*http.Transport); ok {
		transport.TLSClientConfig.InsecureSkipVerify = config.SkipSSL
	}

	// 构建消息
	message := &WebhookMessage{
		Title:     title,
		Content:   content,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     "info",
		Source:    "EmailAlert",
		Tags:      []string{"email", "alert"},
	}

	var response *WebhookResponse
	var lastErr error

	// 重试机制
	for i := 0; i < config.Retries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i*2) * time.Second) // 指数退避
		}

		response, lastErr = w.sendRequest(config, message)
		if lastErr == nil && response.Success {
			break
		}
	}

	if lastErr != nil {
		return response, fmt.Errorf("发送Webhook消息失败: %v", lastErr)
	}

	return response, nil
}

// sendRequest 发送HTTP请求
func (w *WebhookNotifier) sendRequest(config *WebhookConfig, message *WebhookMessage) (*WebhookResponse, error) {
	startTime := time.Now()

	// 构建请求体
	body, err := w.buildRequestBody(config, message)
	if err != nil {
		return nil, fmt.Errorf("构建请求体失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest(config.Method, config.URL, body)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", config.ContentType)

	// 设置自定义头部
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 设置认证
	if err := w.setAuthentication(req, config); err != nil {
		return nil, fmt.Errorf("设置认证失败: %v", err)
	}

	// 发送请求
	resp, err := w.client.Do(req)
	if err != nil {
		return &WebhookResponse{
			Success:    false,
			Error:      err.Error(),
			Duration:   time.Since(startTime),
			StatusCode: 0,
		}, err
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &WebhookResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Success:    false,
			Error:      fmt.Sprintf("读取响应体失败: %v", err),
			Duration:   time.Since(startTime),
		}, err
	}

	// 判断请求是否成功
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	response := &WebhookResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(respBody),
		Success:    success,
		Duration:   time.Since(startTime),
	}

	if !success {
		response.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody))
		return response, fmt.Errorf("Webhook请求失败: HTTP %d", resp.StatusCode)
	}

	return response, nil
}

// buildRequestBody 构建请求体
func (w *WebhookNotifier) buildRequestBody(config *WebhookConfig, message *WebhookMessage) (io.Reader, error) {
	switch strings.ToLower(config.ContentType) {
	case "application/json":
		// 如果有自定义模板，使用模板渲染
		if config.Template != "" {
			return w.buildFromTemplate(config.Template, message)
		}
		// 默认JSON格式
		jsonData, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(jsonData), nil

	case "application/x-www-form-urlencoded":
		// URL编码格式
		data := url.Values{}
		data.Set("title", message.Title)
		data.Set("content", message.Content)
		data.Set("timestamp", message.Timestamp)
		data.Set("level", message.Level)
		data.Set("source", message.Source)
		return strings.NewReader(data.Encode()), nil

	case "text/plain":
		// 纯文本格式
		text := fmt.Sprintf("标题: %s\n内容: %s\n时间: %s",
			message.Title, message.Content, message.Timestamp)
		return strings.NewReader(text), nil

	default:
		// 默认JSON格式
		jsonData, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(jsonData), nil
	}
}

// buildFromTemplate 从模板构建请求体
func (w *WebhookNotifier) buildFromTemplate(template string, message *WebhookMessage) (io.Reader, error) {
	// 简单的变量替换
	content := template
	content = strings.ReplaceAll(content, "{{title}}", message.Title)
	content = strings.ReplaceAll(content, "{{content}}", message.Content)
	content = strings.ReplaceAll(content, "{{timestamp}}", message.Timestamp)
	content = strings.ReplaceAll(content, "{{level}}", message.Level)
	content = strings.ReplaceAll(content, "{{source}}", message.Source)

	return strings.NewReader(content), nil
}

// setAuthentication 设置认证
func (w *WebhookNotifier) setAuthentication(req *http.Request, config *WebhookConfig) error {
	switch strings.ToLower(config.AuthType) {
	case "basic":
		if config.Username == "" || config.Password == "" {
			return fmt.Errorf("Basic认证需要用户名和密码")
		}
		req.SetBasicAuth(config.Username, config.Password)

	case "bearer":
		if config.Token == "" {
			return fmt.Errorf("Bearer认证需要Token")
		}
		req.Header.Set("Authorization", "Bearer "+config.Token)

	case "apikey":
		if config.Token == "" {
			return fmt.Errorf("API Key认证需要Token")
		}
		req.Header.Set("X-API-Key", config.Token)

	case "none", "":
		// 无认证
		break

	default:
		return fmt.Errorf("不支持的认证类型: %s", config.AuthType)
	}

	return nil
}

// TestConnection 测试Webhook连接
func (w *WebhookNotifier) TestConnection(config *WebhookConfig) (*WebhookResponse, error) {
	testTitle := "Webhook连接测试"
	testContent := "这是一条来自邮件告警平台的连接测试消息\n" +
		"发送时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
		"如果您收到此消息，说明Webhook配置正确！"

	return w.SendMessage(config, testTitle, testContent)
}

// ValidateConfig 验证Webhook配置
func (w *WebhookNotifier) ValidateConfig(config *WebhookConfig) error {
	if config.URL == "" {
		return fmt.Errorf("URL不能为空")
	}

	// 验证URL格式
	if _, err := url.Parse(config.URL); err != nil {
		return fmt.Errorf("URL格式错误: %v", err)
	}

	// 验证HTTP方法
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	if config.Method != "" {
		method := strings.ToUpper(config.Method)
		found := false
		for _, vm := range validMethods {
			if method == vm {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("不支持的HTTP方法: %s", config.Method)
		}
	}

	// 验证Content-Type
	validContentTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"text/plain",
		"text/xml",
		"application/xml",
	}
	if config.ContentType != "" {
		contentType := strings.ToLower(config.ContentType)
		found := false
		for _, vct := range validContentTypes {
			if strings.HasPrefix(contentType, vct) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("不支持的Content-Type: %s", config.ContentType)
		}
	}

	// 验证认证配置
	if config.AuthType != "" {
		switch strings.ToLower(config.AuthType) {
		case "basic":
			if config.Username == "" || config.Password == "" {
				return fmt.Errorf("Basic认证需要用户名和密码")
			}
		case "bearer", "apikey":
			if config.Token == "" {
				return fmt.Errorf("%s认证需要Token", config.AuthType)
			}
		case "none":
			// 无认证，合法
		default:
			return fmt.Errorf("不支持的认证类型: %s", config.AuthType)
		}
	}

	// 验证超时和重试配置
	if config.Timeout < 0 || config.Timeout > 300 {
		return fmt.Errorf("超时时间必须在0-300秒之间")
	}
	if config.Retries < 0 || config.Retries > 10 {
		return fmt.Errorf("重试次数必须在0-10次之间")
	}

	return nil
}

// GetSupportedContentTypes 获取支持的Content-Type列表
func (w *WebhookNotifier) GetSupportedContentTypes() []string {
	return []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"text/plain",
		"text/xml",
		"application/xml",
	}
}

// GetSupportedAuthTypes 获取支持的认证类型列表
func (w *WebhookNotifier) GetSupportedAuthTypes() []string {
	return []string{
		"none",
		"basic",
		"bearer",
		"apikey",
	}
}
