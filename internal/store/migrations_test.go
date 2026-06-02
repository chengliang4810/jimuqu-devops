package store

import (
	"strings"
	"testing"
)

func TestAISettingsUserAgentAddColumnSQL(t *testing.T) {
	mysqlSQL := aiSettingsUserAgentAddColumnSQL(true)
	if strings.Contains(strings.ToUpper(mysqlSQL), "DEFAULT") {
		t.Fatalf("expected mysql user_agent alter SQL to avoid DEFAULT for TEXT columns, got %q", mysqlSQL)
	}

	sqliteSQL := aiSettingsUserAgentAddColumnSQL(false)
	if !strings.Contains(strings.ToUpper(sqliteSQL), "DEFAULT ''") {
		t.Fatalf("expected sqlite user_agent alter SQL to include DEFAULT '', got %q", sqliteSQL)
	}
}

func TestDeployConfigMigrationsIncludeDeploySyncModeDefault(t *testing.T) {
	sqliteSQL := strings.Join(sqliteMigrationStatements(), "\n")
	if !strings.Contains(sqliteSQL, "deploy_sync_mode TEXT NOT NULL DEFAULT 'overwrite'") {
		t.Fatalf("expected sqlite deploy_configs to default deploy_sync_mode to overwrite, got %q", sqliteSQL)
	}

	mysqlSQL := strings.Join(mysqlMigrationStatements(), "\n")
	if !strings.Contains(mysqlSQL, "deploy_sync_mode VARCHAR(32) NOT NULL DEFAULT 'overwrite'") {
		t.Fatalf("expected mysql deploy_configs to default deploy_sync_mode to overwrite, got %q", mysqlSQL)
	}
}

func TestDeploySyncModeAddColumnSQL(t *testing.T) {
	sqliteSQL := deploySyncModeAddColumnSQL(false)
	if !strings.Contains(sqliteSQL, "DEFAULT 'overwrite'") {
		t.Fatalf("expected sqlite deploy_sync_mode alter SQL to default overwrite, got %q", sqliteSQL)
	}

	mysqlSQL := deploySyncModeAddColumnSQL(true)
	if !strings.Contains(mysqlSQL, "DEFAULT 'overwrite'") {
		t.Fatalf("expected mysql deploy_sync_mode alter SQL to default overwrite, got %q", mysqlSQL)
	}
}
