package store

import (
	"context"
	"strings"
	"testing"

	"devops-pipeline/internal/model"
)

func TestAppendRunLogQuerySQLite(t *testing.T) {
	query := appendRunLogQuery(false)

	if !strings.Contains(query, "COALESCE(log_text, '') || ?") {
		t.Fatalf("expected sqlite log append query, got %q", query)
	}
	if strings.Contains(query, "CONCAT(") {
		t.Fatalf("did not expect mysql concat query, got %q", query)
	}
}

func TestAppendRunLogQueryMySQL(t *testing.T) {
	query := appendRunLogQuery(true)

	if !strings.Contains(query, "CONCAT(COALESCE(log_text, ''), ?)") {
		t.Fatalf("expected mysql log append query, got %q", query)
	}
	if strings.Contains(query, "|| ?") {
		t.Fatalf("did not expect sqlite concatenation operator, got %q", query)
	}
}

func TestRunSelectQueryWithoutLogText(t *testing.T) {
	query := runSelectQuery(false)

	if !strings.Contains(query, "'' AS log_text") {
		t.Fatalf("expected summary run query to avoid loading log_text, got %q", query)
	}
	if strings.Contains(query, "pipeline_runs.log_text AS log_text") {
		t.Fatalf("did not expect summary run query to select pipeline_runs.log_text, got %q", query)
	}
}

func TestRunSelectQueryWithLogText(t *testing.T) {
	query := runSelectQuery(true)

	if !strings.Contains(query, "pipeline_runs.log_text AS log_text") {
		t.Fatalf("expected detail run query to include log_text, got %q", query)
	}
}

func TestFinalizeActiveRunOnlyUpdatesQueuedOrRunning(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	queued, err := store.CreateRun(ctx, model.RunCreateInput{
		ProjectID:   createRunTestProject(t, store),
		Status:      model.RunStatusQueued,
		TriggerType: model.TriggerTypeManual,
		TriggerRef:  "manual",
	})
	if err != nil {
		t.Fatalf("create queued run: %v", err)
	}

	finalized, err := store.FinalizeActiveRun(ctx, queued.ID, model.RunStatusFailed, "timeout")
	if err != nil {
		t.Fatalf("finalize active run: %v", err)
	}
	if !finalized {
		t.Fatalf("expected queued run to be finalized")
	}

	finalized, err = store.FinalizeActiveRun(ctx, queued.ID, model.RunStatusSuccess, "")
	if err != nil {
		t.Fatalf("finalize terminal run: %v", err)
	}
	if finalized {
		t.Fatalf("did not expect terminal run to be finalized again")
	}

	run, err := store.GetRun(ctx, queued.ID)
	if err != nil {
		t.Fatalf("get finalized run: %v", err)
	}
	if run.Status != model.RunStatusFailed || run.ErrorMessage != "timeout" {
		t.Fatalf("terminal run was unexpectedly overwritten: %+v", run)
	}
}

func createRunTestProject(t *testing.T, store *Store) int64 {
	t.Helper()

	project, err := store.CreateProject(context.Background(), model.ProjectUpsert{
		Name:        "test project",
		RepoURL:     "https://example.com/repo.git",
		Branch:      "main",
		Description: "test",
		GitAuthType: model.GitAuthTypeNone,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	return project.ID
}
