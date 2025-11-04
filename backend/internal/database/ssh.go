package database

import (
	"DevOpsProject/backend/internal/models"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHService SSH服务
type SSHService struct{}

// NewSSHService 创建SSH服务实例
func NewSSHService() *SSHService {
	return &SSHService{}
}

// ExecuteCommand 在指定主机上执行命令
func (s *SSHService) ExecuteCommand(hostID uint, command string, timeout int) (*models.SSHCommandResponse, error) {
	// 获取主机信息
	hostService := NewHostService()
	host, err := hostService.GetHostByID(hostID)
	if err != nil {
		return nil, fmt.Errorf("获取主机信息失败: %v", err)
	}

	// 设置默认超时时间
	if timeout <= 0 {
		timeout = 30
	}

	// 记录开始时间
	startTime := time.Now()
	startTimeStr := startTime.Format("2006-01-02 15:04:05")

	// 建立SSH连接
	client, session, err := s.createSSHConnection(host, timeout)
	if err != nil {
		return &models.SSHCommandResponse{
			HostID:    hostID,
			Command:   command,
			Error:     fmt.Sprintf("SSH连接失败: %v", err),
			ExitCode:  -1,
			Duration:  time.Since(startTime).Milliseconds(),
			StartTime: startTimeStr,
			EndTime:   time.Now().Format("2006-01-02 15:04:05"),
		}, nil
	}
	defer client.Close()
	defer session.Close()

	// 执行命令
	output, err := session.CombinedOutput(command)
	endTime := time.Now()
	endTimeStr := endTime.Format("2006-01-02 15:04:05")
	duration := endTime.Sub(startTime).Milliseconds()

	// 构建响应
	response := &models.SSHCommandResponse{
		HostID:    hostID,
		Command:   command,
		Output:    string(output),
		Duration:  duration,
		StartTime: startTimeStr,
		EndTime:   endTimeStr,
	}

	if err != nil {
		response.Error = err.Error()
		response.ExitCode = 1
	} else {
		response.ExitCode = 0
	}

	return response, nil
}

// createSSHConnection 建立SSH连接
func (s *SSHService) createSSHConnection(host *models.Host, timeout int) (*ssh.Client, *ssh.Session, error) {
	// SSH配置
	config := &ssh.ClientConfig{
		User: host.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(host.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}

	// 建立连接
	address := fmt.Sprintf("%s:%d", host.Host, host.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, nil, fmt.Errorf("连接失败 %s: %v", address, err)
	}

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("创建会话失败: %v", err)
	}

	// 设置终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显
		ssh.TTY_OP_ISPEED: 14400, // 输入速度 = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // 输出速度 = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		client.Close()
		return nil, nil, fmt.Errorf("请求PTY失败: %v", err)
	}

	return client, session, nil
}

// TestConnection 测试主机连接
func (s *SSHService) TestConnection(hostID uint) error {
	// 获取主机信息
	hostService := NewHostService()
	host, err := hostService.GetHostByID(hostID)
	if err != nil {
		return fmt.Errorf("获取主机信息失败: %v", err)
	}

	// 建立SSH连接
	client, session, err := s.createSSHConnection(host, 10)
	if err != nil {
		return fmt.Errorf("SSH连接失败: %v", err)
	}
	defer client.Close()
	defer session.Close()

	// 执行简单命令测试连接
	_, err = session.CombinedOutput("echo 'connection_test'")
	if err != nil {
		return fmt.Errorf("连接测试失败: %v", err)
	}

	return nil
}

// CheckHostStatus 检查主机状态
func (s *SSHService) CheckHostStatus(host *models.Host) string {
	// 尝试建立SSH连接
	client, session, err := s.createSSHConnection(host, 5)
	if err != nil {
		return "offline"
	}
	defer client.Close()
	defer session.Close()

	// 执行简单命令验证连接
	_, err = session.CombinedOutput("echo 'status_check'")
	if err != nil {
		return "offline"
	}

	return "online"
}

// BatchCheckHostStatus 批量检查主机状态
func (s *SSHService) BatchCheckHostStatus(hosts []models.Host) error {
	for i := range hosts {
		status := s.CheckHostStatus(&hosts[i])

		// 更新主机状态
		updates := map[string]interface{}{
			"status":     status,
			"updated_at": models.CustomTime{Time: time.Now()},
		}

		hostService := NewHostService()
		if err := hostService.UpdateHost(hosts[i].ID, updates); err != nil {
			return fmt.Errorf("更新主机 %d 状态失败: %v", hosts[i].ID, err)
		}
	}
	return nil
}

