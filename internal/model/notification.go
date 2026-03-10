package model

import "time"

const (
	ChannelTypeWebhook  = "webhook"
	ChannelTypeWeChat   = "wechat"
	ChannelTypeDingTalk = "dingtalk"
	ChannelTypeFeishu   = "feishu"
	ChannelTypeEmail    = "email"
)

type NotificationChannel struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	IsDefault bool      `json:"is_default"`
	Remark    string    `json:"remark"`
	Config    string    `json:"-"` // JSON格式的配置，不直接返回给前端
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotificationChannelWithConfig struct {
	NotificationChannel
	ConfigMap map[string]any `json:"config"` // 解析后的配置返回给前端
}

type NotificationChannelUpsert struct {
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	IsDefault bool           `json:"is_default"`
	Remark    string         `json:"remark"`
	Config    map[string]any `json:"config"` // 前端传入的配置对象
}

// 通知配置结构
type WebhookConfig struct {
	URL    string `json:"url"`
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

type WeChatConfig struct {
	WebhookURL string `json:"webhook_url"`
	Key        string `json:"key"`
}

type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

type FeishuConfig struct {
	WebhookURL string `json:"webhook_url"`
}

type EmailConfig struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
}

// 通知载荷
type NotificationPayload struct {
	RunID        int64  `json:"run_id"`
	Status       string `json:"status"`
	ProjectID    int64  `json:"project_id"`
	ProjectName  string `json:"project_name"`
	RepoURL      string `json:"repo_url"`
	Branch       string `json:"branch"`
	TriggerType  string `json:"trigger_type"`
	TriggerRef   string `json:"trigger_ref"`
	CommitID     string `json:"commit_id"`
	CommitMessage string `json:"commit_message"`
	Author       string `json:"author"`
	ErrorMessage string `json:"error_message"`
	SentAt       string `json:"sent_at"`
}

type TestNotificationInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}