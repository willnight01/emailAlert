package repository

import (
	"emailAlert/internal/model"

	"gorm.io/gorm"
)

// RuleChannelRepository 规则渠道关联仓库接口
type RuleChannelRepository interface {
	Create(ruleChannel *model.RuleChannel) error
	GetByRuleID(ruleID uint) ([]*model.RuleChannel, error)
	GetByChannelID(channelID uint) ([]*model.RuleChannel, error)
	GetChannelsByRuleID(ruleID uint) ([]*model.Channel, error)
	GetRulesByChannelID(channelID uint) ([]*model.AlertRule, error)
	Delete(ruleID, channelID uint) error
	DeleteByRuleID(ruleID uint) error
	DeleteByChannelID(channelID uint) error
	BatchCreate(ruleID uint, channelIDs []uint, priority int) error
	UpdatePriority(ruleID, channelID uint, priority int) error
}

// ruleChannelRepository 规则渠道关联仓库实现
type ruleChannelRepository struct {
	db *gorm.DB
}

// NewRuleChannelRepository 创建规则渠道关联仓库
func NewRuleChannelRepository(db *gorm.DB) RuleChannelRepository {
	return &ruleChannelRepository{db: db}
}

// Create 创建规则渠道关联
func (r *ruleChannelRepository) Create(ruleChannel *model.RuleChannel) error {
	return r.db.Create(ruleChannel).Error
}

// GetByRuleID 根据规则ID获取关联关系
func (r *ruleChannelRepository) GetByRuleID(ruleID uint) ([]*model.RuleChannel, error) {
	var ruleChannels []*model.RuleChannel
	err := r.db.Where("rule_id = ?", ruleID).Order("priority DESC").Find(&ruleChannels).Error
	return ruleChannels, err
}

// GetByChannelID 根据渠道ID获取关联关系
func (r *ruleChannelRepository) GetByChannelID(channelID uint) ([]*model.RuleChannel, error) {
	var ruleChannels []*model.RuleChannel
	err := r.db.Where("channel_id = ?", channelID).Find(&ruleChannels).Error
	return ruleChannels, err
}

// GetChannelsByRuleID 根据规则ID获取关联的渠道列表
func (r *ruleChannelRepository) GetChannelsByRuleID(ruleID uint) ([]*model.Channel, error) {
	var channels []*model.Channel
	err := r.db.Table("channels").
		Joins("JOIN rule_channels ON channels.id = rule_channels.channel_id").
		Where("rule_channels.rule_id = ? AND channels.status = ?", ruleID, "active").
		Order("rule_channels.priority DESC").
		Find(&channels).Error
	return channels, err
}

// GetRulesByChannelID 根据渠道ID获取关联的规则列表
func (r *ruleChannelRepository) GetRulesByChannelID(channelID uint) ([]*model.AlertRule, error) {
	var rules []*model.AlertRule
	err := r.db.Table("alert_rules").
		Joins("JOIN rule_channels ON alert_rules.id = rule_channels.rule_id").
		Where("rule_channels.channel_id = ? AND alert_rules.status = ?", channelID, "active").
		Find(&rules).Error
	return rules, err
}

// Delete 删除指定的规则渠道关联
func (r *ruleChannelRepository) Delete(ruleID, channelID uint) error {
	return r.db.Where("rule_id = ? AND channel_id = ?", ruleID, channelID).
		Delete(&model.RuleChannel{}).Error
}

// DeleteByRuleID 删除规则的所有渠道关联
func (r *ruleChannelRepository) DeleteByRuleID(ruleID uint) error {
	return r.db.Where("rule_id = ?", ruleID).Delete(&model.RuleChannel{}).Error
}

// DeleteByChannelID 删除渠道的所有规则关联
func (r *ruleChannelRepository) DeleteByChannelID(channelID uint) error {
	return r.db.Where("channel_id = ?", channelID).Delete(&model.RuleChannel{}).Error
}

// BatchCreate 批量创建规则渠道关联
func (r *ruleChannelRepository) BatchCreate(ruleID uint, channelIDs []uint, priority int) error {
	if len(channelIDs) == 0 {
		return nil
	}

	var ruleChannels []model.RuleChannel
	for _, channelID := range channelIDs {
		ruleChannels = append(ruleChannels, model.RuleChannel{
			RuleID:    ruleID,
			ChannelID: channelID,
			Priority:  priority,
		})
	}

	return r.db.Create(&ruleChannels).Error
}

// UpdatePriority 更新规则渠道关联的优先级
func (r *ruleChannelRepository) UpdatePriority(ruleID, channelID uint, priority int) error {
	return r.db.Model(&model.RuleChannel{}).
		Where("rule_id = ? AND channel_id = ?", ruleID, channelID).
		Update("priority", priority).Error
}
