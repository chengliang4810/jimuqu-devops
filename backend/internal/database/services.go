package database

import (
	"DevOpsProject/backend/internal/models"
	"fmt"
	"time"
)

// HostService 主机服务
type HostService struct{}

// NewHostService 创建主机服务实例
func NewHostService() *HostService {
	return &HostService{}
}

// CreateHost 创建主机
// @Summary 创建主机
// @Description 创建新的SSH主机配置
// @Tags hosts
// @Accept json
// @Produce json
// @Param host body models.Host true "主机信息"
// @Success 201 {object} models.Response "创建成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host [post]
func (s *HostService) CreateHost(host *models.Host) error {
	if host.Name == "" || host.Host == "" || host.Username == "" || host.Password == "" {
		return fmt.Errorf("主机名称、地址、用户名和密码不能为空")
	}

	// 设置创建和更新时间
	now := time.Now()
	host.CreatedAt = models.CustomTime{Time: now}
	host.UpdatedAt = models.CustomTime{Time: now}

	return DB.Create(host).Error
}

// GetHostByID 根据ID获取主机
// @Summary 获取主机详情
// @Description 根据ID获取指定主机的详细信息
// @Tags hosts
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "主机不存在"
// @Router /api/host/{id} [get]
func (s *HostService) GetHostByID(id uint) (*models.Host, error) {
	var host models.Host
	err := DB.First(&host, id).Error
	if err != nil {
		return nil, fmt.Errorf("主机不存在: %v", err)
	}
	return &host, nil
}

// GetAllHosts 获取所有主机
// @Summary 获取所有主机
// @Description 获取所有主机配置列表
// @Tags hosts
// @Accept json
// @Produce json
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host [get]
func (s *HostService) GetAllHosts() ([]models.Host, error) {
	var hosts []models.Host
	err := DB.Find(&hosts).Error
	return hosts, err
}

// GetHostsWithPagination 分页获取主机列表
// @Summary 分页获取主机列表
// @Description 分页获取主机配置列表，支持pageNum和pageSize参数，支持主机名模糊查询、IP精确查询、状态精确查询
// @Tags hosts
// @Accept json
// @Produce json
// @Param pageNum query int false "页码" example:"1"
// @Param pageSize query int false "每页条数" example:"10"
// @Param name query string false "主机名模糊查询" example:"测试主机"
// @Param host query string false "IP地址精确查询" example:"192.168.1.100"
// @Param status query string false "主机状态精确查询" example:"online"
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host [get]
func (s *HostService) GetHostsWithPagination(query *models.HostQuery) (*models.PageResult, error) {
	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	var total int64
	var hosts []models.Host

	// 构建查询条件
	db := DB.Model(&models.Host{})

	// 主机名模糊查询
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}

	// IP地址精确查询
	if query.Host != "" {
		db = db.Where("host = ?", query.Host)
	}

	// 主机状态精确查询
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 计算偏移量
	offset := (query.PageNum - 1) * query.PageSize

	// 分页查询
	if err := db.Offset(offset).Limit(query.PageSize).Find(&hosts).Error; err != nil {
		return nil, err
	}

	return &models.PageResult{
		List:  hosts,
		Total: total,
	}, nil
}

// UpdateHost 更新主机
// @Summary 更新主机信息
// @Description 更新指定主机的配置信息
// @Tags hosts
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Param host body map[string]interface{} true "更新的主机信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} models.Response "ID格式错误或请求参数错误"
// @Failure 404 {object} models.Response "主机不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host/{id} [put]
func (s *HostService) UpdateHost(id uint, updates map[string]interface{}) error {
	// 设置更新时间
	updates["updated_at"] = models.CustomTime{Time: time.Now()}
	return DB.Model(&models.Host{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteHost 删除主机
// @Summary 删除主机
// @Description 根据ID删除指定主机（软删除）
// @Tags hosts
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "主机不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host/{id} [delete]
func (s *HostService) DeleteHost(id uint) error {
	return DB.Delete(&models.Host{}, id).Error
}