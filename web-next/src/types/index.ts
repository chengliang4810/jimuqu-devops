// 主机类型
export interface Host {
  id: number;
  name: string;
  address: string;
  port: number;
  username: string;
  created_at: string;
  updated_at: string;
}

// 项目类型
export interface Project {
  id: number;
  name: string;
  branch: string;
  repo_url: string;
  description: string;
  timeout_minutes: number;
  webhook_token: string;
  deploy_config: DeployConfig;
  created_at: string;
  updated_at: string;
}

// 部署配置
export interface DeployConfig {
  host_id: number;
  remote_save_dir: string;
  remote_deploy_dir: string;
  pre_deploy_commands: string;
  post_deploy_commands: string;
  version_count: number;
  notification_channel_id: number | null;
  build_image: string;
  build_commands: string;
  artifact_filter_mode: "none" | "include" | "exclude";
  artifact_rules: string;
}

// 部署记录状态
export type RunStatus = "queued" | "running" | "success" | "failed";

// 部署记录
export interface PipelineRun {
  id: number;
  project_id: number;
  project_name: string;
  branch: string;
  status: RunStatus;
  commit_message: string;
  log_text: string;
  started_at: string;
  finished_at: string | null;
}

// 通知渠道类型
export type NotifyChannelType = "webhook" | "dingtalk" | "wechat" | "feishu";

// 通知渠道
export interface NotifyChannel {
  id: number;
  name: string;
  type: NotifyChannelType;
  webhook_url: string;
  secret: string;
  remark: string;
  created_at: string;
  updated_at: string;
}

// 统计数据
export interface Stats {
  project_count: number;
  host_count: number;
  run_count: number;
  notify_channel_count: number;
}

// 登录请求
export interface LoginRequest {
  username: string;
  password: string;
}

// 登录响应
export interface LoginResponse {
  token: string;
}
