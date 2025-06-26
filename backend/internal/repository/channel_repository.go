package repository

import (
	"emailAlert/internal/model"
	"time"

	"gorm.io/gorm"
)

// ChannelRepository 通知渠道仓库接口
type ChannelRepository interface {
	Create(channel *model.Channel) error
	GetByID(id uint) (*model.Channel, error)
	GetList(page, size int, channelType, status string) ([]*model.Channel, int64, error)
	Update(channel *model.Channel) error
	Delete(id uint) error
	GetByType(channelType string) ([]*model.Channel, error)
	UpdateStatus(id uint, status string) error
	UpdateTestResult(id uint, testResult string, testTime time.Time) error
	GetActiveChannels() ([]*model.Channel, error)
}

// channelRepository 通知渠道仓库实现
type channelRepository struct {
	db *gorm.DB
}

// NewChannelRepository 创建通知渠道仓库
func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

// Create 创建通知渠道
func (r *channelRepository) Create(channel *model.Channel) error {
	return r.db.Create(channel).Error
}

// GetByID 根据ID获取通知渠道
func (r *channelRepository) GetByID(id uint) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.First(&channel, id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// GetList 获取通知渠道列表
func (r *channelRepository) GetList(page, size int, channelType, status string) ([]*model.Channel, int64, error) {
	var channels []*model.Channel
	var total int64

	query := r.db.Model(&model.Channel{})

	// 类型筛选
	if channelType != "" {
		query = query.Where("type = ?", channelType)
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&channels).Error
	if err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// Update 更新通知渠道
func (r *channelRepository) Update(channel *model.Channel) error {
	return r.db.Save(channel).Error
}

// Delete 删除通知渠道（软删除）
func (r *channelRepository) Delete(id uint) error {
	return r.db.Delete(&model.Channel{}, id).Error
}

// GetByType 根据类型获取通知渠道
func (r *channelRepository) GetByType(channelType string) ([]*model.Channel, error) {
	var channels []*model.Channel
	err := r.db.Where("type = ? AND status = ?", channelType, "active").Find(&channels).Error
	return channels, err
}

// UpdateStatus 更新通知渠道状态
func (r *channelRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateTestResult 更新测试结果
func (r *channelRepository) UpdateTestResult(id uint, testResult string, testTime time.Time) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"test_result":  testResult,
		"last_test_at": testTime,
	}).Error
}

// GetActiveChannels 获取所有激活的通知渠道
func (r *channelRepository) GetActiveChannels() ([]*model.Channel, error) {
	var channels []*model.Channel
	err := r.db.Where("status = ?", "active").Find(&channels).Error
	return channels, err
}
