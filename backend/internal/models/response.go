package models

import (
	"database/sql/driver"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// CustomTime 自定义时间类型，用于格式化时间输出
type CustomTime struct {
	time.Time
}

// MarshalJSON 实现JSON序列化，输出 yyyy-MM-dd hh:MM:ss 格式
func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct.IsZero() {
		return []byte("null"), nil
	}
	formatted := fmt.Sprintf(`"%s"`, ct.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

// UnmarshalJSON 实现JSON反序列化
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ct.Time = time.Time{}
		return nil
	}
	// 去掉引号
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// 尝试解析 yyyy-MM-dd hh:MM:ss 格式
	t, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		// 如果失败，尝试解析其他常见格式
		t, err = time.Parse(time.RFC3339, str)
		if err != nil {
			return err
		}
	}
	ct.Time = t
	return nil
}

// Value 实现 driver.Valuer 接口，用于数据库存储
func (ct CustomTime) Value() (driver.Value, error) {
	if ct.IsZero() {
		return nil, nil // 返回 NULL 而不是零值时间
	}
	return ct.Time, nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		ct.Time = t
		return nil
	}
	return fmt.Errorf("cannot convert %v to CustomTime", value)
}

// GormDataType gorm数据类型
func (ct CustomTime) GormDataType() string {
	return "datetime"
}

// Host 主机配置表
type Host struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null;size:255"`
	Host      string         `json:"host" gorm:"not null;size:255"`
	Port      int            `json:"port" gorm:"default:22"`
	Username  string         `json:"username" gorm:"not null;size:255"`
	Password  string         `json:"password" gorm:"not null;size:255"`
	AuthType  string         `json:"auth_type" gorm:"default:password;size:50"` // password, key
	Status    string         `json:"status" gorm:"default:offline;size:50"`   // online, offline
	Remark    string         `json:"remark" gorm:"size:500"`                  // 备注信息
	CreatedAt CustomTime    `json:"created_at"`
	UpdatedAt CustomTime    `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code" example:"200"`           // 状态码：200成功，其他失败
	Message string      `json:"message" example:"操作成功"`    // 消息：成功或失败原因
	Data    interface{} `json:"data,omitempty"`              // 数据：具体返回数据
}

// PageData 分页数据结构
type PageData struct {
	Rows  interface{} `json:"rows"`   // 分页数据项
	Total int64       `json:"total"`  // 数据总条数
}

// PaginationQuery 分页查询参数
type PaginationQuery struct {
	PageNum  int `json:"pageNum" form:"pageNum" example:"1"`      // 页码，从1开始
	PageSize int `json:"pageSize" form:"pageSize" example:"10"`    // 每页条数
}

// HostQuery 主机查询条件
type HostQuery struct {
	PaginationQuery                // 分页参数
	Name             string `json:"name" form:"name" example:"测试主机"`           // 主机名模糊查询
	Host             string `json:"host" form:"host" example:"192.168.1.100"`       // IP地址精确查询
	Status           string `json:"status" form:"status" example:"online"`         // 主机状态精确查询：online/offline
}

// PageResult 分页结果
type PageResult struct {
	List  interface{} `json:"list"`   // 数据列表
	Total int64       `json:"total"`  // 总数
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message" example:"DevOps自动化部署平台运行正常"`
}

// SSHCommandRequest SSH命令执行请求
type SSHCommandRequest struct {
	HostID uint   `json:"host_id" binding:"required" example:"1"`      // 主机ID
	Command string `json:"command" binding:"required" example:"echo 'hello word'"` // 要执行的命令
	Timeout int    `json:"timeout" example:"30"`                         // 超时时间（秒），默认30秒
}

// SSHCommandResponse SSH命令执行响应
type SSHCommandResponse struct {
	HostID    uint   `json:"host_id" example:"1"`              // 主机ID
	Command   string `json:"command" example:"echo 'hello word'"` // 执行的命令
	Output    string `json:"output" example:"hello word"`      // 命令输出
	Error     string `json:"error,omitempty"`                  // 错误信息
	ExitCode  int    `json:"exit_code" example:"0"`           // 退出码
	Duration  int64  `json:"duration" example:"1500"`         // 执行耗时（毫秒）
	StartTime string `json:"start_time" example:"2024-01-01 12:00:00"` // 开始时间
	EndTime   string `json:"end_time" example:"2024-01-01 12:00:01"`   // 结束时间
}

