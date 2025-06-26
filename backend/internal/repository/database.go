package repository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"emailAlert/config"
	"emailAlert/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库连接管理
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg *config.Config) (*Database, error) {
	var db *gorm.DB
	var err error

	// 配置GORM日志级别
	logLevel := logger.Info
	if cfg.Server.Mode == "release" {
		logLevel = logger.Error
	}

	// 创建自定义日志写入器
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logLevel,    // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound错误
			Colorful:                  false,       // 禁用彩色打印
		},
	)

	gormConfig := &gorm.Config{
		Logger: newLogger,
	}

	// 根据数据库类型连接数据库
	switch cfg.Database.Type {
	case "sqlite":
		// 确保数据库目录存在
		dbDir := filepath.Dir(cfg.Database.FilePath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %v", err)
		}

		db, err = gorm.Open(sqlite.Open(cfg.Database.FilePath), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("连接SQLite数据库失败: %v", err)
		}

	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Database,
		)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("连接MySQL数据库失败: %v", err)
		}

	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.Database.Type)
	}

	// 迁移数据库结构
	err = db.AutoMigrate(
		&model.Mailbox{},
		&model.AlertRule{},
		&model.Alert{},
		&model.Channel{},
		&model.Template{},
		&model.RuleChannel{},
		&model.RuleGroupChannel{}, // 新增：规则组渠道关联表
		&model.NotificationLog{},
		&model.RuleGroup{},      // 新增：规则组模型
		&model.MatchCondition{}, // 新增：匹配条件模型
	)
	if err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 添加Channel模型的template_id字段（如果不存在）
	if !db.Migrator().HasColumn(&model.Channel{}, "template_id") {
		err = db.Migrator().AddColumn(&model.Channel{}, "template_id")
		if err != nil {
			log.Printf("添加Channel.template_id字段失败: %v", err)
		} else {
			log.Println("已添加Channel.template_id字段")
		}
	}

	// 添加Alert模型的rule_group_id字段（如果不存在）
	if !db.Migrator().HasColumn(&model.Alert{}, "rule_group_id") {
		err = db.Migrator().AddColumn(&model.Alert{}, "rule_group_id")
		if err != nil {
			log.Printf("添加Alert.rule_group_id字段失败: %v", err)
		} else {
			log.Println("已添加Alert.rule_group_id字段")
		}
	}

	log.Println("数据库迁移完成")

	return &Database{DB: db}, nil
}

// AutoMigrate 自动迁移数据库表结构
func (d *Database) AutoMigrate() error {
	return d.DB.AutoMigrate(
		&model.User{},
		&model.Mailbox{},
		&model.AlertRule{},
		&model.Alert{},
		&model.Template{},
		&model.Channel{},
		&model.RuleChannel{},
		&model.RuleGroupChannel{}, // 新增：规则组渠道关联表
		&model.NotificationLog{},
		&model.RuleGroup{},      // 新增：规则组模型
		&model.MatchCondition{}, // 新增：匹配条件模型
	)
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取数据库连接
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}
