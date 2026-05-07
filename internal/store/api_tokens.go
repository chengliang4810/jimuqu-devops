package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"devops-pipeline/internal/model"
)

func (s *Store) CreateAPIToken(ctx context.Context, input model.APITokenCreateInput) (model.APITokenCreateResponse, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.APITokenCreateResponse{}, fmt.Errorf("api token name is required")
	}

	plainToken, err := generateAPIToken()
	if err != nil {
		return model.APITokenCreateResponse{}, fmt.Errorf("generate api token: %w", err)
	}

	now := nowString()
	tokenHash := hashAPIToken(plainToken)
	tokenPrefix := tokenDisplayPrefix(plainToken)
	var expiresAt any
	if input.ExpiresAt != nil && !input.ExpiresAt.IsZero() {
		expiresAt = input.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}

	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO api_tokens (name, token_hash, token_prefix, expires_at, revoked_at, last_used_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, NULL, NULL, ?, ?)`,
		name, tokenHash, tokenPrefix, expiresAt, now, now,
	)
	if err != nil {
		return model.APITokenCreateResponse{}, fmt.Errorf("insert api token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.APITokenCreateResponse{}, fmt.Errorf("get api token id: %w", err)
	}

	apiToken, err := s.GetAPIToken(ctx, id)
	if err != nil {
		return model.APITokenCreateResponse{}, err
	}

	return model.APITokenCreateResponse{APIToken: apiToken, Token: plainToken}, nil
}

func (s *Store) ListAPITokens(ctx context.Context) ([]model.APIToken, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, token_prefix, expires_at, revoked_at, last_used_at, created_at, updated_at
		 FROM api_tokens
		 ORDER BY id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query api tokens: %w", err)
	}
	defer rows.Close()

	tokens := make([]model.APIToken, 0)
	for rows.Next() {
		token, err := scanAPIToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}

func (s *Store) GetAPIToken(ctx context.Context, id int64) (model.APIToken, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, token_prefix, expires_at, revoked_at, last_used_at, created_at, updated_at
		 FROM api_tokens
		 WHERE id = ?`,
		id,
	)
	return scanAPIToken(row)
}

func (s *Store) ValidateAPIToken(ctx context.Context, plainToken string) (model.APIToken, error) {
	plainToken = strings.TrimSpace(plainToken)
	if !strings.HasPrefix(plainToken, model.APITokenPrefix) {
		return model.APIToken{}, ErrNotFound
	}

	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, token_prefix, expires_at, revoked_at, last_used_at, created_at, updated_at
		 FROM api_tokens
		 WHERE token_hash = ?`,
		hashAPIToken(plainToken),
	)

	token, err := scanAPIToken(row)
	if err != nil {
		return model.APIToken{}, err
	}

	if token.RevokedAt != nil {
		return model.APIToken{}, ErrNotFound
	}
	if token.ExpiresAt != nil && time.Now().UTC().After(token.ExpiresAt.UTC()) {
		return model.APIToken{}, ErrNotFound
	}

	if err := s.MarkAPITokenUsed(ctx, token.ID); err != nil {
		return model.APIToken{}, err
	}
	now := time.Now().UTC()
	token.LastUsedAt = &now

	return token, nil
}

func (s *Store) MarkAPITokenUsed(ctx context.Context, id int64) error {
	now := nowString()
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE api_tokens SET last_used_at = ?, updated_at = ? WHERE id = ?`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("mark api token used: %w", err)
	}
	return nil
}

func (s *Store) RevokeAPIToken(ctx context.Context, id int64) error {
	now := nowString()
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE api_tokens
		 SET revoked_at = COALESCE(revoked_at, ?), updated_at = ?
		 WHERE id = ? AND revoked_at IS NULL`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("revoke api token: %w", err)
	}
	return expectDeleted(result)
}

func scanAPIToken(scan scanner) (model.APIToken, error) {
	var (
		token         model.APIToken
		expiresAtStr  sql.NullString
		revokedAtStr  sql.NullString
		lastUsedAtStr sql.NullString
		createdAtStr  string
		updatedAtStr  string
	)

	err := scan.Scan(
		&token.ID,
		&token.Name,
		&token.Prefix,
		&expiresAtStr,
		&revokedAtStr,
		&lastUsedAtStr,
		&createdAtStr,
		&updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.APIToken{}, ErrNotFound
	}
	if err != nil {
		return model.APIToken{}, fmt.Errorf("scan api token: %w", err)
	}

	if expiresAtStr.Valid && expiresAtStr.String != "" {
		expiresAt, err := parseTime(expiresAtStr.String)
		if err != nil {
			return model.APIToken{}, err
		}
		token.ExpiresAt = &expiresAt
	}
	if revokedAtStr.Valid && revokedAtStr.String != "" {
		revokedAt, err := parseTime(revokedAtStr.String)
		if err != nil {
			return model.APIToken{}, err
		}
		token.RevokedAt = &revokedAt
	}
	if lastUsedAtStr.Valid && lastUsedAtStr.String != "" {
		lastUsedAt, err := parseTime(lastUsedAtStr.String)
		if err != nil {
			return model.APIToken{}, err
		}
		token.LastUsedAt = &lastUsedAt
	}

	createdAt, err := parseTime(createdAtStr)
	if err != nil {
		return model.APIToken{}, err
	}
	updatedAt, err := parseTime(updatedAtStr)
	if err != nil {
		return model.APIToken{}, err
	}
	token.CreatedAt = createdAt
	token.UpdatedAt = updatedAt

	return token, nil
}

func generateAPIToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return model.APITokenPrefix + hex.EncodeToString(buf), nil
}

func hashAPIToken(plainToken string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(plainToken)))
	return hex.EncodeToString(sum[:])
}

func tokenDisplayPrefix(plainToken string) string {
	if len(plainToken) <= 16 {
		return plainToken
	}
	return plainToken[:16]
}
