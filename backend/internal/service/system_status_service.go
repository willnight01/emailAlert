package service

import (
	"context"
	"emailAlert/internal/model"
	"fmt"
	"runtime"
	"time"

	"gorm.io/gorm"
)

// SystemStatusService 系统状态服务
type SystemStatusService struct {
	db                  *gorm.DB
	emailMonitorService *EmailMonitorService
	notificationService NotificationDispatcherService
	mailboxService      *MailboxService
	channelService      ChannelService
	alertService        AlertService
	startTime           time.Time
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name         string      `json:"name"`
	Status       string      `json:"status"` // healthy, unhealthy, degraded
	Message      string      `json:"message"`
	Details      interface{} `json:"details,omitempty"`
	CheckedAt    time.Time   `json:"checked_at"`
	ResponseTime int64       `json:"response_time"` // 毫秒
}

// SystemHealth 系统健康状态
type SystemHealth struct {
	Status    string          `json:"status"` // healthy, degraded, unhealthy
	Version   string          `json:"version"`
	Uptime    string          `json:"uptime"`
	Services  []ServiceStatus `json:"services"`
	Resources SystemResources `json:"resources"`
	Summary   HealthSummary   `json:"summary"`
	CheckedAt time.Time       `json:"checked_at"`
}

// SystemResources 系统资源使用情况
type SystemResources struct {
	CPU        CPUStats    `json:"cpu"`
	Memory     MemoryStats `json:"memory"`
	Goroutines int         `json:"goroutines"`
	GC         GCStats     `json:"gc"`
}

// CPUStats CPU统计信息
type CPUStats struct {
	Cores        int     `json:"cores"`
	UsagePercent float64 `json:"usage_percent"`
}

// MemoryStats 内存统计信息
type MemoryStats struct {
	Alloc        uint64  `json:"alloc"`       // 当前分配内存
	TotalAlloc   uint64  `json:"total_alloc"` // 总分配内存
	Sys          uint64  `json:"sys"`         // 系统内存
	HeapAlloc    uint64  `json:"heap_alloc"`  // 堆内存
	HeapSys      uint64  `json:"heap_sys"`    // 堆系统内存
	UsagePercent float64 `json:"usage_percent"`
}

// GCStats 垃圾回收统计
type GCStats struct {
	NumGC      uint32    `json:"num_gc"`
	LastGC     time.Time `json:"last_gc"`
	TotalPause uint64    `json:"total_pause"`
	AvgPause   uint64    `json:"avg_pause"`
}

// HealthSummary 健康状况摘要
type HealthSummary struct {
	TotalServices     int `json:"total_services"`
	HealthyServices   int `json:"healthy_services"`
	UnhealthyServices int `json:"unhealthy_services"`
	DegradedServices  int `json:"degraded_services"`
}

// SystemStats 系统统计信息
type SystemStats struct {
	Runtime     RuntimeStats     `json:"runtime"`
	Business    BusinessStats    `json:"business"`
	Performance PerformanceStats `json:"performance"`
	Connections ConnectionStats  `json:"connections"`
}

// RuntimeStats 运行时统计
type RuntimeStats struct {
	Uptime    string    `json:"uptime"`
	StartTime time.Time `json:"start_time"`
	Version   string    `json:"version"`
	GoVersion string    `json:"go_version"`
	Platform  string    `json:"platform"`
}

// BusinessStats 业务统计
type BusinessStats struct {
	TotalMailboxes     int     `json:"total_mailboxes"`
	ActiveMailboxes    int     `json:"active_mailboxes"`
	TotalRules         int     `json:"total_rules"`
	ActiveRules        int     `json:"active_rules"`
	TotalChannels      int     `json:"total_channels"`
	ActiveChannels     int     `json:"active_channels"`
	TotalAlerts        int     `json:"total_alerts"`
	TodayAlerts        int     `json:"today_alerts"`
	PendingAlerts      int     `json:"pending_alerts"`
	TotalNotifications int     `json:"total_notifications"`
	TodayNotifications int     `json:"today_notifications"`
	SuccessRate        float64 `json:"success_rate"`
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	AvgProcessingTime float64 `json:"avg_processing_time"` // 毫秒
	MaxProcessingTime float64 `json:"max_processing_time"` // 毫秒
	RequestsPerSecond float64 `json:"requests_per_second"`
	ErrorRate         float64 `json:"error_rate"`
}

