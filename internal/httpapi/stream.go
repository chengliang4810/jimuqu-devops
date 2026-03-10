package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"devops-pipeline/internal/model"
	"devops-pipeline/internal/store"
)

func (s *Server) handleStreamRun(w http.ResponseWriter, r *http.Request) {
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

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	_, _ = fmt.Fprint(w, "retry: 1000\n\n")
	if err = writeRunEvent(w, run); err != nil {
		return
	}
	flusher.Flush()

	if isTerminalStatus(run.Status) {
		return
	}

	lastUpdatedAt := run.UpdatedAt
	pollTicker := time.NewTicker(1 * time.Second)
	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer pollTicker.Stop()
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-pollTicker.C:
			run, err = s.store.GetRun(context.Background(), runID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) || errors.Is(err, context.Canceled) {
					return
				}
				s.logger.Error("stream get run failed", "run_id", runID, "error", err)
				return
			}

			if !run.UpdatedAt.Equal(lastUpdatedAt) {
				lastUpdatedAt = run.UpdatedAt
				if err = writeRunEvent(w, run); err != nil {
					return
				}
				flusher.Flush()
			}

			if isTerminalStatus(run.Status) {
				return
			}
		case <-heartbeatTicker.C:
			if _, err = fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeRunEvent(w http.ResponseWriter, run model.PipelineRun) error {
	body, err := json.Marshal(run)
	if err != nil {
		return err
	}

	if _, err = fmt.Fprintf(w, "event: run\ndata: %s\n\n", body); err != nil {
		return err
	}
	return nil
}

func isTerminalStatus(status string) bool {
	return status == model.RunStatusSuccess || status == model.RunStatusFailed
}