// PingHost ping主机检查网络连通性
func (s *SSHService) PingHost(host string, port int, timeout time.Duration) error {
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

// UploadFile 上传单个文件
func (s *SSHService) UploadFile(hostID uint, remotePath, content, permissions string, overwrite bool) (*models.FileUploadResponse, error) {
	// 获取主机信息
	hostService := NewHostService()
	host, err := hostService.GetHostByID(hostID)
	if err != nil {
		return nil, fmt.Errorf("获取主机信息失败: %v", err)
	}

	// 记录开始时间
	startTime := time.Now()
	startTimeStr := startTime.Format("2006-01-02 15:04:05")

	// 建立SSH连接
	client, session, err := s.createSSHConnection(host, 30)
	if err != nil {
		return &models.FileUploadResponse{
			HostID:     hostID,
			RemotePath: remotePath,
			Duration:   time.Since(startTime).Milliseconds(),
			StartTime:  startTimeStr,
			EndTime:    time.Now().Format("2006-01-02 15:04:05"),
			Message:    fmt.Sprintf("SSH连接失败: %v", err),
		}, nil
	}
	defer client.Close()

	// 检查文件是否已存在
	if !overwrite {
		checkSession, err := client.NewSession()
		if err == nil {
			defer checkSession.Close()
			checkCmd := fmt.Sprintf("test -f %s && echo 'exists' || echo 'not_exists'", remotePath)
			output, err := checkSession.CombinedOutput(checkCmd)
			if err == nil && strings.TrimSpace(string(output)) == "exists" {
				return &models.FileUploadResponse{
					HostID:     hostID,
					RemotePath: remotePath,
					Duration:   time.Since(startTime).Milliseconds(),
					StartTime:  startTimeStr,
					EndTime:    time.Now().Format("2006-01-02 15:04:05"),
					Message:    "文件已存在且不允许覆盖",
				}, nil
			}
		}
	}
	session.Close()

	// 创建目录（如果不存在）
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		mkdirSession, err := client.NewSession()
		if err == nil {
			defer mkdirSession.Close()
			mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
			if _, err := mkdirSession.CombinedOutput(mkdirCmd); err != nil {
				return &models.FileUploadResponse{
					HostID:     hostID,
					RemotePath: remotePath,
					Duration:   time.Since(startTime).Milliseconds(),
					StartTime:  startTimeStr,
					EndTime:    time.Now().Format("2006-01-02 15:04:05"),
					Message:    fmt.Sprintf("创建目录失败: %v", err),
				}, nil
			}
		}
	}

	// 使用echo命令写入文件（简化实现）
	echoSession, err := client.NewSession()
	if err != nil {
		return &models.FileUploadResponse{
			HostID:     hostID,
			RemotePath: remotePath,
			Duration:   time.Since(startTime).Milliseconds(),
			StartTime:  startTimeStr,
			EndTime:    time.Now().Format("2006-01-02 15:04:05"),
			Message:    fmt.Sprintf("创建写入会话失败: %v", err),
		}, nil
	}
	defer echoSession.Close()

	echoCmd := fmt.Sprintf("echo -n %s > %s", fmt.Sprintf("'%s'", strings.ReplaceAll(content, "'", `'"'"'`)), remotePath)
	if _, err := echoSession.CombinedOutput(echoCmd); err != nil {
		return &models.FileUploadResponse{
			HostID:     hostID,
			RemotePath: remotePath,
			Duration:   time.Since(startTime).Milliseconds(),
			StartTime:  startTimeStr,
			EndTime:    time.Now().Format("2006-01-02 15:04:05"),
			Message:    fmt.Sprintf("文件写入失败: %v", err),
		}, nil
	}

	// 设置文件权限
	actualPerms := permissions
	if permissions == "" {
		actualPerms = "0644"
	}
	chmodSession, err := client.NewSession()
	if err == nil {
		defer chmodSession.Close()
		chmodCmd := fmt.Sprintf("chmod %s %s", actualPerms, remotePath)
		chmodSession.CombinedOutput(chmodCmd)
	}

	endTime := time.Now()
	endTimeStr := endTime.Format("2006-01-02 15:04:05")
	duration := endTime.Sub(startTime).Milliseconds()

	return &models.FileUploadResponse{
		HostID:      hostID,
		RemotePath:  remotePath,
		Size:        int64(len(content)),
		Permissions: actualPerms,
		Duration:    duration,
		StartTime:   startTimeStr,
		EndTime:     endTimeStr,
		Message:     "文件上传成功",
	}, nil
}

