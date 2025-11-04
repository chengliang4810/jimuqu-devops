package database

import (
	"encoding/json"
	"errors"
	"DevOpsProject/backend/internal/models"
	"time"
)

// DeployConfigService 部署配置服务
type DeployConfigService struct{}

// NewDeployConfigService 创建部署配置服务
func NewDeployConfigService() *DeployConfigService {
	return &DeployConfigService{}
}

// CreateDeployConfig 创建部署配置
func (s *DeployConfigService) CreateDeployConfig(config *models.DeployConfig) error {
	// 检查项目ID和分支的唯一性
	var existing models.DeployConfig
	err := DB.Where("project_id = ? AND branch = ?", config.ProjectID, config.Branch).First(&existing).Error
	if err == nil {
		return errors.New("该项目的该分支已存在部署配置")
	}

	// 序列化配置为JSON
	configBytes, err := json.Marshal(config.Config)
	if err != nil {
		return err
	}
	config.Config = string(configBytes)

	config.CreatedAt = models.CustomTime{Time: time.Now()}
	config.UpdatedAt = models.CustomTime{Time: time.Now()}

	return DB.Create(config).Error
}

// GetDeployConfigByID 根据ID获取部署配置
func (s *DeployConfigService) GetDeployConfigByID(id uint) (*models.DeployConfigResponse, error) {
	var config models.DeployConfig
	err := DB.First(&config, id).Error
	if err != nil {
		return nil, err
	}

	// 反序列化配置
	var configItems []models.DeployConfigItem
	if config.Config != "" {
		err = json.Unmarshal([]byte(config.Config), &configItems)
		if err != nil {
			return nil, err
		}
	}

	response := &models.DeployConfigResponse{
		ID:        config.ID,
		ProjectID: config.ProjectID,
		Branch:    config.Branch,
		Config:    configItems,
		CreatedAt: config.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: config.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return response, nil
}

// GetDeployConfigsByProjectID 根据项目ID获取部署配置列表
func (s *DeployConfigService) GetDeployConfigsByProjectID(projectID uint) ([]models.DeployConfigResponse, error) {
	var configs []models.DeployConfig
	err := DB.Where("project_id = ?", projectID).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	var responses []models.DeployConfigResponse
	for _, config := range configs {
		// 反序列化配置
		var configItems []models.DeployConfigItem
		if config.Config != "" {
			err = json.Unmarshal([]byte(config.Config), &configItems)
			if err != nil {
				continue
			}
			}

		response := models.DeployConfigResponse{
			ID:        config.ID,
			ProjectID: config.ProjectID,
			Branch:    config.Branch,
			Config:    configItems,
			CreatedAt: config.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: config.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetDeployConfigByProjectAndBranch 根据项目ID和分支获取部署配置
func (s *DeployConfigService) GetDeployConfigByProjectAndBranch(projectID uint, branch string) (*models.DeployConfigResponse, error) {
	var config models.DeployConfig
	err := DB.Where("project_id = ? AND branch = ?", projectID, branch).First(&config).Error
	if err != nil {
		return nil, err
	}

	// 反序列化配置
	var configItems []models.DeployConfigItem
	if config.Config != "" {
		err = json.Unmarshal([]byte(config.Config), &configItems)
		if err != nil {
			return nil, err
		}
	}

	response := &models.DeployConfigResponse{
		ID:        config.ID,
		ProjectID: config.ProjectID,
		Branch:    config.Branch,
		Config:    configItems,
		CreatedAt: config.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: config.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return response, nil
}

// UpdateDeployConfig 更新部署配置
func (s *DeployConfigService) UpdateDeployConfig(id uint, updates map[string]interface{}) error {
	// 检查部署配置是否存在
	var config models.DeployConfig
	if err := DB.First(&config, id).Error; err != nil {
		return errors.New("部署配置不存在")
	}

	// 如果要更新分支或项目ID，需要检查唯一性
	if projectID, ok := updates["project_id"].(uint); ok {
		if branch, branchOk := updates["branch"].(string); branchOk {
			var existing models.DeployConfig
			err := DB.Where("project_id = ? AND branch = ? AND id != ?", projectID, branch, id).First(&existing)
			if err == nil {
				return errors.New("该项目的该分支已存在部署配置")
			}
		}
	}

	// 如果要更新配置内容
	if configItems, ok := updates["config"].([]models.DeployConfigItem); ok {
		configBytes, err := json.Marshal(configItems)
		if err != nil {
			return err
		}
		updates["config"] = string(configBytes)
	}

	updates["updated_at"] = models.CustomTime{Time: time.Now()}

	return DB.Model(&config).Updates(updates).Error
}

// DeleteDeployConfig 删除部署配置（软删除）
func (s *DeployConfigService) DeleteDeployConfig(id uint) error {
	return DB.Delete(&models.DeployConfig{}, id).Error
}

// GetDeployConfigsWithPagination 分页获取部署配置列表
func (s *DeployConfigService) GetDeployConfigsWithPagination(query *models.DeployConfigQuery) (*models.PageData, error) {
	var configs []models.DeployConfig
	var total int64

	db := DB.Model(&models.DeployConfig{})

	// 添加查询条件
	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
	if query.Branch != "" {
		db = db.Where("branch LIKE ?", "%"+query.Branch+"%")
	}

	// 获取总数
	db.Count(&total)

	// 分页查询
	offset := (query.PageNum - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	var responses []models.DeployConfigResponse
	for _, config := range configs {
		// 反序列化配置
		var configItems []models.DeployConfigItem
		if config.Config != "" {
			err = json.Unmarshal([]byte(config.Config), &configItems)
			if err != nil {
				continue
			}
		}

		response := models.DeployConfigResponse{
			ID:        config.ID,
			ProjectID: config.ProjectID,
			Branch:    config.Branch,
			Config:    configItems,
			CreatedAt: config.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: config.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	pageData := &models.PageData{
		Rows:  responses,
		Total: total,
	}

	return pageData, nil
}

// CheckDeployConfigExists 检查部署配置是否存在
func (s *DeployConfigService) CheckDeployConfigExists(projectID uint, branch string, excludeID uint) (bool, error) {
	var count int64
	query := DB.Model(&models.DeployConfig{}).Where("project_id = ? AND branch = ?", projectID, branch)

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	query.Count(&count)

	return count > 0, nil
}

// GetDeployConfigsByBranch 根据分支获取部署配置列表（支持多个项目）
func (s *DeployConfigService) GetDeployConfigsByBranch(branch string) ([]models.DeployConfigResponse, error) {
	var configs []models.DeployConfig
	err := DB.Where("branch = ?", branch).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	var responses []models.DeployConfigResponse
	for _, config := range configs {
		// 反序列化配置
		var configItems []models.DeployConfigItem
		if config.Config != "" {
			err = json.Unmarshal([]byte(config.Config), &configItems)
			if err != nil {
				continue
			}
		}

		response := models.DeployConfigResponse{
			ID:        config.ID,
			ProjectID: config.ProjectID,
			Branch:    config.Branch,
			Config:    configItems,
			CreatedAt: config.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: config.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	return responses, nil
}