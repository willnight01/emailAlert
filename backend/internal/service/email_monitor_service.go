package service

import (
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"emailAlert/pkg/email"
	"fmt"
	"log"
	"sync"
	"time"
)

// LogEntry 日志条目结构
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // info, success, warning, error
	Message   string    `json:"message"`
	MailboxID uint      `json:"mailbox_id,omitempty"`
}

// EmailMonitorService 邮件监控服务
type EmailMonitorService struct {
	mailboxRepo            *repository.MailboxRepository
	alertRepo              *repository.AlertRepository
	enhancedRuleEngine     EnhancedRuleEngineService
	notificationDispatcher NotificationDispatcherService
	monitor                *email.Monitor
	logChannel             chan LogEntry
	logClients             map[chan LogEntry]bool
	logMutex               sync.RWMutex
}

// NewEmailMonitorService 创建邮件监控服务实例
func NewEmailMonitorService(
	mailboxRepo *repository.MailboxRepository,
	alertRepo *repository.AlertRepository,
	enhancedRuleEngine EnhancedRuleEngineService,
	notificationDispatcher NotificationDispatcherService,
) *EmailMonitorService {
	service := &EmailMonitorService{
		mailboxRepo:            mailboxRepo,
		alertRepo:              alertRepo,
		enhancedRuleEngine:     enhancedRuleEngine,
		notificationDispatcher: notificationDispatcher,
		logChannel:             make(chan LogEntry, 100),
		logClients:             make(map[chan LogEntry]bool),
	}

	// 创建邮件监控器，传入处理器
	config := email.DefaultMonitorConfig()
	service.monitor = email.NewMonitor(config, service)

	// 启动日志分发协程
	go service.startLogDispatcher()

	return service
}

// startLogDispatcher 启动日志分发器
func (s *EmailMonitorService) startLogDispatcher() {
	for logEntry := range s.logChannel {
		s.logMutex.RLock()
		for client := range s.logClients {
			select {
			case client <- logEntry:
			default:
				// 如果客户端无法接收，说明连接已断开，删除客户端
				delete(s.logClients, client)
				close(client)
			}
		}
		s.logMutex.RUnlock()
	}
}

// AddLogClient 添加日志客户端
func (s *EmailMonitorService) AddLogClient() chan LogEntry {
	client := make(chan LogEntry, 10)
	s.logMutex.Lock()
	s.logClients[client] = true
	s.logMutex.Unlock()
	return client
}

// RemoveLogClient 移除日志客户端
func (s *EmailMonitorService) RemoveLogClient(client chan LogEntry) {
	s.logMutex.Lock()
	delete(s.logClients, client)
	close(client)
	s.logMutex.Unlock()
}

// addLog 添加日志条目
func (s *EmailMonitorService) addLog(level, message string, mailboxID ...uint) {
	var mbID uint
	if len(mailboxID) > 0 {
		mbID = mailboxID[0]
	}

	logEntry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		MailboxID: mbID,
	}

	select {
	case s.logChannel <- logEntry:
	default:
		// 日志通道满了，丢弃旧日志
		log.Printf("日志通道满，丢弃日志: %s", message)
	}
}

