package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestNormalizeImageSearchQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trim whitespace", input: "  node  ", want: "node"},
		{name: "empty input", input: "   ", wantErr: true},
		{name: "too short", input: "a", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeImageSearchQuery(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestFilterOfficialDockerHubImages(t *testing.T) {
	results := []dockerHubSearchResult{
		{
			Name:        "node",
			Namespace:   "library",
			Description: "node.js runtime",
			StarCount:   123,
			IsOfficial:  true,
		},
		{
			Name:        "custom-node",
			Namespace:   "me",
			Description: "custom build",
			StarCount:   10,
			IsOfficial:  false,
		},
		{
			Name:        "python",
			Namespace:   "library",
			Description: "python runtime",
			StarCount:   500,
			IsOfficial:  true,
		},
	}

	filtered := filterOfficialDockerHubImages(results)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 official images, got %d", len(filtered))
	}

	names := []string{filtered[0].Name, filtered[1].Name}
	expected := []string{"node", "python"}
	for i, want := range expected {
		if names[i] != want {
			t.Fatalf("item %d: expected %q, got %q", i, want, names[i])
		}
	}
}

func TestDockerHubSearchResultUnmarshal(t *testing.T) {
	const payload = `{"results":[{"repo_name":"node","namespace":"library","short_description":"node runtime","star_count":99,"is_official":true}]}`

	var resp dockerHubSearchResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	result := resp.Results[0]
	if result.Name != "node" {
		t.Fatalf("expected repo_name to map to Name, got %q", result.Name)
	}
	if result.Description != "node runtime" {
		t.Fatalf("expected short_description to map to Description, got %q", result.Description)
	}
	if !result.IsOfficial {
		t.Fatalf("expected official flag to be true")
	}
}

func TestSearchOfficialDockerHubImagesPagination(t *testing.T) {
	originalBaseURL := dockerHubSearchBaseURL
	defer func() { dockerHubSearchBaseURL = originalBaseURL }()

	pageCalls := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCalls++
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		if r.URL.Query().Get("page_size") != strconv.Itoa(dockerHubSearchPageSize) {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}

		switch page {
		case "1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"next":"` + server.URL + `/v2/search/repositories/?page=2&page_size=` + strconv.Itoa(dockerHubSearchPageSize) + `&query=node",
				"results":[
					{"repo_name":"node-official","namespace":"library","short_description":"node runtime","star_count":123,"is_official":true},
					{"repo_name":"node-custom","namespace":"me","short_description":"custom build","star_count":10,"is_official":false}
				]
			}`))
		case "2":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"next":"",
				"results":[
					{"repo_name":"python","namespace":"library","short_description":"python runtime","star_count":200,"is_official":true},
					{"repo_name":"redis","namespace":"library","short_description":"redis runtime","star_count":150,"is_official":true}
				]
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	dockerHubSearchBaseURL = server.URL + "/v2/search/repositories/"

	results, err := searchOfficialDockerHubImages(context.Background(), "node", 3, server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	names := []string{results[0].Name, results[1].Name, results[2].Name}
	expected := []string{"node-official", "python", "redis"}
	for i, want := range expected {
		if names[i] != want {
			t.Fatalf("result %d expected %q, got %q", i, want, names[i])
		}
	}

	if pageCalls < 2 {
		t.Fatalf("expected at least 2 page calls, got %d", pageCalls)
	}
}

func TestSearchOfficialDockerHubImagesError(t *testing.T) {
	originalBaseURL := dockerHubSearchBaseURL
	defer func() { dockerHubSearchBaseURL = originalBaseURL }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	dockerHubSearchBaseURL = server.URL + "/v2/search/repositories/"

	_, err := searchOfficialDockerHubImages(context.Background(), "node", 1, server.Client())
	if err == nil {
		t.Fatal("expected error")
	}

	var dhErr *dockerHubSearchError
	if !errors.As(err, &dhErr) {
		t.Fatalf("expected dockerHubSearchError, got %T", err)
	}
}
