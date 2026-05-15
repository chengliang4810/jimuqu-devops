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

const notificationConfigSchemas = {
  webhook: {
    type: "object",
    required: ["url"],
    properties: {
      url: {
        type: "string",
        description: "Webhook endpoint URL",
      },
      token: {
        type: "string",
        description: "Optional bearer token or query token used by the webhook receiver",
      },
      secret: {
        type: "string",
        description: "Optional signing secret used to generate webhook signatures",
      },
    },
    additionalProperties: false,
  },
  wechat: {
    type: "object",
    required: ["webhook_url"],
    properties: {
      webhook_url: {
        type: "string",
        description: "WeChat robot webhook URL",
      },
      key: {
        type: "string",
        description: "Optional robot key used by certain WeChat webhook formats",
      },
    },
    additionalProperties: false,
  },
  dingtalk: {
    type: "object",
    required: ["webhook_url"],
    properties: {
      webhook_url: {
        type: "string",
        description: "DingTalk robot webhook URL",
      },
      secret: {
        type: "string",
        description: "DingTalk robot secret",
      },
    },
    additionalProperties: false,
  },
  feishu: {
    type: "object",
    required: ["webhook_url"],
    properties: {
      webhook_url: {
        type: "string",
        description: "Feishu robot webhook URL",
      },
    },
    additionalProperties: false,
  },
  email: {
    type: "object",
    required: ["smtp_host", "smtp_port", "username", "password", "from", "to"],
    properties: {
      smtp_host: {
        type: "string",
        description: "SMTP server host",
      },
      smtp_port: {
        type: "integer",
        description: "SMTP server port",
      },
      username: {
        type: "string",
        description: "SMTP login username",
      },
      password: {
        type: "string",
        description: "SMTP login password",
      },
      from: {
        type: "string",
        description: "Sender email address",
      },
      to: {
        type: "string",
        description: "Comma-separated recipient email addresses",
      },
      subject: {
        type: "string",
        description: "Optional default email subject",
      },
    },
    additionalProperties: false,
  },
} as const;

export const applySchema = {
  type: "object",
  properties: {
    host: { type: "object" },
    notification_channel: {
      type: "object",
      required: ["name", "type", "config"],
      properties: {
        name: {
          type: "string",
          description: "apply upsert key",
        },
        type: {
          type: "string",
          enum: ["webhook", "wechat", "dingtalk", "feishu", "email"],
          description: "Notification channel type",
        },
        is_default: {
          type: "boolean",
          description: "Whether this channel is the default notification target",
        },
        remark: {
          type: "string",
          description: "Human-readable note for the channel",
        },
        config: {
          type: "object",
          description: "Channel-specific configuration. See the notification_channel oneOf branches for fields by type.",
        },
      },
      oneOf: [
        {
          properties: {
            type: { const: "webhook" },
            config: notificationConfigSchemas.webhook,
          },
          required: ["type", "config"],
        },
        {
          properties: {
            type: { const: "wechat" },
            config: notificationConfigSchemas.wechat,
          },
          required: ["type", "config"],
        },
        {
          properties: {
            type: { const: "dingtalk" },
            config: notificationConfigSchemas.dingtalk,
          },
          required: ["type", "config"],
        },
        {
          properties: {
            type: { const: "feishu" },
            config: notificationConfigSchemas.feishu,
          },
          required: ["type", "config"],
        },
        {
          properties: {
            type: { const: "email" },
            config: notificationConfigSchemas.email,
          },
          required: ["type", "config"],
        },
      ],
      examples: [
        {
          name: "prod-dingtalk",
          type: "dingtalk",
          is_default: true,
          remark: "Production deploy alerts",
          config: {
            webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=xxx",
            secret: "${DINGTALK_SECRET}",
          },
        },
      ],
    },
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
  notification_channel: {
    name: "prod-dingtalk",
    type: "dingtalk",
    is_default: true,
    remark: "Production deploy alerts",
    config: {
      webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=xxx",
      secret: "${DINGTALK_SECRET}",
    },
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
