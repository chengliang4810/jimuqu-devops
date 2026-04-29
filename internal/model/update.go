package model

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ReleaseInfo struct {
	TagName     string         `json:"tag_name"`
	PublishedAt string         `json:"published_at"`
	Body        string         `json:"body"`
	HTMLURL     string         `json:"html_url"`
	Assets      []ReleaseAsset `json:"assets"`
	Message     string         `json:"message,omitempty"`
}

type UpdateStatus struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
}

type UpdateResult struct {
	Message string `json:"message"`
}