// HandleEmail 实现EmailHandler接口，处理收到的邮件
func (s *EmailMonitorService) HandleEmail(mailboxID uint, emailData *email.EmailData) error {
	s.addLog("info", fmt.Sprintf("收到邮件: 主题=%s, 发件人=%s", emailData.Subject, emailData.Sender), mailboxID)

	// 定义北京时区
	beijingTZ, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		s.addLog("error", fmt.Sprintf("加载北京时区失败: %v", err), mailboxID)
		beijingTZ = time.FixedZone("CST", 8*3600) // 使用固定时区作为后备
	}

	// 验证邮件接收时间是否在监控启动时间之后
	status := s.monitor.GetStatus()
	if startTimeStr, ok := status["start_time"].(string); ok {
		// 使用ParseInLocation确保解析时保持北京时区，而不是转换为UTC
		if startTime, err := time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, beijingTZ); err == nil {
			// startTime已经是北京时间了，不需要再次转换

			s.addLog("debug", fmt.Sprintf("北京时间比较: 邮件时间=%s, 启动时间=%s",
				emailData.ReceivedAt.Format("2006-01-02 15:04:05"),
				startTime.Format("2006-01-02 15:04:05")), mailboxID)

			// 验证邮件接收时间是否在监控启动时间之后
			if !emailData.ReceivedAt.IsZero() && !emailData.ReceivedAt.After(startTime) {
				s.addLog("warning", fmt.Sprintf("邮件接收时间(%s)早于或等于监控启动时间(%s)，跳过处理: %s",
					emailData.ReceivedAt.Format("2006-01-02 15:04:05"),
					startTime.Format("2006-01-02 15:04:05"),
					emailData.Subject), mailboxID)
				return nil
			} else {
				s.addLog("info", fmt.Sprintf("邮件时间验证通过: 邮件时间=%s > 监控启动时间=%s",
					emailData.ReceivedAt.Format("2006-01-02 15:04:05"),
					startTime.Format("2006-01-02 15:04:05")), mailboxID)
			}
		} else {
			s.addLog("error", fmt.Sprintf("解析监控启动时间失败: %v", err), mailboxID)
		}
	} else {
		s.addLog("warning", "无法获取监控启动时间，跳过时间验证", mailboxID)
	}

	// 增强的去重检查逻辑
	var isDuplicate bool
	var duplicateCheckErr error

	if emailData.MessageID != "" {
		// 优先使用MessageID进行去重
		isDuplicate, duplicateCheckErr = s.alertRepo.ExistsByMessageID(emailData.MessageID)
		if duplicateCheckErr != nil {
			s.addLog("error", fmt.Sprintf("检查邮件MessageID是否存在失败: %v", duplicateCheckErr), mailboxID)
		}
	} else {
		// MessageID为空时，使用组合字段进行去重检查
		s.addLog("warning", fmt.Sprintf("邮件MessageID为空，使用组合字段进行去重检查: %s", emailData.Subject), mailboxID)

		// 使用 UID + 邮箱ID + 主题 + 发件人 + 接收时间 作为组合唯一标识
		isDuplicate, duplicateCheckErr = s.alertRepo.ExistsByCompositeKey(
			mailboxID,
			emailData.UID,
			emailData.Subject,
			emailData.Sender,
			emailData.ReceivedAt,
		)
		if duplicateCheckErr != nil {
			s.addLog("error", fmt.Sprintf("检查邮件组合字段是否存在失败: %v", duplicateCheckErr), mailboxID)
		}
	}

	if isDuplicate {
		s.addLog("warning", fmt.Sprintf("邮件已存在，跳过处理: %s", emailData.Subject), mailboxID)
		return nil
	}

	// 转换邮件数据格式
	modelEmailData := &model.EmailData{
		UID:         emailData.UID,
		Subject:     emailData.Subject,
		Sender:      emailData.Sender,
		Content:     emailData.Content,
		HTMLContent: emailData.HTMLContent,
		ReceivedAt:  emailData.ReceivedAt,
		MessageID:   emailData.MessageID,
		Size:        emailData.Size,
		Flags:       emailData.Flags,
		// 暂时留空，后续可根据需要扩展
		To: []string{},
		CC: []string{},
		AttachmentNames: func() []string {
			names := make([]string, len(emailData.Attachments))
			for i, att := range emailData.Attachments {
				names[i] = att.Name
			}
			return names
		}(),
	}

	// 使用增强版规则引擎处理邮件
	results, err := s.enhancedRuleEngine.ProcessEmailWithRuleGroups(modelEmailData, mailboxID)
	if err != nil {
		s.addLog("error", fmt.Sprintf("规则引擎处理失败: %v", err), mailboxID)
		return fmt.Errorf("规则引擎处理失败: %v", err)
	}

	// 统计处理结果
	totalMatched := 0
	alertsCreated := 0
	duplicatesSkipped := 0
	errors := 0

	for _, result := range results {
		if result.RuleGroup != nil {
			totalMatched++
		}
		if result.Created {
			alertsCreated++
			// 分发告警通知
			if err := s.notificationDispatcher.DispatchAlert(result.Alert); err != nil {
				s.addLog("warning", fmt.Sprintf("分发告警通知失败: %v", err), mailboxID)
			}
		}
		if result.IsDuplicate {
			duplicatesSkipped++
		}
		if result.Error != "" {
			errors++
			s.addLog("error", fmt.Sprintf("告警处理错误: %s", result.Error), mailboxID)
		}
	}

	if totalMatched == 0 {
		s.addLog("info", fmt.Sprintf("邮件未匹配任何规则: %s", emailData.Subject), mailboxID)
	} else {
		s.addLog("success", fmt.Sprintf("邮件处理完成 - 匹配规则组: %d, 创建告警: %d, 跳过重复: %d, 错误: %d",
			totalMatched, alertsCreated, duplicatesSkipped, errors), mailboxID)
	}

	return nil
}

