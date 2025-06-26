package service

import (
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"errors"
	"fmt"
)

// RuleGroupService 规则组服务接口
type RuleGroupService interface {
	CreateRuleGroup(ruleGroup *model.RuleGroup) error
	GetRuleGroupByID(id uint) (*model.RuleGroup, error)
	GetRuleGroups(page, size int, filters map[string]interface{}) ([]*model.RuleGroup, int64, error)
	UpdateRuleGroup(ruleGroup *model.RuleGroup) error
	UpdateRuleGroupStatus(id uint, status string) error
	DeleteRuleGroup(id uint) error
	GetRuleGroupsByMailboxID(mailboxID uint) ([]*model.RuleGroup, error)
	GetActiveRuleGroups() ([]*model.RuleGroup, error)
	ValidateRuleGroup(ruleGroup *model.RuleGroup) error
	GetRuleGroupWithConditions(id uint) (*model.RuleGroup, error)
	ProcessRuleGroupWithConditions(ruleGroupData *RuleGroupData) error
}

// ruleGroupService 规则组服务实现
type ruleGroupService struct {
	ruleGroupRepo        repository.RuleGroupRepository
	conditionRepo        repository.MatchConditionRepository
	ruleGroupChannelRepo repository.RuleGroupChannelRepository
	mailboxRepo          *repository.MailboxRepository
}

// RuleGroupData 规则组数据结构（包含条件和通知渠道）
type RuleGroupData struct {
	RuleGroup  *model.RuleGroup        `json:"rule_group"`
	Conditions []*model.MatchCondition `json:"conditions"`
	ChannelIDs []uint                  `json:"channel_ids"` // 新增：关联的通知渠道ID列表
}

// NewRuleGroupService 创建新的规则组服务
func NewRuleGroupService(
	ruleGroupRepo repository.RuleGroupRepository,
	conditionRepo repository.MatchConditionRepository,
	ruleGroupChannelRepo repository.RuleGroupChannelRepository,
	mailboxRepo *repository.MailboxRepository,
) RuleGroupService {
	return &ruleGroupService{
		ruleGroupRepo:        ruleGroupRepo,
		conditionRepo:        conditionRepo,
		ruleGroupChannelRepo: ruleGroupChannelRepo,
		mailboxRepo:          mailboxRepo,
	}
}

// CreateRuleGroup 创建规则组
func (s *ruleGroupService) CreateRuleGroup(ruleGroup *model.RuleGroup) error {
	// 验证规则组
	if err := s.ValidateRuleGroup(ruleGroup); err != nil {
		return err
	}

	// 验证邮箱是否存在
	_, err := s.mailboxRepo.GetByID(ruleGroup.MailboxID)
	if err != nil {
		return errors.New("关联的邮箱不存在")
	}

	return s.ruleGroupRepo.Create(ruleGroup)
}

// GetRuleGroupByID 根据ID获取规则组
func (s *ruleGroupService) GetRuleGroupByID(id uint) (*model.RuleGroup, error) {
	return s.ruleGroupRepo.GetByID(id)
}

// GetRuleGroups 获取规则组列表
func (s *ruleGroupService) GetRuleGroups(page, size int, filters map[string]interface{}) ([]*model.RuleGroup, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 10
	}

	return s.ruleGroupRepo.GetAll(page, size, filters)
}

// UpdateRuleGroup 更新规则组
func (s *ruleGroupService) UpdateRuleGroup(ruleGroup *model.RuleGroup) error {
	// 验证规则组是否存在
	existingRuleGroup, err := s.ruleGroupRepo.GetByID(ruleGroup.ID)
	if err != nil {
		return errors.New("规则组不存在")
	}

	// 验证规则组内容
	if err := s.ValidateRuleGroup(ruleGroup); err != nil {
		return err
	}

	// 验证邮箱是否存在
	_, err = s.mailboxRepo.GetByID(ruleGroup.MailboxID)
	if err != nil {
		return errors.New("关联的邮箱不存在")
	}

	// 保留创建时间
	ruleGroup.CreatedAt = existingRuleGroup.CreatedAt

	return s.ruleGroupRepo.Update(ruleGroup)
}

// UpdateRuleGroupStatus 更新规则组状态
func (s *ruleGroupService) UpdateRuleGroupStatus(id uint, status string) error {
	// 验证状态值
	validStatuses := []string{"active", "inactive"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("无效的状态值")
	}

	// 验证规则组是否存在
	existingRuleGroup, err := s.ruleGroupRepo.GetByID(id)
	if err != nil {
		return errors.New("规则组不存在")
	}

	// 更新状态
	existingRuleGroup.Status = status
	return s.ruleGroupRepo.Update(existingRuleGroup)
}

// DeleteRuleGroup 删除规则组
func (s *ruleGroupService) DeleteRuleGroup(id uint) error {
	// 验证规则组是否存在
	_, err := s.ruleGroupRepo.GetByID(id)
	if err != nil {
		return errors.New("规则组不存在")
	}

	// 删除规则组相关的所有条件
	if err := s.conditionRepo.DeleteByRuleGroupID(id); err != nil {
		return fmt.Errorf("删除规则组条件失败: %v", err)
	}

	// 删除规则组相关的所有通知渠道关联
	if err := s.ruleGroupChannelRepo.DeleteByRuleGroupID(id); err != nil {
		return fmt.Errorf("删除规则组通知渠道关联失败: %v", err)
	}

	return s.ruleGroupRepo.Delete(id)
}

