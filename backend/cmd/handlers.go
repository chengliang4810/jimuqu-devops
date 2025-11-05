package main

import (
	"DevOpsProject/backend/internal/database"
	"DevOpsProject/backend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "操作成功",
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, models.Response{
		Code:    code,
		Message: message,
	})
}

// ParsePaginationQuery 解析分页参数
func ParsePaginationQuery(c *gin.Context) (int, int) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页面大小
	}

	return pageNum, pageSize
}

// ParseHostQuery 解析主机查询参数
func ParseHostQuery(c *gin.Context) *models.HostQuery {
	var query models.HostQuery

	// 解析分页参数
	query.PageNum, _ = strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100 // 限制最大页面大小
	}

	// 解析查询条件
	query.Name = c.Query("name")
	query.Host = c.Query("host")
	query.Status = c.Query("status")

	return &query
}

// ParseProjectQuery 解析项目查询参数
func ParseProjectQuery(c *gin.Context) *models.ProjectQuery {
	var query models.ProjectQuery

	// 解析分页参数
	query.PageNum, _ = strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100 // 限制最大页面大小
	}

	// 解析查询条件
	query.Name = c.Query("name")
	query.Code = c.Query("code")

	return &query
}

// @Summary 创建项目
// @Description 创建新的项目
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param request body models.ProjectCreateRequest true "项目信息"
// @Success 201 {object} models.Response "创建成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "创建失败"
// @Router /api/project [post]
func createProjectHandler(service *database.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ProjectCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 转换为项目模型
		project := &models.Project{
			Name:            req.Name,
			Code:            req.Code,
			Remark:          req.Remark,
			GitRepo:         req.GitRepo,
			GitUsername:     req.GitUsername,
			GitPassword:     req.GitPassword,
			WebhookPassword: req.WebhookPassword,
		}

		if err := service.CreateProject(project); err != nil {
			ErrorResponse(c, 500, "创建项目失败: "+err.Error())
			return
		}

		SuccessResponse(c, project)
	}
}

// @Summary 获取项目列表
// @Description 获取所有项目列表，支持分页和条件查询
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param pageNum query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param name query string false "项目名称模糊查询"
// @Param code query string false "项目编码模糊查询"
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/project [get]
func getAllProjectsHandler(service *database.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有分页参数或查询条件
		query := ParseProjectQuery(c)

		// 如果有分页参数或查询条件，使用分页查询
		if c.Query("pageNum") != "" || c.Query("pageSize") != "" ||
			c.Query("name") != "" || c.Query("code") != "" {
			result, err := service.GetProjectsWithPagination(query)
			if err != nil {
				ErrorResponse(c, 500, "获取项目列表失败: "+err.Error())
				return
			}
			SuccessResponse(c, result)
		} else {
			// 否则获取所有项目
			projects, err := service.GetAllProjects()
			if err != nil {
				ErrorResponse(c, 500, "获取项目列表失败: "+err.Error())
				return
			}
			SuccessResponse(c, projects)
		}
	}
}

// @Summary 获取项目详情
// @Description 根据ID获取指定项目的详细信息
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param id path int true "项目ID"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "项目不存在"
// @Router /api/project/{id} [get]
func getProjectHandler(service *database.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		project, err := service.GetProjectByID(uint(id))
		if err != nil {
			ErrorResponse(c, 404, "项目不存在: "+err.Error())
			return
		}

		SuccessResponse(c, project)
	}
}

// @Summary 更新项目
// @Description 根据ID更新指定项目的信息
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param id path int true "项目ID"
// @Param request body models.ProjectUpdateRequest true "更新信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "项目不存在"
// @Failure 500 {object} models.Response "更新失败"
// @Router /api/project/{id} [put]
func updateProjectHandler(service *database.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		var req models.ProjectUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 构建更新字段
		updates := make(map[string]interface{})
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Code != "" {
			updates["code"] = req.Code
		}
		if req.Remark != "" {
			updates["remark"] = req.Remark
		}
		if req.GitRepo != "" {
			updates["git_repo"] = req.GitRepo
		}
		if req.GitUsername != "" {
			updates["git_username"] = req.GitUsername
		}
		// 密码字段直接更新，支持空字符串
		updates["git_password"] = req.GitPassword
		updates["webhook_password"] = req.WebhookPassword

		if err := service.UpdateProject(uint(id), updates); err != nil {
			ErrorResponse(c, 500, "更新项目失败: "+err.Error())
			return
		}

		SuccessResponse(c, nil)
	}
}

