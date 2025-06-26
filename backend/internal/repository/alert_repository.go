package repository

import (
	"emailAlert/internal/model"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AlertRepository 告警记录仓储层
type AlertRepository struct {
	db *gorm.DB
}

// NewAlertRepository 创建告警记录仓储实例
func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Create 创建告警记录
func (r *AlertRepository) Create(alert *model.Alert) error {
	return r.db.Create(alert).Error
}

// GetByID 根据ID获取告警记录
func (r *AlertRepository) GetByID(id uint) (*model.Alert, error) {
	var alert model.Alert
	err := r.db.Preload("Mailbox").Preload("Rule").Preload("RuleGroup").First(&alert, id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// List 获取告警记录列表
func (r *AlertRepository) List(page, pageSize int, status string, mailboxID uint) ([]model.Alert, int64, error) {
	var alerts []model.Alert
	var total int64

	query := r.db.Model(&model.Alert{})

	// 添加过滤条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if mailboxID > 0 {
		query = query.Where("mailbox_id = ?", mailboxID)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Preload("Mailbox").Preload("Rule").Preload("RuleGroup").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&alerts).Error
	if err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// ListWithFilters 获取告警记录列表（支持更多过滤条件）
func (r *AlertRepository) ListWithFilters(page, pageSize int, filters map[string]interface{}) ([]model.Alert, int64, error) {
	var alerts []model.Alert
	var total int64

	query := r.db.Model(&model.Alert{})

	// 添加过滤条件
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if mailboxID, ok := filters["mailbox_id"]; ok && mailboxID != nil {
		query = query.Where("mailbox_id = ?", mailboxID)
	}
	if ruleID, ok := filters["rule_id"]; ok && ruleID != nil {
		query = query.Where("rule_id = ?", ruleID)
	}
	if ruleGroupID, ok := filters["rule_group_id"]; ok && ruleGroupID != nil {
		query = query.Where("rule_group_id = ?", ruleGroupID)
	}
	if subject, ok := filters["subject"]; ok && subject != "" {
		query = query.Where("subject LIKE ?", "%"+subject.(string)+"%")
	}
	if sender, ok := filters["sender"]; ok && sender != "" {
		query = query.Where("sender LIKE ?", "%"+sender.(string)+"%")
	}
	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query = query.Where("created_at >= ?", startDate.(string)+" 00:00:00")
	}
	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query = query.Where("created_at <= ?", endDate.(string)+" 23:59:59")
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 排序
	sortBy := "created_at"
	sortOrder := "desc"
	if sort, ok := filters["sort_by"]; ok && sort != "" {
		sortBy = sort.(string)
	}
	if order, ok := filters["sort_order"]; ok && order != "" {
		sortOrder = order.(string)
	}

	orderClause := sortBy + " " + sortOrder

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Preload("Mailbox").Preload("Rule").Preload("RuleGroup").
		Order(orderClause).
		Offset(offset).Limit(pageSize).
		Find(&alerts).Error
	if err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// Update 更新告警记录
func (r *AlertRepository) Update(id uint, updates *model.Alert) error {
	return r.db.Model(&model.Alert{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除告警记录
func (r *AlertRepository) Delete(id uint) error {
	return r.db.Delete(&model.Alert{}, id).Error
}

// UpdateStatus 更新告警状态
func (r *AlertRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Alert{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateStatusWithDetails 更新告警状态及相关详细信息
func (r *AlertRepository) UpdateStatusWithDetails(id uint, status, sentChannels, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if sentChannels != "" {
		updates["sent_channels"] = sentChannels
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	return r.db.Model(&model.Alert{}).Where("id = ?", id).Updates(updates).Error
}

// ExistsByMessageID 检查指定MessageID的邮件是否已存在
func (r *AlertRepository) ExistsByMessageID(messageID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Alert{}).Where("message_id = ?", messageID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByCompositeKey 当MessageID为空时，使用组合字段检查邮件是否已存在
func (r *AlertRepository) ExistsByCompositeKey(mailboxID uint, uid int, subject, sender string, receivedAt time.Time) (bool, error) {
	var count int64

	// 使用邮箱ID、UID、主题、发件人和接收时间的组合来判断是否重复
	// 由于时间可能有微小差异，我们使用小时级别的精度
	timeThreshold := receivedAt.Truncate(time.Hour)

	err := r.db.Model(&model.Alert{}).
		Where("mailbox_id = ? AND uid = ? AND subject = ? AND sender = ? AND created_at >= ? AND created_at < ?",
			mailboxID, uid, subject, sender, timeThreshold, timeThreshold.Add(time.Hour)).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByMessageID 根据MessageID获取告警记录
func (r *AlertRepository) GetByMessageID(messageID string) (*model.Alert, error) {
	var alert model.Alert
	err := r.db.Preload("Mailbox").Preload("Rule").Preload("RuleGroup").
		Where("message_id = ?", messageID).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// GetStatsByDateRange 获取指定日期范围内的统计信息
func (r *AlertRepository) GetStatsByDateRange(startDate, endDate string) (map[string]interface{}, error) {
	var totalCount int64
	var pendingCount int64
	var sentCount int64
	var failedCount int64

	// 总数统计
	err := r.db.Model(&model.Alert{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}

	// 待处理数统计
	err = r.db.Model(&model.Alert{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, "pending").
		Count(&pendingCount).Error
	if err != nil {
		return nil, err
	}

	// 已发送数统计
	err = r.db.Model(&model.Alert{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, "sent").
		Count(&sentCount).Error
	if err != nil {
		return nil, err
	}

	// 失败数统计
	err = r.db.Model(&model.Alert{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, "failed").
		Count(&failedCount).Error
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_count":   totalCount,
		"pending_count": pendingCount,
		"sent_count":    sentCount,
		"failed_count":  failedCount,
		"success_rate":  0.0,
	}

	// 计算成功率
	if totalCount > 0 {
		stats["success_rate"] = float64(sentCount) / float64(totalCount) * 100
	}

	return stats, nil
}

// GetTodayStats 获取今日统计信息
func (r *AlertRepository) GetTodayStats() (map[string]interface{}, error) {
	today := fmt.Sprintf("%s 00:00:00", time.Now().Format("2006-01-02"))
	tomorrow := fmt.Sprintf("%s 00:00:00", time.Now().AddDate(0, 0, 1).Format("2006-01-02"))

	return r.GetStatsByDateRange(today, tomorrow)
}

// GetAlertsByMailbox 获取指定邮箱的告警记录
func (r *AlertRepository) GetAlertsByMailbox(mailboxID uint, limit int) ([]model.Alert, error) {
	var alerts []model.Alert

	query := r.db.Where("mailbox_id = ?", mailboxID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Preload("Mailbox").Preload("Rule").Preload("RuleGroup").Find(&alerts).Error
	if err != nil {
		return nil, err
	}

	return alerts, nil
}

// BatchUpdateStatus 批量更新状态
func (r *AlertRepository) BatchUpdateStatus(ids []uint, status string) error {
	return r.db.Model(&model.Alert{}).Where("id IN ?", ids).Update("status", status).Error
}

// GetPendingAlerts 获取待处理的告警记录
func (r *AlertRepository) GetPendingAlerts(limit int) ([]model.Alert, error) {
	var alerts []model.Alert

	query := r.db.Where("status = ?", "pending").
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Preload("Mailbox").Preload("Rule").Preload("RuleGroup").Find(&alerts).Error
	if err != nil {
		return nil, err
	}

	return alerts, nil
}
