package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"devops-pipeline/internal/model"
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

func validateSettingKey(key string) error {
	switch key {
	case model.SettingDockerMirrorURL, model.SettingProxyURL, model.SettingRunRetentionDays:
		return nil
	default:
		return errors.New("unsupported setting key")
	}
}

func validateSettingValue(key, value string) error {
	value = strings.TrimSpace(value)
	switch key {
	case model.SettingDockerMirrorURL, model.SettingProxyURL:
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
