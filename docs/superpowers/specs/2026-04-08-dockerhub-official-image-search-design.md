# Docker Hub Official Images Search Design

## Summary

Add a Docker Hub Official Images search capability to the project deployment configuration's `build_image` input.

The feature will:

- keep the existing free-form `build_image` text input behavior
- add backend-proxied search suggestions for Docker Hub Official Images only
- return and fill standard image names such as `node` instead of `library/node`
- preserve current runtime behavior where `proxy_url` affects network access and `docker_mirror_url` affects actual image pull candidates
- normalize pasted or typed image values by trimming leading/trailing spaces and line breaks before validation, search, save, and execution

This design intentionally separates:

- search source: Docker Hub Official Images via backend proxy
- execution source: existing mirror resolution plus raw image name execution logic

That separation keeps the UX simple and avoids conflating "searchable" with "pullable".

## Goals

- Help users discover valid Docker Hub Official Images while editing the build image field
- Reuse existing global proxy configuration for outbound search requests
- Keep mirror acceleration behavior unchanged for build execution
- Allow manual input of full image names and tags without forcing the user through search
- Handle copy-pasted values with extra spaces or line breaks safely and consistently

## Non-Goals

- Searching non-official Docker Hub repositories
- Searching GHCR, Quay, or other registries
- Auto-selecting a tag such as `:latest`
- Rewriting the stored image into mirror-prefixed form
- Preventing manual entry when search yields no result

## Current Context

The current project form uses a plain text input for `build_image`.

Relevant existing behavior:

- `proxy_url` is already loaded by the executor and injected as `HTTP_PROXY`, `HTTPS_PROXY`, `http_proxy`, and `https_proxy`
- `docker_mirror_url` is already used to generate multiple candidate image sources for Docker-based execution
- `build_image` is required and stored as a plain string

This means the system already has the configuration primitives needed for execution. The new work is primarily a search and input-assist feature, not an execution-pipeline redesign.

## User Experience

### Build Image Field

Replace the current plain input interaction with a searchable input pattern while preserving free-form editing.

Behavior:

- user can type any value manually
- user can paste a complete image name such as `node:20`, `golang:1.23-alpine`, or `my-registry.example.com/team/app:1.0`
- when the normalized query length is at least 2 characters, the frontend requests search suggestions
- results appear in a dropdown below the input
- selecting a result fills the input with the recommended standard image name, for example `node`
- the UI does not auto-append tags; users can continue editing to `node:20`
- if no results are found, the user can keep the current manual value

Suggested helper text:

- "仅搜索 Docker Hub 官方镜像，选择后仍可自行补充 tag，例如 `node:20`。"

### Input Normalization

The frontend should normalize values before:

- triggering remote search
- validating the form
- building the request payload

Normalization rules:

- trim leading and trailing whitespace
- replace `\r\n` with `\n` for consistent handling
- collapse outer line breaks so a pasted multi-line value like `\n node:20 \n` becomes `node:20`
- preserve internal characters that are part of the image reference
- do not silently rewrite the registry, namespace, or tag

Examples:

- `  node  ` -> `node`
- `\nnode:20\n` -> `node:20`
- `  docker.io/library/node:20  ` -> `docker.io/library/node:20`

If the normalized value becomes empty, treat it as empty input.

## API Design

Add a new authenticated backend endpoint:

- `GET /api/v1/images/search?q=<query>&limit=<n>`

Request behavior:

- `q` is required after trimming
- reject or short-circuit queries shorter than 2 characters
- `limit` is optional; backend applies a safe upper bound such as 10

Response shape:

```json
{
  "items": [
    {
      "name": "node",
      "display_name": "node",
      "description": "Node.js JavaScript runtime",
      "star_count": 12345
    }
  ]
}
```

Field intent:

- `name`: canonical value to fill into the input, such as `node`
- `display_name`: currently same as `name`, reserved for future display flexibility
- `description`: short repository summary from Docker Hub
- `star_count`: optional ranking/display aid

The API should return only the minimal fields needed by the UI.

## Backend Search Behavior

### Provider

The backend calls Docker Hub search APIs and keeps only Docker Hub Official Images.

Filtering requirement:

- include only official image results
- exclude publisher, user, and organization repositories that are not official images

The implementation can adapt to the exact Docker Hub response format, but the returned app-level contract must remain stable.

### Proxy Handling

