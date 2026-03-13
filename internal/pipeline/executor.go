package pipeline

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"devops-pipeline/internal/model"
	"devops-pipeline/internal/notification"
	"devops-pipeline/internal/store"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Executor struct {
	store         *store.Store
	logger        *slog.Logger
	workspaceRoot string
	artifactRoot  string
	httpClient    *http.Client
	cancelFuncs   map[int64]context.CancelFunc
	cancelMutex   sync.Mutex
	notifySender  *notification.Sender
}

const maxCommandLogTokenSize = 1024 * 1024

var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

type pipelineResult struct {
	Stage           string
	CommitID        string
	CommitMessage   string
	Author          string
	DurationSeconds int64
}

func NewExecutor(store *store.Store, logger *slog.Logger, workspaceRoot, artifactRoot string) *Executor {
	return &Executor{
		store:         store,
		logger:        logger,
		workspaceRoot: workspaceRoot,
		artifactRoot:  artifactRoot,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cancelFuncs:  make(map[int64]context.CancelFunc),
		notifySender: notification.New(logger),
	}
}

func (e *Executor) Trigger(ctx context.Context, projectID int64, triggerType, triggerRef string) (model.PipelineRun, error) {
	if _, err := e.store.GetExecutionBundle(ctx, projectID); err != nil {
		return model.PipelineRun{}, err
	}

	// 停止该项目的所有运行中的部署
	e.cancelRunningDeployments(ctx, projectID)

	run, err := e.store.CreateRun(ctx, model.RunCreateInput{
		ProjectID:   projectID,
		Status:      model.RunStatusQueued,
		TriggerType: triggerType,
		TriggerRef:  triggerRef,
	})
	if err != nil {
		return model.PipelineRun{}, err
	}

	// 创建可取消的context
	runCtx, cancel := context.WithCancel(context.Background())

	// 保存取消函数
	e.cancelMutex.Lock()
	e.cancelFuncs[run.ID] = cancel
	e.cancelMutex.Unlock()

	go e.execute(runCtx, run.ID, projectID, triggerType, triggerRef)
	return run, nil
}

func (e *Executor) cancelRunningDeployments(ctx context.Context, projectID int64) {
	// 获取运行中的部署，只需要获取最近100条记录
	runs, err := e.store.ListAllRuns(ctx, 0, 100)
	if err != nil {
		e.logger.Error("list runs for cancellation failed", "project_id", projectID, "error", err)
		return
	}

	for _, run := range runs {
		if run.ProjectID == projectID && run.Status == model.RunStatusRunning {
			e.logger.Info("cancelling running deployment", "run_id", run.ID, "project_id", projectID)

			// 调用取消函数
			e.cancelMutex.Lock()
			if cancel, exists := e.cancelFuncs[run.ID]; exists {
				cancel()
				delete(e.cancelFuncs, run.ID)
			}
			e.cancelMutex.Unlock()

			// 更新数据库状态
			if err := e.store.FinalizeRun(ctx, run.ID, model.RunStatusFailed, "deployment cancelled by new deployment"); err != nil {
				e.logger.Error("finalize cancelled run failed", "run_id", run.ID, "error", err)
			}

			// 记录取消日志
			logLine := fmt.Sprintf("[%s] deployment cancelled by new deployment\n", time.Now().Local().Format("2006-01-02 15:04:05"))
			if err := e.store.AppendRunLog(ctx, run.ID, logLine); err != nil {
				e.logger.Error("append cancellation log failed", "run_id", run.ID, "error", err)
			}
		}
	}
}

