package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Mailbox 邮箱配置模型
type Mailbox struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`                                       // 邮箱名称
	Email       string `gorm:"size:255;not null;uniqueIndex:idx_mailbox_email_active" json:"email"` // 邮箱地址
	Host        string `gorm:"size:255;not null" json:"host"`                                       // IMAP/POP3服务器地址
	Port        int    `gorm:"not null" json:"port"`                                                // 端口
	Username    string `gorm:"size:255;not null" json:"username"`                                   // 用户名
	Password    string `gorm:"size:255;not null" json:"-"`                                          // 密码（不返回给前端）
	Protocol    string `gorm:"size:10;not null" json:"protocol"`                                    // 协议类型：IMAP/POP3
	SSL         bool   `gorm:"default:true" json:"ssl"`                                             // 是否启用SSL
	Status      string `gorm:"size:20;default:'active'" json:"status"`                              // 状态：active/inactive
	Description string `gorm:"type:text" json:"description"`                                        // 描述
}

// MailboxWithPassword 邮箱配置模型（包含密码，用于编辑）
type MailboxWithPassword struct {
	BaseModel
	Name        string `json:"name"`        // 邮箱名称
	Email       string `json:"email"`       // 邮箱地址
	Host        string `json:"host"`        // IMAP/POP3服务器地址
	Port        int    `json:"port"`        // 端口
	Username    string `json:"username"`    // 用户名
	Password    string `json:"password"`    // 密码（明文返回）
	Protocol    string `json:"protocol"`    // 协议类型：IMAP/POP3
	SSL         bool   `json:"ssl"`         // 是否启用SSL
	Status      string `json:"status"`      // 状态：active/inactive
	Description string `json:"description"` // 描述
}

// AlertRule 告警规则模型
type AlertRule struct {
	BaseModel
	Name        string  `gorm:"size:100;not null" json:"name"`          // 规则名称
	MailboxID   uint    `gorm:"not null" json:"mailbox_id"`             // 关联邮箱ID
	Mailbox     Mailbox `gorm:"foreignKey:MailboxID" json:"mailbox"`    // 关联邮箱
	Priority    int     `gorm:"default:1" json:"priority"`              // 优先级
	MatchType   string  `gorm:"size:20;not null" json:"match_type"`     // 匹配类型：keyword/regex/sender
	MatchValue  string  `gorm:"type:text;not null" json:"match_value"`  // 匹配值
	FieldType   string  `gorm:"size:20;not null" json:"field_type"`     // 匹配字段：subject/body/sender/all
	Status      string  `gorm:"size:20;default:'active'" json:"status"` // 状态：active/inactive
	Description string  `gorm:"type:text" json:"description"`           // 描述
}

// RuleGroup 规则组模型 - 新增，支持多规则组合
type RuleGroup struct {
	BaseModel
	Name        string           `gorm:"size:100;not null" json:"name"`            // 规则组名称
	MailboxID   uint             `gorm:"not null" json:"mailbox_id"`               // 关联邮箱ID
	Mailbox     Mailbox          `gorm:"foreignKey:MailboxID" json:"mailbox"`      // 关联邮箱
	Logic       string           `gorm:"size:10;default:'and'" json:"logic"`       // 规则间逻辑：and/or
	Priority    int              `gorm:"default:1" json:"priority"`                // 优先级 (1-10)
	Status      string           `gorm:"size:20;default:'active'" json:"status"`   // 状态：active/inactive
	Description string           `gorm:"type:text" json:"description"`             // 描述
	Conditions  []MatchCondition `gorm:"foreignKey:RuleGroupID" json:"conditions"` // 关联的匹配条件
	Channels    []Channel        `gorm:"-" json:"channels"`                        // 关联的通知渠道（通过服务层手动加载）
}

// MatchCondition 匹配条件模型 - 新增，支持多维度匹配
type MatchCondition struct {
	BaseModel
	RuleGroupID  uint      `gorm:"not null" json:"rule_group_id"`             // 关联规则组ID
	RuleGroup    RuleGroup `gorm:"foreignKey:RuleGroupID" json:"rule_group"`  // 关联规则组
	FieldType    string    `gorm:"size:20;not null" json:"field_type"`        // 匹配字段：subject/from/to/cc/body/attachment_name
	MatchType    string    `gorm:"size:20;not null" json:"match_type"`        // 匹配类型：equals/contains/startsWith/endsWith/regex/notContains
	Keywords     string    `gorm:"type:text;not null" json:"keywords"`        // 关键词（多个用逗号分隔）
	KeywordLogic string    `gorm:"size:10;default:'or'" json:"keyword_logic"` // 关键词逻辑：and/or
	Priority     int       `gorm:"default:1" json:"priority"`                 // 条件优先级
	Status       string    `gorm:"size:20;default:'active'" json:"status"`    // 状态：active/inactive
	Description  string    `gorm:"type:text" json:"description"`              // 描述
}

// Channel 通知渠道模型
type Channel struct {
	BaseModel
	Name        string     `gorm:"size:100;not null" json:"name"`                   // 渠道名称
	Type        string     `gorm:"size:20;not null" json:"type"`                    // 渠道类型：email/dingtalk/wechat/webhook
	Config      string     `gorm:"type:text;not null" json:"config"`                // 渠道配置（JSON格式）
	Status      string     `gorm:"size:20;default:'active'" json:"status"`          // 状态：active/inactive
	Description string     `gorm:"type:text" json:"description"`                    // 描述
	TestResult  string     `gorm:"type:text" json:"test_result"`                    // 测试结果
	LastTestAt  *time.Time `json:"last_test_at"`                                    // 最后测试时间
	TemplateID  *uint      `gorm:"index" json:"template_id"`                        // 关联的模版ID（可选，为空时使用默认模版）
	Template    *Template  `gorm:"foreignKey:TemplateID" json:"template,omitempty"` // 关联的模版
}

// WeChatConfig 企业微信配置
type WeChatConfig struct {
	Type       string `json:"type"`                  // robot/app - 机器人或应用消息
	WebhookURL string `json:"webhook_url,omitempty"` // 机器人webhook地址
	Key        string `json:"key,omitempty"`         // 机器人key
	CorpID     string `json:"corp_id,omitempty"`     // 企业ID（应用消息用）
	AgentID    int    `json:"agent_id,omitempty"`    // 应用ID（应用消息用）
	Secret     string `json:"secret,omitempty"`      // 应用Secret（应用消息用）
	ToUser     string `json:"to_user,omitempty"`     // 接收人（应用消息用）
	ToParty    string `json:"to_party,omitempty"`    // 接收部门（应用消息用）
	ToTag      string `json:"to_tag,omitempty"`      // 接收标签（应用消息用）
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	Type       string `json:"type"`                  // robot/work - 群机器人或工作通知
	WebhookURL string `json:"webhook_url,omitempty"` // 群机器人webhook地址
	Secret     string `json:"secret,omitempty"`      // 群机器人密钥
	AppKey     string `json:"app_key,omitempty"`     // 应用Key（工作通知用）
	AppSecret  string `json:"app_secret,omitempty"`  // 应用Secret（工作通知用）
	AgentID    int64  `json:"agent_id,omitempty"`    // 应用AgentID（工作通知用）
	UserIDs    string `json:"user_ids,omitempty"`    // 接收用户ID列表（工作通知用）
}

// WebhookConfig 自定义Webhook配置
type WebhookConfig struct {
	URL     string            `json:"url"`     // Webhook地址
	Method  string            `json:"method"`  // 请求方法：POST/PUT
	Headers map[string]string `json:"headers"` // 请求头
	Timeout int               `json:"timeout"` // 超时时间（秒）
}

// EmailConfig 邮件转发配置
type EmailConfig struct {
	Host     string `json:"host"`     // SMTP服务器
	Port     int    `json:"port"`     // 端口
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	SSL      bool   `json:"ssl"`      // 是否启用SSL
	To       string `json:"to"`       // 收件人（多个用逗号分隔）
	CC       string `json:"cc"`       // 抄送（多个用逗号分隔）
	BCC      string `json:"bcc"`      // 密送（多个用逗号分隔）
}

// RuleChannel 规则和渠道的中间表
type RuleChannel struct {
	RuleID    uint `gorm:"primaryKey"`
	ChannelID uint `gorm:"primaryKey"`
	Priority  int  `gorm:"default:1"` // 渠道优先级
	CreatedAt time.Time
}

// RuleGroupChannel 规则组和渠道的中间表
type RuleGroupChannel struct {
	RuleGroupID uint `gorm:"primaryKey"`
	ChannelID   uint `gorm:"primaryKey"`
	Priority    int  `gorm:"default:1"` // 渠道优先级
	CreatedAt   time.Time
}

// NotificationLog 通知发送日志
type NotificationLog struct {
	BaseModel
	ChannelID    uint       `gorm:"not null" json:"channel_id"` // 渠道ID
	Channel      Channel    `gorm:"foreignKey:ChannelID" json:"channel"`
	AlertID      uint       `gorm:"not null" json:"alert_id"` // 告警ID
	Alert        Alert      `gorm:"foreignKey:AlertID" json:"alert"`
	Content      string     `gorm:"type:longtext" json:"content"`   // 发送内容
	Status       string     `gorm:"size:20;not null" json:"status"` // 发送状态：pending/success/failed
	ErrorMsg     string     `gorm:"type:text" json:"error_msg"`     // 错误信息
	ResponseData string     `gorm:"type:text" json:"response_data"` // 响应数据
	SentAt       *time.Time `json:"sent_at"`                        // 发送时间
	RetryCount   int        `gorm:"default:0" json:"retry_count"`   // 重试次数
}

// Template 消息模版模型
type Template struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`          // 模版名称
	Type        string `gorm:"size:20;not null" json:"type"`           // 模版类型：email/dingtalk/wechat/markdown
	Subject     string `gorm:"size:255" json:"subject"`                // 主题模版（邮件用）
	Content     string `gorm:"type:text;not null" json:"content"`      // 内容模版
	Variables   string `gorm:"type:text" json:"variables"`             // 可用变量说明（JSON格式）
	IsDefault   bool   `gorm:"default:false" json:"is_default"`        // 是否为默认模版
	Status      string `gorm:"size:20;default:'active'" json:"status"` // 状态：active/inactive
	Description string `gorm:"type:text" json:"description"`           // 描述
}

