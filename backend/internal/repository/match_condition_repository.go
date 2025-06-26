package repository

import (
	"emailAlert/internal/model"

	"gorm.io/gorm"
)

// MatchConditionRepository 匹配条件仓库接口
type MatchConditionRepository interface {
	Create(condition *model.MatchCondition) error
	GetByID(id uint) (*model.MatchCondition, error)
	GetAll(page, size int, filters map[string]interface{}) ([]*model.MatchCondition, int64, error)
	Update(condition *model.MatchCondition) error
	Delete(id uint) error
	GetByRuleGroupID(ruleGroupID uint) ([]*model.MatchCondition, error)
	GetActiveConditions() ([]*model.MatchCondition, error)
	BatchCreate(conditions []*model.MatchCondition) error
	BatchUpdate(conditions []*model.MatchCondition) error
	DeleteByRuleGroupID(ruleGroupID uint) error
}

// matchConditionRepository 匹配条件仓库实现
type matchConditionRepository struct {
	db *gorm.DB
}

// NewMatchConditionRepository 创建新的匹配条件仓库
func NewMatchConditionRepository(db *gorm.DB) MatchConditionRepository {
	return &matchConditionRepository{db: db}
}

// Create 创建匹配条件
func (r *matchConditionRepository) Create(condition *model.MatchCondition) error {
	return r.db.Create(condition).Error
}

// GetByID 根据ID获取匹配条件
func (r *matchConditionRepository) GetByID(id uint) (*model.MatchCondition, error) {
	var condition model.MatchCondition
	err := r.db.Preload("RuleGroup").First(&condition, id).Error
	if err != nil {
		return nil, err
	}
	return &condition, nil
}

// GetAll 获取所有匹配条件（带分页）
func (r *matchConditionRepository) GetAll(page, size int, filters map[string]interface{}) ([]*model.MatchCondition, int64, error) {
	var conditions []*model.MatchCondition
	var total int64

	query := r.db.Model(&model.MatchCondition{})

	// 应用过滤条件
	if ruleGroupID, ok := filters["rule_group_id"]; ok && ruleGroupID != nil {
		query = query.Where("rule_group_id = ?", ruleGroupID)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if fieldType, ok := filters["field_type"]; ok && fieldType != "" {
		query = query.Where("field_type = ?", fieldType)
	}
	if matchType, ok := filters["match_type"]; ok && matchType != "" {
		query = query.Where("match_type = ?", matchType)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err = query.Preload("RuleGroup").
		Order("priority DESC, created_at DESC").
		Offset(offset).Limit(size).Find(&conditions).Error

	return conditions, total, err
}

// Update 更新匹配条件
func (r *matchConditionRepository) Update(condition *model.MatchCondition) error {
	return r.db.Save(condition).Error
}

// Delete 删除匹配条件（软删除）
func (r *matchConditionRepository) Delete(id uint) error {
	return r.db.Delete(&model.MatchCondition{}, id).Error
}

// GetByRuleGroupID 根据规则组ID获取匹配条件
func (r *matchConditionRepository) GetByRuleGroupID(ruleGroupID uint) ([]*model.MatchCondition, error) {
	var conditions []*model.MatchCondition
	err := r.db.Where("rule_group_id = ? AND status = ?", ruleGroupID, "active").
		Order("priority DESC").Find(&conditions).Error
	return conditions, err
}

// GetActiveConditions 获取所有激活的匹配条件
func (r *matchConditionRepository) GetActiveConditions() ([]*model.MatchCondition, error) {
	var conditions []*model.MatchCondition
	err := r.db.Where("status = ?", "active").
		Preload("RuleGroup").
		Order("priority DESC").Find(&conditions).Error
	return conditions, err
}

// BatchCreate 批量创建匹配条件
func (r *matchConditionRepository) BatchCreate(conditions []*model.MatchCondition) error {
	if len(conditions) == 0 {
		return nil
	}
	return r.db.Create(&conditions).Error
}

// BatchUpdate 批量更新匹配条件
func (r *matchConditionRepository) BatchUpdate(conditions []*model.MatchCondition) error {
	if len(conditions) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, condition := range conditions {
			if err := tx.Save(condition).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteByRuleGroupID 删除指定规则组的所有条件
func (r *matchConditionRepository) DeleteByRuleGroupID(ruleGroupID uint) error {
	return r.db.Where("rule_group_id = ?", ruleGroupID).Delete(&model.MatchCondition{}).Error
}
