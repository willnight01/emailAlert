package api

import (
	"emailAlert/internal/model"
	"emailAlert/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RuleGroupHandler 规则组处理器
type RuleGroupHandler struct {
	ruleGroupService service.RuleGroupService
	mailboxService   *service.MailboxService
	channelService   service.ChannelService
}

// NewRuleGroupHandler 创建新的规则组处理器
func NewRuleGroupHandler(ruleGroupService service.RuleGroupService, mailboxService *service.MailboxService, channelService service.ChannelService) *RuleGroupHandler {
	return &RuleGroupHandler{
		ruleGroupService: ruleGroupService,
		mailboxService:   mailboxService,
		channelService:   channelService,
	}
}

// CreateRuleGroup 创建规则组
func (h *RuleGroupHandler) CreateRuleGroup(c *gin.Context) {
	var ruleGroup model.RuleGroup
	if err := c.ShouldBindJSON(&ruleGroup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 设置默认状态
	if ruleGroup.Status == "" {
		ruleGroup.Status = "active"
	}

	if err := h.ruleGroupService.CreateRuleGroup(&ruleGroup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建规则组成功",
		"data":    ruleGroup,
	})
}

// CreateRuleGroupWithConditions 创建规则组及其条件
func (h *RuleGroupHandler) CreateRuleGroupWithConditions(c *gin.Context) {
	var ruleGroupData service.RuleGroupData
	if err := c.ShouldBindJSON(&ruleGroupData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 设置默认状态
	if ruleGroupData.RuleGroup.Status == "" {
		ruleGroupData.RuleGroup.Status = "active"
	}

	// 为条件设置默认状态
	for _, condition := range ruleGroupData.Conditions {
		if condition.Status == "" {
			condition.Status = "active"
		}
	}

	if err := h.ruleGroupService.ProcessRuleGroupWithConditions(&ruleGroupData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建规则组和条件成功",
		"data":    ruleGroupData,
	})
}

// GetRuleGroups 获取规则组列表
func (h *RuleGroupHandler) GetRuleGroups(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// 获取过滤参数
	filters := make(map[string]interface{})
	if mailboxID := c.Query("mailbox_id"); mailboxID != "" {
		if id, err := strconv.ParseUint(mailboxID, 10, 32); err == nil {
			filters["mailbox_id"] = uint(id)
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if logic := c.Query("logic"); logic != "" {
		filters["logic"] = logic
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	ruleGroups, total, err := h.ruleGroupService.GetRuleGroups(page, size, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取规则组列表失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取规则组列表成功",
		"data": gin.H{
			"items": ruleGroups,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// GetRuleGroup 获取规则组详情
func (h *RuleGroupHandler) GetRuleGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则组ID",
			"data":    nil,
		})
		return
	}

	ruleGroup, err := h.ruleGroupService.GetRuleGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "规则组不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取规则组详情成功",
		"data":    ruleGroup,
	})
}

// GetRuleGroupWithConditions 获取规则组及其条件
func (h *RuleGroupHandler) GetRuleGroupWithConditions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则组ID",
			"data":    nil,
		})
		return
	}

	ruleGroup, err := h.ruleGroupService.GetRuleGroupWithConditions(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "规则组不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取规则组详情成功",
		"data":    ruleGroup,
	})
}

// UpdateRuleGroup 更新规则组
func (h *RuleGroupHandler) UpdateRuleGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则组ID: " + c.Param("id"),
			"data":    nil,
		})
		return
	}

	var ruleGroup model.RuleGroup
	if err := c.ShouldBindJSON(&ruleGroup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	ruleGroup.ID = uint(id)
	if err := h.ruleGroupService.UpdateRuleGroup(&ruleGroup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 重新获取更新后的完整规则组数据（包含关联信息）
	updatedRuleGroup, err := h.ruleGroupService.GetRuleGroupByID(uint(id))
	if err != nil {
		// 如果获取失败，仍然返回成功，但数据可能不完整
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "更新规则组成功",
			"data":    ruleGroup,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新规则组成功",
		"data":    updatedRuleGroup,
	})
}

// UpdateRuleGroupStatus 更新规则组状态
func (h *RuleGroupHandler) UpdateRuleGroupStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则组ID: " + c.Param("id"),
			"data":    nil,
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 验证状态值
	validStatuses := []string{"active", "inactive"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的状态值，支持的状态: active, inactive",
			"data":    nil,
		})
		return
	}

	err = h.ruleGroupService.UpdateRuleGroupStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新规则组状态失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新规则组状态成功",
		"data":    nil,
	})
}

