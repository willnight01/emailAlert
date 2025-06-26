package repository

import (
	"emailAlert/internal/model"
	"time"

	"gorm.io/gorm"
)

// NotificationLogRepository 通知日志仓库接口
type NotificationLogRepository interface {
	Create(log *model.NotificationLog) error
	GetByID(id uint) (*model.NotificationLog, error)
	GetByIDWithDetails(id uint) (*model.NotificationLog, error)
	GetByAlertID(alertID uint) ([]*model.NotificationLog, error)
	GetByChannelID(channelID uint, page, size int) ([]*model.NotificationLog, int64, error)
	GetNotificationLogsWithDetails(conditions map[string]interface{}, offset, size int, startTime, endTime *time.Time) ([]*model.NotificationLog, int64, error)
	GetFailedLogs(maxRetryCount int) ([]*model.NotificationLog, error)
	UpdateStatus(id uint, status, errorMsg, responseData string) error
	UpdateSentAt(id uint, sentAt time.Time) error
	IncrementRetryCount(id uint) error
	Update(log *model.NotificationLog) error
	UpdateContent(id uint, content string) error
	Delete(id uint) error
	GetStats(startTime, endTime time.Time) (map[string]interface{}, error)
	GetStatistics(startTime, endTime time.Time) (map[string]interface{}, error)
	GetTodayStats() (map[string]interface{}, error)
}

// notificationLogRepository 通知日志仓库实现
type notificationLogRepository struct {
	db *gorm.DB
}

// NewNotificationLogRepository 创建通知日志仓库
func NewNotificationLogRepository(db *gorm.DB) NotificationLogRepository {
	return &notificationLogRepository{db: db}
}

// Create 创建通知日志
func (r *notificationLogRepository) Create(log *model.NotificationLog) error {
	return r.db.Create(log).Error
}