// ConnectionStats 连接统计
type ConnectionStats struct {
	DatabaseConnections int `json:"database_connections"`
	ActiveConnections   int `json:"active_connections"`
	MaxConnections      int `json:"max_connections"`
}

// NewSystemStatusService 创建系统状态服务
func NewSystemStatusService(
	db *gorm.DB,
	emailMonitorService *EmailMonitorService,
	notificationService NotificationDispatcherService,
	mailboxService *MailboxService,
	channelService ChannelService,
	alertService AlertService,
) *SystemStatusService {
	return &SystemStatusService{
		db:                  db,
		emailMonitorService: emailMonitorService,
		notificationService: notificationService,
		mailboxService:      mailboxService,
		channelService:      channelService,
		alertService:        alertService,
		startTime:           time.Now(),
	}
}

// GetSystemHealth 获取系统健康状态
func (s *SystemStatusService) GetSystemHealth() (*SystemHealth, error) {
	checkTime := time.Now()

	// 并发检查各服务状态
	services := make([]ServiceStatus, 0)

	// 检查数据库状态
	if dbStatus := s.checkDatabaseHealth(); dbStatus != nil {
		services = append(services, *dbStatus)
	}

	// 检查邮件监控服务状态
	if monitorStatus := s.checkEmailMonitorHealth(); monitorStatus != nil {
		services = append(services, *monitorStatus)
	}

	// 检查通知服务状态
	if notificationStatus := s.checkNotificationHealth(); notificationStatus != nil {
		services = append(services, *notificationStatus)
	}

	// 检查缓存服务状态（如果有）
	if cacheStatus := s.checkCacheHealth(); cacheStatus != nil {
		services = append(services, *cacheStatus)
	}

	// 计算系统整体状态
	overallStatus := s.calculateOverallStatus(services)

	// 获取系统资源使用情况
	resources := s.getSystemResources()

	// 计算摘要
	summary := s.calculateHealthSummary(services)

	return &SystemHealth{
		Status:    overallStatus,
		Version:   "1.0.0",
		Uptime:    s.getUptime(),
		Services:  services,
		Resources: resources,
		Summary:   summary,
		CheckedAt: checkTime,
	}, nil
}

// GetSystemStats 获取系统统计信息
func (s *SystemStatusService) GetSystemStats() (*SystemStats, error) {
	// 获取运行时信息
	runtimeStats := RuntimeStats{
		Uptime:    s.getUptime(),
		StartTime: s.startTime,
		Version:   "1.0.0",
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	// 获取业务统计
	business, err := s.getBusinessStats()
	if err != nil {
		return nil, fmt.Errorf("获取业务统计失败: %v", err)
	}

	// 获取性能统计
	performance := s.getPerformanceStats()

	// 获取连接统计
	connections := s.getConnectionStats()

	return &SystemStats{
		Runtime:     runtimeStats,
		Business:    *business,
		Performance: performance,
		Connections: connections,
	}, nil
}

// checkDatabaseHealth 检查数据库健康状态
func (s *SystemStatusService) checkDatabaseHealth() *ServiceStatus {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := &ServiceStatus{
		Name:      "数据库连接",
		CheckedAt: time.Now(),
	}

	// 获取底层数据库连接
	sqlDB, err := s.db.DB()
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("获取数据库连接失败: %v", err)
		return status
	}

	// 执行ping检查
	if err := sqlDB.PingContext(ctx); err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("数据库连接失败: %v", err)
		return status
	}

	// 检查连接池状态
	dbStats := sqlDB.Stats()

	status.Status = "healthy"
	status.Message = "数据库连接正常"
	status.ResponseTime = time.Since(start).Milliseconds()
	status.Details = map[string]interface{}{
		"open_connections": dbStats.OpenConnections,
		"in_use":           dbStats.InUse,
		"idle":             dbStats.Idle,
		"max_open":         dbStats.MaxOpenConnections,
		"wait_count":       dbStats.WaitCount,
		"wait_duration":    dbStats.WaitDuration.String(),
	}

	return status
}

