package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Addr          string
	DataDir       string
	DBDriver      string
	DBSource      string
	WorkspaceDir  string
	ArtifactDir   string
	Secret        string
	JWTSecret     string
	AdminUsername string
	AdminPassword string
}

func Load() Config {
	dataDir := env("APP_DATA_DIR", filepath.Join(".", "data"))
	dbDriver := env("APP_DB_DRIVER", "sqlite")
	defaultSQLitePath := filepath.Join(dataDir, "pipeline.db")
	dbSource := env("APP_DB_SOURCE", defaultSQLitePath)

	return Config{
		Addr:          env("APP_ADDR", ":18080"),
		DataDir:       dataDir,
		DBDriver:      dbDriver,
		DBSource:      dbSource,
		WorkspaceDir:  env("APP_WORKSPACE_DIR", filepath.Join(dataDir, "workspaces")),
		ArtifactDir:   env("APP_ARTIFACT_DIR", filepath.Join(dataDir, "artifacts")),
		Secret:        env("APP_SECRET", "change-me-in-production"),
		JWTSecret:     env("JWT_SECRET", "change-me-in-production"),
		AdminUsername: env("ADMIN_USERNAME", "admin"),
		AdminPassword: env("ADMIN_PASSWORD", "admin123"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
