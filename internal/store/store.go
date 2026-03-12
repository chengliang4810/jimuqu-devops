package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	cryptoutil "devops-pipeline/internal/crypto"
	"devops-pipeline/internal/model"
)

var (
	ErrNotFound = errors.New("store: not found")
	ErrConflict = errors.New("store: conflict")
)

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	for _, statement := range s.migrationStatements() {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}

	if err := s.ensureSortOrderColumn(ctx, "hosts"); err != nil {
		return err
	}
	if err := s.ensureSortOrderColumn(ctx, "projects"); err != nil {
		return err
	}
	if err := s.ensureSortOrderColumn(ctx, "notification_channels"); err != nil {
		return err
	}

	if err := s.initializeSortOrder(ctx, "hosts"); err != nil {
		return err
	}
	if err := s.initializeSortOrder(ctx, "projects"); err != nil {
		return err
	}
	if err := s.initializeSortOrder(ctx, "notification_channels"); err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateHost(ctx context.Context, input model.HostUpsert) (model.Host, error) {
	encryptedPassword, err := s.cipher.Encrypt(valueOrEmpty(input.Password))
	if err != nil {
		return model.Host{}, fmt.Errorf("encrypt host password: %w", err)
	}

	now := nowString()
	nextSortOrder, err := s.nextSortOrder(ctx, s.db, "hosts")
	if err != nil {
		return model.Host{}, err
	}
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO hosts (sort_order, name, address, port, username, password_cipher, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		nextSortOrder, input.Name, input.Address, input.Port, input.Username, encryptedPassword, now, now,
	)
	if err != nil {
		return model.Host{}, fmt.Errorf("insert host: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Host{}, fmt.Errorf("get host id: %w", err)
	}

	return s.GetHost(ctx, id)
}

func (s *Store) UpdateHost(ctx context.Context, id int64, input model.HostUpsert) (model.Host, error) {
	host, err := s.GetHost(ctx, id)
	if err != nil {
		return model.Host{}, err
	}

	encryptedPassword, err := s.cipher.Encrypt(host.Password)
	if err != nil {
		return model.Host{}, fmt.Errorf("encrypt current password: %w", err)
	}
	if input.Password != nil {
		encryptedPassword, err = s.cipher.Encrypt(*input.Password)
		if err != nil {
			return model.Host{}, fmt.Errorf("encrypt host password: %w", err)
		}
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE hosts
		 SET name = ?, address = ?, port = ?, username = ?, password_cipher = ?, updated_at = ?
		 WHERE id = ?`,
		input.Name, input.Address, input.Port, input.Username, encryptedPassword, nowString(), id,
	)
	if err != nil {
		return model.Host{}, fmt.Errorf("update host: %w", err)
	}

	return s.GetHost(ctx, id)
}

func (s *Store) DeleteHost(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM hosts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete host: %w", err)
	}
	return expectDeleted(result)
}

func (s *Store) ReorderHosts(ctx context.Context, ids []int64) error {
	return s.reorderRecords(ctx, "hosts", ids)
}

func (s *Store) GetHost(ctx context.Context, id int64) (model.Host, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, sort_order, name, address, port, username, password_cipher, created_at, updated_at
		 FROM hosts
		 WHERE id = ?`,
		id,
	)
	return s.scanHost(row)
}

func (s *Store) ListHosts(ctx context.Context) ([]model.Host, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, sort_order, name, address, port, username, password_cipher, created_at, updated_at
		 FROM hosts
		 ORDER BY sort_order DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query hosts: %w", err)
	}
	defer rows.Close()

	var hosts []model.Host
	for rows.Next() {
		host, err := s.scanHost(rows)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}

	return hosts, rows.Err()
}

