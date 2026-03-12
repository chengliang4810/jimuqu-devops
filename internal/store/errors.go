package store

import (
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type conflictError struct {
	message string
}

func (e conflictError) Error() string {
	return e.message
}

func (e conflictError) Unwrap() error {
	return ErrConflict
}

func newConflictError(message string) error {
	return conflictError{message: message}
}

func wrapProjectMutationError(action string, err error) error {
	if isProjectRepoBranchConflict(err) {
		return newConflictError("project with the same repository URL and branch already exists")
	}
	return fmt.Errorf("%s project: %w", action, err)
}

func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrConflict) {
		return true
	}
	return isUniqueConstraintError(err) || isForeignKeyConstraintError(err)
}

func isProjectRepoBranchConflict(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "uniq_projects_repo_branch")
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed: projects.repo_url, projects.branch")
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed") || strings.Contains(message, "duplicate entry")
}

func isForeignKeyConstraintError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1451 || mysqlErr.Number == 1452
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "foreign key constraint failed") || strings.Contains(message, "cannot delete or update a parent row") || strings.Contains(message, "cannot add or update a child row")
}
