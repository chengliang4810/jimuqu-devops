package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"devops-pipeline/internal/model"
	"devops-pipeline/internal/update"
	"devops-pipeline/internal/version"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) handleListSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.ListSettings(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleUpdateSetting(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(chi.URLParam(r, "key"))
	if err := validateSettingKey(key); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	var input model.SettingUpsert
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}
	if err := validateSettingValue(key, input.Value); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	setting, err := s.store.SetSetting(r.Context(), key, strings.TrimSpace(input.Value))
	if err != nil {
		s.writeError(w, err)
		return
	}

	if key == model.SettingRunRetentionDays {
		if err := s.store.ApplyRunRetention(r.Context()); err != nil {
			s.writeError(w, err)
			return
		}
	}

	writeJSON(w, http.StatusOK, setting)
}

func (s *Server) handleExportBackup(w http.ResponseWriter, r *http.Request) {
	backup, err := s.store.ExportBackup(r.Context(), version.RepoURL, version.Current())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, backup)
}

func (s *Server) handleImportBackup(w http.ResponseWriter, r *http.Request) {
	var backup model.BackupData
	if err := decodeJSON(r.Body, &backup); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	result, err := s.store.ImportBackup(r.Context(), backup)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetAdminProfile(w http.ResponseWriter, r *http.Request) {
	admin, err := s.store.GetAdminUser(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, model.AccountProfile{Username: admin.Username})
}

func (s *Server) handleChangeAdminUsername(w http.ResponseWriter, r *http.Request) {
	var input model.ChangeUsernameInput
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	input.NewUsername = strings.TrimSpace(input.NewUsername)
	if input.NewUsername == "" {
		s.writeBadRequest(w, errors.New("new_username is required"))
		return
	}

	if err := s.store.UpdateAdminUsername(r.Context(), input.NewUsername); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleChangeAdminPassword(w http.ResponseWriter, r *http.Request) {
	var input model.ChangePasswordInput
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	if strings.TrimSpace(input.OldPassword) == "" || strings.TrimSpace(input.NewPassword) == "" {
		s.writeBadRequest(w, errors.New("old_password and new_password are required"))
		return
	}
	if len(input.NewPassword) < 6 {
		s.writeBadRequest(w, errors.New("new password must be at least 6 characters"))
		return
	}

	admin, err := s.store.GetAdminUser(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(input.OldPassword)); err != nil {
		s.writeBadRequest(w, errors.New("old password is incorrect"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := s.store.UpdateAdminPassword(r.Context(), string(passwordHash)); err != nil {
		s.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleClearRuns(w http.ResponseWriter, r *http.Request) {
	affected, err := s.store.ClearRuns(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"cleared": affected})
}

func (s *Server) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.SystemInfo{
		RepoURL: version.RepoURL,
		Version: version.Current(),
	})
}

func (s *Server) handleGetLatestRelease(w http.ResponseWriter, r *http.Request) {
	proxyURL := s.getProxyURL(r)
	release, err := update.GetLatestRelease(r.Context(), proxyURL)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, release)
}

func (s *Server) handleGetUpdateStatus(w http.ResponseWriter, r *http.Request) {
	proxyURL := s.getProxyURL(r)
	status, err := update.GetUpdateStatus(r.Context(), proxyURL)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleApplyUpdate(w http.ResponseWriter, r *http.Request) {
	activeRuns, err := s.store.CountActiveRuns(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	if activeRuns > 0 {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "存在正在部署中的任务，请等待部署完成后再更新"})
		return
	}

	proxyURL := s.getProxyURL(r)
	result, err := update.ApplyUpdate(r.Context(), proxyURL)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
	update.ScheduleRestartAndExit()
}

func validateSettingKey(key string) error {
	switch key {
	case model.SettingDockerMirrorURL, model.SettingGitDockerImage, model.SettingBuildCacheDirs, model.SettingPublicBaseURL, model.SettingProxyURL, model.SettingRunRetentionDays:
		return nil
	default:
		return errors.New("unsupported setting key")
	}
}

func (s *Server) getProxyURL(r *http.Request) string {
	proxyURL, err := s.store.GetSettingValue(r.Context(), model.SettingProxyURL)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(proxyURL)
}

func validateSettingValue(key, value string) error {
	value = strings.TrimSpace(value)
	switch key {
	case model.SettingDockerMirrorURL, model.SettingProxyURL, model.SettingPublicBaseURL:
		return nil
	case model.SettingGitDockerImage:
		if value == "" {
			return errors.New("git_docker_image is required")
		}
		return nil
	case model.SettingBuildCacheDirs:
		for _, cacheDir := range strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n") {
			trimmed := strings.TrimSpace(cacheDir)
			if trimmed == "" {
				continue
			}
			if model.NormalizeCacheDir(trimmed) == "" {
				return errors.New("build_cache_dirs must contain absolute container paths such as /root/.m2")
			}
		}
		return nil
	case model.SettingRunRetentionDays:
		if value == "" {
			return errors.New("run_retention_days is required")
		}
		days, err := strconv.Atoi(value)
		if err != nil || days <= 0 {
			return errors.New("run_retention_days must be a positive integer")
		}
		return nil
	default:
		return errors.New("unsupported setting key")
	}
}
