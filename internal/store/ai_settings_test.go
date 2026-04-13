package store

import (
	"context"
	"testing"

	cryptoutil "devops-pipeline/internal/crypto"
	"devops-pipeline/internal/model"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := Open(DriverSQLite, "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	store := New(db, cryptoutil.New("test-secret"), DriverSQLite)
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return store
}

func TestAISettingsDefault(t *testing.T) {
	store := newTestStore(t)

	settings, err := store.GetAISettings(context.Background())
	if err != nil {
		t.Fatalf("get ai settings: %v", err)
	}

	if settings.Enabled {
		t.Fatalf("expected ai settings disabled by default")
	}
	if settings.Protocol != model.AIProtocolOpenAI {
		t.Fatalf("expected default protocol %q, got %q", model.AIProtocolOpenAI, settings.Protocol)
	}
	if settings.BaseURL != "" || settings.APIKey != "" || settings.Model != "" || settings.UserAgent != "" {
		t.Fatalf("expected default ai settings fields to be empty, got %+v", settings)
	}
}

func TestAISettingsRoundTrip(t *testing.T) {
	store := newTestStore(t)

	saved, err := store.SetAISettings(context.Background(), model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolOpenAI,
		BaseURL:  "https://example.com/v1",
		APIKey:   "plain-api-key",
		Model:    "gpt-4.1-mini",
		UserAgent: "Codex Desktop/0.115.0-alpha.11 (Windows 10.0.22621; x86_64) unknown (Codex Desktop; 26.311.21342)",
	})
	if err != nil {
		t.Fatalf("set ai settings: %v", err)
	}

	if !saved.Enabled || saved.BaseURL != "https://example.com/v1" || saved.APIKey != "plain-api-key" || saved.Model != "gpt-4.1-mini" || saved.UserAgent == "" {
		t.Fatalf("unexpected saved ai settings: %+v", saved)
	}

	loaded, err := store.GetAISettings(context.Background())
	if err != nil {
		t.Fatalf("reload ai settings: %v", err)
	}

	if loaded != saved {
		t.Fatalf("expected reloaded ai settings to match saved settings:\nwant: %+v\ngot:  %+v", saved, loaded)
	}
}

func TestBackupExportImportIncludesAISettings(t *testing.T) {
	source := newTestStore(t)
	ctx := context.Background()

	saved, err := source.SetAISettings(ctx, model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolOpenAI,
		BaseURL:  "https://proxy.example.com/openai",
		APIKey:   "visible-key",
		Model:    "gpt-5-mini",
		UserAgent: "Codex Desktop/0.115.0-alpha.11 (Windows 10.0.22621; x86_64) unknown (Codex Desktop; 26.311.21342)",
	})
	if err != nil {
		t.Fatalf("set ai settings: %v", err)
	}

	backup, err := source.ExportBackup(ctx, "https://repo.example.com/demo", "v1.0.0")
	if err != nil {
		t.Fatalf("export backup: %v", err)
	}

	if backup.AISettings == nil {
		t.Fatalf("expected backup to include ai settings")
	}
	if *backup.AISettings != saved {
		t.Fatalf("expected backup ai settings to match saved settings:\nwant: %+v\ngot:  %+v", saved, *backup.AISettings)
	}

	target := newTestStore(t)
	if _, err := target.ImportBackup(ctx, backup); err != nil {
		t.Fatalf("import backup: %v", err)
	}

	restored, err := target.GetAISettings(ctx)
	if err != nil {
		t.Fatalf("get restored ai settings: %v", err)
	}

	if restored.Enabled != saved.Enabled || restored.Protocol != saved.Protocol || restored.BaseURL != saved.BaseURL || restored.APIKey != saved.APIKey || restored.Model != saved.Model || restored.UserAgent != saved.UserAgent {
		t.Fatalf("expected restored ai settings payload to match saved settings:\nwant: %+v\ngot:  %+v", saved, restored)
	}
	if restored.CreatedAt.IsZero() || restored.UpdatedAt.IsZero() {
		t.Fatalf("expected restored ai settings timestamps to be populated, got %+v", restored)
	}
}
