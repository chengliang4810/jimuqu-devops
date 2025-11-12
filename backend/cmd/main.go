package main

import (
	"DevOpsProject/backend/internal/database"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "DevOpsProject/backend/docs"
)

// @title DevOps主机管理API
// @version 1.0
// @description DevOps自动化部署平台 - 主机管理API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8089
// @BasePath /

func main() {
	// 初始化数据库 - 使用MySQL
	config := database.Config{
		Type:     "mysql",
		Host:     "localhost",
		Port:     "3306",
		Database: "devops_platform",
		Username: "root",
		Password: "P@ssw0rd",
	}

	if err := database.InitDatabase(config); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 自动迁移数据库表
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 初始化默认数据
	if err := database.InitDefaultData(); err != nil {
		log.Fatal("初始化默认数据失败:", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建服务实例
	hostService := database.NewHostService()
	sshService := database.NewSSHService()
	dockerService := database.NewDockerService()
	projectService := database.NewProjectService()
	deployConfigService := database.NewDeployConfigService()
	deployRecordService := database.NewDeployRecordService()

	// 项目相关API
	projectGroup := r.Group("/api/project")
	{
		projectGroup.POST("", createProjectHandler(projectService))
		projectGroup.GET("", getAllProjectsHandler(projectService))
		projectGroup.GET("/:id", getProjectHandler(projectService))
		projectGroup.PUT("/:id", updateProjectHandler(projectService))
		projectGroup.DELETE("/:id", deleteProjectHandler(projectService))
		projectGroup.GET("/code/:code", getProjectByCodeHandler(projectService))
		projectGroup.GET("/check-code", checkProjectCodeHandler(projectService))
	}

	// 部署配置相关API
	deployConfigGroup := r.Group("/api/deploy-config")
	{
		deployConfigGroup.POST("", createDeployConfigHandler(deployConfigService))
		deployConfigGroup.GET("", getDeployConfigsHandler(deployConfigService))
		deployConfigGroup.GET("/:id", getDeployConfigHandler(deployConfigService))
		deployConfigGroup.PUT("/:id", updateDeployConfigHandler(deployConfigService))
		deployConfigGroup.DELETE("/:id", deleteDeployConfigHandler(deployConfigService))
		deployConfigGroup.GET("/project/:project_id", getDeployConfigsByProjectHandler(deployConfigService))
		deployConfigGroup.GET("/project/:project_id/branch/:branch", getDeployConfigByProjectAndBranchHandler(deployConfigService))
	}

	// 部署记录相关API
	deployRecordGroup := r.Group("/api/deploy-record")
	{
		deployRecordGroup.POST("", createDeployRecordHandler(deployRecordService))
		deployRecordGroup.GET("", getDeployRecordsHandler(deployRecordService))
		deployRecordGroup.GET("/:id", getDeployRecordHandler(deployRecordService))
		deployRecordGroup.PUT("/:id", updateDeployRecordHandler(deployRecordService))
		deployRecordGroup.DELETE("/:id", deleteDeployRecordHandler(deployRecordService))
		deployRecordGroup.GET("/project/:project_id", getDeployRecordsByProjectHandler(deployRecordService))
		deployRecordGroup.GET("/branch/:branch", getDeployRecordsByBranchHandler(deployRecordService))
		deployRecordGroup.GET("/status/:status", getDeployRecordsByStatusHandler(deployRecordService))
		deployRecordGroup.GET("/project/:project_id/branch/:branch/latest", getLatestDeployRecordByProjectAndBranchHandler(deployRecordService))
		deployRecordGroup.GET("/stats", getDeployRecordStatsHandler(deployRecordService))
	}

	// 主机相关API
	hostGroup := r.Group("/api/host")
	{
		hostGroup.POST("", createHostHandler(hostService))
		hostGroup.GET("", getAllHostsHandler(hostService))
		hostGroup.GET("/:id", getHostHandler(hostService))
		hostGroup.PUT("/:id", updateHostHandler(hostService))
		hostGroup.DELETE("/:id", deleteHostHandler(hostService))

		// SSH相关API
		hostGroup.POST("/execute", executeCommandHandler(sshService))
		hostGroup.POST("/:id/test", testConnectionHandler(sshService))
		hostGroup.POST("/batch-check", batchCheckStatusHandler(sshService))

		// 文件上传相关API
		hostGroup.POST("/upload/file", uploadFileHandler(sshService))
		hostGroup.POST("/upload/directory", uploadDirectoryHandler(sshService))

		// Docker相关API
		hostGroup.POST("/docker/info", dockerInfoHandler(dockerService))
		hostGroup.POST("/docker/build", dockerBuildHandler(dockerService))
		hostGroup.POST("/docker/run", dockerRunHandler(dockerService))
		hostGroup.POST("/docker/execute", dockerExecuteHandler(dockerService))
	}

	// 认证相关API
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", loginHandler)
		authGroup.POST("/logout", logoutHandler)
		authGroup.POST("/refresh", refreshTokenHandler)
		authGroup.GET("/codes", getAccessCodesHandler)
	}

	// 用户相关API
	userGroup := r.Group("/user")
	{
		userGroup.GET("/info", getUserInfoHandler)
	}

	// 健康检查
	r.GET("/health", healthHandler)

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	fmt.Println("DevOps主机管理平台启动在端口 :8089")
	fmt.Println("Swagger文档地址: http://localhost:8089/swagger/index.html")
	log.Fatal(r.Run(":8089"))
}