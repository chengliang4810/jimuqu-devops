# Go Pipeline

一个使用 Go 实现的轻量流水线后端，目标是解决以下场景：

- 通过 Git 仓库 webhook 触发项目构建与自动部署
- 一个项目唯一对应一个 Git 仓库分支
- 通过 SSH 账号密码管理目标主机
- 使用 Docker 容器隔离编译环境
- 支持部署配置、项目复制、运行记录和成功/失败通知

当前实现包含后端 API 服务和内置管理后台页面。

## 核心模型

### 1. 项目管理

一个项目绑定：

- 项目名称
- Git 仓库地址
- 分支名
- 描述
- webhook token

约束：

- 同一个仓库通过 `repo_url + branch` 唯一确定一个项目
- 支持一键复制项目
- 复制项目时复用源项目仓库地址，并通过新分支生成新项目

### 2. 主机管理

一个主机包含：

- 名称
- 地址
- 端口
- SSH 用户名
- SSH 密码

说明：

- SSH 密码在 SQLite 中以本地密钥做 AES 加密存储

### 3. 部署配置

一个项目只有一份部署配置，包含：

- 目标主机
- 编译镜像名
- 编译命令列表
- 制品过滤模式：`none | include | exclude`
- 制品过滤规则
- 远程保存目录
- 远程部署目录
- 部署前命令
- 部署后命令
- 通知 webhook 地址
- 通知 bearer token

## 流水线流程

固定流程：

1. `git clone` 指定仓库分支
2. 在 Docker 容器中执行编译命令
3. 对编译结果执行包含/排除过滤
4. 通过 SSH/SFTP 上传到远程保存目录
5. 执行部署前命令
6. 将保存目录内容同步到部署目录
7. 执行部署后命令
8. 发送通知

通知规则：

- 全部成功，发送成功通知
- 任意环节失败，发送失败通知

## 技术选型

- HTTP 路由：`chi`
- 存储：`SQLite`
- SSH/SFTP：`golang.org/x/crypto/ssh` + `github.com/pkg/sftp`
- 编译容器：本机 `docker` CLI
- Git 克隆：本机 `git` CLI

## 目录结构

```text
cmd/server              启动入口
internal/app            应用装配
internal/config         配置加载
internal/crypto         敏感信息加密
internal/httpapi        HTTP API
internal/model          领域模型
internal/pipeline       流水线执行器
internal/store          SQLite 存储
data/                   运行数据目录
```

## 启动要求

运行服务的机器需要具备：

- Go 1.25+
- Git
- Docker
- 可访问目标主机的 SSH 网络
- 目标主机默认按 Linux Shell 行为执行部署命令

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `APP_ADDR` | `:18080` | HTTP 监听地址 |
| `APP_DATA_DIR` | `./data` | 数据根目录 |
| `APP_DB_PATH` | `./data/pipeline.db` | SQLite 数据库路径 |
| `APP_WORKSPACE_DIR` | `./data/workspaces` | 本地克隆目录 |
| `APP_ARTIFACT_DIR` | `./data/artifacts` | 本地制品目录 |
| `APP_SECRET` | `change-me-in-production` | 本地加密密钥 |

## 运行

```bash
go run ./cmd/server
```

管理后台入口：

```text
http://127.0.0.1:18080/
```

健康检查：

```bash
curl http://127.0.0.1:18080/healthz
```

## API 概览

管理后台页面会直接调用以下 API。

### 主机

- `POST /api/v1/hosts`
- `GET /api/v1/hosts`
- `GET /api/v1/hosts/{hostID}`
- `PUT /api/v1/hosts/{hostID}`
- `DELETE /api/v1/hosts/{hostID}`

### 项目

- `POST /api/v1/projects`
- `GET /api/v1/projects`
- `GET /api/v1/projects/{projectID}`
- `PUT /api/v1/projects/{projectID}`
- `DELETE /api/v1/projects/{projectID}`
- `POST /api/v1/projects/{projectID}/clone`

### 部署配置

- `PUT /api/v1/projects/{projectID}/deploy-config`
- `GET /api/v1/projects/{projectID}/deploy-config`

### 执行记录

- `POST /api/v1/projects/{projectID}/trigger`
- `GET /api/v1/projects/{projectID}/runs`
- `GET /api/v1/runs/{runID}`
- `GET /api/v1/runs/{runID}/stream`

### Webhook

- `POST /api/v1/webhooks/{token}`

## 示例

### 1. 创建主机

```bash
curl -X POST http://127.0.0.1:18080/api/v1/hosts \
  -H "Content-Type: application/json" \
  -d '{
    "name":"prod-1",
    "address":"192.168.1.10",
    "port":22,
    "username":"root",
    "password":"123456"
  }'
```

### 2. 创建项目

```bash
curl -X POST http://127.0.0.1:18080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name":"portal-prod",
    "repo_url":"https://github.com/example/portal.git",
    "branch":"main",
    "description":"生产环境"
  }'
```

### 3. 配置部署

```bash
curl -X PUT http://127.0.0.1:18080/api/v1/projects/1/deploy-config \
  -H "Content-Type: application/json" \
  -d '{
    "host_id":1,
    "build_image":"node:20",
    "build_commands":[
      "pnpm install",
      "pnpm run build"
    ],
    "artifact_filter_mode":"include",
    "artifact_rules":[
      "dist",
      "package.json"
    ],
    "remote_save_dir":"/data/releases",
    "remote_deploy_dir":"/data/apps/portal",
    "pre_deploy_commands":[
      "mkdir -p /data/apps/portal"
    ],
    "post_deploy_commands":[
      "docker restart app"
    ],
    "notify_webhook_url":"https://example.com/hooks/deploy"
  }'
```

### 4. 手动触发

```bash
curl -X POST http://127.0.0.1:18080/api/v1/projects/1/trigger
```

### 5. 复制项目

```bash
curl -X POST http://127.0.0.1:18080/api/v1/projects/1/clone \
  -H "Content-Type: application/json" \
  -d '{
    "name":"portal-test",
    "branch":"test"
  }'
```

### 6. 配置 Git Webhook

创建项目后，接口返回 `webhook_token`。将 Git 仓库 webhook 指向：

```text
POST http://your-server:18080/api/v1/webhooks/{webhook_token}
```

当前支持从以下字段识别分支：

- `ref`，例如 `refs/heads/main`
- `branch`
- Bitbucket 风格 `push.changes[0].new.name`
- Header `X-Git-Ref`

## 运行记录

每次触发会生成一条流水线记录，包含：

- 触发方式
- 状态
- 日志文本
- 错误信息
- 开始时间
- 结束时间

管理后台在查看某条运行记录时，会通过 `SSE` 接口实时流式刷新日志与状态。

## 当前实现约束

- 远程部署默认面向 Linux 主机，依赖 `sh`、`cp`、`find`
- Docker 编译阶段默认要求镜像内存在 `sh`
- 通知当前实现为 webhook 回调
- 已提供内置管理后台页面
- 没有做分布式队列，执行器在当前服务进程内异步运行

## 后续可扩展项

- 增加 webhook 签名校验
- 增加通知渠道，如企业微信、钉钉、飞书、邮件
- 增加多步骤编译 UI 编排
- 增加部署策略，如蓝绿、灰度、回滚
- 增加并发控制、任务队列和审计日志
