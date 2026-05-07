# Agent 使用 Jimuqu DevOps CLI

Agent 的默认入口是 `jimuqu-devops apply`，不要直接操作数据库，也不要保存管理员密码。

## 推荐流程

```bash
jimuqu-devops status --json
jimuqu-devops host list --json
jimuqu-devops project list --json
jimuqu-devops apply --file jimuqu-devops.yml --trigger --watch --json
```

如果当前环境没有登录态，优先使用环境变量：

```bash
JIMUQU_DEVOPS_URL=http://127.0.0.1:18080
JIMUQU_DEVOPS_TOKEN=jdp_xxx
```

## 规则

- 所有读取和部署命令都加 `--json`。`--watch` / `--follow` 会输出 JSONL 事件流，每行都是一个 JSON 对象。
- 需要长期自动化时使用 API Token，不要求用户提供管理员密码。
- 创建或更新部署前先查询 `host list` 和 `project list`。
- 不在回复中打印密码、Git token、SSH key 或完整 API Token。
- `apply` 文件中的敏感值使用 `${ENV_NAME}` 引用环境变量。

## 能力发现

```bash
jimuqu-devops agent commands --json
jimuqu-devops agent examples --json
jimuqu-devops agent schema --json
```

## 常见操作

```bash
jimuqu-devops token create --name codex-agent --json
jimuqu-devops deploy trigger <project-id> --watch --json
jimuqu-devops run logs <run-id> --follow --json
jimuqu-devops api get /projects --json
```