// checkEmailMonitorHealth 检查邮件监控服务健康状态
func (s *SystemStatusService) checkEmailMonitorHealth() *ServiceStatus {
	start := time.Now()

	status := &ServiceStatus{
		Name:      "邮件监控服务",
		CheckedAt: time.Now(),
	}

	if s.emailMonitorService == nil {
		status.Status = "unhealthy"
		status.Message = "邮件监控服务未初始化"
		return status
	}

	// 检查监控服务是否运行
	isRunning := s.emailMonitorService.IsRunning()
	monitorStatus := s.emailMonitorService.GetStatus()

	if isRunning {
		status.Status = "healthy"
		status.Message = "邮件监控服务运行中"
	} else {
		status.Status = "degraded"
		status.Message = "邮件监控服务已停止"
	}

	status.ResponseTime = time.Since(start).Milliseconds()
	status.Details = monitorStatus

	return status
}

// checkNotificationHealth 检查通知服务健康状态
func (s *SystemStatusService) checkNotificationHealth() *ServiceStatus {
	start := time.Now()

	status := &ServiceStatus{
		Name:      "通知服务",
		CheckedAt: time.Now(),
	}

	if s.notificationService == nil {
		status.Status = "unhealthy"
		status.Message = "通知服务未初始化"
		return status
	}

	// 检查通知渠道配置
	channels, _, err := s.channelService.GetChannelList(1, 100, "", "")
	if err != nil {
		status.Status = "degraded"
		status.Message = fmt.Sprintf("获取通知渠道失败: %v", err)
		return status
	}

	activeChannels := 0
	for _, channel := range channels {
		if channel.Status == "active" {
			activeChannels++
		}
	}

	if activeChannels > 0 {
		status.Status = "healthy"
		status.Message = fmt.Sprintf("通知服务正常，活跃渠道: %d", activeChannels)
	} else {
		status.Status = "degraded"
		status.Message = "未配置活跃通知渠道"
	}

	status.ResponseTime = time.Since(start).Milliseconds()
	status.Details = map[string]interface{}{
		"total_channels":  len(channels),
		"active_channels": activeChannels,
	}

	return status
}

// checkCacheHealth 检查缓存服务健康状态
func (s *SystemStatusService) checkCacheHealth() *ServiceStatus {
	start := time.Now()

	status := &ServiceStatus{
		Name:      "缓存服务",
		CheckedAt: time.Now(),
	}

	// 这里可以添加Redis或其他缓存服务的健康检查
	// 目前假设使用内存缓存
	status.Status = "healthy"
	status.Message = "内存缓存正常"
	status.ResponseTime = time.Since(start).Milliseconds()
	status.Details = map[string]interface{}{
		"type": "memory",
	}

	return status
}

// calculateOverallStatus 计算系统整体状态
func (s *SystemStatusService) calculateOverallStatus(services []ServiceStatus) string {
	if len(services) == 0 {
		return "unknown"
	}

	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, service := range services {
		switch service.Status {
		case "healthy":
			healthyCount++
		case "degraded":
			degradedCount++
		case "unhealthy":
			unhealthyCount++
		}
	}

	// 如果有任何服务不健康，整体状态为不健康
	if unhealthyCount > 0 {
		return "unhealthy"
	}

	// 如果有降级服务，整体状态为降级
	if degradedCount > 0 {
		return "degraded"
	}

	// 所有服务都健康
	return "healthy"
}

// getSystemResources 获取系统资源使用情况
func (s *SystemStatusService) getSystemResources() SystemResources {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 获取最近的GC信息
	var lastGC time.Time
	if m.NumGC > 0 {
		lastGC = time.Unix(0, int64(m.LastGC))
	}

	return SystemResources{
		CPU: CPUStats{
			Cores:        runtime.NumCPU(),
			UsagePercent: 0, // 需要使用第三方库来获取CPU使用率
		},
		Memory: MemoryStats{
			Alloc:        m.Alloc,
			TotalAlloc:   m.TotalAlloc,
			Sys:          m.Sys,
			HeapAlloc:    m.HeapAlloc,
			HeapSys:      m.HeapSys,
			UsagePercent: float64(m.HeapAlloc) / float64(m.HeapSys) * 100,
		},
		Goroutines: runtime.NumGoroutine(),
		GC: GCStats{
			NumGC:      m.NumGC,
			LastGC:     lastGC,
			TotalPause: m.PauseTotalNs,
			AvgPause: func() uint64 {
				if m.NumGC > 0 {
					return m.PauseTotalNs / uint64(m.NumGC)
				}
				return 0
			}(),
		},
	}
}

