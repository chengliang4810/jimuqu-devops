package pipeline

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"devops-pipeline/internal/model"
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

func TestDockerBuildRunArgsIdentifyDeploymentContainer(t *testing.T) {
	args := dockerBuildRunArgs(42, "/tmp/source", "node:20-alpine", "npm run build", nil, nil)
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "--name jimuqu-devops-build-42") {
		t.Fatalf("expected run-specific container name, got %q", joined)
	}
	if !strings.Contains(joined, "--label jimuqu-devops.build=true") {
		t.Fatalf("expected managed build label, got %q", joined)
	}
}

func TestDeploySyncCommandOverwriteKeepsExistingTargetFiles(t *testing.T) {
	command := buildDeploySyncCommand(model.DeploySyncModeOverwrite, "/tmp/run", "/data/app")

	if strings.Contains(command, "rm -rf") || strings.Contains(command, "find ") {
		t.Fatalf("overwrite mode must not delete target contents, got %q", command)
	}
	if !strings.Contains(command, "cp -a '/tmp/run'/. '/data/app'/") {
		t.Fatalf("expected overwrite mode to copy artifacts into target, got %q", command)
	}
}

func TestDeploySyncCommandCleanRemovesExistingTargetFiles(t *testing.T) {
	command := buildDeploySyncCommand(model.DeploySyncModeClean, "/tmp/run", "/data/app")

	if !strings.Contains(command, "find '/data/app' -mindepth 1 -maxdepth 1 -exec rm -rf {} +") {
		t.Fatalf("expected clean mode to delete target contents, got %q", command)
	}
	if !strings.Contains(command, "cp -a '/tmp/run'/. '/data/app'/") {
		t.Fatalf("expected clean mode to copy artifacts into target, got %q", command)
	}
}
