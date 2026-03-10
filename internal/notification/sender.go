package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"devops-pipeline/internal/model"
)

type Sender struct {
	logger *slog.Logger
	client *http.Client
}

func New(logger *slog.Logger) *Sender {
	return &Sender{
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			},
		},
	}
}

func (s *Sender) Send(channel model.NotificationChannel, payload model.NotificationPayload) error {
	switch channel.Type {
	case model.ChannelTypeWebhook:
		return s.sendWebhook(channel, payload)
	case model.ChannelTypeWeChat:
		return s.sendWeChat(channel, payload)
	case model.ChannelTypeDingTalk:
		return s.sendDingTalk(channel, payload)
	case model.ChannelTypeFeishu:
		return s.sendFeishu(channel, payload)
	case model.ChannelTypeEmail:
		return s.sendEmail(channel, payload)
	default:
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}

func (s *Sender) sendWebhook(channel model.NotificationChannel, payload model.NotificationPayload) error {
	var config model.WebhookConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("unmarshal webhook config: %w", err)
	}

	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", config.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 添加Token认证
	if config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+config.Token)
	}

	// 添加签名
	if config.Secret != "" {
		signature := generateHMACSignature(config.Secret, body)
		req.Header.Set("X-Signature", "sha256="+signature)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("webhook notification sent", "channel_id", channel.ID, "run_id", payload.RunID)
	return nil
}

