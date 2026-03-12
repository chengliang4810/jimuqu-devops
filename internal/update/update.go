package update

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"devops-pipeline/internal/model"
	"devops-pipeline/internal/version"
)

const requestTimeout = 30 * time.Second

func GetLatestRelease(ctx context.Context, proxyURL string) (model.ReleaseInfo, error) {
	apiURL, err := latestReleaseAPIURL()
	if err != nil {
		return model.ReleaseInfo{}, err
	}

	body, err := doRequest(ctx, apiURL, proxyURL)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), `"status":"404"`) || strings.Contains(err.Error(), `"status": "404"`) {
			return model.ReleaseInfo{Message: "no published release found"}, nil
		}
		return model.ReleaseInfo{}, err
	}

	var release model.ReleaseInfo
	if err := json.Unmarshal(body, &release); err != nil {
		return model.ReleaseInfo{}, fmt.Errorf("decode latest release: %w", err)
	}
	if release.Message != "" {
		return model.ReleaseInfo{}, fmt.Errorf("get latest release failed: %s", release.Message)
	}
	return release, nil
}

func GetUpdateStatus(ctx context.Context, proxyURL string) (model.UpdateStatus, error) {
	release, err := GetLatestRelease(ctx, proxyURL)
	if err != nil {
		return model.UpdateStatus{}, err
	}

	currentVersion := version.Current()
	return model.UpdateStatus{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		HasUpdate:      currentVersion != "" && currentVersion != "dev" && release.TagName != "" && currentVersion != release.TagName,
	}, nil
}

func ApplyUpdate(ctx context.Context, proxyURL string) (model.UpdateResult, error) {
	if version.Current() == "dev" {
		return model.UpdateResult{}, fmt.Errorf("online update requires a packaged release build")
	}

	release, err := GetLatestRelease(ctx, proxyURL)
	if err != nil {
		return model.UpdateResult{}, err
	}
	if release.Message != "" {
		return model.UpdateResult{}, fmt.Errorf("no published release found")
	}

	assetName, err := releaseAssetName()
	if err != nil {
		return model.UpdateResult{}, err
	}

	downloadURL := ""
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return model.UpdateResult{}, fmt.Errorf("release asset not found: %s", assetName)
	}

	archiveData, err := doRequest(ctx, downloadURL, proxyURL)
	if err != nil {
		return model.UpdateResult{}, fmt.Errorf("download release asset: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return model.UpdateResult{}, fmt.Errorf("resolve executable path: %w", err)
	}
	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return model.UpdateResult{}, fmt.Errorf("resolve absolute executable path: %w", err)
	}
	targetDir := filepath.Dir(execPath)

	if runtime.GOOS == "windows" {
		if err := applyWindowsUpdate(archiveData, targetDir, execPath); err != nil {
			return model.UpdateResult{}, err
		}
		return model.UpdateResult{Message: "update downloaded, application will restart shortly"}, nil
	}

	if err := applyUnixUpdate(archiveData, targetDir, execPath); err != nil {
		return model.UpdateResult{}, err
	}

	return model.UpdateResult{Message: "update applied, application will restart shortly"}, nil
}

func ScheduleRestartAndExit() {
	go func() {
		time.Sleep(1200 * time.Millisecond)
		os.Exit(0)
	}()
}

func doRequest(ctx context.Context, targetURL, proxyURL string) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", version.AppName)

	client, err := buildHTTPClient(proxyURL)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", targetURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("request %s failed: %s", targetURL, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func buildHTTPClient(proxyURL string) (*http.Client, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if strings.TrimSpace(proxyURL) != "" {
		parsed, err := url.Parse(strings.TrimSpace(proxyURL))
		if err != nil {
			return nil, fmt.Errorf("parse proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsed)
	}

	return &http.Client{Transport: transport}, nil
}

func latestReleaseAPIURL() (string, error) {
	owner, repo, err := parseRepoOwnerRepo(version.RepoURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo), nil
}

func parseRepoOwnerRepo(repoURL string) (string, string, error) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(repoURL), ".git")
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("parse repo url: %w", err)
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) < 2 {
		return "", "", fmt.Errorf("invalid repo url: %s", repoURL)
	}
	return segments[0], segments[1], nil
}

