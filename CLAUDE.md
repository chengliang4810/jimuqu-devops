# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个使用 Go 实现的轻量级 CI/CD 流水线后端服务，支持：
- Git webhook 触发项目构建与自动部署
- 通过 SSH 部署到远程主机
- Docker 容器隔离编译环境
- SQLite 数据存储，敏感信息 AES 加密

## 常用命令

```bash
# 运行服务
go run ./cmd/server

# 构建
go build -o server.exe ./cmd/server

# 健康检查
curl http://127.0.0.1:18080/healthz
```

## 架构

```
cmd/server/main.go      - 启动入口，信号处理
internal/app/app.go     - 应用装配：初始化 Store、Executor、HTTP Handler
internal/config/        - 环境变量配置加载
internal/model/         - 领域模型定义（Host、Project、DeployConfig、PipelineRun）
internal/store/         - SQLite 存储层，含数据库迁移和 CRUD 操作
internal/crypto/        - AES-GCM 加密工具，用于敏感字段加密
internal/pipeline/      - 流水线执行器：git clone → docker build → artifact filter → SSH deploy → notify
internal/httpapi/       - HTTP API 路由和处理器，内置管理后台 UI
```

## 流水线执行流程

1. `git clone --depth 1 --single-branch` 克隆指定分支
2. Docker 容器中执行编译命令
3. 制品过滤（none/include/exclude 模式）
4. SSH/SFTP 上传到远程保存目录
5. 执行部署前命令 → 同步到部署目录 → 执行部署后命令
6. Webhook 通知（成功/失败）

## 关键约束

- 项目通过 `repo_url + branch` 唯一确定
- SQLite 单连接（`SetMaxOpenConns(1)`），适合单进程部署
- 敏感字段（SSH密码、通知Token）使用 AES-GCM 加密存储
- 远程部署依赖 Linux Shell（`sh`、`cp`、`find`）
- Docker 编译镜像需包含 `sh`
- 执行器在服务进程内异步运行（无分布式队列）

## 环境变量

| 变量 | 默认值 |
|------|--------|
| `APP_ADDR` | `:18080` |
| `APP_DATA_DIR` | `./data` |
| `APP_DB_PATH` | `./data/pipeline.db` |
| `APP_WORKSPACE_DIR` | `./data/workspaces` |
| `APP_ARTIFACT_DIR` | `./data/artifacts` |
| `APP_SECRET` | `change-me-in-production` |

## Webhook 触发

- URL: `POST /api/v1/webhooks/{token}`
- 支持从 `ref`、`branch`、Bitbucket 风格 `push.changes[0].new.name`、Header `X-Git-Ref` 识别分支