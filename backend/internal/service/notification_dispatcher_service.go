package service

import (
	"context"
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// NotificationDispatcherService 通知分发服务接口
type NotificationDispatcherService interface {
	DispatchAlert(alert *model.Alert) error
	ProcessPendingAlerts() error
	RetryFailedNotifications() error
	StartBackgroundProcessor(ctx context.Context) error
	GetDispatchStats() (map[string]interface{}, error)
}

// notificationDispatcherService 通知分发服务实现
type notificationDispatcherService struct {
	ruleChannelRepo      repository.RuleChannelRepository      // 旧架构兼容
	ruleGroupChannelRepo repository.RuleGroupChannelRepository // 新架构
	notificationLogRepo  repository.NotificationLogRepository
	alertRepo            *repository.AlertRepository
	channelService       ChannelService
	templateService      *TemplateService

	// 配置参数
	maxRetryCount int
	retryInterval time.Duration
	maxWorkers    int
	batchSize     int

	// 工作队列
	alertQueue     chan *model.Alert
	retryQueue     chan *model.NotificationLog
	workerWg       sync.WaitGroup
	processingLock sync.RWMutex
}

// NotificationTask 通知任务
type NotificationTask struct {
	Alert   *model.Alert   `json:"alert"`
	Channel *model.Channel `json:"channel"`
	Content string         `json:"content"`
	Subject string         `json:"subject"`
	LogID   uint           `json:"log_id"`
}

// NotificationResult 通知结果
type NotificationResult struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	ResponseData string `json:"response_data,omitempty"`
}

// NewNotificationDispatcherService 创建通知分发服务
func NewNotificationDispatcherService(
	ruleChannelRepo repository.RuleChannelRepository,
	ruleGroupChannelRepo repository.RuleGroupChannelRepository,
	notificationLogRepo repository.NotificationLogRepository,
	alertRepo *repository.AlertRepository,
	channelService ChannelService,
	templateService *TemplateService,
) NotificationDispatcherService {
	return &notificationDispatcherService{
		ruleChannelRepo:      ruleChannelRepo,
		ruleGroupChannelRepo: ruleGroupChannelRepo,
		notificationLogRepo:  notificationLogRepo,
		alertRepo:            alertRepo,
		channelService:       channelService,
		templateService:      templateService,
		maxRetryCount:        3,
		retryInterval:        5 * time.Minute,
		maxWorkers:           10,
		batchSize:            50,
		alertQueue:           make(chan *model.Alert, 100),
		retryQueue:           make(chan *model.NotificationLog, 100),
	}
}

// DispatchAlert 分发告警通知
func (s *notificationDispatcherService) DispatchAlert(alert *model.Alert) error {
	// 非阻塞方式将告警添加到队列
	select {
	case s.alertQueue <- alert:
		log.Printf("告警 %d 已添加到分发队列", alert.ID)
		return nil
	default:
		// 队列满时直接处理
		return s.processAlert(alert)
	}
}

// processAlert 处理单个告警
func (s *notificationDispatcherService) processAlert(alert *model.Alert) error {
	var channels []*model.Channel
	var err error

	// 优先使用新的规则组架构
	if alert.RuleGroupID > 0 {
		channels, err = s.ruleGroupChannelRepo.GetChannelsByRuleGroupID(alert.RuleGroupID)
		if err != nil {
			return fmt.Errorf("获取规则组渠道失败: %v", err)
		}
		log.Printf("告警 %d 使用规则组 %d 获取到 %d 个通知渠道", alert.ID, alert.RuleGroupID, len(channels))
	} else if alert.RuleID > 0 {
		// 向后兼容旧的规则架构
		channels, err = s.ruleChannelRepo.GetChannelsByRuleID(alert.RuleID)
		if err != nil {
			return fmt.Errorf("获取规则渠道失败: %v", err)
		}
		log.Printf("告警 %d 使用旧规则 %d 获取到 %d 个通知渠道", alert.ID, alert.RuleID, len(channels))
	} else {
		log.Printf("告警 %d 没有匹配规则或规则组，跳过通知", alert.ID)
		return nil
	}

	if len(channels) == 0 {
		log.Printf("告警 %d 没有配置通知渠道", alert.ID)
		return nil
	}

	// 为每个渠道创建通知任务
	var wg sync.WaitGroup
	sentChannels := make([]string, 0, len(channels))
	var mu sync.Mutex

	for _, channel := range channels {
		wg.Add(1)
		go func(ch *model.Channel) {
			defer wg.Done()

			if err := s.sendNotificationToChannel(alert, ch); err != nil {
				log.Printf("发送通知到渠道 %s 失败: %v", ch.Name, err)
			} else {
				mu.Lock()
				sentChannels = append(sentChannels, ch.Name)
				mu.Unlock()
			}
		}(channel)
	}

	wg.Wait()

	// 更新告警状态
	status := "sent"
	if len(sentChannels) == 0 {
		status = "failed"
	}

	return s.alertRepo.UpdateStatusWithDetails(alert.ID, status, strings.Join(sentChannels, ","), "")
}

