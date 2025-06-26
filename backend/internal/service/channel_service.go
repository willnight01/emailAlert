package service

import (
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"emailAlert/pkg/notification"
	"encoding/json"
	"fmt"
	"time"
)

// ChannelService 通知渠道服务接口
type ChannelService interface {
	CreateChannel(channel *model.Channel) error
	GetChannel(id uint) (*model.Channel, error)
	GetChannelList(page, size int, channelType, status string) ([]*model.Channel, int64, error)
	GetActiveChannels() ([]*model.Channel, error)
	UpdateChannel(channel *model.Channel) error
	DeleteChannel(id uint) error
	TestChannel(id uint) error
	TestChannelConfig(channelType, config string) error
	UpdateChannelStatus(id uint, status string) error
	GetChannelTypes() []string
	SendNotification(channelID uint, title, content string) error
}

// channelService 通知渠道服务实现
type channelService struct {
	channelRepo      repository.ChannelRepository
	wechatNotifier   *notification.WeChatNotifier
	dingtalkNotifier *notification.DingTalkNotifier
	webhookNotifier  *notification.WebhookNotifier
	emailNotifier    *notification.EmailNotifier
}

// NewChannelService 创建通知渠道服务
func NewChannelService(channelRepo repository.ChannelRepository) ChannelService {
	return &channelService{
		channelRepo:      channelRepo,
		wechatNotifier:   notification.NewWeChatNotifier(),
		dingtalkNotifier: notification.NewDingTalkNotifier(),
		webhookNotifier:  notification.NewWebhookNotifier(),
		emailNotifier:    notification.NewEmailNotifier(),
	}
}

// CreateChannel 创建通知渠道
func (s *channelService) CreateChannel(channel *model.Channel) error {
	// 验证配置
	if err := s.validateChannelConfig(channel.Type, channel.Config); err != nil {
		return fmt.Errorf("渠道配置验证失败: %v", err)
	}

	return s.channelRepo.Create(channel)
}

// GetChannel 获取通知渠道
func (s *channelService) GetChannel(id uint) (*model.Channel, error) {
	return s.channelRepo.GetByID(id)
}

// GetChannelList 获取通知渠道列表
func (s *channelService) GetChannelList(page, size int, channelType, status string) ([]*model.Channel, int64, error) {
	return s.channelRepo.GetList(page, size, channelType, status)
}

// GetActiveChannels 获取所有激活的通知渠道
func (s *channelService) GetActiveChannels() ([]*model.Channel, error) {
	return s.channelRepo.GetActiveChannels()
}

// UpdateChannel 更新通知渠道
func (s *channelService) UpdateChannel(channel *model.Channel) error {
	// 验证配置
	if err := s.validateChannelConfig(channel.Type, channel.Config); err != nil {
		return fmt.Errorf("渠道配置验证失败: %v", err)
	}

	return s.channelRepo.Update(channel)
}

// DeleteChannel 删除通知渠道
func (s *channelService) DeleteChannel(id uint) error {
	return s.channelRepo.Delete(id)
}

// TestChannel 测试通知渠道
func (s *channelService) TestChannel(id uint) error {
	channel, err := s.channelRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("获取渠道信息失败: %v", err)
	}

	testResult := "测试成功"
	testTime := time.Now()

	err = s.testChannelByType(channel.Type, channel.Config)
	if err != nil {
		testResult = fmt.Sprintf("测试失败: %v", err)
	}

	// 更新测试结果
	s.channelRepo.UpdateTestResult(id, testResult, testTime)

	return err
}

// TestChannelConfig 测试渠道配置
func (s *channelService) TestChannelConfig(channelType, config string) error {
	return s.testChannelByType(channelType, config)
}

// UpdateChannelStatus 更新渠道状态
func (s *channelService) UpdateChannelStatus(id uint, status string) error {
	return s.channelRepo.UpdateStatus(id, status)
}

// GetChannelTypes 获取支持的渠道类型
func (s *channelService) GetChannelTypes() []string {
	return []string{"wechat", "dingtalk", "email", "webhook"}
}

// SendNotification 发送通知
func (s *channelService) SendNotification(channelID uint, title, content string) error {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("获取渠道信息失败: %v", err)
	}

	if channel.Status != "active" {
		return fmt.Errorf("渠道已停用")
	}

	return s.sendNotificationByType(channel.Type, channel.Config, title, content)
}