// Start 启动邮件监控
func (s *EmailMonitorService) Start() error {
	s.addLog("info", "正在启动邮件监控服务...")

	// 获取活跃的邮箱配置
	mailboxes, err := s.mailboxRepo.GetActiveMailboxes()
	if err != nil {
		s.addLog("error", fmt.Sprintf("获取活跃邮箱配置失败: %v", err))
		return fmt.Errorf("获取活跃邮箱配置失败: %v", err)
	}

	s.addLog("info", fmt.Sprintf("找到 %d 个活跃邮箱配置", len(mailboxes)))

	// 转换为监控器需要的格式
	var monitorConfigs []email.MailboxConfig
	for _, mb := range mailboxes {
		// 直接使用明文密码
		config := email.MailboxConfig{
			ID:       mb.ID,
			Name:     mb.Name,
			Email:    mb.Email,
			Host:     mb.Host,
			Port:     mb.Port,
			Username: mb.Username,
			Password: mb.Password,
			Protocol: mb.Protocol,
			SSL:      mb.SSL,
			Status:   mb.Status,
		}
		monitorConfigs = append(monitorConfigs, config)
		s.addLog("info", fmt.Sprintf("准备监控邮箱: %s (%s)", mb.Name, mb.Email), mb.ID)
	}

	// 更新监控器的邮箱列表
	s.monitor.UpdateMailboxes(monitorConfigs)

	// 启动监控
	err = s.monitor.Start()
	if err != nil {
		s.addLog("error", fmt.Sprintf("启动监控失败: %v", err))
		return err
	}

	s.addLog("success", "邮件监控服务已启动")
	return nil
}

// Stop 停止邮件监控
func (s *EmailMonitorService) Stop() error {
	s.addLog("info", "正在停止邮件监控服务...")
	s.monitor.Stop()
	s.addLog("success", "邮件监控服务已停止")
	return nil
}

// IsRunning 检查是否正在运行
func (s *EmailMonitorService) IsRunning() bool {
	return s.monitor.IsRunning()
}

// GetStatus 获取监控状态
func (s *EmailMonitorService) GetStatus() map[string]interface{} {
	return s.monitor.GetStatus()
}

