# Jimuqu DevOps CLI Agent Skill

Use this file as the complete Agent guide for installing and using the Jimuqu DevOps CLI.

## Install

Install the CLI with npm:

```bash
npm install -g @jimuqu/devops-cli
jimuqu-devops --version
```

The npm package is `@jimuqu/devops-cli`; the command name is `jimuqu-devops`.

## Required Configuration

Prefer environment variables for Agent and CI usage:

```bash
export JIMUQU_DEVOPS_URL=http://127.0.0.1:18080
export JIMUQU_DEVOPS_TOKEN=jdp_xxx
```

On Windows PowerShell:

```powershell
$env:JIMUQU_DEVOPS_URL = "http://127.0.0.1:18080"
$env:JIMUQU_DEVOPS_TOKEN = "jdp_xxx"
```

Verify connectivity:

```bash
jimuqu-devops status --json
jimuqu-devops whoami --json
```

If a token is not available and a human is operating the terminal, login once:

```bash
jimuqu-devops login --url http://127.0.0.1:18080 --username admin
```

The CLI stores url, username, and token only. It does not store passwords.

## API Token

For Agents, use API Tokens instead of asking for an admin password. Create one once and store it outside the conversation:

```bash
jimuqu-devops token create --name codex-agent --json
jimuqu-devops token list --json
jimuqu-devops token revoke <token-id> --json
```

Full token values are only returned on creation. Never print full API tokens, passwords, Git tokens, or SSH keys in final replies.

## Agent Rules

- Use `jimuqu-devops`; do not read or write the database directly.
- Add `--json` to commands whenever possible.
- Treat `--watch` and `--follow` output as JSONL: one JSON object per line.
- Before creating deployment resources, inspect existing resources with `host list`, `project list`, and `notify list`.
- Prefer declarative `apply --file ...` for project setup and updates.
- Use environment variable references such as `${SSH_PASSWORD}` and `${GIT_TOKEN}` in apply files for secrets.
- Do not persist user passwords in files.

## Discovery Commands

Use these commands to learn the exact installed CLI surface:

```bash
jimuqu-devops agent commands --json
jimuqu-devops agent examples --json
jimuqu-devops agent schema --json
jimuqu-devops --help
```

## Recommended Deployment Workflow

1. Check server and current resources:

```bash
jimuqu-devops status --json
jimuqu-devops host list --json
jimuqu-devops project list --json
jimuqu-devops notify list --json
```

2. Create or update a declarative file named `jimuqu-devops.yml`.

3. Apply configuration, trigger deployment, and watch logs:

```bash
jimuqu-devops apply --file jimuqu-devops.yml --trigger --watch --json
```

4. If a deployment is already running or the run id is known:

```bash
jimuqu-devops run logs <run-id> --follow --json
jimuqu-devops run cancel <run-id> --json
jimuqu-devops deploy trigger <project-id> --watch --json
```

## Declarative Apply Example

```yaml
host:
  name: prod-1
  address: 127.0.0.1
  port: 22
  username: root
  password: ${SSH_PASSWORD}

notification_channel:
  name: prod-dingtalk
  type: dingtalk
  is_default: true
  remark: Production deploy alerts
  config:
    webhook_url: https://oapi.dingtalk.com/robot/send?access_token=xxx
    secret: ${DINGTALK_SECRET}

project:
  name: api-server
  repo_url: https://github.com/example/api-server.git
  branch: main
  description: API server
  git_auth_type: none

deploy_config:
  build_image: node:22-alpine
  build_commands:
    - npm install
    - npm run build
  artifact_filter_mode: include
  artifact_rules:
    - dist/**
  remote_save_dir: /opt/releases/api-server
  remote_deploy_dir: /opt/api-server
  pre_deploy_commands: []
  post_deploy_commands:
    - pm2 restart api-server
  version_count: 5
  timeout_seconds: 1800
```

`apply` upserts resources by natural key:

- host: `name`
- notification channel: `name`
- project: `repo_url + branch`

## Notification Channel Body

`notify create --file` and `notify update --file` accept the same body shape as the `notification_channel` block in `apply` files. Use this to define notification channels declaratively without inspecting server or frontend source code.

### DingTalk JSON example:

```json
{
  "name": "prod-dingtalk",
  "type": "dingtalk",
  "is_default": true,
  "remark": "Production deploy alerts",
  "config": {
    "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    "secret": "${DINGTALK_SECRET}"
  }
}
```

### Supported Notification Channel Types

| type | required config fields | optional config fields |
| --- | --- | --- |
| webhook | `url` | `token`, `secret` |
| wechat | `webhook_url` | `key` |
| dingtalk | `webhook_url` | `secret` |
| feishu | `webhook_url` | _none_ |
| email | `smtp_host`, `smtp_port`, `username`, `password`, `from`, `to` | `subject` |

### Command Examples

```bash
jimuqu-devops notify create --file dingtalk-channel.json --json
jimuqu-devops notify test <channel-id> --json
```

## Common Commands

```bash
jimuqu-devops host list --json
jimuqu-devops host get <host-id> --json
jimuqu-devops host create --name prod --address 1.2.3.4 --port 22 --username root --password "$SSH_PASSWORD" --json

jimuqu-devops project list --json
jimuqu-devops project get <project-id> --json
jimuqu-devops project create --file project.json --json

jimuqu-devops deploy-config get <project-id> --json
jimuqu-devops deploy-config set <project-id> --file deploy-config.yml --json

jimuqu-devops deploy trigger <project-id> --watch --json
jimuqu-devops run list --json
jimuqu-devops run get <run-id> --json
jimuqu-devops run logs <run-id> --follow --json
jimuqu-devops run cancel <run-id> --json

jimuqu-devops notify list --json
jimuqu-devops notify test <channel-id> --json

jimuqu-devops setting list --json
```

## Raw API Escape Hatch

For unsupported CLI operations, use the authenticated API helper:

```bash
jimuqu-devops api get /projects --json
jimuqu-devops api post /projects/1/trigger --json
jimuqu-devops api put /settings/run_retention_days --data '{"value":"30"}' --json
```