// validateChannelConfig 验证渠道配置
func (s *channelService) validateChannelConfig(channelType, config string) error {
	switch channelType {
	case "wechat":
		var wechatConfig model.WeChatConfig
		if err := json.Unmarshal([]byte(config), &wechatConfig); err != nil {
			return fmt.Errorf("企业微信配置格式错误: %v", err)
		}
		return s.validateWeChatConfig(&wechatConfig)
	case "dingtalk":
		var dingConfig model.DingTalkConfig
		if err := json.Unmarshal([]byte(config), &dingConfig); err != nil {
			return fmt.Errorf("钉钉配置格式错误: %v", err)
		}
		return s.validateDingTalkConfig(&dingConfig)
	case "email":
		var emailConfig notification.EmailConfig
		if err := json.Unmarshal([]byte(config), &emailConfig); err != nil {
			return fmt.Errorf("邮件配置格式错误: %v", err)
		}
		return s.emailNotifier.ValidateConfig(&emailConfig)
	case "webhook":
		var webhookConfig notification.WebhookConfig
		if err := json.Unmarshal([]byte(config), &webhookConfig); err != nil {
			return fmt.Errorf("Webhook配置格式错误: %v", err)
		}
		return s.webhookNotifier.ValidateConfig(&webhookConfig)
	default:
		return fmt.Errorf("不支持的渠道类型: %s", channelType)
	}
}

// validateWeChatConfig 验证企业微信配置
func (s *channelService) validateWeChatConfig(config *model.WeChatConfig) error {
	if config.Type == "" {
		return fmt.Errorf("企业微信类型不能为空")
	}

	switch config.Type {
	case "robot":
		if config.WebhookURL == "" {
			return fmt.Errorf("机器人Webhook地址不能为空")
		}
	case "app":
		if config.CorpID == "" || config.Secret == "" || config.AgentID == 0 {
			return fmt.Errorf("企业微信应用配置不完整")
		}
		if config.ToUser == "" && config.ToParty == "" && config.ToTag == "" {
			return fmt.Errorf("必须指定接收人、部门或标签")
		}
	default:
		return fmt.Errorf("不支持的企业微信类型: %s", config.Type)
	}

	return nil
}

// validateDingTalkConfig 验证钉钉配置
func (s *channelService) validateDingTalkConfig(config *model.DingTalkConfig) error {
	if config.Type == "" {
		return fmt.Errorf("钉钉类型不能为空")
	}

	switch config.Type {
	case "robot":
		if config.WebhookURL == "" {
			return fmt.Errorf("群机器人Webhook地址不能为空")
		}
	case "work":
		if config.AppKey == "" || config.AppSecret == "" || config.AgentID == 0 {
			return fmt.Errorf("钉钉应用配置不完整")
		}
		if config.UserIDs == "" {
			return fmt.Errorf("接收用户ID不能为空")
		}
	default:
		return fmt.Errorf("不支持的钉钉类型: %s", config.Type)
	}

	return nil
}

// testChannelByType 根据类型测试渠道
func (s *channelService) testChannelByType(channelType, config string) error {
	switch channelType {
	case "wechat":
		return s.testWeChatChannel(config)
	case "dingtalk":
		return s.testDingTalkChannel(config)
	case "email":
		return s.testEmailChannel(config)
	case "webhook":
		return s.testWebhookChannel(config)
	default:
		return fmt.Errorf("不支持的渠道类型: %s", channelType)
	}
}

// testWeChatChannel 测试企业微信渠道
func (s *channelService) testWeChatChannel(config string) error {
	var wechatConfig model.WeChatConfig
	if err := json.Unmarshal([]byte(config), &wechatConfig); err != nil {
		return err
	}

	switch wechatConfig.Type {
	case "robot":
		return s.wechatNotifier.TestRobotConnection(wechatConfig.WebhookURL, wechatConfig.Key)
	case "app":
		return s.wechatNotifier.TestAppConnection(wechatConfig.CorpID, wechatConfig.Secret, wechatConfig.AgentID, wechatConfig.ToUser)
	default:
		return fmt.Errorf("不支持的企业微信类型: %s", wechatConfig.Type)
	}
}

// testDingTalkChannel 测试钉钉渠道
func (s *channelService) testDingTalkChannel(config string) error {
	var dingConfig model.DingTalkConfig
	if err := json.Unmarshal([]byte(config), &dingConfig); err != nil {
		return err
	}

	switch dingConfig.Type {
	case "robot":
		return s.dingtalkNotifier.TestRobotConnection(dingConfig.WebhookURL, dingConfig.Secret)
	case "work":
		return s.dingtalkNotifier.TestWorkConnection(dingConfig.AppKey, dingConfig.AppSecret, dingConfig.AgentID, dingConfig.UserIDs)
	default:
		return fmt.Errorf("不支持的钉钉类型: %s", dingConfig.Type)
	}
}

