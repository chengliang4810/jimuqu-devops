package database

import (
	"DevOpsProject/backend/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
	_ "github.com/go-sql-driver/mysql"
)

var DB *gorm.DB

// Config 数据库配置
type Config struct {
	Type     string // sqlite, mysql
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

// InitDatabase 初始化数据库连接
func InitDatabase(config Config) error {
	var err error

	// 设置日志级别
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	switch config.Type {
	case "sqlite":
		// SQLite连接 - 使用modernc.org/sqlite纯Go驱动
		dsn := config.Database
		if dsn == "" {
			dsn = "./devops.db"
		}

		// 创建自定义Dialector，明确使用modernc驱动
		DB, err = gorm.Open(&sqlite.Dialector{
			DSN: dsn,
		}, gormConfig)
	case "mysql":
		// MySQL连接 - 先连接到MySQL服务器，创建数据库，然后连接到数据库
		// 首先连接到MySQL服务器（不指定数据库）
		serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port)

		serverDB, err := gorm.Open(mysql.Open(serverDSN), gormConfig)
		if err != nil {
			return fmt.Errorf("连接MySQL服务器失败: %v", err)
		}

		// 创建数据库（如果不存在）
		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", config.Database)
		if err := serverDB.Exec(createDBSQL).Error; err != nil {
			return fmt.Errorf("创建数据库失败: %v", err)
		}

		// 关闭服务器连接
		sqlServerDB, _ := serverDB.DB()
		sqlServerDB.Close()

		// 连接到指定的数据库
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database)
		DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 获取底层的sql.DB对象进行连接池配置
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)   // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)  // 最大打开连接数
	sqlDB.SetConnMaxLifetime(3600) // 连接最大生存时间

	log.Printf("数据库连接成功，类型: %s", config.Type)
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	// 对于SQLite，使用更安全的方式迁移
	if DB.Dialector.Name() == "sqlite" {
		// 先检查并手动创建表结构
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("获取数据库连接失败: %v", err)
		}

		// 设置SQLite参数
		sqlDB.Exec("PRAGMA foreign_keys = OFF")
		sqlDB.Exec("PRAGMA journal_mode = WAL")
		sqlDB.Exec("PRAGMA synchronous = NORMAL")

		// 手动创建表
		tables := []string{
			`CREATE TABLE IF NOT EXISTS hosts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				host TEXT NOT NULL,
				port INTEGER DEFAULT 22,
				username TEXT NOT NULL,
				password TEXT NOT NULL,
				auth_type TEXT DEFAULT 'password',
				status TEXT DEFAULT 'active',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				deleted_at DATETIME
			)`,
			`CREATE TABLE IF NOT EXISTS projects (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				code TEXT NOT NULL UNIQUE,
				remark TEXT,
				git_repo TEXT,
				git_username TEXT,
				git_password TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				deleted_at DATETIME
			)`,
		}

		for _, tableSQL := range tables {
			if _, err := sqlDB.Exec(tableSQL); err != nil {
				log.Printf("创建表失败: %v", err)
				return fmt.Errorf("数据库表创建失败: %v", err)
			}
		}

		// 创建索引
		indexes := []string{
			`CREATE INDEX IF NOT EXISTS idx_hosts_deleted_at ON hosts(deleted_at)`,
			`CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_code ON projects(code)`,
		}

		for _, indexSQL := range indexes {
			if _, err := sqlDB.Exec(indexSQL); err != nil {
				log.Printf("创建索引失败: %v", err)
			}
		}

		// 重新启用外键约束
		sqlDB.Exec("PRAGMA foreign_keys = ON")
	} else {
		// MySQL直接迁移
		err := DB.AutoMigrate(&models.Host{}, &models.Project{}, &models.DeployConfig{}, &models.DeployRecord{})
		if err != nil {
			return fmt.Errorf("数据库迁移失败: %v", err)
		}
	}

	log.Println("数据库表结构迁移完成")
	return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	// 检查是否已有主机数据
	var count int64
	DB.Model(&models.Host{}).Count(&count)
	if count > 0 {
		log.Println("数据库已有数据，跳过初始化")
		return nil
	}

	// 插入测试主机
	testHost := models.Host{
		Name:     "测试主机",
		Host:     "127.0.0.1",
		Port:     22,
		Username: "root",
		Password: "password",
		AuthType: "password",
		Status:   "inactive",
	}
	if err := DB.Create(&testHost).Error; err != nil {
		return fmt.Errorf("插入测试主机失败: %v", err)
	}

	log.Println("默认数据初始化完成")
	return nil
}