func (s *Store) CreateProject(ctx context.Context, input model.ProjectUpsert) (model.Project, error) {
	now := nowString()
	token, err := randomToken()
	if err != nil {
		return model.Project{}, fmt.Errorf("generate webhook token: %w", err)
	}

	// 处理Git认证信息
	gitAuthType := input.GitAuthType
	if gitAuthType == "" {
		gitAuthType = model.GitAuthTypeNone
	}

	gitUsernameCipher, err := s.cipher.Encrypt(valueOrEmpty(input.GitUsername))
	if err != nil {
		return model.Project{}, fmt.Errorf("encrypt git username: %w", err)
	}

	gitPasswordCipher, err := s.cipher.Encrypt(valueOrEmpty(input.GitPassword))
	if err != nil {
		return model.Project{}, fmt.Errorf("encrypt git password: %w", err)
	}

	gitSSHKeyCipher, err := s.cipher.Encrypt(valueOrEmpty(input.GitSSHKey))
	if err != nil {
		return model.Project{}, fmt.Errorf("encrypt git ssh key: %w", err)
	}

	nextSortOrder, err := s.nextSortOrder(ctx, s.db, "projects")
	if err != nil {
		return model.Project{}, err
	}
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO projects (sort_order, name, repo_url, branch, description, webhook_token, git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		nextSortOrder, input.Name, input.RepoURL, input.Branch, input.Description, token, gitAuthType, gitUsernameCipher, gitPasswordCipher, gitSSHKeyCipher, now, now,
	)
	if err != nil {
		return model.Project{}, wrapProjectMutationError("insert", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Project{}, fmt.Errorf("get project id: %w", err)
	}

	return s.GetProject(ctx, id)
}

func (s *Store) UpdateProject(ctx context.Context, id int64, input model.ProjectUpsert) (model.Project, error) {
	currentProject, err := s.GetProject(ctx, id)
	if err != nil {
		return model.Project{}, err
	}

	// 处理Git认证信息
	gitAuthType := input.GitAuthType
	if gitAuthType == "" {
		gitAuthType = currentProject.GitAuthType
	}

	// 获取或加密Git认证信息
	gitUsernameCipher, err := s.cipher.Encrypt(valueOrEmpty(input.GitUsername))
	if err != nil {
		return model.Project{}, fmt.Errorf("encrypt git username: %w", err)
	}
	if input.GitUsername == nil && currentProject.GitUsername != "" {
		gitUsernameCipher, err = s.cipher.Encrypt(currentProject.GitUsername)
		if err != nil {
			return model.Project{}, fmt.Errorf("encrypt current git username: %w", err)
		}
	}

	gitPasswordCipher, err := s.cipher.Encrypt(valueOrEmpty(input.GitPassword))
	if err != nil {
		return model.Project{}, fmt.Errorf("encrypt git password: %w", err)
	}
	if input.GitPassword == nil && currentProject.GitPassword != "" {
		gitPasswordCipher, err = s.cipher.Encrypt(currentProject.GitPassword)
		if err != nil {
			return model.Project{}, fmt.Errorf("encrypt current git password: %w", err)
		}
	}

	gitSSHKeyCipher, err := s.cipher.Encrypt(valueOrEmpty(input.GitSSHKey))
	if err != nil {
		return model.Project{}, fmt.Errorf("encrypt git ssh key: %w", err)
	}
	if input.GitSSHKey == nil && currentProject.GitSSHKey != "" {
		gitSSHKeyCipher, err = s.cipher.Encrypt(currentProject.GitSSHKey)
		if err != nil {
			return model.Project{}, fmt.Errorf("encrypt current git ssh key: %w", err)
		}
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE projects
		 SET name = ?, repo_url = ?, branch = ?, description = ?, git_auth_type = ?, git_username_cipher = ?, git_password_cipher = ?, git_ssh_key_cipher = ?, updated_at = ?
		 WHERE id = ?`,
		input.Name, input.RepoURL, input.Branch, input.Description, gitAuthType, gitUsernameCipher, gitPasswordCipher, gitSSHKeyCipher, nowString(), id,
	)
	if err != nil {
		return model.Project{}, wrapProjectMutationError("update", err)
	}

	return s.GetProject(ctx, id)
}

func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return expectDeleted(result)
}

func (s *Store) ReorderProjects(ctx context.Context, ids []int64) error {
	return s.reorderRecords(ctx, "projects", ids)
}

func (s *Store) GetProject(ctx context.Context, id int64) (model.Project, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT projects.id, projects.sort_order, projects.name, projects.repo_url, projects.branch, projects.description, projects.webhook_token,
		        (CASE WHEN deploy_configs.id IS NOT NULL THEN 1 ELSE 0 END) as has_deploy_config,
		        projects.git_auth_type, projects.git_username_cipher, projects.git_password_cipher, projects.git_ssh_key_cipher,
		        projects.created_at, projects.updated_at
		 FROM projects
		 LEFT JOIN deploy_configs ON deploy_configs.project_id = projects.id
		 WHERE projects.id = ?`,
		id,
	)
	return s.scanProjectWithGitAuth(row)
}

func (s *Store) GetProjectByWebhookToken(ctx context.Context, token string) (model.Project, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT projects.id, projects.sort_order, projects.name, projects.repo_url, projects.branch, projects.description, projects.webhook_token,
		        (CASE WHEN deploy_configs.id IS NOT NULL THEN 1 ELSE 0 END) as has_deploy_config,
		        projects.git_auth_type, projects.git_username_cipher, projects.git_password_cipher, projects.git_ssh_key_cipher,
		        projects.created_at, projects.updated_at
		 FROM projects
		 LEFT JOIN deploy_configs ON deploy_configs.project_id = projects.id
		 WHERE projects.webhook_token = ?`,
		token,
	)
	return s.scanProjectWithGitAuth(row)
}

func (s *Store) ListProjects(ctx context.Context) ([]model.Project, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT projects.id, projects.sort_order, projects.name, projects.repo_url, projects.branch, projects.description, projects.webhook_token,
		        (CASE WHEN deploy_configs.id IS NOT NULL THEN 1 ELSE 0 END) as has_deploy_config,
		        projects.git_auth_type, projects.git_username_cipher, projects.git_password_cipher, projects.git_ssh_key_cipher,
		        projects.created_at, projects.updated_at
		 FROM projects
		 LEFT JOIN deploy_configs ON deploy_configs.project_id = projects.id
		 ORDER BY projects.sort_order DESC, projects.id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		project, err := s.scanProjectWithGitAuth(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (s *Store) CloneProject(ctx context.Context, sourceID int64, input model.ProjectCloneInput) (model.ProjectDetail, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.ProjectDetail{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	sourceProjectRow := tx.QueryRowContext(
		ctx,
		`SELECT projects.id, projects.sort_order, projects.name, projects.repo_url, projects.branch, projects.description, projects.webhook_token,
		        (CASE WHEN deploy_configs.id IS NOT NULL THEN 1 ELSE 0 END) as has_deploy_config,
		        projects.git_auth_type, projects.git_username_cipher, projects.git_password_cipher, projects.git_ssh_key_cipher,
		        projects.created_at, projects.updated_at
		 FROM projects
		 LEFT JOIN deploy_configs ON deploy_configs.project_id = projects.id
		 WHERE projects.id = ?`,
		sourceID,
	)

	sourceProject, err := s.scanProjectWithGitAuth(sourceProjectRow)
	if err != nil {
		return model.ProjectDetail{}, err
	}

	token, err := randomToken()
	if err != nil {
		return model.ProjectDetail{}, fmt.Errorf("generate webhook token: %w", err)
	}

	description := input.Description
	if description == "" {
		description = sourceProject.Description
	}

	// 加密Git认证信息（复制项目时保留Git认证）
	gitUsernameCipher, err := s.cipher.Encrypt(sourceProject.GitUsername)
	if err != nil {
		return model.ProjectDetail{}, fmt.Errorf("encrypt git username: %w", err)
	}

	gitPasswordCipher, err := s.cipher.Encrypt(sourceProject.GitPassword)
	if err != nil {
		return model.ProjectDetail{}, fmt.Errorf("encrypt git password: %w", err)
	}

	gitSSHKeyCipher, err := s.cipher.Encrypt(sourceProject.GitSSHKey)
	if err != nil {
		return model.ProjectDetail{}, fmt.Errorf("encrypt git ssh key: %w", err)
	}

	now := nowString()
	nextSortOrder, err := s.nextSortOrder(ctx, tx, "projects")
	if err != nil {
		return model.ProjectDetail{}, err
	}
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO projects (sort_order, name, repo_url, branch, description, webhook_token, git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		nextSortOrder, input.Name, sourceProject.RepoURL, input.Branch, description, token, sourceProject.GitAuthType, gitUsernameCipher, gitPasswordCipher, gitSSHKeyCipher, now, now,
	)
	if err != nil {
		return model.ProjectDetail{}, wrapProjectMutationError("clone", err)
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		return model.ProjectDetail{}, fmt.Errorf("get cloned project id: %w", err)
	}

	config, err := s.getDeployConfigWithExecutor(ctx, tx, sourceID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return model.ProjectDetail{}, err
	}

	if err == nil {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO deploy_configs (
				project_id, host_id, build_image, build_commands_json, artifact_filter_mode,
				artifact_rules_json, remote_save_dir, remote_deploy_dir, pre_deploy_commands_json,
				post_deploy_commands_json, timeout_seconds, notify_webhook_url, notify_token_cipher, notification_channel_id, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			projectID,
			config.HostID,
			config.BuildImage,
			mustMarshal(config.BuildCommands),
			config.ArtifactFilterMode,
			mustMarshal(config.ArtifactRules),
			config.RemoteSaveDir,
			config.RemoteDeployDir,
			mustMarshal(config.PreDeployCommands),
			mustMarshal(config.PostDeployCommands),
			config.TimeoutSeconds,
			config.NotifyWebhookURL,
			mustEncryptString(s.cipher, config.NotifyBearerToken),
			config.NotificationChannelID,
			now,
			now,
		)
		if err != nil {
			return model.ProjectDetail{}, fmt.Errorf("clone deploy config: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return model.ProjectDetail{}, fmt.Errorf("commit clone transaction: %w", err)
	}

	return s.GetProjectDetail(ctx, projectID)
}

func (s *Store) UpsertDeployConfig(ctx context.Context, projectID int64, input model.DeployConfigUpsert) (model.DeployConfig, error) {
	if _, err := s.GetProject(ctx, projectID); err != nil {
		return model.DeployConfig{}, err
	}
	if _, err := s.GetHost(ctx, input.HostID); err != nil {
		return model.DeployConfig{}, err
	}

	tokenCipher := ""
	if input.NotifyBearerToken != nil {
		var err error
		tokenCipher, err = s.cipher.Encrypt(*input.NotifyBearerToken)
		if err != nil {
			return model.DeployConfig{}, fmt.Errorf("encrypt notify token: %w", err)
		}
	}

	currentConfig, err := s.GetDeployConfigByProjectID(ctx, projectID)
	if err == nil && input.NotifyBearerToken == nil {
		tokenCipher, err = s.cipher.Encrypt(currentConfig.NotifyBearerToken)
		if err != nil {
			return model.DeployConfig{}, fmt.Errorf("encrypt existing notify token: %w", err)
		}
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return model.DeployConfig{}, err
	}

	now := nowString()
	query := deployConfigUpsertQuery(s.isMySQL())
	_, err = s.db.ExecContext(
		ctx,
		query,
		projectID,
		input.HostID,
		input.BuildImage,
		mustMarshal(input.BuildCommands),
		input.ArtifactFilterMode,
		mustMarshal(input.ArtifactRules),
		input.RemoteSaveDir,
		input.RemoteDeployDir,
		mustMarshal(input.PreDeployCommands),
		mustMarshal(input.PostDeployCommands),
		input.TimeoutSeconds,
		input.NotifyWebhookURL,
		tokenCipher,
		input.NotificationChannelID,
		now,
		now,
	)
	if err != nil {
		return model.DeployConfig{}, fmt.Errorf("upsert deploy config: %w", err)
	}

	return s.GetDeployConfigByProjectID(ctx, projectID)
}

func deployConfigUpsertQuery(isMySQL bool) string {
	query := `INSERT INTO deploy_configs (
		project_id, host_id, build_image, build_commands_json, artifact_filter_mode,
		artifact_rules_json, remote_save_dir, remote_deploy_dir, pre_deploy_commands_json,
		post_deploy_commands_json, timeout_seconds, notify_webhook_url, notify_token_cipher, notification_channel_id, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(project_id) DO UPDATE SET
		host_id = excluded.host_id,
		build_image = excluded.build_image,
		build_commands_json = excluded.build_commands_json,
		artifact_filter_mode = excluded.artifact_filter_mode,
		artifact_rules_json = excluded.artifact_rules_json,
		remote_save_dir = excluded.remote_save_dir,
		remote_deploy_dir = excluded.remote_deploy_dir,
		pre_deploy_commands_json = excluded.pre_deploy_commands_json,
		post_deploy_commands_json = excluded.post_deploy_commands_json,
		timeout_seconds = excluded.timeout_seconds,
		notify_webhook_url = excluded.notify_webhook_url,
		notify_token_cipher = excluded.notify_token_cipher,
		notification_channel_id = excluded.notification_channel_id,
		updated_at = excluded.updated_at`
	if isMySQL {
		query = `INSERT INTO deploy_configs (
			project_id, host_id, build_image, build_commands_json, artifact_filter_mode,
			artifact_rules_json, remote_save_dir, remote_deploy_dir, pre_deploy_commands_json,
			post_deploy_commands_json, timeout_seconds, notify_webhook_url, notify_token_cipher, notification_channel_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			host_id = VALUES(host_id),
			build_image = VALUES(build_image),
			build_commands_json = VALUES(build_commands_json),
			artifact_filter_mode = VALUES(artifact_filter_mode),
			artifact_rules_json = VALUES(artifact_rules_json),
			remote_save_dir = VALUES(remote_save_dir),
			remote_deploy_dir = VALUES(remote_deploy_dir),
			pre_deploy_commands_json = VALUES(pre_deploy_commands_json),
			post_deploy_commands_json = VALUES(post_deploy_commands_json),
			timeout_seconds = VALUES(timeout_seconds),
			notify_webhook_url = VALUES(notify_webhook_url),
			notify_token_cipher = VALUES(notify_token_cipher),
			notification_channel_id = VALUES(notification_channel_id),
			updated_at = VALUES(updated_at)`
	}
	return query
}

func (s *Store) GetDeployConfigByProjectID(ctx context.Context, projectID int64) (model.DeployConfig, error) {
	return s.getDeployConfigWithExecutor(ctx, s.db, projectID)
}

func (s *Store) getDeployConfigWithExecutor(ctx context.Context, queryer queryRowContext, projectID int64) (model.DeployConfig, error) {
	row := queryer.QueryRowContext(
		ctx,
		`SELECT id, project_id, host_id, build_image, build_commands_json, artifact_filter_mode,
		        artifact_rules_json, remote_save_dir, remote_deploy_dir, pre_deploy_commands_json,
		        post_deploy_commands_json, timeout_seconds, notify_webhook_url, notify_token_cipher, notification_channel_id, created_at, updated_at
		 FROM deploy_configs
		 WHERE project_id = ?`,
		projectID,
	)
	return s.scanDeployConfig(row)
}

func (s *Store) GetProjectDetail(ctx context.Context, projectID int64) (model.ProjectDetail, error) {
	project, err := s.GetProject(ctx, projectID)
	if err != nil {
		return model.ProjectDetail{}, err
	}

	detail := model.ProjectDetail{Project: project}

	config, err := s.GetDeployConfigByProjectID(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return detail, nil
		}
		return model.ProjectDetail{}, err
	}
	detail.DeployConfig = &config

	host, err := s.GetHost(ctx, config.HostID)
	if err != nil {
		return model.ProjectDetail{}, err
	}
	detail.Host = &host

	return detail, nil
}

func (s *Store) GetExecutionBundle(ctx context.Context, projectID int64) (model.ExecutionBundle, error) {
	project, err := s.GetProject(ctx, projectID)
	if err != nil {
		return model.ExecutionBundle{}, err
	}

	config, err := s.GetDeployConfigByProjectID(ctx, projectID)
	if err != nil {
		return model.ExecutionBundle{}, err
	}

	host, err := s.GetHost(ctx, config.HostID)
	if err != nil {
		return model.ExecutionBundle{}, err
	}

	return model.ExecutionBundle{
		Project:      project,
		DeployConfig: config,
		Host:         host,
	}, nil
}

func (s *Store) CreateRun(ctx context.Context, input model.RunCreateInput) (model.PipelineRun, error) {
	now := nowString()
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO pipeline_runs (project_id, status, trigger_type, trigger_ref, log_text, error_message, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.ProjectID, input.Status, input.TriggerType, input.TriggerRef, "", "", now, now,
	)
	if err != nil {
		return model.PipelineRun{}, fmt.Errorf("insert run: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.PipelineRun{}, fmt.Errorf("get run id: %w", err)
	}

	return s.GetRun(ctx, id)
}

func (s *Store) MarkRunRunning(ctx context.Context, runID int64) error {
	now := nowString()
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE pipeline_runs
		 SET status = ?, started_at = ?, updated_at = ?
		 WHERE id = ?`,
		model.RunStatusRunning, now, now, runID,
	)
	if err != nil {
		return fmt.Errorf("mark run running: %w", err)
	}
	return nil
}

func (s *Store) AppendRunLog(ctx context.Context, runID int64, line string) error {
	query := appendRunLogQuery(s.isMySQL())
	_, err := s.db.ExecContext(
		ctx,
		query,
		line,
		nowString(),
		runID,
	)
	if err != nil {
		return fmt.Errorf("append run log: %w", err)
	}
	return nil
}

func appendRunLogQuery(isMySQL bool) string {
	if isMySQL {
		return `UPDATE pipeline_runs
		 SET log_text = CONCAT(COALESCE(log_text, ''), ?), updated_at = ?
		 WHERE id = ?`
	}
	return `UPDATE pipeline_runs
	 SET log_text = COALESCE(log_text, '') || ?, updated_at = ?
	 WHERE id = ?`
}

func (s *Store) FinalizeRun(ctx context.Context, runID int64, status string, errorMessage string) error {
	now := nowString()
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE pipeline_runs
		 SET status = ?, error_message = ?, finished_at = ?, updated_at = ?
		 WHERE id = ?`,
		status, errorMessage, now, now, runID,
	)
	if err != nil {
		return fmt.Errorf("finalize run: %w", err)
	}
	return nil
}

func (s *Store) GetRun(ctx context.Context, runID int64) (model.PipelineRun, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT pipeline_runs.id, pipeline_runs.project_id, projects.name, projects.branch,
		        pipeline_runs.status, pipeline_runs.trigger_type, pipeline_runs.trigger_ref,
		        pipeline_runs.log_text, pipeline_runs.error_message,
		        pipeline_runs.started_at, pipeline_runs.finished_at, pipeline_runs.created_at, pipeline_runs.updated_at
		 FROM pipeline_runs
		 JOIN projects ON projects.id = pipeline_runs.project_id
		 WHERE pipeline_runs.id = ?`,
		runID,
	)
	return scanRun(row)
}

func (s *Store) ListRunsByProject(ctx context.Context, projectID int64) ([]model.PipelineRun, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT pipeline_runs.id, pipeline_runs.project_id, projects.name, projects.branch,
		        pipeline_runs.status, pipeline_runs.trigger_type, pipeline_runs.trigger_ref,
		        pipeline_runs.log_text, pipeline_runs.error_message,
		        pipeline_runs.started_at, pipeline_runs.finished_at, pipeline_runs.created_at, pipeline_runs.updated_at
		 FROM pipeline_runs
		 JOIN projects ON projects.id = pipeline_runs.project_id
		 WHERE pipeline_runs.project_id = ?
		 ORDER BY pipeline_runs.id DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("query runs: %w", err)
	}
	defer rows.Close()

	var runs []model.PipelineRun
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

func (s *Store) ListAllRuns(ctx context.Context, offset, limit int) ([]model.PipelineRun, error) {
	query := `SELECT pipeline_runs.id, pipeline_runs.project_id, projects.name, projects.branch,
		        pipeline_runs.status, pipeline_runs.trigger_type, pipeline_runs.trigger_ref,
		        pipeline_runs.log_text, pipeline_runs.error_message,
		        pipeline_runs.started_at, pipeline_runs.finished_at, pipeline_runs.created_at, pipeline_runs.updated_at
		 FROM pipeline_runs
		 JOIN projects ON projects.id = pipeline_runs.project_id
		 ORDER BY pipeline_runs.id DESC
		 LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query all runs: %w", err)
	}
	defer rows.Close()

	var runs []model.PipelineRun
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

func (s *Store) GetHomeDashboard(ctx context.Context) (model.HomeDashboard, error) {
	projects, err := s.ListProjects(ctx)
	if err != nil {
		return model.HomeDashboard{}, err
	}

	runs, err := s.listRunsForDashboard(ctx)
	if err != nil {
		return model.HomeDashboard{}, err
	}

	now := time.Now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dailyStart := today.AddDate(0, 0, -(54*7 - 1))

	dailyMap := make(map[string]*model.HomeStatsDaily, 54*7)
	dailyStats := make([]model.HomeStatsDaily, 0, 54*7)
	for index := 0; index < 54*7; index++ {
		current := dailyStart.AddDate(0, 0, index)
		key := current.Format("2006-01-02")
		entry := model.HomeStatsDaily{Date: key}
		dailyStats = append(dailyStats, entry)
		dailyMap[key] = &dailyStats[len(dailyStats)-1]
	}

	hourlyStats := make([]model.HomeStatsHourly, 24)
	for hour := 0; hour < 24; hour++ {
		hourlyStats[hour] = model.HomeStatsHourly{Hour: fmt.Sprintf("%02d", hour)}
	}

	projectStatsMap := make(map[int64]*model.HomeProjectRank, len(projects))
	for _, project := range projects {
		projectCopy := model.HomeProjectRank{
			ProjectID:   project.ID,
			ProjectName: project.Name,
			Branch:      project.Branch,
		}
		projectStatsMap[project.ID] = &projectCopy
	}

	var total model.HomeStatsTotal
	total.ProjectCount = int64(len(projects))
	var totalDurationSeconds int64
	var finishedRunCount int64

	for _, run := range runs {
		total.DeployCount += 1
		runTime := dashboardRunTime(run)
		if total.LastDeployAt == "" || runTime.After(parseDashboardTime(total.LastDeployAt)) {
			total.LastDeployAt = runTime.Format(time.RFC3339)
		}

		projectStat, exists := projectStatsMap[run.ProjectID]
		if !exists {
			projectStat = &model.HomeProjectRank{
				ProjectID:   run.ProjectID,
				ProjectName: run.ProjectName,
				Branch:      run.Branch,
			}
			projectStatsMap[run.ProjectID] = projectStat
		}
		projectStat.DeployCount += 1
		if projectStat.LastDeployAt == "" || runTime.After(parseDashboardTime(projectStat.LastDeployAt)) {
			projectStat.LastDeployAt = runTime.Format(time.RFC3339)
		}

		if dayEntry := dailyMap[runTime.Format("2006-01-02")]; dayEntry != nil {
			dayEntry.DeployCount += 1
			incrementDashboardStatus(dayEntry, run.Status)
		}

		if runTime.Year() == today.Year() && runTime.YearDay() == today.YearDay() {
			hourEntry := &hourlyStats[runTime.Hour()]
			hourEntry.DeployCount += 1
			incrementDashboardStatus(hourEntry, run.Status)
		}

		incrementDashboardStatus(&total, run.Status)
		incrementProjectStatus(projectStat, run.Status)

		if run.StartedAt != nil && run.FinishedAt != nil {
			durationSeconds := int64(run.FinishedAt.Sub(*run.StartedAt).Seconds())
			if durationSeconds >= 0 {
				totalDurationSeconds += durationSeconds
				finishedRunCount += 1
			}
		}
	}

	for index := range dailyStats {
		dailyStats[index].SuccessRate = dashboardSuccessRate(dailyStats[index].SuccessCount, dailyStats[index].FailedCount)
	}
	for index := range hourlyStats {
		hourlyStats[index].SuccessRate = dashboardSuccessRate(hourlyStats[index].SuccessCount, hourlyStats[index].FailedCount)
	}

	projectStats := make([]model.HomeProjectRank, 0, len(projectStatsMap))
	for _, projectStat := range projectStatsMap {
		projectStat.SuccessRate = dashboardSuccessRate(projectStat.SuccessCount, projectStat.FailedCount)
		projectStats = append(projectStats, *projectStat)
	}
	sort.Slice(projectStats, func(left, right int) bool {
		if projectStats[left].DeployCount != projectStats[right].DeployCount {
			return projectStats[left].DeployCount > projectStats[right].DeployCount
		}
		if projectStats[left].SuccessRate != projectStats[right].SuccessRate {
			return projectStats[left].SuccessRate > projectStats[right].SuccessRate
		}
		return parseDashboardTime(projectStats[left].LastDeployAt).After(parseDashboardTime(projectStats[right].LastDeployAt))
	})

	total.SuccessRate = dashboardSuccessRate(total.SuccessCount, total.FailedCount)
	if total.ProjectCount > 0 {
		total.AverageDeployPerProj = float64(total.DeployCount) / float64(total.ProjectCount)
	}
	if finishedRunCount > 0 {
		total.AverageDurationSec = totalDurationSeconds / finishedRunCount
	}

	return model.HomeDashboard{
		Total:    total,
		Daily:    dailyStats,
		Hourly:   hourlyStats,
		Projects: projectStats,
	}, nil
}

func (s *Store) listRunsForDashboard(ctx context.Context) ([]model.PipelineRun, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT pipeline_runs.id, pipeline_runs.project_id, projects.name, projects.branch,
		        pipeline_runs.status, pipeline_runs.trigger_type, pipeline_runs.trigger_ref,
		        pipeline_runs.log_text, pipeline_runs.error_message,
		        pipeline_runs.started_at, pipeline_runs.finished_at, pipeline_runs.created_at, pipeline_runs.updated_at
		 FROM pipeline_runs
		 JOIN projects ON projects.id = pipeline_runs.project_id
		 ORDER BY pipeline_runs.id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query dashboard runs: %w", err)
	}
	defer rows.Close()

	var runs []model.PipelineRun
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

type queryRowContext interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *Store) nextSortOrder(ctx context.Context, queryer queryRowContext, table string) (int64, error) {
	row := queryer.QueryRowContext(ctx, fmt.Sprintf(`SELECT COALESCE(MAX(sort_order), 0) + 1 FROM %s`, table))

	var nextSortOrder int64
	if err := row.Scan(&nextSortOrder); err != nil {
		return 0, fmt.Errorf("query next sort order for %s: %w", table, err)
	}

	return nextSortOrder, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func (s *Store) scanHost(scan scanner) (model.Host, error) {
	var (
		host            model.Host
		passwordCipher  string
		createdAtString string
		updatedAtString string
	)

	err := scan.Scan(
		&host.ID,
		&host.SortOrder,
		&host.Name,
		&host.Address,
		&host.Port,
		&host.Username,
		&passwordCipher,
		&createdAtString,
		&updatedAtString,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Host{}, ErrNotFound
	}
	if err != nil {
		return model.Host{}, fmt.Errorf("scan host: %w", err)
	}

	host.Password, err = s.cipher.Decrypt(passwordCipher)
	if err != nil {
		return model.Host{}, fmt.Errorf("decrypt host password: %w", err)
	}
	host.HasPassword = host.Password != ""
	host.CreatedAt, err = parseTime(createdAtString)
	if err != nil {
		return model.Host{}, err
	}
	host.UpdatedAt, err = parseTime(updatedAtString)
	if err != nil {
		return model.Host{}, err
	}

	return host, nil
}

func scanProject(scan scanner) (model.Project, error) {
	var (
		project         model.Project
		hasDeployConfig int64
		gitAuthType     string
		createdAtString string
		updatedAtString string
	)

	err := scan.Scan(
		&project.ID,
		&project.SortOrder,
		&project.Name,
		&project.RepoURL,
		&project.Branch,
		&project.Description,
		&project.WebhookToken,
		&hasDeployConfig,
		&gitAuthType,
		&createdAtString,
		&updatedAtString,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Project{}, ErrNotFound
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("scan project: %w", err)
	}

	project.HasDeployConfig = hasDeployConfig == 1
	project.GitAuthType = gitAuthType

	// 判断是否有Git认证信息（不实际解密，仅标记）
	project.HasGitAuth = project.GitAuthType != model.GitAuthTypeNone

	project.CreatedAt, err = parseTime(createdAtString)
	if err != nil {
		return model.Project{}, err
	}
	project.UpdatedAt, err = parseTime(updatedAtString)
	if err != nil {
		return model.Project{}, err
	}

	return project, nil
}

func (s *Store) scanProjectWithGitAuth(scan scanner) (model.Project, error) {
	var (
		project           model.Project
		hasDeployConfig   int64
		gitAuthType       string
		gitUsernameCipher string
		gitPasswordCipher string
		gitSSHKeyCipher   string
		createdAtString   string
		updatedAtString   string
	)

	err := scan.Scan(
		&project.ID,
		&project.SortOrder,
		&project.Name,
		&project.RepoURL,
		&project.Branch,
		&project.Description,
		&project.WebhookToken,
		&hasDeployConfig,
		&gitAuthType,
		&gitUsernameCipher,
		&gitPasswordCipher,
		&gitSSHKeyCipher,
		&createdAtString,
		&updatedAtString,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Project{}, ErrNotFound
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("scan project: %w", err)
	}

	project.HasDeployConfig = hasDeployConfig == 1
	project.GitAuthType = gitAuthType

	// 解密Git认证信息
	if gitUsernameCipher != "" {
		project.GitUsername, err = s.cipher.Decrypt(gitUsernameCipher)
		if err != nil {
			return model.Project{}, fmt.Errorf("decrypt git username: %w", err)
		}
	}

	if gitPasswordCipher != "" {
		project.GitPassword, err = s.cipher.Decrypt(gitPasswordCipher)
		if err != nil {
			return model.Project{}, fmt.Errorf("decrypt git password: %w", err)
		}
	}

	if gitSSHKeyCipher != "" {
		project.GitSSHKey, err = s.cipher.Decrypt(gitSSHKeyCipher)
		if err != nil {
			return model.Project{}, fmt.Errorf("decrypt git ssh key: %w", err)
		}
	}

	// 判断是否有Git认证信息
	project.HasGitAuth = project.GitAuthType != model.GitAuthTypeNone &&
		(project.GitUsername != "" || project.GitPassword != "" || project.GitSSHKey != "")

	project.CreatedAt, err = parseTime(createdAtString)
	if err != nil {
		return model.Project{}, err
	}
	project.UpdatedAt, err = parseTime(updatedAtString)
	if err != nil {
		return model.Project{}, err
	}

	return project, nil
}

func (s *Store) scanDeployConfig(scan scanner) (model.DeployConfig, error) {
	var (
		config                model.DeployConfig
		buildCommandsJSON     string
		artifactRulesJSON     string
		preDeployCommands     string
		postDeployCommands    string
		timeoutSeconds        sql.NullInt64
		notifyTokenCipher     string
		notificationChannelID sql.NullInt64
		createdAtString       string
		updatedAtString       string
	)

	err := scan.Scan(
		&config.ID,
		&config.ProjectID,
		&config.HostID,
		&config.BuildImage,
		&buildCommandsJSON,
		&config.ArtifactFilterMode,
		&artifactRulesJSON,
		&config.RemoteSaveDir,
		&config.RemoteDeployDir,
		&preDeployCommands,
		&postDeployCommands,
		&timeoutSeconds,
		&config.NotifyWebhookURL,
		&notifyTokenCipher,
		&notificationChannelID,
		&createdAtString,
		&updatedAtString,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.DeployConfig{}, ErrNotFound
	}
	if err != nil {
		return model.DeployConfig{}, fmt.Errorf("scan deploy config: %w", err)
	}

	if timeoutSeconds.Valid {
		config.TimeoutSeconds = int(timeoutSeconds.Int64)
	} else {
		config.TimeoutSeconds = 0 // 默认值
	}

	if notificationChannelID.Valid {
		id := notificationChannelID.Int64
		config.NotificationChannelID = &id
	}

	if err = json.Unmarshal([]byte(buildCommandsJSON), &config.BuildCommands); err != nil {
		return model.DeployConfig{}, fmt.Errorf("unmarshal build commands: %w", err)
	}
	if err = json.Unmarshal([]byte(artifactRulesJSON), &config.ArtifactRules); err != nil {
		return model.DeployConfig{}, fmt.Errorf("unmarshal artifact rules: %w", err)
	}
	if err = json.Unmarshal([]byte(preDeployCommands), &config.PreDeployCommands); err != nil {
		return model.DeployConfig{}, fmt.Errorf("unmarshal pre-deploy commands: %w", err)
	}
	if err = json.Unmarshal([]byte(postDeployCommands), &config.PostDeployCommands); err != nil {
		return model.DeployConfig{}, fmt.Errorf("unmarshal post-deploy commands: %w", err)
	}

	config.NotifyBearerToken, err = s.cipher.Decrypt(notifyTokenCipher)
	if err != nil {
		return model.DeployConfig{}, fmt.Errorf("decrypt notify token: %w", err)
	}
	config.HasNotifyToken = config.NotifyBearerToken != ""

	config.CreatedAt, err = parseTime(createdAtString)
	if err != nil {
		return model.DeployConfig{}, err
	}
	config.UpdatedAt, err = parseTime(updatedAtString)
	if err != nil {
		return model.DeployConfig{}, err
	}

	return config, nil
}

func scanRun(scan scanner) (model.PipelineRun, error) {
	var (
		run              model.PipelineRun
		startedAtString  sql.NullString
		finishedAtString sql.NullString
		createdAtString  string
		updatedAtString  string
	)

	err := scan.Scan(
		&run.ID,
		&run.ProjectID,
		&run.ProjectName,
		&run.Branch,
		&run.Status,
		&run.TriggerType,
		&run.TriggerRef,
		&run.LogText,
		&run.ErrorMessage,
		&startedAtString,
		&finishedAtString,
		&createdAtString,
		&updatedAtString,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.PipelineRun{}, ErrNotFound
	}
	if err != nil {
		return model.PipelineRun{}, fmt.Errorf("scan run: %w", err)
	}

	createdAt, err := parseTime(createdAtString)
	if err != nil {
		return model.PipelineRun{}, err
	}
	updatedAt, err := parseTime(updatedAtString)
	if err != nil {
		return model.PipelineRun{}, err
	}
	run.CreatedAt = createdAt
	run.UpdatedAt = updatedAt

	if startedAtString.Valid {
		startedAt, err := parseTime(startedAtString.String)
		if err != nil {
			return model.PipelineRun{}, err
		}
		run.StartedAt = &startedAt
	}
	if finishedAtString.Valid {
		finishedAt, err := parseTime(finishedAtString.String)
		if err != nil {
			return model.PipelineRun{}, err
		}
		run.FinishedAt = &finishedAt
	}

	return run, nil
}

func incrementDashboardStatus(target interface{}, status string) {
	switch stats := target.(type) {
	case *model.HomeStatsTotal:
		switch status {
		case model.RunStatusSuccess:
			stats.SuccessCount += 1
		case model.RunStatusFailed:
			stats.FailedCount += 1
		case model.RunStatusRunning:
			stats.RunningCount += 1
		case model.RunStatusQueued:
			stats.QueuedCount += 1
		}
	case *model.HomeStatsDaily:
		switch status {
		case model.RunStatusSuccess:
			stats.SuccessCount += 1
		case model.RunStatusFailed:
			stats.FailedCount += 1
		case model.RunStatusRunning:
			stats.RunningCount += 1
		case model.RunStatusQueued:
			stats.QueuedCount += 1
		}
	case *model.HomeStatsHourly:
		switch status {
		case model.RunStatusSuccess:
			stats.SuccessCount += 1
		case model.RunStatusFailed:
			stats.FailedCount += 1
		case model.RunStatusRunning:
			stats.RunningCount += 1
		case model.RunStatusQueued:
			stats.QueuedCount += 1
		}
	}
}

func incrementProjectStatus(target *model.HomeProjectRank, status string) {
	switch status {
	case model.RunStatusSuccess:
		target.SuccessCount += 1
	case model.RunStatusFailed:
		target.FailedCount += 1
	case model.RunStatusRunning:
		target.RunningCount += 1
	case model.RunStatusQueued:
		target.QueuedCount += 1
	}
}

func dashboardSuccessRate(successCount, failedCount int64) float64 {
	total := successCount + failedCount
	if total == 0 {
		return 0
	}
	return float64(successCount) * 100 / float64(total)
}

func dashboardRunTime(run model.PipelineRun) time.Time {
	if run.StartedAt != nil {
		return run.StartedAt.In(time.Local)
	}
	return run.CreatedAt.In(time.Local)
}

func parseDashboardTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", value, err)
	}
	return parsed, nil
}

func nowString() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func expectDeleted(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func mustMarshal(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	data, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func mustEncryptString(cipher *cryptoutil.Cipher, value string) string {
	encrypted, err := cipher.Encrypt(value)
	if err != nil {
		panic(err)
	}
	return encrypted
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// 通知渠道管理

func (s *Store) CreateNotificationChannel(ctx context.Context, input model.NotificationChannelUpsert) (model.NotificationChannel, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("marshal channel config: %w", err)
	}

	now := nowString()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("begin create notification channel transaction: %w", err)
	}
	defer tx.Rollback()

	if input.IsDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE notification_channels SET is_default = 0 WHERE is_default = 1`); err != nil {
			return model.NotificationChannel{}, fmt.Errorf("reset default notification channels: %w", err)
		}
	}

	nextSortOrder, err := s.nextSortOrder(ctx, tx, "notification_channels")
	if err != nil {
		return model.NotificationChannel{}, err
	}
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO notification_channels (sort_order, name, type, is_default, remark, config_json, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		nextSortOrder, input.Name, input.Type, boolToInt(input.IsDefault), input.Remark, string(configJSON), now, now,
	)
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("insert notification channel: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("get channel id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.NotificationChannel{}, fmt.Errorf("commit create notification channel transaction: %w", err)
	}

	return s.GetNotificationChannel(ctx, id)
}