func (e *Executor) execute(ctx context.Context, runID, projectID int64, triggerType, triggerRef string) {
	// 清理取消函数
	defer func() {
		e.cancelMutex.Lock()
		delete(e.cancelFuncs, runID)
		e.cancelMutex.Unlock()
	}()

	if err := e.store.MarkRunRunning(ctx, runID); err != nil {
		e.logger.Error("mark run running failed", "run_id", runID, "error", err)
		return
	}

	bundle, err := e.store.GetExecutionBundle(ctx, projectID)
	if err != nil {
		_ = e.store.FinalizeRun(ctx, runID, model.RunStatusFailed, err.Error())
		e.logger.Error("load execution bundle failed", "run_id", runID, "error", err)
		return
	}

	logf := func(format string, args ...any) {
		line := fmt.Sprintf("[%s] %s\n", time.Now().Local().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
		if appendErr := e.store.AppendRunLog(ctx, runID, line); appendErr != nil {
			e.logger.Error("append run log failed", "run_id", runID, "error", appendErr)
		}
		e.logger.Info("pipeline", "run_id", runID, "message", strings.TrimSpace(fmt.Sprintf(format, args...)))
	}

	logf("pipeline start: project=%s branch=%s trigger=%s", bundle.Project.Name, bundle.Project.Branch, triggerType)
	startedAt := time.Now()

	// 设置超时context
	timeout := time.Duration(bundle.DeployConfig.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute // 默认30分钟
	}
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	// 使用带超时的context执行pipeline
	result, execErr := e.runPipeline(timeoutCtx, runID, bundle, logf)
	result.DurationSeconds = int64(time.Since(startedAt).Seconds())
	finalStatus := model.RunStatusSuccess
	finalError := ""

	if execErr != nil {
		// 检查是否是超时错误
		if timeoutCtx.Err() == context.DeadlineExceeded {
			finalStatus = model.RunStatusFailed
			finalError = fmt.Sprintf("deployment timeout after %d seconds", bundle.DeployConfig.TimeoutSeconds)
			logf("deployment timeout after %d seconds", bundle.DeployConfig.TimeoutSeconds)
		} else if ctx.Err() == context.Canceled {
			// 被手动取消，不需要记录错误
			return
		} else {
			finalStatus = model.RunStatusFailed
			finalError = execErr.Error()
			logf("pipeline failed at stage=%s: %v", displayStage(result.Stage), execErr)
		}
	} else {
		logf("pipeline finished without stage error")
	}

	notifyStatus := finalStatus
	notifyErr := e.sendNotification(ctx, bundle, runID, notifyStatus, finalError, triggerType, triggerRef, result, logf)
	if notifyErr != nil {
		logf("notification failed: %v", notifyErr)
		logf("notification failure ignored, keeping deployment status=%s", finalStatus)
	} else {
		logf("notification stage completed")
	}

	if err := e.store.FinalizeRun(ctx, runID, finalStatus, finalError); err != nil {
		e.logger.Error("finalize run failed", "run_id", runID, "error", err)
		return
	}

	logf("pipeline finalized with status=%s", finalStatus)
}

func (e *Executor) runPipeline(ctx context.Context, runID int64, bundle model.ExecutionBundle, logf func(string, ...any)) (pipelineResult, error) {
	result := pipelineResult{}
	workspaceDir := filepath.Join(e.workspaceRoot, fmt.Sprintf("run-%d", runID))
	sourceDir := filepath.Join(workspaceDir, "source")
	artifactDir := filepath.Join(e.artifactRoot, fmt.Sprintf("run-%d", runID))

	if err := os.RemoveAll(workspaceDir); err != nil {
		return result, fmt.Errorf("cleanup workspace: %w", err)
	}
	if err := os.RemoveAll(artifactDir); err != nil {
		return result, fmt.Errorf("cleanup artifact dir: %w", err)
	}
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		return result, fmt.Errorf("create source dir: %w", err)
	}
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return result, fmt.Errorf("create artifact dir: %w", err)
	}

	result.Stage = "git-clone"
	logf("stage git-clone: cloning %s#%s", bundle.Project.RepoURL, bundle.Project.Branch)
	if err := e.runGitCloneWithAuth(ctx, bundle.Project, sourceDir, logf); err != nil {
		return result, fmt.Errorf("git clone failed: %w", err)
	}

	if commitInfo, err := e.readGitMetadata(ctx, sourceDir); err == nil {
		result.CommitID = commitInfo.CommitID
		result.CommitMessage = commitInfo.CommitMessage
		result.Author = commitInfo.Author
		if result.CommitID != "" {
			logf("git metadata: commit=%s author=%s", shortCommit(result.CommitID), result.Author)
		}
	} else {
		logf("git metadata unavailable: %v", err)
	}

	result.Stage = "build"
	logf("stage build: image=%s", bundle.DeployConfig.BuildImage)
	if err := e.runDockerBuildWithLogging(ctx, sourceDir, bundle.DeployConfig.BuildImage, bundle.DeployConfig.BuildCommands, logf); err != nil {
		return result, fmt.Errorf("docker build stage failed: %w", err)
	}

	result.Stage = "artifact-filter"
	logf("stage artifact-filter: mode=%s rules=%d", bundle.DeployConfig.ArtifactFilterMode, len(bundle.DeployConfig.ArtifactRules))
	if err := filterArtifacts(sourceDir, artifactDir, bundle.DeployConfig.ArtifactFilterMode, bundle.DeployConfig.ArtifactRules); err != nil {
		return result, fmt.Errorf("filter artifacts: %w", err)
	}

	result.Stage = "deploy"
	logf("stage deploy: host=%s:%d", bundle.Host.Address, bundle.Host.Port)
	if err := e.deployToRemote(bundle, artifactDir, runID, logf); err != nil {
		return result, err
	}

	result.Stage = "completed"
	return result, nil
}

