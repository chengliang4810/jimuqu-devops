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
	"strings"
	"sync"
	"time"

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

	// 设置超时context
	timeout := time.Duration(bundle.DeployConfig.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute // 默认30分钟
	}
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	// 使用带超时的context执行pipeline
	execErr := e.runPipeline(timeoutCtx, runID, bundle, logf)
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
			logf("pipeline failed: %v", execErr)
		}
	} else {
		logf("pipeline finished without stage error")
	}

	notifyStatus := finalStatus
	notifyErr := e.sendNotification(ctx, bundle, runID, notifyStatus, finalError, triggerType, triggerRef)
	if notifyErr != nil {
		logf("notification failed: %v", notifyErr)
		if finalError == "" {
			finalError = notifyErr.Error()
		} else {
			finalError = finalError + "; notification: " + notifyErr.Error()
		}
		finalStatus = model.RunStatusFailed
	} else {
		logf("notification stage completed")
	}

	if err := e.store.FinalizeRun(ctx, runID, finalStatus, finalError); err != nil {
		e.logger.Error("finalize run failed", "run_id", runID, "error", err)
		return
	}

	logf("pipeline finalized with status=%s", finalStatus)
}

func (e *Executor) runPipeline(ctx context.Context, runID int64, bundle model.ExecutionBundle, logf func(string, ...any)) error {
	workspaceDir := filepath.Join(e.workspaceRoot, fmt.Sprintf("run-%d", runID))
	sourceDir := filepath.Join(workspaceDir, "source")
	artifactDir := filepath.Join(e.artifactRoot, fmt.Sprintf("run-%d", runID))

	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("cleanup workspace: %w", err)
	}
	if err := os.RemoveAll(artifactDir); err != nil {
		return fmt.Errorf("cleanup artifact dir: %w", err)
	}
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		return fmt.Errorf("create source dir: %w", err)
	}
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return fmt.Errorf("create artifact dir: %w", err)
	}

	logf("stage git-clone: cloning %s#%s", bundle.Project.RepoURL, bundle.Project.Branch)
	if err := e.runGitCloneWithAuth(ctx, bundle.Project, sourceDir, logf); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	logf("stage build: image=%s", bundle.DeployConfig.BuildImage)
	if err := e.runDockerBuildWithLogging(ctx, sourceDir, bundle.DeployConfig.BuildImage, bundle.DeployConfig.BuildCommands, logf); err != nil {
		return fmt.Errorf("docker build stage failed: %w", err)
	}

	logf("stage artifact-filter: mode=%s rules=%d", bundle.DeployConfig.ArtifactFilterMode, len(bundle.DeployConfig.ArtifactRules))
	if err := filterArtifacts(sourceDir, artifactDir, bundle.DeployConfig.ArtifactFilterMode, bundle.DeployConfig.ArtifactRules); err != nil {
		return fmt.Errorf("filter artifacts: %w", err)
	}

	logf("stage deploy: host=%s:%d", bundle.Host.Address, bundle.Host.Port)
	if err := e.deployToRemote(bundle, artifactDir, runID, logf); err != nil {
		return err
	}

	return nil
}

func (e *Executor) runLocalCommand(ctx context.Context, name string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (e *Executor) runLocalCommandWithLogging(ctx context.Context, logf func(string, ...any), name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()

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

	// 实时读取stdout并记录到日志
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				logf("%s", line)
			}
		}
	}()

	// 实时读取stderr并记录到日志
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				logf("%s", line)
			}
		}
	}()

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

	mountDir := filepath.ToSlash(absSourceDir)
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", mountDir),
		"-w", "/workspace",
		image,
		"sh", "-lc", script,
	}
	return e.runLocalCommand(ctx, "docker", args)
}

