#!/bin/bash

echo "=================================="
echo "  DevOps平台开发模式启动脚本"
echo "=================================="
echo ""

# 检查Node.js
if ! command -v node &> /dev/null; then
    echo "❌ 错误：未检测到Node.js，请先安装Node.js 20+"
    echo "下载地址：https://nodejs.org/"
    exit 1
fi

NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ $NODE_VERSION -lt 20 ]; then
    echo "❌ 错误：Node.js版本过低，需要20+，当前版本: $(node -v)"
    exit 1
fi

echo "✅ Node.js版本检查通过: $(node -v)"
echo ""

# 检查pnpm
if ! command -v pnpm &> /dev/null; then
    echo "⚠️  未检测到pnpm，正在安装..."
    npm install -g pnpm
    echo "✅ pnpm安装完成"
else
    echo "✅ pnpm已安装: $(pnpm -v)"
fi
echo ""

# 启动后端服务
echo "🚀 启动后端服务..."
docker-compose up -d redis

# 等待Redis启动
sleep 3
echo "✅ 后端服务已启动"
echo ""

# 安装前端依赖
echo "📦 安装前端依赖..."
cd frontend
if [ ! -d "node_modules" ]; then
    pnpm install
else
    echo "✅ 依赖已存在，跳过安装"
fi
echo ""

# 启动前端开发服务器
echo "🎨 启动前端开发服务器..."
echo ""
echo "=================================="
echo "  开发服务器启动成功！"
echo "=================================="
echo ""
echo "前端地址: http://localhost:5173"
echo "后端地址: http://localhost:8000"
echo "API文档:  http://localhost:8000/docs"
echo ""
echo "按 Ctrl+C 停止前端服务器"
echo "重新运行此脚本启动前端"
echo "=================================="
echo ""

pnpm dev