func (e *Executor) runLocalCommand(ctx context.Context, name string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	env, err := e.buildCommandEnv(ctx)
	if err != nil {
		return "", err
	}
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (e *Executor) runLocalCommandWithLogging(ctx context.Context, logf func(string, ...any), name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	env, err := e.buildCommandEnv(ctx)
	if err != nil {
		return err
	}
	cmd.Env = env

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	go e.streamCommandOutput(stdoutPipe, logf)
	go e.streamCommandOutput(stderrPipe, logf)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (e *Executor) runDockerBuild(ctx context.Context, sourceDir, image string, commands []string) (string, error) {
	script := "set -eu\n" + strings.Join(commands, "\n")
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return "", fmt.Errorf("resolve source dir: %w", err)
	}
	candidateImages, err := e.resolveBuildImages(ctx, image)
	if err != nil {
		return "", err
	}
	envArgs, err := e.dockerEnvArgs(ctx)
	if err != nil {
		return "", err
	}

	var combinedOutput strings.Builder
	for _, candidateImage := range candidateImages {
		output, runErr := e.runDockerCommand(ctx, absSourceDir, candidateImage, script, envArgs)
		if runErr == nil {
			return output, nil
		}
		if combinedOutput.Len() > 0 {
			combinedOutput.WriteString("\n")
		}
		combinedOutput.WriteString(fmt.Sprintf("image %s failed: %v", candidateImage, runErr))
		if strings.TrimSpace(output) != "" {
			combinedOutput.WriteString("\n")
			combinedOutput.WriteString(strings.TrimSpace(output))
		}
	}

	return combinedOutput.String(), fmt.Errorf("all docker mirror candidates failed")
}

func (e *Executor) runDockerBuildWithLogging(ctx context.Context, sourceDir, image string, commands []string, logf func(string, ...any)) error {
	script := "set -eu\n" + strings.Join(commands, "\n")
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("resolve source dir: %w", err)
	}
	candidateImages, err := e.resolveBuildImages(ctx, image)
	if err != nil {
		return err
	}
	envArgs, err := e.dockerEnvArgs(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, candidateImage := range candidateImages {
		logf("stage build: trying image source=%s", candidateImage)
		runErr := e.runDockerCommandWithLogging(ctx, absSourceDir, candidateImage, script, envArgs, logf)
		if runErr == nil {
			return nil
		}
		lastErr = runErr
		logf("stage build: image source failed=%s error=%v", candidateImage, runErr)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("all docker mirror candidates failed")
	}
	return lastErr
}

func (e *Executor) runDockerCommand(ctx context.Context, absSourceDir, image, script string, envArgs []string) (string, error) {
	mountDir := filepath.ToSlash(absSourceDir)
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", mountDir),
		"-w", "/workspace",
	}
	args = append(args, envArgs...)
	args = append(args, image, "sh", "-lc", script)
	return e.runLocalCommand(ctx, "docker", args)
}

func (e *Executor) runDockerCommandWithLogging(ctx context.Context, absSourceDir, image, script string, envArgs []string, logf func(string, ...any)) error {
	mountDir := filepath.ToSlash(absSourceDir)
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", mountDir),
		"-w", "/workspace",
	}
	args = append(args, envArgs...)
	args = append(args, image, "sh", "-lc", script)
	return e.runLocalCommandWithLogging(ctx, logf, "docker", args)
}

func (e *Executor) buildCommandEnv(ctx context.Context) ([]string, error) {
	env := os.Environ()

	proxyURL, err := e.store.GetSettingValue(ctx, model.SettingProxyURL)
	if err != nil {
		return nil, fmt.Errorf("load proxy setting: %w", err)
	}
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return env, nil
	}

	env = append(env,
		"HTTP_PROXY="+proxyURL,
		"HTTPS_PROXY="+proxyURL,
		"http_proxy="+proxyURL,
		"https_proxy="+proxyURL,
	)
	return env, nil
}

func (e *Executor) dockerEnvArgs(ctx context.Context) ([]string, error) {
	proxyURL, err := e.store.GetSettingValue(ctx, model.SettingProxyURL)
	if err != nil {
		return nil, fmt.Errorf("load proxy setting: %w", err)
	}
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return nil, nil
	}

	return []string{
		"-e", "HTTP_PROXY=" + proxyURL,
		"-e", "HTTPS_PROXY=" + proxyURL,
		"-e", "http_proxy=" + proxyURL,
		"-e", "https_proxy=" + proxyURL,
	}, nil
}

