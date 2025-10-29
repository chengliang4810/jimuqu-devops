#!/bin/bash

echo "创建前端目录结构..."

# 创建目录
mkdir -p frontend/src/{router,stores,utils,views/devops/{dashboard,projects,deployments/components}}

echo "前端目录创建完成"
echo ""
echo "需要手动创建的文件："
echo "1. frontend/src/router/index.ts - 主路由"
echo "2. frontend/src/router/devops.ts - DevOps路由"
echo "3. frontend/src/stores/projects.ts - 项目状态管理"
echo "4. frontend/src/stores/deployments.ts - 部署状态管理"
echo "5. frontend/src/stores/dashboard.ts - 仪表盘状态管理"
echo "6. frontend/src/utils/api.ts - API客户端"
echo "7. frontend/src/views/devops/dashboard/index.vue - 仪表盘页面"
echo "8. frontend/src/views/devops/projects/index.vue - 项目列表页面"
echo "9. frontend/src/views/devops/projects/components/ProjectModal.vue - 项目模态框"
echo "10. frontend/src/views/devops/deployments/index.vue - 部署列表页面"
echo "11. frontend/src/views/devops/deployments/components/LogModal.vue - 日志模态框"
