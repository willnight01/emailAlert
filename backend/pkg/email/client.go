package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// Client IMAP客户端
type Client struct {
	host     string
	port     int
	username string
	password string
	ssl      bool
}

// NewClient 创建新的邮箱客户端
func NewClient(host string, port int, username, password string, ssl bool) *Client {
	return &Client{
		host:     host,
		port:     port,
		username: username,
		password: password,
		ssl:      ssl,
	}
}

// connect 创建IMAP连接
func (c *Client) connect() (*client.Client, error) {
	// 构建连接地址
	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	var conn *client.Client
	var err error

	if c.ssl {
		// SSL/TLS连接
		tlsConfig := &tls.Config{
			ServerName: c.host,
		}
		conn, err = client.DialTLS(addr, tlsConfig)
	} else {
		// 普通连接
		conn, err = client.Dial(addr)
	}

	if err != nil {
		return nil, fmt.Errorf("连接服务器失败: %v", err)
	}

	// 登录认证
	err = conn.Login(c.username, c.password)
	if err != nil {
		conn.Logout()
		return nil, fmt.Errorf("认证失败: %v", err)
	}

	// 针对126/163邮箱发送ID命令
	if c.shouldSendID() {
		err = c.sendIMAPID(conn)
		if err != nil {
			// ID发送失败不影响连接，只记录日志
			fmt.Printf("发送IMAP ID失败: %v\n", err)
		}
	}

	return conn, nil
}

// TestConnection 测试邮箱连接（带超时控制）
func (c *Client) TestConnection() error {
	// 设置15秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 创建一个channel来接收结果
	resultChan := make(chan error, 1)

	go func() {
		conn, err := c.connect()
		if err != nil {
			resultChan <- err
			return
		}
		defer conn.Logout()

		// 连接和认证成功就算测试成功
		// 不强制要求文件夹访问成功（某些邮箱服务商有安全限制）
		resultChan <- nil
	}()

	// 等待结果或超时
	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("连接超时，请检查邮箱配置和网络连接")
	}
}

// GetFolders 获取邮箱文件夹列表
func (c *Client) GetFolders() ([]string, error) {
	// 设置15秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 创建一个channel来接收结果
	type result struct {
		folders []string
		err     error
	}
	resultChan := make(chan result, 1)

	go func() {
		conn, err := c.connect()
		if err != nil {
			resultChan <- result{nil, err}
			return
		}
		defer conn.Logout()

		// 获取文件夹列表
		mailboxes := make(chan *imap.MailboxInfo, 10)
		done := make(chan error, 1)
		go func() {
			done <- conn.List("", "*", mailboxes)
		}()

		var folders []string
		for m := range mailboxes {
			folders = append(folders, m.Name)
		}

		if err := <-done; err != nil {
			resultChan <- result{nil, fmt.Errorf("获取文件夹列表失败: %v", err)}
			return
		}

		// 如果没有获取到文件夹，添加默认文件夹
		if len(folders) == 0 {
			folders = append(folders, "INBOX")
		}

		resultChan <- result{folders, nil}
	}()

	// 等待结果或超时
	select {
	case res := <-resultChan:
		return res.folders, res.err
	case <-ctx.Done():
		return nil, fmt.Errorf("获取文件夹列表超时")
	}
}

// GetEmailCount 获取指定文件夹的邮件数量
func (c *Client) GetEmailCount(folder string) (int, error) {
	conn, err := c.connect()
	if err != nil {
		return 0, err
	}
	defer conn.Logout()

	// 选择文件夹
	mbox, err := conn.Select(folder, true)
	if err != nil {
		return 0, fmt.Errorf("选择文件夹失败: %v", err)
	}

	return int(mbox.Messages), nil
}

// ConnectionInfo 连接信息结构
type ConnectionInfo struct {
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	Username string    `json:"username"`
	SSL      bool      `json:"ssl"`
	Status   string    `json:"status"`    // connected, disconnected, error
	Message  string    `json:"message"`   // 连接状态消息
	Folders  []string  `json:"folders"`   // 文件夹列表
	TestTime time.Time `json:"test_time"` // 测试时间
}

// GetConnectionInfo 获取完整的连接信息
func (c *Client) GetConnectionInfo() *ConnectionInfo {
	info := &ConnectionInfo{
		Host:     c.host,
		Port:     c.port,
		Username: c.username,
		SSL:      c.ssl,
		TestTime: time.Now(),
	}

	// 测试连接
	err := c.TestConnection()
	if err != nil {
		info.Status = "error"
		info.Message = err.Error()
		return info
	}

	// 连接成功
	info.Status = "connected"
	info.Message = "连接成功"
	info.Folders = []string{"INBOX"} // 提供默认文件夹

	// 注意：暂时不获取文件夹列表，因为某些邮箱服务商有安全限制
	// 可以在后续版本中添加可选的文件夹获取功能

	return info
}

// shouldSendID 判断是否需要发送IMAP ID命令
func (c *Client) shouldSendID() bool {
	// 针对126/163/阿里云企业邮箱需要发送ID命令
	return strings.Contains(c.host, "126.com") ||
		strings.Contains(c.host, "163.com") ||
		strings.Contains(c.host, "qiye.aliyun.com")
}

// sendIMAPID 发送IMAP ID命令，用于解决126/163邮箱的"Unsafe Login"问题
func (c *Client) sendIMAPID(conn *client.Client) error {
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