// @Summary 删除项目
// @Description 根据ID删除指定项目（软删除）
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param id path int true "项目ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "项目不存在"
// @Failure 500 {object} models.Response "删除失败"
// @Router /api/project/{id} [delete]
func deleteProjectHandler(service *database.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		if err := service.DeleteProject(uint(id)); err != nil {
			ErrorResponse(c, 500, "删除项目失败: "+err.Error())
			return
		}

		SuccessResponse(c, nil)
	}
}

// @Summary 根据项目编码获取项目
// @Description 根据项目编码获取项目信息
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param code path string true "项目编码"
// @Success 200 {object} models.Response "获取成功"
// @Failure 404 {object} models.Response "项目不存在"
// @Router /api/project/code/{code} [get]
func getProjectByCodeHandler(service *database.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		if code == "" {
			ErrorResponse(c, 400, "项目编码不能为空")
			return
		}

		project, err := service.GetProjectByCode(code)
		if err != nil {
			ErrorResponse(c, 404, "项目不存在: "+err.Error())
			return
		}

		SuccessResponse(c, project)
	}
}

// @Summary 检查项目编码是否存在
// @Description 检查项目编码是否已存在，用于新增和更新时的唯一性验证
// @Tags 项目管理
// @Accept json
// @Produce json
// @Param code query string true "项目编码"
// @Param excludeId query int false "排除的项目ID（用于更新时检查）"
// @Success 200 {object} models.Response "检查完成"
// @Failure 400 {object} models.Response "请求参数错误"
// @Router /api/project/check-code [get]
func checkProjectCodeHandler(service *database.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			ErrorResponse(c, 400, "项目编码不能为空")
			return
		}

		var excludeID uint
		if excludeIdStr := c.Query("excludeId"); excludeIdStr != "" {
			if id, err := strconv.ParseUint(excludeIdStr, 10, 32); err == nil {
				excludeID = uint(id)
			}
		}

		exists, err := service.CheckProjectCodeExists(code, excludeID)
		if err != nil {
			ErrorResponse(c, 500, "检查项目编码失败: "+err.Error())
			return
		}

		SuccessResponse(c, gin.H{
			"exists":    exists,
			"code":      code,
			"excludeId": excludeID,
		})
	}
}

// 主机相关处理器
// @Summary 创建主机
// @Description 创建新的SSH主机配置
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param host body models.Host true "主机信息"
// @Success 201 {object} models.Response "创建成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host [post]
func createHostHandler(service *database.HostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var host models.Host
		if err := c.ShouldBindJSON(&host); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		if err := service.CreateHost(&host); err != nil {
			ErrorResponse(c, 500, "创建主机失败: "+err.Error())
			return
		}

		SuccessResponse(c, host)
	}
}

// @Summary 分页获取主机列表
// @Description 分页获取主机配置列表，支持pageNum和pageSize参数，支持主机名模糊查询、IP精确查询、状态精确查询
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param name query string false "主机名模糊查询"
// @Param host query string false "IP地址精确查询"
// @Param status query string false "主机状态精确查询"
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host [get]
func getAllHostsHandler(service *database.HostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有分页参数或查询条件
		query := ParseHostQuery(c)

		// 如果有分页参数或查询条件，使用分页查询
		if c.Query("pageNum") != "" || c.Query("pageSize") != "" ||
			c.Query("name") != "" || c.Query("host") != "" || c.Query("status") != "" {
			result, err := service.GetHostsWithPagination(query)
			if err != nil {
				ErrorResponse(c, 500, "获取主机列表失败: "+err.Error())
				return
			}
			SuccessResponse(c, result)
		} else {
			// 否则获取所有主机
			hosts, err := service.GetAllHosts()
			if err != nil {
				ErrorResponse(c, 500, "获取主机列表失败: "+err.Error())
				return
			}
			SuccessResponse(c, hosts)
		}
	}
}

