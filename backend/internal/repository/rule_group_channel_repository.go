package repository

import (
	"emailAlert/internal/model"

	"gorm.io/gorm"
)

// RuleGroupChannelRepository 规则组渠道关联仓库接口
type RuleGroupChannelRepository interface {
	Create(ruleGroupChannel *model.RuleGroupChannel) error
	GetByRuleGroupID(ruleGroupID uint) ([]*model.RuleGroupChannel, error)
	GetByChannelID(channelID uint) ([]*model.RuleGroupChannel, error)
	GetChannelsByRuleGroupID(ruleGroupID uint) ([]*model.Channel, error)
	GetRuleGroupsByChannelID(channelID uint) ([]*model.RuleGroup, error)
	Delete(ruleGroupID, channelID uint) error
	DeleteByRuleGroupID(ruleGroupID uint) error
	DeleteByChannelID(channelID uint) error
	BatchCreate(ruleGroupID uint, channelIDs []uint, priority int) error
	UpdatePriority(ruleGroupID, channelID uint, priority int) error
}

// ruleGroupChannelRepository 规则组渠道关联仓库实现
type ruleGroupChannelRepository struct {
	db *gorm.DB
}

// NewRuleGroupChannelRepository 创建规则组渠道关联仓库
func NewRuleGroupChannelRepository(db *gorm.DB) RuleGroupChannelRepository {
	return &ruleGroupChannelRepository{db: db}
}

// Create 创建规则组渠道关联
func (r *ruleGroupChannelRepository) Create(ruleGroupChannel *model.RuleGroupChannel) error {
	return r.db.Create(ruleGroupChannel).Error
}

// GetByRuleGroupID 根据规则组ID获取关联
func (r *ruleGroupChannelRepository) GetByRuleGroupID(ruleGroupID uint) ([]*model.RuleGroupChannel, error) {
	var ruleGroupChannels []*model.RuleGroupChannel
	err := r.db.Where("rule_group_id = ?", ruleGroupID).Order("priority DESC").Find(&ruleGroupChannels).Error
	return ruleGroupChannels, err
}

// GetByChannelID 根据渠道ID获取关联
func (r *ruleGroupChannelRepository) GetByChannelID(channelID uint) ([]*model.RuleGroupChannel, error) {
	var ruleGroupChannels []*model.RuleGroupChannel
	err := r.db.Where("channel_id = ?", channelID).Find(&ruleGroupChannels).Error
	return ruleGroupChannels, err
}

// GetChannelsByRuleGroupID 根据规则组ID获取通知渠道
func (r *ruleGroupChannelRepository) GetChannelsByRuleGroupID(ruleGroupID uint) ([]*model.Channel, error) {
	var channels []*model.Channel
	err := r.db.Table("channels").
		Joins("JOIN rule_group_channels ON channels.id = rule_group_channels.channel_id").
		Where("rule_group_channels.rule_group_id = ? AND channels.status = ?", ruleGroupID, "active").
		Order("rule_group_channels.priority DESC").
		Find(&channels).Error
	return channels, err
}

// GetRuleGroupsByChannelID 根据渠道ID获取规则组
func (r *ruleGroupChannelRepository) GetRuleGroupsByChannelID(channelID uint) ([]*model.RuleGroup, error) {
	var ruleGroups []*model.RuleGroup
	err := r.db.Table("rule_groups").
		Joins("JOIN rule_group_channels ON rule_groups.id = rule_group_channels.rule_group_id").
		Where("rule_group_channels.channel_id = ? AND rule_groups.status = ?", channelID, "active").
		Find(&ruleGroups).Error
	return ruleGroups, err
}

// Delete 删除规则组渠道关联
func (r *ruleGroupChannelRepository) Delete(ruleGroupID, channelID uint) error {
	return r.db.Where("rule_group_id = ? AND channel_id = ?", ruleGroupID, channelID).
		Delete(&model.RuleGroupChannel{}).Error
}

// DeleteByRuleGroupID 根据规则组ID删除所有关联
func (r *ruleGroupChannelRepository) DeleteByRuleGroupID(ruleGroupID uint) error {
	return r.db.Where("rule_group_id = ?", ruleGroupID).Delete(&model.RuleGroupChannel{}).Error
}

// DeleteByChannelID 根据渠道ID删除所有关联
func (r *ruleGroupChannelRepository) DeleteByChannelID(channelID uint) error {
	return r.db.Where("channel_id = ?", channelID).Delete(&model.RuleGroupChannel{}).Error
}

// BatchCreate 批量创建规则组渠道关联
func (r *ruleGroupChannelRepository) BatchCreate(ruleGroupID uint, channelIDs []uint, priority int) error {
	if len(channelIDs) == 0 {
		return nil
	}

	var ruleGroupChannels []model.RuleGroupChannel
	for _, channelID := range channelIDs {
		ruleGroupChannels = append(ruleGroupChannels, model.RuleGroupChannel{
			RuleGroupID: ruleGroupID,
			ChannelID:   channelID,
			Priority:    priority,
		})
	}

	return r.db.Create(&ruleGroupChannels).Error
}

// UpdatePriority 更新规则组渠道关联的优先级
func (r *ruleGroupChannelRepository) UpdatePriority(ruleGroupID, channelID uint, priority int) error {
	return r.db.Model(&model.RuleGroupChannel{}).
		Where("rule_group_id = ? AND channel_id = ?", ruleGroupID, channelID).
		Update("priority", priority).Error
}
