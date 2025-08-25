@echo off
chcp 65001 > nul

echo ================================================
echo    Jimuqu DevOps 平台 - Docker 一键启动脚本
echo ================================================
echo.

REM 检查Docker是否运行
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] Docker 未运行或未安装
    echo 请确保 Docker Desktop 已启动
    pause
    exit /b 1
)

echo [信息] Docker 运行正常

REM 检查docker-compose是否存在
docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] docker compose 命令不可用
    echo 请确保使用 Docker Desktop 或安装 docker-compose
    pause
    exit /b 1
)

echo [信息] Docker Compose 可用

echo.
echo [选择] 请选择操作:
echo 1. 首次启动（构建并启动所有服务）
echo 2. 正常启动（启动现有服务）
echo 3. 重新构建并启动
echo 4. 停止所有服务
echo 5. 查看服务状态
echo 6. 查看日志
echo 7. 清理所有数据（危险操作）
echo 0. 退出
echo.

set /p choice=请输入选择 (0-7): 

if "%choice%"=="1" goto first_start
if "%choice%"=="2" goto normal_start
if "%choice%"=="3" goto rebuild_start
if "%choice%"=="4" goto stop_services
if "%choice%"=="5" goto show_status
if "%choice%"=="6" goto show_logs
if "%choice%"=="7" goto cleanup
if "%choice%"=="0" goto exit
goto invalid_choice

:first_start
echo.
echo [信息] 首次启动 - 构建并启动所有服务...
echo [警告] 首次构建可能需要较长时间，请耐心等待
echo.
docker compose up --build -d
if %errorlevel% equ 0 (
    echo.
    echo [成功] 服务启动成功！
    echo.
    echo 访问地址:
    echo   前端界面: http://localhost
    echo   后端API:  http://localhost:8080
    echo   数据库:   localhost:3306
    echo.
    echo 默认数据库信息:
    echo   数据库: jimuqu_devops
    echo   用户名: devops
    echo   密码:   devops123
) else (
    echo [错误] 服务启动失败，请检查日志
)
goto end

:normal_start
echo.
echo [信息] 正常启动所有服务...
docker compose up -d
if %errorlevel% equ 0 (
    echo [成功] 服务启动成功！
    echo 访问地址: http://localhost
) else (
    echo [错误] 服务启动失败
)
goto end

:rebuild_start
echo.
echo [信息] 重新构建并启动...
docker compose down
docker compose build --no-cache
docker compose up -d
if %errorlevel% equ 0 (
    echo [成功] 服务重新构建并启动成功！
) else (
    echo [错误] 重新构建失败
)
goto end

:stop_services
echo.
echo [信息] 停止所有服务...
docker compose down
echo [完成] 所有服务已停止
goto end

:show_status
echo.
echo [信息] 服务状态:
docker compose ps
echo.
echo [信息] 容器状态:
docker ps -a --filter "label=com.docker.compose.project=jimuqu-devops"
goto end

:show_logs
echo.
echo [信息] 显示最近日志 (按 Ctrl+C 退出):
docker compose logs -f --tail=50
goto end

:cleanup
echo.
echo [警告] 这将删除所有数据，包括数据库数据！
set /p confirm=确定要继续吗？(输入 YES 确认): 
if not "%confirm%"=="YES" (
    echo [取消] 操作已取消
    goto end
)
echo.
echo [信息] 停止并删除所有服务和数据...
docker compose down -v --remove-orphans
docker system prune -f --volumes
echo [完成] 清理完成
goto end

:invalid_choice
echo [错误] 无效选择，请重新运行脚本
goto end

:end
echo.
echo [提示] 常用命令:
echo   查看日志: docker compose logs -f [服务名]
echo   进入容器: docker compose exec [服务名] /bin/bash
echo   重启服务: docker compose restart [服务名]
echo.
pause

:exit