package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// EmailData 邮件数据结构
type EmailData struct {
	UID         int              `json:"uid"`
	Subject     string           `json:"subject"`
	Sender      string           `json:"sender"`
	Content     string           `json:"content"`
	HTMLContent string           `json:"html_content"`
	ReceivedAt  time.Time        `json:"received_at"`
	Size        uint64           `json:"size"`
	Flags       []string         `json:"flags"`
	MessageID   string           `json:"message_id"`
	Attachments []AttachmentData `json:"attachments"`
}

// AttachmentData 附件数据结构
type AttachmentData struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	Content  []byte `json:"content"`
}

// MailboxConfig 邮箱配置结构
type MailboxConfig struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Protocol string `json:"protocol"`
	SSL      bool   `json:"ssl"`
	Status   string `json:"status"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	CheckInterval time.Duration // 检查间隔
	MaxRetries    int           // 最大重试次数
	Folder        string        // 监控的文件夹，默认为INBOX
	MarkAsRead    bool          // 是否标记为已读
	OnlyUnread    bool          // 是否只处理未读邮件
}

// DefaultMonitorConfig 默认监控配置
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		CheckInterval: 30 * time.Second,
		MaxRetries:    3,
		Folder:        "INBOX",
		MarkAsRead:    false,
		OnlyUnread:    false, // 临时改为false，监控所有邮件
	}
}

// EmailHandler 邮件处理器接口
type EmailHandler interface {
	HandleEmail(mailboxID uint, email *EmailData) error
}

// Monitor 邮件监控器
type Monitor struct {
	mailboxes    []MailboxConfig
	config       *MonitorConfig
	handler      EmailHandler
	stopChan     chan struct{}
	wg           sync.WaitGroup
	isRunning    bool
	mutex        sync.RWMutex
	lastCheckUID map[uint]int // 记录每个邮箱的最后检查UID
	startTime    time.Time    // 监控启动时间，只处理此时间之后的邮件
	parser       *EmailParser // 邮件解析器
}

// NewMonitor 创建新的邮件监控器
func NewMonitor(config *MonitorConfig, handler EmailHandler) *Monitor {
	if config == nil {
		config = DefaultMonitorConfig()
	}

	return &Monitor{
		config:       config,
		handler:      handler,
		stopChan:     make(chan struct{}),
		lastCheckUID: make(map[uint]int),
		parser:       NewEmailParser(), // 初始化邮件解析器
	}
}

// AddMailbox 添加邮箱到监控列表
func (m *Monitor) AddMailbox(mailbox MailboxConfig) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已存在
	for i, mb := range m.mailboxes {
		if mb.ID == mailbox.ID {
			m.mailboxes[i] = mailbox // 更新配置
			return
		}
	}

	m.mailboxes = append(m.mailboxes, mailbox)
	log.Printf("邮箱监控: 添加邮箱 %s (%s) 到监控列表", mailbox.Name, mailbox.Email)
}

// RemoveMailbox 从监控列表移除邮箱
func (m *Monitor) RemoveMailbox(mailboxID uint) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, mb := range m.mailboxes {
		if mb.ID == mailboxID {
			m.mailboxes = append(m.mailboxes[:i], m.mailboxes[i+1:]...)
			delete(m.lastCheckUID, mailboxID)
			log.Printf("邮箱监控: 从监控列表移除邮箱 ID %d", mailboxID)
			return
		}
	}
}

// UpdateMailboxes 更新邮箱列表
func (m *Monitor) UpdateMailboxes(mailboxes []MailboxConfig) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mailboxes = mailboxes
	// 清理不存在的邮箱的lastCheckUID
	existingIDs := make(map[uint]bool)
	for _, mb := range mailboxes {
		existingIDs[mb.ID] = true
	}

	for id := range m.lastCheckUID {
		if !existingIDs[id] {
			delete(m.lastCheckUID, id)
		}
	}

	log.Printf("邮箱监控: 更新邮箱列表，当前监控 %d 个邮箱", len(mailboxes))
}

// Start 开始监控
func (m *Monitor) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("邮箱监控已在运行中")
	}

	m.isRunning = true
	m.stopChan = make(chan struct{})

	// 使用北京时间作为启动时间
	beijingTZ, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		beijingTZ = time.FixedZone("CST", 8*3600) // 使用固定时区作为后备
	}
	m.startTime = time.Now().In(beijingTZ) // 记录监控启动时间（北京时间）

	log.Printf("邮箱监控: 开始监控，检查间隔 %v，启动时间 %v", m.config.CheckInterval, m.startTime.Format("2006-01-02 15:04:05"))

	// 为每个邮箱启动一个goroutine
	for _, mailbox := range m.mailboxes {
		if mailbox.Status == "active" {
			m.wg.Add(1)
			go m.monitorMailbox(mailbox)
		}
	}

	return nil
}

// Stop 停止监控
func (m *Monitor) Stop() {
	m.mutex.Lock()
	if !m.isRunning {
		m.mutex.Unlock()
		return
	}

	m.isRunning = false
	close(m.stopChan)
	m.mutex.Unlock()

	log.Printf("邮箱监控: 正在停止监控...")

	// 使用带超时的等待，避免无限阻塞
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("邮箱监控: 监控已停止")
	case <-time.After(10 * time.Second):
		log.Printf("邮箱监控: 停止超时，强制结束")
	}
}

// IsRunning 检查是否正在运行
func (m *Monitor) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isRunning
}

// GetStatus 获取监控状态
func (m *Monitor) GetStatus() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	status := map[string]interface{}{
		"is_running":     m.isRunning,
		"mailbox_count":  len(m.mailboxes),
		"check_interval": m.config.CheckInterval.String(),
		"last_check_uid": m.lastCheckUID,
	}

	// 如果监控正在运行，添加启动时间信息
	if m.isRunning && !m.startTime.IsZero() {
		status["start_time"] = m.startTime.Format("2006-01-02 15:04:05")
		status["running_duration"] = time.Since(m.startTime).String()
	}

	return status
}

// monitorMailbox 监控单个邮箱
func (m *Monitor) monitorMailbox(mailboxConfig MailboxConfig) {
	defer m.wg.Done()

	log.Printf("邮箱监控: 开始监控邮箱 %s (%s)", mailboxConfig.Name, mailboxConfig.Email)

	// 在开始监控前验证邮箱访问权限
	if err := m.validateMailboxAccess(mailboxConfig); err != nil {
		log.Printf("邮箱监控: 邮箱 %s 访问验证失败，跳过监控: %v", mailboxConfig.Name, err)
		return
	}

	log.Printf("邮箱监控: 邮箱 %s 访问验证成功，开始监控", mailboxConfig.Name)

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	// 立即执行一次检查
	m.checkMailbox(mailboxConfig)

	for {
		select {
		case <-ticker.C:
			m.checkMailbox(mailboxConfig)
		case <-m.stopChan:
			log.Printf("邮箱监控: 停止监控邮箱 %s", mailboxConfig.Name)
			return
		}
	}
}

// checkMailbox 检查邮箱是否有新邮件
func (m *Monitor) checkMailbox(mailboxConfig MailboxConfig) {
	retries := 0
	maxRetries := m.config.MaxRetries

	for retries <= maxRetries {
		err := m.fetchNewEmails(mailboxConfig)
		if err == nil {
			break
		}

		retries++
		if retries <= maxRetries {
			log.Printf("邮箱监控: 检查邮箱 %s 失败，重试 %d/%d: %v", mailboxConfig.Name, retries, maxRetries, err)
			time.Sleep(time.Duration(retries) * time.Second)
		} else {
			log.Printf("邮箱监控: 检查邮箱 %s 失败，已达最大重试次数: %v", mailboxConfig.Name, err)
		}
	}
}

// validateMailboxAccess 验证邮箱是否支持完整访问
func (m *Monitor) validateMailboxAccess(mailboxConfig MailboxConfig) error {
	addr := fmt.Sprintf("%s:%d", mailboxConfig.Host, mailboxConfig.Port)

	var conn *client.Client
	var err error

	if mailboxConfig.SSL {
		tlsConfig := &tls.Config{
			ServerName: mailboxConfig.Host,
		}
		conn, err = client.DialTLS(addr, tlsConfig)
	} else {
		conn, err = client.Dial(addr)
	}

	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}
	defer conn.Logout()

	// 登录
	if err := conn.Login(mailboxConfig.Username, mailboxConfig.Password); err != nil {
		return fmt.Errorf("登录失败: %v", err)
	}

	// 针对126/163邮箱发送ID命令
	if strings.Contains(mailboxConfig.Host, "126.com") || strings.Contains(mailboxConfig.Host, "163.com") {
		err = sendIMAPIDInMonitor(conn)
		if err != nil {
			log.Printf("发送IMAP ID失败: %v", err)
		} else {
			log.Printf("邮箱监控: 检测到126/163邮箱，已发送ID命令")
		}
	}

	// 尝试访问INBOX文件夹
	_, err = conn.Select("INBOX", true)
	if err != nil {
		if strings.Contains(err.Error(), "Unsafe Login") ||
			strings.Contains(err.Error(), "authorization") ||
			strings.Contains(err.Error(), "auth") {
			return fmt.Errorf("邮箱安全限制，无法访问文件夹。请检查邮箱设置：\n1. 确保已开启IMAP服务\n2. 如果是126/163等邮箱，请使用授权码而非密码\n3. 检查邮箱安全设置\n错误详情: %v", err)
		}
		return fmt.Errorf("无法访问邮箱文件夹: %v", err)
	}

	return nil
}

// fetchNewEmails 获取新邮件
func (m *Monitor) fetchNewEmails(mailboxConfig MailboxConfig) error {
	// 创建IMAP连接
	addr := fmt.Sprintf("%s:%d", mailboxConfig.Host, mailboxConfig.Port)

	var conn *client.Client
	var err error

	if mailboxConfig.SSL {
		tlsConfig := &tls.Config{
			ServerName: mailboxConfig.Host,
		}
		conn, err = client.DialTLS(addr, tlsConfig)
	} else {
		conn, err = client.Dial(addr)
	}

	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}
	defer conn.Logout()

	// 登录
	if err := conn.Login(mailboxConfig.Username, mailboxConfig.Password); err != nil {
		return fmt.Errorf("登录失败: %v", err)
	}

	// 针对126/163/阿里云企业邮箱发送ID命令
	if strings.Contains(mailboxConfig.Host, "126.com") ||
		strings.Contains(mailboxConfig.Host, "163.com") ||
		strings.Contains(mailboxConfig.Host, "qiye.aliyun.com") {
		err = sendIMAPIDInMonitor(conn)
		if err != nil {
			log.Printf("发送IMAP ID失败: %v", err)
		} else {
			log.Printf("邮箱监控: 检测到企业邮箱，已发送ID命令")
		}
	}

	// 选择邮箱文件夹
	mbox, err := conn.Select(m.config.Folder, true)
	if err != nil {
		// 检查是否是安全限制错误
		if strings.Contains(err.Error(), "Unsafe Login") ||
			strings.Contains(err.Error(), "authorization") ||
			strings.Contains(err.Error(), "auth") {
			return fmt.Errorf("邮箱安全限制: %s。请检查邮箱设置，确保已开启IMAP服务并使用正确的授权码", err.Error())
		}
		return fmt.Errorf("选择文件夹 %s 失败: %v", m.config.Folder, err)
	}

	// 如果邮箱为空，返回
	if mbox.Messages == 0 {
		return nil
	}

	// 获取上次检查的UID
	lastUID := m.getLastCheckUID(mailboxConfig.ID)

	// 修复阿里云企业邮箱兼容性问题 - 完全避免使用搜索
	var uids []uint32
	var searchErr error

	// 针对阿里云企业邮箱：直接获取邮件而不使用搜索
	if strings.Contains(mailboxConfig.Host, "qiye.aliyun.com") {
		// 阿里云企业邮箱：直接使用UID范围，避免搜索命令
		if lastUID > 0 {
			// 从上次检查的UID+1开始到最新的UID
			if mbox.UidNext > uint32(lastUID+1) {
				seqset := &imap.SeqSet{}
				seqset.AddRange(uint32(lastUID+1), mbox.UidNext-1)

				// 直接获取这个范围的邮件UID和时间信息
				messages := make(chan *imap.Message, 100)
				done := make(chan error, 1)
				go func() {
					done <- conn.UidFetch(seqset, []imap.FetchItem{imap.FetchUid, imap.FetchEnvelope}, messages)
				}()

				for msg := range messages {
					// 只处理监控启动时间之后的邮件
					if msg.Envelope != nil && msg.Envelope.Date.After(m.startTime) {
						uids = append(uids, msg.Uid)
					}
				}

				if fetchErr := <-done; fetchErr != nil {
					return fmt.Errorf("获取邮件UID失败: %v", fetchErr)
				}
			}
		} else {
			// 首次检查：只获取监控启动时间之后的邮件
			if mbox.UidNext > 1 {
				seqset := &imap.SeqSet{}
				seqset.AddRange(1, mbox.UidNext-1)

				// 获取所有邮件的UID和时间信息，然后过滤
				messages := make(chan *imap.Message, 100)
				done := make(chan error, 1)
				go func() {
					done <- conn.UidFetch(seqset, []imap.FetchItem{imap.FetchUid, imap.FetchEnvelope}, messages)
				}()

				for msg := range messages {
					// 只处理监控启动时间之后的邮件
					if msg.Envelope != nil && msg.Envelope.Date.After(m.startTime) {
						uids = append(uids, msg.Uid)
					}
				}

				if fetchErr := <-done; fetchErr != nil {
					return fmt.Errorf("获取邮件UID失败: %v", fetchErr)
				}
			}
		}
		log.Printf("阿里云企业邮箱: 获取到 %d 个符合时间条件的UID（启动时间之后）", len(uids))
	} else {
		// 其他邮箱：使用搜索逻辑，添加时间过滤
		var criteria *imap.SearchCriteria
		if lastUID > 0 {
			// 只获取UID大于lastUID且时间在启动时间之后的邮件
			criteria = &imap.SearchCriteria{
				Uid:   &imap.SeqSet{},
				Since: m.startTime, // 添加时间过滤
			}
			criteria.Uid.AddRange(uint32(lastUID+1), mbox.UidNext-1)
		} else {
			// 首次检查：只获取监控启动时间之后的邮件
			criteria = &imap.SearchCriteria{
				Since: m.startTime, // 只获取启动时间之后的邮件
			}
		}

		// 如果只处理未读邮件
		if m.config.OnlyUnread {
			criteria.WithoutFlags = []string{imap.SeenFlag}
		}

		// 搜索邮件
		uids, searchErr = conn.UidSearch(criteria)
		if searchErr != nil {
			return fmt.Errorf("搜索邮件失败: %v", searchErr)
		}
		log.Printf("邮箱 %s: 搜索到 %d 个符合时间条件的UID（启动时间之后）", mailboxConfig.Name, len(uids))
	}

	if len(uids) == 0 {
		return nil
	}

	log.Printf("邮箱监控: 在邮箱 %s 中找到 %d 封新邮件", mailboxConfig.Name, len(uids))

	// 获取邮件详情
	seqset := &imap.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	// 获取邮件头和正文
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822Size, imap.FetchRFC822}
	messages := make(chan *imap.Message, len(uids))

	done := make(chan error, 1)
	go func() {
		done <- conn.UidFetch(seqset, items, messages)
	}()

	// 处理每封邮件
	var maxUID uint32
	for msg := range messages {
		if msg.Uid > maxUID {
			maxUID = msg.Uid
		}

		emailData := m.convertToEmailData(msg)
		if emailData != nil {
			// 调用处理器处理邮件
			if m.handler != nil {
				if err := m.handler.HandleEmail(mailboxConfig.ID, emailData); err != nil {
					log.Printf("邮箱监控: 处理邮件失败: %v", err)
				}
			}
		}
	}

	if err := <-done; err != nil {
		return fmt.Errorf("获取邮件失败: %v", err)
	}

	// 更新最后检查的UID
	if maxUID > 0 {
		m.setLastCheckUID(mailboxConfig.ID, int(maxUID))
	}

	return nil
}

// convertToEmailData 转换邮件数据
func (m *Monitor) convertToEmailData(msg *imap.Message) *EmailData {
	if msg.Envelope == nil {
		return nil
	}

	// 使用北京时间进行时间验证
	beijingTZ, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		beijingTZ = time.FixedZone("CST", 8*3600) // 使用固定时区作为后备
	}

	// 添加调试日志，查看邮件原始时间和转换过程
	log.Printf("邮件原始时间调试: Date=%v, Location=%v", msg.Envelope.Date, msg.Envelope.Date.Location())

	// 二次时间验证：确保邮件时间在监控启动时间之后（都转换为北京时间）
	emailTimeBeijing := msg.Envelope.Date.In(beijingTZ)
	log.Printf("邮件时间转换调试: 原始时间=%s, 北京时间=%s",
		msg.Envelope.Date.Format("2006-01-02 15:04:05 -0700 MST"),
		emailTimeBeijing.Format("2006-01-02 15:04:05"))

	if !msg.Envelope.Date.IsZero() {
		startTimeBeijing := m.startTime.In(beijingTZ)

		if !emailTimeBeijing.After(startTimeBeijing) {
			log.Printf("邮件时间验证失败(北京时间): 邮件时间=%s, 监控启动时间=%s, 跳过处理",
				emailTimeBeijing.Format("2006-01-02 15:04:05"),
				startTimeBeijing.Format("2006-01-02 15:04:05"))
			return nil
		}
	}

	emailData := &EmailData{
		UID:        int(msg.Uid),
		Subject:    msg.Envelope.Subject,
		Size:       uint64(msg.Size),
		Flags:      msg.Flags,
		MessageID:  msg.Envelope.MessageId,
		ReceivedAt: emailTimeBeijing, // 统一使用北京时间
	}

	// 处理发件人
	if len(msg.Envelope.From) > 0 {
		emailData.Sender = msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
	}

	// 提取邮件正文 - 使用专业的邮件解析器
	rfc822Section := &imap.BodySectionName{}
	if rfc822Body, exists := msg.Body[rfc822Section]; exists && rfc822Body != nil {
		// 读取完整的RFC822格式邮件
		buf := make([]byte, 0, 8192)
		tempBuf := make([]byte, 4096)

		for {
			n, err := rfc822Body.Read(tempBuf)
			if n > 0 {
				buf = append(buf, tempBuf[:n]...)
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("读取RFC822邮件正文失败: %v", err)
				}
				break
			}
		}

		if len(buf) > 0 {
			// 使用专业的邮件解析器解析邮件内容
			if m.parser != nil {
				textContent, htmlContent, err := m.parser.ParseContent(string(buf))
				if err != nil {
					log.Printf("邮件解析失败: %v", err)
					// 如果解析失败，使用原始方法作为后备
					emailData.Content = m.extractContentFromRFC822(string(buf))
				} else {
					emailData.Content = textContent
					emailData.HTMLContent = htmlContent

					// 如果没有纯文本内容，但有HTML内容，从HTML中提取文本
					if emailData.Content == "" && emailData.HTMLContent != "" {
						emailData.Content = m.parser.stripHTMLTags(emailData.HTMLContent)
					}
				}
			} else {
				// 后备方法：使用原始的解析方式
				emailData.Content = m.extractContentFromRFC822(string(buf))
			}
		}
	} else if len(msg.Body) > 0 {
		// 如果没有RFC822格式，尝试从Body字段读取
		for _, body := range msg.Body {
			if body != nil {
				buf := make([]byte, 0, 8192)
				tempBuf := make([]byte, 4096)

				for {
					n, err := body.Read(tempBuf)
					if n > 0 {
						buf = append(buf, tempBuf[:n]...)
					}
					if err != nil {
						if err != io.EOF {
							log.Printf("读取邮件正文失败: %v", err)
						}
						break
					}
				}

				if len(buf) > 0 {
					// 使用专业的邮件解析器解析邮件内容
					if m.parser != nil {
						textContent, htmlContent, err := m.parser.ParseContent(string(buf))
						if err != nil {
							log.Printf("邮件解析失败: %v", err)
							emailData.Content = string(buf)
						} else {
							emailData.Content = textContent
							emailData.HTMLContent = htmlContent

							// 如果没有纯文本内容，但有HTML内容，从HTML中提取文本
							if emailData.Content == "" && emailData.HTMLContent != "" {
								emailData.Content = m.parser.stripHTMLTags(emailData.HTMLContent)
							}
						}
					} else {
						emailData.Content = string(buf)
					}
					break
				}
			}
		}
	}

	// 如果没有读取到正文内容，尝试从信封中获取一些基本信息作为内容
	if emailData.Content == "" && msg.Envelope != nil {
		contentParts := []string{}
		if msg.Envelope.Subject != "" {
			contentParts = append(contentParts, "主题: "+msg.Envelope.Subject)
		}
		if len(msg.Envelope.From) > 0 {
			from := msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
			contentParts = append(contentParts, "发件人: "+from)
		}
		if !msg.Envelope.Date.IsZero() {
			contentParts = append(contentParts, "时间: "+msg.Envelope.Date.Format("2006-01-02 15:04:05"))
		}

		if len(contentParts) > 0 {
			emailData.Content = strings.Join(contentParts, "\n") + "\n\n[注: 未能获取邮件正文内容，显示基本信息]"
		} else {
			emailData.Content = "[未能获取邮件内容]"
		}
	}

	return emailData
}

// getLastCheckUID 获取最后检查的UID
func (m *Monitor) getLastCheckUID(mailboxID uint) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if uid, exists := m.lastCheckUID[mailboxID]; exists {
		return uid
	}
	return 0
}

// setLastCheckUID 设置最后检查的UID
func (m *Monitor) setLastCheckUID(mailboxID uint, uid int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.lastCheckUID[mailboxID] = uid
}

// sendIMAPIDInMonitor 发送IMAP ID命令，用于解决126/163邮箱的"Unsafe Login"问题
func sendIMAPIDInMonitor(conn *client.Client) error {
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

// extractContentFromRFC822 从RFC822格式的邮件中提取正文内容
func (m *Monitor) extractContentFromRFC822(rawEmail string) string {
	// 简单的RFC822解析，查找正文部分
	lines := strings.Split(rawEmail, "\n")

	// 找到空行，表示头部结束，正文开始
	var bodyStartIndex int
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			bodyStartIndex = i + 1
			break
		}
	}

	if bodyStartIndex >= len(lines) {
		return ""
	}

	// 提取正文部分
	bodyLines := lines[bodyStartIndex:]
	bodyContent := strings.Join(bodyLines, "\n")

	// 处理多部分邮件
	if strings.Contains(strings.ToLower(rawEmail), "content-type: multipart/") {
		// 查找text/plain或text/html部分
		parts := strings.Split(bodyContent, "--")
		for _, part := range parts {
			partLower := strings.ToLower(part)
			if strings.Contains(partLower, "content-type: text/plain") ||
				strings.Contains(partLower, "content-type: text/html") {

				// 找到这个部分的正文
				partLines := strings.Split(part, "\n")
				var partBodyStart int
				var encoding string

				// 解析头部，查找编码方式
				for i, line := range partLines {
					lineLower := strings.ToLower(strings.TrimSpace(line))
					if lineLower == "" {
						partBodyStart = i + 1
						break
					}
					if strings.HasPrefix(lineLower, "content-transfer-encoding:") {
						encoding = strings.TrimSpace(strings.Split(line, ":")[1])
					}
				}

				if partBodyStart < len(partLines) {
					partBody := strings.Join(partLines[partBodyStart:], "\n")
					partBody = strings.TrimSpace(partBody)

					if partBody != "" {
						// 根据编码方式解码内容
						decodedContent := m.decodeContent(partBody, encoding)
						if decodedContent != "" {
							return decodedContent
						}
					}
				}
			}
		}
	}

	// 如果不是多部分邮件，直接返回正文
	return strings.TrimSpace(bodyContent)
}

// decodeContent 根据编码方式解码内容
func (m *Monitor) decodeContent(content, encoding string) string {
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	switch encoding {
	case "base64":
		// Base64解码
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			log.Printf("Base64解码失败: %v", err)
			return content // 返回原始内容
		}
		return string(decoded)

	case "quoted-printable":
		// Quoted-Printable解码（简化实现）
		return m.decodeQuotedPrintable(content)

	default:
		// 无编码或不支持的编码，直接返回
		return content
	}
}

// decodeQuotedPrintable 简单的Quoted-Printable解码
func (m *Monitor) decodeQuotedPrintable(content string) string {
	// 简化的Quoted-Printable解码
	content = strings.ReplaceAll(content, "=\r\n", "")
	content = strings.ReplaceAll(content, "=\n", "")

	// 解码=XX格式的字符
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		decoded := ""
		i := 0
		for i < len(line) {
			if i < len(line)-2 && line[i] == '=' {
				// 尝试解码=XX格式
				if hex := line[i+1 : i+3]; len(hex) == 2 {
					if b, err := strconv.ParseInt(hex, 16, 8); err == nil {
						decoded += string(byte(b))
						i += 3
						continue
					}
				}
			}
			decoded += string(line[i])
			i++
		}
		result = append(result, decoded)
	}

	return strings.Join(result, "\n")
}