// sendNotificationToChannel 向指定渠道发送通知
func (s *notificationDispatcherService) sendNotificationToChannel(alert *model.Alert, channel *model.Channel) error {
	// 创建通知日志记录
	notificationLog := &model.NotificationLog{
		ChannelID: channel.ID,
		AlertID:   alert.ID,
		Status:    "pending",
		Content:   "",
	}

	if err := s.notificationLogRepo.Create(notificationLog); err != nil {
		return fmt.Errorf("创建通知日志失败: %v", err)
	}

	// 生成通知内容
	content, subject, err := s.generateNotificationContent(alert, channel)
	if err != nil {
		s.notificationLogRepo.UpdateStatus(notificationLog.ID, "failed",
			fmt.Sprintf("生成通知内容失败: %v", err), "")
		return err
	}

	// 更新日志内容到数据库
	if err := s.notificationLogRepo.UpdateContent(notificationLog.ID, content); err != nil {
		log.Printf("更新通知日志内容失败: %v", err)
	}
	notificationLog.Content = content

	// 发送通知
	result := s.sendNotification(channel, subject, content)

	// 更新通知状态
	status := "failed"
	if result.Success {
		status = "success"
	}

	err = s.notificationLogRepo.UpdateStatus(notificationLog.ID, status, result.Error, result.ResponseData)
	if err != nil {
		log.Printf("更新通知日志状态失败: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("通知发送失败: %s", result.Error)
	}

	return nil
}

// generateNotificationContent 生成通知内容
func (s *notificationDispatcherService) generateNotificationContent(alert *model.Alert, channel *model.Channel) (string, string, error) {
	// 获取渲染数据
	renderData := s.buildRenderData(alert)

	var template *model.Template
	var err error

	// 如果渠道指定了模版，优先使用渠道模版
	if channel.TemplateID != nil && *channel.TemplateID > 0 {
		template, err = s.templateService.GetByID(*channel.TemplateID)
		if err != nil {
			log.Printf("获取渠道指定模版失败，将使用默认模版: %v", err)
		}
	}

	// 如果没有渠道模版或获取失败，使用默认模版
	if template == nil {
		template, err = s.templateService.GetDefaultByType(channel.Type)
		if err != nil {
			log.Printf("获取默认模版失败，使用简单格式: %v", err)
			return s.generateSimpleContent(alert, channel.Type), alert.Subject, nil
		}
	}

	// 渲染模版
	result, err := s.templateService.Render(template.ID, renderData)
	if err != nil {
		log.Printf("渲染模版失败，使用简单格式: %v", err)
		return s.generateSimpleContent(alert, channel.Type), alert.Subject, nil
	}

	// 处理消息长度限制
	content := s.processMessageLength(result.Content, channel.Type)
	subject := result.Subject
	if subject == "" {
		subject = fmt.Sprintf("[告警] %s", alert.Subject)
	}

	return content, subject, nil
}

// buildRenderData 构建模版渲染数据
func (s *notificationDispatcherService) buildRenderData(alert *model.Alert) *model.TemplateRenderData {
	now := time.Now()

	// 构建邮件数据
	emailData := &model.EmailData{
		Subject:    alert.Subject,
		Sender:     alert.Sender,
		Content:    alert.Content,
		ReceivedAt: alert.ReceivedAt,
		MessageID:  "",
	}

	// 构建渲染数据
	renderData := &model.TemplateRenderData{
		Email: emailData,
		Alert: alert,
		System: model.SystemInfo{
			AppName:     "邮件告警平台",
			AppVersion:  "1.0.0",
			ServerName:  "localhost",
			Environment: "production",
		},
		Time: model.TimeInfo{
			Now:       now,
			NowFormat: now.Format("2006-01-02 15:04:05"),
			NowUnix:   now.Unix(),
			Today:     now.Format("2006-01-02"),
			Yesterday: now.AddDate(0, 0, -1).Format("2006-01-02"),
		},
	}

	// 如果有关联规则，添加规则信息
	if alert.RuleID > 0 && alert.Rule.ID > 0 {
		renderData.Rule = &alert.Rule
	}

	// 如果有关联邮箱，添加邮箱信息
	if alert.MailboxID > 0 && alert.Mailbox.ID > 0 {
		renderData.Mailbox = &alert.Mailbox
	}

	return renderData
}

// generateSimpleContent 生成简单格式的通知内容
func (s *notificationDispatcherService) generateSimpleContent(alert *model.Alert, channelType string) string {
	switch channelType {
	case "dingtalk":
		return fmt.Sprintf("## 邮件告警通知\n\n**主题：** %s\n**发件人：** %s\n**时间：** %s\n\n**内容：**\n%s",
			alert.Subject, alert.Sender, alert.ReceivedAt.Format("2006-01-02 15:04:05"), alert.Content)
	case "wechat":
		return fmt.Sprintf("邮件告警通知\n主题：%s\n发件人：%s\n时间：%s\n\n内容：\n%s",
			alert.Subject, alert.Sender, alert.ReceivedAt.Format("2006-01-02 15:04:05"), alert.Content)
	case "email":
		return fmt.Sprintf("<h2>邮件告警通知</h2><p><strong>主题：</strong>%s</p><p><strong>发件人：</strong>%s</p><p><strong>时间：</strong>%s</p><p><strong>内容：</strong></p><pre>%s</pre>",
			alert.Subject, alert.Sender, alert.ReceivedAt.Format("2006-01-02 15:04:05"), alert.Content)
	default:
		return fmt.Sprintf("邮件告警通知\n主题：%s\n发件人：%s\n时间：%s\n内容：%s",
			alert.Subject, alert.Sender, alert.ReceivedAt.Format("2006-01-02 15:04:05"), alert.Content)
	}
}

// processMessageLength 处理消息长度限制
func (s *notificationDispatcherService) processMessageLength(content, channelType string) string {
	var maxLength int
	switch channelType {
	case "dingtalk":
		maxLength = 20000 // 钉钉消息限制
	case "wechat":
		maxLength = 2048 // 企业微信消息限制
	case "webhook":
		maxLength = 50000 // 一般webhook限制
	case "email":
		return content // 邮件通常没有严格限制
	default:
		maxLength = 4000
	}

	if len(content) <= maxLength {
		return content
	}

	// 截断内容并添加提示
	truncated := content[:maxLength-100]
	return truncated + "\n\n...(内容过长已截断)..."
}

// sendNotification 发送通知
func (s *notificationDispatcherService) sendNotification(channel *model.Channel, subject, content string) NotificationResult {
	err := s.channelService.SendNotification(channel.ID, subject, content)
	if err != nil {
		return NotificationResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return NotificationResult{
		Success:      true,
		ResponseData: "发送成功",
	}
}

// ProcessPendingAlerts 处理待处理的告警
func (s *notificationDispatcherService) ProcessPendingAlerts() error {
	// 获取状态为pending的告警
	alerts, err := s.alertRepo.GetPendingAlerts(s.batchSize)
	if err != nil {
		return fmt.Errorf("获取待处理告警失败: %v", err)
	}

	if len(alerts) == 0 {
		return nil
	}

	log.Printf("开始处理 %d 个待处理告警", len(alerts))

	// 并发处理告警
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.maxWorkers)

	for _, alert := range alerts {
		wg.Add(1)
		go func(a *model.Alert) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := s.processAlert(a); err != nil {
				log.Printf("处理告警 %d 失败: %v", a.ID, err)
			}
		}(&alert)
	}

	wg.Wait()
	log.Printf("完成处理 %d 个待处理告警", len(alerts))

	return nil
}