// FileUploadRequest 文件上传请求
type FileUploadRequest struct {
	HostID      uint   `json:"host_id" binding:"required" example:"1"`      // 主机ID
	RemotePath  string `json:"remote_path" binding:"required" example:"/root/test.txt"` // 远程路径
	Content     string `json:"content" binding:"required" example:"Hello World"`        // 文件内容
	Permissions string `json:"permissions" example:"0644"`           // 文件权限（可选）
	Overwrite   bool   `json:"overwrite" example:"true"`               // 是否覆盖已存在的文件
}

// FileUploadResponse 文件上传响应
type FileUploadResponse struct {
	HostID      uint   `json:"host_id" example:"1"`              // 主机ID
	RemotePath  string `json:"remote_path" example:"/root/test.txt"` // 远程路径
	Size        int64  `json:"size" example:"1024"`              // 文件大小（字节）
	Permissions string `json:"permissions" example:"0644"`       // 实际文件权限
	Duration    int64  `json:"duration" example:"500"`           // 上传耗时（毫秒）
	StartTime   string `json:"start_time" example:"2024-01-01 12:00:00"` // 开始时间
	EndTime     string `json:"end_time" example:"2024-01-01 12:00:01"`   // 结束时间
	Message     string `json:"message" example:"文件上传成功"`       // 状态消息
}

// DirectoryUploadRequest 目录上传请求
type DirectoryUploadRequest struct {
	HostID     uint                `json:"host_id" binding:"required" example:"1"`      // 主机ID
	RemotePath string              `json:"remote_path" binding:"required" example:"/root/test_dir"` // 远程目录路径
	Files      []DirectoryFileItem `json:"files" binding:"required"`                    // 文件列表
	Overwrite  bool                `json:"overwrite" example:"true"`                   // 是否覆盖已存在的文件
}

// DirectoryFileItem 目录文件项
type DirectoryFileItem struct {
	Path        string `json:"path" binding:"required" example:"subdir/file.txt"`        // 相对路径
	Content     string `json:"content" binding:"required" example:"Hello World"`        // 文件内容
	Permissions string `json:"permissions" example:"0644"`       // 文件权限（可选）
	IsDir       bool   `json:"is_dir" example:"false"`            // 是否为目录
}

// DirectoryUploadResponse 目录上传响应
type DirectoryUploadResponse struct {
	HostID       uint                     `json:"host_id" example:"1"`              // 主机ID
	RemotePath   string                   `json:"remote_path" example:"/root/test_dir"` // 远程目录路径
	TotalFiles   int                      `json:"total_files" example:"5"`           // 总文件数
	SuccessFiles int                      `json:"success_files" example:"5"`         // 成功上传文件数
	FailedFiles  []DirectoryUploadError   `json:"failed_files,omitempty"`           // 失败文件列表
	Duration     int64                    `json:"duration" example:"2000"`          // 上传耗时（毫秒）
	StartTime    string                   `json:"start_time" example:"2024-01-01 12:00:00"` // 开始时间
	EndTime      string                   `json:"end_time" example:"2024-01-01 12:00:02"`   // 结束时间
	Message      string                   `json:"message" example:"目录上传完成"`       // 状态消息
}

// DirectoryUploadError 目录上传错误
type DirectoryUploadError struct {
	Path   string `json:"path" example:"subdir/file.txt"` // 文件路径
	Error  string `json:"error" example:"Permission denied"` // 错误信息
}

// DockerBuildRequest Docker编译请求
type DockerBuildRequest struct {
	HostID      uint   `json:"host_id" binding:"required" example:"1"`      // 主机ID
	ImageName   string `json:"image_name" binding:"required" example:"devops-app"` // 镜像名称
	ImageTag    string `json:"image_tag" example:"latest"`                    // 镜像标签
	Dockerfile  string `json:"dockerfile" example:"Dockerfile"`              // Dockerfile路径
	ContextPath string `json:"context_path" example:"."`                     // 构建上下文路径
	BuildArgs   map[string]string `json:"build_args,omitempty"`               // 构建参数
	Timeout     int    `json:"timeout" example:"300"`                         // 超时时间（秒）
}