// @Summary 获取主机详情
// @Description 根据ID获取指定主机的详细信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "主机不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host/{id} [get]
func getHostHandler(service *database.HostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		host, err := service.GetHostByID(uint(id))
		if err != nil {
			ErrorResponse(c, 404, "主机不存在: "+err.Error())
			return
		}

		SuccessResponse(c, host)
	}
}

// @Summary 更新主机信息
// @Description 更新指定主机的配置信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Param host body map[string]interface{} true "更新的主机信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} models.Response "ID格式错误或请求参数错误"
// @Failure 404 {object} models.Response "主机不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host/{id} [put]
func updateHostHandler(service *database.HostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		if err := service.UpdateHost(uint(id), updates); err != nil {
			ErrorResponse(c, 500, "更新主机失败: "+err.Error())
			return
		}

		SuccessResponse(c, nil)
	}
}

// @Summary 删除主机
// @Description 根据ID删除指定主机（软删除）
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "主机不存在"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/host/{id} [delete]
func deleteHostHandler(service *database.HostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		if err := service.DeleteHost(uint(id)); err != nil {
			ErrorResponse(c, 500, "删除主机失败: "+err.Error())
			return
		}

		SuccessResponse(c, nil)
	}
}

// SSH相关处理器
// @Summary 执行SSH命令
// @Description 在指定主机上执行SSH命令
// @Tags SSH操作
// @Accept json
// @Produce json
// @Param request body models.SSHCommandRequest true "命令执行请求"
// @Success 200 {object} models.Response "执行成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "执行失败"
// @Router /api/host/execute [post]
func executeCommandHandler(sshService *database.SSHService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SSHCommandRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 执行命令
		result, err := sshService.ExecuteCommand(req.HostID, req.Command, req.Timeout)
		if err != nil {
			ErrorResponse(c, 500, "执行命令失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 测试SSH连接
// @Description 测试指定主机的SSH连接是否正常
// @Tags SSH操作
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} models.Response "连接成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 500 {object} models.Response "连接失败"
// @Router /api/host/{id}/test [post]
func testConnectionHandler(sshService *database.SSHService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		if err := sshService.TestConnection(uint(id)); err != nil {
			ErrorResponse(c, 500, "连接测试失败: "+err.Error())
			return
		}

		SuccessResponse(c, gin.H{
			"host_id":   id,
			"status":    "success",
			"message":   "SSH连接测试成功",
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		})
	}
}

// @Summary 批量检查主机状态
// @Description 批量检查所有主机的SSH连接状态并更新
// @Tags SSH操作
// @Accept json
// @Produce json
// @Success 200 {object} models.Response "检查完成"
// @Failure 500 {object} models.Response "检查失败"
// @Router /api/host/batch-check [post]
func batchCheckStatusHandler(sshService *database.SSHService) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostService := database.NewHostService()

		// 获取所有主机
		hosts, err := hostService.GetAllHosts()
		if err != nil {
			ErrorResponse(c, 500, "获取主机列表失败: "+err.Error())
			return
		}

		if len(hosts) == 0 {
			SuccessResponse(c, gin.H{
				"message":       "没有需要检查的主机",
				"checked_count": 0,
			})
			return
		}

		// 批量检查状态
		err = sshService.BatchCheckHostStatus(hosts)
		if err != nil {
			ErrorResponse(c, 500, "批量检查主机状态失败: "+err.Error())
			return
		}

		SuccessResponse(c, gin.H{
			"message":       "主机状态检查完成",
			"checked_count": len(hosts),
			"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
		})
	}
}

