package database

import (
	"DevOpsProject/backend/internal/models"
	"errors"
	"time"
)

// DeployRecordService 部署记录服务
type DeployRecordService struct{}

// NewDeployRecordService 创建部署记录服务
func NewDeployRecordService() *DeployRecordService {
	return &DeployRecordService{}
}

// CreateDeployRecord 创建部署记录
func (s *DeployRecordService) CreateDeployRecord(record *models.DeployRecord) error {
	record.CreatedAt = models.CustomTime{Time: time.Now()}
	record.UpdatedAt = models.CustomTime{Time: time.Now()}
	return DB.Create(record).Error
}

// GetDeployRecordByID 根据ID获取部署记录
func (s *DeployRecordService) GetDeployRecordByID(id uint) (*models.DeployRecordResponse, error) {
	var record models.DeployRecord
	err := DB.First(&record, id).Error
	if err != nil {
		return nil, err
	}

	response := &models.DeployRecordResponse{
		ID:          record.ID,
		ProjectID:   record.ProjectID,
		ProjectName: record.ProjectName,
		Branch:      record.Branch,
		StartTime:   record.StartTime.Format("2006-01-02 15:04:05"),
		Duration:    record.Duration,
		LogPath:     record.LogPath,
		Status:      record.Status,
		CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return response, nil
}

// GetDeployRecordsByProjectID 根据项目ID获取部署记录列表
func (s *DeployRecordService) GetDeployRecordsByProjectID(projectID uint) ([]models.DeployRecordResponse, error) {
	var records []models.DeployRecord
	err := DB.Where("project_id = ?", projectID).Order("start_time DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	var responses []models.DeployRecordResponse
	for _, record := range records {
		response := models.DeployRecordResponse{
			ID:          record.ID,
			ProjectID:   record.ProjectID,
			ProjectName: record.ProjectName,
			Branch:      record.Branch,
			StartTime:   record.StartTime.Format("2006-01-02 15:04:05"),
			Duration:    record.Duration,
			LogPath:     record.LogPath,
			Status:      record.Status,
			CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// UpdateDeployRecord 更新部署记录
func (s *DeployRecordService) UpdateDeployRecord(id uint, updates map[string]interface{}) error {
	// 检查部署记录是否存在
	var record models.DeployRecord
	if err := DB.First(&record, id).Error; err != nil {
		return errors.New("部署记录不存在")
	}

	// 如果要更新开始时间，需要解析字符串为时间
	if startTime, ok := updates["start_time"].(string); ok {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", startTime)
		if err != nil {
			return errors.New("开始时间格式错误，应为 yyyy-MM-dd HH:mm:ss")
		}
		updates["start_time"] = models.CustomTime{Time: parsedTime}
	}

	updates["updated_at"] = models.CustomTime{Time: time.Now()}

	return DB.Model(&record).Updates(updates).Error
}

// DeleteDeployRecord 删除部署记录（软删除）
func (s *DeployRecordService) DeleteDeployRecord(id uint) error {
	return DB.Delete(&models.DeployRecord{}, id).Error
}

// GetDeployRecordsWithPagination 分页获取部署记录列表
func (s *DeployRecordService) GetDeployRecordsWithPagination(query *models.DeployRecordQuery) (*models.PageData, error) {
	var records []models.DeployRecord
	var total int64

	db := DB.Model(&models.DeployRecord{})

	// 添加查询条件
	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
	if query.ProjectName != "" {
		db = db.Where("project_name LIKE ?", "%"+query.ProjectName+"%")
	}
	if query.Branch != "" {
		db = db.Where("branch LIKE ?", "%"+query.Branch+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTimeStart != "" {
		db = db.Where("start_time >= ?", query.StartTimeStart)
	}
	if query.StartTimeEnd != "" {
		db = db.Where("start_time <= ?", query.StartTimeEnd)
	}

	// 获取总数
	db.Count(&total)

	// 分页查询，按开始时间倒序排列
	offset := (query.PageNum - 1) * query.PageSize
	err := db.Order("start_time DESC").Offset(offset).Limit(query.PageSize).Find(&records).Error
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	var responses []models.DeployRecordResponse
	for _, record := range records {
		response := models.DeployRecordResponse{
			ID:          record.ID,
			ProjectID:   record.ProjectID,
			ProjectName: record.ProjectName,
			Branch:      record.Branch,
			StartTime:   record.StartTime.Format("2006-01-02 15:04:05"),
			Duration:    record.Duration,
			LogPath:     record.LogPath,
			Status:      record.Status,
			CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	pageData := &models.PageData{
		Rows:  responses,
		Total: total,
	}

	return pageData, nil
}

// GetDeployRecordsByBranch 根据分支获取部署记录列表
func (s *DeployRecordService) GetDeployRecordsByBranch(branch string) ([]models.DeployRecordResponse, error) {
	var records []models.DeployRecord
	err := DB.Where("branch = ?", branch).Order("start_time DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	var responses []models.DeployRecordResponse
	for _, record := range records {
		response := models.DeployRecordResponse{
			ID:          record.ID,
			ProjectID:   record.ProjectID,
			ProjectName: record.ProjectName,
			Branch:      record.Branch,
			StartTime:   record.StartTime.Format("2006-01-02 15:04:05"),
			Duration:    record.Duration,
			LogPath:     record.LogPath,
			Status:      record.Status,
			CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetDeployRecordsByStatus 根据状态获取部署记录列表
func (s *DeployRecordService) GetDeployRecordsByStatus(status string) ([]models.DeployRecordResponse, error) {
	var records []models.DeployRecord
	err := DB.Where("status = ?", status).Order("start_time DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	var responses []models.DeployRecordResponse
	for _, record := range records {
		response := models.DeployRecordResponse{
			ID:          record.ID,
			ProjectID:   record.ProjectID,
			ProjectName: record.ProjectName,
			Branch:      record.Branch,
			StartTime:   record.StartTime.Format("2006-01-02 15:04:05"),
			Duration:    record.Duration,
			LogPath:     record.LogPath,
			Status:      record.Status,
			CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetLatestDeployRecordByProjectAndBranch 获取指定项目和分支的最新部署记录
func (s *DeployRecordService) GetLatestDeployRecordByProjectAndBranch(projectID uint, branch string) (*models.DeployRecordResponse, error) {
	var record models.DeployRecord
	err := DB.Where("project_id = ? AND branch = ?", projectID, branch).Order("start_time DESC").First(&record).Error
	if err != nil {
		return nil, err
	}

	response := &models.DeployRecordResponse{
		ID:          record.ID,
		ProjectID:   record.ProjectID,
		ProjectName: record.ProjectName,
		Branch:      record.Branch,
		StartTime:   record.StartTime.Format("2006-01-02 15:04:05"),
		Duration:    record.Duration,
		LogPath:     record.LogPath,
		Status:      record.Status,
		CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return response, nil
}

// GetDeployRecordStats 获取部署记录统计信息
func (s *DeployRecordService) GetDeployRecordStats(projectID uint) (map[string]int64, error) {
	stats := make(map[string]int64)

	db := DB.Model(&models.DeployRecord{})
	if projectID > 0 {
		db = db.Where("project_id = ?", projectID)
	}

	// 总部署次数
	var total int64
	db.Count(&total)
	stats["total"] = total

	// 成功次数
	var success int64
	db.Where("status = ?", "success").Count(&success)
	stats["success"] = success

	// 失败次数
	var failed int64
	db.Where("status = ?", "failed").Count(&failed)
	stats["failed"] = failed

	// 运行中次数
	var running int64
	db.Where("status = ?", "running").Count(&running)
	stats["running"] = running

	return stats, nil
}