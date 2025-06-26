package main

import (
	"context"
	"emailAlert/config"
	"emailAlert/internal/api"
	"emailAlert/internal/repository"
	"emailAlert/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	db, err := repository.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")

	// 初始化默认模版
	templateRepo := repository.NewTemplateRepository(db.GetDB())
	templateInitService := service.NewTemplateInitService(templateRepo)
	if err := templateInitService.InitDefaultTemplates(); err != nil {
		log.Printf("默认模版初始化失败: %v", err)
	} else {
		log.Println("默认模版初始化完成")
	}

	// 初始化通知分发服务（重要：启动后台处理器）
	alertRepo := repository.NewAlertRepository(db.GetDB())
	channelRepo := repository.NewChannelRepository(db.GetDB())
	ruleChannelRepo := repository.NewRuleChannelRepository(db.GetDB())
	ruleGroupChannelRepo := repository.NewRuleGroupChannelRepository(db.GetDB())
	notificationLogRepo := repository.NewNotificationLogRepository(db.GetDB())
	templateService := service.NewTemplateService(templateRepo)
	channelService := service.NewChannelService(channelRepo)

	// 创建通知分发服务
	notificationDispatcherService := service.NewNotificationDispatcherService(
		ruleChannelRepo,
		ruleGroupChannelRepo,
		notificationLogRepo,
		alertRepo,
		channelService,
		templateService,
	)

	// 启动通知分发后台处理器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := notificationDispatcherService.StartBackgroundProcessor(ctx); err != nil {
		log.Printf("启动通知分发后台处理器失败: %v", err)
	} else {
		log.Println("通知分发后台处理器启动成功")
	}

	// 处理已有的待处理告警
	if err := notificationDispatcherService.ProcessPendingAlerts(); err != nil {
		log.Printf("处理待处理告警失败: %v", err)
	} else {
		log.Println("开始处理待处理告警")
	}

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin实例
	router := gin.Default()

	// 设置路由
	api.SetupRoutes(router, cfg, db)

	// 启动服务器
	log.Printf("邮件告警平台启动中，监听端口: %s", cfg.Server.Port)
	log.Printf("数据库文件: %s", cfg.Database.FilePath)

	// 处理优雅关闭
	go func() {
		if err := router.Run(cfg.Server.Host + ":" + cfg.Server.Port); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 取消后台处理器
	cancel()
	log.Println("服务器已关闭")
}
