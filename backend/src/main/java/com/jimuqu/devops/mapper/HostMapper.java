package com.jimuqu.devops.mapper;

import com.jimuqu.devops.entity.Host;
import com.jimuqu.devops.common.enums.HostStatus;
import org.noear.solon.annotation.Component;
import java.util.List;
import java.util.Optional;

/**
 * 主机数据访问层
 */
@Component
public interface HostMapper {
    
    /**
     * 插入主机
     */
    int insert(Host host);
    
    /**
     * 根据ID删除主机
     */
    int deleteById(Long id);
    
    /**
     * 更新主机
     */
    int update(Host host);
    
    /**
     * 根据ID查询主机
     */
    Optional<Host> findById(Long id);
    
    /**
     * 根据名称查询主机
     */
    Optional<Host> findByName(String name);
    
    /**
     * 分页查询主机列表
     */
    List<Host> findByPage(int offset, int limit, String keyword);
    
    /**
     * 统计主机总数
     */
    long count(String keyword);
    
    /**
     * 根据状态查询主机列表
     */
    List<Host> findByStatus(HostStatus status);
    
    /**
     * 批量更新主机状态
     */
    int updateStatusBatch(List<Long> ids, HostStatus status);
}