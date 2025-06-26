package service

import (
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"emailAlert/pkg/email"
	"errors"
	"fmt"
)

// MailboxService 邮箱配置服务层
type MailboxService struct {
	mailboxRepo *repository.MailboxRepository
}

// NewMailboxService 创建邮箱配置服务实例
func NewMailboxService(mailboxRepo *repository.MailboxRepository) *MailboxService {
	return &MailboxService{mailboxRepo: mailboxRepo}
}

// CreateMailboxRequest 创建邮箱配置请求结构
type CreateMailboxRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Host        string `json:"host" binding:"required"`
	Port        int    `json:"port" binding:"required,min=1,max=65535"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Protocol    string `json:"protocol" binding:"required,oneof=IMAP POP3"`
	SSL         bool   `json:"ssl"`
	Description string `json:"description"`
}

// UpdateMailboxRequest 更新邮箱配置请求结构
type UpdateMailboxRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email" binding:"omitempty,email"`
	Host        string `json:"host"`
	Port        int    `json:"port" binding:"omitempty,min=1,max=65535"`
	Username    string `json:"username"`
	Password    string `json:"password"` // 为空时不更新密码
	Protocol    string `json:"protocol" binding:"omitempty,oneof=IMAP POP3"`
	SSL         bool   `json:"ssl"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
	Description string `json:"description"`
}

// MailboxListResponse 邮箱列表响应结构
type MailboxListResponse struct {
	List  []model.Mailbox `json:"list"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

// Create 创建邮箱配置
func (s *MailboxService) Create(req *CreateMailboxRequest) (*model.Mailbox, error) {
	// 检查邮箱地址是否已存在
	exists, err := s.mailboxRepo.EmailExists(req.Email, 0)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱地址失败: %v", err)
	}
	if exists {
		return nil, errors.New("邮箱地址已存在")
	}

	// 直接使用明文密码

	// 创建邮箱配置模型
	mailbox := &model.Mailbox{
		Name:        req.Name,
		Email:       req.Email,
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		Password:    req.Password,
		Protocol:    req.Protocol,
		SSL:         req.SSL,
		Status:      "active",
		Description: req.Description,
	}

	// 保存到数据库
	err = s.mailboxRepo.Create(mailbox)
	if err != nil {
		return nil, fmt.Errorf("创建邮箱配置失败: %v", err)
	}

	// 返回时不包含密码
	mailbox.Password = ""
	return mailbox, nil
}

// GetByID 根据ID获取邮箱配置
func (s *MailboxService) GetByID(id uint) (*model.Mailbox, error) {
	mailbox, err := s.mailboxRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 不返回密码
	mailbox.Password = ""
	return mailbox, nil
}

// GetByIDWithPassword 根据ID获取邮箱配置（包含密码，用于编辑）
func (s *MailboxService) GetByIDWithPassword(id uint) (*model.MailboxWithPassword, error) {
	mailbox, err := s.mailboxRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 转换为包含密码的结构
	result := &model.MailboxWithPassword{
		BaseModel:   mailbox.BaseModel,
		Name:        mailbox.Name,
		Email:       mailbox.Email,
		Host:        mailbox.Host,
		Port:        mailbox.Port,
		Username:    mailbox.Username,
		Password:    mailbox.Password, // 返回明文密码
		Protocol:    mailbox.Protocol,
		SSL:         mailbox.SSL,
		Status:      mailbox.Status,
		Description: mailbox.Description,
	}

	return result, nil
}

