package com.jimuqu.devops.mapper.impl;

import com.jimuqu.devops.entity.Host;
import com.jimuqu.devops.mapper.HostMapper;
import com.jimuqu.devops.common.enums.HostStatus;
import org.noear.solon.annotation.Component;
import org.apache.commons.lang3.StringUtils;

import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;
import java.util.stream.Collectors;

/**
 * 主机数据访问层实现（内存版本）
 * 生产环境应使用数据库实现
 */
@Component
public class HostMapperImpl implements HostMapper {
    
    private final Map<Long, Host> hostStore = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(1);
    
    @Override
    public int insert(Host host) {
        Long id = idGenerator.getAndIncrement();
        host.setId(id);
        host.setCreateTime(LocalDateTime.now());
        host.setUpdateTime(LocalDateTime.now());
        hostStore.put(id, host);
        return 1;
    }
    
    @Override
    public int deleteById(Long id) {
        Host removed = hostStore.remove(id);
        return removed != null ? 1 : 0;
    }
    
    @Override
    public int update(Host host) {
        if (host.getId() == null || !hostStore.containsKey(host.getId())) {
            return 0;
        }
        host.setUpdateTime(LocalDateTime.now());
        hostStore.put(host.getId(), host);
        return 1;
    }
    
    @Override
    public Optional<Host> findById(Long id) {
        return Optional.ofNullable(hostStore.get(id));
    }
    
    @Override
    public Optional<Host> findByName(String name) {
        return hostStore.values().stream()
            .filter(host -> name.equals(host.getName()))
            .findFirst();
    }
    
    @Override
    public List<Host> findByPage(int offset, int limit, String keyword) {
        List<Host> filtered = hostStore.values().stream()
            .filter(host -> StringUtils.isBlank(keyword) || 
                          host.getName().contains(keyword) || 
                          host.getHostIp().contains(keyword))
            .sorted((h1, h2) -> h2.getCreateTime().compareTo(h1.getCreateTime()))
            .collect(Collectors.toList());
            
        return filtered.stream()
            .skip(offset)
            .limit(limit)
            .collect(Collectors.toList());
    }
    
    @Override
    public long count(String keyword) {
        return hostStore.values().stream()
            .filter(host -> StringUtils.isBlank(keyword) || 
                          host.getName().contains(keyword) || 
                          host.getHostIp().contains(keyword))
            .count();
    }
    
    @Override
    public List<Host> findByStatus(HostStatus status) {
        return hostStore.values().stream()
            .filter(host -> status.equals(host.getStatus()))
            .collect(Collectors.toList());
    }
    
    @Override
    public int updateStatusBatch(List<Long> ids, HostStatus status) {
        int count = 0;
        for (Long id : ids) {
            Host host = hostStore.get(id);
            if (host != null) {
                host.setStatus(status);
                host.setUpdateTime(LocalDateTime.now());
                count++;
            }
        }
        return count;
    }
}