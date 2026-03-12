package store

import (
	"errors"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func TestWrapProjectMutationErrorReturnsConflict(t *testing.T) {
	err := wrapProjectMutationError("insert", &mysqlDriver.MySQLError{
		Number:  1062,
		Message: "Duplicate entry 'https://example.com/repo.git-main' for key 'projects.uniq_projects_repo_branch'",
	})

	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	if got, want := err.Error(), "project with the same repository URL and branch already exists"; got != want {
		t.Fatalf("unexpected error message: got %q want %q", got, want)
	}
}

func TestWrapProjectMutationErrorPassesThroughOtherErrors(t *testing.T) {
	original := errors.New("boom")
	err := wrapProjectMutationError("insert", original)

	if errors.Is(err, ErrConflict) {
		t.Fatalf("did not expect ErrConflict, got %v", err)
	}

	if !errors.Is(err, original) {
		t.Fatalf("expected original error to be wrapped, got %v", err)
	}
}

func TestIsConstraintError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "store conflict",
			err:  newConflictError("duplicate"),
			want: true,
		},
		{
			name: "mysql duplicate",
			err: &mysqlDriver.MySQLError{
				Number:  1062,
				Message: "Duplicate entry 'x' for key 'projects.uniq_projects_repo_branch'",
			},
			want: true,
		},
		{
			name: "mysql foreign key delete",
			err: &mysqlDriver.MySQLError{
				Number:  1451,
				Message: "Cannot delete or update a parent row: a foreign key constraint fails",
			},
			want: true,
		},
		{
			name: "sqlite unique",
			err:  errors.New("UNIQUE constraint failed: projects.repo_url, projects.branch"),
			want: true,
		},
		{
			name: "sqlite foreign key",
			err:  errors.New("FOREIGN KEY constraint failed"),
			want: true,
		},
		{
			name: "non constraint",
			err:  errors.New("network timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConstraintError(tt.err); got != tt.want {
				t.Fatalf("IsConstraintError() = %v, want %v", got, tt.want)
			}
		})
	}
}