// UploadDirectory 上传目录
func (s *SSHService) UploadDirectory(hostID uint, remotePath string, files []models.DirectoryFileItem, overwrite bool) (*models.DirectoryUploadResponse, error) {
	// 获取主机信息
	hostService := NewHostService()
	host, err := hostService.GetHostByID(hostID)
	if err != nil {
		return nil, fmt.Errorf("获取主机信息失败: %v", err)
	}

	// 记录开始时间
	startTime := time.Now()
	startTimeStr := startTime.Format("2006-01-02 15:04:05")

	// 建立SSH连接
	client, session, err := s.createSSHConnection(host, 60)
	if err != nil {
		return &models.DirectoryUploadResponse{
			HostID:       hostID,
			RemotePath:   remotePath,
			TotalFiles:   len(files),
			SuccessFiles: 0,
			Duration:     time.Since(startTime).Milliseconds(),
			StartTime:    startTimeStr,
			EndTime:      time.Now().Format("2006-01-02 15:04:05"),
			Message:      fmt.Sprintf("SSH连接失败: %v", err),
		}, nil
	}
	defer client.Close()
	defer session.Close()

	// 创建远程目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remotePath)
	if _, err := session.CombinedOutput(mkdirCmd); err != nil {
		return &models.DirectoryUploadResponse{
			HostID:       hostID,
			RemotePath:   remotePath,
			TotalFiles:   len(files),
			SuccessFiles: 0,
			Duration:     time.Since(startTime).Milliseconds(),
			StartTime:    startTimeStr,
			EndTime:      time.Now().Format("2006-01-02 15:04:05"),
			Message:      fmt.Sprintf("创建远程目录失败: %v", err),
		}, nil
	}

	var successCount int
	var failedFiles []models.DirectoryUploadError

	// 逐个上传文件
	for _, file := range files {
		if file.IsDir {
			// 创建子目录
			subDirPath := filepath.Join(remotePath, file.Path)
			mkdirCmd := fmt.Sprintf("mkdir -p %s", subDirPath)
			if _, err := session.CombinedOutput(mkdirCmd); err != nil {
				failedFiles = append(failedFiles, models.DirectoryUploadError{
					Path:  file.Path,
					Error: fmt.Sprintf("创建目录失败: %v", err),
				})
			} else {
				successCount++
			}
		} else {
			// 上传文件
			fullRemotePath := filepath.Join(remotePath, file.Path)

			// 检查文件是否已存在
			if !overwrite {
				checkCmd := fmt.Sprintf("test -f %s && echo 'exists' || echo 'not_exists'", fullRemotePath)
				output, err := session.CombinedOutput(checkCmd)
				if err == nil && strings.TrimSpace(string(output)) == "exists" {
					failedFiles = append(failedFiles, models.DirectoryUploadError{
						Path:  file.Path,
						Error: "文件已存在且不允许覆盖",
					})
					continue
				}
			}

			// 创建文件目录
			fileDir := filepath.Dir(fullRemotePath)
			if fileDir != "." {
				mkdirCmd := fmt.Sprintf("mkdir -p %s", fileDir)
				session.CombinedOutput(mkdirCmd)
			}

			// 使用echo命令写入文件
			echoCmd := fmt.Sprintf("echo -n %s > %s", fmt.Sprintf("'%s'", strings.ReplaceAll(file.Content, "'", `'"'"'`)), fullRemotePath)
			if _, err := session.CombinedOutput(echoCmd); err != nil {
				failedFiles = append(failedFiles, models.DirectoryUploadError{
					Path:  file.Path,
					Error: fmt.Sprintf("上传失败: %v", err),
				})
			} else {
				// 设置文件权限
				perms := file.Permissions
				if perms == "" {
					perms = "0644"
				}
				chmodCmd := fmt.Sprintf("chmod %s %s", perms, fullRemotePath)
				session.CombinedOutput(chmodCmd)

				successCount++
			}
		}
	}

	endTime := time.Now()
	endTimeStr := endTime.Format("2006-01-02 15:04:05")
	duration := endTime.Sub(startTime).Milliseconds()

	message := "目录上传完成"
	if len(failedFiles) > 0 {
		message = fmt.Sprintf("目录上传完成，成功%d个，失败%d个", successCount, len(failedFiles))
	}

	return &models.DirectoryUploadResponse{
		HostID:       hostID,
		RemotePath:   remotePath,
		TotalFiles:   len(files),
		SuccessFiles: successCount,
		FailedFiles:  failedFiles,
		Duration:     duration,
		StartTime:    startTimeStr,
		EndTime:      endTimeStr,
		Message:      message,
	}, nil
}