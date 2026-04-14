package notification

import (
	"strings"
	"testing"

	"devops-pipeline/internal/model"
)

func TestBuildExecutionInfoPlainSuccessOmitsFailureStage(t *testing.T) {
	payload := model.NotificationPayload{
		RunID:           114,
		RunURL:          "https://devops.jimuqu.com/?view=logs&run_id=114",
		Status:          "success",
		TriggerType:     "webhook",
		Stage:           "build",
		DurationSeconds: 886,
	}

	got := buildExecutionInfoPlain(payload)

	if strings.Contains(got, "失败阶段") {
		t.Fatalf("expected success execution info to omit failure stage, got %q", got)
	}
	if !strings.Contains(got, "运行记录: 114 (https://devops.jimuqu.com/?view=logs&run_id=114)") {
		t.Fatalf("expected run link in execution info, got %q", got)
	}
}

func TestBuildNotificationErrorPlainSuccessOmitsErrorLine(t *testing.T) {
	payload := model.NotificationPayload{
		Status:       "success",
		ErrorMessage: "docker build stage failed: command failed: exit status 125",
	}

	got := buildNotificationErrorPlain(payload)

	if got != "" {
		t.Fatalf("expected success notification to omit error line, got %q", got)
	}
}

func TestBuildExecutionInfoMarkdownSuccessOmitsFailureStage(t *testing.T) {
	payload := model.NotificationPayload{
		RunID:           114,
		RunURL:          "https://devops.jimuqu.com/?view=logs&run_id=114",
		Status:          "success",
		TriggerType:     "webhook",
		Stage:           "build",
		DurationSeconds: 886,
	}

	got := buildExecutionInfoMarkdown(payload)

	if strings.Contains(got, "失败阶段") {
		t.Fatalf("expected success execution info to omit failure stage, got %q", got)
	}
	if !strings.Contains(got, "**运行记录**: [114](https://devops.jimuqu.com/?view=logs&run_id=114)") {
		t.Fatalf("expected run link in execution info, got %q", got)
	}
}

func TestBuildNotificationErrorMarkdownSuccessOmitsErrorLine(t *testing.T) {
	payload := model.NotificationPayload{
		Status:       "success",
		ErrorMessage: "docker build stage failed: command failed: exit status 125",
	}

	got := buildNotificationErrorMarkdown(payload)

	if got != "" {
		t.Fatalf("expected success notification to omit markdown error info, got %q", got)
	}
}