func (s *Store) UpdateNotificationChannel(ctx context.Context, id int64, input model.NotificationChannelUpsert) (model.NotificationChannel, error) {
	if _, err := s.GetNotificationChannel(ctx, id); err != nil {
		return model.NotificationChannel{}, err
	}

	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("marshal channel config: %w", err)
	}

	now := nowString()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("begin update notification channel transaction: %w", err)
	}
	defer tx.Rollback()

	if input.IsDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE notification_channels SET is_default = 0 WHERE id != ? AND is_default = 1`, id); err != nil {
			return model.NotificationChannel{}, fmt.Errorf("reset default notification channels: %w", err)
		}
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE notification_channels
		 SET name = ?, type = ?, is_default = ?, remark = ?, config_json = ?, updated_at = ?
		 WHERE id = ?`,
		input.Name, input.Type, boolToInt(input.IsDefault), input.Remark, string(configJSON), now, id,
	)
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("update notification channel: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.NotificationChannel{}, fmt.Errorf("commit update notification channel transaction: %w", err)
	}

	return s.GetNotificationChannel(ctx, id)
}

func (s *Store) DeleteNotificationChannel(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM notification_channels WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete notification channel: %w", err)
	}
	return expectDeleted(result)
}

func (s *Store) ReorderNotificationChannels(ctx context.Context, ids []int64) error {
	return s.reorderRecords(ctx, "notification_channels", ids)
}

