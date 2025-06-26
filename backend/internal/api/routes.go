package api

import (
	"emailAlert/config"
	"emailAlert/internal/middleware"
	"emailAlert/internal/repository"
	"emailAlert/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, cfg *config.Config, db *repository.Database) {
	// 初始化认证服务
	authService := service.NewAuthService("./config/users.json")
	authHandler := NewAuthHandler(authService)

	// 设置中间件
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// 认证路由（不需要认证）
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
		auth.GET("/profile", authHandler.Profile)
	}

	// 用户管理路由（需要认证和管理员权限）
	users := router.Group("/api/v1/users")
	users.Use(middleware.AuthMiddleware(authService))
	{
		users.GET("", authHandler.GetUsersHandler)
		users.POST("", authHandler.CreateUserHandler)
		users.PUT("/:username", authHandler.UpdateUserHandler)
		users.DELETE("/:username", authHandler.DeleteUserHandler)
	}

	// 健康检查
	router.GET("/health", healthCheck)

	// 初始化仓库层
	mailboxRepo := repository.NewMailboxRepository(db.GetDB())
	alertRepo := repository.NewAlertRepository(db.GetDB())
	templateRepo := repository.NewTemplateRepository(db.GetDB())
	channelRepo := repository.NewChannelRepository(db.GetDB())
	ruleChannelRepo := repository.NewRuleChannelRepository(db.GetDB())
	notificationLogRepo := repository.NewNotificationLogRepository(db.GetDB())
	// 规则组和匹配条件仓库
	ruleGroupRepo := repository.NewRuleGroupRepository(db.GetDB())
	matchConditionRepo := repository.NewMatchConditionRepository(db.GetDB())
	ruleGroupChannelRepo := repository.NewRuleGroupChannelRepository(db.GetDB())

	// 初始化基础服务层
	mailboxService := service.NewMailboxService(mailboxRepo)
	templateService := service.NewTemplateService(templateRepo)
	channelService := service.NewChannelService(channelRepo)
	// 规则组服务的初始化
	ruleGroupService := service.NewRuleGroupService(ruleGroupRepo, matchConditionRepo, ruleGroupChannelRepo, mailboxRepo)
	// 增强版规则引擎初始化
	enhancedRuleEngineService := service.NewEnhancedRuleEngineService(ruleGroupRepo, matchConditionRepo, *alertRepo)

	// 初始化通知分发服务
	notificationDispatcherService := service.NewNotificationDispatcherService(
		ruleChannelRepo,
		ruleGroupChannelRepo,
		notificationLogRepo,
		alertRepo,
		channelService,
		templateService,
	)

	// 初始化告警服务（传入通知分发服务以支持重试功能）
	alertService := service.NewAlertService(alertRepo, notificationDispatcherService)

	// 初始化邮件监控服务（使用新的规则引擎）
	emailMonitorService := service.NewEmailMonitorService(
		mailboxRepo,
		alertRepo,
		enhancedRuleEngineService,
		notificationDispatcherService,
	)

	// 初始化处理器
	mailboxHandler := NewMailboxHandler(mailboxService)
	emailMonitorHandler := NewEmailMonitorHandler(emailMonitorService)
	templateHandler := NewTemplateHandler(templateService)
	channelHandler := NewChannelHandler(channelService)
	alertHandler := NewAlertHandler(alertService)
	// 规则组处理器
	ruleGroupHandler := NewRuleGroupHandler(ruleGroupService, mailboxService, channelService)
	// 通知日志处理器
	notificationLogHandler := NewNotificationLogHandler(notificationLogRepo, alertRepo, channelRepo)

	// API版本分组（需要认证）
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(authService))
	{
		// 邮箱管理路由 - 重新组织确保路由优先级正确
		mailboxes := v1.Group("/mailboxes")
		{
			// 基础CRUD路由
			mailboxes.GET("", mailboxHandler.GetMailboxes)
			mailboxes.POST("", mailboxHandler.CreateMailbox)

			// 特殊功能路由（必须在参数化路由之前）
			mailboxes.POST("/config-test", mailboxHandler.TestMailboxConfig)         // 测试邮箱配置
			mailboxes.POST("/diagnose-config", mailboxHandler.DiagnoseMailboxConfig) // 诊断邮箱配置

			// 参数化路由（放在最后）
			mailboxes.GET("/:id", mailboxHandler.GetMailbox)
			mailboxes.GET("/:id/edit", mailboxHandler.GetMailboxWithPassword) // 获取包含密码的邮箱信息（用于编辑）
			mailboxes.PUT("/:id", mailboxHandler.UpdateMailbox)
			mailboxes.DELETE("/:id", mailboxHandler.DeleteMailbox)
			mailboxes.PUT("/:id/status", mailboxHandler.UpdateMailboxStatus)
			mailboxes.POST("/:id/test", mailboxHandler.TestMailbox)         // 测试现有邮箱连接
			mailboxes.POST("/:id/diagnose", mailboxHandler.DiagnoseMailbox) // 诊断现有邮箱
		}

		// 邮件监控路由
		emailMonitorHandler.RegisterRoutes(v1)

		// 规则组路由（替代旧的告警规则路由）
		ruleGroups := v1.Group("/rule-groups")
		{
			// 基础CRUD路由
			ruleGroups.GET("", ruleGroupHandler.GetRuleGroups)
			ruleGroups.POST("", ruleGroupHandler.CreateRuleGroup)
			ruleGroups.GET("/:id", ruleGroupHandler.GetRuleGroup)
			ruleGroups.PUT("/:id", ruleGroupHandler.UpdateRuleGroup)
			ruleGroups.PUT("/:id/status", ruleGroupHandler.UpdateRuleGroupStatus)
			ruleGroups.DELETE("/:id", ruleGroupHandler.DeleteRuleGroup)

			// 扩展功能路由
			ruleGroups.POST("/with-conditions", ruleGroupHandler.CreateRuleGroupWithConditions)
			ruleGroups.GET("/:id/with-conditions", ruleGroupHandler.GetRuleGroupWithConditions)
			ruleGroups.PUT("/:id/with-conditions", ruleGroupHandler.UpdateRuleGroupWithConditions)
			ruleGroups.POST("/test", ruleGroupHandler.TestRuleGroup)

			// 选项接口
			ruleGroups.GET("/mailbox-options", ruleGroupHandler.GetMailboxOptions)
			ruleGroups.GET("/match-type-options", ruleGroupHandler.GetMatchTypeOptions)
			ruleGroups.GET("/field-type-options", ruleGroupHandler.GetFieldTypeOptions)
			ruleGroups.GET("/channel-options", ruleGroupHandler.GetChannelOptions)
		}

		// 告警历史路由
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", alertHandler.GetAlerts)
			alerts.GET("/:id", alertHandler.GetAlert)
			alerts.PUT("/:id/status", alertHandler.UpdateAlertStatus)
			alerts.DELETE("/:id", alertHandler.DeleteAlert)
			alerts.POST("/:id/retry", alertHandler.RetryAlert)
			alerts.GET("/stats", alertHandler.GetAlertStats)
			alerts.GET("/trends", alertHandler.GetAlertTrends)
			alerts.POST("/batch-update", alertHandler.BatchUpdateAlerts)
		}

		// 通知渠道路由
		// 注意：特殊路由必须在参数化路由之前注册
		v1.GET("/channels", channelHandler.GetChannels)
		v1.POST("/channels", channelHandler.CreateChannel)
		v1.GET("/channels/types", channelHandler.GetChannelTypes)
		v1.POST("/channels/config-test", channelHandler.TestChannelConfig)

		// 参数化路由放在最后
		v1.GET("/channels/:id", channelHandler.GetChannel)
		v1.PUT("/channels/:id", channelHandler.UpdateChannel)
		v1.DELETE("/channels/:id", channelHandler.DeleteChannel)
		v1.POST("/channels/:id/test", channelHandler.TestChannel)
		v1.PUT("/channels/:id/status", channelHandler.UpdateChannelStatus)
		v1.POST("/channels/:id/send", channelHandler.SendNotification)

		// 模版管理路由
		templates := v1.Group("/templates")
		{
			// 基础CRUD路由
			templates.GET("", templateHandler.GetTemplates)
			templates.POST("", templateHandler.CreateTemplate)

			// 特殊功能路由（必须在参数化路由之前）
			templates.POST("/preview", templateHandler.PreviewTemplate)
			templates.GET("/variables", templateHandler.GetAvailableVariables)
			templates.GET("/type/:type", templateHandler.GetTemplatesByType)
			templates.GET("/default/:type", templateHandler.GetDefaultTemplate)

			// 参数化路由（放在最后）
			templates.GET("/:id", templateHandler.GetTemplate)
			templates.PUT("/:id", templateHandler.UpdateTemplate)
			templates.DELETE("/:id", templateHandler.DeleteTemplate)
			templates.PUT("/:id/default", templateHandler.SetDefaultTemplate)
			templates.POST("/:id/render", templateHandler.RenderTemplate)
		}

		// 通知日志路由
		notificationLogs := v1.Group("/notification-logs")
		{
			notificationLogs.GET("", notificationLogHandler.GetNotificationLogs)
			notificationLogs.GET("/:id", notificationLogHandler.GetNotificationLog)
			notificationLogs.DELETE("/:id", notificationLogHandler.DeleteNotificationLog)
			notificationLogs.POST("/:id/retry", notificationLogHandler.RetryNotificationLog)
			notificationLogs.GET("/stats", notificationLogHandler.GetNotificationLogStats)
		}

		// 系统状态路由
		// 初始化系统状态服务和处理器
		systemStatusService := service.NewSystemStatusService(
			db.GetDB(),
			emailMonitorService,
			notificationDispatcherService,
			mailboxService,
			channelService,
			alertService,
		)
		systemStatusHandler := NewSystemStatusHandler(systemStatusService)

		system := v1.Group("/system")
		{
			system.GET("/status", systemStatusHandler.GetSystemStatus)
			system.GET("/stats", systemStatusHandler.GetSystemStats)
			system.GET("/health", systemStatusHandler.GetSystemHealth)
			system.POST("/cleanup", systemStatusHandler.CleanupHistoryData)
		}
	}
}

// healthCheck 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "服务运行正常",
		"data": gin.H{
			"status":  "healthy",
			"service": "emailAlert",
		},
	})
}

// 系统状态处理函数（临时实现）
func getSystemStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取系统状态成功",
		"data": gin.H{
			"status":    "running",
			"uptime":    "24h",
			"version":   "1.0.0",
			"mailboxes": 0,
			"rules":     0,
			"alerts":    0,
		},
	})
}

func getSystemStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取系统统计成功",
		"data": gin.H{
			"total_alerts": 0,
			"today_alerts": 0,
			"success_rate": "100%",
			"channels": gin.H{
				"email":    0,
				"dingtalk": 0,
				"wechat":   0,
			},
		},
	})
}
