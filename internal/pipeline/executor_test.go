package pipeline

import (
	"errors"
	"fmt"
	"testing"
)

func TestUserVisibleErrorMessagePrefersDetailedCommandOutput(t *testing.T) {
	err := fmt.Errorf("docker build stage failed: %w", newCommandFailureError(
		errors.New("command failed: exit status 125"),
		[]string{
			"stage build: trying image source=node:20-alpine",
			"command failed: exit status 125",
			"npm ERR! Missing script: \"build\"",
		},
	))

	got := userVisibleErrorMessage(err)

	if got != `npm ERR! Missing script: "build"` {
		t.Fatalf("expected concrete command output, got %q", got)
	}
}

func TestUserVisibleErrorMessageFallsBackToWrappedError(t *testing.T) {
	err := fmt.Errorf("docker build stage failed: %w", errors.New("command failed: exit status 125"))

	got := userVisibleErrorMessage(err)

	if got != "docker build stage failed: command failed: exit status 125" {
		t.Fatalf("expected wrapped error fallback, got %q", got)
	}
}

func TestSummarizeCommandFailureDetailSkipsGenericLines(t *testing.T) {
	got := summarizeCommandFailureDetail([]string{
		"stage build: trying image source=node:20-alpine",
		"command failed: exit status 125",
		"",
		"docker: Error response from daemon: pull access denied for private-image",
	})

	if got != "docker: Error response from daemon: pull access denied for private-image" {
		t.Fatalf("expected concrete docker error, got %q", got)
	}
}
