package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cryptoutil "devops-pipeline/internal/crypto"
	"devops-pipeline/internal/model"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("store: not found")

type Store struct {
	db     *sql.DB
	cipher *cryptoutil.Cipher
}

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)

	if _, err = db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return db, nil
}

func New(db *sql.DB, cipher *cryptoutil.Cipher) *Store {
	return &Store{db: db, cipher: cipher}
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS hosts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			address TEXT NOT NULL,
			port INTEGER NOT NULL,
			username TEXT NOT NULL,
			password_cipher TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			repo_url TEXT NOT NULL,
			branch TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			webhook_token TEXT NOT NULL UNIQUE,
			git_auth_type TEXT NOT NULL DEFAULT 'none',
			git_username_cipher TEXT NOT NULL DEFAULT '',
			git_password_cipher TEXT NOT NULL DEFAULT '',
			git_ssh_key_cipher TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			UNIQUE(repo_url, branch)
		);`,
		`CREATE TABLE IF NOT EXISTS deploy_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL UNIQUE,
			host_id INTEGER NOT NULL,
			build_image TEXT NOT NULL,
			build_commands_json TEXT NOT NULL,
			artifact_filter_mode TEXT NOT NULL,
			artifact_rules_json TEXT NOT NULL,
			remote_save_dir TEXT NOT NULL,
			remote_deploy_dir TEXT NOT NULL,
			pre_deploy_commands_json TEXT NOT NULL,
			post_deploy_commands_json TEXT NOT NULL,
			timeout_seconds INTEGER NOT NULL DEFAULT 1800,
			version_count INTEGER NOT NULL DEFAULT 5,
			notify_webhook_url TEXT NOT NULL DEFAULT '',
			notify_token_cipher TEXT NOT NULL DEFAULT '',
			notification_channel_id INTEGER,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
			FOREIGN KEY(host_id) REFERENCES hosts(id) ON DELETE RESTRICT,
			FOREIGN KEY(notification_channel_id) REFERENCES notification_channels(id) ON DELETE SET NULL
		);`,
		`CREATE TABLE IF NOT EXISTS pipeline_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			trigger_type TEXT NOT NULL,
			trigger_ref TEXT NOT NULL DEFAULT '',
			log_text TEXT NOT NULL DEFAULT '',
			error_message TEXT NOT NULL DEFAULT '',
			started_at TEXT,
			finished_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS notification_channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			is_default INTEGER NOT NULL DEFAULT 0,
			remark TEXT NOT NULL DEFAULT '',
			config_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TRIGGER IF NOT EXISTS unique_default_channel
			AFTER UPDATE OF is_default ON notification_channels
			WHEN NEW.is_default = 1
		BEGIN
			UPDATE notification_channels SET is_default = 0 WHERE id != NEW.id AND is_default = 1;
		END;`,
		`CREATE TRIGGER IF NOT EXISTS insert_default_channel
			AFTER INSERT ON notification_channels
			WHEN NEW.is_default = 1
		BEGIN
			UPDATE notification_channels SET is_default = 0 WHERE id != NEW.id AND is_default = 1;
		END;`,
		`CREATE TABLE IF NOT EXISTS admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TRIGGER IF NOT EXISTS prevent_multiple_admins
			BEFORE INSERT ON admin_users
			WHEN (SELECT COUNT(*) FROM admin_users) >= 1
		BEGIN
			SELECT RAISE(ABORT, 'Only one admin user is allowed');
		END;`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}

	return nil
}

func (s *Store) CreateHost(ctx context.Context, input model.HostUpsert) (model.Host, error) {
	encryptedPassword, err := s.cipher.Encrypt(valueOrEmpty(input.Password))
	if err != nil {
		return model.Host{}, fmt.Errorf("encrypt host password: %w", err)
	}

	now := nowString()
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO hosts (name, address, port, username, password_cipher, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.Address, input.Port, input.Username, encryptedPassword, now, now,
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

func (s *Store) GetHost(ctx context.Context, id int64) (model.Host, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, address, port, username, password_cipher, created_at, updated_at
		 FROM hosts
		 WHERE id = ?`,
		id,
	)
	return s.scanHost(row)
}

func (s *Store) ListHosts(ctx context.Context) ([]model.Host, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, address, port, username, password_cipher, created_at, updated_at
		 FROM hosts
		 ORDER BY id DESC`,
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

	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO projects (name, repo_url, branch, description, webhook_token, git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.RepoURL, input.Branch, input.Description, token, gitAuthType, gitUsernameCipher, gitPasswordCipher, gitSSHKeyCipher, now, now,
	)
	if err != nil {
		return model.Project{}, fmt.Errorf("insert project: %w", err)
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
		return model.Project{}, fmt.Errorf("update project: %w", err)
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

func (s *Store) GetProject(ctx context.Context, id int64) (model.Project, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, repo_url, branch, description, webhook_token,
		        git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher,
		        created_at, updated_at
		 FROM projects
		 WHERE id = ?`,
		id,
	)
	return s.scanProjectWithGitAuth(row)
}