// 文件上传相关处理器
// @Summary 上传单个文件
// @Description 通过SSH上传单个文件到指定主机
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param request body models.FileUploadRequest true "文件上传请求"
// @Success 200 {object} models.Response "上传成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "上传失败"
// @Router /api/host/upload/file [post]
func uploadFileHandler(sshService *database.SSHService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.FileUploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 上传文件
		result, err := sshService.UploadFile(req.HostID, req.RemotePath, req.Content, req.Permissions, req.Overwrite)
		if err != nil {
			ErrorResponse(c, 500, "文件上传失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 上传目录
// @Description 通过SSH上传整个目录到指定主机
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param request body models.DirectoryUploadRequest true "目录上传请求"
// @Success 200 {object} models.Response "上传成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "上传失败"
// @Router /api/host/upload/directory [post]
func uploadDirectoryHandler(sshService *database.SSHService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.DirectoryUploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 上传目录
		result, err := sshService.UploadDirectory(req.HostID, req.RemotePath, req.Files, req.Overwrite)
		if err != nil {
			ErrorResponse(c, 500, "目录上传失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// Docker相关处理器
// @Summary 获取Docker信息
// @Description 获取指定主机的Docker信息和状态
// @Tags Docker管理
// @Accept json
// @Produce json
// @Param request body models.DockerInfoRequest true "Docker信息请求"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/host/docker/info [post]
func dockerInfoHandler(dockerService *database.DockerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.DockerInfoRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 获取Docker信息
		result, err := dockerService.GetDockerInfo(req.HostID)
		if err != nil {
			ErrorResponse(c, 500, "获取Docker信息失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 构建Docker镜像
// @Description 在指定主机上构建Docker镜像
// @Tags Docker管理
// @Accept json
// @Produce json
// @Param request body models.DockerBuildRequest true "Docker构建请求"
// @Success 200 {object} models.Response "构建成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "构建失败"
// @Router /api/host/docker/build [post]
func dockerBuildHandler(dockerService *database.DockerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.DockerBuildRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 构建镜像
		result, err := dockerService.BuildDockerImage(req.HostID, &req)
		if err != nil {
			ErrorResponse(c, 500, "构建Docker镜像失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 运行Docker容器
// @Description 在指定主机上运行Docker容器
// @Tags Docker管理
// @Accept json
// @Produce json
// @Param request body models.DockerRunRequest true "Docker运行请求"
// @Success 200 {object} models.Response "运行成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "运行失败"
// @Router /api/host/docker/run [post]
func dockerRunHandler(dockerService *database.DockerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.DockerRunRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 运行容器
		result, err := dockerService.RunDockerContainer(req.HostID, &req)
		if err != nil {
			ErrorResponse(c, 500, "运行Docker容器失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 通过SSH执行Docker命令
// @Description 通过SSH连接在指定主机上执行Docker命令
// @Tags Docker管理
// @Accept json
// @Produce json
// @Param request body models.SSHCommandRequest true "Docker命令执行请求"
// @Success 200 {object} models.Response "执行成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "执行失败"
// @Router /api/host/docker/execute [post]
func dockerExecuteHandler(dockerService *database.DockerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SSHCommandRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 通过SSH执行Docker命令
		result, err := dockerService.ExecuteDockerCommandViaSSH(req.HostID, req.Command, req.Timeout)
		if err != nil {
			ErrorResponse(c, 500, "执行Docker命令失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 健康检查
// @Description 检查服务运行状态
// @Tags 心跳接口
// @Accept json
// @Produce json
// @Success 200 {object} models.Response "服务正常"
// @Router /health [get]
func healthHandler(c *gin.Context) {
	SuccessResponse(c, struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "ok",
		Message: "DevOps自动化部署平台运行正常",
	})
}

// ParseDeployConfigQuery 解析部署配置查询参数
func ParseDeployConfigQuery(c *gin.Context) *models.DeployConfigQuery {
	var query models.DeployConfigQuery

	// 解析分页参数
	query.PageNum, _ = strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100 // 限制最大页面大小
	}

	// 解析查询条件
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		if id, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			query.ProjectID = uint(id)
		}
	}
	query.Branch = c.Query("branch")

	return &query
}

// @Summary 创建部署配置
// @Description 为指定项目创建新的部署配置，每个项目的每个分支只能有一个配置
// @Tags 部署配置管理
// @Accept json
// @Produce json
// @Param request body models.DeployConfigCreateRequest true "部署配置信息"
// @Success 201 {object} models.Response "创建成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "创建失败"
// @Router /api/deploy-config [post]
func createDeployConfigHandler(service *database.DeployConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.DeployConfigCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 转换为部署配置模型
		config := &models.DeployConfig{
			ProjectID: req.ProjectID,
			Branch:    req.Branch,
		}

		if err := service.CreateDeployConfig(config); err != nil {
			ErrorResponse(c, 500, "创建部署配置失败: "+err.Error())
			return
		}

		// 更新配置内容
		updates := make(map[string]interface{})
		updates["config"] = req.Config
		if err := service.UpdateDeployConfig(config.ID, updates); err != nil {
			ErrorResponse(c, 500, "更新配置内容失败: "+err.Error())
			return
		}

		// 获取创建后的配置详情
		result, err := service.GetDeployConfigByID(config.ID)
		if err != nil {
			ErrorResponse(c, 500, "获取创建后的配置失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 获取部署配置详情
// @Description 根据ID获取指定部署配置的详细信息
// @Tags 部署配置管理
// @Accept json
// @Produce json
// @Param id path int true "部署配置ID"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "部署配置不存在"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-config/{id} [get]
func getDeployConfigHandler(service *database.DeployConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		config, err := service.GetDeployConfigByID(uint(id))
		if err != nil {
			ErrorResponse(c, 404, "部署配置不存在: "+err.Error())
			return
		}

		SuccessResponse(c, config)
	}
}

// @Summary 获取部署配置列表
// @Description 分页获取部署配置列表，支持按项目ID和分支名称过滤
// @Tags 部署配置管理
// @Accept json
// @Produce json
// @Param pageNum query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param project_id query int false "项目ID精确查询"
// @Param branch query string false "分支名称模糊查询"
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-config [get]
func getDeployConfigsHandler(service *database.DeployConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := ParseDeployConfigQuery(c)

		result, err := service.GetDeployConfigsWithPagination(query)
		if err != nil {
			ErrorResponse(c, 500, "获取部署配置列表失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 根据项目ID获取部署配置列表
// @Description 获取指定项目的所有部署配置
// @Tags 部署配置管理
// @Accept json
// @Produce json
// @Param project_id path int true "项目ID"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-config/project/{project_id} [get]
func getDeployConfigsByProjectHandler(service *database.DeployConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectIDStr := c.Param("project_id")
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "项目ID格式错误: ID必须是数字")
			return
		}

		configs, err := service.GetDeployConfigsByProjectID(uint(projectID))
		if err != nil {
			ErrorResponse(c, 500, "获取项目部署配置失败: "+err.Error())
			return
		}

		SuccessResponse(c, configs)
	}
}

// @Summary 根据项目和分支获取部署配置
// @Description 根据项目ID和分支名称获取特定的部署配置
// @Tags 部署配置管理
// @Accept json
// @Produce json
// @Param project_id path int true "项目ID"
// @Param branch path string true "分支名称"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "参数格式错误"
// @Failure 404 {object} models.Response "部署配置不存在"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-config/project/{project_id}/branch/{branch} [get]
func getDeployConfigByProjectAndBranchHandler(service *database.DeployConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectIDStr := c.Param("project_id")
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "项目ID格式错误: ID必须是数字")
			return
		}

		branch := c.Param("branch")
		if branch == "" {
			ErrorResponse(c, 400, "分支名称不能为空")
			return
		}

		config, err := service.GetDeployConfigByProjectAndBranch(uint(projectID), branch)
		if err != nil {
			ErrorResponse(c, 404, "部署配置不存在: "+err.Error())
			return
		}

		SuccessResponse(c, config)
	}
}

// @Summary 更新部署配置
// @Description 根据ID更新指定部署配置的信息
// @Tags 部署配置管理
// @Accept json
// @Produce json
// @Param id path int true "部署配置ID"
// @Param request body models.DeployConfigUpdateRequest true "更新信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "部署配置不存在"
// @Failure 500 {object} models.Response "更新失败"
// @Router /api/deploy-config/{id} [put]
func updateDeployConfigHandler(service *database.DeployConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		var req models.DeployConfigUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 构建更新字段
		updates := make(map[string]interface{})
		if req.Branch != "" {
			updates["branch"] = req.Branch
		}
		if req.Config != nil {
			updates["config"] = req.Config
		}

		if err := service.UpdateDeployConfig(uint(id), updates); err != nil {
			ErrorResponse(c, 500, "更新部署配置失败: "+err.Error())
			return
		}

		// 获取更新后的配置详情
		result, err := service.GetDeployConfigByID(uint(id))
		if err != nil {
			ErrorResponse(c, 500, "获取更新后的配置失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 删除部署配置
// @Description 根据ID删除指定部署配置（软删除）
// @Tags 部署配置管理
// @Accept json
// @Produce json
// @Param id path int true "部署配置ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "部署配置不存在"
// @Failure 500 {object} models.Response "删除失败"
// @Router /api/deploy-config/{id} [delete]
func deleteDeployConfigHandler(service *database.DeployConfigService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		// 检查配置是否存在
		_, err = service.GetDeployConfigByID(uint(id))
		if err != nil {
			ErrorResponse(c, 404, "部署配置不存在: "+err.Error())
			return
		}

		if err := service.DeleteDeployConfig(uint(id)); err != nil {
			ErrorResponse(c, 500, "删除部署配置失败: "+err.Error())
			return
		}

		SuccessResponse(c, gin.H{
			"id":      id,
			"message": "部署配置删除成功",
		})
	}
}

// ParseDeployRecordQuery 解析部署记录查询参数
func ParseDeployRecordQuery(c *gin.Context) *models.DeployRecordQuery {
	var query models.DeployRecordQuery

	// 解析分页参数
	query.PageNum, _ = strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100 // 限制最大页面大小
	}

	// 解析查询条件
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		if id, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			query.ProjectID = uint(id)
		}
	}
	query.ProjectName = c.Query("project_name")
	query.Branch = c.Query("branch")
	query.Status = c.Query("status")
	query.StartTimeStart = c.Query("start_time_start")
	query.StartTimeEnd = c.Query("start_time_end")

	return &query
}

// @Summary 创建部署记录
// @Description 创建新的部署记录
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param request body models.DeployRecordCreateRequest true "部署记录信息"
// @Success 201 {object} models.Response "创建成功"
// @Failure 400 {object} models.Response "请求参数错误"
// @Failure 500 {object} models.Response "创建失败"
// @Router /api/deploy-record [post]
func createDeployRecordHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.DeployRecordCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 解析开始时间
		startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err != nil {
			ErrorResponse(c, 400, "开始时间格式错误，应为 yyyy-MM-dd HH:mm:ss")
			return
		}

		// 转换为部署记录模型
		record := &models.DeployRecord{
			ProjectID:   req.ProjectID,
			ProjectName: req.ProjectName,
			Branch:      req.Branch,
			StartTime:   models.CustomTime{Time: startTime},
			Duration:    req.Duration,
			LogPath:     req.LogPath,
			Status:      req.Status,
		}

		if err := service.CreateDeployRecord(record); err != nil {
			ErrorResponse(c, 500, "创建部署记录失败: "+err.Error())
			return
		}

		// 获取创建后的记录详情
		result, err := service.GetDeployRecordByID(record.ID)
		if err != nil {
			ErrorResponse(c, 500, "获取创建后的记录失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 获取部署记录详情
// @Description 根据ID获取指定部署记录的详细信息
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param id path int true "部署记录ID"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "部署记录不存在"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-record/{id} [get]
func getDeployRecordHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		record, err := service.GetDeployRecordByID(uint(id))
		if err != nil {
			ErrorResponse(c, 404, "部署记录不存在: "+err.Error())
			return
		}

		SuccessResponse(c, record)
	}
}

// @Summary 获取部署记录列表
// @Description 分页获取部署记录列表，支持按项目ID、项目名称、分支、状态、时间范围过滤
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param pageNum query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param project_id query int false "项目ID精确查询"
// @Param project_name query string false "项目名称模糊查询"
// @Param branch query string false "分支模糊查询"
// @Param status query string false "状态精确查询"
// @Param start_time_start query string false "开始时间范围查询-开始"
// @Param start_time_end query string false "开始时间范围查询-结束"
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-record [get]
func getDeployRecordsHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := ParseDeployRecordQuery(c)

		result, err := service.GetDeployRecordsWithPagination(query)
		if err != nil {
			ErrorResponse(c, 500, "获取部署记录列表失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 更新部署记录
// @Description 根据ID更新指定部署记录的信息
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param id path int true "部署记录ID"
// @Param request body models.DeployRecordUpdateRequest true "更新信息"
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "部署记录不存在"
// @Failure 500 {object} models.Response "更新失败"
// @Router /api/deploy-record/{id} [put]
func updateDeployRecordHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		var req models.DeployRecordUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, 400, "请求参数错误: "+err.Error())
			return
		}

		// 构建更新字段
		updates := make(map[string]interface{})
		if req.ProjectID > 0 {
			updates["project_id"] = req.ProjectID
		}
		if req.ProjectName != "" {
			updates["project_name"] = req.ProjectName
		}
		if req.Branch != "" {
			updates["branch"] = req.Branch
		}
		if req.StartTime != "" {
			updates["start_time"] = req.StartTime
		}
		if req.Duration > 0 {
			updates["duration"] = req.Duration
		}
		if req.LogPath != "" {
			updates["log_path"] = req.LogPath
		}
		if req.Status != "" {
			updates["status"] = req.Status
		}

		if err := service.UpdateDeployRecord(uint(id), updates); err != nil {
			ErrorResponse(c, 500, "更新部署记录失败: "+err.Error())
			return
		}

		// 获取更新后的记录详情
		result, err := service.GetDeployRecordByID(uint(id))
		if err != nil {
			ErrorResponse(c, 500, "获取更新后的记录失败: "+err.Error())
			return
		}

		SuccessResponse(c, result)
	}
}

// @Summary 删除部署记录
// @Description 根据ID删除指定部署记录（软删除）
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param id path int true "部署记录ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 404 {object} models.Response "部署记录不存在"
// @Failure 500 {object} models.Response "删除失败"
// @Router /api/deploy-record/{id} [delete]
func deleteDeployRecordHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "ID格式错误: ID必须是数字")
			return
		}

		// 检查记录是否存在
		_, err = service.GetDeployRecordByID(uint(id))
		if err != nil {
			ErrorResponse(c, 404, "部署记录不存在: "+err.Error())
			return
		}

		if err := service.DeleteDeployRecord(uint(id)); err != nil {
			ErrorResponse(c, 500, "删除部署记录失败: "+err.Error())
			return
		}

		SuccessResponse(c, gin.H{
			"id":      id,
			"message": "部署记录删除成功",
		})
	}
}

// @Summary 根据项目ID获取部署记录列表
// @Description 获取指定项目的所有部署记录
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param project_id path int true "项目ID"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "ID格式错误"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-record/project/{project_id} [get]
func getDeployRecordsByProjectHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectIDStr := c.Param("project_id")
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "项目ID格式错误: ID必须是数字")
			return
		}

		records, err := service.GetDeployRecordsByProjectID(uint(projectID))
		if err != nil {
			ErrorResponse(c, 500, "获取项目部署记录失败: "+err.Error())
			return
		}

		SuccessResponse(c, records)
	}
}

// @Summary 根据分支获取部署记录列表
// @Description 获取指定分支的所有部署记录
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param branch path string true "分支名称"
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-record/branch/{branch} [get]
func getDeployRecordsByBranchHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		branch := c.Param("branch")
		if branch == "" {
			ErrorResponse(c, 400, "分支名称不能为空")
			return
		}

		records, err := service.GetDeployRecordsByBranch(branch)
		if err != nil {
			ErrorResponse(c, 500, "获取分支部署记录失败: "+err.Error())
			return
		}

		SuccessResponse(c, records)
	}
}

