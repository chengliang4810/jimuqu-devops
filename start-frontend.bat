@echo off
chcp 65001 > nul

echo === Jimuqu DevOps 前端项目启动脚本 ===

REM 检查是否安装了pnpm
pnpm --version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到pnpm，请先安装pnpm
    echo 安装命令: npm install -g pnpm
    pause
    exit /b 1
)

REM 进入前端目录
cd /d "%~dp0frontend"

echo 正在安装依赖...
pnpm install

if %errorlevel% neq 0 (
    echo 依赖安装失败
    pause
    exit /b 1
)

echo 正在启动开发服务器...
pnpm dev

echo 前端服务已启动，请在浏览器中访问: http://localhost:3000
pause