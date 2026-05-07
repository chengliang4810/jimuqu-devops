package store

import (
	"context"
	"testing"
	"time"

	"devops-pipeline/internal/model"
)

func TestCreateAPITokenListsMetadataOnlyAndValidates(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateAPIToken(ctx, model.APITokenCreateInput{Name: "agent"})
	if err != nil {
		t.Fatalf("CreateAPIToken() error = %v", err)
	}
	if created.Token == "" || created.APIToken.Prefix == "" {
		t.Fatalf("expected token and prefix, got %#v", created)
	}

	tokens, err := store.ListAPITokens(ctx)
	if err != nil {
		t.Fatalf("ListAPITokens() error = %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Prefix == created.Token {
		t.Fatalf("list returned full token")
	}

	validated, err := store.ValidateAPIToken(ctx, created.Token)
	if err != nil {
		t.Fatalf("ValidateAPIToken() error = %v", err)
	}
	if validated.Name != "agent" || validated.LastUsedAt == nil {
		t.Fatalf("unexpected validated token: %#v", validated)
	}
}

func TestRevokedAPITokenCannotValidate(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateAPIToken(ctx, model.APITokenCreateInput{Name: "agent"})
	if err != nil {
		t.Fatalf("CreateAPIToken() error = %v", err)
	}
	if err := store.RevokeAPIToken(ctx, created.APIToken.ID); err != nil {
		t.Fatalf("RevokeAPIToken() error = %v", err)
	}

	if _, err := store.ValidateAPIToken(ctx, created.Token); err == nil {
		t.Fatalf("expected revoked token validation to fail")
	}
}

func TestExpiredAPITokenCannotValidate(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	expiresAt := time.Now().UTC().Add(-time.Hour)

	created, err := store.CreateAPIToken(ctx, model.APITokenCreateInput{Name: "agent", ExpiresAt: &expiresAt})
	if err != nil {
		t.Fatalf("CreateAPIToken() error = %v", err)
	}

	if _, err := store.ValidateAPIToken(ctx, created.Token); err == nil {
		t.Fatalf("expected expired token validation to fail")
	}
}
