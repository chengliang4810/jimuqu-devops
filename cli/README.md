# Jimuqu DevOps CLI

`@jimuqu/devops-cli` provides the `jimuqu-devops` command for managing Jimuqu DevOps from a terminal, CI job, or AI Agent.

It can inspect resources, create API tokens, manage hosts/projects/deploy configs, trigger deployments, stream logs, cancel runs, and apply declarative JSON/YAML deployment files.

## Install

```bash
npm install -g @jimuqu/devops-cli
jimuqu-devops --version
```

## Give This CLI To An Agent

Send the Agent this skill file URL:

```text
https://cdn.jsdelivr.net/npm/@jimuqu/devops-cli@latest/SKILL.md
```

Or use a fixed version for reproducible behavior. Replace `<version>` with the CLI version you want the Agent to use:

```text
https://cdn.jsdelivr.net/npm/@jimuqu/devops-cli@<version>/SKILL.md
```

Tell the Agent:

```text
Read this SKILL.md, install the CLI with npm, use JSON output, and deploy through Jimuqu DevOps:
https://cdn.jsdelivr.net/npm/@jimuqu/devops-cli@latest/SKILL.md
```

The skill file explains installation, authentication, API token usage, declarative `apply`, deployment triggers, log streaming, and safety rules for secrets.

## Configure

Agents and CI should use environment variables:

```bash
export JIMUQU_DEVOPS_URL=http://127.0.0.1:18080
export JIMUQU_DEVOPS_TOKEN=jdp_xxx
jimuqu-devops status --json
```

Windows PowerShell:

```powershell
$env:JIMUQU_DEVOPS_URL = "http://127.0.0.1:18080"
$env:JIMUQU_DEVOPS_TOKEN = "jdp_xxx"
jimuqu-devops status --json
```

Human operators can login once:

```bash
jimuqu-devops login --url http://127.0.0.1:18080 --username admin
jimuqu-devops status
```

The CLI stores url, username, and token only. It does not store passwords.

## API Tokens

Create a long-lived token for Agents or CI:

```bash
jimuqu-devops token create --name codex-agent --json
jimuqu-devops token list --json
jimuqu-devops token revoke <token-id> --json
```

Full token values are shown only once at creation time.

## Common Workflow

```bash
jimuqu-devops status --json
jimuqu-devops host list --json
jimuqu-devops project list --json
jimuqu-devops apply --file jimuqu-devops.yml --trigger --watch --json
```

`--watch` and `--follow` output JSONL events when `--json` is enabled.

## Declarative Apply

`apply` supports `.json`, `.yaml`, and `.yml` files. It upserts resources by natural key:

- host: `name`
- notification channel: `name`
- project: `repo_url + branch`

Sensitive values can reference local environment variables:

```yaml
host:
  password: ${SSH_PASSWORD}
project:
  git_password: ${GIT_TOKEN}
```

Run:

```bash
jimuqu-devops apply --file jimuqu-devops.yml --trigger --watch --json
```

## Discovery

Use these commands to inspect the exact installed command surface:

```bash
jimuqu-devops agent commands --json
jimuqu-devops agent examples --json
jimuqu-devops agent schema --json
jimuqu-devops --help
```

## More Documentation

- GitHub repository: https://github.com/chengliang4810/jimuqu-devops
- Agent skill via npm CDN: https://cdn.jsdelivr.net/npm/@jimuqu/devops-cli@latest/SKILL.md
- Agent skill via GitHub raw: https://raw.githubusercontent.com/chengliang4810/jimuqu-devops/main/agent-skills/jimuqu-devops/SKILL.md
