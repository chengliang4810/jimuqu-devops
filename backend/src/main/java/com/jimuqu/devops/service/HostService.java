package com.jimuqu.devops.service;

import com.jimuqu.devops.entity.Host;
import com.jimuqu.devops.common.dto.PageResult;
import com.jimuqu.devops.common.enums.HostStatus;

import java.util.List;
import java.util.Optional;

/**
 * 主机管理服务接口
 */
public interface HostService {
    
    /**
     * 创建主机
     */
    Host createHost(Host host);
    
    /**
     * 删除主机
     */
    boolean deleteHost(Long id);
    
    /**
     * 更新主机
     */
    Host updateHost(Host host);
    
    /**
     * 根据ID查询主机
     */
    Optional<Host> getHostById(Long id);
    
    /**
     * 分页查询主机列表
     */
    PageResult<Host> getHostPage(int page, int size, String keyword);
    
    /**
     * 获取所有在线主机
     */
    List<Host> getOnlineHosts();
    
    /**
     * 测试主机连接
     */
    boolean testConnection(Long id);
    
    /**
     * 批量测试主机连接并更新状态
     */
    void updateHostsStatus();
    
    /**
     * 检查主机名是否存在
     */
    boolean existsByName(String name);
}