// UpdateRuleGroupWithConditions 更新规则组及其条件
func (h *RuleGroupHandler) UpdateRuleGroupWithConditions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则组ID: " + c.Param("id"),
			"data":    nil,
		})
		return
	}

	var ruleGroupData service.RuleGroupData
	if err := c.ShouldBindJSON(&ruleGroupData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	ruleGroupData.RuleGroup.ID = uint(id)

	// 为条件设置默认状态
	for _, condition := range ruleGroupData.Conditions {
		if condition.Status == "" {
			condition.Status = "active"
		}
	}

	if err := h.ruleGroupService.ProcessRuleGroupWithConditions(&ruleGroupData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新规则组和条件成功",
		"data":    ruleGroupData,
	})
}

// DeleteRuleGroup 删除规则组
func (h *RuleGroupHandler) DeleteRuleGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则组ID",
			"data":    nil,
		})
		return
	}

	if err := h.ruleGroupService.DeleteRuleGroup(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除规则组成功",
		"data":    nil,
	})
}

// TestRuleGroup 测试规则组
func (h *RuleGroupHandler) TestRuleGroup(c *gin.Context) {
	var request struct {
		RuleGroupData service.RuleGroupData `json:"rule_group_data"`
		TestEmail     model.EmailData       `json:"test_email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数格式错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// TODO: 实现规则组测试逻辑
	// 这里可以调用增强版规则引擎进行测试

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "测试规则组成功",
		"data": gin.H{
			"matched":         true, // 临时返回值
			"rule_group_data": request.RuleGroupData,
			"test_email":      request.TestEmail,
		},
	})
}

// GetMailboxOptions 获取邮箱选项（用于规则组创建时选择邮箱）
func (h *RuleGroupHandler) GetMailboxOptions(c *gin.Context) {
	result, err := h.mailboxService.List(1, 100, "active")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取邮箱选项失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 转换为选项格式
	options := make([]gin.H, len(result.List))
	for i, mailbox := range result.List {
		options[i] = gin.H{
			"value": mailbox.ID,
			"label": mailbox.Name + " (" + mailbox.Email + ")",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取邮箱选项成功",
		"data":    options,
	})
}

// GetMatchTypeOptions 获取匹配类型选项
func (h *RuleGroupHandler) GetMatchTypeOptions(c *gin.Context) {
	options := []gin.H{
		{"value": "equals", "label": "完全匹配"},
		{"value": "contains", "label": "包含匹配"},
		{"value": "startsWith", "label": "前缀匹配"},
		{"value": "endsWith", "label": "后缀匹配"},
		{"value": "regex", "label": "正则表达式"},
		{"value": "notContains", "label": "不包含"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取匹配类型选项成功",
		"data":    options,
	})
}

// GetFieldTypeOptions 获取字段类型选项
func (h *RuleGroupHandler) GetFieldTypeOptions(c *gin.Context) {
	options := []gin.H{
		{"value": "subject", "label": "邮件主题"},
		{"value": "from", "label": "发件人"},
		{"value": "to", "label": "收件人"},
		{"value": "cc", "label": "抄送人"},
		{"value": "body", "label": "邮件正文"},
		{"value": "attachment_name", "label": "附件名称"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取字段类型选项成功",
		"data":    options,
	})
}

// GetChannelOptions 获取通知渠道选项
func (h *RuleGroupHandler) GetChannelOptions(c *gin.Context) {
	// 获取所有激活的通知渠道
	channels, err := h.channelService.GetActiveChannels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取通知渠道失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 转换为选项格式
	options := make([]gin.H, len(channels))
	for i, channel := range channels {
		options[i] = gin.H{
			"label": channel.Name + " (" + getChannelTypeText(channel.Type) + ")",
			"value": channel.ID,
			"type":  channel.Type,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取通知渠道选项成功",
		"data":    options,
	})
}

// getChannelTypeText 获取渠道类型文本
func getChannelTypeText(channelType string) string {
	switch channelType {
	case "wechat":
		return "企业微信"
	case "dingtalk":
		return "钉钉"
	case "email":
		return "邮件"
	case "webhook":
		return "Webhook"
	default:
		return channelType
	}
}
