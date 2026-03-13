package model

import "time"

const (
	SettingDockerMirrorURL  = "docker_mirror_url"
	SettingGitDockerImage   = "git_docker_image"
	SettingPublicBaseURL    = "public_base_url"
	SettingProxyURL         = "proxy_url"
	SettingRunRetentionDays = "run_retention_days"
)

var DefaultSettings = map[string]string{
	SettingDockerMirrorURL:  "",
	SettingGitDockerImage:   "alpine/git:latest",
	SettingPublicBaseURL:    "",
	SettingProxyURL:         "",
	SettingRunRetentionDays: "30",
}

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SettingUpsert struct {
	Value string `json:"value"`
}

type ChangeUsernameInput struct {
	NewUsername string `json:"new_username"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type AccountProfile struct {
	Username string `json:"username"`
}

type BackupMeta struct {
	SchemaVersion int    `json:"schema_version"`
	ExportedAt    string `json:"exported_at"`
	RepoURL       string `json:"repo_url"`
	Version       string `json:"version"`
}

type BackupHost struct {
	ID        int64  `json:"id"`
	SortOrder int64  `json:"sort_order"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type BackupProject struct {
	ID           int64   `json:"id"`
	SortOrder    int64   `json:"sort_order"`
	Name         string  `json:"name"`
	RepoURL      string  `json:"repo_url"`
	Branch       string  `json:"branch"`
	Description  string  `json:"description"`
	WebhookToken string  `json:"webhook_token"`
	GitAuthType  string  `json:"git_auth_type"`
	GitUsername  *string `json:"git_username,omitempty"`
	GitPassword  *string `json:"git_password,omitempty"`
	GitSSHKey    *string `json:"git_ssh_key,omitempty"`
}

type BackupDeployConfig struct {
	ProjectID             int64    `json:"project_id"`
	HostID                int64    `json:"host_id"`
	BuildImage            string   `json:"build_image"`
	BuildCommands         []string `json:"build_commands"`
	ArtifactFilterMode    string   `json:"artifact_filter_mode"`
	ArtifactRules         []string `json:"artifact_rules"`
	RemoteSaveDir         string   `json:"remote_save_dir"`
	RemoteDeployDir       string   `json:"remote_deploy_dir"`
	PreDeployCommands     []string `json:"pre_deploy_commands"`
	PostDeployCommands    []string `json:"post_deploy_commands"`
	VersionCount          int      `json:"version_count"`
	TimeoutSeconds        int      `json:"timeout_seconds"`
	NotifyWebhookURL      string   `json:"notify_webhook_url"`
	NotifyBearerToken     *string  `json:"notify_bearer_token,omitempty"`
	NotificationChannelID *int64   `json:"notification_channel_id"`
}

type BackupProjectBundle struct {
	Project      BackupProject       `json:"project"`
	DeployConfig *BackupDeployConfig `json:"deploy_config,omitempty"`
}

type BackupData struct {
	Meta                 BackupMeta                      `json:"meta"`
	Hosts                []BackupHost                    `json:"hosts"`
	Projects             []BackupProjectBundle           `json:"projects"`
	NotificationChannels []NotificationChannelWithConfig `json:"notification_channels"`
	Settings             []Setting                       `json:"settings"`
}

type BackupRestoreResult struct {
	RowsAffected map[string]int `json:"rows_affected"`
}

type SystemInfo struct {
	RepoURL string `json:"repo_url"`
	Version string `json:"version"`
}
