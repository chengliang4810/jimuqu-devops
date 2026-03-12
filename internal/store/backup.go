package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"devops-pipeline/internal/model"
)

func (s *Store) ExportBackup(ctx context.Context, repoURL, version string) (model.BackupData, error) {
	hosts, err := s.ListHosts(ctx)
	if err != nil {
		return model.BackupData{}, err
	}

	projects, err := s.ListProjects(ctx)
	if err != nil {
		return model.BackupData{}, err
	}

	channels, err := s.ListNotificationChannels(ctx)
	if err != nil {
		return model.BackupData{}, err
	}

	settings, err := s.ListSettings(ctx)
	if err != nil {
		return model.BackupData{}, err
	}

	backup := model.BackupData{
		Meta: model.BackupMeta{
			SchemaVersion: 1,
			ExportedAt:    nowString(),
			RepoURL:       repoURL,
			Version:       version,
		},
		Hosts:                make([]model.BackupHost, 0, len(hosts)),
		Projects:             make([]model.BackupProjectBundle, 0, len(projects)),
		NotificationChannels: make([]model.NotificationChannelWithConfig, 0, len(channels)),
		Settings:             settings,
	}

	for _, host := range hosts {
		backup.Hosts = append(backup.Hosts, model.BackupHost{
			ID:        host.ID,
			SortOrder: host.SortOrder,
			Name:      host.Name,
			Address:   host.Address,
			Port:      host.Port,
			Username:  host.Username,
			Password:  host.Password,
		})
	}

	for _, project := range projects {
		detail, err := s.GetProjectDetail(ctx, project.ID)
		if err != nil {
			return model.BackupData{}, err
		}

		bundle := model.BackupProjectBundle{
			Project: model.BackupProject{
				ID:           detail.Project.ID,
				SortOrder:    detail.Project.SortOrder,
				Name:         detail.Project.Name,
				RepoURL:      detail.Project.RepoURL,
				Branch:       detail.Project.Branch,
				Description:  detail.Project.Description,
				WebhookToken: detail.Project.WebhookToken,
				GitAuthType:  detail.Project.GitAuthType,
				GitUsername:  optionalString(detail.Project.GitUsername),
				GitPassword:  optionalString(detail.Project.GitPassword),
				GitSSHKey:    optionalString(detail.Project.GitSSHKey),
			},
		}

		if detail.DeployConfig != nil {
			bundle.DeployConfig = &model.BackupDeployConfig{
				ProjectID:             detail.DeployConfig.ProjectID,
				HostID:                detail.DeployConfig.HostID,
				BuildImage:            detail.DeployConfig.BuildImage,
				BuildCommands:         detail.DeployConfig.BuildCommands,
				ArtifactFilterMode:    detail.DeployConfig.ArtifactFilterMode,
				ArtifactRules:         detail.DeployConfig.ArtifactRules,
				RemoteSaveDir:         detail.DeployConfig.RemoteSaveDir,
				RemoteDeployDir:       detail.DeployConfig.RemoteDeployDir,
				PreDeployCommands:     detail.DeployConfig.PreDeployCommands,
				PostDeployCommands:    detail.DeployConfig.PostDeployCommands,
				TimeoutSeconds:        detail.DeployConfig.TimeoutSeconds,
				NotifyWebhookURL:      detail.DeployConfig.NotifyWebhookURL,
				NotifyBearerToken:     optionalString(detail.DeployConfig.NotifyBearerToken),
				NotificationChannelID: detail.DeployConfig.NotificationChannelID,
			}
		}

		backup.Projects = append(backup.Projects, bundle)
	}

	for _, channel := range channels {
		detail, err := s.GetNotificationChannelWithConfig(ctx, channel.ID)
		if err != nil {
			return model.BackupData{}, err
		}
		backup.NotificationChannels = append(backup.NotificationChannels, detail)
	}

	return backup, nil
}