// RefreshMailboxes 刷新邮箱配置（当邮箱配置发生变化时调用）
func (s *EmailMonitorService) RefreshMailboxes() error {
	s.addLog("info", "正在刷新邮箱配置...")

	if !s.monitor.IsRunning() {
		s.addLog("warning", "监控服务未运行，无法刷新配置")
		return nil // 如果监控未运行，不需要刷新
	}

	// 获取最新的活跃邮箱配置
	mailboxes, err := s.mailboxRepo.GetActiveMailboxes()
	if err != nil {
		s.addLog("error", fmt.Sprintf("获取活跃邮箱配置失败: %v", err))
		return fmt.Errorf("获取活跃邮箱配置失败: %v", err)
	}

	s.addLog("info", fmt.Sprintf("发现 %d 个活跃邮箱配置", len(mailboxes)))

	// 转换为监控器需要的格式
	var monitorConfigs []email.MailboxConfig
	for _, mb := range mailboxes {
		// 直接使用明文密码
		config := email.MailboxConfig{
			ID:       mb.ID,
			Name:     mb.Name,
			Email:    mb.Email,
			Host:     mb.Host,
			Port:     mb.Port,
			Username: mb.Username,
			Password: mb.Password,
			Protocol: mb.Protocol,
			SSL:      mb.SSL,
			Status:   mb.Status,
		}
		monitorConfigs = append(monitorConfigs, config)
	}

	// 更新监控器的邮箱列表
	s.monitor.UpdateMailboxes(monitorConfigs)

	s.addLog("success", fmt.Sprintf("邮箱配置已刷新，当前监控 %d 个邮箱", len(monitorConfigs)))
	return nil
}

// AddMailbox 添加邮箱到监控
func (s *EmailMonitorService) AddMailbox(mailbox model.Mailbox) error {
	if !s.monitor.IsRunning() {
		return nil // 如果监控未运行，不需要添加
	}

	// 直接使用明文密码
	config := email.MailboxConfig{
		ID:       mailbox.ID,
		Name:     mailbox.Name,
		Email:    mailbox.Email,
		Host:     mailbox.Host,
		Port:     mailbox.Port,
		Username: mailbox.Username,
		Password: mailbox.Password,
		Protocol: mailbox.Protocol,
		SSL:      mailbox.SSL,
		Status:   mailbox.Status,
	}

	s.monitor.AddMailbox(config)
	return nil
}

// RemoveMailbox 从监控中移除邮箱
func (s *EmailMonitorService) RemoveMailbox(mailboxID uint) {
	s.monitor.RemoveMailbox(mailboxID)
}

// UpdateMonitorConfig 更新监控配置
func (s *EmailMonitorService) UpdateMonitorConfig(config *email.MonitorConfig) error {
	// 如果监控正在运行，需要重启以应用新配置
	wasRunning := s.monitor.IsRunning()
	if wasRunning {
		s.monitor.Stop()
	}

	// 创建新的监控器
	s.monitor = email.NewMonitor(config, s)

	if wasRunning {
		return s.Start()
	}

	return nil
}

// GetEmailStats 获取邮件统计信息
func (s *EmailMonitorService) GetEmailStats() (map[string]interface{}, error) {
	// 获取活跃邮箱数量
	activeMailboxes, err := s.mailboxRepo.GetActiveMailboxes()
	if err != nil {
		return nil, fmt.Errorf("获取活跃邮箱数失败: %v", err)
	}

	// 获取今日邮件统计
	todayStats, err := s.alertRepo.GetTodayStats()
	if err != nil {
		return nil, fmt.Errorf("获取今日邮件统计失败: %v", err)
	}

	// 获取待处理告警数
	pendingAlerts, err := s.alertRepo.GetPendingAlerts(0) // 0表示获取所有
	if err != nil {
		return nil, fmt.Errorf("获取待处理告警数失败: %v", err)
	}

	// 计算错误次数（失败状态的告警数）
	var errorCount int64 = 0
	if failedCount, exists := todayStats["failed_count"]; exists {
		if count, ok := failedCount.(int64); ok {
			errorCount = count
		}
	}

	// 获取监控状态
	monitorStatus := s.monitor.GetStatus()
	isRunning, _ := monitorStatus["is_running"].(bool)

	stats := map[string]interface{}{
		"activeMailboxes": len(activeMailboxes),
		"todayEmails":     todayStats["total_count"],
		"errorCount":      errorCount,
		"pendingAlerts":   len(pendingAlerts),
		"isRunning":       isRunning,
		"mailboxCount":    monitorStatus["mailbox_count"],
		"checkInterval":   monitorStatus["check_interval"],
		"lastCheckUID":    monitorStatus["last_check_uid"],
	}

	return stats, nil
}
