package store

import (
	"strings"
	"testing"
)

func TestAppendRunLogQuerySQLite(t *testing.T) {
	query := appendRunLogQuery(false)

	if !strings.Contains(query, "COALESCE(log_text, '') || ?") {
		t.Fatalf("expected sqlite log append query, got %q", query)
	}
	if strings.Contains(query, "CONCAT(") {
		t.Fatalf("did not expect mysql concat query, got %q", query)
	}
}

func TestAppendRunLogQueryMySQL(t *testing.T) {
	query := appendRunLogQuery(true)

	if !strings.Contains(query, "CONCAT(COALESCE(log_text, ''), ?)") {
		t.Fatalf("expected mysql log append query, got %q", query)
	}
	if strings.Contains(query, "|| ?") {
		t.Fatalf("did not expect sqlite concatenation operator, got %q", query)
	}
}

func TestRunSelectQueryWithoutLogText(t *testing.T) {
	query := runSelectQuery(false)

	if !strings.Contains(query, "'' AS log_text") {
		t.Fatalf("expected summary run query to avoid loading log_text, got %q", query)
	}
	if strings.Contains(query, "pipeline_runs.log_text AS log_text") {
		t.Fatalf("did not expect summary run query to select pipeline_runs.log_text, got %q", query)
	}
}

func TestRunSelectQueryWithLogText(t *testing.T) {
	query := runSelectQuery(true)

	if !strings.Contains(query, "pipeline_runs.log_text AS log_text") {
		t.Fatalf("expected detail run query to include log_text, got %q", query)
	}
}
