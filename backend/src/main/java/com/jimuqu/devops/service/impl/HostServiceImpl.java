package com.jimuqu.devops.service.impl;

import com.jimuqu.devops.entity.Host;
import com.jimuqu.devops.mapper.HostMapper;
import com.jimuqu.devops.service.HostService;
import com.jimuqu.devops.common.dto.PageResult;
import com.jimuqu.devops.common.enums.HostStatus;
import com.jimuqu.devops.util.PasswordUtil;
import com.jimuqu.devops.util.SshUtil;
import lombok.extern.slf4j.Slf4j;
import org.noear.solon.annotation.Component;
import org.noear.solon.annotation.Inject;
import org.apache.commons.lang3.StringUtils;

import java.util.List;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.Executor;
import java.util.concurrent.Executors;

/**
 * 主机管理服务实现
 */
@Slf4j
@Component
public class HostServiceImpl implements HostService {
    
    @Inject
    private HostMapper hostMapper;
    
    private final Executor executor = Executors.newFixedThreadPool(10);
    
    @Override
    public Host createHost(Host host) {
        // 验证主机名是否重复
        if (existsByName(host.getName())) {
            throw new IllegalArgumentException("主机名已存在: " + host.getName());
        }
        
        // 加密密码
        if (StringUtils.isNotBlank(host.getPassword())) {
            host.setPassword(PasswordUtil.encode(host.getPassword()));
        }
        
        // 设置默认状态
        host.setStatus(HostStatus.OFFLINE);
        
        hostMapper.insert(host);
        
        // 异步测试连接
        testConnectionAsync(host);
        
        return host;
    }
    
    @Override
    public boolean deleteHost(Long id) {
        return hostMapper.deleteById(id) > 0;
    }
    
    @Override
    public Host updateHost(Host host) {
        Optional<Host> existingOpt = hostMapper.findById(host.getId());
        if (existingOpt.isEmpty()) {
            throw new IllegalArgumentException("主机不存在: " + host.getId());
        }
        
        Host existing = existingOpt.get();
        
        // 检查名称是否重复（排除自己）
        if (!existing.getName().equals(host.getName()) && existsByName(host.getName())) {
            throw new IllegalArgumentException("主机名已存在: " + host.getName());
        }
        
        // 如果密码有变化，重新加密
        if (StringUtils.isNotBlank(host.getPassword()) && 
            !PasswordUtil.matches(host.getPassword(), existing.getPassword())) {
            host.setPassword(PasswordUtil.encode(host.getPassword()));
        } else {
            // 保持原密码
            host.setPassword(existing.getPassword());
        }
        
        hostMapper.update(host);
        
        // 异步测试连接
        testConnectionAsync(host);
        
        return host;
    }
    
    @Override
    public Optional<Host> getHostById(Long id) {
        return hostMapper.findById(id);
    }
    
    @Override
    public PageResult<Host> getHostPage(int page, int size, String keyword) {
        int offset = (page - 1) * size;
        List<Host> hosts = hostMapper.findByPage(offset, size, keyword);
        long total = hostMapper.count(keyword);
        
        // 隐藏敏感信息
        hosts.forEach(this::hideSensitiveInfo);
        
        return new PageResult<>(hosts, total, (long) page, (long) size);
    }
    
    @Override
    public List<Host> getOnlineHosts() {
        List<Host> hosts = hostMapper.findByStatus(HostStatus.ONLINE);
        hosts.forEach(this::hideSensitiveInfo);
        return hosts;
    }
    
    @Override
    public boolean testConnection(Long id) {
        Optional<Host> hostOpt = hostMapper.findById(id);
        if (hostOpt.isEmpty()) {
            return false;
        }
        
        Host host = hostOpt.get();
        boolean connected = testConnectionInternal(host);
        
        // 更新主机状态
        host.setStatus(connected ? HostStatus.ONLINE : HostStatus.OFFLINE);
        hostMapper.update(host);
        
        return connected;
    }
    
    @Override
    public void updateHostsStatus() {
        List<Host> allHosts = hostMapper.findByPage(0, Integer.MAX_VALUE, null);
        
        // 异步批量测试连接
        List<CompletableFuture<Void>> futures = allHosts.stream()
            .map(host -> CompletableFuture.runAsync(() -> {
                boolean connected = testConnectionInternal(host);
                host.setStatus(connected ? HostStatus.ONLINE : HostStatus.OFFLINE);
                hostMapper.update(host);
            }, executor))
            .toList();
        
        // 等待所有测试完成
        CompletableFuture.allOf(futures.toArray(new CompletableFuture[0]))
            .whenComplete((result, throwable) -> {
                if (throwable != null) {
                    log.error("批量测试主机连接时发生异常", throwable);
                } else {
                    log.info("已完成 {} 台主机的连接状态更新", allHosts.size());
                }
            });
    }
    
    @Override
    public boolean existsByName(String name) {
        return hostMapper.findByName(name).isPresent();
    }
    
    /**
     * 异步测试连接
     */
    private void testConnectionAsync(Host host) {
        CompletableFuture.runAsync(() -> {
            boolean connected = testConnectionInternal(host);
            host.setStatus(connected ? HostStatus.ONLINE : HostStatus.OFFLINE);
            hostMapper.update(host);
        }, executor);
    }
    
    /**
     * 内部测试连接方法
     */
    private boolean testConnectionInternal(Host host) {
        try {
            // 这里需要解密密码，为简化演示，假设密码未加密
            // 实际实现时需要先解密密码
            String password = host.getPassword();
            String privateKey = host.getPrivateKey();
            
            return SshUtil.testConnection(
                host.getHostIp(),
                host.getPort(),
                host.getUsername(),
                password,
                privateKey
            );
        } catch (Exception e) {
            log.error("测试主机连接失败: {}", host.getName(), e);
            return false;
        }
    }
    
    /**
     * 隐藏敏感信息
     */
    private void hideSensitiveInfo(Host host) {
        host.setPassword("******");
        if (StringUtils.isNotBlank(host.getPrivateKey())) {
            host.setPrivateKey("******");
        }
    }
}