# 项目结构说明

## 📂 完整目录结构

```
devops-platform/
│
├── app/                          # FastAPI应用主目录
│   ├── api/                      # API路由模块
│   │   ├── __init__.py
│   │   ├── deployments.py        # 部署记录API
│   │   ├── projects.py           # 项目管理API
│   │   └── webhook.py            # Webhook接收器
│   │
│   ├── models/                   # 数据模型
│   │   ├── __init__.py
│   │   ├── project.py            # 项目模型
│   │   └── deployment.py         # 部署记录模型
│   │
│   ├── services/                 # 业务逻辑层
│   │   ├── deployment_service.py # 部署服务
│   │   ├── docker_service.py     # Docker编译服务
│   │   └── ssh_service.py        # SSH部署服务
│   │
│   ├── __init__.py
│   ├── auth.py                   # 认证中间件
│   ├── config.py                 # 配置管理
│   ├── database.py               # 数据库工具
│   ├── main.py                   # FastAPI入口
│   ├── schemas.py                # Pydantic模型
│   └── websocket.py              # WebSocket处理
│
├── worker/                       # Celery Worker
│   └── tasks.py                  # 异步任务
│
├── static/                       # 前端静态文件
│   ├── css/
│   │   └── style.css             # 页面样式
│   ├── js/
│   │   ├── api.js                # API调用封装
│   │   └── app.js                # 前端主逻辑
│   └── index.html                # 主页面
│
├── .env.example                  # 环境变量示例
├── .gitignore                    # Git忽略文件
├── Dockerfile                    # Docker镜像构建文件
├── docker-compose.yml            # Docker编排配置
├── QUICKSTART.md                 # 快速开始指南
├── README.md                     # 项目说明文档
├── requirements.txt              # Python依赖
└── start.sh                      # 启动脚本
```

## 📄 核心文件说明

### 后端服务 (app/)

#### 配置与入口
- **config.py**: 统一配置管理，读取环境变量
- **main.py**: FastAPI主应用，整合所有模块
- **database.py**: 数据库连接和会话管理
- **auth.py**: HTTP Basic认证中间件

#### 数据模型
- **models/project.py**: 项目实体定义
- **models/deployment.py**: 部署记录实体定义
- **schemas.py**: API请求/响应模型（Pydantic）

#### API路由
- **api/projects.py**: 项目CRUD接口
- **api/deployments.py**: 部署管理接口
- **api/webhook.py**: GitHub/GitLab Webhook接收

#### 业务逻辑
- **services/docker_service.py**: Docker容器编译逻辑
- **services/ssh_service.py**: SSH文件传输和命令执行
- **services/deployment_service.py**: 部署流程编排
- **websocket.py**: WebSocket实时日志推送

### 异步任务 (worker/)

- **tasks.py**: Celery异步任务，Web和Worker进程共享

### 前端界面 (static/)

- **index.html**: 单页应用主页面
- **css/style.css**: 响应式样式
- **js/api.js**: HTTP API调用封装
- **js/app.js**: 前端交互逻辑

### 部署配置

- **Dockerfile**: 单容器镜像定义
- **docker-compose.yml**: 多容器编排（Web + Worker + Redis）
- **start.sh**: 一键启动脚本
- **requirements.txt**: Python依赖列表

## 🔄 数据流

### 1. 手动部署流程

```
用户点击部署
    ↓
FastAPI创建Deployment记录
    ↓
Celery异步任务提交到队列
    ↓
Worker执行部署任务
    ├─→ Docker编译
    ├─→ SSH上传
    ├─→ 执行重启命令
    └─→ 推送实时日志到WebSocket
    ↓
前端实时显示日志
```

### 2. Webhook自动部署流程

```
Git仓库推送代码
    ↓
GitHub/GitLab发送Webhook
    ↓
Webhook接收器验证签名
    ↓
FastAPI创建Deployment记录
    ↓
提交Celery异步任务
    ↓
(同手动部署流程)
```

## 🏗️ 架构特点

### 1. 分层架构
- **API层**: RESTful接口，认证和参数验证
- **Service层**: 业务逻辑封装
- **模型层**: 数据持久化

### 2. 异步处理
- **Web服务**: 只负责接收请求，响应快速
- **Worker进程**: 异步执行长时间任务
- **Redis**: 任务队列和缓存

### 3. 容器化
- **Docker编译**: 隔离不同语言编译环境
- **资源限制**: 容器内存和CPU限制
- **一键部署**: Docker Compose编排

### 4. 实时通信
- **WebSocket**: 实时日志推送
- **长连接**: 部署过程可视化
- **状态同步**: 前端实时更新

## 📊 技术栈

| 组件 | 技术选型 | 用途 |
|------|----------|------|
| Web框架 | FastAPI 0.104 | 异步API框架 |
| 任务队列 | Celery 5.3 + Redis | 异步任务处理 |
| 数据库 | SQLAlchemy + SQLite | ORM和持久化 |
| 容器 | Docker SDK | 动态编译环境 |
| 传输 | Paramiko (SSH) | 远程文件传输 |
| 前端 | 原生HTML/CSS/JS | 轻量级SPA |
| 实时通信 | WebSocket | 实时日志推送 |
| 构建 | Docker Compose | 多容器编排 |

## 🚀 快速启动

```bash
# 1. 启动所有服务
./start.sh

# 2. 访问平台
# Web: http://localhost:8000
# API文档: http://localhost:8000/docs

# 3. 登录
# 用户名: admin
# 密码: admin123
```

## 📝 扩展建议

### 前端优化
- 替换为Vue3/React前端框架
- 添加ECharts数据可视化
- 移动端适配优化

### 功能增强
- 邮件/钉钉通知
- 多环境部署（dev/staging/prod）
- 自动回滚机制
- 部署审批流程

### 安全加固
- JWT Token认证
- HTTPS强制跳转
- 审计日志
- 权限控制

### 性能优化
- 数据库迁移到PostgreSQL
- Redis集群
- CDN加速静态资源
- 缓存优化

---

更多详情请查看 [README.md](README.md) 和 [QUICKSTART.md](QUICKSTART.md)
