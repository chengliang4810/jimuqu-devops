#!/bin/bash

echo "=== Jimuqu DevOps 前端项目启动脚本 ==="

# 检查是否安装了pnpm
if ! command -v pnpm &> /dev/null
then
    echo "错误: 未找到pnpm，请先安装pnpm"
    echo "安装命令: npm install -g pnpm"
    exit 1
fi

# 进入前端目录
cd "$(dirname "$0")/frontend"

echo "正在安装依赖..."
pnpm install

echo "正在启动开发服务器..."
pnpm dev

echo "前端服务已启动，请在浏览器中访问: http://localhost:3000"