// DockerBuildResponse Docker编译响应
type DockerBuildResponse struct {
	HostID    uint   `json:"host_id" example:"1"`              // 主机ID
	ImageID   string `json:"image_id" example:"sha256:abc123"` // 镜像ID
	ImageName string `json:"image_name" example:"devops-app"`   // 镜像名称
	ImageTag  string `json:"image_tag" example:"latest"`       // 镜像标签
	Size      int64  `json:"size" example:"125829120"`          // 镜像大小（字节）
	Duration  int64  `json:"duration" example:"45000"`          // 编译耗时（毫秒）
	StartTime string `json:"start_time" example:"2024-01-01 12:00:00"` // 开始时间
	EndTime   string `json:"end_time" example:"2024-01-01 12:00:45"`   // 结束时间
	Message   string `json:"message" example:"Docker镜像编译成功"`   // 状态消息
	Logs      []string `json:"logs,omitempty"`                    // 编译日志
}

// DockerRunRequest Docker运行请求
type DockerRunRequest struct {
	HostID      uint                   `json:"host_id" binding:"required" example:"1"` // 主机ID
	ImageName   string                 `json:"image_name" binding:"required" example:"devops-app"` // 镜像名称
	ImageTag    string                 `json:"image_tag" example:"latest"`       // 镜像标签
	ContainerName string               `json:"container_name" example:"test-container"` // 容器名称
	Command     string                 `json:"command" example:"./app"`           // 运行命令
	EnvVars     map[string]string      `json:"env_vars,omitempty"`             // 环境变量
	PortBindings map[string]string      `json:"port_bindings,omitempty"`       // 端口绑定
	Volumes     []DockerVolumeMapping   `json:"volumes,omitempty"`              // 卷映射
	Detach      bool                   `json:"detach" example:"true"`            // 是否后台运行
	Timeout     int                    `json:"timeout" example:"60"`             // 超时时间（秒）
}

// DockerVolumeMapping Docker卷映射
type DockerVolumeMapping struct {
	HostPath      string `json:"host_path" example:"/host/path"`      // 主机路径
	ContainerPath string `json:"container_path" example:"/container/path"` // 容器路径
	Mode          string `json:"mode" example:"rw"`                   // 读写模式：rw/ro
}

// DockerRunResponse Docker运行响应
type DockerRunResponse struct {
	HostID        uint                   `json:"host_id" example:"1"`              // 主机ID
	ContainerID   string                 `json:"container_id" example:"abc123def"`  // 容器ID
	ContainerName string                 `json:"container_name" example:"test-container"` // 容器名称
	ImageName     string                 `json:"image_name" example:"devops-app"`   // 镜像名称
	ImageTag      string                 `json:"image_tag" example:"latest"`       // 镜像标签
	Status        string                 `json:"status" example:"running"`         // 容器状态
	Command       string                 `json:"command" example:"./app"`         // 执行命令
	ExitCode      int                    `json:"exit_code" example:"0"`           // 退出码
	Duration      int64                  `json:"duration" example:"1500"`         // 执行耗时（毫秒）
	StartTime     string                 `json:"start_time" example:"2024-01-01 12:00:00"` // 开始时间
	EndTime       string                 `json:"end_time" example:"2024-01-01 12:00:02"`   // 结束时间
	Message       string                 `json:"message" example:"容器运行成功"`     // 状态消息
	Logs          []string               `json:"logs,omitempty"`                  // 容器日志
	PortBindings  map[string]string      `json:"port_bindings,omitempty"`       // 实际端口绑定
}

// DockerInfoRequest Docker信息请求
type DockerInfoRequest struct {
	HostID uint `json:"host_id" binding:"required" example:"1"` // 主机ID
}

// DockerInfoResponse Docker信息响应
type DockerInfoResponse struct {
	HostID          uint                   `json:"host_id" example:"1"`        // 主机ID
	Version         string                 `json:"version" example:"24.0.6"`   // Docker版本
	Architecture    string                 `json:"architecture" example:"x86_64"` // 架构
	NCPU            int                    `json:"ncpu" example:"8"`            // CPU核心数
	MemTotal        int64                  `json:"mem_total" example:"16777216000"` // 总内存
	ImagesCount     int                    `json:"images_count" example:"25"`   // 镜像数量
	ContainersCount int                    `json:"containers_count" example:"15"` // 容器数量
	RunningCount    int                    `json:"running_count" example:"3"`    // 运行中容器数
	Message         string                 `json:"message" example:"Docker连接正常"` // 状态消息
	Duration        int64                  `json:"duration" example:"500"`       // 查询耗时（毫秒）
	Timestamp       string                 `json:"timestamp" example:"2024-01-01 12:00:00"` // 查询时间
}

