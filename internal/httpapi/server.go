package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"devops-pipeline/internal/auth"
	"devops-pipeline/internal/config"
	"devops-pipeline/internal/crypto"
	"devops-pipeline/internal/model"
	"devops-pipeline/internal/pipeline"
	"devops-pipeline/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	store      *store.Store
	executor   *pipeline.Executor
	logger     *slog.Logger
	config     config.Config
	jwtManager *auth.JWTManager
}

func New(store *store.Store, executor *pipeline.Executor, logger *slog.Logger, cfg config.Config) http.Handler {
	jwtManager := auth.NewJWTManager(cfg.Secret)
	server := &Server{
		store:      store,
		executor:   executor,
		logger:     logger,
		config:     cfg,
		jwtManager: jwtManager,
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(corsMiddleware)

	router.Get("/healthz", server.handleHealth)

	router.Route("/api/v1", func(r chi.Router) {
		// 公开接口 - 不需要认证
		r.Post("/admin/login", server.handleAdminLogin)
		r.Post("/webhooks/{token}", server.handleWebhook)

		// 需要认证的接口
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(server.jwtManager))

			r.Route("/hosts", func(r chi.Router) {
				r.Get("/", server.handleListHosts)
				r.Post("/", server.handleCreateHost)
				r.Put("/reorder", server.handleReorderHosts)
				r.Route("/{hostID}", func(r chi.Router) {
					r.Get("/", server.handleGetHost)
					r.Put("/", server.handleUpdateHost)
					r.Delete("/", server.handleDeleteHost)
				})
			})

			r.Route("/projects", func(r chi.Router) {
				r.Get("/", server.handleListProjects)
				r.Post("/", server.handleCreateProject)
				r.Put("/reorder", server.handleReorderProjects)
				r.Route("/{projectID}", func(r chi.Router) {
					r.Get("/", server.handleGetProject)
					r.Put("/", server.handleUpdateProject)
					r.Delete("/", server.handleDeleteProject)
					r.Post("/clone", server.handleCloneProject)
					r.Put("/deploy-config", server.handleUpsertDeployConfig)
					r.Get("/deploy-config", server.handleGetDeployConfig)
					r.Get("/runs", server.handleListProjectRuns)
					r.Post("/trigger", server.handleTriggerProject)
				})
			})

			r.Route("/notification-channels", func(r chi.Router) {
				r.Get("/", server.handleListNotificationChannels)
				r.Post("/", server.handleCreateNotificationChannel)
				r.Put("/reorder", server.handleReorderNotificationChannels)
				r.Route("/{channelID}", func(r chi.Router) {
					r.Get("/", server.handleGetNotificationChannel)
					r.Put("/", server.handleUpdateNotificationChannel)
					r.Delete("/", server.handleDeleteNotificationChannel)
					r.Put("/default", server.handleSetDefaultNotificationChannel)
					r.Post("/test", server.handleTestNotificationChannel)
				})
			})

			r.Route("/settings", func(r chi.Router) {
				r.Get("/", server.handleListSettings)
				r.Get("/backup", server.handleExportBackup)
				r.Post("/restore", server.handleImportBackup)
				r.Put("/{key}", server.handleUpdateSetting)
			})

			r.Route("/update", func(r chi.Router) {
				r.Get("/", server.handleGetLatestRelease)
				r.Get("/now-version", server.handleGetUpdateStatus)
				r.Post("/", server.handleApplyUpdate)
			})

			r.Route("/admin", func(r chi.Router) {
				r.Get("/profile", server.handleGetAdminProfile)
				r.Put("/username", server.handleChangeAdminUsername)
				r.Put("/password", server.handleChangeAdminPassword)
			})

			r.Get("/runs", server.handleListAllRuns)
			r.Delete("/runs", server.handleClearRuns)
			r.Get("/runs/{runID}", server.handleGetRun)
			r.Get("/stats", server.handleStats)
			r.Get("/dashboard/home", server.handleHomeDashboard)
			r.Get("/system/info", server.handleSystemInfo)
		})

		// 流式接口移到认证组外面，支持查询参数传递token
		r.Get("/runs/{runID}/stream", server.handleStreamRun)
	})

	server.mountStaticUI(router)

	return router
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.store.ApplyRunRetention(ctx); err != nil {
		s.writeError(w, err)
		return
	}

	// 获取各模块数量
	hosts, err := s.store.ListHosts(ctx)
	if err != nil {
		s.writeError(w, err)
		return
	}

	projects, err := s.store.ListProjects(ctx)
	if err != nil {
		s.writeError(w, err)
		return
	}

	runs, err := s.store.ListAllRuns(ctx, 0, 1000)
	if err != nil {
		s.writeError(w, err)
		return
	}

	channels, err := s.store.ListNotificationChannels(ctx)
	if err != nil {
		s.writeError(w, err)
		return
	}

	stats := map[string]int{
		"host_count":           len(hosts),
		"project_count":        len(projects),
		"run_count":            len(runs),
		"notify_channel_count": len(channels),
	}

	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleHomeDashboard(w http.ResponseWriter, r *http.Request) {
	if err := s.store.ApplyRunRetention(r.Context()); err != nil {
		s.writeError(w, err)
		return
	}
	dashboard, err := s.store.GetHomeDashboard(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

func (s *Server) handleListHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := s.store.ListHosts(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, hosts)
}

func (s *Server) handleCreateHost(w http.ResponseWriter, r *http.Request) {
	var input model.HostUpsert
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	if input.Port == 0 {
		input.Port = 22
	}
	if err := validateHostInput(input, true); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	host, err := s.store.CreateHost(r.Context(), input)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, host)
}

