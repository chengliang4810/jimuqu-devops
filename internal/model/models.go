package model

import (
	"path"
	"strings"
	"time"
)

const (
	ArtifactFilterNone    = "none"
	ArtifactFilterInclude = "include"
	ArtifactFilterExclude = "exclude"

	RunStatusQueued  = "queued"
	RunStatusRunning = "running"
	RunStatusSuccess = "success"
	RunStatusFailed  = "failed"

	TriggerTypeWebhook = "webhook"
	TriggerTypeManual  = "manual"

	GitAuthTypeNone     = "none"
	GitAuthTypeUsername = "username" // 用户名密码认证
	GitAuthTypeToken    = "token"    // Token认证
	GitAuthTypeSSH      = "ssh"      // SSH密钥认证
)

type Host struct {
	ID          int64     `json:"id"`
	SortOrder   int64     `json:"sort_order"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Port        int       `json:"port"`
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	HasPassword bool      `json:"has_password"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type HostUpsert struct {
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Port     int     `json:"port"`
	Username string  `json:"username"`
	Password *string `json:"password"`
}

type Project struct {
	ID              int64     `json:"id"`
	SortOrder       int64     `json:"sort_order"`
	Name            string    `json:"name"`
	RepoURL         string    `json:"repo_url"`
	Branch          string    `json:"branch"`
	Description     string    `json:"description"`
	WebhookToken    string    `json:"webhook_token"`
	HasDeployConfig bool      `json:"has_deploy_config"`
	GitAuthType     string    `json:"git_auth_type"` // none/username/token/ssh
	GitUsername     string    `json:"git_username"`  // Git用户名（加密）
	GitPassword     string    `json:"-"`             // Git密码/Token（加密）
	GitSSHKey       string    `json:"-"`             // SSH私钥（加密）
	HasGitAuth      bool      `json:"has_git_auth"`  // 是否配置了Git认证
	HasGitPassword  bool      `json:"has_git_password"`
	HasGitSSHKey    bool      `json:"has_git_ssh_key"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProjectUpsert struct {
	Name        string  `json:"name"`
	RepoURL     string  `json:"repo_url"`
	Branch      string  `json:"branch"`
	Description string  `json:"description"`
	GitAuthType string  `json:"git_auth_type"` // none/username/token/ssh
	GitUsername *string `json:"git_username"`  // Git用户名
	GitPassword *string `json:"git_password"`  // Git密码/Token
	GitSSHKey   *string `json:"git_ssh_key"`   // SSH私钥
}

type ProjectDetailUpsert struct {
	Name         string             `json:"name"`
	RepoURL      string             `json:"repo_url"`
	Branch       string             `json:"branch"`
	Description  string             `json:"description"`
	GitAuthType  string             `json:"git_auth_type"`
	GitUsername  *string            `json:"git_username"`
	GitPassword  *string            `json:"git_password"`
	GitSSHKey    *string            `json:"git_ssh_key"`
	DeployConfig DeployConfigUpsert `json:"deploy_config"`
}

func (p ProjectDetailUpsert) ProjectUpsert() ProjectUpsert {
	return ProjectUpsert{
		Name:        p.Name,
		RepoURL:     p.RepoURL,
		Branch:      p.Branch,
		Description: p.Description,
		GitAuthType: p.GitAuthType,
		GitUsername: p.GitUsername,
		GitPassword: p.GitPassword,
		GitSSHKey:   p.GitSSHKey,
	}
}

type ProjectCloneInput struct {
	Name        string `json:"name"`
	Branch      string `json:"branch"`
	Description string `json:"description"`
}

type DeployConfig struct {
	ID                    int64     `json:"id"`
	ProjectID             int64     `json:"project_id"`
	HostID                int64     `json:"host_id"`
	BuildImage            string    `json:"build_image"`
	BuildCommands         []string  `json:"build_commands"`
	CacheDirs             []string  `json:"cache_dirs"`
	ArtifactFilterMode    string    `json:"artifact_filter_mode"`
	ArtifactRules         []string  `json:"artifact_rules"`
	RemoteSaveDir         string    `json:"remote_save_dir"`
	RemoteDeployDir       string    `json:"remote_deploy_dir"`
	PreDeployCommands     []string  `json:"pre_deploy_commands"`
	PostDeployCommands    []string  `json:"post_deploy_commands"`
	VersionCount          int       `json:"version_count"`
	TimeoutSeconds        int       `json:"timeout_seconds"`
	NotifyWebhookURL      string    `json:"notify_webhook_url"`
	NotifyBearerToken     string    `json:"-"`
	HasNotifyToken        bool      `json:"has_notify_token"`
	NotificationChannelID *int64    `json:"notification_channel_id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type DeployConfigUpsert struct {
	HostID                int64    `json:"host_id"`
	BuildImage            string   `json:"build_image"`
	BuildCommands         []string `json:"build_commands"`
	CacheDirs             []string `json:"cache_dirs"`
	ArtifactFilterMode    string   `json:"artifact_filter_mode"`
	ArtifactRules         []string `json:"artifact_rules"`
	RemoteSaveDir         string   `json:"remote_save_dir"`
	RemoteDeployDir       string   `json:"remote_deploy_dir"`
	PreDeployCommands     []string `json:"pre_deploy_commands"`
	PostDeployCommands    []string `json:"post_deploy_commands"`
	VersionCount          int      `json:"version_count"`
	TimeoutSeconds        int      `json:"timeout_seconds"`
	NotifyWebhookURL      string   `json:"notify_webhook_url"`
	NotifyBearerToken     *string  `json:"notify_bearer_token"`
	NotificationChannelID *int64   `json:"notification_channel_id"`
}

type ProjectDetail struct {
	Project      Project       `json:"project"`
	DeployConfig *DeployConfig `json:"deploy_config,omitempty"`
	Host         *Host         `json:"host,omitempty"`
}

type ExecutionBundle struct {
	Project      Project
	DeployConfig DeployConfig
	Host         Host
}

type PipelineRun struct {
	ID            int64      `json:"id"`
	ProjectID     int64      `json:"project_id"`
	ProjectName   string     `json:"project_name"`
	Branch        string     `json:"branch"`
	Status        string     `json:"status"`
	TriggerType   string     `json:"trigger_type"`
	TriggerRef    string     `json:"trigger_ref"`
	CommitID      string     `json:"commit_id"`
	CommitMessage string     `json:"commit_message"`
	Author        string     `json:"author"`
	LogText       string     `json:"log_text"`
	ErrorMessage  string     `json:"error_message"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type RunCreateInput struct {
	ProjectID   int64
	Status      string
	TriggerType string
	TriggerRef  string
}

type ReorderInput struct {
	IDs []int64 `json:"ids"`
}

var defaultDeployCacheDirs = []string{
	"/root/.m2",
	"/root/.gradle/caches",
	"/root/.npm",
	"/root/.yarn",
	"/go/pkg/mod",
	"/root/.cache",
}

func DefaultDeployCacheDirs() []string {
	return append([]string(nil), defaultDeployCacheDirs...)
}

func NormalizeCacheDirs(cacheDirs []string) []string {
	if len(cacheDirs) == 0 {
		return DefaultDeployCacheDirs()
	}

	normalized := make([]string, 0, len(cacheDirs))
	seen := make(map[string]struct{}, len(cacheDirs))
	for _, dir := range cacheDirs {
		cleaned := NormalizeCacheDir(dir)
		if cleaned == "" {
			continue
		}
		if _, exists := seen[cleaned]; exists {
			continue
		}
		seen[cleaned] = struct{}{}
		normalized = append(normalized, cleaned)
	}

	if len(normalized) == 0 {
		return DefaultDeployCacheDirs()
	}

	return normalized
}

func NormalizeCacheDir(cacheDir string) string {
	cleaned := strings.TrimSpace(strings.ReplaceAll(cacheDir, "\\", "/"))
	if cleaned == "" {
		return ""
	}

	cleaned = path.Clean(cleaned)
	if cleaned == "." || cleaned == "/" || !strings.HasPrefix(cleaned, "/") {
		return ""
	}

	return cleaned
}
