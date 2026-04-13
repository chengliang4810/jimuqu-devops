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
	var capturedUserAgent string
	var capturedRequest openAIChatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAuth = r.Header.Get("Authorization")
		capturedUserAgent = r.Header.Get("User-Agent")

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
		UserAgent: "Codex Desktop/0.115.0-alpha.11 (Windows 10.0.22621; x86_64) unknown (Codex Desktop; 26.311.21342)",
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
	if capturedUserAgent != "Codex Desktop/0.115.0-alpha.11 (Windows 10.0.22621; x86_64) unknown (Codex Desktop; 26.311.21342)" {
		t.Fatalf("expected user agent header, got %q", capturedUserAgent)
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

func TestNormalizeAISettingsInputSupportsMultipleProtocols(t *testing.T) {
	testCases := []string{
		model.AIProtocolOpenAI,
		model.AIProtocolOpenAIResponses,
		model.AIProtocolAnthropic,
		model.AIProtocolGemini,
	}

	for _, protocol := range testCases {
		t.Run(protocol, func(t *testing.T) {
			normalized, err := normalizeAISettingsInput(model.AISettings{
				Enabled:  true,
				Protocol: protocol,
				BaseURL:  "https://example.com/v1",
				APIKey:   "plain-api-key",
				Model:    "demo-model",
			})
			if err != nil {
				t.Fatalf("expected protocol %q to be accepted, got error %v", protocol, err)
			}
			if normalized.Protocol != protocol {
				t.Fatalf("expected normalized protocol %q, got %q", protocol, normalized.Protocol)
			}
		})
	}
}

func TestRequestAIInterpretationOpenAIResponses(t *testing.T) {
	var capturedPath string
	var capturedAuth string
	var capturedRequest map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAuth = r.Header.Get("Authorization")

		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"output_text": "失败摘要\n部署失败\n\n可能原因\n脚本不存在\n\n建议操作\n检查脚本路径",
		})
	}))
	defer server.Close()

	content, err := requestAIInterpretation(context.Background(), server.Client(), model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolOpenAIResponses,
		BaseURL:  server.URL + "/v1",
		APIKey:   "plain-api-key",
		Model:    "gpt-5-mini",
	}, aiInterpretationInput{
		ProjectName:   "demo",
		Branch:        "main",
		Status:        model.RunStatusFailed,
		CommitMessage: "fix: build",
		ErrorMessage:  "build failed",
		LogText:       "npm ERR! missing script: build",
	})
	if err != nil {
		t.Fatalf("request ai interpretation: %v", err)
	}

	if capturedPath != "/v1/responses" {
		t.Fatalf("expected request path /v1/responses, got %q", capturedPath)
	}
	if capturedAuth != "Bearer plain-api-key" {
		t.Fatalf("expected bearer token header, got %q", capturedAuth)
	}
	if capturedRequest["model"] != "gpt-5-mini" {
		t.Fatalf("expected request model gpt-5-mini, got %#v", capturedRequest["model"])
	}
	if strings.TrimSpace(capturedRequest["input"].(string)) == "" {
		t.Fatalf("expected non-empty input prompt")
	}
	if !strings.Contains(content, "失败摘要") {
		t.Fatalf("expected returned interpretation content, got %q", content)
	}
}

func TestRequestAIInterpretationOpenAIResponsesOmitsEmptyUserAgent(t *testing.T) {
	var capturedUserAgents []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserAgents = r.Header.Values("User-Agent")

		writeJSON(w, http.StatusOK, map[string]any{
			"output_text": "失败摘要\n部署失败\n\n可能原因\n脚本不存在\n\n建议操作\n检查脚本路径",
		})
	}))
	defer server.Close()

	_, err := requestAIInterpretation(context.Background(), server.Client(), model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolOpenAIResponses,
		BaseURL:  server.URL + "/v1",
		APIKey:   "plain-api-key",
		Model:    "gpt-5-mini",
	}, aiInterpretationInput{
		ProjectName:   "demo",
		Branch:        "main",
		Status:        model.RunStatusFailed,
		ErrorMessage:  "build failed",
		LogText:       "npm ERR! missing script: build",
	})
	if err != nil {
		t.Fatalf("request ai interpretation: %v", err)
	}

	if len(capturedUserAgents) != 0 {
		t.Fatalf("expected no user-agent header when configuration is empty, got %#v", capturedUserAgents)
	}
}