func (s *Store) GetNotificationChannel(ctx context.Context, id int64) (model.NotificationChannel, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, sort_order, name, type, is_default, remark, config_json, created_at, updated_at
		 FROM notification_channels
		 WHERE id = ?`,
		id,
	)
	return s.scanNotificationChannel(row)
}

func (s *Store) GetNotificationChannelWithConfig(ctx context.Context, id int64) (model.NotificationChannelWithConfig, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, sort_order, name, type, is_default, remark, config_json, created_at, updated_at
		 FROM notification_channels
		 WHERE id = ?`,
		id,
	)
	return s.scanNotificationChannelWithConfig(row)
}

func (s *Store) ListNotificationChannels(ctx context.Context) ([]model.NotificationChannel, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, sort_order, name, type, is_default, remark, config_json, created_at, updated_at
		 FROM notification_channels
		 ORDER BY sort_order DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query notification channels: %w", err)
	}
	defer rows.Close()

	var channels []model.NotificationChannel
	for rows.Next() {
		channel, err := s.scanNotificationChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, rows.Err()
}

func (s *Store) GetDefaultNotificationChannel(ctx context.Context) (model.NotificationChannel, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, sort_order, name, type, is_default, remark, config_json, created_at, updated_at
		 FROM notification_channels
		 WHERE is_default = 1
		 ORDER BY sort_order DESC, id DESC
		 LIMIT 1`,
	)
	return s.scanNotificationChannel(row)
}

func (s *Store) SetDefaultNotificationChannel(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin set default notification channel transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE notification_channels SET is_default = 0 WHERE is_default = 1`); err != nil {
		return fmt.Errorf("reset default notification channels: %w", err)
	}
	result, err := tx.ExecContext(ctx, `UPDATE notification_channels SET is_default = 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("set default notification channel: %w", err)
	}
	if err := expectDeleted(result); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit set default notification channel transaction: %w", err)
	}
	return nil
}

func (s *Store) scanNotificationChannel(scan scanner) (model.NotificationChannel, error) {
	var (
		channel      model.NotificationChannel
		configJSON   string
		createdAtStr string
		updatedAtStr string
	)

	err := scan.Scan(
		&channel.ID,
		&channel.SortOrder,
		&channel.Name,
		&channel.Type,
		&channel.IsDefault,
		&channel.Remark,
		&configJSON,
		&createdAtStr,
		&updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.NotificationChannel{}, ErrNotFound
	}
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("scan notification channel: %w", err)
	}

	channel.Config = configJSON

	createdAt, err := parseTime(createdAtStr)
	if err != nil {
		return model.NotificationChannel{}, err
	}
	updatedAt, err := parseTime(updatedAtStr)
	if err != nil {
		return model.NotificationChannel{}, err
	}
	channel.CreatedAt = createdAt
	channel.UpdatedAt = updatedAt

	return channel, nil
}

func (s *Store) scanNotificationChannelWithConfig(scan scanner) (model.NotificationChannelWithConfig, error) {
	var (
		channel      model.NotificationChannelWithConfig
		configJSON   string
		createdAtStr string
		updatedAtStr string
	)

	err := scan.Scan(
		&channel.ID,
		&channel.SortOrder,
		&channel.Name,
		&channel.Type,
		&channel.IsDefault,
		&channel.Remark,
		&configJSON,
		&createdAtStr,
		&updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.NotificationChannelWithConfig{}, ErrNotFound
	}
	if err != nil {
		return model.NotificationChannelWithConfig{}, fmt.Errorf("scan notification channel: %w", err)
	}

	if err = json.Unmarshal([]byte(configJSON), &channel.ConfigMap); err != nil {
		return model.NotificationChannelWithConfig{}, fmt.Errorf("unmarshal channel config: %w", err)
	}

	createdAt, err := parseTime(createdAtStr)
	if err != nil {
		return model.NotificationChannelWithConfig{}, err
	}
	updatedAt, err := parseTime(updatedAtStr)
	if err != nil {
		return model.NotificationChannelWithConfig{}, err
	}
	channel.CreatedAt = createdAt
	channel.UpdatedAt = updatedAt

	return channel, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (s *Store) ensureSortOrderColumn(ctx context.Context, table string) error {
	if s.isMySQL() {
		var count int
		if err := s.db.QueryRowContext(
			ctx,
			`SELECT COUNT(*) FROM information_schema.COLUMNS
			 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = 'sort_order'`,
			table,
		).Scan(&count); err != nil {
			return fmt.Errorf("read %s columns: %w", table, err)
		}
		if count > 0 {
			return nil
		}
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s ADD COLUMN sort_order BIGINT NOT NULL DEFAULT 0`, table)); err != nil {
			return fmt.Errorf("add sort_order to %s: %w", table, err)
		}
		return nil
	}

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return fmt.Errorf("read %s columns: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid          int
			name         string
			columnType   string
			notNull      int
			defaultValue sql.NullString
			primaryKey   int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("scan %s columns: %w", table, err)
		}
		if name == "sort_order" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate %s columns: %w", table, err)
	}

	if _, err := s.db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0`, table)); err != nil {
		return fmt.Errorf("add sort_order to %s: %w", table, err)
	}

	return nil
}

func (s *Store) initializeSortOrder(ctx context.Context, table string) error {
	if _, err := s.db.ExecContext(ctx, fmt.Sprintf(`UPDATE %s SET sort_order = id WHERE sort_order = 0`, table)); err != nil {
		return fmt.Errorf("initialize %s sort_order: %w", table, err)
	}
	return nil
}

func (s *Store) reorderRecords(ctx context.Context, table string, ids []int64) error {
	if len(ids) == 0 {
		return errors.New("ids are required")
	}

	existingIDs, err := s.listRecordIDs(ctx, table)
	if err != nil {
		return err
	}
	if len(existingIDs) != len(ids) {
		return fmt.Errorf("ids length mismatch for %s", table)
	}

	expected := make(map[int64]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		expected[id] = struct{}{}
	}
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := expected[id]; !ok {
			return fmt.Errorf("invalid id %d for %s", id, table)
		}
		if _, duplicated := seen[id]; duplicated {
			return fmt.Errorf("duplicate id %d for %s", id, table)
		}
		seen[id] = struct{}{}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reorder %s transaction: %w", table, err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`UPDATE %s SET sort_order = ?, updated_at = ? WHERE id = ?`, table)
	now := nowString()
	total := len(ids)
	for index, id := range ids {
		sortOrder := total - index
		if _, err := tx.ExecContext(ctx, query, sortOrder, now, id); err != nil {
			return fmt.Errorf("reorder %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reorder %s: %w", table, err)
	}

	return nil
}

func (s *Store) listRecordIDs(ctx context.Context, table string) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`SELECT id FROM %s ORDER BY sort_order DESC, id DESC`, table))
	if err != nil {
		return nil, fmt.Errorf("query %s ids: %w", table, err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan %s id: %w", table, err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s ids: %w", table, err)
	}

	return ids, nil
}

// 管理员用户管理

func (s *Store) GetAdminUser(ctx context.Context) (model.AdminUser, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, created_at, updated_at
		 FROM admin_users
		 LIMIT 1`,
	)
	return s.scanAdminUser(row)
}