func (s *Sender) sendWeChat(channel model.NotificationChannel, payload model.NotificationPayload) error {
	var config model.WeChatConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("unmarshal wechat config: %w", err)
	}

	if config.WebhookURL == "" {
		return fmt.Errorf("wechat webhook URL is required")
	}

	// 构造企业微信消息格式
	content := fmt.Sprintf(
		"## 部署通知\n**项目**: %s\n**分支**: %s\n**状态**: %s\n**触发方式**: %s\n**错误信息**: %s\n**时间**: %s",
		payload.ProjectName,
		payload.Branch,
		payload.Status,
		payload.TriggerType,
		payload.ErrorMessage,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	message := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]any{
			"content": content,
		},
	}

	if config.Key != "" {
		message["sign"] = generateWeChatSign(config.Key, time.Now().Unix())
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("wechat returned status %d: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("wechat notification sent", "channel_id", channel.ID, "run_id", payload.RunID)
	return nil
}

func (s *Sender) sendDingTalk(channel model.NotificationChannel, payload model.NotificationPayload) error {
	var config model.DingTalkConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("unmarshal dingtalk config: %w", err)
	}

	if config.WebhookURL == "" {
		return fmt.Errorf("dingtalk webhook URL is required")
	}

	// 构造钉钉消息格式
	text := fmt.Sprintf(
		"【部署通知】\n项目: %s\n分支: %s\n状态: %s\n触发方式: %s\n错误信息: %s\n时间: %s",
		payload.ProjectName,
		payload.Branch,
		payload.Status,
		payload.TriggerType,
		payload.ErrorMessage,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	message := map[string]any{
		"msgtype": "text",
		"text": map[string]any{
			"content": text,
		},
	}

	// 添加签名
	if config.Secret != "" {
		timestamp := time.Now().UnixMilli()
		signature := generateDingTalkSign(config.Secret, timestamp)
		message["timestamp"] = timestamp
		message["sign"] = signature
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("dingtalk returned status %d: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("dingtalk notification sent", "channel_id", channel.ID, "run_id", payload.RunID)
	return nil
}

func (s *Sender) sendFeishu(channel model.NotificationChannel, payload model.NotificationPayload) error {
	var config model.FeishuConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("unmarshal feishu config: %w", err)
	}

	if config.WebhookURL == "" {
		return fmt.Errorf("feishu webhook URL is required")
	}

	// 构造飞书消息格式
	title := "部署通知"
	if payload.Status == "failed" {
		title = "部署失败通知"
	} else if payload.Status == "success" {
		title = "部署成功通知"
	}

	content := map[string]any{
		"tag": "div",
		"text": fmt.Sprintf(
			"项目: %s\n分支: %s\n状态: %s\n触发方式: %s\n错误信息: %s\n时间: %s",
			payload.ProjectName,
			payload.Branch,
			payload.Status,
			payload.TriggerType,
			payload.ErrorMessage,
			time.Now().Format("2006-01-02 15:04:05"),
		),
	}

	message := map[string]any{
		"msg_type": "interactive",
		"card": map[string]any{
			"header": map[string]any{
				"title": map[string]any{
					"tag":     "plain_text",
					"content": title,
				},
			},
			"elements": []any{
				map[string]any{
					"tag": "div",
					"text": map[string]any{
						"tag":     "lark_md",
						"content": strings.ReplaceAll(content["text"].(string), "\n", "\\n"),
					},
				},
			},
		},
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("feishu returned status %d: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("feishu notification sent", "channel_id", channel.ID, "run_id", payload.RunID)
	return nil
}

func (s *Sender) sendEmail(channel model.NotificationChannel, payload model.NotificationPayload) error {
	var config model.EmailConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("unmarshal email config: %w", err)
	}

	if config.SMTPHost == "" || config.Username == "" || config.To == "" {
		return fmt.Errorf("email config is incomplete")
	}

	smtpPort := config.SMTPPort
	if smtpPort == 0 {
		smtpPort = 587
	}

	// 构造邮件内容
	subject := config.Subject
	if subject == "" {
		subject = fmt.Sprintf("部署通知: %s - %s", payload.ProjectName, payload.Status)
	}

	body := fmt.Sprintf(
		"部署通知\n\n"+
			"项目: %s\n"+
			"分支: %s\n"+
			"状态: %s\n"+
			"触发方式: %s\n"+
			"错误信息: %s\n"+
			"时间: %s\n",
		payload.ProjectName,
		payload.Branch,
		payload.Status,
		payload.TriggerType,
		payload.ErrorMessage,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// 发送邮件
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
	smtpAddr := fmt.Sprintf("%s:%d", config.SMTPHost, smtpPort)

	to := strings.Split(config.To, ",")
	for _, recipient := range to {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}

		from := config.From
		if from == "" {
			from = config.Username
		}

		message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, recipient, subject, body)

		err := smtp.SendMail(smtpAddr, auth, from, []string{recipient}, []byte(message))
		if err != nil {
			s.logger.Error("failed to send email", "recipient", recipient, "error", err)
			return fmt.Errorf("send email to %s: %w", recipient, err)
		}
	}

	s.logger.Info("email notification sent", "channel_id", channel.ID, "run_id", payload.RunID)
	return nil
}

func (s *Sender) TestChannel(channel model.NotificationChannel, input model.TestNotificationInput) error {
	if input.Title == "" {
		input.Title = "通知渠道测试"
	}
	if input.Content == "" {
		input.Content = "这是一条测试消息，如果您收到此消息，说明通知渠道配置正确。"
	}

	testPayload := model.NotificationPayload{
		RunID:        0,
		Status:       "testing",
		ProjectID:    0,
		ProjectName:  "测试项目",
		RepoURL:      "https://example.com/repo.git",
		Branch:       "main",
		TriggerType:  "manual",
		TriggerRef:   "test",
		ErrorMessage: input.Content,
		SentAt:       time.Now().UTC().Format(time.RFC3339),
	}

	// 为不同渠道定制测试消息内容
	switch channel.Type {
	case model.ChannelTypeWeChat:
		return s.sendWeChat(channel, testPayload)
	case model.ChannelTypeDingTalk:
		return s.sendDingTalk(channel, testPayload)
	case model.ChannelTypeFeishu:
		return s.sendFeishu(channel, testPayload)
	case model.ChannelTypeWebhook:
		// Webhook使用原始payload
		return s.sendWebhook(channel, model.NotificationPayload{
			RunID:        0,
			Status:       "testing",
			ProjectID:    0,
			ProjectName:  input.Title,
			RepoURL:      "https://example.com/repo.git",
			Branch:       "main",
			TriggerType:  "manual",
			TriggerRef:   "test",
			ErrorMessage: input.Content,
			SentAt:       time.Now().UTC().Format(time.RFC3339),
		})
	case model.ChannelTypeEmail:
		return s.sendEmail(channel, testPayload)
	default:
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}

func generateHMACSignature(secret string, data []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func generateWeChatSign(key string, timestamp int64) string {
	data := fmt.Sprintf("%d", timestamp) + key
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func generateDingTalkSign(secret string, timestamp int64) string {
	data := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}