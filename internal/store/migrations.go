package store

import "fmt"

func (s *Store) migrationStatements() []string {
	if s.isMySQL() {
		return mysqlMigrationStatements()
	}
	return sqliteMigrationStatements()
}

func sqliteMigrationStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS hosts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sort_order INTEGER NOT NULL DEFAULT 0,
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
			sort_order INTEGER NOT NULL DEFAULT 0,
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
		`CREATE TABLE IF NOT EXISTS notification_channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sort_order INTEGER NOT NULL DEFAULT 0,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			is_default INTEGER NOT NULL DEFAULT 0,
			remark TEXT NOT NULL DEFAULT '',
			config_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
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
		`CREATE TABLE IF NOT EXISTS admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
	}
}

func mysqlMigrationStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS hosts (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			sort_order BIGINT NOT NULL DEFAULT 0,
			name VARCHAR(255) NOT NULL,
			address VARCHAR(255) NOT NULL,
			port INT NOT NULL,
			username VARCHAR(255) NOT NULL,
			password_cipher TEXT NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			updated_at VARCHAR(64) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
		`CREATE TABLE IF NOT EXISTS projects (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			sort_order BIGINT NOT NULL DEFAULT 0,
			name VARCHAR(255) NOT NULL,
			repo_url TEXT NOT NULL,
			branch VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			webhook_token VARCHAR(255) NOT NULL,
			git_auth_type VARCHAR(32) NOT NULL DEFAULT 'none',
			git_username_cipher TEXT NOT NULL,
			git_password_cipher TEXT NOT NULL,
			git_ssh_key_cipher LONGTEXT NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			updated_at VARCHAR(64) NOT NULL,
			UNIQUE KEY uniq_projects_repo_branch (repo_url(255), branch),
			UNIQUE KEY uniq_projects_webhook_token (webhook_token)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
		`CREATE TABLE IF NOT EXISTS notification_channels (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			sort_order BIGINT NOT NULL DEFAULT 0,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(64) NOT NULL,
			is_default TINYINT(1) NOT NULL DEFAULT 0,
			remark TEXT NOT NULL,
			config_json LONGTEXT NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			updated_at VARCHAR(64) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
		`CREATE TABLE IF NOT EXISTS deploy_configs (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			project_id BIGINT NOT NULL,
			host_id BIGINT NOT NULL,
			build_image VARCHAR(255) NOT NULL,
			build_commands_json LONGTEXT NOT NULL,
			artifact_filter_mode VARCHAR(32) NOT NULL,
			artifact_rules_json LONGTEXT NOT NULL,
			remote_save_dir TEXT NOT NULL,
			remote_deploy_dir TEXT NOT NULL,
			pre_deploy_commands_json LONGTEXT NOT NULL,
			post_deploy_commands_json LONGTEXT NOT NULL,
			timeout_seconds INT NOT NULL DEFAULT 1800,
			version_count INT NOT NULL DEFAULT 5,
			notify_webhook_url TEXT NOT NULL,
			notify_token_cipher TEXT NOT NULL,
			notification_channel_id BIGINT NULL,
			created_at VARCHAR(64) NOT NULL,
			updated_at VARCHAR(64) NOT NULL,
			UNIQUE KEY uniq_deploy_configs_project_id (project_id),
			CONSTRAINT fk_deploy_configs_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			CONSTRAINT fk_deploy_configs_host FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE RESTRICT,
			CONSTRAINT fk_deploy_configs_notification_channel FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
		`CREATE TABLE IF NOT EXISTS pipeline_runs (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			project_id BIGINT NOT NULL,
			status VARCHAR(32) NOT NULL,
			trigger_type VARCHAR(32) NOT NULL,
			trigger_ref TEXT NOT NULL,
			log_text LONGTEXT NOT NULL,
			error_message LONGTEXT NOT NULL,
			started_at VARCHAR(64) NULL,
			finished_at VARCHAR(64) NULL,
			created_at VARCHAR(64) NOT NULL,
			updated_at VARCHAR(64) NOT NULL,
			CONSTRAINT fk_pipeline_runs_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
		`CREATE TABLE IF NOT EXISTS admin_users (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(191) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			updated_at VARCHAR(64) NOT NULL,
			UNIQUE KEY uniq_admin_users_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
		`CREATE TABLE IF NOT EXISTS settings (
			` + "`key`" + ` VARCHAR(191) NOT NULL PRIMARY KEY,
			` + "`value`" + ` LONGTEXT NOT NULL,
			created_at VARCHAR(64) NOT NULL,
			updated_at VARCHAR(64) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
	}
}

func columnExistsQuery(driver, table, column string) (string, []any) {
	if driver == DriverMySQL {
		return `SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`, []any{table, column}
	}
	return fmt.Sprintf(`PRAGMA table_info(%s)`, table), nil
}
