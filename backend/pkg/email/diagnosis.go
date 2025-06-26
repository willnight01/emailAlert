package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// DiagnosisResult 诊断结果
type DiagnosisResult struct {
	Step       string `json:"step"`                 // 诊断步骤
	Success    bool   `json:"success"`              // 是否成功
	Message    string `json:"message"`              // 消息
	Suggestion string `json:"suggestion,omitempty"` // 建议
}

// EmailDiagnosis 邮箱诊断器
type EmailDiagnosis struct {
	Results []DiagnosisResult `json:"results"`
}

// sendIMAPID 发送IMAP ID命令，用于解决126/163邮箱的"Unsafe Login"问题
func sendIMAPID(conn *client.Client) error {
	// 构造ID命令参数
	// 根据CSDN文章的解决方案，发送客户端标识信息
	idArgs := []interface{}{
		"name", "EmailAlert",
		"version", "1.0.0",
		"vendor", "EmailAlert System",
	}

	// 创建ID命令
	cmd := &imap.Command{
		Name:      "ID",
		Arguments: []interface{}{idArgs},
	}

	// 发送命令
	status, err := conn.Execute(cmd, nil)
	if err != nil {
		return fmt.Errorf("执行ID命令失败: %v", err)
	}

	if status.Type != imap.StatusRespOk {
		return fmt.Errorf("ID命令执行状态异常: %v", status.Type)
	}

	log.Printf("IMAP ID命令发送成功")
	return nil
}

// DiagnoseMailbox 诊断邮箱连接问题
func DiagnoseMailbox(mailboxConfig MailboxConfig) *EmailDiagnosis {
	diagnosis := &EmailDiagnosis{}

	// 步骤1: 基础连接测试
	diagnosis.testConnection(mailboxConfig)

	// 步骤2: 认证测试
	diagnosis.testAuthentication(mailboxConfig)

	// 步骤3: IMAP权限测试
	diagnosis.testIMAPAccess(mailboxConfig)

	// 步骤4: 特定邮箱提供商建议
	diagnosis.provideProviderSuggestions(mailboxConfig)

	return diagnosis
}

// testConnection 测试基础连接
func (d *EmailDiagnosis) testConnection(config MailboxConfig) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	var conn *client.Client
	var err error

	if config.SSL {
		tlsConfig := &tls.Config{
			ServerName: config.Host,
		}
		conn, err = client.DialTLS(addr, tlsConfig)
	} else {
		conn, err = client.Dial(addr)
	}

	if err != nil {
		d.Results = append(d.Results, DiagnosisResult{
			Step:       "连接测试",
			Success:    false,
			Message:    fmt.Sprintf("无法连接到服务器 %s", addr),
			Suggestion: "请检查服务器地址、端口号和网络连接",
		})
		return
	}
	defer conn.Logout()

	d.Results = append(d.Results, DiagnosisResult{
		Step:    "连接测试",
		Success: true,
		Message: fmt.Sprintf("成功连接到服务器 %s", addr),
	})
}

// testAuthentication 测试认证
func (d *EmailDiagnosis) testAuthentication(config MailboxConfig) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	var conn *client.Client
	var err error

	if config.SSL {
		tlsConfig := &tls.Config{
			ServerName: config.Host,
		}
		conn, err = client.DialTLS(addr, tlsConfig)
	} else {
		conn, err = client.Dial(addr)
	}

	if err != nil {
		d.Results = append(d.Results, DiagnosisResult{
			Step:    "认证测试",
			Success: false,
			Message: "连接失败，跳过认证测试",
		})
		return
	}
	defer conn.Logout()

	// 尝试登录
	err = conn.Login(config.Username, config.Password)
	if err != nil {
		suggestion := "请检查用户名和密码"
		if strings.Contains(err.Error(), "authentication") ||
			strings.Contains(err.Error(), "auth") {
			suggestion = "认证失败。如果是126/163/Gmail等邮箱，可能需要使用授权码而非密码"
		}

		d.Results = append(d.Results, DiagnosisResult{
			Step:       "认证测试",
			Success:    false,
			Message:    fmt.Sprintf("认证失败: %v", err),
			Suggestion: suggestion,
		})
		return
	}

	d.Results = append(d.Results, DiagnosisResult{
		Step:    "认证测试",
		Success: true,
		Message: "认证成功",
	})
}