// Project 项目模型
type Project struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:255"`        // 项目名称
	Code        string         `json:"code" gorm:"not null;uniqueIndex;size:100"` // 项目编码（唯一）
	Remark      string         `json:"remark" gorm:"size:1000"`               // 备注
	GitRepo     string         `json:"git_repo" gorm:"size:500"`               // Git仓库地址
	GitUsername string         `json:"git_username" gorm:"size:255"`           // Git用户名
	GitPassword string         `json:"git_password" gorm:"size:255"`           // Git密码
	CreatedAt   CustomTime     `json:"created_at"`                             // 创建时间
	UpdatedAt   CustomTime     `json:"updated_at"`                             // 更新时间
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`                         // 软删除
}

// ProjectQuery 项目查询条件
type ProjectQuery struct {
	PaginationQuery                // 分页参数
	Name             string `json:"name" form:"name" example:"项目名称"`           // 项目名称模糊查询
	Code             string `json:"code" form:"code" example:"PROJECT001"`       // 项目编码模糊查询
}

// ProjectCreateRequest 创建项目请求
type ProjectCreateRequest struct {
	Name        string `json:"name" binding:"required" example:"我的项目"`        // 项目名称
	Code        string `json:"code" binding:"required" example:"PROJECT001"`    // 项目编码
	Remark      string `json:"remark" example:"项目备注信息"`                     // 备注
	GitRepo     string `json:"git_repo" example:"https://github.com/user/repo.git"` // Git仓库地址
	GitUsername string `json:"git_username" example:"gituser"`                 // Git用户名
	GitPassword string `json:"git_password" example:"gitpassword"`             // Git密码
}

// ProjectUpdateRequest 更新项目请求
type ProjectUpdateRequest struct {
	Name        string `json:"name" example:"我的项目"`                          // 项目名称
	Code        string `json:"code" example:"PROJECT001"`                      // 项目编码
	Remark      string `json:"remark" example:"项目备注信息"`                     // 备注
	GitRepo     string `json:"git_repo" example:"https://github.com/user/repo.git"` // Git仓库地址
	GitUsername string `json:"git_username" example:"gituser"`                 // Git用户名
	GitPassword string `json:"git_password" example:"gitpassword"`             // Git密码
}

// DeployConfig 部署配置表
type DeployConfig struct {
	ID        uint        `json:"id" gorm:"primaryKey"`                              // 主键
	ProjectID uint        `json:"project_id" gorm:"not null;index"`              // 项目ID
	Branch    string      `json:"branch" gorm:"not null;size:255"`                // 分支
	Config    string      `json:"config" gorm:"type:json"`                           // 配置内容（JSON数组）
	CreatedAt CustomTime `json:"created_at"`                                    // 创建时间
	UpdatedAt CustomTime `json:"updated_at"`                                    // 更新时间
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                            // 软删除
}

// DeployConfigCreateRequest 创建部署配置请求
type DeployConfigCreateRequest struct {
	ProjectID uint                   `json:"project_id" binding:"required" example:"1"`     // 项目ID
	Branch    string                 `json:"branch" binding:"required" example:"main"` // 分支
	Config    []DeployConfigItem     `json:"config" binding:"required"`                // 配置内容数组
}

// DeployConfigUpdateRequest 更新部署配置请求
type DeployConfigUpdateRequest struct {
	Branch string             `json:"branch" example:"main"`                 // 分支
	Config []DeployConfigItem `json:"config" example:"[{\"env\":\"prod\"}]"` // 配置内容数组
}

// DeployConfigItem 部署配置项
type DeployConfigItem struct {
	Key   string      `json:"key" example:"DATABASE_URL"`     // 配置键
	Value interface{} `json:"value" example:"localhost:5432"` // 配置值
	Desc  string      `json:"desc" example:"数据库连接地址"`  // 配置描述
}

// DeployConfigQuery 部署配置查询条件
type DeployConfigQuery struct {
	PaginationQuery                      // 分页参数
	ProjectID    uint   `json:"project_id" form:"project_id" example:"1"`    // 项目ID精确查询
	Branch        string `json:"branch" form:"branch" example:"main"`        // 分支模糊查询
}

// DeployConfigResponse 部署配置响应
type DeployConfigResponse struct {
	ID        uint                `json:"id" example:"1"`              // 主键
	ProjectID uint                `json:"project_id" example:"1"`        // 项目ID
	Branch    string              `json:"branch" example:"main"`        // 分支
	Config    []DeployConfigItem   `json:"config" example:"[{\"env\":\"prod\"}]"` // 配置内容数组
	CreatedAt string              `json:"created_at" example:"2024-01-01 12:00:00"` // 创建时间
	UpdatedAt string              `json:"updated_at" example:"2024-01-01 12:30:00"` // 更新时间
}

// DeployRecord 部署记录表
type DeployRecord struct {
	ID          uint        `json:"id" gorm:"primaryKey"`                    // 主键
	ProjectID   uint        `json:"project_id" gorm:"not null;index"`        // 项目ID
	ProjectName string      `json:"project_name" gorm:"not null;size:255"`   // 项目名称
	Branch      string      `json:"branch" gorm:"not null;size:255"`          // 分支
	StartTime   CustomTime  `json:"start_time" gorm:"not null"`               // 开始执行时间
	Duration    int64       `json:"duration" gorm:"default:0"`                // 耗时（秒）
	LogPath     string      `json:"log_path" gorm:"size:500"`                 // 日志文件路径
	Status      string      `json:"status" gorm:"default:running;size:50"`    // 部署状态：running, success, failed
	CreatedAt   CustomTime  `json:"created_at"`                                 // 创建时间
	UpdatedAt   CustomTime  `json:"updated_at"`                                 // 更新时间
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`                        // 软删除
}

