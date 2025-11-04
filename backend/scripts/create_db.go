package main

import (
	"DevOpsProject/internal/database"
	"fmt"
	"log"
)

func main() {
	fmt.Println("=== 创建MySQL数据库 ===")

	// 连接MySQL服务器（不指定数据库）
	config := database.Config{
		Type:     "mysql",
		Host:     "localhost",
		Port:     "3306",
		Database: "", // 不指定数据库
		Username: "root",
		Password: "P@ssw0rd",
	}

	if err := database.InitDatabase(config); err != nil {
		log.Fatal("连接MySQL服务器失败:", err)
	}

	// 创建数据库
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Fatal("获取数据库连接失败:", err)
	}

	// 创建devops_platform数据库
	_, err = sqlDB.Exec("CREATE DATABASE IF NOT EXISTS devops_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	if err != nil {
		log.Fatal("创建数据库失败:", err)
	}

	fmt.Println("✅ 数据库 devops_platform 创建成功！")

	// 关闭连接
	sqlDB.Close()
	database.CloseDatabase()
	fmt.Println("=== 数据库创建完成 ===")
}