package repository

import (
	"emailAlert/internal/model"
	"errors"

	"gorm.io/gorm"
)

// TemplateRepository 模版数据访问层
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository 创建模版仓库实例
func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create 创建模版
func (r *TemplateRepository) Create(template *model.Template) error {
	return r.db.Create(template).Error
}

// GetByID 根据ID获取模版
func (r *TemplateRepository) GetByID(id uint) (*model.Template, error) {
	var template model.Template
	err := r.db.First(&template, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模版不存在")
		}
		return nil, err
	}
	return &template, nil
}

// GetByName 根据名称获取模版
func (r *TemplateRepository) GetByName(name string) (*model.Template, error) {
	var template model.Template
	err := r.db.Where("name = ?", name).First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模版不存在")
		}
		return nil, err
	}
	return &template, nil
}

// List 获取模版列表
func (r *TemplateRepository) List(page, pageSize int, templateType, status string) ([]model.Template, int64, error) {
	var templates []model.Template
	var total int64

	query := r.db.Model(&model.Template{})

	// 类型过滤
	if templateType != "" {
		query = query.Where("type = ?", templateType)
	}

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&templates).Error
	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Update 更新模版
func (r *TemplateRepository) Update(id uint, template *model.Template) error {
	return r.db.Model(&model.Template{}).Where("id = ?", id).Updates(template).Error
}

// Delete 删除模版
func (r *TemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.Template{}, id).Error
}

// UpdateStatus 更新模版状态
func (r *TemplateRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Template{}).Where("id = ?", id).Update("status", status).Error
}

// GetActiveTemplates 获取所有活跃的模版
func (r *TemplateRepository) GetActiveTemplates() ([]model.Template, error) {
	var templates []model.Template
	err := r.db.Where("status = ?", "active").Find(&templates).Error
	return templates, err
}

// GetByType 根据类型获取模版列表
func (r *TemplateRepository) GetByType(templateType string) ([]model.Template, error) {
	var templates []model.Template
	err := r.db.Where("type = ? AND status = ?", templateType, "active").Find(&templates).Error
	return templates, err
}

// GetDefaultByType 获取指定类型的默认模版
func (r *TemplateRepository) GetDefaultByType(templateType string) (*model.Template, error) {
	var template model.Template
	err := r.db.Where("type = ? AND is_default = ? AND status = ?", templateType, true, "active").First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("默认模版不存在")
		}
		return nil, err
	}
	return &template, nil
}

// SetDefault 设置默认模版
func (r *TemplateRepository) SetDefault(id uint, templateType string) error {
	// 先取消同类型的其他默认模版
	if err := r.db.Model(&model.Template{}).Where("type = ? AND id != ?", templateType, id).Update("is_default", false).Error; err != nil {
		return err
	}

	// 设置当前模版为默认
	return r.db.Model(&model.Template{}).Where("id = ?", id).Update("is_default", true).Error
}

// NameExists 检查模版名称是否已存在
func (r *TemplateRepository) NameExists(name string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&model.Template{}).Where("name = ?", name).Where("deleted_at IS NULL")

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}
