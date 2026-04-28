package update

import "testing"

func TestReleaseTagFromURL(t *testing.T) {
	got, err := releaseTagFromURL("https://github.com/chengliang4810/jimuqu-devops/releases/tag/v0.0.37")
	if err != nil {
		t.Fatalf("releaseTagFromURL returned error: %v", err)
	}
	if got != "v0.0.37" {
		t.Fatalf("expected v0.0.37, got %q", got)
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name      string
		candidate string
		current   string
		want      bool
	}{
		{name: "newer patch", candidate: "v0.0.37", current: "v0.0.36", want: true},
		{name: "same version", candidate: "v0.0.37", current: "0.0.37", want: false},
		{name: "older version", candidate: "v0.0.36", current: "v0.0.37", want: false},
		{name: "newer minor", candidate: "v0.1.0", current: "v0.0.99", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNewerVersion(tt.candidate, tt.current); got != tt.want {
				t.Fatalf("isNewerVersion(%q, %q) = %v, want %v", tt.candidate, tt.current, got, tt.want)
			}
		})
	}
}

func TestReleaseDownloadURL(t *testing.T) {
	got := releaseDownloadURL("chengliang4810", "jimuqu-devops", "v0.0.37", "jimuqu-devops-linux-x86_64.zip")
	want := "https://github.com/chengliang4810/jimuqu-devops/releases/download/v0.0.37/jimuqu-devops-linux-x86_64.zip"
	if got != want {
		t.Fatalf("unexpected download URL:\nwant: %s\ngot:  %s", want, got)
	}
}
