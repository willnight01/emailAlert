package service

import (
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// EnhancedRuleEngineService 增强版规则执行引擎服务接口
type EnhancedRuleEngineService interface {
	ProcessEmailWithRuleGroups(emailData *model.EmailData, mailboxID uint) ([]*EnhancedAlertResult, error)
	MatchRuleGroups(emailData *model.EmailData, ruleGroups []*model.RuleGroup) ([]*RuleGroupMatchResult, error)
	MatchConditions(emailData *model.EmailData, conditions []*model.MatchCondition) ([]*ConditionMatchResult, error)
	MatchSingleCondition(emailData *model.EmailData, condition *model.MatchCondition) (bool, string, error)
	ExtractEmailFields(emailData *model.EmailData) map[string]string
	GetEnhancedRuleEngineStats() (map[string]interface{}, error)
}

// enhancedRuleEngineService 增强版规则执行引擎服务实现
type enhancedRuleEngineService struct {
	ruleGroupRepo repository.RuleGroupRepository
	conditionRepo repository.MatchConditionRepository
	alertRepo     repository.AlertRepository
}

// EnhancedAlertResult 增强版告警处理结果
type EnhancedAlertResult struct {
	RuleGroup    *model.RuleGroup        `json:"rule_group"`
	Alert        *model.Alert            `json:"alert"`
	IsDuplicate  bool                    `json:"is_duplicate"`
	Created      bool                    `json:"created"`
	Error        string                  `json:"error,omitempty"`
	MatchDetails []*ConditionMatchResult `json:"match_details"`
}

// RuleGroupMatchResult 规则组匹配结果
type RuleGroupMatchResult struct {
	RuleGroup        *model.RuleGroup        `json:"rule_group"`
	Matched          bool                    `json:"matched"`
	MatchedCount     int                     `json:"matched_count"`
	TotalCount       int                     `json:"total_count"`
	Logic            string                  `json:"logic"`
	Reason           string                  `json:"reason"`
	ConditionResults []*ConditionMatchResult `json:"condition_results"`
}

// ConditionMatchResult 条件匹配结果
type ConditionMatchResult struct {
	Condition       *model.MatchCondition `json:"condition"`
	Matched         bool                  `json:"matched"`
	MatchedKeywords []string              `json:"matched_keywords"`
	FieldContent    string                `json:"field_content"`
	Reason          string                `json:"reason"`
}

// NewEnhancedRuleEngineService 创建新的增强版规则执行引擎服务
func NewEnhancedRuleEngineService(
	ruleGroupRepo repository.RuleGroupRepository,
	conditionRepo repository.MatchConditionRepository,
	alertRepo repository.AlertRepository,
) EnhancedRuleEngineService {
	return &enhancedRuleEngineService{
		ruleGroupRepo: ruleGroupRepo,
		conditionRepo: conditionRepo,
		alertRepo:     alertRepo,
	}
}

// ProcessEmailWithRuleGroups 使用规则组处理邮件
func (s *enhancedRuleEngineService) ProcessEmailWithRuleGroups(emailData *model.EmailData, mailboxID uint) ([]*EnhancedAlertResult, error) {
	var results []*EnhancedAlertResult

	// 1. 获取该邮箱的所有激活规则组
	ruleGroups, err := s.ruleGroupRepo.GetByMailboxID(mailboxID)
	if err != nil {
		return nil, fmt.Errorf("获取邮箱规则组失败: %v", err)
	}

	if len(ruleGroups) == 0 {
		log.Printf("邮箱 %d 没有配置规则组", mailboxID)
		return results, nil
	}

	// 2. 执行规则组匹配
	matchResults, err := s.MatchRuleGroups(emailData, ruleGroups)
	if err != nil {
		return nil, fmt.Errorf("规则组匹配失败: %v", err)
	}

	// 3. 处理匹配到的规则组（按优先级排序）
	for _, matchResult := range matchResults {
		if !matchResult.Matched {
			continue
		}

		result := &EnhancedAlertResult{
			RuleGroup:    matchResult.RuleGroup,
			MatchDetails: matchResult.ConditionResults,
		}

		// 4. 检查是否重复告警（基于MessageID和规则组ID）
		isDuplicate, err := s.CheckDuplicateByRuleGroup(emailData, matchResult.RuleGroup.ID)
		if err != nil {
			result.Error = fmt.Sprintf("检查重复告警失败: %v", err)
			results = append(results, result)
			continue
		}

		result.IsDuplicate = isDuplicate
		if isDuplicate {
			log.Printf("邮件 %s 在规则组 %s 下重复告警，跳过", emailData.MessageID, matchResult.RuleGroup.Name)
			results = append(results, result)
			continue
		}

		// 5. 创建告警记录
		alert, err := s.CreateAlertFromRuleGroup(emailData, matchResult.RuleGroup)
		if err != nil {
			result.Error = fmt.Sprintf("创建告警失败: %v", err)
			results = append(results, result)
			continue
		}

		result.Alert = alert
		result.Created = true
		results = append(results, result)

		log.Printf("邮件 %s 匹配规则组 %s，创建告警 ID: %d",
			emailData.MessageID, matchResult.RuleGroup.Name, alert.ID)
	}

	return results, nil
}

// MatchRuleGroups 执行规则组匹配
func (s *enhancedRuleEngineService) MatchRuleGroups(emailData *model.EmailData, ruleGroups []*model.RuleGroup) ([]*RuleGroupMatchResult, error) {
	var results []*RuleGroupMatchResult

	for _, ruleGroup := range ruleGroups {
		if ruleGroup.Status != "active" {
			continue
		}

		result := &RuleGroupMatchResult{
			RuleGroup: ruleGroup,
			Logic:     ruleGroup.Logic,
		}

		// 获取规则组的所有激活条件
		conditions, err := s.conditionRepo.GetByRuleGroupID(ruleGroup.ID)
		if err != nil {
			result.Matched = false
			result.Reason = fmt.Sprintf("获取规则组条件失败: %v", err)
			results = append(results, result)
			continue
		}

		if len(conditions) == 0 {
			result.Matched = false
			result.Reason = "规则组没有配置匹配条件"
			results = append(results, result)
			continue
		}

		// 执行条件匹配
		conditionResults, err := s.MatchConditions(emailData, conditions)
		if err != nil {
			result.Matched = false
			result.Reason = fmt.Sprintf("条件匹配失败: %v", err)
			results = append(results, result)
			continue
		}

		result.ConditionResults = conditionResults
		result.TotalCount = len(conditionResults)

		// 统计匹配成功的条件数量
		matchedCount := 0
		for _, condResult := range conditionResults {
			if condResult.Matched {
				matchedCount++
			}
		}
		result.MatchedCount = matchedCount

		// 根据规则组逻辑判断是否匹配
		if ruleGroup.Logic == "and" {
			result.Matched = matchedCount == result.TotalCount
			if result.Matched {
				result.Reason = fmt.Sprintf("所有 %d 个条件都匹配成功 (AND逻辑)", result.TotalCount)
			} else {
				result.Reason = fmt.Sprintf("只有 %d/%d 个条件匹配成功，AND逻辑要求全部匹配", matchedCount, result.TotalCount)
			}
		} else { // or逻辑
			result.Matched = matchedCount > 0
			if result.Matched {
				result.Reason = fmt.Sprintf("%d/%d 个条件匹配成功 (OR逻辑)", matchedCount, result.TotalCount)
			} else {
				result.Reason = "没有任何条件匹配成功"
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// MatchConditions 执行条件匹配
func (s *enhancedRuleEngineService) MatchConditions(emailData *model.EmailData, conditions []*model.MatchCondition) ([]*ConditionMatchResult, error) {
	var results []*ConditionMatchResult

	for _, condition := range conditions {
		if condition.Status != "active" {
			continue
		}

		matched, reason, err := s.MatchSingleCondition(emailData, condition)
		if err != nil {
			return nil, fmt.Errorf("匹配条件 %d 失败: %v", condition.ID, err)
		}

		result := &ConditionMatchResult{
			Condition: condition,
			Matched:   matched,
			Reason:    reason,
		}

		// 提取字段内容用于调试
		emailFields := s.ExtractEmailFields(emailData)
		if fieldContent, exists := emailFields[condition.FieldType]; exists {
			result.FieldContent = fieldContent
		}

		results = append(results, result)
	}

	return results, nil
}

// MatchSingleCondition 匹配单个条件
func (s *enhancedRuleEngineService) MatchSingleCondition(emailData *model.EmailData, condition *model.MatchCondition) (bool, string, error) {
	// 提取邮件字段内容
	emailFields := s.ExtractEmailFields(emailData)

	// 获取要匹配的字段内容
	fieldContent, exists := emailFields[condition.FieldType]
	if !exists {
		return false, fmt.Sprintf("不支持的字段类型: %s", condition.FieldType), nil
	}

	// 解析关键词列表
	keywords := strings.Split(condition.Keywords, ",")
	var trimmedKeywords []string
	for _, keyword := range keywords {
		if trimmed := strings.TrimSpace(keyword); trimmed != "" {
			trimmedKeywords = append(trimmedKeywords, trimmed)
		}
	}

	if len(trimmedKeywords) == 0 {
		return false, "没有有效的关键词", nil
	}

	// 执行关键词匹配
	var matchedKeywords []string
	for _, keyword := range trimmedKeywords {
		matched, err := s.MatchKeywordWithType(fieldContent, keyword, condition.MatchType)
		if err != nil {
			return false, fmt.Sprintf("匹配关键词 '%s' 失败: %v", keyword, err), err
		}
		if matched {
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}

	// 根据关键词逻辑判断最终结果
	keywordMatched := false
	var reason string

	if condition.KeywordLogic == "and" {
		keywordMatched = len(matchedKeywords) == len(trimmedKeywords)
		if keywordMatched {
			reason = fmt.Sprintf("所有关键词都匹配成功: [%s]", strings.Join(matchedKeywords, ", "))
		} else {
			reason = fmt.Sprintf("只有部分关键词匹配: [%s]，AND逻辑要求全部匹配", strings.Join(matchedKeywords, ", "))
		}
	} else { // or逻辑
		keywordMatched = len(matchedKeywords) > 0
		if keywordMatched {
			reason = fmt.Sprintf("匹配成功的关键词: [%s]", strings.Join(matchedKeywords, ", "))
		} else {
			reason = "没有任何关键词匹配成功"
		}
	}

	return keywordMatched, reason, nil
}

// MatchKeywordWithType 根据匹配类型执行关键词匹配
func (s *enhancedRuleEngineService) MatchKeywordWithType(content, keyword, matchType string) (bool, error) {
	switch matchType {
	case "equals":
		// 完全匹配
		return content == keyword, nil

	case "contains":
		// 包含匹配（不区分大小写）
		return strings.Contains(strings.ToLower(content), strings.ToLower(keyword)), nil

	case "startsWith":
		// 前缀匹配（不区分大小写）
		return strings.HasPrefix(strings.ToLower(content), strings.ToLower(keyword)), nil

	case "endsWith":
		// 后缀匹配（不区分大小写）
		return strings.HasSuffix(strings.ToLower(content), strings.ToLower(keyword)), nil

	case "regex":
		// 正则表达式匹配
		regex, err := regexp.Compile(keyword)
		if err != nil {
			return false, fmt.Errorf("正则表达式编译失败: %v", err)
		}
		return regex.MatchString(content), nil

	case "notContains":
		// 不包含匹配（不区分大小写）
		return !strings.Contains(strings.ToLower(content), strings.ToLower(keyword)), nil

	default:
		return false, fmt.Errorf("不支持的匹配类型: %s", matchType)
	}
}

// ExtractEmailFields 提取邮件字段内容
func (s *enhancedRuleEngineService) ExtractEmailFields(emailData *model.EmailData) map[string]string {
	fields := map[string]string{
		"subject": emailData.Subject,
		"from":    emailData.Sender,
		"body":    emailData.Content,
	}

	// 处理多值字段
	if len(emailData.To) > 0 {
		fields["to"] = strings.Join(emailData.To, ", ")
	}

	if len(emailData.CC) > 0 {
		fields["cc"] = strings.Join(emailData.CC, ", ")
	}

	if len(emailData.AttachmentNames) > 0 {
		fields["attachment_name"] = strings.Join(emailData.AttachmentNames, ", ")
	}

	return fields
}

// CheckDuplicateByRuleGroup 检查规则组重复告警
func (s *enhancedRuleEngineService) CheckDuplicateByRuleGroup(emailData *model.EmailData, ruleGroupID uint) (bool, error) {
	// 基于MessageID检查是否已存在告警（可以扩展为基于规则组的更精确检查）
	return s.alertRepo.ExistsByMessageID(emailData.MessageID)
}

// CreateAlertFromRuleGroup 从规则组创建告警
func (s *enhancedRuleEngineService) CreateAlertFromRuleGroup(emailData *model.EmailData, ruleGroup *model.RuleGroup) (*model.Alert, error) {
	alert := &model.Alert{
		MailboxID:    ruleGroup.MailboxID,
		RuleID:       0,            // 旧架构字段，保持为0
		RuleGroupID:  ruleGroup.ID, // 新架构字段，关联规则组
		Subject:      emailData.Subject,
		Sender:       emailData.Sender,
		Content:      emailData.Content,
		MessageID:    emailData.MessageID,
		ReceivedAt:   emailData.ReceivedAt,
		Status:       "pending",
		SentChannels: "",
		ErrorMsg:     "",
		RetryCount:   0,
	}

	err := s.alertRepo.Create(alert)
	if err != nil {
		return nil, err
	}

	return alert, nil
}

// GetEnhancedRuleEngineStats 获取增强版规则引擎统计信息
func (s *enhancedRuleEngineService) GetEnhancedRuleEngineStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取激活规则组数量
	activeRuleGroups, err := s.ruleGroupRepo.GetActiveRuleGroups()
	if err != nil {
		return nil, err
	}
	stats["active_rule_groups"] = len(activeRuleGroups)

	// 获取激活条件数量
	activeConditions, err := s.conditionRepo.GetActiveConditions()
	if err != nil {
		return nil, err
	}
	stats["active_conditions"] = len(activeConditions)

	// 获取今日告警统计
	todayStats, err := s.alertRepo.GetTodayStats()
	if err != nil {
		return nil, err
	}

	// 合并统计信息
	for key, value := range todayStats {
		stats[key] = value
	}

	return stats, nil
}
