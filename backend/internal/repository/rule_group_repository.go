package repository

import (
	"emailAlert/internal/model"

	"gorm.io/gorm"
)

// RuleGroupRepository 规则组仓库接口
type RuleGroupRepository interface {
	Create(ruleGroup *model.RuleGroup) error
	GetByID(id uint) (*model.RuleGroup, error)
	GetAll(page, size int, filters map[string]interface{}) ([]*model.RuleGroup, int64, error)
	Update(ruleGroup *model.RuleGroup) error
	Delete(id uint) error
	GetByMailboxID(mailboxID uint) ([]*model.RuleGroup, error)
	GetActiveRuleGroups() ([]*model.RuleGroup, error)
	GetWithConditions(id uint) (*model.RuleGroup, error)
}

// ruleGroupRepository 规则组仓库实现
type ruleGroupRepository struct {
	db *gorm.DB
}

// NewRuleGroupRepository 创建新的规则组仓库
func NewRuleGroupRepository(db *gorm.DB) RuleGroupRepository {
	return &ruleGroupRepository{db: db}
}

// Create 创建规则组
func (r *ruleGroupRepository) Create(ruleGroup *model.RuleGroup) error {
	return r.db.Create(ruleGroup).Error
}

// GetByID 根据ID获取规则组
func (r *ruleGroupRepository) GetByID(id uint) (*model.RuleGroup, error) {
	var ruleGroup model.RuleGroup
	err := r.db.Preload("Mailbox").First(&ruleGroup, id).Error
	if err != nil {
		return nil, err
	}
	return &ruleGroup, nil
}

// GetAll 获取所有规则组（带分页）
func (r *ruleGroupRepository) GetAll(page, size int, filters map[string]interface{}) ([]*model.RuleGroup, int64, error) {
	var ruleGroups []*model.RuleGroup
	var total int64

	query := r.db.Model(&model.RuleGroup{})

	// 应用过滤条件
	if mailboxID, ok := filters["mailbox_id"]; ok && mailboxID != nil {
		query = query.Where("mailbox_id = ?", mailboxID)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if logic, ok := filters["logic"]; ok && logic != "" {
		query = query.Where("logic = ?", logic)
	}
	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+name.(string)+"%")
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err = query.Preload("Mailbox").Preload("Conditions").
		Order("priority DESC, created_at DESC").
		Offset(offset).Limit(size).Find(&ruleGroups).Error

	return ruleGroups, total, err
}

// Update 更新规则组
func (r *ruleGroupRepository) Update(ruleGroup *model.RuleGroup) error {
	return r.db.Save(ruleGroup).Error
}

// Delete 删除规则组（软删除）
func (r *ruleGroupRepository) Delete(id uint) error {
	return r.db.Delete(&model.RuleGroup{}, id).Error
}

// GetByMailboxID 根据邮箱ID获取规则组
func (r *ruleGroupRepository) GetByMailboxID(mailboxID uint) ([]*model.RuleGroup, error) {
	var ruleGroups []*model.RuleGroup
	err := r.db.Where("mailbox_id = ? AND status = ?", mailboxID, "active").
		Preload("Conditions", "status = ?", "active").
		Order("priority DESC").Find(&ruleGroups).Error
	return ruleGroups, err
}

// GetActiveRuleGroups 获取所有激活的规则组
func (r *ruleGroupRepository) GetActiveRuleGroups() ([]*model.RuleGroup, error) {
	var ruleGroups []*model.RuleGroup
	err := r.db.Where("status = ?", "active").
		Preload("Mailbox").
		Preload("Conditions", "status = ?", "active").
		Order("priority DESC").Find(&ruleGroups).Error
	return ruleGroups, err
}

// GetWithConditions 获取规则组及其所有条件
func (r *ruleGroupRepository) GetWithConditions(id uint) (*model.RuleGroup, error) {
	var ruleGroup model.RuleGroup
	err := r.db.Preload("Mailbox").
		Preload("Conditions", func(db *gorm.DB) *gorm.DB {
			return db.Order("priority DESC")
		}).
		First(&ruleGroup, id).Error
	if err != nil {
		return nil, err
	}
	return &ruleGroup, nil
}
