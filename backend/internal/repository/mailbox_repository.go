package repository

import (
	"emailAlert/internal/model"
	"errors"

	"gorm.io/gorm"
)

// MailboxRepository 邮箱配置数据访问层
type MailboxRepository struct {
	db *gorm.DB
}

// NewMailboxRepository 创建邮箱配置仓库实例
func NewMailboxRepository(db *gorm.DB) *MailboxRepository {
	return &MailboxRepository{db: db}
}

// Create 创建邮箱配置
func (r *MailboxRepository) Create(mailbox *model.Mailbox) error {
	return r.db.Create(mailbox).Error
}

// GetByID 根据ID获取邮箱配置
func (r *MailboxRepository) GetByID(id uint) (*model.Mailbox, error) {
	var mailbox model.Mailbox
	err := r.db.First(&mailbox, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("邮箱配置不存在")
		}
		return nil, err
	}
	return &mailbox, nil
}

// GetByEmail 根据邮箱地址获取配置
func (r *MailboxRepository) GetByEmail(email string) (*model.Mailbox, error) {
	var mailbox model.Mailbox
	err := r.db.Where("email = ?", email).First(&mailbox).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("邮箱配置不存在")
		}
		return nil, err
	}
	return &mailbox, nil
}

// List 获取邮箱配置列表
func (r *MailboxRepository) List(page, pageSize int, status string) ([]model.Mailbox, int64, error) {
	var mailboxes []model.Mailbox
	var total int64

	query := r.db.Model(&model.Mailbox{})

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&mailboxes).Error
	if err != nil {
		return nil, 0, err
	}

	return mailboxes, total, nil
}

// Update 更新邮箱配置
func (r *MailboxRepository) Update(id uint, mailbox *model.Mailbox) error {
	return r.db.Model(&model.Mailbox{}).Where("id = ?", id).Updates(mailbox).Error
}

// Delete 删除邮箱配置
func (r *MailboxRepository) Delete(id uint) error {
	return r.db.Delete(&model.Mailbox{}, id).Error
}

// UpdateStatus 更新邮箱状态
func (r *MailboxRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Mailbox{}).Where("id = ?", id).Update("status", status).Error
}

// GetActiveMailboxes 获取所有活跃的邮箱配置
func (r *MailboxRepository) GetActiveMailboxes() ([]model.Mailbox, error) {
	var mailboxes []model.Mailbox
	err := r.db.Where("status = ?", "active").Find(&mailboxes).Error
	return mailboxes, err
}

// EmailExists 检查邮箱地址是否已存在（排除软删除的记录）
func (r *MailboxRepository) EmailExists(email string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&model.Mailbox{}).Where("email = ?", email).Where("deleted_at IS NULL")

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}
