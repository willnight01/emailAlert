package service

import (
	"emailAlert/internal/model"
	"emailAlert/internal/repository"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// AlertService 告警服务接口
type AlertService interface {
	GetAlerts(page, size int, filters map[string]interface{}) ([]*model.Alert, int64, error)
	GetAlertByID(id uint) (*model.Alert, error)
	CreateAlert(alert *model.Alert) error
	UpdateAlertStatus(id uint, status string) error
	DeleteAlert(id uint) error
	RetryAlert(id uint) error
	BatchUpdateAlerts(ids []uint, status string) error
	GetAlertStats(startDate, endDate string) (map[string]interface{}, error)
	GetAlertTrends(period string) ([]map[string]interface{}, error)
	GetTodayStats() (map[string]interface{}, error)
}

// alertService 告警服务实现
type alertService struct {
	alertRepo                     *repository.AlertRepository
	notificationDispatcherService NotificationDispatcherService
}

// NewAlertService 创建告警服务实例
func NewAlertService(alertRepo *repository.AlertRepository, notificationDispatcherService NotificationDispatcherService) AlertService {
	return &alertService{
		alertRepo:                     alertRepo,
		notificationDispatcherService: notificationDispatcherService,
	}
}

// GetAlerts 获取告警列表
func (s *alertService) GetAlerts(page, size int, filters map[string]interface{}) ([]*model.Alert, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	// 使用新的ListWithFilters方法支持更多过滤条件
	alerts, total, err := s.alertRepo.ListWithFilters(page, size, filters)
	if err != nil {
		return nil, 0, err
	}

	// 转换为指针切片
	var alertPtrs []*model.Alert
	for i := range alerts {
		alertPtrs = append(alertPtrs, &alerts[i])
	}

	return alertPtrs, total, nil
}

// GetAlertByID 根据ID获取告警详情
func (s *alertService) GetAlertByID(id uint) (*model.Alert, error) {
	alert, err := s.alertRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("告警记录不存在")
		}
		return nil, err
	}
	return alert, nil
}

// CreateAlert 创建告警记录
func (s *alertService) CreateAlert(alert *model.Alert) error {
	return s.alertRepo.Create(alert)
}

// UpdateAlertStatus 更新告警状态
func (s *alertService) UpdateAlertStatus(id uint, status string) error {
	// 先验证记录是否存在
	_, err := s.GetAlertByID(id)
	if err != nil {
		return err
	}

	// 更新状态
	err = s.alertRepo.UpdateStatus(id, status)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAlert 删除告警记录
func (s *alertService) DeleteAlert(id uint) error {
	// 先验证记录是否存在
	_, err := s.GetAlertByID(id)
	if err != nil {
		return err
	}

	// 删除记录
	err = s.alertRepo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

// RetryAlert 重试告警发送
func (s *alertService) RetryAlert(id uint) error {
	// 获取告警记录
	alert, err := s.GetAlertByID(id)
	if err != nil {
		return err
	}

	// 验证可重试的状态
	allowedStatuses := []string{"failed", "pending", "sent"}
	isAllowed := false
	for _, status := range allowedStatuses {
		if alert.Status == status {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return fmt.Errorf("状态为 %s 的告警不支持重试", alert.Status)
	}

	// 记录原始状态用于日志
	originalStatus := alert.Status

	// 重置状态为待处理，增加重试次数
	alert.Status = "pending"
	alert.RetryCount = alert.RetryCount + 1
	alert.ErrorMsg = ""

	err = s.alertRepo.Update(id, alert)
	if err != nil {
		return err
	}

	// 立即触发重新发送
	if s.notificationDispatcherService != nil {
		log.Printf("告警 %d 开始重新发送 - RuleGroupID: %d, RuleID: %d", alert.ID, alert.RuleGroupID, alert.RuleID)
		if err := s.notificationDispatcherService.DispatchAlert(alert); err != nil {
			// 如果分发失败，记录错误但不返回失败（状态已更新）
			log.Printf("告警 %d 重试分发失败: %v", alert.ID, err)
			alert.Status = "failed"
			alert.ErrorMsg = fmt.Sprintf("重试分发失败: %v", err)
			s.alertRepo.Update(id, alert)
			return fmt.Errorf("重试发送失败: %v", err)
		}
		log.Printf("告警 %d 重新发送完成", alert.ID)
	} else {
		log.Printf("警告: 通知分发服务未初始化，无法重试告警 %d", alert.ID)
	}

	// 记录重试操作日志
	log.Printf("告警 %d 重试发送 - 原状态: %s, 重试次数: %d", id, originalStatus, alert.RetryCount)

	return nil
}

// BatchUpdateAlerts 批量更新告警状态
func (s *alertService) BatchUpdateAlerts(ids []uint, status string) error {
	if len(ids) == 0 {
		return errors.New("请选择要更新的告警记录")
	}

	err := s.alertRepo.BatchUpdateStatus(ids, status)
	if err != nil {
		return err
	}

	return nil
}

// GetAlertStats 获取告警统计信息
func (s *alertService) GetAlertStats(startDate, endDate string) (map[string]interface{}, error) {
	return s.alertRepo.GetStatsByDateRange(startDate, endDate)
}

// GetAlertTrends 获取告警趋势数据
func (s *alertService) GetAlertTrends(period string) ([]map[string]interface{}, error) {
	var trends []map[string]interface{}

	var days int
	switch period {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	default:
		days = 7
	}

	// 生成时间范围
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	// 按天生成统计数据
	for i := 0; i < days; i++ {
		date := startTime.AddDate(0, 0, i)
		startDate := date.Format("2006-01-02 00:00:00")
		endDate := date.Format("2006-01-02 23:59:59")

		dayStats, err := s.alertRepo.GetStatsByDateRange(startDate, endDate)
		if err != nil {
			return nil, err
		}

		trend := map[string]interface{}{
			"date":          date.Format("2006-01-02"),
			"total_count":   dayStats["total_count"],
			"sent_count":    dayStats["sent_count"],
			"failed_count":  dayStats["failed_count"],
			"pending_count": dayStats["pending_count"],
		}
		trends = append(trends, trend)
	}

	return trends, nil
}

// GetTodayStats 获取今日统计
func (s *alertService) GetTodayStats() (map[string]interface{}, error) {
	return s.alertRepo.GetTodayStats()
}
