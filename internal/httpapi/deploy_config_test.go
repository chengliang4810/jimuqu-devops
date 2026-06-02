package httpapi

import (
	"strings"
	"testing"

	"devops-pipeline/internal/model"
)

func validDeployConfigInput() model.DeployConfigUpsert {
	return model.DeployConfigUpsert{
		HostID:             1,
		BuildImage:         "node:20",
		BuildCommands:      []string{"pnpm build"},
		ArtifactFilterMode: model.ArtifactFilterInclude,
		ArtifactRules:      []string{"dist"},
		RemoteSaveDir:      "/data/releases",
		RemoteDeployDir:    "/data/app",
		VersionCount:       5,
		TimeoutSeconds:     1800,
	}
}

func TestValidateDeployConfigInputAcceptsDeploySyncModes(t *testing.T) {
	for _, mode := range []string{"", model.DeploySyncModeOverwrite, model.DeploySyncModeClean} {
		input := validDeployConfigInput()
		input.DeploySyncMode = mode

		if err := validateDeployConfigInput(input); err != nil {
			t.Fatalf("expected deploy_sync_mode %q to be accepted, got %v", mode, err)
		}
	}
}

func TestValidateDeployConfigInputRejectsInvalidDeploySyncMode(t *testing.T) {
	input := validDeployConfigInput()
	input.DeploySyncMode = "replace"

	err := validateDeployConfigInput(input)
	if err == nil {
		t.Fatal("expected invalid deploy_sync_mode to be rejected")
	}
	if !strings.Contains(err.Error(), "deploy_sync_mode") {
		t.Fatalf("expected deploy_sync_mode error, got %v", err)
	}
}
