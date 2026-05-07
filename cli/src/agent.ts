export const agentCommands = {
  commands: [
    {
      name: "status",
      cli: "jimuqu-devops status --json",
      purpose: "检查服务连接和当前凭据是否可用",
    },
    {
      name: "apply",
      cli: "jimuqu-devops apply --file jimuqu-devops.yml --trigger --watch --json",
      purpose: "按声明式文件创建或更新资源并触发部署",
    },
    {
      name: "token.create",
      cli: "jimuqu-devops token create --name codex-agent --json",
      purpose: "创建长期 API Token，只显示一次明文",
    },
    {
      name: "run.logs",
      cli: "jimuqu-devops run logs <run-id> --follow --json",
      purpose: "查看部署日志",
    },
  ],
};

export const applySchema = {
  type: "object",
  properties: {
    host: { type: "object" },
    notification_channel: { type: "object" },
    project: { type: "object" },
    deploy_config: { type: "object" },
  },
};

export const applyExample = {
  host: {
    name: "prod-1",
    address: "127.0.0.1",
    port: 22,
    username: "root",
    password: "${SSH_PASSWORD}",
  },
  project: {
    name: "api-server",
    repo_url: "https://github.com/example/api-server.git",
    branch: "main",
    description: "API server",
    git_auth_type: "none",
  },
  deploy_config: {
    build_image: "node:22-alpine",
    build_commands: ["npm install", "npm run build"],
    artifact_filter_mode: "include",
    artifact_rules: ["dist/**"],
    remote_save_dir: "/opt/releases/api-server",
    remote_deploy_dir: "/opt/api-server",
    pre_deploy_commands: [],
    post_deploy_commands: ["pm2 restart api-server"],
    version_count: 5,
    timeout_seconds: 1800,
  },
};
