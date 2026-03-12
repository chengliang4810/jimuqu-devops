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