// calculateHealthSummary 计算健康状况摘要
func (s *SystemStatusService) calculateHealthSummary(services []ServiceStatus) HealthSummary {
	summary := HealthSummary{
		TotalServices: len(services),
	}

	for _, service := range services {
		switch service.Status {
		case "healthy":
			summary.HealthyServices++
		case "degraded":
			summary.DegradedServices++
		case "unhealthy":
			summary.UnhealthyServices++
		}
	}

	return summary
}

// getUptime 获取系统运行时间
func (s *SystemStatusService) getUptime() string {
	duration := time.Since(s.startTime)

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		return fmt.Sprintf("%d分钟", minutes)
	}
}

// getBusinessStats 获取业务统计信息
func (s *SystemStatusService) getBusinessStats() (*BusinessStats, error) {
	stats := &BusinessStats{}

	// 获取邮箱统计
	if s.mailboxService != nil {
		mailboxResponse, err := s.mailboxService.List(1, 1000, "")
		if err == nil {
			stats.TotalMailboxes = len(mailboxResponse.List)
			for _, mb := range mailboxResponse.List {
				if mb.Status == "active" {
					stats.ActiveMailboxes++
				}
			}
		}
	}

	// 获取规则统计
	var totalRules int64
	if err := s.db.Model(&model.RuleGroup{}).Count(&totalRules).Error; err == nil {
		stats.TotalRules = int(totalRules)
	}

	var activeRules int64
	if err := s.db.Model(&model.RuleGroup{}).Where("status = ?", "active").Count(&activeRules).Error; err == nil {
		stats.ActiveRules = int(activeRules)
	}

	// 获取渠道统计
	if s.channelService != nil {
		channels, _, err := s.channelService.GetChannelList(1, 1000, "", "")
		if err == nil {
			stats.TotalChannels = len(channels)
			for _, ch := range channels {
				if ch.Status == "active" {
					stats.ActiveChannels++
				}
			}
		}
	}

	// 获取告警统计
	if s.alertService != nil {
		// 获取总告警数
		_, totalCount, err := s.alertService.GetAlerts(1, 1, map[string]interface{}{})
		if err == nil {
			stats.TotalAlerts = int(totalCount)
		}

		// 获取今日告警数
		today := time.Now().Format("2006-01-02")
		todayFilters := map[string]interface{}{
			"date_from": today,
			"date_to":   today,
		}
		_, todayCount, err := s.alertService.GetAlerts(1, 1, todayFilters)
		if err == nil {
			stats.TodayAlerts = int(todayCount)
		}

		// 获取待处理告警数
		pendingFilters := map[string]interface{}{
			"status": "pending",
		}
		_, pendingCount, err := s.alertService.GetAlerts(1, 1, pendingFilters)
		if err == nil {
			stats.PendingAlerts = int(pendingCount)
		}

		// 计算成功率
		if stats.TodayAlerts > 0 {
			successFilters := map[string]interface{}{
				"status":    "sent",
				"date_from": today,
				"date_to":   today,
			}
			_, successCount, err := s.alertService.GetAlerts(1, 1, successFilters)
			if err == nil {
				stats.SuccessRate = float64(successCount) / float64(stats.TodayAlerts) * 100
			}
		}
	}

	// 获取通知记录统计
	var totalNotifications int64
	if err := s.db.Model(&model.NotificationLog{}).Count(&totalNotifications).Error; err == nil {
		stats.TotalNotifications = int(totalNotifications)
	}

	// 获取今日通知记录数
	var todayNotifications int64
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	if err := s.db.Model(&model.NotificationLog{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Count(&todayNotifications).Error; err == nil {
		stats.TodayNotifications = int(todayNotifications)
	}

	return stats, nil
}

// getPerformanceStats 获取性能统计信息
func (s *SystemStatusService) getPerformanceStats() PerformanceStats {
	// 这里可以添加实际的性能统计逻辑
	// 目前返回模拟数据
	return PerformanceStats{
		AvgProcessingTime: 150.5,  // 毫秒
		MaxProcessingTime: 2500.0, // 毫秒
		RequestsPerSecond: 25.3,
		ErrorRate:         0.5, // 百分比
	}
}

// getConnectionStats 获取连接统计信息
func (s *SystemStatusService) getConnectionStats() ConnectionStats {
	stats := ConnectionStats{
		MaxConnections: 100, // 默认最大连接数
	}

	// 获取数据库连接统计
	if sqlDB, err := s.db.DB(); err == nil {
		dbStats := sqlDB.Stats()
		stats.DatabaseConnections = dbStats.OpenConnections
		stats.ActiveConnections = dbStats.InUse
		if dbStats.MaxOpenConnections > 0 {
			stats.MaxConnections = dbStats.MaxOpenConnections
		}
	}

	return stats
}

// CleanupHistoryData 清理历史数据
func (s *SystemStatusService) CleanupHistoryData(dataType string, timeRange string) (*CleanupResult, error) {
	var cutoffTime time.Time
	now := time.Now()

	// 根据时间范围计算截止时间
	switch timeRange {
	case "1month":
		cutoffTime = now.AddDate(0, -1, 0)
	case "3months":
		cutoffTime = now.AddDate(0, -3, 0)
	case "6months":
		cutoffTime = now.AddDate(0, -6, 0)
	case "1year":
		cutoffTime = now.AddDate(-1, 0, 0)
	case "2years":
		cutoffTime = now.AddDate(-2, 0, 0)
	case "all":
		cutoffTime = time.Time{} // 最早时间，清理所有数据
	default:
		return nil, fmt.Errorf("不支持的时间范围: %s", timeRange)
	}

	result := &CleanupResult{
		DataType:    dataType,
		TimeRange:   timeRange,
		CutoffTime:  cutoffTime,
		StartTime:   now,
		DeletedRows: 0,
	}

	var err error
	switch dataType {
	case "alerts":
		result.DeletedRows, err = s.cleanupAlerts(cutoffTime)
	case "notifications":
		result.DeletedRows, err = s.cleanupNotifications(cutoffTime)
	case "both":
		alertRows, alertErr := s.cleanupAlerts(cutoffTime)
		if alertErr != nil {
			return nil, fmt.Errorf("清理告警数据失败: %v", alertErr)
		}
		notificationRows, notificationErr := s.cleanupNotifications(cutoffTime)
		if notificationErr != nil {
			return nil, fmt.Errorf("清理通知数据失败: %v", notificationErr)
		}
		result.DeletedRows = alertRows + notificationRows
	default:
		return nil, fmt.Errorf("不支持的数据类型: %s", dataType)
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// cleanupAlerts 清理告警历史数据
func (s *SystemStatusService) cleanupAlerts(cutoffTime time.Time) (int64, error) {
	var result *gorm.DB

	if cutoffTime.IsZero() {
		// 清理所有数据 - 使用正确的表名
		result = s.db.Exec("DELETE FROM alerts")
		if result.Error != nil {
			return 0, result.Error
		}
		// 重置自增ID
		s.db.Exec("DELETE FROM sqlite_sequence WHERE name = 'alerts'")
	} else {
		// 清理指定时间之前的数据
		result = s.db.Exec("DELETE FROM alerts WHERE created_at < ?", cutoffTime)
	}

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// cleanupNotifications 清理通知历史数据
func (s *SystemStatusService) cleanupNotifications(cutoffTime time.Time) (int64, error) {
	var result *gorm.DB

	if cutoffTime.IsZero() {
		// 清理所有数据 - 使用正确的表名
		result = s.db.Exec("DELETE FROM notification_logs")
		if result.Error != nil {
			return 0, result.Error
		}
		// 重置自增ID
		s.db.Exec("DELETE FROM sqlite_sequence WHERE name = 'notification_logs'")
	} else {
		// 清理指定时间之前的数据
		result = s.db.Exec("DELETE FROM notification_logs WHERE created_at < ?", cutoffTime)
	}

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// CleanupResult 清理结果
type CleanupResult struct {
	DataType    string        `json:"data_type"`    // 数据类型
	TimeRange   string        `json:"time_range"`   // 时间范围
	CutoffTime  time.Time     `json:"cutoff_time"`  // 截止时间
	StartTime   time.Time     `json:"start_time"`   // 开始时间
	EndTime     time.Time     `json:"end_time"`     // 结束时间
	Duration    time.Duration `json:"duration"`     // 执行耗时
	DeletedRows int64         `json:"deleted_rows"` // 删除行数
	Success     bool          `json:"success"`      // 是否成功
	Error       string        `json:"error"`        // 错误信息
}