// sendNotificationByType 根据类型发送通知
func (s *channelService) sendNotificationByType(channelType, config, title, content string) error {
	switch channelType {
	case "wechat":
		return s.sendWeChatNotification(config, title, content)
	case "dingtalk":
		return s.sendDingTalkNotification(config, title, content)
	case "email":
		return s.sendEmailNotification(config, title, content)
	case "webhook":
		return s.sendWebhookNotification(config, title, content)
	default:
		return fmt.Errorf("不支持的渠道类型: %s", channelType)
	}
}

// sendWeChatNotification 发送企业微信通知
func (s *channelService) sendWeChatNotification(config, title, content string) error {
	var wechatConfig model.WeChatConfig
	if err := json.Unmarshal([]byte(config), &wechatConfig); err != nil {
		return err
	}

	message := fmt.Sprintf("**%s**\n\n%s", title, content)

	switch wechatConfig.Type {
	case "robot":
		return s.wechatNotifier.SendRobotMessage(wechatConfig.WebhookURL, wechatConfig.Key, message, "markdown")
	case "app":
		return s.wechatNotifier.SendAppMessage(wechatConfig.CorpID, wechatConfig.Secret, wechatConfig.AgentID,
			wechatConfig.ToUser, wechatConfig.ToParty, wechatConfig.ToTag, title, content)
	default:
		return fmt.Errorf("不支持的企业微信类型: %s", wechatConfig.Type)
	}
}

// sendDingTalkNotification 发送钉钉通知
func (s *channelService) sendDingTalkNotification(config, title, content string) error {
	var dingConfig model.DingTalkConfig
	if err := json.Unmarshal([]byte(config), &dingConfig); err != nil {
		return err
	}

	switch dingConfig.Type {
	case "robot":
		// 钉钉群机器人支持markdown格式
		message := fmt.Sprintf("## %s\n\n%s", title, content)
		return s.dingtalkNotifier.SendRobotMessage(dingConfig.WebhookURL, dingConfig.Secret, message, title, "markdown")
	case "work":
		// 钉钉工作通知使用OA消息格式
		return s.dingtalkNotifier.SendWorkMessage(dingConfig.AppKey, dingConfig.AppSecret, dingConfig.AgentID, dingConfig.UserIDs, title, content)
	default:
		return fmt.Errorf("不支持的钉钉类型: %s", dingConfig.Type)
	}
}

// testWebhookChannel 测试Webhook渠道
func (s *channelService) testWebhookChannel(config string) error {
	var webhookConfig notification.WebhookConfig
	if err := json.Unmarshal([]byte(config), &webhookConfig); err != nil {
		return err
	}

	// 先验证配置
	if err := s.webhookNotifier.ValidateConfig(&webhookConfig); err != nil {
		return err
	}

	// 然后测试连接
	response, err := s.webhookNotifier.TestConnection(&webhookConfig)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("Webhook测试失败: %s", response.Error)
	}

	return nil
}

// sendWebhookNotification 发送Webhook通知
func (s *channelService) sendWebhookNotification(config, title, content string) error {
	var webhookConfig notification.WebhookConfig
	if err := json.Unmarshal([]byte(config), &webhookConfig); err != nil {
		return err
	}

	response, err := s.webhookNotifier.SendMessage(&webhookConfig, title, content)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("Webhook消息发送失败: %s", response.Error)
	}

	return nil
}

// testEmailChannel 测试邮件渠道
func (s *channelService) testEmailChannel(config string) error {
	var emailConfig notification.EmailConfig
	if err := json.Unmarshal([]byte(config), &emailConfig); err != nil {
		return err
	}

	// 先验证配置
	if err := s.emailNotifier.ValidateConfig(&emailConfig); err != nil {
		return err
	}

	// 然后测试连接
	response, err := s.emailNotifier.TestConnection(&emailConfig)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("邮件测试失败: %s", response.Error)
	}

	return nil
}

// sendEmailNotification 发送邮件通知
func (s *channelService) sendEmailNotification(config, title, content string) error {
	var emailConfig notification.EmailConfig
	if err := json.Unmarshal([]byte(config), &emailConfig); err != nil {
		return err
	}

	response, err := s.emailNotifier.SendMessage(&emailConfig, title, content)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("邮件发送失败: %s", response.Error)
	}

	return nil
}
