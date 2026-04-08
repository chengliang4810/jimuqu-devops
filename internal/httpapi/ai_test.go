package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"devops-pipeline/internal/config"
	cryptoutil "devops-pipeline/internal/crypto"
	"devops-pipeline/internal/model"
	"devops-pipeline/internal/store"

	"github.com/go-chi/chi/v5"
)

func newTestHTTPStore(t *testing.T) *store.Store {
	t.Helper()

	db, err := store.Open(store.DriverSQLite, "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	appStore := store.New(db, cryptoutil.New("httpapi-test-secret"), store.DriverSQLite)
	if err := appStore.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return appStore
}

func newTestServer(t *testing.T, appStore *store.Store) *Server {
	t.Helper()

	return &Server{
		store:  appStore,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		config: config.Config{Secret: "httpapi-test-secret"},
	}
}

func addURLParam(r *http.Request, key, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeContext))
}

func newInterpretRequest(runID int64) *http.Request {
	return addURLParam(
		httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+strconv.FormatInt(runID, 10)+"/interpret", nil),
		"runID",
		strconv.FormatInt(runID, 10),
	)
}

func createTestProjectAndRun(t *testing.T, appStore *store.Store, status string) model.PipelineRun {
	t.Helper()

	project, err := appStore.CreateProject(context.Background(), model.ProjectUpsert{
		Name:        "AI Demo",
		RepoURL:     "https://example.com/demo.git",
		Branch:      "main",
		Description: "demo",
		GitAuthType: model.GitAuthTypeNone,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	run, err := appStore.CreateRun(context.Background(), model.RunCreateInput{
		ProjectID:   project.ID,
		Status:      model.RunStatusQueued,
		TriggerType: model.TriggerTypeManual,
		TriggerRef:  "manual",
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	if err := appStore.AppendRunLog(context.Background(), run.ID, "step 1: build failed\nnpm ERR! missing script: build\n"); err != nil {
		t.Fatalf("append run log: %v", err)
	}
	if err := appStore.FinalizeRun(context.Background(), run.ID, status, "docker build exited with code 127"); err != nil {
		t.Fatalf("finalize run: %v", err)
	}

	loaded, err := appStore.GetRun(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("reload run: %v", err)
	}

	return loaded
}

func TestTruncateAILog(t *testing.T) {
	truncated, changed := truncateAILog(strings.Repeat("a", aiLogCharacterLimit+100))
	if !changed {
		t.Fatalf("expected log truncation to report changed=true")
	}
	if len(truncated) != aiLogCharacterLimit {
		t.Fatalf("expected truncated log length %d, got %d", aiLogCharacterLimit, len(truncated))
	}
}

func TestRequestOpenAIInterpretation(t *testing.T) {
	var capturedPath string
	var capturedAuth string
	var capturedRequest openAIChatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAuth = r.Header.Get("Authorization")

		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]string{
						"content": "失败摘要\n构建脚本缺失\n\n可能原因\npackage.json 未定义 build\n\n建议操作\n补充 build 脚本后重试",
					},
				},
			},
		})
	}))
	defer server.Close()

	content, err := requestOpenAIInterpretation(context.Background(), server.Client(), model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolOpenAI,
		BaseURL:  server.URL + "/v1",
		APIKey:   "plain-api-key",
		Model:    "gpt-test",
	}, aiInterpretationInput{
		ProjectName:   "demo",
		Branch:        "main",
		Status:        model.RunStatusFailed,
		CommitMessage: "",
		ErrorMessage:  "build failed",
		LogText:       "npm ERR! missing script: build",
	})
	if err != nil {
		t.Fatalf("request openai interpretation: %v", err)
	}

	if capturedPath != "/v1/chat/completions" {
		t.Fatalf("expected request path /v1/chat/completions, got %q", capturedPath)
	}
	if capturedAuth != "Bearer plain-api-key" {
		t.Fatalf("expected bearer token header, got %q", capturedAuth)
	}
	if capturedRequest.Model != "gpt-test" {
		t.Fatalf("expected request model gpt-test, got %q", capturedRequest.Model)
	}
	if len(capturedRequest.Messages) == 0 {
		t.Fatalf("expected at least one request message")
	}
	if !strings.Contains(content, "失败摘要") {
		t.Fatalf("expected returned interpretation content, got %q", content)
	}
}

func TestHandleInterpretRunRejectsNonFailedRun(t *testing.T) {
	appStore := newTestHTTPStore(t)
	server := newTestServer(t, appStore)
	run := createTestProjectAndRun(t, appStore, model.RunStatusSuccess)

	if _, err := appStore.SetAISettings(context.Background(), model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolOpenAI,
		BaseURL:  "https://example.com/v1",
		APIKey:   "plain-api-key",
		Model:    "gpt-test",
	}); err != nil {
		t.Fatalf("set ai settings: %v", err)
	}

	request := newInterpretRequest(run.ID)
	recorder := httptest.NewRecorder()

	server.handleInterpretRun(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "only failed runs can be interpreted") {
		t.Fatalf("expected non-failed rejection message, got %s", recorder.Body.String())
	}
}

func TestHandleInterpretRunRequiresEnabledSettings(t *testing.T) {
	appStore := newTestHTTPStore(t)
	server := newTestServer(t, appStore)
	run := createTestProjectAndRun(t, appStore, model.RunStatusFailed)

	recorder := httptest.NewRecorder()

	server.handleInterpretRun(recorder, newInterpretRequest(run.ID))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "ai interpretation is disabled") {
		t.Fatalf("expected disabled ai error, got %s", recorder.Body.String())
	}
}

func TestHandleInterpretRunReturnsInterpretation(t *testing.T) {
	appStore := newTestHTTPStore(t)
	server := newTestServer(t, appStore)
	run := createTestProjectAndRun(t, appStore, model.RunStatusFailed)

	openAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]string{
						"content": "失败摘要\n构建失败\n\n可能原因\n命令不存在\n\n建议操作\n补充脚本后重试",
					},
				},
			},
		})
	}))
	defer openAIServer.Close()

	if _, err := appStore.SetAISettings(context.Background(), model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolOpenAI,
		BaseURL:  openAIServer.URL + "/v1",
		APIKey:   "plain-api-key",
		Model:    "gpt-test",
	}); err != nil {
		t.Fatalf("set ai settings: %v", err)
	}

	request := newInterpretRequest(run.ID)
	recorder := httptest.NewRecorder()

	server.handleInterpretRun(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	var response model.AIInterpretationResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.RunID != run.ID {
		t.Fatalf("expected run id %d, got %d", run.ID, response.RunID)
	}
	if response.Protocol != model.AIProtocolOpenAI {
		t.Fatalf("expected protocol %q, got %q", model.AIProtocolOpenAI, response.Protocol)
	}
	if response.Model != "gpt-test" {
		t.Fatalf("expected response model gpt-test, got %q", response.Model)
	}
	if !strings.Contains(response.Content, "失败摘要") {
		t.Fatalf("expected interpretation content, got %q", response.Content)
	}
}
