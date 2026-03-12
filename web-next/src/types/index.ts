// 主机类型
export interface Host {
  id: number;
  sort_order: number;
  name: string;
  address: string;
  port: number;
  username: string;
  has_password?: boolean;
  created_at: string;
  updated_at: string;
}

// 项目类型
export interface Project {
  id: number;
  sort_order: number;
  name: string;
  branch: string;
  repo_url: string;
  description: string;
  webhook_token: string;
  has_deploy_config: boolean;
  git_auth_type: "none" | "username" | "token" | "ssh";
  has_git_auth: boolean;
  created_at: string;
  updated_at: string;
}

// 部署配置
export interface DeployConfig {
  id?: number;
  project_id?: number;
  host_id: number;
  remote_save_dir: string;
  remote_deploy_dir: string;
  pre_deploy_commands: string[];
  post_deploy_commands: string[];
  version_count?: number;
  notification_channel_id: number | null;
  build_image: string;
  build_commands: string[];
  artifact_filter_mode: "none" | "include" | "exclude";
  artifact_rules: string[];
  timeout_seconds?: number;
  notify_webhook_url?: string;
  has_notify_token?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface ProjectDetail {
  project: Project;
  deploy_config?: DeployConfig | null;
  host?: Host | null;
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
  sort_order: number;
  name: string;
  type: NotifyChannelType;
  is_default?: boolean;
  remark: string;
  created_at: string;
  updated_at: string;
}

export interface NotifyChannelDetail extends NotifyChannel {
  config: Record<string, unknown>;
}

// 统计数据
export interface Stats {
  project_count: number;
  host_count: number;
  run_count: number;
  notify_channel_count: number;
}

export interface HomeStatsTotal {
  deploy_count: number;
  success_count: number;
  failed_count: number;
  running_count: number;
  queued_count: number;
  project_count: number;
  success_rate: number;
  average_deploy_per_project: number;
  average_deploy_duration_seconds: number;
  last_deploy_at: string;
}

export interface HomeStatsDaily {
  date: string;
  deploy_count: number;
  success_count: number;
  failed_count: number;
  running_count: number;
  queued_count: number;
  success_rate: number;
}

export interface HomeStatsHourly {
  hour: string;
  deploy_count: number;
  success_count: number;
  failed_count: number;
  running_count: number;
  queued_count: number;
  success_rate: number;
}

export interface HomeProjectRank {
  project_id: number;
  project_name: string;
  branch: string;
  deploy_count: number;
  success_count: number;
  failed_count: number;
  running_count: number;
  queued_count: number;
  success_rate: number;
  last_deploy_at: string;
}

export interface HomeDashboard {
  total: HomeStatsTotal;
  daily: HomeStatsDaily[];
  hourly: HomeStatsHourly[];
  projects: HomeProjectRank[];
}

export type SettingKey = "docker_mirror_url" | "proxy_url" | "run_retention_days";

export interface Setting {
  key: SettingKey;
  value: string;
  created_at?: string;
  updated_at?: string;
}

export interface AccountProfile {
  username: string;
}

export interface BackupRestoreResult {
  rows_affected: Record<string, number>;
}

export interface SystemInfo {
  repo_url: string;
  version: string;
}

export interface ReleaseAsset {
  name: string;
  browser_download_url: string;
}

export interface ReleaseInfo {
  tag_name: string;
  published_at: string;
  body: string;
  html_url: string;
  assets: ReleaseAsset[];
}

export interface UpdateStatus {
  current_version: string;
  latest_version: string;
  has_update: boolean;
}

export interface UpdateResult {
  message: string;
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