// @Summary 根据状态获取部署记录列表
// @Description 获取指定状态的所有部署记录
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param status path string true "状态"
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-record/status/{status} [get]
func getDeployRecordsByStatusHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Param("status")
		if status == "" {
			ErrorResponse(c, 400, "状态不能为空")
			return
		}

		records, err := service.GetDeployRecordsByStatus(status)
		if err != nil {
			ErrorResponse(c, 500, "获取状态部署记录失败: "+err.Error())
			return
		}

		SuccessResponse(c, records)
	}
}

// @Summary 获取项目和分支的最新部署记录
// @Description 获取指定项目和分支的最新部署记录
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param project_id path int true "项目ID"
// @Param branch path string true "分支名称"
// @Success 200 {object} models.Response "获取成功"
// @Failure 400 {object} models.Response "参数格式错误"
// @Failure 404 {object} models.Response "部署记录不存在"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-record/project/{project_id}/branch/{branch}/latest [get]
func getLatestDeployRecordByProjectAndBranchHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectIDStr := c.Param("project_id")
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			ErrorResponse(c, 400, "项目ID格式错误: ID必须是数字")
			return
		}

		branch := c.Param("branch")
		if branch == "" {
			ErrorResponse(c, 400, "分支名称不能为空")
			return
		}

		record, err := service.GetLatestDeployRecordByProjectAndBranch(uint(projectID), branch)
		if err != nil {
			ErrorResponse(c, 404, "部署记录不存在: "+err.Error())
			return
		}

		SuccessResponse(c, record)
	}
}

