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
