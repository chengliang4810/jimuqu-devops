package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"devops-pipeline/internal/config"
	cryptoutil "devops-pipeline/internal/crypto"
	"devops-pipeline/internal/httpapi"
	"devops-pipeline/internal/pipeline"
	"devops-pipeline/internal/store"

	"golang.org/x/crypto/bcrypt"
)

type App struct {
	store    *store.Store
	executor *pipeline.Executor
	handler  http.Handler
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	dirs := []string{
		cfg.DataDir,
		cfg.WorkspaceDir,
		cfg.ArtifactDir,
	}
	if cfg.DBDriver == "" || cfg.DBDriver == store.DriverSQLite {
		dirs = append(dirs, filepath.Dir(cfg.DBSource))
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create data dir %s: %w", dir, err)
		}
	}

	db, err := store.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		return nil, err
	}

	appStore := store.New(db, cryptoutil.New(cfg.Secret), cfg.DBDriver)
	if err = appStore.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	// 初始化管理员用户
	if err = initializeAdminUser(context.Background(), appStore, cfg); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize admin user: %w", err)
	}
	if err = appStore.ApplyRunRetention(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply run retention: %w", err)
	}

	executor := pipeline.NewExecutor(appStore, logger, cfg.WorkspaceDir, cfg.ArtifactDir)

	return &App{
		store:    appStore,
		executor: executor,
		handler:  httpapi.New(appStore, executor, logger, cfg),
	}, nil
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) Close() error {
	return a.store.Close()
}

// FixRunningRuns 修复所有卡在运行中的部署记录
func (a *App) FixRunningRuns(ctx context.Context) error {
	// 获取最近1000条记录进行修复，避免数据量过大
	runs, err := a.store.ListAllRuns(ctx, 0, 1000)
	if err != nil {
		return err
	}

	fixedCount := 0
	for _, run := range runs {
		if run.Status == "running" {
			if err := a.store.FinalizeRun(ctx, run.ID, "failed", "deployment interrupted by server restart"); err != nil {
				continue
			}
			fixedCount++
		}
	}

	if fixedCount > 0 {
		fmt.Printf("Fixed %d interrupted deployment records\n", fixedCount)
	}

	return nil
}

func initializeAdminUser(ctx context.Context, store *store.Store, cfg config.Config) error {
	// 检查管理员用户是否已存在
	exists, err := store.AdminUserExists(ctx)
	if err != nil {
		return err
	}

	// 如果管理员用户不存在，则创建
	if !exists {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generate password hash: %w", err)
		}

		_, err = store.CreateAdminUser(ctx, cfg.AdminUsername, string(passwordHash))
		if err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}

		// 记录日志
		fmt.Printf("Admin user created: username=%s\n", cfg.AdminUsername)
		fmt.Println("Please change the default password for security!")
	}

	return nil
}