// @Summary 获取部署记录统计信息
// @Description 获取部署记录的统计信息，包括总次数、成功次数、失败次数、运行中次数
// @Tags 部署记录管理
// @Accept json
// @Produce json
// @Param project_id query int false "项目ID（可选，不传则统计所有项目）"
// @Success 200 {object} models.Response "获取成功"
// @Failure 500 {object} models.Response "获取失败"
// @Router /api/deploy-record/stats [get]
func getDeployRecordStatsHandler(service *database.DeployRecordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var projectID uint
		if projectIDStr := c.Query("project_id"); projectIDStr != "" {
			if id, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
				projectID = uint(id)
			}
		}

		stats, err := service.GetDeployRecordStats(projectID)
		if err != nil {
			ErrorResponse(c, 500, "获取部署记录统计失败: "+err.Error())
			return
		}

		SuccessResponse(c, stats)
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
}

// loginHandler 处理登录请求
// @Summary 用户登录
// @Description 用户登录认证
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} models.Response{data=LoginResponse} "登录成功"
// @Failure 401 {object} models.Response "登录失败"
// @Router /auth/login [post]
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, 400, "请求参数错误: "+err.Error())
		return
	}

	// 简单的用户名密码验证
	if req.Username == "admin" && req.Password == "admin" {
		// 生成简单的token（实际项目中应该使用JWT）
		token := "mock-jwt-token-" + time.Now().Format("20060102150405")

		loginResp := LoginResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   3600, // 1小时
		}

		SuccessResponse(c, loginResp)
		return
	}

	ErrorResponse(c, 401, "用户名或密码错误")
}

