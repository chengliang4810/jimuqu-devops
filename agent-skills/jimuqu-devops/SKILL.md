# Jimuqu DevOps

Use this skill when configuring or deploying projects through Jimuqu DevOps.

## Rules

- Use the `jimuqu-devops` CLI; do not call the database directly.
- Prefer `--json` for all commands that return data.
- Do not ask users for admin passwords if `JIMUQU_DEVOPS_TOKEN` or an existing CLI login is available.
- Do not print full API tokens, passwords, Git tokens, or SSH keys in final replies.
- Use environment variable references such as `${SSH_PASSWORD}` in apply files for secrets.

## Recommended Workflow

```bash
jimuqu-devops status --json
jimuqu-devops host list --json
jimuqu-devops project list --json
jimuqu-devops apply --file jimuqu-devops.yml --trigger --watch --json
```

## API Token

For long-running automation, create an API Token once:

```bash
jimuqu-devops token create --name codex-agent --json
```

Store it outside the conversation as `JIMUQU_DEVOPS_TOKEN`.

## Discovery

```bash
jimuqu-devops agent commands --json
jimuqu-devops agent examples --json
jimuqu-devops agent schema --json
```