// GetByID 根据ID获取通知日志
func (r *notificationLogRepository) GetByID(id uint) (*model.NotificationLog, error) {
	var log model.NotificationLog
	err := r.db.Preload("Channel").Preload("Alert").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByIDWithDetails 根据ID获取通知日志及其详细信息
func (r *notificationLogRepository) GetByIDWithDetails(id uint) (*model.NotificationLog, error) {
	var log model.NotificationLog
	err := r.db.Preload("Channel").Preload("Alert").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByAlertID 根据告警ID获取通知日志列表
func (r *notificationLogRepository) GetByAlertID(alertID uint) ([]*model.NotificationLog, error) {
	var logs []*model.NotificationLog
	err := r.db.Preload("Channel").Where("alert_id = ?", alertID).
		Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// GetByChannelID 根据渠道ID获取通知日志列表（分页）
func (r *notificationLogRepository) GetByChannelID(channelID uint, page, size int) ([]*model.NotificationLog, int64, error) {
	var logs []*model.NotificationLog
	var total int64

	query := r.db.Model(&model.NotificationLog{}).Where("channel_id = ?", channelID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err := query.Preload("Alert").Offset(offset).Limit(size).
		Order("created_at DESC").Find(&logs).Error

	return logs, total, err
}

// GetNotificationLogsWithDetails 获取满足条件的通知日志及其详细信息
func (r *notificationLogRepository) GetNotificationLogsWithDetails(conditions map[string]interface{}, offset, size int, startTime, endTime *time.Time) ([]*model.NotificationLog, int64, error) {
	var logs []*model.NotificationLog
	var total int64

	query := r.db.Model(&model.NotificationLog{})

	// 构建查询条件
	for key, value := range conditions {
		if key == "content_like" {
			query = query.Where("content LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	// 添加时间范围条件
	if startTime != nil && endTime != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startTime, *endTime)
	} else if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	} else if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Preload("Channel").Preload("Alert").Offset(offset).Limit(size).
		Order("created_at DESC").Find(&logs).Error

	return logs, total, err
}

// GetFailedLogs 获取失败的通知日志（用于重试）
func (r *notificationLogRepository) GetFailedLogs(maxRetryCount int) ([]*model.NotificationLog, error) {
	var logs []*model.NotificationLog
	err := r.db.Preload("Channel").Preload("Alert").
		Where("status = ? AND retry_count < ?", "failed", maxRetryCount).
		Order("created_at ASC").Find(&logs).Error
	return logs, err
}

// UpdateStatus 更新通知状态
func (r *notificationLogRepository) UpdateStatus(id uint, status, errorMsg, responseData string) error {
	updates := map[string]interface{}{
		"status":        status,
		"error_msg":     errorMsg,
		"response_data": responseData,
	}

	if status == "success" {
		updates["sent_at"] = time.Now()
	}

	return r.db.Model(&model.NotificationLog{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateSentAt 更新发送时间
func (r *notificationLogRepository) UpdateSentAt(id uint, sentAt time.Time) error {
	return r.db.Model(&model.NotificationLog{}).Where("id = ?", id).
		Update("sent_at", sentAt).Error
}

// IncrementRetryCount 增加重试次数
func (r *notificationLogRepository) IncrementRetryCount(id uint) error {
	return r.db.Model(&model.NotificationLog{}).Where("id = ?", id).
		Update("retry_count", gorm.Expr("retry_count + 1")).Error
}

// Update 更新通知日志
func (r *notificationLogRepository) Update(log *model.NotificationLog) error {
	return r.db.Save(log).Error
}

// UpdateContent 更新通知内容
func (r *notificationLogRepository) UpdateContent(id uint, content string) error {
	return r.db.Model(&model.NotificationLog{}).Where("id = ?", id).
		Update("content", content).Error
}

// Delete 删除通知日志
func (r *notificationLogRepository) Delete(id uint) error {
	return r.db.Delete(&model.NotificationLog{}, id).Error
}

// GetStats 获取指定时间范围的统计信息
func (r *notificationLogRepository) GetStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总发送量
	var totalCount int64
	if err := r.db.Model(&model.NotificationLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// 成功发送量
	var successCount int64
	if err := r.db.Model(&model.NotificationLog{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "success").
		Count(&successCount).Error; err != nil {
		return nil, err
	}
	stats["success_count"] = successCount

	// 失败发送量
	var failedCount int64
	if err := r.db.Model(&model.NotificationLog{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "failed").
		Count(&failedCount).Error; err != nil {
		return nil, err
	}
	stats["failed_count"] = failedCount

	// 待处理量
	var pendingCount int64
	if err := r.db.Model(&model.NotificationLog{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "pending").
		Count(&pendingCount).Error; err != nil {
		return nil, err
	}
	stats["pending_count"] = pendingCount

	// 成功率
	if totalCount > 0 {
		stats["success_rate"] = float64(successCount) / float64(totalCount) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	// 按渠道统计
	var channelStats []struct {
		ChannelID uint   `json:"channel_id"`
		Type      string `json:"type"`
		Count     int64  `json:"count"`
		Success   int64  `json:"success"`
	}
	if err := r.db.Table("notification_logs").
		Select("channel_id, channels.type, COUNT(*) as count, SUM(CASE WHEN notification_logs.status = 'success' THEN 1 ELSE 0 END) as success").
		Joins("LEFT JOIN channels ON notification_logs.channel_id = channels.id").
		Where("notification_logs.created_at BETWEEN ? AND ?", startTime, endTime).
		Group("channel_id, channels.type").
		Scan(&channelStats).Error; err != nil {
		return nil, err
	}
	stats["channel_stats"] = channelStats

	return stats, nil
}

// GetStatistics 获取指定时间范围的统计信息
func (r *notificationLogRepository) GetStatistics(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总发送量
	var totalCount int64
	if err := r.db.Model(&model.NotificationLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// 成功发送量
	var successCount int64
	if err := r.db.Model(&model.NotificationLog{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "success").
		Count(&successCount).Error; err != nil {
		return nil, err
	}
	stats["success_count"] = successCount

	// 失败发送量
	var failedCount int64
	if err := r.db.Model(&model.NotificationLog{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "failed").
		Count(&failedCount).Error; err != nil {
		return nil, err
	}
	stats["failed_count"] = failedCount

	// 待处理量
	var pendingCount int64
	if err := r.db.Model(&model.NotificationLog{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startTime, endTime, "pending").
		Count(&pendingCount).Error; err != nil {
		return nil, err
	}
	stats["pending_count"] = pendingCount

	// 成功率
	if totalCount > 0 {
		stats["success_rate"] = float64(successCount) / float64(totalCount) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	// 按渠道统计
	var channelStats []struct {
		ChannelID uint   `json:"channel_id"`
		Type      string `json:"type"`
		Count     int64  `json:"count"`
		Success   int64  `json:"success"`
	}
	if err := r.db.Table("notification_logs").
		Select("channel_id, channels.type, COUNT(*) as count, SUM(CASE WHEN notification_logs.status = 'success' THEN 1 ELSE 0 END) as success").
		Joins("LEFT JOIN channels ON notification_logs.channel_id = channels.id").
		Where("notification_logs.created_at BETWEEN ? AND ?", startTime, endTime).
		Group("channel_id, channels.type").
		Scan(&channelStats).Error; err != nil {
		return nil, err
	}
	stats["channel_stats"] = channelStats

	return stats, nil
}

// GetTodayStats 获取今日统计信息
func (r *notificationLogRepository) GetTodayStats() (map[string]interface{}, error) {
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endTime := startTime.Add(24 * time.Hour)

	return r.GetStats(startTime, endTime)
}
