package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"devops-pipeline/internal/model"
)

const aiSettingsSingletonID = 1

func (s *Store) GetAISettings(ctx context.Context) (model.AISettings, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT enabled, protocol, base_url, api_key, model, user_agent, created_at, updated_at
		 FROM ai_settings
		 WHERE id = ?`,
		aiSettingsSingletonID,
	)

	settings, err := scanAISettings(row)
	if errors.Is(err, ErrNotFound) {
		return model.DefaultAISettings, nil
	}
	return settings, err
}

func (s *Store) SetAISettings(ctx context.Context, input model.AISettings) (model.AISettings, error) {
	now := nowString()
	query := `INSERT INTO ai_settings (id, enabled, protocol, base_url, api_key, model, user_agent, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			enabled = excluded.enabled,
			protocol = excluded.protocol,
			base_url = excluded.base_url,
			api_key = excluded.api_key,
			model = excluded.model,
			user_agent = excluded.user_agent,
			updated_at = excluded.updated_at`
	if s.isMySQL() {
		query = `INSERT INTO ai_settings (id, enabled, protocol, base_url, api_key, model, user_agent, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				enabled = VALUES(enabled),
				protocol = VALUES(protocol),
				base_url = VALUES(base_url),
				api_key = VALUES(api_key),
				model = VALUES(model),
				user_agent = VALUES(user_agent),
				updated_at = VALUES(updated_at)`
	}

	if input.Protocol == "" {
		input.Protocol = model.AIProtocolOpenAI
	}

	if _, err := s.db.ExecContext(
		ctx,
		query,
		aiSettingsSingletonID,
		boolToInt(input.Enabled),
		input.Protocol,
		input.BaseURL,
		input.APIKey,
		input.Model,
		input.UserAgent,
		now,
		now,
	); err != nil {
		return model.AISettings{}, fmt.Errorf("set ai settings: %w", err)
	}

	return s.GetAISettings(ctx)
}

func scanAISettings(scan scanner) (model.AISettings, error) {
	var (
		settings      model.AISettings
		enabled       int
		createdAtStr  sql.NullString
		updatedAtStr  sql.NullString
	)

	err := scan.Scan(
		&enabled,
		&settings.Protocol,
		&settings.BaseURL,
		&settings.APIKey,
		&settings.Model,
		&settings.UserAgent,
		&createdAtStr,
		&updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.AISettings{}, ErrNotFound
	}
	if err != nil {
		return model.AISettings{}, fmt.Errorf("scan ai settings: %w", err)
	}

	settings.Enabled = enabled == 1
	if settings.Protocol == "" {
		settings.Protocol = model.AIProtocolOpenAI
	}
	if createdAtStr.Valid {
		settings.CreatedAt, err = parseTime(createdAtStr.String)
		if err != nil {
			return model.AISettings{}, err
		}
	}
	if updatedAtStr.Valid {
		settings.UpdatedAt, err = parseTime(updatedAtStr.String)
		if err != nil {
			return model.AISettings{}, err
		}
	}

	return settings, nil
}