// RetryFailedNotifications 重试失败的通知
func (s *notificationDispatcherService) RetryFailedNotifications() error {
	// 获取失败的通知日志
	failedLogs, err := s.notificationLogRepo.GetFailedLogs(s.maxRetryCount)
	if err != nil {
		return fmt.Errorf("获取失败通知日志失败: %v", err)
	}

	if len(failedLogs) == 0 {
		return nil
	}

	log.Printf("开始重试 %d 个失败通知", len(failedLogs))

	for _, logEntry := range failedLogs {
		// 增加重试次数
		s.notificationLogRepo.IncrementRetryCount(logEntry.ID)

		// 重新发送通知
		result := s.sendNotification(&logEntry.Channel, "重试通知", logEntry.Content)

		status := "failed"
		if result.Success {
			status = "success"
		}

		// 更新状态
		s.notificationLogRepo.UpdateStatus(logEntry.ID, status, result.Error, result.ResponseData)

		if result.Success {
			log.Printf("重试通知 %d 成功", logEntry.ID)
		} else {
			log.Printf("重试通知 %d 失败: %s", logEntry.ID, result.Error)
		}
	}

	return nil
}

// StartBackgroundProcessor 启动后台处理器
func (s *notificationDispatcherService) StartBackgroundProcessor(ctx context.Context) error {
	log.Println("启动通知分发后台处理器")

	// 启动告警处理工作器
	for i := 0; i < s.maxWorkers; i++ {
		s.workerWg.Add(1)
		go s.alertWorker(ctx)
	}

	// 启动重试工作器
	s.workerWg.Add(1)
	go s.retryWorker(ctx)

	// 启动定时任务
	s.workerWg.Add(1)
	go s.scheduledTasks(ctx)

	// 等待所有工作器完成
	go func() {
		s.workerWg.Wait()
		log.Println("通知分发后台处理器已停止")
	}()

	return nil
}