// logoutHandler 处理登出请求
// @Summary 用户登出
// @Description 用户登出
// @Tags 认证管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response "登出成功"
// @Router /auth/logout [post]
func logoutHandler(c *gin.Context) {
	SuccessResponse(c, gin.H{"message": "登出成功"})
}

// refreshTokenHandler 刷新token
// @Summary 刷新访问令牌
// @Description 刷新用户的访问令牌
// @Tags 认证管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=string} "刷新成功"
// @Router /auth/refresh [post]
func refreshTokenHandler(c *gin.Context) {
	newToken := "refreshed-jwt-token-" + time.Now().Format("20060102150405")
	SuccessResponse(c, newToken)
}

// getAccessCodesHandler 获取用户权限码
// @Summary 获取用户权限码
// @Description 获取当前用户的权限码列表
// @Tags 认证管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]string} "获取成功"
// @Router /auth/codes [get]
func getAccessCodesHandler(c *gin.Context) {
	// 返回模拟的权限码
	codes := []string{
		"devops:host:view",
		"devops:host:create",
		"devops:host:update",
		"devops:host:delete",
		"devops:project:view",
		"devops:project:create",
		"devops:project:update",
		"devops:project:delete",
		"devops:deploy:view",
		"devops:deploy:create",
	}
	SuccessResponse(c, codes)
}

// UserInfo 用户信息结构
type UserInfo struct {
	ID       uint     `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Avatar   string   `json:"avatar"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// getUserInfoHandler 获取当前用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=UserInfo} "获取成功"
// @Failure 401 {object} models.Response "未授权"
// @Router /user/info [get]
func getUserInfoHandler(c *gin.Context) {
	userInfo := UserInfo{
		ID:       1,
		Username: "admin",
		Nickname: "管理员",
		Avatar:   "",
		Email:    "admin@devops.com",
		Roles:    []string{"admin"},
	}
	SuccessResponse(c, userInfo)
}
