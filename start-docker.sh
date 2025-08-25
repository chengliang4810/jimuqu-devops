#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}    Jimuqu DevOps 平台 - Docker 一键启动脚本${NC}"
echo -e "${BLUE}================================================${NC}"
echo

# 检查Docker是否运行
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}[错误] Docker 未运行或未安装${NC}"
    echo "请确保 Docker 已正确安装并启动"
    exit 1
fi

echo -e "${GREEN}[信息] Docker 运行正常${NC}"

# 检查docker-compose是否存在
if ! command -v docker compose >/dev/null 2>&1; then
    echo -e "${RED}[错误] docker compose 命令不可用${NC}"
    echo "请安装 Docker Compose"
    exit 1
fi

echo -e "${GREEN}[信息] Docker Compose 可用${NC}"

show_menu() {
    echo
    echo -e "${YELLOW}[选择] 请选择操作:${NC}"
    echo "1. 首次启动（构建并启动所有服务）"
    echo "2. 正常启动（启动现有服务）"
    echo "3. 重新构建并启动"
    echo "4. 停止所有服务"
    echo "5. 查看服务状态"
    echo "6. 查看日志"
    echo "7. 清理所有数据（危险操作）"
    echo "0. 退出"
    echo
}

first_start() {
    echo
    echo -e "${BLUE}[信息] 首次启动 - 构建并启动所有服务...${NC}"
    echo -e "${YELLOW}[警告] 首次构建可能需要较长时间，请耐心等待${NC}"
    echo
    
    if docker compose up --build -d; then
        echo
        echo -e "${GREEN}[成功] 服务启动成功！${NC}"
        echo
        echo -e "${BLUE}访问地址:${NC}"
        echo -e "  前端界面: ${GREEN}http://localhost${NC}"
        echo -e "  后端API:  ${GREEN}http://localhost:8080${NC}"
        echo -e "  数据库:   ${GREEN}localhost:3306${NC}"
        echo
        echo -e "${BLUE}默认数据库信息:${NC}"
        echo -e "  数据库: ${GREEN}jimuqu_devops${NC}"
        echo -e "  用户名: ${GREEN}devops${NC}"
        echo -e "  密码:   ${GREEN}devops123${NC}"
    else
        echo -e "${RED}[错误] 服务启动失败，请检查日志${NC}"
    fi
}

normal_start() {
    echo
    echo -e "${BLUE}[信息] 正常启动所有服务...${NC}"
    
    if docker compose up -d; then
        echo -e "${GREEN}[成功] 服务启动成功！${NC}"
        echo -e "访问地址: ${GREEN}http://localhost${NC}"
    else
        echo -e "${RED}[错误] 服务启动失败${NC}"
    fi
}

rebuild_start() {
    echo
    echo -e "${BLUE}[信息] 重新构建并启动...${NC}"
    
    docker compose down
    docker compose build --no-cache
    
    if docker compose up -d; then
        echo -e "${GREEN}[成功] 服务重新构建并启动成功！${NC}"
    else
        echo -e "${RED}[错误] 重新构建失败${NC}"
    fi
}

stop_services() {
    echo
    echo -e "${BLUE}[信息] 停止所有服务...${NC}"
    docker compose down
    echo -e "${GREEN}[完成] 所有服务已停止${NC}"
}

show_status() {
    echo
    echo -e "${BLUE}[信息] 服务状态:${NC}"
    docker compose ps
    echo
    echo -e "${BLUE}[信息] 容器状态:${NC}"
    docker ps -a --filter "label=com.docker.compose.project=jimuqu-devops"
}

show_logs() {
    echo
    echo -e "${BLUE}[信息] 显示最近日志 (按 Ctrl+C 退出):${NC}"
    docker compose logs -f --tail=50
}

cleanup() {
    echo
    echo -e "${RED}[警告] 这将删除所有数据，包括数据库数据！${NC}"
    read -p "确定要继续吗？(输入 YES 确认): " confirm
    
    if [ "$confirm" != "YES" ]; then
        echo -e "${YELLOW}[取消] 操作已取消${NC}"
        return
    fi
    
    echo
    echo -e "${BLUE}[信息] 停止并删除所有服务和数据...${NC}"
    docker compose down -v --remove-orphans
    docker system prune -f --volumes
    echo -e "${GREEN}[完成] 清理完成${NC}"
}

show_tips() {
    echo
    echo -e "${BLUE}[提示] 常用命令:${NC}"
    echo -e "  查看日志: ${GREEN}docker compose logs -f [服务名]${NC}"
    echo -e "  进入容器: ${GREEN}docker compose exec [服务名] /bin/bash${NC}"
    echo -e "  重启服务: ${GREEN}docker compose restart [服务名]${NC}"
    echo
}

# 主循环
while true; do
    show_menu
    read -p "请输入选择 (0-7): " choice
    
    case $choice in
        1)
            first_start
            show_tips
            ;;
        2)
            normal_start
            show_tips
            ;;
        3)
            rebuild_start
            show_tips
            ;;
        4)
            stop_services
            ;;
        5)
            show_status
            ;;
        6)
            show_logs
            ;;
        7)
            cleanup
            ;;
        0)
            echo -e "${GREEN}[退出] 再见！${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}[错误] 无效选择，请重新输入${NC}"
            ;;
    esac
    
    echo
    read -p "按回车键继续..." dummy
done