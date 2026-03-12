package store

import (
	"strings"
	"testing"
)

func TestDeployConfigUpsertQuerySQLite(t *testing.T) {
	query := deployConfigUpsertQuery(false)

	if !strings.Contains(query, "ON CONFLICT(project_id) DO UPDATE SET") {
		t.Fatalf("expected sqlite upsert query, got %q", query)
	}
	if strings.Contains(query, "ON DUPLICATE KEY UPDATE") {
		t.Fatalf("did not expect mysql upsert clause, got %q", query)
	}
}

func TestDeployConfigUpsertQueryMySQL(t *testing.T) {
	query := deployConfigUpsertQuery(true)

	if !strings.Contains(query, "ON DUPLICATE KEY UPDATE") {
		t.Fatalf("expected mysql upsert query, got %q", query)
	}
	if strings.Contains(query, "ON CONFLICT(project_id) DO UPDATE SET") {
		t.Fatalf("did not expect sqlite upsert clause, got %q", query)
	}
	if !strings.Contains(query, "notification_channel_id = VALUES(notification_channel_id)") {
		t.Fatalf("expected notification channel to be updated, got %q", query)
	}
}