func (s *Server) handleGetHost(w http.ResponseWriter, r *http.Request) {
	hostID, err := parseInt64Param(r, "hostID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	host, err := s.store.GetHost(r.Context(), hostID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, host)
}

func (s *Server) handleUpdateHost(w http.ResponseWriter, r *http.Request) {
	hostID, err := parseInt64Param(r, "hostID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	var input model.HostUpsert
	if err = decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if input.Port == 0 {
		input.Port = 22
	}
	if err = validateHostInput(input, false); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	host, err := s.store.UpdateHost(r.Context(), hostID, input)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, host)
}

func (s *Server) handleDeleteHost(w http.ResponseWriter, r *http.Request) {
	hostID, err := parseInt64Param(r, "hostID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if err = s.store.DeleteHost(r.Context(), hostID); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReorderHosts(w http.ResponseWriter, r *http.Request) {
	var input model.ReorderInput
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if len(input.IDs) == 0 {
		s.writeBadRequest(w, errors.New("ids are required"))
		return
	}
	if err := s.store.ReorderHosts(r.Context(), input.IDs); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.store.ListProjects(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var input model.ProjectUpsert
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if err := validateProjectInput(input, nil); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	project, err := s.store.CreateProject(r.Context(), input)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseInt64Param(r, "projectID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	project, err := s.store.GetProjectDetail(r.Context(), projectID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseInt64Param(r, "projectID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	var input model.ProjectUpsert
	if err = decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	currentProject, err := s.store.GetProject(r.Context(), projectID)
	if err != nil {
		s.writeError(w, err)
		return
	}

	if err = validateProjectInput(input, &currentProject); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	project, err := s.store.UpdateProject(r.Context(), projectID, input)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseInt64Param(r, "projectID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if err = s.store.DeleteProject(r.Context(), projectID); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReorderProjects(w http.ResponseWriter, r *http.Request) {
	var input model.ReorderInput
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if len(input.IDs) == 0 {
		s.writeBadRequest(w, errors.New("ids are required"))
		return
	}
	if err := s.store.ReorderProjects(r.Context(), input.IDs); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCloneProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseInt64Param(r, "projectID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	var input model.ProjectCloneInput
	if err = decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Branch) == "" {
		s.writeBadRequest(w, errors.New("name and branch are required"))
		return
	}

	project, err := s.store.CloneProject(r.Context(), projectID, input)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (s *Server) handleUpsertDeployConfig(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseInt64Param(r, "projectID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	var input model.DeployConfigUpsert
	if err = decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if err = validateDeployConfigInput(input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	config, err := s.store.UpsertDeployConfig(r.Context(), projectID, input)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func (s *Server) handleGetDeployConfig(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseInt64Param(r, "projectID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	config, err := s.store.GetDeployConfigByProjectID(r.Context(), projectID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func (s *Server) handleListProjectRuns(w http.ResponseWriter, r *http.Request) {
	if err := s.store.ApplyRunRetention(r.Context()); err != nil {
		s.writeError(w, err)
		return
	}
	projectID, err := parseInt64Param(r, "projectID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	runs, err := s.store.ListRunsByProject(r.Context(), projectID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (s *Server) handleListAllRuns(w http.ResponseWriter, r *http.Request) {
	if err := s.store.ApplyRunRetention(r.Context()); err != nil {
		s.writeError(w, err)
		return
	}
	// 解析分页参数，默认显示前100条记录
	offset := 0
	limit := 100

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	runs, err := s.store.ListAllRuns(r.Context(), offset, limit)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (s *Server) handleTriggerProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseInt64Param(r, "projectID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	run, err := s.executor.Trigger(r.Context(), projectID, model.TriggerTypeManual, "manual")
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	if err := s.store.ApplyRunRetention(r.Context()); err != nil {
		s.writeError(w, err)
		return
	}
	runID, err := parseInt64Param(r, "runID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	run, err := s.store.GetRun(r.Context(), runID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if strings.TrimSpace(token) == "" {
		s.writeBadRequest(w, errors.New("missing webhook token"))
		return
	}

	project, err := s.store.GetProjectByWebhookToken(r.Context(), token)
	if err != nil {
		s.writeError(w, err)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		s.writeBadRequest(w, fmt.Errorf("read webhook body: %w", err))
		return
	}

	branch, triggerRef, err := extractWebhookBranch(r.Context(), body, r.Header)
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if branch != "" && branch != project.Branch {
		s.writeBadRequest(w, fmt.Errorf("branch mismatch: expected %s got %s", project.Branch, branch))
		return
	}
	if triggerRef == "" {
		triggerRef = branch
	}

	run, err := s.executor.Trigger(r.Context(), project.ID, model.TriggerTypeWebhook, triggerRef)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}

func (s *Server) writeBadRequest(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
}

func (s *Server) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case store.IsConstraintError(err):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	default:
		s.logger.Error("request failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func decodeJSON(body io.ReadCloser, target any) error {
	defer body.Close()
	decoder := json.NewDecoder(io.LimitReader(body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseInt64Param(r *http.Request, key string) (int64, error) {
	raw := chi.URLParam(r, key)
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}

func validateHostInput(input model.HostUpsert, requirePassword bool) error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("host name is required")
	}
	if strings.TrimSpace(input.Address) == "" {
		return errors.New("host address is required")
	}
	if input.Port <= 0 {
		return errors.New("host port must be positive")
	}
	if strings.TrimSpace(input.Username) == "" {
		return errors.New("host username is required")
	}
	if requirePassword && (input.Password == nil || strings.TrimSpace(*input.Password) == "") {
		return errors.New("host password is required")
	}
	return nil
}

func validateProjectInput(input model.ProjectUpsert, current *model.Project) error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("project name is required")
	}
	if strings.TrimSpace(input.RepoURL) == "" {
		return errors.New("repo_url is required")
	}
	if strings.TrimSpace(input.Branch) == "" {
		return errors.New("branch is required")
	}

	// 验证Git认证配置
	gitAuthType := input.GitAuthType
	if gitAuthType == "" {
		gitAuthType = model.GitAuthTypeNone
	}

	validGitTypes := map[string]bool{
		model.GitAuthTypeNone:     true,
		model.GitAuthTypeUsername: true,
		model.GitAuthTypeToken:    true,
		model.GitAuthTypeSSH:      true,
	}

	if !validGitTypes[gitAuthType] {
		return errors.New("invalid git_auth_type, must be one of: none, username, token, ssh")
	}

	// 根据认证类型验证必填字段
	switch gitAuthType {
	case model.GitAuthTypeUsername, model.GitAuthTypeToken:
		gitUsername := trimStringPtr(input.GitUsername)
		gitPassword := trimStringPtr(input.GitPassword)
		existingUsername := ""
		existingPassword := ""
		if current != nil {
			existingUsername = strings.TrimSpace(current.GitUsername)
			existingPassword = strings.TrimSpace(current.GitPassword)
		}
		if gitUsername == "" && existingUsername == "" {
			return errors.New("git_username is required for username/token authentication")
		}
		if gitPassword == "" && existingPassword == "" {
			return errors.New("git_password/token is required for username/token authentication")
		}
	case model.GitAuthTypeSSH:
		gitSSHKey := trimStringPtr(input.GitSSHKey)
		existingSSHKey := ""
		if current != nil {
			existingSSHKey = strings.TrimSpace(current.GitSSHKey)
		}
		if gitSSHKey == "" && existingSSHKey == "" {
			return errors.New("git_ssh_key is required for ssh authentication")
		}
	}

	return nil
}

func trimStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func validateDeployConfigInput(input model.DeployConfigUpsert) error {
	if input.HostID <= 0 {
		return errors.New("host_id is required")
	}
	if strings.TrimSpace(input.BuildImage) == "" {
		return errors.New("build_image is required")
	}
	if len(input.BuildCommands) == 0 {
		return errors.New("build_commands cannot be empty")
	}
	switch input.ArtifactFilterMode {
	case "", model.ArtifactFilterNone, model.ArtifactFilterInclude, model.ArtifactFilterExclude:
	default:
		return errors.New("artifact_filter_mode must be one of none/include/exclude")
	}
	// 当过滤规则为空时，自动将模式设为none
	if len(input.ArtifactRules) == 0 {
		input.ArtifactFilterMode = model.ArtifactFilterNone
	}
	if strings.TrimSpace(input.RemoteSaveDir) == "" {
		return errors.New("remote_save_dir is required")
	}
	if strings.TrimSpace(input.RemoteDeployDir) == "" {
		return errors.New("remote_deploy_dir is required")
	}
	return nil
}

type webhookPayload struct {
	Ref    string `json:"ref"`
	Branch string `json:"branch"`
	After  string `json:"after"`
	Push   struct {
		Changes []struct {
			New struct {
				Name string `json:"name"`
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
}

func extractWebhookBranch(ctx context.Context, body []byte, headers http.Header) (branch string, triggerRef string, err error) {
	if len(body) == 0 {
		ref := headers.Get("X-Git-Ref")
		return normalizeRef(ref), ref, nil
	}

	var payload webhookPayload
	if err = json.Unmarshal(body, &payload); err != nil {
		return "", "", fmt.Errorf("invalid webhook payload: %w", err)
	}

	switch {
	case payload.Ref != "":
		return normalizeRef(payload.Ref), payload.Ref, nil
	case payload.Branch != "":
		return strings.TrimSpace(payload.Branch), payload.Branch, nil
	case len(payload.Push.Changes) > 0 && payload.Push.Changes[0].New.Name != "":
		name := payload.Push.Changes[0].New.Name
		return name, name, nil
	case headers.Get("X-Git-Ref") != "":
		ref := headers.Get("X-Git-Ref")
		return normalizeRef(ref), ref, nil
	default:
		return "", "", nil
	}
}

func normalizeRef(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "refs/heads/")
	return ref
}

// 通知渠道管理处理器

func (s *Server) handleListNotificationChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := s.store.ListNotificationChannels(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, channels)
}

func (s *Server) handleCreateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	var input model.NotificationChannelUpsert
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	if err := validateNotificationChannelInput(input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	channel, err := s.store.CreateNotificationChannel(r.Context(), input)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, channel)
}

func (s *Server) handleGetNotificationChannel(w http.ResponseWriter, r *http.Request) {
	channelID, err := parseInt64Param(r, "channelID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	channel, err := s.store.GetNotificationChannelWithConfig(r.Context(), channelID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, channel)
}

func (s *Server) handleUpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	channelID, err := parseInt64Param(r, "channelID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	var input model.NotificationChannelUpsert
	if err = decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	if err := validateNotificationChannelInput(input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	channel, err := s.store.UpdateNotificationChannel(r.Context(), channelID, input)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, channel)
}

func (s *Server) handleDeleteNotificationChannel(w http.ResponseWriter, r *http.Request) {
	channelID, err := parseInt64Param(r, "channelID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	if err = s.store.DeleteNotificationChannel(r.Context(), channelID); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSetDefaultNotificationChannel(w http.ResponseWriter, r *http.Request) {
	channelID, err := parseInt64Param(r, "channelID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	if err = s.store.SetDefaultNotificationChannel(r.Context(), channelID); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReorderNotificationChannels(w http.ResponseWriter, r *http.Request) {
	var input model.ReorderInput
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if len(input.IDs) == 0 {
		s.writeBadRequest(w, errors.New("ids are required"))
		return
	}
	if err := s.store.ReorderNotificationChannels(r.Context(), input.IDs); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTestNotificationChannel(w http.ResponseWriter, r *http.Request) {
	channelID, err := parseInt64Param(r, "channelID")
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	var input model.TestNotificationInput
	if err = decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	channel, err := s.store.GetNotificationChannel(r.Context(), channelID)
	if err != nil {
		s.writeError(w, err)
		return
	}

	// 发送测试通知
	s.logger.Info("测试通知渠道", "channel_id", channelID, "channel_name", channel.Name, "type", channel.Type)

	if err = s.sendTestNotification(&channel, input); err != nil {
		s.logger.Error("发送测试通知失败", "error", err, "channel_id", channelID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.logger.Info("测试通知发送成功", "channel_id", channelID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func validateNotificationChannelInput(input model.NotificationChannelUpsert) error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("channel name is required")
	}

	validTypes := map[string]bool{
		model.ChannelTypeWebhook:  true,
		model.ChannelTypeWeChat:   true,
		model.ChannelTypeDingTalk: true,
		model.ChannelTypeFeishu:   true,
		model.ChannelTypeEmail:    true,
	}
	if !validTypes[input.Type] {
		return errors.New("invalid channel type, must be one of: webhook, wechat, dingtalk, feishu, email")
	}

	// 根据类型验证配置
	switch input.Type {
	case model.ChannelTypeWebhook:
		if url, ok := input.Config["url"].(string); !ok || strings.TrimSpace(url) == "" {
			return errors.New("webhook url is required")
		}
	case model.ChannelTypeWeChat:
		if url, ok := input.Config["webhook_url"].(string); !ok || strings.TrimSpace(url) == "" {
			return errors.New("wechat webhook_url is required")
		}
	case model.ChannelTypeDingTalk:
		if url, ok := input.Config["webhook_url"].(string); !ok || strings.TrimSpace(url) == "" {
			return errors.New("dingtalk webhook_url is required")
		}
	case model.ChannelTypeFeishu:
		if url, ok := input.Config["webhook_url"].(string); !ok || strings.TrimSpace(url) == "" {
			return errors.New("feishu webhook_url is required")
		}
	case model.ChannelTypeEmail:
		if host, ok := input.Config["smtp_host"].(string); !ok || strings.TrimSpace(host) == "" {
			return errors.New("email smtp_host is required")
		}
		if username, ok := input.Config["username"].(string); !ok || strings.TrimSpace(username) == "" {
			return errors.New("email username is required")
		}
		if to, ok := input.Config["to"].(string); !ok || strings.TrimSpace(to) == "" {
			return errors.New("email to is required")
		}
	}

	return nil
}

// 管理员认证处理器

func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var input model.LoginRequest
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	if strings.TrimSpace(input.Username) == "" || strings.TrimSpace(input.Password) == "" {
		s.writeBadRequest(w, errors.New("username and password are required"))
		return
	}

	// 获取管理员用户
	admin, err := s.store.GetAdminUser(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(input.Password)); err != nil {
		s.writeBadRequest(w, errors.New("invalid username or password"))
		return
	}

	// 生成JWT token
	jwtManager := auth.NewJWTManager(s.config.Secret)
	token, err := jwtManager.GenerateToken(admin.ID, admin.Username, 24*time.Hour) // 24小时有效期
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	response := model.LoginResponse{
		Token:    token,
		Username: admin.Username,
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) sendTestNotification(channel *model.NotificationChannel, input model.TestNotificationInput) error {
	var config map[string]any
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	switch channel.Type {
	case model.ChannelTypeDingTalk:
		return s.sendDingTalkNotification(config, input.Title, input.Content)
	case model.ChannelTypeWeChat:
		return s.sendWeChatNotification(config, input.Title, input.Content)
	case model.ChannelTypeFeishu:
		return s.sendFeishuNotification(config, input.Title, input.Content)
	case model.ChannelTypeWebhook:
		return s.sendWebhookNotification(config, input.Title, input.Content)
	default:
		return fmt.Errorf("不支持的通知类型: %s", channel.Type)
	}
}

func (s *Server) sendDingTalkNotification(config map[string]any, title, content string) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return errors.New("钉钉 Webhook URL 为空")
	}

	// 如果有签名，计算签名并添加到URL
	if secret, ok := config["secret"].(string); ok && secret != "" {
		timestamp := time.Now().UnixMilli()
		signStr := fmt.Sprintf("%d\n%s", timestamp, secret)
		hmac := crypto.HMACSHA256([]byte(signStr), []byte(secret))
		sign := base64.StdEncoding.EncodeToString(hmac)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, url.QueryEscape(sign))
		s.logger.Info("钉钉通知使用签名", "timestamp", timestamp)
	}

	// 钉钉机器人要求消息中必须包含关键词，添加常见关键词
	messageContent := fmt.Sprintf("%s\n%s\n\n【通知】【部署】", title, content)

	message := map[string]any{
		"msgtype": "text",
		"text": map[string]string{
			"content": messageContent,
		},
	}

	return s.sendHTTPRequest(webhookURL, message)
}

func (s *Server) sendWeChatNotification(config map[string]any, title, content string) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return errors.New("企业微信 Webhook URL 为空")
	}

	message := map[string]any{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n%s", title, content),
		},
	}

	return s.sendHTTPRequest(webhookURL, message)
}

func (s *Server) sendFeishuNotification(config map[string]any, title, content string) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return errors.New("飞书 Webhook URL 为空")
	}

	message := map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": fmt.Sprintf("%s\n%s", title, content),
		},
	}

	return s.sendHTTPRequest(webhookURL, message)
}

func (s *Server) sendWebhookNotification(config map[string]any, title, content string) error {
	webhookURL, ok := config["url"].(string)
	if !ok || webhookURL == "" {
		return errors.New("Webhook URL 为空")
	}

	message := map[string]any{
		"title":   title,
		"content": content,
	}

	return s.sendHTTPRequest(webhookURL, message)
}

func (s *Server) sendHTTPRequest(url string, payload any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	s.logger.Info("发送HTTP请求", "url", url, "payload", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	s.logger.Info("收到响应", "status", resp.StatusCode, "body", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