// alertWorker 告警处理工作器
func (s *notificationDispatcherService) alertWorker(ctx context.Context) {
	defer s.workerWg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case alert := <-s.alertQueue:
			if err := s.processAlert(alert); err != nil {
				log.Printf("处理告警 %d 失败: %v", alert.ID, err)
			}
		case retryLog := <-s.retryQueue:
			s.processRetryNotification(retryLog)
		}
	}
}

// retryWorker 重试工作器
func (s *notificationDispatcherService) retryWorker(ctx context.Context) {
	defer s.workerWg.Done()
	ticker := time.NewTicker(s.retryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.RetryFailedNotifications(); err != nil {
				log.Printf("重试失败通知时出错: %v", err)
			}
		}
	}
}

// scheduledTasks 定时任务
func (s *notificationDispatcherService) scheduledTasks(ctx context.Context) {
	defer s.workerWg.Done()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 定期处理待处理的告警
			if err := s.ProcessPendingAlerts(); err != nil {
				log.Printf("处理待处理告警时出错: %v", err)
			}
		}
	}
}

// processRetryNotification 处理重试通知
func (s *notificationDispatcherService) processRetryNotification(logEntry *model.NotificationLog) {
	// 增加重试次数
	s.notificationLogRepo.IncrementRetryCount(logEntry.ID)

	// 重新发送通知
	result := s.sendNotification(&logEntry.Channel, "重试通知", logEntry.Content)

	status := "failed"
	if result.Success {
		status = "success"
	}

	// 更新状态
	s.notificationLogRepo.UpdateStatus(logEntry.ID, status, result.Error, result.ResponseData)
}

// GetDispatchStats 获取分发统计信息
func (s *notificationDispatcherService) GetDispatchStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取通知日志统计
	logStats, err := s.notificationLogRepo.GetTodayStats()
	if err != nil {
		return nil, fmt.Errorf("获取通知日志统计失败: %v", err)
	}

	// 合并统计信息
	for key, value := range logStats {
		stats[key] = value
	}

	// 添加系统状态
	stats["queue_size"] = len(s.alertQueue)
	stats["retry_queue_size"] = len(s.retryQueue)
	stats["max_workers"] = s.maxWorkers
	stats["max_retry_count"] = s.maxRetryCount

	return stats, nil
}