func (e *Executor) resolveBuildImages(ctx context.Context, image string) ([]string, error) {
	mirrorValue, err := e.store.GetSettingValue(ctx, model.SettingDockerMirrorURL)
	if err != nil {
		return nil, fmt.Errorf("load docker mirror setting: %w", err)
	}

	candidates := make([]string, 0, 4)
	seen := make(map[string]struct{})
	addCandidate := func(candidate string) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return
		}
		if _, exists := seen[candidate]; exists {
			return
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}

	for _, line := range strings.Split(strings.ReplaceAll(mirrorValue, "\r\n", "\n"), "\n") {
		mirrorURL := strings.TrimSpace(line)
		if mirrorURL == "" {
			continue
		}
		mirrorURL = strings.TrimPrefix(mirrorURL, "https://")
		mirrorURL = strings.TrimPrefix(mirrorURL, "http://")
		mirrorURL = strings.TrimSuffix(mirrorURL, "/")
		addCandidate(mirrorURL + "/" + strings.TrimPrefix(image, "/"))
	}

	addCandidate(image)

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no docker image candidates resolved")
	}

	return candidates, nil
}

func filterArtifacts(sourceDir, artifactDir, mode string, rules []string) error {
	normalizedRules := normalizeRules(rules)
	return filepath.WalkDir(sourceDir, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(sourceDir, current)
		if err != nil {
			return fmt.Errorf("build relative path: %w", err)
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}

		if rel == ".git" || strings.HasPrefix(rel, ".git/") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		switch mode {
		case model.ArtifactFilterInclude:
			if entry.IsDir() {
				if matchesRule(rel, normalizedRules) || hasDescendantRule(rel, normalizedRules) {
					return os.MkdirAll(filepath.Join(artifactDir, filepath.FromSlash(rel)), 0o755)
				}
				return filepath.SkipDir
			}
			if !matchesRule(rel, normalizedRules) {
				return nil
			}
		case model.ArtifactFilterExclude:
			if matchesRule(rel, normalizedRules) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		case "", model.ArtifactFilterNone:
		default:
			return fmt.Errorf("unsupported artifact filter mode %q", mode)
		}

		target := filepath.Join(artifactDir, filepath.FromSlash(rel))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		return copyFile(current, target)
	})
}

func (e *Executor) deployToRemote(bundle model.ExecutionBundle, artifactDir string, runID int64, logf func(string, ...any)) error {
	if err := validateRemoteDir(bundle.DeployConfig.RemoteSaveDir); err != nil {
		return fmt.Errorf("invalid remote save dir: %w", err)
	}
	if err := validateRemoteDir(bundle.DeployConfig.RemoteDeployDir); err != nil {
		return fmt.Errorf("invalid remote deploy dir: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            bundle.Host.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(bundle.Host.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	address := fmt.Sprintf("%s:%d", bundle.Host.Address, bundle.Host.Port)
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return fmt.Errorf("ssh dial failed: %w", err)
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("create sftp client: %w", err)
	}
	defer sftpClient.Close()

	saveRunDir := path.Join(
		bundle.DeployConfig.RemoteSaveDir,
		sanitizeName(bundle.Project.Name),
		fmt.Sprintf("run-%d", runID),
	)

	logf("deploy preparing remote save dir: %s", saveRunDir)
	if err = uploadDir(sftpClient, artifactDir, saveRunDir, logf); err != nil {
		return fmt.Errorf("upload artifacts: %w", err)
	}

	for _, command := range bundle.DeployConfig.PreDeployCommands {
		logf("deploy pre-command: %s", command)
		if err := runRemoteCommandWithLogging(client, command, logf); err != nil {
			return fmt.Errorf("pre-deploy command failed: %w", err)
		}
	}

	deployCommand := fmt.Sprintf(
		"mkdir -p %s && find %s -mindepth 1 -maxdepth 1 -exec rm -rf {} + && cp -a %s/. %s/",
		shellQuote(bundle.DeployConfig.RemoteDeployDir),
		shellQuote(bundle.DeployConfig.RemoteDeployDir),
		shellQuote(saveRunDir),
		shellQuote(bundle.DeployConfig.RemoteDeployDir),
	)
	logf("deploy syncing save dir to target dir")
	if err := runRemoteCommandWithLogging(client, deployCommand, logf); err != nil {
		return fmt.Errorf("deploy copy failed: %w", err)
	}

	for _, command := range bundle.DeployConfig.PostDeployCommands {
		logf("deploy post-command: %s", command)
		if err := runRemoteCommandWithLogging(client, command, logf); err != nil {
			return fmt.Errorf("post-deploy command failed: %w", err)
		}
	}

	return nil
}

func uploadDir(client *sftp.Client, localDir, remoteDir string, logf func(string, ...any)) error {
	if err := client.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("mkdir remote root: %w", err)
	}

	return filepath.WalkDir(localDir, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(localDir, current)
		if err != nil {
			return fmt.Errorf("build relative path: %w", err)
		}
		if rel == "." {
			return nil
		}

		remotePath := path.Join(remoteDir, filepath.ToSlash(rel))
		if entry.IsDir() {
			if err = client.MkdirAll(remotePath); err != nil {
				return fmt.Errorf("mkdir remote dir %s: %w", remotePath, err)
			}
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("read file info: %w", err)
		}
		if err = client.MkdirAll(path.Dir(remotePath)); err != nil {
			return fmt.Errorf("mkdir remote parent %s: %w", path.Dir(remotePath), err)
		}

		sourceFile, err := os.Open(current)
		if err != nil {
			return fmt.Errorf("open local file %s: %w", current, err)
		}
		defer sourceFile.Close()

		targetFile, err := client.Create(remotePath)
		if err != nil {
			return fmt.Errorf("create remote file %s: %w", remotePath, err)
		}

		fileSize := info.Size()
		logf("uploading: %s (%d bytes)", rel, fileSize)
		if _, err = io.Copy(targetFile, sourceFile); err != nil {
			targetFile.Close()
			return fmt.Errorf("copy to remote file %s: %w", remotePath, err)
		}
		logf("uploaded: %s (completed)", rel)

		if err = targetFile.Close(); err != nil {
			return fmt.Errorf("close remote file %s: %w", remotePath, err)
		}
		if err = client.Chmod(remotePath, info.Mode().Perm()); err != nil {
			return fmt.Errorf("chmod remote file %s: %w", remotePath, err)
		}

		return nil
	})
}

func runRemoteCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new ssh session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	return string(output), err
}

func runRemoteCommandWithLogging(client *ssh.Client, command string, logf func(string, ...any)) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new ssh session: %w", err)
	}
	defer session.Close()

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	if err := session.Start(command); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	go streamCommandOutput(stdoutPipe, logf)
	go streamCommandOutput(stderrPipe, logf)

	if err := session.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (e *Executor) sendNotification(
	ctx context.Context,
	bundle model.ExecutionBundle,
	runID int64,
	status string,
	errorMessage string,
	triggerType string,
	triggerRef string,
	result pipelineResult,
	logf func(string, ...any),
) error {
	// 构造通知载荷
	payload := model.NotificationPayload{
		RunID:           runID,
		Status:          status,
		ProjectID:       bundle.Project.ID,
		ProjectName:     bundle.Project.Name,
		RepoURL:         bundle.Project.RepoURL,
		Branch:          bundle.Project.Branch,
		TriggerType:     triggerType,
		TriggerRef:      triggerRef,
		CommitID:        result.CommitID,
		CommitMessage:   result.CommitMessage,
		Author:          result.Author,
		Stage:           result.Stage,
		HostName:        bundle.Host.Name,
		HostAddress:     bundle.Host.Address,
		RemoteDeployDir: bundle.DeployConfig.RemoteDeployDir,
		DurationSeconds: result.DurationSeconds,
		RunURL:          e.buildRunURL(ctx, runID),
		ErrorMessage:    errorMessage,
		SentAt:          time.Now().UTC().Format(time.RFC3339),
	}

	e.logger.Info("sendNotification called", "run_id", runID, "project", bundle.Project.Name, "status", status,
		"notification_channel_id", bundle.DeployConfig.NotificationChannelID)
	logf("stage notification: preparing to send %s notification", status)

	// 优先使用配置的通知渠道
	if bundle.DeployConfig.NotificationChannelID != nil {
		channel, err := e.store.GetNotificationChannel(ctx, *bundle.DeployConfig.NotificationChannelID)
		if err != nil {
			// 如果获取通知渠道失败，尝试使用默认渠道
			e.logger.Warn("failed to get configured notification channel, trying default", "channel_id", *bundle.DeployConfig.NotificationChannelID, "error", err)
			logf("notification: configured channel not found, trying default channel")
			return e.sendDefaultNotification(ctx, bundle, payload, logf)
		}

		logf("notification: sending via configured channel (type=%s, id=%d)", channel.Type, channel.ID)
		e.logger.Info("sending notification via configured channel", "channel_id", channel.ID, "channel_type", channel.Type)

		// 记录发送前的详细信息 - 从Config字段解析
		var configMap map[string]any
		if channel.Config != "" {
			if err := json.Unmarshal([]byte(channel.Config), &configMap); err == nil {
				if webhookURL, ok := configMap["webhook_url"].(string); ok {
					logf("notification: channel webhook_url=%s", webhookURL)
				} else if url, ok := configMap["url"].(string); ok {
					logf("notification: channel webhook_url=%s", url)
				}
				if secret, ok := configMap["secret"].(string); ok && secret != "" {
					logf("notification: channel has secret configured (length=%d)", len(secret))
				}
			}
		}

		// 设置日志输出函数，让通知发送器能输出到部署记录
		e.notifySender.SetLogf(logf)

		if err := e.notifySender.Send(channel, payload); err != nil {
			logf("notification: failed to send via configured channel: %v", err)
			logf("notification: error details: %s", err.Error())
			return fmt.Errorf("send notification via channel: %w", err)
		}

		logf("notification: successfully sent via channel (id=%d)", channel.ID)
		e.logger.Info("notification sent via channel", "channel_id", channel.ID, "run_id", runID)
		return nil
	}

	logf("notification: no channel configured, trying default channel")
	e.logger.Info("no notification channel configured, trying default", "run_id", runID)
	// 如果没有配置通知渠道，尝试使用系统默认渠道
	return e.sendDefaultNotification(ctx, bundle, payload, logf)
}

