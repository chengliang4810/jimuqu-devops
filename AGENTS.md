# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## 项目概述

这是一个轻量级 DevOps / CI/CD 系统，当前仓库包含：

- Go 后端服务：负责认证、项目/主机/通知渠道管理、Webhook 触发、流水线执行、运行记录查询
- 内置管理后台：后端通过 `internal/httpapi/web` 提供一套嵌入式管理界面
- `web-next` 前端：基于 Next.js 的独立前端实现，作为新界面工程推进中

后端核心能力包括：

- Git webhook 触发项目构建与自动部署
- 手动触发部署与运行日志流式查看
- 通过 SSH / SFTP 上传制品并执行远程命令
- Docker 容器隔离编译环境
- SQLite 数据存储
- 敏感字段 AES-GCM 加密存储
- 管理员用户名密码登录，JWT 鉴权
- 通知渠道管理（Webhook、企业微信、钉钉、飞书、Email）

## 仓库结构

```text
cmd/server/main.go             - 启动入口，加载配置、修复中断任务、启动 HTTP 服务、优雅退出
internal/app/app.go            - 应用装配：初始化目录、Store、Executor、HTTP Handler、管理员账号
internal/auth/                 - JWT 生成与校验
internal/config/config.go      - 环境变量配置加载
internal/crypto/               - AES-GCM 与 HMAC 工具
internal/httpapi/              - HTTP API、认证中间件、SSE/日志流、内置管理界面
internal/httpapi/web/          - 嵌入式静态后台资源
internal/model/                - 领域模型：管理员、主机、项目、部署配置、运行记录、通知渠道
internal/notification/         - 通知发送相关逻辑
internal/pipeline/executor.go  - 流水线执行器：clone/build/filter/upload/deploy/notify
internal/store/store.go        - SQLite 存储、迁移、CRUD、加密字段落库
web-next/                      - 独立前端工程（Next.js + React + Tailwind CSS v4）
web-next/src/app/              - App Router 页面入口
web-next/src/components/       - 页面模块与 UI 组件
web-next/src/stores/           - Zustand 状态管理
web-next/src/lib,src/api/      - 前端 API 封装
data/                          - 运行期数据目录（数据库、工作区、制品等）
```

## 常用命令

### 后端

```bash
# 运行服务
go run ./cmd/server

# 构建
go build -o server.exe ./cmd/server

# 测试
go test ./...

# 健康检查
curl http://127.0.0.1:18080/healthz
```

### 前端 (`web-next`)

```bash
# 安装依赖
pnpm install

# 启动开发环境
pnpm dev

# 构建
pnpm build

# 启动生产服务
pnpm start
```

## 运行与配置

后端默认监听 `:18080`，首次启动会自动：

- 创建 `data`、`workspaces`、`artifacts` 等目录
- 初始化 SQLite 数据库和表结构
- 在管理员不存在时创建默认账号

默认管理员账号来自环境变量：

- 用户名：`admin`
- 密码：`admin123`

前端 `web-next` 可通过 `.env.local` 配置：

- `NEXT_PUBLIC_API_BASE_URL`：当前端与后端分离部署时指定 API 基础地址

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `APP_ADDR` | `:18080` | 服务监听地址 |
| `APP_DATA_DIR` | `./data` | 数据目录 |
| `APP_DB_DRIVER` | `sqlite` | 数据库驱动，支持 `sqlite` / `mysql` |
| `APP_DB_SOURCE` | `./data/pipeline.db` | 数据源；SQLite 填文件路径，MySQL 填 DSN |
| `APP_WORKSPACE_DIR` | `./data/workspaces` | Git 工作区目录 |
| `APP_ARTIFACT_DIR` | `./data/artifacts` | 制品目录 |
| `APP_SECRET` | `change-me-in-production` | AES 加密密钥 |
| `JWT_SECRET` | `change-me-in-production` | JWT 签名密钥 |
| `ADMIN_USERNAME` | `admin` | 初始管理员用户名 |
| `ADMIN_PASSWORD` | `admin123` | 初始管理员密码 |

## API 与业务要点

主要后端接口位于 `/api/v1`：

- `POST /api/v1/admin/login`：管理员登录
- `POST /api/v1/webhooks/{token}`：Webhook 触发部署
- `GET/POST/PUT/DELETE /api/v1/hosts`：主机管理
- `GET/POST/PUT/DELETE /api/v1/projects`：项目管理
- `PUT/GET /api/v1/projects/{id}/deploy-config`：部署配置
- `POST /api/v1/projects/{id}/trigger`：手动触发部署
- `GET /api/v1/runs/{runID}/stream`：运行日志流式输出
- `GET/POST/PUT/DELETE /api/v1/notification-channels`：通知渠道管理

Webhook 分支识别支持：

- `ref`
- `branch`
- Bitbucket 风格 `push.changes[0].new.name`
- Header `X-Git-Ref`

## 流水线执行流程

1. `git clone --depth 1 --single-branch` 拉取指定分支代码
2. 按项目配置处理 Git 认证（none / username / token / ssh）
3. 在 Docker 容器中执行编译命令
4. 按 `none/include/exclude` 规则过滤制品
5. 通过 SSH / SFTP 上传到远程保存目录
6. 执行部署前命令、同步到部署目录、执行部署后命令
7. 写入运行记录并发送通知

## 关键约束

- 项目通过 `repo_url + branch` 唯一确定
- SQLite 单连接（适合单进程部署）
- 执行器在服务进程内异步运行，没有分布式任务队列
- 远程部署依赖 Linux Shell 命令，如 `sh`、`cp`、`find`
- Docker 编译镜像必须包含 `sh`
- 敏感字段不能明文落库，需走现有加密逻辑
- 修改认证、通知、流水线逻辑时，要同时关注存储层与 API 校验逻辑

## 前端注意事项

`web-next` 是独立前端工程，开发时请优先修改：

- `web-next/src/**` 源码
- 不要编辑 `web-next/.next/**`、`web-next/out/**`、`web-next/node_modules/**` 生成产物

当前仓库前端现状：

- Next.js 15.1.6
- React 19
- Tailwind CSS v4
- `@tailwindcss/postcss`
- `motion`
- `sonner`
- `zustand`
- App Router 目录结构

## 开发建议

- 涉及后端改动时，优先查看 `internal/httpapi/server.go`、`internal/pipeline/executor.go`、`internal/store/store.go`
- 涉及鉴权改动时，同时检查 `internal/auth`、`internal/httpapi/middleware.go`、登录接口
- 涉及通知渠道改动时，同时检查模型、校验逻辑、发送逻辑、存储加密逻辑
- 涉及前端改动时，先确认是内置后台 `internal/httpapi/web` 还是独立前端 `web-next`
- 若同时修改前后端接口，保持字段命名和校验规则一致