// GetRuleGroupsByMailboxID 根据邮箱ID获取规则组
func (s *ruleGroupService) GetRuleGroupsByMailboxID(mailboxID uint) ([]*model.RuleGroup, error) {
	return s.ruleGroupRepo.GetByMailboxID(mailboxID)
}

// GetActiveRuleGroups 获取所有激活的规则组
func (s *ruleGroupService) GetActiveRuleGroups() ([]*model.RuleGroup, error) {
	return s.ruleGroupRepo.GetActiveRuleGroups()
}

// ValidateRuleGroup 验证规则组
func (s *ruleGroupService) ValidateRuleGroup(ruleGroup *model.RuleGroup) error {
	if ruleGroup.Name == "" {
		return errors.New("规则组名称不能为空")
	}

	if ruleGroup.MailboxID == 0 {
		return errors.New("必须选择邮箱")
	}

	// 验证逻辑类型
	validLogicTypes := []string{"and", "or"}
	if ruleGroup.Logic == "" {
		ruleGroup.Logic = "and" // 默认使用AND逻辑
	} else if !contains(validLogicTypes, ruleGroup.Logic) {
		return errors.New("无效的逻辑类型")
	}

	// 验证状态
	validStatuses := []string{"active", "inactive"}
	if ruleGroup.Status != "" && !contains(validStatuses, ruleGroup.Status) {
		return errors.New("无效的状态")
	}

	// 验证优先级
	if ruleGroup.Priority < 1 || ruleGroup.Priority > 10 {
		ruleGroup.Priority = 1 // 设置默认优先级
	}

	return nil
}

// GetRuleGroupWithConditions 获取规则组及其条件和通知渠道
func (s *ruleGroupService) GetRuleGroupWithConditions(id uint) (*model.RuleGroup, error) {
	// 获取规则组及其条件
	ruleGroup, err := s.ruleGroupRepo.GetWithConditions(id)
	if err != nil {
		return nil, err
	}

	// 获取关联的通知渠道
	channels, err := s.ruleGroupChannelRepo.GetChannelsByRuleGroupID(id)
	if err != nil {
		return nil, fmt.Errorf("获取规则组通知渠道失败: %v", err)
	}

	// 将通知渠道添加到规则组中
	if len(channels) > 0 {
		ruleGroup.Channels = make([]model.Channel, len(channels))
		for i, channel := range channels {
			ruleGroup.Channels[i] = *channel
		}
	}

	return ruleGroup, nil
}

// ProcessRuleGroupWithConditions 处理规则组和条件的创建/更新
func (s *ruleGroupService) ProcessRuleGroupWithConditions(ruleGroupData *RuleGroupData) error {
	// 验证规则组
	if err := s.ValidateRuleGroup(ruleGroupData.RuleGroup); err != nil {
		return err
	}

	// 如果是新建规则组
	if ruleGroupData.RuleGroup.ID == 0 {
		// 创建规则组
		if err := s.ruleGroupRepo.Create(ruleGroupData.RuleGroup); err != nil {
			return fmt.Errorf("创建规则组失败: %v", err)
		}

		// 设置条件的规则组ID
		for _, condition := range ruleGroupData.Conditions {
			condition.RuleGroupID = ruleGroupData.RuleGroup.ID
		}

		// 批量创建条件
		if len(ruleGroupData.Conditions) > 0 {
			if err := s.conditionRepo.BatchCreate(ruleGroupData.Conditions); err != nil {
				return fmt.Errorf("创建匹配条件失败: %v", err)
			}
		}

		// 创建通知渠道关联
		if len(ruleGroupData.ChannelIDs) > 0 {
			if err := s.ruleGroupChannelRepo.BatchCreate(ruleGroupData.RuleGroup.ID, ruleGroupData.ChannelIDs, 1); err != nil {
				return fmt.Errorf("关联通知渠道失败: %v", err)
			}
		}
	} else {
		// 更新规则组
		if err := s.ruleGroupRepo.Update(ruleGroupData.RuleGroup); err != nil {
			return fmt.Errorf("更新规则组失败: %v", err)
		}

		// 删除原有条件
		if err := s.conditionRepo.DeleteByRuleGroupID(ruleGroupData.RuleGroup.ID); err != nil {
			return fmt.Errorf("删除原有条件失败: %v", err)
		}

		// 设置条件的规则组ID
		for _, condition := range ruleGroupData.Conditions {
			condition.RuleGroupID = ruleGroupData.RuleGroup.ID
			condition.ID = 0 // 重置ID，作为新条件创建
		}

		// 批量创建新条件
		if len(ruleGroupData.Conditions) > 0 {
			if err := s.conditionRepo.BatchCreate(ruleGroupData.Conditions); err != nil {
				return fmt.Errorf("创建新匹配条件失败: %v", err)
			}
		}

		// 更新通知渠道关联
		// 先删除原有关联
		if err := s.ruleGroupChannelRepo.DeleteByRuleGroupID(ruleGroupData.RuleGroup.ID); err != nil {
			return fmt.Errorf("删除原有通知渠道关联失败: %v", err)
		}

		// 创建新的关联
		if len(ruleGroupData.ChannelIDs) > 0 {
			if err := s.ruleGroupChannelRepo.BatchCreate(ruleGroupData.RuleGroup.ID, ruleGroupData.ChannelIDs, 1); err != nil {
				return fmt.Errorf("关联通知渠道失败: %v", err)
			}
		}
	}

	return nil
}

// contains 检查字符串切片是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
