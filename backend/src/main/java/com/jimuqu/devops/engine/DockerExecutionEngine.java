package com.jimuqu.devops.engine;

import com.github.dockerjava.api.DockerClient;
import com.github.dockerjava.api.command.CreateContainerResponse;
import com.github.dockerjava.api.model.Bind;
import com.github.dockerjava.api.model.HostConfig;
import com.github.dockerjava.api.model.Volume;
import com.github.dockerjava.core.DefaultDockerClientConfig;
import com.github.dockerjava.core.DockerClientImpl;
import com.github.dockerjava.httpclient5.ApacheDockerHttpClient;
import com.github.dockerjava.api.async.ResultCallbackTemplate;
import com.github.dockerjava.api.model.Frame;
import lombok.extern.slf4j.Slf4j;
import org.noear.solon.annotation.Component;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;
import java.util.Map;

/**
 * Docker执行引擎
 */
@Slf4j
@Component
public class DockerExecutionEngine {
    
    private final DockerClient dockerClient;
    
    public DockerExecutionEngine() {
        DefaultDockerClientConfig config = DefaultDockerClientConfig.createDefaultConfigBuilder().build();
        ApacheDockerHttpClient httpClient = new ApacheDockerHttpClient.Builder()
            .dockerHost(config.getDockerHost())
            .sslConfig(config.getSSLConfig())
            .build();
        this.dockerClient = DockerClientImpl.getInstance(config, httpClient);
    }
    
    /**
     * 容器执行结果
     */
    public static class ContainerResult {
        private final boolean success;
        private final String output;
        private final String error;
        private final int exitCode;
        
        public ContainerResult(boolean success, String output, String error, int exitCode) {
            this.success = success;
            this.output = output;
            this.error = error;
            this.exitCode = exitCode;
        }
        
        public boolean isSuccess() {
            return success;
        }
        
        public String getOutput() {
            return output;
        }
        
        public String getError() {
            return error;
        }
        
        public int getExitCode() {
            return exitCode;
        }
    }
    
    /**
     * 执行Docker容器
     */
    public ContainerResult runContainer(String image, List<String> commands, 
                                      String workspacePath, Map<String, String> environment) {
        String containerId = null;
        try {
            // 确保工作空间目录存在
            Path workspace = Paths.get(workspacePath);
            if (!Files.exists(workspace)) {
                Files.createDirectories(workspace);
            }
            
            // 构建执行命令
            String[] cmdArray = {"/bin/sh", "-c", String.join(" && ", commands)};
            
            // 创建容器
            CreateContainerResponse container = dockerClient.createContainerCmd(image)
                .withCmd(cmdArray)
                .withWorkingDir("/workspace")
                .withHostConfig(HostConfig.newHostConfig()
                    .withBinds(new Bind(workspacePath, new Volume("/workspace"))))
                .withEnv(environment.entrySet().stream()
                    .map(entry -> entry.getKey() + "=" + entry.getValue())
                    .toArray(String[]::new))
                .exec();
            
            containerId = container.getId();
            
            // 启动容器
            dockerClient.startContainerCmd(containerId).exec();
            
            // 等待容器执行完成
            int exitCode = dockerClient.waitContainerCmd(containerId)
                .start()
                .awaitStatusCode();
            
            // 获取日志
            String logs = dockerClient.logsCmd(containerId)
                .withStdOut(true)
                .withStdErr(true)
                .exec(new LogContainerResultCallback())
                .awaitCompletion()
                .toString();
            
            boolean success = exitCode == 0;
            return new ContainerResult(success, logs, success ? "" : logs, exitCode);
            
        } catch (Exception e) {
            log.error("Docker容器执行失败", e);
            return new ContainerResult(false, "", e.getMessage(), -1);
        } finally {
            // 清理容器
            if (containerId != null) {
                try {
                    dockerClient.removeContainerCmd(containerId).withForce(true).exec();
                } catch (Exception e) {
                    log.warn("清理容器失败: {}", e.getMessage());
                }
            }
        }
    }
    
    /**
     * 创建工作空间
     */
    public String createWorkspace(String basePath, String appName, int buildNumber) {
        String workspacePath = Paths.get(basePath, appName, String.valueOf(buildNumber)).toString();
        try {
            Files.createDirectories(Paths.get(workspacePath));
            return workspacePath;
        } catch (IOException e) {
            throw new RuntimeException("创建工作空间失败: " + e.getMessage(), e);
        }
    }
    
    /**
     * 清理工作空间
     */
    public void cleanupWorkspace(String workspacePath) {
        try {
            Path path = Paths.get(workspacePath);
            if (Files.exists(path)) {
                Files.walk(path)
                    .sorted((a, b) -> b.compareTo(a)) // 先删除文件，再删除目录
                    .forEach(p -> {
                        try {
                            Files.delete(p);
                        } catch (IOException e) {
                            log.warn("删除文件失败: {}", p, e);
                        }
                    });
            }
        } catch (Exception e) {
            log.warn("清理工作空间失败: {}", workspacePath, e);
        }
    }
    
    /**
     * 日志回调处理器
     */
    private static class LogContainerResultCallback extends ResultCallbackTemplate<LogContainerResultCallback, Frame> {
        private final StringBuilder logs = new StringBuilder();
        
        @Override
        public void onNext(Frame frame) {
            logs.append(new String(frame.getPayload()));
        }
        
        @Override
        public String toString() {
            return logs.toString();
        }
    }
}