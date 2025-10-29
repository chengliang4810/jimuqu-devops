# DevOps自动化部署平台

一个基于 **FastAPI + Celery + Docker + Vue3** 开发的现代化DevOps自动化部署平台，支持多语言项目的自动化编译、部署和监控。

## ✨ 特性

- 🚀 **多语言支持**：Java、Python、Node.js、Go
- 📦 **容器化编译**：使用Docker隔离编译环境
- 🔄 **Webhook集成**：GitHub/GitLab自动触发部署
- 📊 **实时日志**：WebSocket实时查看部署日志
- 🔐 **简单认证**：基于环境变量的单用户认证
- 📈 **部署统计**：成功率、耗时等数据统计
- 🎨 **现代化UI**：基于Vue3 + Naive UI + TailwindCSS
- 📱 **响应式设计**：支持桌面和移动端

## 🏗️ 架构

```
┌───────────────────────────────────────────────┐
│              前端 (Vue3 + Vite)               │
│        http://localhost:5173                  │
└─────────────────┬─────────────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────────────┐
│              FastAPI 后端服务                  │
│            http://localhost:8000              │
│  - 项目管理API  - 部署记录API                  │
│  - Webhook接收  - WebSocket日志                │
└─────────────────┬─────────────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────────────┐
│            Celery 任务队列 (Redis)              │
│  - 任务调度  - 结果存储                        │
└─────────────────┬─────────────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────────────┐
│            Celery Worker 进程                  │
│  - Docker编译引擎  - SSH部署服务               │
└─────────────────┬─────────────────────────────┘
                  │
        ┌─────────┴─────────┐
        ▼                   ▼
   ┌─────────┐       ┌──────────┐
   │  Docker │       │   SSH    │
   │ (编译)  │       │  (部署)  │
   └─────────┘       └──────────┘
```

## 📋 系统要求

- Docker >= 20.10
- Docker Compose >= 2.0
- Node.js >= 20 (开发前端)
- 目标主机需要支持SSH访问

## 🚀 快速开始

### 方式一：一键启动（推荐）

```bash
# 开发模式（前端 + 后端）
./dev-start.sh

# 生产模式（仅后端）
./start.sh
```

### 方式二：手动启动

#### 启动后端服务
```bash
docker-compose up -d
```

#### 启动前端开发服务器
```bash
cd frontend
pnpm install
pnpm dev
```

## 🌐 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| **前端界面** | http://localhost:5173 | Vue3管理界面 |
| **API服务** | http://localhost:8000 | FastAPI后端 |
| **API文档** | http://localhost:8000/docs | Swagger文档 |
| **Flower监控** | http://localhost:5555 | Celery监控 |
| **生产前端** | http://localhost:8000 | 打包后的前端 |

默认登录凭据：
- 用户名：`admin`
- 密码：`admin123`

## 📱 前端技术栈

基于 **lithe-admin** 开源框架：

- ✅ **Vue 3.5** + Composition API
- ✅ **TypeScript** - 完整类型支持
- ✅ **Vite 7** - 极速构建工具
- ✅ **TailwindCSS 4** - 现代化样式
- ✅ **Naive UI** - Vue3组件库
- ✅ **Pinia** - 状态管理
- ✅ **Vue Router** - 路由管理
- ✅ **ECharts** - 数据图表

## 📚 目录结构

```
devops-platform/
├── app/                    # FastAPI后端
│   ├── api/               # API路由
│   ├── models/            # 数据模型
│   ├── services/          # 业务逻辑
│   ├── main.py            # 应用入口
│   └── websocket.py       # WebSocket处理
├── frontend/              # Vue3前端 (lithe-admin)
│   ├── src/
│   │   ├── views/         # 页面组件
│   │   │   ├── devops/    # DevOps功能页面
│   │   │   ├── sign-in/   # 登录页
│   │   │   └── error-page/# 错误页
│   │   ├── router/        # 路由配置
│   │   ├── stores/        # Pinia状态管理
│   │   ├── utils/         # 工具函数
│   │   └── components/    # 通用组件
│   ├── package.json       # 前端依赖
│   └── vite.config.ts     # Vite配置
├── worker/                # Celery异步任务
│   └── tasks.py           # 部署任务
├── docker-compose.yml     # Docker编排
├── Dockerfile             # 容器镜像
├── requirements.txt       # Python依赖
└── README.md              # 项目文档
```

## 📝 使用说明