func releaseAssetName() (string, error) {
	switch runtime.GOOS {
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			return version.AppName + "-windows-amd64.zip", nil
		case "386":
			return version.AppName + "-windows-386.zip", nil
		case "arm64":
			return version.AppName + "-windows-arm64.zip", nil
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return version.AppName + "-linux-amd64.zip", nil
		case "386":
			return version.AppName + "-linux-386.zip", nil
		case "arm64":
			return version.AppName + "-linux-arm64.zip", nil
		case "arm":
			return version.AppName + "-linux-armv7.zip", nil
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return version.AppName + "-darwin-amd64.zip", nil
		case "arm64":
			return version.AppName + "-darwin-arm64.zip", nil
		}
	}

	return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func applyUnixUpdate(archiveData []byte, targetDir, execPath string) error {
	stagingDir, err := os.MkdirTemp(targetDir, ".update-*")
	if err != nil {
		return fmt.Errorf("create update staging dir: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	if err := unzipToDir(archiveData, stagingDir); err != nil {
		return err
	}
	if err := copyDirContents(stagingDir, targetDir, execPath); err != nil {
		return err
	}

	return startDetachedUnix(execPath, os.Args[1:])
}

func applyWindowsUpdate(archiveData []byte, targetDir, execPath string) error {
	stagingDir, err := os.MkdirTemp(targetDir, "update-*")
	if err != nil {
		return fmt.Errorf("create windows update staging dir: %w", err)
	}

	if err := unzipToDir(archiveData, stagingDir); err != nil {
		return err
	}

	return startDetachedWindowsUpdater(stagingDir, targetDir, execPath, os.Args[1:])
}

func unzipToDir(data []byte, dest string) error {
	zipReader, err := zip.NewReader(bytesReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}

	for _, file := range zipReader.File {
		targetPath := filepath.Join(dest, file.Name)
		if !isPathInDestination(targetPath, dest) {
			return fmt.Errorf("invalid archive entry path: %s", file.Name)
		}

		info := file.FileInfo()
		if info.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("create directory %s: %w", targetPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("create parent directory %s: %w", targetPath, err)
		}

		source, err := file.Open()
		if err != nil {
			return fmt.Errorf("open archive file %s: %w", file.Name, err)
		}

		mode := info.Mode().Perm()
		if mode == 0 {
			mode = 0o644
		}
		target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			source.Close()
			return fmt.Errorf("create target file %s: %w", targetPath, err)
		}

		if _, err := io.Copy(target, source); err != nil {
			target.Close()
			source.Close()
			return fmt.Errorf("extract file %s: %w", file.Name, err)
		}

		target.Close()
		source.Close()
	}

	return nil
}

func copyDirContents(sourceDir, targetDir, execPath string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == sourceDir {
			return nil
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, relativePath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		if targetPath == execPath {
			tempExecPath := execPath + ".new"
			if err := copyFile(path, tempExecPath, info.Mode()); err != nil {
				return err
			}
			return os.Rename(tempExecPath, execPath)
		}

		return copyFile(path, targetPath, info.Mode())
	})
}

func copyFile(sourcePath, targetPath string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create target dir %s: %w", targetPath, err)
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open source file %s: %w", sourcePath, err)
	}
	defer source.Close()

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return fmt.Errorf("open target file %s: %w", targetPath, err)
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", sourcePath, targetPath, err)
	}
	return nil
}

func startDetachedUnix(execPath string, args []string) error {
	quotedArgs := make([]string, 0, len(args)+1)
	quotedArgs = append(quotedArgs, shellQuote(execPath))
	for _, arg := range args {
		quotedArgs = append(quotedArgs, shellQuote(arg))
	}

	command := exec.Command("sh", "-c", fmt.Sprintf("sleep 1; exec %s >/dev/null 2>&1 &", strings.Join(quotedArgs, " ")))
	if err := command.Start(); err != nil {
		return fmt.Errorf("schedule restart: %w", err)
	}
	return nil
}

func startDetachedWindowsUpdater(stagingDir, targetDir, execPath string, args []string) error {
	updaterPath := filepath.Join(stagingDir, "update.bat")
	quotedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		quotedArgs = append(quotedArgs, quoteWindowsArg(arg))
	}

	script := strings.Join([]string{
		"@echo off",
		"setlocal",
		"timeout /t 2 /nobreak >nul",
		fmt.Sprintf("xcopy /E /I /Y %s %s >nul", quoteWindowsPath(filepath.Join(stagingDir, "*")), quoteWindowsPath(targetDir)),
		fmt.Sprintf("start \"\" %s %s", quoteWindowsPath(execPath), strings.Join(quotedArgs, " ")),
		"endlocal",
	}, "\r\n")

	if err := os.WriteFile(updaterPath, []byte(script), 0o700); err != nil {
		return fmt.Errorf("write updater script: %w", err)
	}

	command := exec.Command("cmd", "/C", "start", "", updaterPath)
	if err := command.Start(); err != nil {
		return fmt.Errorf("schedule windows updater: %w", err)
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func quoteWindowsArg(value string) string {
	if value == "" {
		return `""`
	}
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

func quoteWindowsPath(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func isPathInDestination(path, destination string) bool {
	relativePath, err := filepath.Rel(destination, path)
	if err != nil {
		return false
	}
	return filepath.IsLocal(relativePath)
}

type byteReaderAt struct {
	data []byte
}

func bytesReader(data []byte) *byteReaderAt {
	return &byteReaderAt{data: data}
}

func (reader *byteReaderAt) ReadAt(buffer []byte, offset int64) (int, error) {
	if offset >= int64(len(reader.data)) {
		return 0, io.EOF
	}
	n := copy(buffer, reader.data[offset:])
	if n < len(buffer) {
		return n, io.EOF
	}
	return n, nil
}