func (e *Executor) runDockerBuildWithLogging(ctx context.Context, sourceDir, image string, commands []string, logf func(string, ...any)) error {
	script := "set -eu\n" + strings.Join(commands, "\n")
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("resolve source dir: %w", err)
	}

	mountDir := filepath.ToSlash(absSourceDir)
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", mountDir),
		"-w", "/workspace",
		image,
		"sh", "-lc", script,
	}
	return e.runLocalCommandWithLogging(ctx, logf, "docker", args)
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

	// 实时读取stdout并记录到日志
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				logf("%s", line)
			}
		}
	}()

	// 实时读取stderr并记录到日志
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				logf("%s", line)
			}
		}
	}()

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
) error {
	// 构造通知载荷
	payload := model.NotificationPayload{
		RunID:        runID,
		Status:       status,
		ProjectID:    bundle.Project.ID,
		ProjectName:  bundle.Project.Name,
		RepoURL:      bundle.Project.RepoURL,
		Branch:       bundle.Project.Branch,
		TriggerType:  triggerType,
		TriggerRef:   triggerRef,
		ErrorMessage: errorMessage,
		SentAt:       time.Now().UTC().Format(time.RFC3339),
	}

	// 优先使用配置的通知渠道
	if bundle.DeployConfig.NotificationChannelID != nil {
		channel, err := e.store.GetNotificationChannel(ctx, *bundle.DeployConfig.NotificationChannelID)
		if err != nil {
			// 如果获取通知渠道失败，尝试使用默认渠道
			e.logger.Warn("failed to get configured notification channel, trying default", "channel_id", *bundle.DeployConfig.NotificationChannelID, "error", err)
			return e.sendDefaultNotification(ctx, bundle, payload)
		}

		if err := e.notifySender.Send(channel, payload); err != nil {
			return fmt.Errorf("send notification via channel: %w", err)
		}

		e.logger.Info("notification sent via channel", "channel_id", channel.ID, "run_id", runID)
		return nil
	}

	// 如果没有配置通知渠道，尝试使用系统默认渠道
	return e.sendDefaultNotification(ctx, bundle, payload)
}

func (e *Executor) sendDefaultNotification(ctx context.Context, bundle model.ExecutionBundle, payload model.NotificationPayload) error {
	channel, err := e.store.GetDefaultNotificationChannel(ctx)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			e.logger.Error("failed to get default notification channel", "error", err)
		}

		// 如果没有通知渠道，尝试使用原来的webhook方式
		return e.sendLegacyWebhook(ctx, bundle, payload)
	}

	if err := e.notifySender.Send(channel, payload); err != nil {
		return fmt.Errorf("send notification via default channel: %w", err)
	}

	e.logger.Info("notification sent via default channel", "channel_id", channel.ID, "run_id", payload.RunID)
	return nil
}

func (e *Executor) sendLegacyWebhook(ctx context.Context, bundle model.ExecutionBundle, payload model.NotificationPayload) error {
	webhookURL := strings.TrimSpace(bundle.DeployConfig.NotifyWebhookURL)
	if webhookURL == "" {
		return nil // 没有配置webhook，直接跳过
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
		return fmt.Errorf("marshal notification payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if bundle.DeployConfig.NotifyBearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bundle.DeployConfig.NotifyBearerToken)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send notification request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
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
	// 根据Git认证类型选择克隆方式
	switch project.GitAuthType {
	case model.GitAuthTypeNone:
		// 无需认证，直接克隆
		return e.runLocalCommandWithLogging(
			ctx,
			logf,
			"git",
			[]string{"clone", "--depth", "1", "--single-branch", "--branch", project.Branch, "--progress", project.RepoURL, sourceDir},
		)
		
	case model.GitAuthTypeUsername, model.GitAuthTypeToken:
		// 使用用户名密码或Token认证
		if project.GitUsername == "" || project.GitPassword == "" {
			return fmt.Errorf("git username/password is required for %s authentication", project.GitAuthType)
		}

		// 构造带认证的URL（注意：这个方式会在URL中暴露密码，仅作为示例）
		// 生产环境中更安全的做法是使用Git凭证管理器或SSH
		authURL := e.constructAuthenticatedURL(project.RepoURL, project.GitUsername, project.GitPassword)
		return e.runLocalCommandWithLogging(
			ctx,
			logf,
			"git",
			[]string{"clone", "--depth", "1", "--single-branch", "--branch", project.Branch, "--progress", authURL, sourceDir},
		)
		
	case model.GitAuthTypeSSH:
		// 使用SSH密钥认证
		if project.GitSSHKey == "" {
			return fmt.Errorf("ssh key is required for ssh authentication")
		}

		return e.runGitCloneWithSSH(ctx, project, sourceDir, logf)
		
	default:
		return fmt.Errorf("unsupported git authentication type: %s", project.GitAuthType)
	}
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

func (e *Executor) runGitCloneWithSSH(ctx context.Context, project model.Project, sourceDir string, logf func(string, ...any)) error {
	// 创建临时SSH密钥文件
	sshKeyFile, err := e.createTempSSHKeyFile(project.GitSSHKey)
	if err != nil {
		return fmt.Errorf("create temporary ssh key file: %w", err)
	}
	defer os.Remove(sshKeyFile) // 清理临时文件

	// 构造SSH命令
	sshCmd := []string{
		"git",
		"-c", fmt.Sprintf("core.sshCommand=ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", sshKeyFile),
		"clone",
		"--depth", "1",
		"--single-branch",
		"--branch", project.Branch,
		"--progress",
		project.RepoURL,
		sourceDir,
	}

	return e.runLocalCommandWithLogging(ctx, logf, "git", sshCmd[1:])
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