func (e *Executor) sendDefaultNotification(ctx context.Context, bundle model.ExecutionBundle, payload model.NotificationPayload, logf func(string, ...any)) error {
	e.logger.Info("sendDefaultNotification: trying to get default channel", "run_id", payload.RunID)
	logf("notification: looking for default notification channel")
	channel, err := e.store.GetDefaultNotificationChannel(ctx)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			e.logger.Error("failed to get default notification channel", "error", err)
			logf("notification: error getting default channel: %v", err)
		} else {
			e.logger.Info("no default notification channel found", "run_id", payload.RunID)
			logf("notification: no default channel found")
		}

		// 如果没有通知渠道，尝试使用原来的webhook方式
		return e.sendLegacyWebhook(ctx, bundle, payload, logf)
	}

	logf("notification: sending via default channel (type=%s, id=%d)", channel.Type, channel.ID)
	e.logger.Info("sending notification via default channel", "channel_id", channel.ID, "channel_type", channel.Type, "run_id", payload.RunID)

	// 记录发送前的详细信息 - 从Config字段解析
	var configMap map[string]any
	if channel.Config != "" {
		if err := json.Unmarshal([]byte(channel.Config), &configMap); err == nil {
			if webhookURL, ok := configMap["webhook_url"].(string); ok {
				logf("notification: channel webhook_url=%s", webhookURL)
			} else if url, ok := configMap["url"].(string); ok {
				logf("notification: channel webhook_url=%s", url)
			}
			if secret, ok := configMap["secret"].(string); ok && secret != "" {
				logf("notification: channel has secret configured (length=%d)", len(secret))
			}
		}
	}

	// 设置日志输出函数，让通知发送器能输出到部署记录
	e.notifySender.SetLogf(logf)

	if err := e.notifySender.Send(channel, payload); err != nil {
		logf("notification: failed to send via default channel: %v", err)
		logf("notification: error details: %s", err.Error())
		return fmt.Errorf("send notification via default channel: %w", err)
	}

	logf("notification: successfully sent via default channel (id=%d)", channel.ID)
	e.logger.Info("notification sent via default channel", "channel_id", channel.ID, "run_id", payload.RunID)
	return nil
}

