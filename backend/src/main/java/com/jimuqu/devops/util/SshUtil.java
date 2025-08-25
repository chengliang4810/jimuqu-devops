package com.jimuqu.devops.util;

import com.jcraft.jsch.*;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;

/**
 * SSH工具类
 */
@Slf4j
public class SshUtil {
    
    private static final int CONNECTION_TIMEOUT = 30000;  // 30秒连接超时
    private static final int COMMAND_TIMEOUT = 300000;    // 5分钟命令超时
    
    /**
     * SSH连接结果
     */
    public static class SshResult {
        private final boolean success;
        private final String output;
        private final String error;
        private final int exitCode;
        
        public SshResult(boolean success, String output, String error, int exitCode) {
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
     * 测试SSH连接
     */
    public static boolean testConnection(String host, int port, String username, String password, String privateKey) {
        JSch jsch = new JSch();
        Session session = null;
        
        try {
            session = jsch.getSession(username, host, port);
            
            // 设置认证方式
            if (StringUtils.isNotBlank(privateKey)) {
                // 使用私钥认证
                jsch.addIdentity("key", privateKey.getBytes(), null, null);
            } else if (StringUtils.isNotBlank(password)) {
                // 使用密码认证
                session.setPassword(password);
            } else {
                log.error("SSH连接失败: 未提供密码或私钥");
                return false;
            }
            
            // 跳过主机密钥检查
            session.setConfig("StrictHostKeyChecking", "no");
            session.setTimeout(CONNECTION_TIMEOUT);
            
            session.connect();
            
            // 执行简单命令测试
            SshResult result = executeCommand(session, "echo 'connection test'");
            return result.isSuccess();
            
        } catch (Exception e) {
            log.error("SSH连接测试失败: {}", e.getMessage());
            return false;
        } finally {
            if (session != null && session.isConnected()) {
                session.disconnect();
            }
        }
    }
    
    /**
     * 执行SSH命令
     */
    public static SshResult executeCommand(String host, int port, String username, String password, String privateKey, String command) {
        JSch jsch = new JSch();
        Session session = null;
        
        try {
            session = jsch.getSession(username, host, port);
            
            if (StringUtils.isNotBlank(privateKey)) {
                jsch.addIdentity("key", privateKey.getBytes(), null, null);
            } else if (StringUtils.isNotBlank(password)) {
                session.setPassword(password);
            } else {
                return new SshResult(false, "", "未提供密码或私钥", -1);
            }
            
            session.setConfig("StrictHostKeyChecking", "no");
            session.setTimeout(CONNECTION_TIMEOUT);
            session.connect();
            
            return executeCommand(session, command);
            
        } catch (Exception e) {
            log.error("SSH命令执行失败: {}", e.getMessage());
            return new SshResult(false, "", e.getMessage(), -1);
        } finally {
            if (session != null && session.isConnected()) {
                session.disconnect();
            }
        }
    }
    
    /**
     * 在已建立的会话中执行命令
     */
    private static SshResult executeCommand(Session session, String command) {
        ChannelExec channel = null;
        ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
        ByteArrayOutputStream errorStream = new ByteArrayOutputStream();
        
        try {
            channel = (ChannelExec) session.openChannel("exec");
            channel.setCommand(command);
            channel.setInputStream(null);
            channel.setOutputStream(outputStream);
            channel.setErrStream(errorStream);
            
            channel.connect(COMMAND_TIMEOUT);
            
            // 等待命令执行完成
            while (!channel.isClosed()) {
                Thread.sleep(100);
            }
            
            int exitCode = channel.getExitStatus();
            String output = outputStream.toString(StandardCharsets.UTF_8);
            String error = errorStream.toString(StandardCharsets.UTF_8);
            
            return new SshResult(exitCode == 0, output, error, exitCode);
            
        } catch (Exception e) {
            log.error("SSH命令执行异常: {}", e.getMessage());
            return new SshResult(false, "", e.getMessage(), -1);
        } finally {
            if (channel != null && channel.isConnected()) {
                channel.disconnect();
            }
            try {
                outputStream.close();
                errorStream.close();
            } catch (IOException e) {
                log.warn("关闭输出流失败: {}", e.getMessage());
            }
        }
    }
}