func (s *Store) ImportBackup(ctx context.Context, backup model.BackupData) (model.BackupRestoreResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.BackupRestoreResult{}, fmt.Errorf("begin backup import transaction: %w", err)
	}
	defer tx.Rollback()

	rowsAffected := map[string]int{
		"hosts":                 0,
		"projects":              0,
		"deploy_configs":        0,
		"notification_channels": 0,
		"settings":              0,
	}

	for _, statement := range []string{
		`DELETE FROM pipeline_runs`,
		`DELETE FROM deploy_configs`,
		`DELETE FROM projects`,
		`DELETE FROM hosts`,
		`DELETE FROM notification_channels`,
		`DELETE FROM settings`,
	} {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("clear data before restore: %w", err)
		}
	}

	now := nowString()

	for _, host := range backup.Hosts {
		passwordCipher, err := s.cipher.Encrypt(host.Password)
		if err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("encrypt host password: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO hosts (id, sort_order, name, address, port, username, password_cipher, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			host.ID, host.SortOrder, host.Name, host.Address, host.Port, host.Username, passwordCipher, now, now,
		); err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("restore host %d: %w", host.ID, err)
		}
		rowsAffected["hosts"] += 1
	}

	for _, channel := range backup.NotificationChannels {
		configJSON, err := json.Marshal(channel.ConfigMap)
		if err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("marshal notification channel config: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO notification_channels (id, sort_order, name, type, is_default, remark, config_json, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			channel.ID,
			channel.SortOrder,
			channel.Name,
			channel.Type,
			boolToInt(channel.IsDefault),
			channel.Remark,
			string(configJSON),
			now,
			now,
		); err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("restore notification channel %d: %w", channel.ID, err)
		}
		rowsAffected["notification_channels"] += 1
	}

	for _, setting := range backup.Settings {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO settings (` + "`key`" + `, ` + "`value`" + `, created_at, updated_at)
			 VALUES (?, ?, ?, ?)`,
			setting.Key, setting.Value, now, now,
		); err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("restore setting %s: %w", setting.Key, err)
		}
		rowsAffected["settings"] += 1
	}

	for _, bundle := range backup.Projects {
		project := bundle.Project
		gitUsernameCipher, err := s.cipher.Encrypt(valueOrEmpty(project.GitUsername))
		if err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("encrypt git username: %w", err)
		}
		gitPasswordCipher, err := s.cipher.Encrypt(valueOrEmpty(project.GitPassword))
		if err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("encrypt git password: %w", err)
		}
		gitSSHKeyCipher, err := s.cipher.Encrypt(valueOrEmpty(project.GitSSHKey))
		if err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("encrypt git ssh key: %w", err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO projects (id, sort_order, name, repo_url, branch, description, webhook_token, git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			project.ID,
			project.SortOrder,
			project.Name,
			project.RepoURL,
			project.Branch,
			project.Description,
			project.WebhookToken,
			project.GitAuthType,
			gitUsernameCipher,
			gitPasswordCipher,
			gitSSHKeyCipher,
			now,
			now,
		); err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("restore project %d: %w", project.ID, err)
		}
		rowsAffected["projects"] += 1

		if bundle.DeployConfig == nil {
			continue
		}

		tokenCipher, err := s.cipher.Encrypt(valueOrEmpty(bundle.DeployConfig.NotifyBearerToken))
		if err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("encrypt notify bearer token: %w", err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO deploy_configs (
				project_id, host_id, build_image, build_commands_json, artifact_filter_mode,
				artifact_rules_json, remote_save_dir, remote_deploy_dir, pre_deploy_commands_json,
				post_deploy_commands_json, timeout_seconds, notify_webhook_url, notify_token_cipher, notification_channel_id, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			bundle.DeployConfig.ProjectID,
			bundle.DeployConfig.HostID,
			bundle.DeployConfig.BuildImage,
			mustMarshal(bundle.DeployConfig.BuildCommands),
			bundle.DeployConfig.ArtifactFilterMode,
			mustMarshal(bundle.DeployConfig.ArtifactRules),
			bundle.DeployConfig.RemoteSaveDir,
			bundle.DeployConfig.RemoteDeployDir,
			mustMarshal(bundle.DeployConfig.PreDeployCommands),
			mustMarshal(bundle.DeployConfig.PostDeployCommands),
			bundle.DeployConfig.TimeoutSeconds,
			bundle.DeployConfig.NotifyWebhookURL,
			tokenCipher,
			bundle.DeployConfig.NotificationChannelID,
			now,
			now,
		); err != nil {
			return model.BackupRestoreResult{}, fmt.Errorf("restore deploy config for project %d: %w", bundle.Project.ID, err)
		}
		rowsAffected["deploy_configs"] += 1
	}

	for _, table := range []string{"hosts", "projects", "notification_channels", "deploy_configs", "pipeline_runs"} {
		if err := s.resetAutoIncrement(ctx, tx, table); err != nil {
			return model.BackupRestoreResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.BackupRestoreResult{}, fmt.Errorf("commit backup import transaction: %w", err)
	}

	return model.BackupRestoreResult{RowsAffected: rowsAffected}, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	next := value
	return &next
}

func (s *Store) resetAutoIncrement(ctx context.Context, tx *sql.Tx, table string) error {
	if s.isMySQL() {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s AUTO_INCREMENT = 1`, table)); err != nil {
			return fmt.Errorf("reset mysql auto increment for %s: %w", table, err)
		}
		return nil
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sqlite_sequence WHERE name = ?`, table); err != nil {
		return fmt.Errorf("reset sqlite sequence for %s: %w", table, err)
	}
	return nil
}
