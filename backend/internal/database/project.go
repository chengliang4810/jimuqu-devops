package database

import (
	"DevOpsProject/backend/internal/models"
	"fmt"
	"time"
)

// ProjectService 项目服务
type ProjectService struct{}

// NewProjectService 创建项目服务实例
func NewProjectService() *ProjectService {
	return &ProjectService{}
}

// CreateProject 创建项目
func (s *ProjectService) CreateProject(project *models.Project) error {
	if project.Name == "" || project.Code == "" {
		return fmt.Errorf("项目名称和项目编码不能为空")
	}

	// 检查项目编码是否已存在
	var existingProject models.Project
	if err := DB.Where("code = ?", project.Code).First(&existingProject).Error; err == nil {
		return fmt.Errorf("项目编码 '%s' 已存在", project.Code)
	}

	// 设置创建和更新时间
	now := time.Now()
	project.CreatedAt = models.CustomTime{Time: now}
	project.UpdatedAt = models.CustomTime{Time: now}

	return DB.Create(project).Error
}

// GetProjectByID 根据ID获取项目
func (s *ProjectService) GetProjectByID(id uint) (*models.Project, error) {
	var project models.Project
	if err := DB.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetAllProjects 获取所有项目
func (s *ProjectService) GetAllProjects() ([]models.Project, error) {
	var projects []models.Project
	if err := DB.Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProjectsWithPagination 分页查询项目
func (s *ProjectService) GetProjectsWithPagination(query *models.ProjectQuery) (*models.PageResult, error) {
	var projects []models.Project
	var total int64

	// 构建查询条件
	db := DB.Model(&models.Project{})

	// 添加查询条件
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Code != "" {
		db = db.Where("code LIKE ?", "%"+query.Code+"%")
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (query.PageNum - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, err
	}

	return &models.PageResult{
		List:  projects,
		Total: total,
	}, nil
}

// UpdateProject 更新项目
func (s *ProjectService) UpdateProject(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}

	// 检查项目是否存在
	var project models.Project
	if err := DB.First(&project, id).Error; err != nil {
		return fmt.Errorf("项目不存在: %v", err)
	}

	// 如果更新项目编码，检查是否重复
	if code, ok := updates["code"]; ok {
		codeStr, ok := code.(string)
		if !ok {
			return fmt.Errorf("项目编码格式错误")
		}
		if codeStr != project.Code {
			var existingProject models.Project
			if err := DB.Where("code = ? AND id != ?", codeStr, id).First(&existingProject).Error; err == nil {
				return fmt.Errorf("项目编码 '%s' 已存在", codeStr)
			}
		}
	}

	// 设置更新时间
	updates["updated_at"] = models.CustomTime{Time: time.Now()}

	return DB.Model(&project).Updates(updates).Error
}

// DeleteProject 删除项目（软删除）
func (s *ProjectService) DeleteProject(id uint) error {
	var project models.Project
	if err := DB.First(&project, id).Error; err != nil {
		return fmt.Errorf("项目不存在: %v", err)
	}

	return DB.Delete(&project).Error
}

// GetProjectByCode 根据项目编码获取项目
func (s *ProjectService) GetProjectByCode(code string) (*models.Project, error) {
	var project models.Project
	if err := DB.Where("code = ?", code).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// CheckProjectCodeExists 检查项目编码是否存在
func (s *ProjectService) CheckProjectCodeExists(code string, excludeID ...uint) (bool, error) {
	var count int64
	db := DB.Model(&models.Project{}).Where("code = ?", code)

	if len(excludeID) > 0 {
		db = db.Where("id != ?", excludeID[0])
	}

	if err := db.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}