func (e *Executor) sendLegacyWebhook(ctx context.Context, bundle model.ExecutionBundle, payload model.NotificationPayload, logf func(string, ...any)) error {
	webhookURL := strings.TrimSpace(bundle.DeployConfig.NotifyWebhookURL)
	if webhookURL == "" {
		e.logger.Info("sendLegacyWebhook: no webhook URL configured, skipping notification", "run_id", payload.RunID)
		logf("notification: no webhook URL configured, skipping notification")
		return nil // 没有配置webhook，直接跳过
	}

	logf("notification: sending via legacy webhook")
	e.logger.Info("sendLegacyWebhook: sending notification via legacy webhook", "run_id", payload.RunID, "webhook_url", webhookURL)

	// 记录webhook URL
	logf("notification: webhook_url=%s", webhookURL)
	if bundle.DeployConfig.NotifyBearerToken != "" {
		logf("notification: webhook has bearer token configured")
	}

	// 使用原来的webhook方式发送通知
	payloadMap := map[string]any{
		"run_id":        payload.RunID,
		"status":        payload.Status,
		"project_id":    payload.ProjectID,
		"project_name":  payload.ProjectName,
		"repo_url":      payload.RepoURL,
		"branch":        payload.Branch,
		"trigger_type":  payload.TriggerType,
		"trigger_ref":   payload.TriggerRef,
		"error_message": payload.ErrorMessage,
		"sent_at":       payload.SentAt,
	}

	body, err := json.Marshal(payloadMap)
	if err != nil {
		logf("notification: failed to marshal payload: %v", err)
		return fmt.Errorf("marshal notification payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		logf("notification: failed to create request: %v", err)
		return fmt.Errorf("build notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if bundle.DeployConfig.NotifyBearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bundle.DeployConfig.NotifyBearerToken)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		logf("notification: failed to send webhook request: %v", err)
		logf("notification: error details: %s", err.Error())
		return fmt.Errorf("send notification request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体并记录
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	logf("notification: webhook response status_code=%d body=%s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	e.logger.Info("sendLegacyWebhook: webhook response", "run_id", payload.RunID, "status_code", resp.StatusCode, "response_body", string(responseBody))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logf("notification: successfully sent via webhook")
		return nil
	}

	logf("notification: webhook returned error status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	return fmt.Errorf("notification returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
}

func copyFile(source, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("open target file: %w", err)
	}
	defer targetFile.Close()

	if _, err = io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	return nil
}

func normalizeRules(rules []string) []string {
	var normalized []string
	for _, rule := range rules {
		cleaned := strings.TrimSpace(filepath.ToSlash(rule))
		cleaned = strings.TrimPrefix(cleaned, "./")
		cleaned = path.Clean(cleaned)
		if cleaned == "." || cleaned == "" {
			continue
		}
		normalized = append(normalized, cleaned)
	}
	return normalized
}

func matchesRule(rel string, rules []string) bool {
	for _, rule := range rules {
		if rel == rule || strings.HasPrefix(rel, rule+"/") {
			return true
		}
	}
	return false
}

func hasDescendantRule(rel string, rules []string) bool {
	prefix := rel + "/"
	for _, rule := range rules {
		if strings.HasPrefix(rule, prefix) {
			return true
		}
	}
	return false
}

func validateRemoteDir(dir string) error {
	cleaned := path.Clean(strings.TrimSpace(dir))
	switch cleaned {
	case "", ".", "/":
		return fmt.Errorf("unsafe path %q", dir)
	default:
		return nil
	}
}

func wrapCommandOutput(output string, err error) error {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return err
	}
	return fmt.Errorf("%w; output: %s", err, trimmed)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\"'\"'`) + "'"
}

func sanitizeName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "-", "<", "-", ">", "-", "|", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "project"
	}
	return value
}

func (e *Executor) runGitCloneWithAuth(ctx context.Context, project model.Project, sourceDir string, logf func(string, ...any)) error {
	absWorkspaceDir, err := filepath.Abs(filepath.Dir(sourceDir))
	if err != nil {
		return fmt.Errorf("resolve git workspace dir: %w", err)
	}

	containerSourceDir := path.Join("/workspace", filepath.Base(sourceDir))
	gitImage, err := e.store.GetSettingValue(ctx, model.SettingGitDockerImage)
	if err != nil {
		return fmt.Errorf("load git docker image setting: %w", err)
	}

	candidateImages, err := e.resolveBuildImages(ctx, gitImage)
	if err != nil {
		return err
	}
	envArgs, err := e.dockerEnvArgs(ctx)
	if err != nil {
		return err
	}

	gitArgs := []string{
		"clone",
		"--depth", "1",
		"--single-branch",
		"--branch", project.Branch,
		"--progress",
		project.RepoURL,
		containerSourceDir,
	}
	extraArgs := []string{}

	switch project.GitAuthType {
	case "", model.GitAuthTypeNone:
	case model.GitAuthTypeUsername, model.GitAuthTypeToken:
		if project.GitUsername == "" || project.GitPassword == "" {
			return fmt.Errorf("git username/password is required for %s authentication", project.GitAuthType)
		}
		authURL := e.constructAuthenticatedURL(project.RepoURL, project.GitUsername, project.GitPassword)
		gitArgs[len(gitArgs)-2] = authURL
	case model.GitAuthTypeSSH:
		if project.GitSSHKey == "" {
			return fmt.Errorf("ssh key is required for ssh authentication")
		}

		sshKeyFile, err := e.createTempSSHKeyFile(project.GitSSHKey)
		if err != nil {
			return fmt.Errorf("create temporary ssh key file: %w", err)
		}
		defer os.Remove(sshKeyFile)

		containerKeyPath := "/tmp/git_ssh_key"
		extraArgs = append(extraArgs, "-v", fmt.Sprintf("%s:%s:ro", filepath.ToSlash(sshKeyFile), containerKeyPath))
		gitArgs = append([]string{
			"-c",
			fmt.Sprintf("core.sshCommand=ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", containerKeyPath),
		}, gitArgs...)
	default:
		return fmt.Errorf("unsupported git authentication type: %s", project.GitAuthType)
	}

	var lastErr error
	for _, candidateImage := range candidateImages {
		logf("stage git-clone: trying image source=%s", candidateImage)
		runErr := e.runDockerGitCommandWithLogging(ctx, absWorkspaceDir, candidateImage, gitArgs, envArgs, extraArgs, logf)
		if runErr == nil {
			return nil
		}
		lastErr = runErr
		logf("stage git-clone: image source failed=%s error=%v", candidateImage, runErr)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("all docker mirror candidates failed")
	}
	return lastErr
}

func (e *Executor) constructAuthenticatedURL(repoURL, username, password string) string {
	// 构造带认证的URL，对密码中的特殊字符进行URL编码
	// 格式：https://username:encoded_password@repo-url

	// 解析URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		// 如果URL解析失败，使用简单替换作为后备
		return strings.Replace(repoURL, "://", fmt.Sprintf("://%s:%s@", username, url.QueryEscape(password)), 1)
	}

	// 设置用户信息
	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String()
}