func (s *Store) CreateAdminUser(ctx context.Context, username, passwordHash string) (model.AdminUser, error) {
	now := nowString()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.AdminUser{}, fmt.Errorf("begin create admin user transaction: %w", err)
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return model.AdminUser{}, fmt.Errorf("count admin users: %w", err)
	}
	if count > 0 {
		return model.AdminUser{}, fmt.Errorf("only one admin user is allowed")
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO admin_users (username, password_hash, created_at, updated_at)
		 VALUES (?, ?, ?, ?)`,
		username, passwordHash, now, now,
	)
	if err != nil {
		return model.AdminUser{}, fmt.Errorf("create admin user: %w", err)
	}

	_, err = result.LastInsertId()
	if err != nil {
		return model.AdminUser{}, fmt.Errorf("get admin user id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.AdminUser{}, fmt.Errorf("commit create admin user transaction: %w", err)
	}

	return s.GetAdminUser(ctx)
}

func (s *Store) UpdateAdminPassword(ctx context.Context, passwordHash string) error {
	now := nowString()
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE admin_users SET password_hash = ?, updated_at = ?`,
		passwordHash, now,
	)
	if err != nil {
		return fmt.Errorf("update admin password: %w", err)
	}
	return nil
}

func (s *Store) AdminUserExists(ctx context.Context) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check admin user exists: %w", err)
	}
	return count > 0, nil
}

func (s *Store) scanAdminUser(scan scanner) (model.AdminUser, error) {
	var (
		user         model.AdminUser
		createdAtStr string
		updatedAtStr string
	)

	err := scan.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&createdAtStr,
		&updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.AdminUser{}, ErrNotFound
	}
	if err != nil {
		return model.AdminUser{}, fmt.Errorf("scan admin user: %w", err)
	}

	createdAt, err := parseTime(createdAtStr)
	if err != nil {
		return model.AdminUser{}, err
	}
	updatedAt, err := parseTime(updatedAtStr)
	if err != nil {
		return model.AdminUser{}, err
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return user, nil
}