Outbound search requests should reuse the existing global `proxy_url` setting.

Behavior:

- if `proxy_url` is configured, use it in the backend HTTP client transport
- if `proxy_url` is empty, use direct outbound access
- if proxy configuration is invalid or upstream is unreachable, return a clear application error

This keeps search behavior aligned with the deployment environment without pushing proxy concerns into the browser.

### Timeout, Rate Limit, and Caching

Recommended baseline:

- short timeout, such as 3 to 5 seconds
- simple in-memory per-query cache with a short TTL, such as 1 to 5 minutes
- optional lightweight rate limiting later if abuse becomes a concern

Caching is not required for the first implementation, but the backend contract should allow adding it without frontend changes.

## Frontend Search Behavior

### Request Rules

- debounce user input by about 300 ms
- only search when the normalized query length is at least 2
- cancel or ignore stale requests when the query changes quickly
- show a loading state in the dropdown while waiting

### Result Presentation

Each item should show:

- repository name, such as `node`
- short description when available

Sorting should prioritize backend order. The frontend should not invent its own ranking logic in the first version.

### Selection Behavior

On selection:

- fill the input with `item.name`
- close the dropdown
- keep the field editable

Do not:

- auto-append `:latest`
- convert `node` into `library/node`
- rewrite to mirror-prefixed addresses

## Validation and Persistence

Validation should use the normalized `build_image` value.

Rules:

- required after trimming
- preserve user-specified full references when they are non-empty after normalization
- keep the stored value as the standard/manual string the user entered or selected

Examples:

- selecting `node` and then editing to `node:20` stores `node:20`
- pasting `  node:20  ` stores `node:20`
- pasting only spaces and line breaks fails required validation

## Relationship With Mirror Acceleration

Search and execution must remain separate concerns.

Search:

- queries Docker Hub Official Images only
- does not depend on mirror acceleration

Execution:

- uses the stored image string
- applies existing `docker_mirror_url` candidate resolution
- still falls back to the original image string

This distinction matters because:

- a result can be searchable but not currently pullable in the user's network
- a manually entered image can be pullable even if it is not searchable through the official-image search feature

The UI should communicate this clearly enough to reduce confusion.

## Error Handling

Frontend-visible errors should be actionable and specific.

Examples:

- "官方镜像搜索失败，请检查网络代理配置或服务器外网连通性"
- "请输入至少 2 个字符再搜索"

Behavioral guidance:

- search failures must not block manual input
- form submission should still be allowed if the current normalized `build_image` is valid
- empty-result searches should not be treated as errors

## Security and Operational Notes

- do not expose Docker Hub credentials; search is anonymous unless future needs change
- sanitize query parameters before logging
- avoid storing raw upstream payloads in the database
- keep the new endpoint behind existing authenticated API protections

## Testing Strategy

### Backend

- unit tests for query normalization and minimum-length handling
- tests for official-image filtering against representative Docker Hub payloads
- tests that proxy configuration is wired into the HTTP client path
- tests for upstream failure mapping

### Frontend

- component tests for input normalization before submit
- tests for debounce and stale-request handling
- tests that selecting `node` fills `node`
- tests that whitespace-only pasted values fail validation
- tests that manual full image values remain editable and submittable

### Manual Verification

- search `no` and choose `node`, then save as `node:20`
- paste `  node:20  ` and confirm saved value is normalized
- paste a value with line breaks and confirm normalization
- disable outbound access or misconfigure proxy and confirm search shows a clear error while manual entry still works
- confirm build execution still uses mirror candidates exactly as before

## Implementation Notes

Likely touch points:

- project deployment configuration form in `web-next/src/components/modules/projects/index.tsx`
- new backend HTTP handler under `internal/httpapi`
- backend request/client logic near settings-aware services or a small dedicated Docker Hub search helper
- shared request/response types as needed

Recommended implementation order:

1. add backend search endpoint and filtering logic
2. add frontend searchable input behavior
3. normalize `build_image` consistently in form state and submit path
4. add tests for backend filtering and frontend input handling

## Open Decisions Resolved In This Spec

- Search scope: Docker Hub Official Images only
- Search path: backend proxy, not browser-direct
- Filled value format: recommended short form such as `node`
- Tag behavior: user-controlled, no auto-append
- Whitespace handling: trim leading/trailing spaces and line breaks before validation, search, save, and execution