// List 获取邮箱配置列表
func (s *MailboxService) List(page, pageSize int, status string) (*MailboxListResponse, error) {
	// 参数验证
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	mailboxes, total, err := s.mailboxRepo.List(page, pageSize, status)
	if err != nil {
		return nil, fmt.Errorf("获取邮箱配置列表失败: %v", err)
	}

	// 清除密码字段
	for i := range mailboxes {
		mailboxes[i].Password = ""
	}

	return &MailboxListResponse{
		List:  mailboxes,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

// Update 更新邮箱配置
func (s *MailboxService) Update(id uint, req *UpdateMailboxRequest) (*model.Mailbox, error) {
	// 获取原有配置
	existingMailbox, err := s.mailboxRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 检查邮箱地址是否被其他配置使用
	if req.Email != "" && req.Email != existingMailbox.Email {
		exists, err := s.mailboxRepo.EmailExists(req.Email, id)
		if err != nil {
			return nil, fmt.Errorf("检查邮箱地址失败: %v", err)
		}
		if exists {
			return nil, errors.New("邮箱地址已被其他配置使用")
		}
	}

	// 准备更新数据
	updateData := &model.Mailbox{}

	if req.Name != "" {
		updateData.Name = req.Name
	}
	if req.Email != "" {
		updateData.Email = req.Email
	}
	if req.Host != "" {
		updateData.Host = req.Host
	}
	if req.Port > 0 {
		updateData.Port = req.Port
	}
	if req.Username != "" {
		updateData.Username = req.Username
	}
	if req.Password != "" {
		// 直接使用明文密码
		updateData.Password = req.Password
	}
	if req.Protocol != "" {
		updateData.Protocol = req.Protocol
	}
	updateData.SSL = req.SSL
	if req.Status != "" {
		updateData.Status = req.Status
	}
	updateData.Description = req.Description

	// 更新数据库
	err = s.mailboxRepo.Update(id, updateData)
	if err != nil {
		return nil, fmt.Errorf("更新邮箱配置失败: %v", err)
	}

	// 返回更新后的配置
	return s.GetByID(id)
}

// Delete 删除邮箱配置
func (s *MailboxService) Delete(id uint) error {
	// 检查配置是否存在
	_, err := s.mailboxRepo.GetByID(id)
	if err != nil {
		return err
	}

	// TODO: 检查是否有关联的告警规则，如果有则不允许删除

	return s.mailboxRepo.Delete(id)
}

// TestConnection 测试邮箱连接
func (s *MailboxService) TestConnection(id uint) (*email.ConnectionInfo, error) {
	// 获取邮箱配置
	mailbox, err := s.mailboxRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 直接使用明文密码
	// 创建邮件客户端
	client := email.NewClient(mailbox.Host, mailbox.Port, mailbox.Username, mailbox.Password, mailbox.SSL)

	// 获取连接信息
	return client.GetConnectionInfo(), nil
}

// TestConnectionWithConfig 使用配置参数测试连接
func (s *MailboxService) TestConnectionWithConfig(req *CreateMailboxRequest) (*email.ConnectionInfo, error) {
	// 创建邮件客户端
	client := email.NewClient(req.Host, req.Port, req.Username, req.Password, req.SSL)

	// 获取连接信息
	return client.GetConnectionInfo(), nil
}

// UpdateStatus 更新邮箱状态
func (s *MailboxService) UpdateStatus(id uint, status string) error {
	// 验证状态值
	if status != "active" && status != "inactive" {
		return errors.New("无效的状态值")
	}

	// 检查配置是否存在
	_, err := s.mailboxRepo.GetByID(id)
	if err != nil {
		return err
	}

	return s.mailboxRepo.UpdateStatus(id, status)
}

// GetActiveMailboxes 获取所有活跃的邮箱配置
func (s *MailboxService) GetActiveMailboxes() ([]model.Mailbox, error) {
	mailboxes, err := s.mailboxRepo.GetActiveMailboxes()
	if err != nil {
		return nil, err
	}

	// 清除密码字段
	for i := range mailboxes {
		mailboxes[i].Password = ""
	}

	return mailboxes, nil
}

// DiagnoseMailbox 诊断邮箱连接问题
func (s *MailboxService) DiagnoseMailbox(id uint) (*email.EmailDiagnosis, error) {
	// 获取邮箱配置
	mailbox, err := s.mailboxRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取邮箱配置失败: %v", err)
	}

	// 转换为邮箱配置结构
	mailboxConfig := email.MailboxConfig{
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

	// 执行诊断
	return email.DiagnoseMailbox(mailboxConfig), nil
}

// DiagnoseMailboxWithConfig 使用配置诊断邮箱连接问题
func (s *MailboxService) DiagnoseMailboxWithConfig(req *CreateMailboxRequest) *email.EmailDiagnosis {
	// 转换为邮箱配置结构
	mailboxConfig := email.MailboxConfig{
		Name:     req.Name,
		Email:    req.Email,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Protocol: req.Protocol,
		SSL:      req.SSL,
		Status:   "active",
	}

	// 执行诊断
	return email.DiagnoseMailbox(mailboxConfig)
}