// testIMAPAccess 测试IMAP访问权限
func (d *EmailDiagnosis) testIMAPAccess(config MailboxConfig) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	var conn *client.Client
	var err error

	if config.SSL {
		tlsConfig := &tls.Config{
			ServerName: config.Host,
		}
		conn, err = client.DialTLS(addr, tlsConfig)
	} else {
		conn, err = client.Dial(addr)
	}

	if err != nil {
		d.Results = append(d.Results, DiagnosisResult{
			Step:    "IMAP访问测试",
			Success: false,
			Message: "连接失败，跳过IMAP访问测试",
		})
		return
	}
	defer conn.Logout()

	// 先登录
	err = conn.Login(config.Username, config.Password)
	if err != nil {
		d.Results = append(d.Results, DiagnosisResult{
			Step:    "IMAP访问测试",
			Success: false,
			Message: "登录失败，跳过IMAP访问测试",
		})
		return
	}

	// 针对126/163/阿里云企业邮箱发送ID命令
	if strings.Contains(config.Host, "126.com") ||
		strings.Contains(config.Host, "163.com") ||
		strings.Contains(config.Host, "qiye.aliyun.com") {
		err = sendIMAPID(conn)
		if err != nil {
			log.Printf("发送IMAP ID失败: %v", err)
		}
	}

	// 尝试访问INBOX（使用只读模式）
	_, err = conn.Select("INBOX", true) // readonly=true
	if err != nil {
		suggestion := ""
		if strings.Contains(err.Error(), "Unsafe Login") {
			suggestion = "检测到'Unsafe Login'错误。请：\n1. 开启邮箱的IMAP服务\n2. 使用授权码替代密码\n3. 检查邮箱安全设置"
		} else if strings.Contains(err.Error(), "authorization") ||
			strings.Contains(err.Error(), "auth") {
			suggestion = "权限不足。请检查邮箱IMAP权限设置"
		} else {
			suggestion = "无法访问邮箱文件夹，请检查邮箱设置"
		}

		d.Results = append(d.Results, DiagnosisResult{
			Step:       "IMAP访问测试",
			Success:    false,
			Message:    fmt.Sprintf("无法访问INBOX文件夹: %v", err),
			Suggestion: suggestion,
		})
		return
	}

	d.Results = append(d.Results, DiagnosisResult{
		Step:    "IMAP访问测试",
		Success: true,
		Message: "成功访问INBOX文件夹，邮件监控应该可以正常工作",
	})
}

// provideProviderSuggestions 提供特定邮箱提供商的建议
func (d *EmailDiagnosis) provideProviderSuggestions(config MailboxConfig) {
	host := strings.ToLower(config.Host)

	suggestions := []string{}

	if strings.Contains(host, "126.com") || strings.Contains(host, "imap.126.com") {
		suggestions = append(suggestions, "126邮箱设置建议：")
		suggestions = append(suggestions, "1. 登录126邮箱网页版")
		suggestions = append(suggestions, "2. 进入设置 > POP3/SMTP/IMAP")
		suggestions = append(suggestions, "3. 开启IMAP/SMTP服务")
		suggestions = append(suggestions, "4. 使用授权码作为密码，而非登录密码")
		suggestions = append(suggestions, "5. 服务器地址：imap.126.com，端口：993，启用SSL")
	} else if strings.Contains(host, "163.com") || strings.Contains(host, "imap.163.com") {
		suggestions = append(suggestions, "163邮箱设置建议：")
		suggestions = append(suggestions, "1. 登录163邮箱网页版")
		suggestions = append(suggestions, "2. 进入设置 > POP3/SMTP/IMAP")
		suggestions = append(suggestions, "3. 开启IMAP/SMTP服务")
		suggestions = append(suggestions, "4. 使用授权码作为密码")
		suggestions = append(suggestions, "5. 服务器地址：imap.163.com，端口：993，启用SSL")
	} else if strings.Contains(host, "qq.com") || strings.Contains(host, "imap.qq.com") {
		suggestions = append(suggestions, "QQ邮箱设置建议：")
		suggestions = append(suggestions, "1. 登录QQ邮箱网页版")
		suggestions = append(suggestions, "2. 进入设置 > 账户")
		suggestions = append(suggestions, "3. 开启IMAP/SMTP服务")
		suggestions = append(suggestions, "4. 生成授权码作为密码")
		suggestions = append(suggestions, "5. 服务器地址：imap.qq.com，端口：993，启用SSL")
	} else if strings.Contains(host, "gmail.com") || strings.Contains(host, "imap.gmail.com") {
		suggestions = append(suggestions, "Gmail设置建议：")
		suggestions = append(suggestions, "1. 开启两步验证")
		suggestions = append(suggestions, "2. 生成应用专用密码")
		suggestions = append(suggestions, "3. 使用应用专用密码而非Google账号密码")
		suggestions = append(suggestions, "4. 服务器地址：imap.gmail.com，端口：993，启用SSL")
	}

	if len(suggestions) > 0 {
		d.Results = append(d.Results, DiagnosisResult{
			Step:       "邮箱提供商建议",
			Success:    true,
			Message:    "根据您的邮箱提供商，提供以下设置建议",
			Suggestion: strings.Join(suggestions, "\n"),
		})
	}
}

// PrintDiagnosis 打印诊断结果
func (d *EmailDiagnosis) PrintDiagnosis() {
	log.Println("=== 邮箱诊断结果 ===")
	for i, result := range d.Results {
		status := "✓"
		if !result.Success {
			status = "✗"
		}

		log.Printf("%d. %s %s: %s", i+1, status, result.Step, result.Message)
		if result.Suggestion != "" {
			log.Printf("   建议: %s", result.Suggestion)
		}
	}
	log.Println("=== 诊断完成 ===")
}
