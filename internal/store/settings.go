package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"devops-pipeline/internal/model"
)

func (s *Store) ListSettings(ctx context.Context) ([]model.Setting, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT ` + "`key`" + `, ` + "`value`" + `, created_at, updated_at
		 FROM settings
		 ORDER BY ` + "`key`" + ` ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	defer rows.Close()

	settingsByKey := make(map[string]model.Setting, len(model.DefaultSettings))
	for rows.Next() {
		setting, err := scanSetting(rows)
		if err != nil {
			return nil, err
		}
		settingsByKey[setting.Key] = setting
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(model.DefaultSettings))
	for key := range model.DefaultSettings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	settings := make([]model.Setting, 0, len(model.DefaultSettings))
	for _, key := range keys {
		defaultValue := model.DefaultSettings[key]
		if setting, ok := settingsByKey[key]; ok {
			settings = append(settings, setting)
			continue
		}
		settings = append(settings, model.Setting{Key: key, Value: defaultValue})
	}

	return settings, nil
}

func (s *Store) GetSetting(ctx context.Context, key string) (model.Setting, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT ` + "`key`" + `, ` + "`value`" + `, created_at, updated_at
		 FROM settings
		 WHERE ` + "`key`" + ` = ?`,
		key,
	)

	setting, err := scanSetting(row)
	if errors.Is(err, ErrNotFound) {
		defaultValue, ok := model.DefaultSettings[key]
		if !ok {
			return model.Setting{}, ErrNotFound
		}
		return model.Setting{Key: key, Value: defaultValue}, nil
	}
	return setting, err
}

func (s *Store) GetSettingValue(ctx context.Context, key string) (string, error) {
	setting, err := s.GetSetting(ctx, key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *Store) SetSetting(ctx context.Context, key, value string) (model.Setting, error) {
	now := nowString()
	query := `INSERT INTO settings (` + "`key`" + `, ` + "`value`" + `, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(` + "`key`" + `) DO UPDATE SET ` + "`value`" + ` = excluded.` + "`value`" + `, updated_at = excluded.updated_at`
	if s.isMySQL() {
		query = `INSERT INTO settings (` + "`key`" + `, ` + "`value`" + `, created_at, updated_at)
			VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE ` + "`value`" + ` = VALUES(` + "`value`" + `), updated_at = VALUES(updated_at)`
	}
	_, err := s.db.ExecContext(ctx, query, key, value, now, now)
	if err != nil {
		return model.Setting{}, fmt.Errorf("set setting %s: %w", key, err)
	}
	return s.GetSetting(ctx, key)
}

func (s *Store) UpdateAdminUsername(ctx context.Context, username string) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE admin_users SET username = ?, updated_at = ?`,
		username, nowString(),
	)
	if err != nil {
		return fmt.Errorf("update admin username: %w", err)
	}
	return nil
}

func (s *Store) ClearRuns(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM pipeline_runs`)
	if err != nil {
		return 0, fmt.Errorf("clear runs: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read cleared runs count: %w", err)
	}
	return affected, nil
}

func (s *Store) DeleteRunsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := s.db.ExecContext(
		ctx,
		`DELETE FROM pipeline_runs
		 WHERE COALESCE(finished_at, started_at, created_at) < ?`,
		cutoff.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return 0, fmt.Errorf("delete expired runs: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read deleted runs count: %w", err)
	}
	return affected, nil
}

func (s *Store) ApplyRunRetention(ctx context.Context) error {
	value, err := s.GetSettingValue(ctx, model.SettingRunRetentionDays)
	if err != nil {
		return err
	}

	days, err := strconv.Atoi(value)
	if err != nil || days <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	_, err = s.DeleteRunsOlderThan(ctx, cutoff)
	return err
}

func scanSetting(scan scanner) (model.Setting, error) {
	var (
		setting      model.Setting
		createdAtStr sql.NullString
		updatedAtStr sql.NullString
	)

	err := scan.Scan(&setting.Key, &setting.Value, &createdAtStr, &updatedAtStr)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Setting{}, ErrNotFound
	}
	if err != nil {
		return model.Setting{}, fmt.Errorf("scan setting: %w", err)
	}

	if createdAtStr.Valid {
		setting.CreatedAt, err = parseTime(createdAtStr.String)
		if err != nil {
			return model.Setting{}, err
		}
	}
	if updatedAtStr.Valid {
		setting.UpdatedAt, err = parseTime(updatedAtStr.String)
		if err != nil {
			return model.Setting{}, err
		}
	}

	return setting, nil
}