func (e *Executor) createTempSSHKeyFile(privateKey string) (string, error) {
	// 创建临时目录
	tempDir := os.TempDir()
	keyFile := filepath.Join(tempDir, fmt.Sprintf("git_key_%d", time.Now().UnixNano()))

	// 写入SSH私钥
	if err := os.WriteFile(keyFile, []byte(privateKey), 0o600); err != nil {
		return "", fmt.Errorf("write ssh key file: %w", err)
	}

	return keyFile, nil
}

func (e *Executor) runDockerGitCommandWithLogging(ctx context.Context, absWorkspaceDir, image string, gitArgs, envArgs, extraArgs []string, logf func(string, ...any)) error {
	mountDir := filepath.ToSlash(absWorkspaceDir)
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", mountDir),
		"-w", "/workspace",
		"--entrypoint", "git",
	}
	args = append(args, extraArgs...)
	args = append(args, envArgs...)
	args = append(args, image)
	args = append(args, gitArgs...)
	return e.runLocalCommandWithLogging(ctx, logf, "docker", args)
}

type gitMetadata struct {
	CommitID      string
	CommitMessage string
	Author        string
}

func (e *Executor) readGitMetadata(ctx context.Context, sourceDir string) (gitMetadata, error) {
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return gitMetadata{}, fmt.Errorf("resolve source dir: %w", err)
	}

	gitImage, err := e.store.GetSettingValue(ctx, model.SettingGitDockerImage)
	if err != nil {
		return gitMetadata{}, fmt.Errorf("load git docker image setting: %w", err)
	}

	candidateImages, err := e.resolveBuildImages(ctx, gitImage)
	if err != nil {
		return gitMetadata{}, err
	}
	envArgs, err := e.dockerEnvArgs(ctx)
	if err != nil {
		return gitMetadata{}, err
	}

	format := "%H%n%s%n%an"
	var lastErr error
	for _, candidateImage := range candidateImages {
		output, runErr := e.runDockerGitReadOnly(ctx, absSourceDir, candidateImage, []string{"log", "-1", "--pretty=format:" + format}, envArgs)
		if runErr == nil {
			parts := strings.SplitN(strings.TrimSpace(output), "\n", 3)
			if len(parts) < 3 {
				return gitMetadata{}, fmt.Errorf("unexpected git log output")
			}
			return gitMetadata{
				CommitID:      strings.TrimSpace(parts[0]),
				CommitMessage: strings.TrimSpace(parts[1]),
				Author:        strings.TrimSpace(parts[2]),
			}, nil
		}
		lastErr = runErr
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no docker image candidates resolved")
	}
	return gitMetadata{}, lastErr
}

func (e *Executor) runDockerGitReadOnly(ctx context.Context, absSourceDir, image string, gitArgs, envArgs []string) (string, error) {
	mountDir := filepath.ToSlash(absSourceDir)
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", mountDir),
		"-w", "/workspace",
		"--entrypoint", "git",
	}
	args = append(args, envArgs...)
	args = append(args, image)
	args = append(args, gitArgs...)
	return e.runLocalCommand(ctx, "docker", args)
}

func shortCommit(commitID string) string {
	if len(commitID) <= 12 {
		return commitID
	}
	return commitID[:12]
}

func displayStage(stage string) string {
	switch strings.TrimSpace(stage) {
	case "", "completed":
		return "completed"
	default:
		return stage
	}
}

func (e *Executor) buildRunURL(ctx context.Context, runID int64) string {
	baseURL, err := e.store.GetSettingValue(ctx, model.SettingPublicBaseURL)
	if err != nil {
		return ""
	}

	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}

	baseURL = strings.TrimRight(baseURL, "/")
	return fmt.Sprintf("%s/?view=logs&run_id=%d", baseURL, runID)
}

func (e *Executor) streamCommandOutput(reader io.Reader, logf func(string, ...any)) {
	streamCommandOutput(reader, logf)
}

func streamCommandOutput(reader io.Reader, logf func(string, ...any)) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), maxCommandLogTokenSize)
	scanner.Split(scanCommandLines)
	for scanner.Scan() {
		line := sanitizeCommandLogLine(scanner.Text())
		if line != "" {
			logf("%s", line)
		}
	}
}

func scanCommandLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	for index, value := range data {
		if value == '\n' || value == '\r' {
			return index + 1, data[:index], nil
		}
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func sanitizeCommandLogLine(line string) string {
	line = strings.TrimSpace(strings.ReplaceAll(line, "\x00", ""))
	if line == "" {
		return ""
	}

	line = ansiEscapePattern.ReplaceAllString(line, "")
	line = strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return r
		case r == utf8.RuneError:
			return -1
		case unicode.IsControl(r):
			return -1
		default:
			return r
		}
	}, line)

	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	return line
}