// Alert 告警记录模型
type Alert struct {
	BaseModel
	MailboxID    uint      `gorm:"not null" json:"mailbox_id"`               // 邮箱ID
	Mailbox      Mailbox   `gorm:"foreignKey:MailboxID" json:"mailbox"`      // 关联邮箱
	RuleID       uint      `gorm:"default:0" json:"rule_id"`                 // 规则ID（已废弃，兼容用）
	Rule         AlertRule `gorm:"foreignKey:RuleID" json:"rule"`            // 关联规则（已废弃，兼容用）
	RuleGroupID  uint      `gorm:"default:0" json:"rule_group_id"`           // 规则组ID（新架构）
	RuleGroup    RuleGroup `gorm:"foreignKey:RuleGroupID" json:"rule_group"` // 关联规则组（新架构）
	Subject      string    `gorm:"size:500" json:"subject"`                  // 邮件主题
	Sender       string    `gorm:"size:255" json:"sender"`                   // 发件人
	Content      string    `gorm:"type:longtext" json:"content"`             // 邮件内容
	MessageID    string    `gorm:"size:255;index" json:"message_id"`         // 邮件MessageID（用于去重）
	ReceivedAt   time.Time `gorm:"not null" json:"received_at"`              // 邮件接收时间
	Status       string    `gorm:"size:20;default:'pending'" json:"status"`  // 处理状态：pending/sent/failed
	SentChannels string    `gorm:"type:text" json:"sent_channels"`           // 已发送的渠道
	ErrorMsg     string    `gorm:"type:text" json:"error_msg"`               // 错误信息
	RetryCount   int       `gorm:"default:0" json:"retry_count"`             // 重试次数
}

