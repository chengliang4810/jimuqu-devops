package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var dockerHubSearchBaseURL = "https://hub.docker.com/v2/search/repositories/"

const (
	defaultImageSearchLimit   = 20
	maxImageSearchLimit       = 100
	minImageSearchQueryLength = 2
	dockerHubSearchTimeout    = 10 * time.Second
	dockerHubSearchPageSize   = maxImageSearchLimit
	dockerHubSearchMaxPages   = 10
)

func normalizeImageSearchQuery(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("query is required")
	}
	if utf8.RuneCountInString(trimmed) < minImageSearchQueryLength {
		return "", errors.New("query must be at least 2 characters")
	}
	return trimmed, nil
}

func parseImageSearchLimit(r *http.Request) int {
	limit := defaultImageSearchLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			if parsed > maxImageSearchLimit {
				parsed = maxImageSearchLimit
			}
			if parsed >= 1 {
				limit = parsed
			}
		}
	}
	return limit
}

type imageSearchResponse struct {
	Items []imageSearchItem `json:"items"`
}

type imageSearchItem struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	StarCount   int    `json:"star_count"`
}

type dockerHubSearchResponse struct {
	Next    string                  `json:"next"`
	Results []dockerHubSearchResult `json:"results"`
}

type dockerHubSearchResult struct {
	Name        string `json:"repo_name"`
	Namespace   string `json:"namespace"`
	Description string `json:"short_description"`
	StarCount   int    `json:"star_count"`
	IsOfficial  bool   `json:"is_official"`
}

func filterOfficialDockerHubImages(items []dockerHubSearchResult) []dockerHubSearchResult {
	filtered := make([]dockerHubSearchResult, 0, len(items))
	for _, item := range items {
		if item.IsOfficial {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func mapToImageSearchItems(results []dockerHubSearchResult) []imageSearchItem {
	items := make([]imageSearchItem, 0, len(results))
	for _, result := range results {
		items = append(items, imageSearchItem{
			Name:        result.Name,
			DisplayName: result.Name,
			Description: result.Description,
			StarCount:   result.StarCount,
		})
	}
	return items
}

func newDockerHubHTTPClient(proxyURL string) *http.Client {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		transport = &http.Transport{}
	} else {
		transport = transport.Clone()
	}
	if proxyURL != "" {
		fixed := proxyURL
		transport.Proxy = func(*http.Request) (*url.URL, error) {
			return url.Parse(fixed)
		}
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}

	return &http.Client{
		Timeout:   dockerHubSearchTimeout,
		Transport: transport,
	}
}

type dockerHubSearchError struct {
	Err error
}

func (e *dockerHubSearchError) Error() string {
	if e.Err == nil {
		return "docker hub search failed"
	}
	return fmt.Sprintf("docker hub search failed: %v", e.Err)
}

func (e *dockerHubSearchError) Unwrap() error {
	return e.Err
}

func wrapDockerHubError(err error) error {
	return &dockerHubSearchError{Err: err}
}

func searchOfficialDockerHubImages(ctx context.Context, query string, limit int, client *http.Client) ([]dockerHubSearchResult, error) {
	if client == nil {
		client = newDockerHubHTTPClient("")
	}

	nextURL, err := buildDockerHubSearchURL(query)
	if err != nil {
		return nil, wrapDockerHubError(err)
	}

	var collected []dockerHubSearchResult
	pageCount := 0
	for nextURL != "" && len(collected) < limit && pageCount < dockerHubSearchMaxPages {
		pageCount++

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, wrapDockerHubError(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, wrapDockerHubError(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, wrapDockerHubError(fmt.Errorf("read docker hub response: %w", err))
		}

		if resp.StatusCode != http.StatusOK {
			message := strings.TrimSpace(string(body))
			if message == "" {
				message = http.StatusText(resp.StatusCode)
			}
			return nil, wrapDockerHubError(fmt.Errorf("docker hub search failed: %d %s", resp.StatusCode, message))
		}

		var parsed dockerHubSearchResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, wrapDockerHubError(fmt.Errorf("decode docker hub response: %w", err))
		}

		for _, result := range parsed.Results {
			if !result.IsOfficial {
				continue
			}
			collected = append(collected, result)
			if len(collected) >= limit {
				break
			}
		}

		nextURL = strings.TrimSpace(parsed.Next)
	}

	if len(collected) > limit {
		collected = collected[:limit]
	}

	return collected, nil
}

func buildDockerHubSearchURL(query string) (string, error) {
	u, err := url.Parse(dockerHubSearchBaseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("page", "1")
	q.Set("page_size", strconv.Itoa(dockerHubSearchPageSize))
	u.RawQuery = q.Encode()

	return u.String(), nil
}
