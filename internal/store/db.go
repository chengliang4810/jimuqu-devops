package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	cryptoutil "devops-pipeline/internal/crypto"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

const (
	DriverSQLite = "sqlite"
	DriverMySQL  = "mysql"
)

type Store struct {
	db     *sql.DB
	cipher *cryptoutil.Cipher
	driver string
}

func Open(driver, source string) (*sql.DB, error) {
	switch normalizeDriver(driver) {
	case DriverSQLite:
		db, err := sql.Open(DriverSQLite, source)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}

		db.SetMaxOpenConns(1)
		db.SetConnMaxLifetime(0)

		if _, err = db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
			return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
		}

		return db, nil
	case DriverMySQL:
		if strings.TrimSpace(source) == "" {
			return nil, fmt.Errorf("open mysql: APP_DB_SOURCE is required")
		}

		db, err := sql.Open(DriverMySQL, source)
		if err != nil {
			return nil, fmt.Errorf("open mysql: %w", err)
		}

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(30 * time.Minute)

		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("ping mysql: %w", err)
		}

		return db, nil
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", driver)
	}
}

func New(db *sql.DB, cipher *cryptoutil.Cipher, driver string) *Store {
	return &Store{db: db, cipher: cipher, driver: normalizeDriver(driver)}
}

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", DriverSQLite:
		return DriverSQLite
	case DriverMySQL:
		return DriverMySQL
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func (s *Store) isSQLite() bool {
	return s.driver == "" || s.driver == DriverSQLite
}

func (s *Store) isMySQL() bool {
	return s.driver == DriverMySQL
}