### 1. 创建项目

访问 **http://localhost:5173** → 项目管理 → 新建项目

配置信息：
- **项目名称**：唯一标识
- **Git地址**：仓库HTTP(S)地址
- **开发语言**：Java/Python/Node.js/Go
- **部署路径**：目标主机目录
- **目标主机**：服务器IP
- **SSH信息**：登录凭据

### 2. 配置Git Webhook

#### GitHub配置
- **Payload URL**: `http://你的IP:8000/api/webhook/github/1`
- **Content type**: `application/json`
- **Secret**: 项目配置的Webhook密钥
- **Events**: Push events

#### GitLab配置
- **URL**: `http://你的IP:8000/api/webhook/gitlab/1`
- **Secret token**: 项目配置的Webhook密钥
- **Trigger events**: Push events

### 3. 触发部署

**自动触发**：
- 推送代码到Git仓库
- Webhook自动触发部署

**手动触发**：
- 进入部署记录页面
- 点击"执行"按钮

### 4. 查看日志

- 进入部署记录
- 点击"查看日志"
- 实时查看编译和部署过程

## 🔧 开发指南

### 前端开发

```bash
# 启动开发服务器
cd frontend
pnpm dev

# 构建生产版本
pnpm build

# 代码检查
pnpm lint:check
```

详细文档：[FRONTEND_QUICKSTART.md](FRONTEND_QUICKSTART.md)

### 后端开发

```bash
# 安装Python依赖
pip install -r requirements.txt

# 启动Redis
redis-server

# 启动Celery Worker
celery -A worker.tasks worker --loglevel=info

# 启动Web服务
uvicorn app.main:app --reload
```

### API文档

启动服务后访问：http://localhost:8000/docs

主要接口：
- `/api/projects` - 项目管理
- `/api/deployments` - 部署管理
- `/api/webhook/github/{id}` - GitHub Webhook
- `/api/webhook/gitlab/{id}` - GitLab Webhook
- `/ws/deployments/{id}` - 实时日志

## 🔒 安全建议

1. **修改默认密码**：生产环境修改`ADMIN_USERNAME`和`ADMIN_PASSWORD`
2. **使用SSH密钥**：建议使用密钥而非密码连接目标主机
3. **HTTPS部署**：生产环境使用Nginx反向代理并启用HTTPS
4. **限制访问**：配置防火墙仅允许必要端口访问
5. **Webhook安全**：启用Webhook Secret验证

## 🐛 常见问题

### Q: 前端无法访问后端API？

A: 检查后端服务是否启动：
```bash
docker-compose ps
```

### Q: WebSocket连接失败？

A: 检查防火墙设置，确认WebSocket端口开放

### Q: Docker编译失败？

A: 检查Docker daemon是否正常运行，容器是否有足够权限

### Q: SSH连接失败？

A: 确认目标主机SSH服务正常，用户名密码或密钥正确

### Q: 如何查看详细日志？

A: 查看容器日志：
```bash
docker-compose logs -f web
docker-compose logs -f worker
```

## 📄 文档

- [README.md](README.md) - 项目总览
- [QUICKSTART.md](QUICKSTART.md) - 快速开始
- [FRONTEND_QUICKSTART.md](FRONTEND_QUICKSTART.md) - 前端开发指南
- [FRONTEND_INTEGRATION.md](FRONTEND_INTEGRATION.md) - 前端集成详情
- [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) - 项目结构说明
- [SUMMARY.md](SUMMARY.md) - 交付总结

## 🎯 下一步

- [ ] 配置HTTPS访问
- [ ] 添加邮件/钉钉通知
- [ ] 配置多环境部署（dev/staging/prod）
- [ ] 集成堡垒机
- [ ] 添加审批流程
- [ ] 配置自动回滚
- [ ] 添加监控告警

## 📈 性能优化

### 前端优化
- 路由懒加载
- 组件按需导入
- 图片CDN加速
- 开启Gzip压缩

### 后端优化
- 数据库索引优化
- Redis缓存
- Celery队列调优
- Docker资源限制

### 部署优化
- Nginx反向代理
- 静态资源CDN
- 数据库读写分离
- 容器健康检查

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📧 联系方式

如有问题，请提交Issue或联系：[你的邮箱]

## 📄 许可证

MIT License

---

**感谢使用 DevOps自动化部署平台！** 🚀

如有问题，请查看文档或提交Issue。