// User 用户模型（后续扩展）
type User struct {
	BaseModel
	Username string `gorm:"size:50;not null;unique" json:"username"`
	Email    string `gorm:"size:255;not null;unique" json:"email"`
	Password string `gorm:"size:255;not null" json:"-"`
	Role     string `gorm:"size:20;default:'user'" json:"role"` // user/admin
	Status   string `gorm:"size:20;default:'active'" json:"status"`
}

// EmailData 邮件数据结构（用于规则匹配）- 升级版本
type EmailData struct {
	UID             int       `json:"uid"`
	Subject         string    `json:"subject"`
	Sender          string    `json:"sender"`  // 发件人 (From)
	To              []string  `json:"to"`      // 收件人列表
	CC              []string  `json:"cc"`      // 抄送人列表
	BCC             []string  `json:"bcc"`     // 密送人列表
	Content         string    `json:"content"` // 邮件正文
	HTMLContent     string    `json:"html_content"`
	AttachmentNames []string  `json:"attachment_names"` // 附件名称列表
	ReceivedAt      time.Time `json:"received_at"`
	MessageID       string    `json:"message_id"`
	Size            uint64    `json:"size"`
	Flags           []string  `json:"flags"`
}

// TemplateVariable 模版变量定义
type TemplateVariable struct {
	Name        string `json:"name"`        // 变量名
	Description string `json:"description"` // 变量描述
	Example     string `json:"example"`     // 示例值
	Category    string `json:"category"`    // 变量分类：email/alert/system/time
}

// TemplateRenderData 模版渲染数据
type TemplateRenderData struct {
	Email   *EmailData `json:"email"`   // 邮件数据
	Alert   *Alert     `json:"alert"`   // 告警数据
	Rule    *AlertRule `json:"rule"`    // 规则数据
	Mailbox *Mailbox   `json:"mailbox"` // 邮箱数据
	System  SystemInfo `json:"system"`  // 系统信息
	Time    TimeInfo   `json:"time"`    // 时间信息
}

// SystemInfo 系统信息
type SystemInfo struct {
	AppName     string `json:"app_name"`
	AppVersion  string `json:"app_version"`
	ServerName  string `json:"server_name"`
	Environment string `json:"environment"`
}

// TimeInfo 时间信息
type TimeInfo struct {
	Now       time.Time `json:"now"`
	NowFormat string    `json:"now_format"`
	NowUnix   int64     `json:"now_unix"`
	Today     string    `json:"today"`
	Yesterday string    `json:"yesterday"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	SessionID string `json:"session_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
}