func (s *Store) GetProjectByWebhookToken(ctx context.Context, token string) (model.Project, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, repo_url, branch, description, webhook_token,
		        git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher,
		        created_at, updated_at
		 FROM projects
		 WHERE webhook_token = ?`,
		token,
	)
	return s.scanProjectWithGitAuth(row)
}

func (s *Store) ListProjects(ctx context.Context) ([]model.Project, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, repo_url, branch, description, webhook_token,
		        git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher,
		        created_at, updated_at
		 FROM projects
		 ORDER BY id DESC`,
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
		`SELECT id, name, repo_url, branch, description, webhook_token,
		        git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher,
		        created_at, updated_at
		 FROM projects
		 WHERE id = ?`,
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
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO projects (name, repo_url, branch, description, webhook_token, git_auth_type, git_username_cipher, git_password_cipher, git_ssh_key_cipher, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Name, sourceProject.RepoURL, input.Branch, description, token, sourceProject.GitAuthType, gitUsernameCipher, gitPasswordCipher, gitSSHKeyCipher, now, now,
	)
	if err != nil {
		return model.ProjectDetail{}, fmt.Errorf("clone project: %w", err)
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
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO deploy_configs (
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
			updated_at = excluded.updated_at`,
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
		`INSERT INTO pipeline_runs (project_id, status, trigger_type, trigger_ref, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.ProjectID, input.Status, input.TriggerType, input.TriggerRef, now, now,
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
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE pipeline_runs
		 SET log_text = COALESCE(log_text, '') || ?, updated_at = ?
		 WHERE id = ?`,
		line,
		nowString(),
		runID,
	)
	if err != nil {
		return fmt.Errorf("append run log: %w", err)
	}
	return nil
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
		`SELECT id, project_id, status, trigger_type, trigger_ref, log_text, error_message,
		        started_at, finished_at, created_at, updated_at
		 FROM pipeline_runs
		 WHERE id = ?`,
		runID,
	)
	return scanRun(row)
}

func (s *Store) ListRunsByProject(ctx context.Context, projectID int64) ([]model.PipelineRun, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, project_id, status, trigger_type, trigger_ref, log_text, error_message,
		        started_at, finished_at, created_at, updated_at
		 FROM pipeline_runs
		 WHERE project_id = ?
		 ORDER BY id DESC`,
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
	query := `SELECT id, project_id, status, trigger_type, trigger_ref, log_text, error_message,
		        started_at, finished_at, created_at, updated_at
		 FROM pipeline_runs
		 ORDER BY id DESC
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

type queryRowContext interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
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
		project              model.Project
		hasDeployConfig      int64
		gitAuthType          string
		createdAtString      string
		updatedAtString      string
	)

	err := scan.Scan(
		&project.ID,
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
		project              model.Project
		gitAuthType          string
		gitUsernameCipher    string
		gitPasswordCipher    string
		gitSSHKeyCipher      string
		createdAtString      string
		updatedAtString      string
	)

	err := scan.Scan(
		&project.ID,
		&project.Name,
		&project.RepoURL,
		&project.Branch,
		&project.Description,
		&project.WebhookToken,
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
		config                   model.DeployConfig
		buildCommandsJSON        string
		artifactRulesJSON        string
		preDeployCommands        string
		postDeployCommands       string
		timeoutSeconds           sql.NullInt64
		notifyTokenCipher        string
		notificationChannelID    sql.NullInt64
		createdAtString          string
		updatedAtString          string
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
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO notification_channels (name, type, is_default, remark, config_json, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.Type, boolToInt(input.IsDefault), input.Remark, string(configJSON), now, now,
	)
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("insert notification channel: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("get channel id: %w", err)
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
	_, err = s.db.ExecContext(
		ctx,
		`UPDATE notification_channels
		 SET name = ?, type = ?, is_default = ?, remark = ?, config_json = ?, updated_at = ?
		 WHERE id = ?`,
		input.Name, input.Type, boolToInt(input.IsDefault), input.Remark, string(configJSON), now, id,
	)
	if err != nil {
		return model.NotificationChannel{}, fmt.Errorf("update notification channel: %w", err)
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

func (s *Store) GetNotificationChannel(ctx context.Context, id int64) (model.NotificationChannel, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, type, is_default, remark, config_json, created_at, updated_at
		 FROM notification_channels
		 WHERE id = ?`,
		id,
	)
	return s.scanNotificationChannel(row)
}

func (s *Store) GetNotificationChannelWithConfig(ctx context.Context, id int64) (model.NotificationChannelWithConfig, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, name, type, is_default, remark, config_json, created_at, updated_at
		 FROM notification_channels
		 WHERE id = ?`,
		id,
	)
	return s.scanNotificationChannelWithConfig(row)
}

func (s *Store) ListNotificationChannels(ctx context.Context) ([]model.NotificationChannel, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, type, is_default, remark, config_json, created_at, updated_at
		 FROM notification_channels
		 ORDER BY is_default DESC, id DESC`,
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
		`SELECT id, name, type, is_default, remark, config_json, created_at, updated_at
		 FROM notification_channels
		 WHERE is_default = 1
		 LIMIT 1`,
	)
	return s.scanNotificationChannel(row)
}

func (s *Store) SetDefaultNotificationChannel(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE notification_channels SET is_default = 1 WHERE id = ?`,
		id,
	)
	if err != nil {
		return fmt.Errorf("set default notification channel: %w", err)
	}
	return nil
}

func (s *Store) scanNotificationChannel(scan scanner) (model.NotificationChannel, error) {
	var (
		channel        model.NotificationChannel
		configJSON     string
		createdAtStr   string
		updatedAtStr   string
	)

	err := scan.Scan(
		&channel.ID,
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
		channel        model.NotificationChannelWithConfig
		configJSON     string
		createdAtStr   string
		updatedAtStr   string
	)

	err := scan.Scan(
		&channel.ID,
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
	result, err := s.db.ExecContext(
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
		user          model.AdminUser
		createdAtStr  string
		updatedAtStr  string
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
