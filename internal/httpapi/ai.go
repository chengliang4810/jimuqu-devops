package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"devops-pipeline/internal/model"
)

const (
	aiLogCharacterLimit = 12000
	aiRequestTimeout    = 30 * time.Second
)

type aiInterpretationInput struct {
	ProjectName    string
	Branch         string
	Status         string
	CommitMessage  string
	ErrorMessage   string
	LogText        string
	LogTruncated   bool
}

type openAIChatCompletionRequest struct {
	Model       string                      `json:"model"`
	Messages    []openAIChatCompletionMessage `json:"messages"`
	Temperature float64                     `json:"temperature,omitempty"`
}

type openAIChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatCompletionResponse struct {
	Choices []struct {
		Message openAIChatCompletionMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type openAIResponsesRequest struct {
	Model       string  `json:"model"`
	Input       string  `json:"input"`
	Temperature float64 `json:"temperature,omitempty"`
}

type openAIResponsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type anthropicMessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessagesRequest struct {
	Model       string                    `json:"model"`
	System      string                    `json:"system"`
	Messages    []anthropicMessageRequest `json:"messages"`
	Temperature float64                   `json:"temperature,omitempty"`
	MaxTokens   int                       `json:"max_tokens"`
}

type anthropicMessagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature float64 `json:"temperature,omitempty"`
}

type geminiGenerateContentRequest struct {
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	Contents          []geminiContent         `json:"contents"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *Server) handleGetAISettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.GetAISettings(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleGetAISettingsStatus(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.GetAISettings(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, model.AISettingsStatus{Enabled: settings.Enabled})
}

func (s *Server) handleUpdateAISettings(w http.ResponseWriter, r *http.Request) {
	var input model.AISettings
	if err := decodeJSON(r.Body, &input); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	normalized, err := normalizeAISettingsInput(input)
	if err != nil {
		s.writeBadRequest(w, err)
		return
	}

	settings, err := s.store.SetAISettings(r.Context(), normalized)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleInterpretRun(w http.ResponseWriter, r *http.Request) {
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
	if run.Status != model.RunStatusFailed {
		s.writeBadRequest(w, errors.New("only failed runs can be interpreted"))
		return
	}

	settings, err := s.store.GetAISettings(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !settings.Enabled {
		s.writeBadRequest(w, errors.New("ai interpretation is disabled"))
		return
	}
	if _, err := normalizeAISettingsInput(settings); err != nil {
		s.writeBadRequest(w, err)
		return
	}

	logText := run.LogText
	if strings.TrimSpace(logText) == "" {
		runLog, err := s.store.GetRunLog(r.Context(), runID)
		if err == nil {
			logText = runLog.LogText
		}
	}
	logText, logTruncated := truncateAILog(logText)

	content, err := requestAIInterpretation(
		r.Context(),
		newAIHTTPClient(s.getProxyURL(r)),
		settings,
		aiInterpretationInput{
			ProjectName:   run.ProjectName,
			Branch:        run.Branch,
			Status:        run.Status,
			CommitMessage: run.CommitMessage,
			ErrorMessage:  run.ErrorMessage,
			LogText:       logText,
			LogTruncated:  logTruncated,
		},
	)
	if err != nil {
		s.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, model.AIInterpretationResponse{
		RunID:        run.ID,
		Protocol:     settings.Protocol,
		Model:        settings.Model,
		Content:      content,
		LogTruncated: logTruncated,
	})
}

func normalizeAISettingsInput(input model.AISettings) (model.AISettings, error) {
	input.Protocol = strings.TrimSpace(input.Protocol)
	input.BaseURL = strings.TrimSpace(input.BaseURL)
	input.APIKey = strings.TrimSpace(input.APIKey)
	input.Model = strings.TrimSpace(input.Model)

	if input.Protocol == "" {
		input.Protocol = model.AIProtocolOpenAI
	}
	switch input.Protocol {
	case model.AIProtocolOpenAI, model.AIProtocolOpenAIResponses, model.AIProtocolAnthropic, model.AIProtocolGemini:
	default:
		return model.AISettings{}, errors.New("unsupported ai protocol")
	}

	if !input.Enabled {
		return input, nil
	}
	if input.BaseURL == "" {
		return model.AISettings{}, errors.New("base_url is required when ai interpretation is enabled")
	}
	if input.APIKey == "" {
		return model.AISettings{}, errors.New("api_key is required when ai interpretation is enabled")
	}
	if input.Model == "" {
		return model.AISettings{}, errors.New("model is required when ai interpretation is enabled")
	}

	return input, nil
}

func truncateAILog(logText string) (string, bool) {
	runes := []rune(logText)
	if len(runes) <= aiLogCharacterLimit {
		return logText, false
	}
	return string(runes[len(runes)-aiLogCharacterLimit:]), true
}

func requestAIInterpretation(ctx context.Context, client *http.Client, settings model.AISettings, input aiInterpretationInput) (string, error) {
	switch settings.Protocol {
	case "", model.AIProtocolOpenAI:
		return requestOpenAIInterpretation(ctx, client, settings, input)
	case model.AIProtocolOpenAIResponses:
		return requestOpenAIResponsesInterpretation(ctx, client, settings, input)
	case model.AIProtocolAnthropic:
		return requestAnthropicInterpretation(ctx, client, settings, input)
	case model.AIProtocolGemini:
		return requestGeminiInterpretation(ctx, client, settings, input)
	default:
		return "", errors.New("unsupported ai protocol")
	}
}

func requestOpenAIInterpretation(ctx context.Context, client *http.Client, settings model.AISettings, input aiInterpretationInput) (string, error) {
	if client == nil {
		client = newAIHTTPClient("")
	}

	endpoint := strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/") + "/chat/completions"
	body, err := json.Marshal(openAIChatCompletionRequest{
		Model: settings.Model,
		Messages: []openAIChatCompletionMessage{
			{
				Role: "system",
				Content: "你是一个 DevOps 部署故障分析助手。请始终使用中文回答，并严格输出以下三个小节：失败摘要、可能原因、建议操作。",
			},
			{
				Role:    "user",
				Content: buildAIInterpretationPrompt(input),
			},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, aiRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create openai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+settings.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request ai interpretation: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read ai response: %w", err)
	}

	var parsed openAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode ai response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			return "", fmt.Errorf("ai response failed: %s", parsed.Error.Message)
		}
		return "", fmt.Errorf("ai response failed with status %d", resp.StatusCode)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", errors.New("ai response did not include interpretation content")
	}

	return parsed.Choices[0].Message.Content, nil
}

func requestOpenAIResponsesInterpretation(ctx context.Context, client *http.Client, settings model.AISettings, input aiInterpretationInput) (string, error) {
	responseBody, err := doAIJSONRequest(
		ctx,
		client,
		strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")+"/responses",
		map[string]string{
			"Authorization": "Bearer " + settings.APIKey,
		},
		openAIResponsesRequest{
			Model:       settings.Model,
			Input:       buildAIInterpretationPrompt(input),
			Temperature: 0.2,
		},
	)
	if err != nil {
		return "", err
	}

	var parsed openAIResponsesResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return "", fmt.Errorf("decode ai response: %w", err)
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return "", fmt.Errorf("ai response failed: %s", parsed.Error.Message)
	}
	if strings.TrimSpace(parsed.OutputText) != "" {
		return parsed.OutputText, nil
	}

	var contentParts []string
	for _, outputItem := range parsed.Output {
		for _, content := range outputItem.Content {
			if strings.TrimSpace(content.Text) != "" {
				contentParts = append(contentParts, content.Text)
			}
		}
	}
	if len(contentParts) == 0 {
		return "", errors.New("ai response did not include interpretation content")
	}
	return strings.Join(contentParts, "\n"), nil
}

func requestAnthropicInterpretation(ctx context.Context, client *http.Client, settings model.AISettings, input aiInterpretationInput) (string, error) {
	responseBody, err := doAIJSONRequest(
		ctx,
		client,
		strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")+"/messages",
		map[string]string{
			"x-api-key":         settings.APIKey,
			"anthropic-version": "2023-06-01",
		},
		anthropicMessagesRequest{
			Model:  settings.Model,
			System: "你是一个 DevOps 部署故障分析助手。请始终使用中文回答，并严格输出以下三个小节：失败摘要、可能原因、建议操作。",
			Messages: []anthropicMessageRequest{
				{
					Role:    "user",
					Content: buildAIInterpretationPrompt(input),
				},
			},
			Temperature: 0.2,
			MaxTokens:   1200,
		},
	)
	if err != nil {
		return "", err
	}

	var parsed anthropicMessagesResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return "", fmt.Errorf("decode ai response: %w", err)
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return "", fmt.Errorf("ai response failed: %s", parsed.Error.Message)
	}

	var contentParts []string
	for _, content := range parsed.Content {
		if strings.TrimSpace(content.Text) != "" {
			contentParts = append(contentParts, content.Text)
		}
	}
	if len(contentParts) == 0 {
		return "", errors.New("ai response did not include interpretation content")
	}
	return strings.Join(contentParts, "\n"), nil
}

func requestGeminiInterpretation(ctx context.Context, client *http.Client, settings model.AISettings, input aiInterpretationInput) (string, error) {
	responseBody, err := doAIJSONRequest(
		ctx,
		client,
		strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")+"/models/"+url.PathEscape(settings.Model)+":generateContent",
		map[string]string{
			"x-goog-api-key": settings.APIKey,
		},
		geminiGenerateContentRequest{
			SystemInstruction: &geminiContent{
				Parts: []geminiPart{
					{
						Text: "你是一个 DevOps 部署故障分析助手。请始终使用中文回答，并严格输出以下三个小节：失败摘要、可能原因、建议操作。",
					},
				},
			},
			Contents: []geminiContent{
				{
					Role: "user",
					Parts: []geminiPart{
						{
							Text: buildAIInterpretationPrompt(input),
						},
					},
				},
			},
			GenerationConfig: &geminiGenerationConfig{
				Temperature: 0.2,
			},
		},
	)
	if err != nil {
		return "", err
	}

	var parsed geminiGenerateContentResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return "", fmt.Errorf("decode ai response: %w", err)
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return "", fmt.Errorf("ai response failed: %s", parsed.Error.Message)
	}
	if len(parsed.Candidates) == 0 {
		return "", errors.New("ai response did not include interpretation content")
	}

	var contentParts []string
	for _, part := range parsed.Candidates[0].Content.Parts {
		if strings.TrimSpace(part.Text) != "" {
			contentParts = append(contentParts, part.Text)
		}
	}
	if len(contentParts) == 0 {
		return "", errors.New("ai response did not include interpretation content")
	}
	return strings.Join(contentParts, "\n"), nil
}

func doAIJSONRequest(ctx context.Context, client *http.Client, endpoint string, headers map[string]string, requestBody any) ([]byte, error) {
	if client == nil {
		client = newAIHTTPClient("")
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal ai request: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, aiRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create ai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request ai interpretation: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ai response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("ai response failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return responseBody, nil
}

func buildAIInterpretationPrompt(input aiInterpretationInput) string {
	var builder strings.Builder
	builder.WriteString("请分析以下部署失败日志，并输出“失败摘要 / 可能原因 / 建议操作”三个部分。\n\n")
	builder.WriteString(fmt.Sprintf("项目: %s\n", emptyFallback(input.ProjectName, "-")))
	builder.WriteString(fmt.Sprintf("分支: %s\n", emptyFallback(input.Branch, "-")))
	builder.WriteString(fmt.Sprintf("状态: %s\n", emptyFallback(input.Status, "-")))
	builder.WriteString(fmt.Sprintf("提交信息: %s\n", emptyFallback(input.CommitMessage, "-")))
	builder.WriteString(fmt.Sprintf("错误信息: %s\n", emptyFallback(input.ErrorMessage, "-")))
	if input.LogTruncated {
		builder.WriteString("日志说明: 已截取最近 12000 个字符\n")
	}
	builder.WriteString("\n日志内容:\n")
	builder.WriteString(emptyFallback(input.LogText, "暂无日志"))
	return builder.String()
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func newAIHTTPClient(proxyURL string) *http.Client {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		transport = &http.Transport{}
	} else {
		transport = transport.Clone()
	}

	transport.DialContext = (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	transport.ForceAttemptHTTP2 = true
	transport.MaxIdleConns = 20
	transport.IdleConnTimeout = 90 * time.Second
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.ExpectContinueTimeout = 1 * time.Second

	if strings.TrimSpace(proxyURL) != "" {
		fixed := strings.TrimSpace(proxyURL)
		transport.Proxy = func(*http.Request) (*url.URL, error) {
			return url.Parse(fixed)
		}
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}

	return &http.Client{
		Timeout:   aiRequestTimeout,
		Transport: transport,
	}
}
