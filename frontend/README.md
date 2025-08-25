# Jimuqu DevOps 前端项目说明

## 项目概述

基于Vue3 + TypeScript + NaiveUI + Vite构建的现代化DevOps管理后台，提供直观的用户界面管理Docker容器化部署平台。

## 技术栈

- **前端框架**: Vue 3.5.17
- **开发语言**: TypeScript 5.8.3
- **构建工具**: Vite 7.0.4
- **UI组件库**: Naive UI 2.42.0
- **状态管理**: Pinia 3.0.3
- **路由管理**: Vue Router 4.5.1
- **图表库**: ECharts 5.6.0
- **图标库**: Iconify
- **CSS框架**: UnoCSS
- **包管理器**: pnpm 10.5.0+

## 功能特性

### 🏠 仪表盘
- 系统状态总览
- 统计数据展示（应用数量、构建次数、成功率等）
- 构建趋势图表
- 最近构建列表
- 主机状态监控
- 快速操作入口

### 🖥️ 主机管理
- 主机信息的增删改查
- SSH连接配置和测试
- 批量主机状态更新
- 主机状态实时监控
- 支持密码和私钥认证

### 📦 应用管理
- 应用配置管理
- Git仓库集成
- 环境变量配置
- Webhook配置
- 自动/手动触发设置
- 一键构建部署

### 🔨 构建管理
- 构建历史查看
- 实时构建日志
- 构建步骤详情
- 构建状态监控
- 失败构建分析

## 页面结构

```
src/views/devops/
├── dashboard/          # 仪表盘
│   └── index.vue
├── host/              # 主机管理
│   └── index.vue  
├── application/       # 应用管理
│   └── index.vue
└── build/             # 构建管理
    └── index.vue
```

## 快速开始

### 环境要求
- Node.js 20.19.0+
- pnpm 10.5.0+

### 安装依赖
```bash
cd frontend
pnpm install
```

### 启动开发服务器
```bash
pnpm dev
```

### 构建生产版本
```bash
pnpm build
```

### 预览生产构建
```bash
pnpm preview
```

## 项目配置

### 环境变量
```bash
# 开发环境
VITE_API_BASE_URL=http://localhost:8080/api

# 生产环境  
VITE_API_BASE_URL=https://your-domain.com/api
```

### 代理配置
开发环境下，Vite会自动代理API请求到后端服务器：

```typescript
// vite.config.ts
export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
```

## 页面功能详解

### 仪表盘页面
- **统计卡片**: 显示应用总数、构建次数、成功率、在线主机数
- **趋势图表**: 最近7天的构建趋势线图
- **状态分布**: 构建状态的饼图分析
- **最近构建**: 最新的构建记录表格
- **主机状态**: 实时显示各主机连接状态
- **快速操作**: 常用功能的快捷入口

### 主机管理页面
- **主机列表**: 分页显示所有配置的主机
- **搜索过滤**: 支持按主机名称和IP搜索
- **添加主机**: 表单录入主机连接信息
- **编辑主机**: 修改主机配置
- **连接测试**: 实时测试SSH连接状态
- **批量操作**: 批量更新主机状态

### 应用管理页面
- **应用列表**: 显示所有配置的应用项目
- **Git配置**: 设置仓库地址、分支、认证信息
- **环境变量**: 动态配置构建环境变量
- **Webhook设置**: 配置自动触发构建
- **一键构建**: 手动触发应用构建
- **流水线配置**: 跳转到流水线配置页面

### 构建管理页面
- **构建列表**: 显示所有构建历史记录
- **状态过滤**: 按构建状态筛选记录
- **日志查看**: 实时查看构建详细日志
- **步骤追踪**: 查看每个构建步骤的执行情况
- **操作管理**: 取消正在运行的构建

## 组件特性

### 响应式设计
- 支持桌面端和移动端
- 自适应布局和组件
- 触摸友好的交互

### 实时更新
- 构建状态实时刷新
- 主机连接状态监控
- 自动刷新机制

### 用户体验
- 加载状态指示
- 错误提示和处理
- 操作确认对话框
- 友好的表单验证

## API接口对接

前端页面通过Fetch API与后端服务通信：

```typescript
// 主机管理API示例
const getHosts = async (page: number, size: number, keyword?: string) => {
  const params = new URLSearchParams({
    page: page.toString(),
    size: size.toString()
  });
  if (keyword) params.append('keyword', keyword);
  
  const response = await fetch(`/api/hosts?${params}`);
  return await response.json();
};

// 触发构建API示例  
const triggerBuild = async (applicationId: number) => {
  const response = await fetch(`/api/applications/${applicationId}/build`, {
    method: 'POST'
  });
  return await response.json();
};
```

## 开发指南

### 添加新页面
1. 在`src/views/devops/`下创建新的页面组件
2. 在`src/router/elegant/routes.ts`中添加路由配置
3. 在`src/router/elegant/imports.ts`中添加组件导入
4. 更新导航菜单配置

### 样式开发
- 使用UnoCSS原子化CSS类
- 组件内样式使用scoped
- 遵循NaiveUI设计规范

### 状态管理
- 使用Pinia进行全局状态管理
- 页面级状态使用Vue3 Composition API

## 部署说明

### 构建优化
```bash
# 生产构建
pnpm build

# 分析构建包大小
pnpm preview --host 0.0.0.0
```

### Nginx配置
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        root /var/www/devops-frontend;
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 故障排除

### 常见问题
1. **依赖安装失败**: 检查Node.js和pnpm版本
2. **API请求失败**: 确认后端服务是否启动
3. **路由不工作**: 检查路由配置和组件导入
4. **样式异常**: 清除浏览器缓存，重新构建

### 开发工具
- Vue DevTools: 调试Vue组件和状态
- Network面板: 检查API请求
- Console: 查看错误信息和日志

## 更新日志

### v1.0.0 (2024-01-15)
- ✨ 完成DevOps平台前端初始版本
- 🎉 实现仪表盘、主机管理、应用管理、构建管理页面
- 🚀 集成NaiveUI组件库和ECharts图表
- 📱 支持响应式设计和移动端适配
- 🔧 配置完整的开发和构建环境