// DeployRecordCreateRequest 创建部署记录请求
type DeployRecordCreateRequest struct {
	ProjectID   uint   `json:"project_id" binding:"required" example:"1"`     // 项目ID
	ProjectName string `json:"project_name" binding:"required" example:"我的项目"` // 项目名称
	Branch      string `json:"branch" binding:"required" example:"main"`     // 分支
	StartTime   string `json:"start_time" binding:"required" example:"2024-01-01 12:00:00"` // 开始执行时间
	Duration    int64  `json:"duration" example:"300"`                         // 耗时（秒）
	LogPath     string `json:"log_path" example:"/var/log/deploy/20240101_120000.log"` // 日志文件路径
	Status      string `json:"status" example:"running"`                       // 部署状态：running, success, failed
}

// DeployRecordUpdateRequest 更新部署记录请求
type DeployRecordUpdateRequest struct {
	ProjectID   uint   `json:"project_id" example:"1"`                        // 项目ID
	ProjectName string `json:"project_name" example:"我的项目"`                 // 项目名称
	Branch      string `json:"branch" example:"main"`                         // 分支
	StartTime   string `json:"start_time" example:"2024-01-01 12:00:00"`     // 开始执行时间
	Duration    int64  `json:"duration" example:"300"`                         // 耗时（秒）
	LogPath     string `json:"log_path" example:"/var/log/deploy/20240101_120000.log"` // 日志文件路径
	Status      string `json:"status" example:"success"`                       // 部署状态：running, success, failed
}

// DeployRecordQuery 部署记录查询条件
type DeployRecordQuery struct {
	PaginationQuery                        // 分页参数
	ProjectID      uint   `json:"project_id" form:"project_id" example:"1"`    // 项目ID精确查询
	ProjectName    string `json:"project_name" form:"project_name" example:"我的项目"` // 项目名称模糊查询
	Branch         string `json:"branch" form:"branch" example:"main"`        // 分支模糊查询
	Status         string `json:"status" form:"status" example:"success"`     // 状态精确查询：running, success, failed
	StartTimeStart string `json:"start_time_start" form:"start_time_start" example:"2024-01-01 00:00:00"` // 开始时间范围查询-开始
	StartTimeEnd   string `json:"start_time_end" form:"start_time_end" example:"2024-01-31 23:59:59"`     // 开始时间范围查询-结束
}

// DeployRecordResponse 部署记录响应
type DeployRecordResponse struct {
	ID          uint   `json:"id" example:"1"`                        // 主键
	ProjectID   uint   `json:"project_id" example:"1"`                // 项目ID
	ProjectName string `json:"project_name" example:"我的项目"`         // 项目名称
	Branch      string `json:"branch" example:"main"`                 // 分支
	StartTime   string `json:"start_time" example:"2024-01-01 12:00:00"` // 开始执行时间
	Duration    int64  `json:"duration" example:"300"`                 // 耗时（秒）
	LogPath     string `json:"log_path" example:"/var/log/deploy/20240101_120000.log"` // 日志文件路径
	Status      string `json:"status" example:"success"`               // 部署状态：running, success, failed
	CreatedAt   string `json:"created_at" example:"2024-01-01 12:00:00"` // 创建时间
	UpdatedAt   string `json:"updated_at" example:"2024-01-01 12:05:00"` // 更新时间
}