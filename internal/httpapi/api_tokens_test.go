package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"devops-pipeline/internal/config"
	cryptoutil "devops-pipeline/internal/crypto"
	"devops-pipeline/internal/model"
	"devops-pipeline/internal/pipeline"
	"devops-pipeline/internal/store"
)

func TestAPITokenAuthenticatesProtectedRoutesAndRevocation(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(store.DriverSQLite, ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	appStore := store.New(db, cryptoutil.New("test-secret"), store.DriverSQLite)
	if err := appStore.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	created, err := appStore.CreateAPIToken(ctx, model.APITokenCreateInput{Name: "agent"})
	if err != nil {
		t.Fatalf("CreateAPIToken() error = %v", err)
	}

	logger := slog.New(slog.DiscardHandler)
	executor := pipeline.NewExecutor(appStore, logger, t.TempDir(), t.TempDir(), t.TempDir())
	handler := New(appStore, executor, logger, config.Config{Secret: "test-secret"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/hosts", nil)
	req.Header.Set("Authorization", "Bearer "+created.Token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected API token to authenticate, status=%d body=%s", rec.Code, rec.Body.String())
	}

	if err := appStore.RevokeAPIToken(ctx, created.APIToken.ID); err != nil {
		t.Fatalf("RevokeAPIToken() error = %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/hosts", nil)
	req.Header.Set("Authorization", "Bearer "+created.Token)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked API token to be rejected, status=%d body=%s", rec.Code, rec.Body.String())
	}
}
