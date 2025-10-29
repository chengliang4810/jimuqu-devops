#!/bin/bash

# DevOps部署平台启动脚本

set -e

echo "=================================="
echo "  DevOps部署平台启动脚本"
echo "=================================="
echo ""

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ 错误：未检测到Docker，请先安装Docker"
    exit 1
fi

# 检查Docker Compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ 错误：未检测到Docker Compose，请先安装Docker Compose"
    exit 1
fi

echo "✅ Docker环境检查通过"
echo ""

# 创建必要的目录
echo "📁 创建必要目录..."
mkdir -p logs data
echo "✅ 目录创建完成"
echo ""

# 检查是否存在.env文件
if [ ! -f .env ]; then
    echo "⚠️  未检测到.env文件，复制示例配置..."
    cp .env.example .env
    echo "✅ 已创建.env文件，请根据需要修改配置"
    echo ""
fi

# 构建镜像
echo "🔨 构建Docker镜像..."
docker-compose build
echo "✅ 镜像构建完成"
echo ""

# 启动服务
echo "🚀 启动服务..."
docker-compose up -d
echo ""

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo ""
echo "📊 服务状态："
docker-compose ps
echo ""

# 显示访问地址
echo "✅ 启动完成！"
echo ""
echo "=================================="
echo "  访问地址"
echo "=================================="
echo "🌐 Web界面:     http://localhost:8000"
echo "📚 API文档:     http://localhost:8000/docs"
echo "📊 Flower监控:  http://localhost:5555"
echo ""
echo "默认登录凭据："
echo "  用户名: admin"
echo "  密码:   admin123"
echo ""
echo "=================================="
echo "  常用命令"
echo "=================================="
echo "查看日志:      docker-compose logs -f"
echo "停止服务:      docker-compose down"
echo "重启服务:      docker-compose restart"
echo "查看状态:      docker-compose ps"
echo ""
echo "按 Ctrl+C 退出"
echo "=================================="