func TestRequestAIInterpretationAnthropic(t *testing.T) {
	var capturedPath string
	var capturedAPIKey string
	var capturedVersion string
	var capturedRequest map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAPIKey = r.Header.Get("x-api-key")
		capturedVersion = r.Header.Get("anthropic-version")

		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": "失败摘要\n构建失败\n\n可能原因\n依赖缺失\n\n建议操作\n补充依赖"},
			},
		})
	}))
	defer server.Close()

	content, err := requestAIInterpretation(context.Background(), server.Client(), model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolAnthropic,
		BaseURL:  server.URL + "/v1",
		APIKey:   "plain-api-key",
		Model:    "claude-sonnet-4-5",
	}, aiInterpretationInput{
		ProjectName:   "demo",
		Branch:        "main",
		Status:        model.RunStatusFailed,
		CommitMessage: "fix: deploy",
		ErrorMessage:  "deploy failed",
		LogText:       "sh: npm: not found",
	})
	if err != nil {
		t.Fatalf("request ai interpretation: %v", err)
	}

	if capturedPath != "/v1/messages" {
		t.Fatalf("expected request path /v1/messages, got %q", capturedPath)
	}
	if capturedAPIKey != "plain-api-key" {
		t.Fatalf("expected anthropic api key header, got %q", capturedAPIKey)
	}
	if capturedVersion != "2023-06-01" {
		t.Fatalf("expected anthropic-version header, got %q", capturedVersion)
	}
	if capturedRequest["model"] != "claude-sonnet-4-5" {
		t.Fatalf("expected request model claude-sonnet-4-5, got %#v", capturedRequest["model"])
	}
	if strings.TrimSpace(capturedRequest["system"].(string)) == "" {
		t.Fatalf("expected non-empty system prompt")
	}
	if !strings.Contains(content, "失败摘要") {
		t.Fatalf("expected returned interpretation content, got %q", content)
	}
}

func TestRequestAIInterpretationGemini(t *testing.T) {
	var capturedPath string
	var capturedAPIKey string
	var capturedRequest map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAPIKey = r.Header.Get("x-goog-api-key")

		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{"text": "失败摘要\n构建失败\n\n可能原因\n环境变量缺失\n\n建议操作\n补齐变量"},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	content, err := requestAIInterpretation(context.Background(), server.Client(), model.AISettings{
		Enabled:  true,
		Protocol: model.AIProtocolGemini,
		BaseURL:  server.URL + "/v1beta",
		APIKey:   "plain-api-key",
		Model:    "gemini-2.5-flash",
	}, aiInterpretationInput{
		ProjectName:   "demo",
		Branch:        "main",
		Status:        model.RunStatusFailed,
		CommitMessage: "fix: ci",
		ErrorMessage:  "ci failed",
		LogText:       "pnpm: command not found",
	})
	if err != nil {
		t.Fatalf("request ai interpretation: %v", err)
	}

	if capturedPath != "/v1beta/models/gemini-2.5-flash:generateContent" {
		t.Fatalf("expected gemini request path, got %q", capturedPath)
	}
	if capturedAPIKey != "plain-api-key" {
		t.Fatalf("expected x-goog-api-key header, got %q", capturedAPIKey)
	}
	systemInstruction, ok := capturedRequest["systemInstruction"].(map[string]any)
	if !ok || systemInstruction == nil {
		t.Fatalf("expected systemInstruction object, got %#v", capturedRequest["systemInstruction"])
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
