# Jimuqu DevOps CLI

CLI 包名为 `@jimuqu/devops-cli`，安装后提供 `jimuqu-devops` 命令。

```bash
npm install -g @jimuqu/devops-cli
```

## 登录与凭据

人使用时推荐登录一次：

```bash
jimuqu-devops login --url http://127.0.0.1:18080 --username admin
jimuqu-devops status
```

CLI 只保存服务地址、用户名和登录后的 token，不保存密码。Agent 或 CI 推荐使用 API Token：

```bash
export JIMUQU_DEVOPS_URL=http://127.0.0.1:18080
export JIMUQU_DEVOPS_TOKEN=jdp_xxx
jimuqu-devops status --json
```

创建 API Token：

```bash
jimuqu-devops token create --name codex-agent --json
jimuqu-devops token list
jimuqu-devops token revoke <token-id>
```

## 常用命令

```bash
jimuqu-devops host list
jimuqu-devops host create --name prod --address 1.2.3.4 --port 22 --username root --password "$SSH_PASSWORD"

jimuqu-devops project list
jimuqu-devops project get <project-id>
jimuqu-devops project create --file project.json

jimuqu-devops deploy-config get <project-id>
jimuqu-devops deploy-config set <project-id> --file deploy-config.yml

jimuqu-devops deploy trigger <project-id> --watch
jimuqu-devops run logs <run-id> --follow
jimuqu-devops run cancel <run-id>

jimuqu-devops notify list
jimuqu-devops notify test <channel-id>
jimuqu-devops setting list
```

所有命令都支持 `--json`，用于 Agent 或脚本稳定解析。`--watch` / `--follow` 会输出 JSONL 事件流，每行都是一个 JSON 对象。

## 声明式 apply

Agent 推荐生成 `jimuqu-devops.yml` 后执行：

```bash
jimuqu-devops apply --file jimuqu-devops.yml --trigger --watch --json
```

`apply` 会按顺序处理主机、通知渠道、项目、部署配置。已有资源按自然键更新：主机按 `name`，通知渠道按 `name`，项目按 `repo_url + branch`。

敏感字段可引用环境变量：

```yaml
host:
  password: ${SSH_PASSWORD}
project:
  git_password: ${GIT_TOKEN}
```

## 原始 API

不常用操作可以通过低层 API 命令调用：

```bash
jimuqu-devops api get /projects --json
jimuqu-devops api post /projects/1/trigger --json
jimuqu-devops api put /settings/run_retention_days --data '{"value":"30"}'
```
