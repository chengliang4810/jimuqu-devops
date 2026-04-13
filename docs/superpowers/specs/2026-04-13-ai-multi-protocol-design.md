# AI Multi-Protocol Interpretation Design

**Goal**

Extend the existing failed-run AI interpretation feature from a single OpenAI-compatible Chat Completions flow to a single active configuration that can target multiple protocols: OpenAI-compatible Chat Completions, OpenAI Responses, Anthropic Claude Messages, and Google Gemini GenerateContent.

**Current State**

- The backend stores one `ai_settings` row with `enabled`, `protocol`, `base_url`, `api_key`, and `model`.
- `POST /api/v1/runs/{runID}/interpret` only supports failed runs and currently sends one OpenAI-compatible Chat Completions request.
- The frontend settings page only exposes one protocol option and the failed-run detail dialog only cares whether AI is enabled.

**Decisions**

1. Keep a single active configuration.
   - No multi-provider profile storage.
   - No schema migration is required if we preserve the existing `protocol` column and expand its allowed values.

2. Preserve the existing stored value `openai` as the legacy and ongoing identifier for OpenAI-compatible Chat Completions.
   - UI text will describe it as "OpenAI 兼容 / Chat Completions".
   - This avoids breaking existing saved settings.

3. Add protocol values:
   - `openai`
   - `openai_responses`
   - `anthropic`
   - `gemini`

4. Keep the public API surface unchanged.
   - `GET/PUT /api/v1/settings/ai` stays the same shape.
   - `POST /api/v1/runs/{runID}/interpret` stays the same.
   - `AIInterpretationResponse` remains protocol-agnostic.

5. Reuse the same interpretation prompt across providers.
   - The backend will build one Chinese analysis prompt.
   - Each provider adapter will map that prompt into its own request payload.

**Architecture**

- Introduce a single backend dispatcher, `requestAIInterpretation(...)`, that switches on `settings.Protocol`.
- Keep provider-specific request/response translation in focused helper functions inside `internal/httpapi/ai.go`.
- Keep current storage encryption behavior untouched because `base_url`, `api_key`, and `model` already flow through the existing store layer.

**Protocol Mapping**

- `openai`
  - Endpoint: `{base_url}/chat/completions`
  - Auth: `Authorization: Bearer <api_key>`
  - Payload: `model`, `messages`, `temperature`

- `openai_responses`
  - Endpoint: `{base_url}/responses`
  - Auth: `Authorization: Bearer <api_key>`
  - Payload: `model`, `input`, `temperature`
  - Response extraction prefers `output_text`, then falls back to structured output parsing.

- `anthropic`
  - Endpoint: `{base_url}/messages`
  - Auth: `x-api-key: <api_key>`
  - Required header: `anthropic-version: 2023-06-01`
  - Payload: `model`, `system`, `messages`, `temperature`, `max_tokens`
  - Response extraction concatenates text content blocks.

- `gemini`
  - Endpoint: `{base_url}/models/{model}:generateContent`
  - Auth: `x-goog-api-key: <api_key>`
  - Payload: `contents`, `systemInstruction`, `generationConfig`
  - Response extraction concatenates text parts from the first candidate content.

**Validation**

- AI disabled: keep existing permissive behavior.
- AI enabled:
  - `protocol` must be one of the four supported values.
  - `base_url`, `api_key`, and `model` remain required for all protocols.

**Frontend Changes**

- Update `AISettings["protocol"]` to a union of the four values.
- Expand the settings page select options.
- Add protocol-specific labels and placeholders so the single-form configuration is still understandable.
- Keep the failed-run detail dialog unchanged except that it will now trigger whichever protocol is configured.

**Testing**

- Backend:
  - Add request tests for OpenAI Responses, Anthropic, and Gemini request shapes and response parsing.
  - Add normalization tests for supported protocol values.
  - Keep existing failed-run and disabled-AI coverage intact.

- Frontend:
  - Update settings tests for the expanded protocol union.
  - Add a test that switching protocol updates the submitted payload.
  - Add a test that protocol-specific helper text or placeholder changes with selection.

**Non-Goals**

- Multi-profile provider storage
- Streaming AI analysis responses
- Provider-specific advanced settings beyond the existing common fields
- Azure/OpenRouter-specific modes because OpenAI-compatible covers the common generic case
