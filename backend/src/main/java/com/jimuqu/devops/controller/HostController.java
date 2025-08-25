package com.jimuqu.devops.controller;

import com.jimuqu.devops.entity.Host;
import com.jimuqu.devops.service.HostService;
import com.jimuqu.devops.common.dto.Result;
import com.jimuqu.devops.common.dto.PageResult;
import lombok.extern.slf4j.Slf4j;
import org.noear.solon.annotation.*;
import org.noear.solon.core.handle.MethodType;

import java.util.List;
import java.util.Optional;

/**
 * 主机管理控制器
 */
@Slf4j
@Controller
@Mapping("/api/hosts")
public class HostController {
    
    @Inject
    private HostService hostService;
    
    /**
     * 分页查询主机列表
     */
    @Mapping(value = "", method = MethodType.GET)
    public Result<PageResult<Host>> listHosts(
            @Param(defaultValue = "1") int page,
            @Param(defaultValue = "10") int size,
            @Param(required = false) String keyword) {
        
        try {
            PageResult<Host> result = hostService.getHostPage(page, size, keyword);
            return Result.success(result);
        } catch (Exception e) {
            log.error("查询主机列表失败", e);
            return Result.error("查询主机列表失败: " + e.getMessage());
        }
    }
    
    /**
     * 根据ID查询主机详情
     */
    @Mapping(value = "/{id}", method = MethodType.GET)
    public Result<Host> getHost(@Path Long id) {
        try {
            Optional<Host> hostOpt = hostService.getHostById(id);
            if (hostOpt.isPresent()) {
                return Result.success(hostOpt.get());
            } else {
                return Result.notFound("主机不存在");
            }
        } catch (Exception e) {
            log.error("查询主机详情失败", e);
            return Result.error("查询主机详情失败: " + e.getMessage());
        }
    }
    
    /**
     * 创建主机
     */
    @Mapping(value = "", method = MethodType.POST)
    public Result<Host> createHost(@Body Host host) {
        try {
            // 基本参数验证
            if (host.getName() == null || host.getName().trim().isEmpty()) {
                return Result.badRequest("主机名称不能为空");
            }
            if (host.getHostIp() == null || host.getHostIp().trim().isEmpty()) {
                return Result.badRequest("主机IP不能为空");
            }
            if (host.getUsername() == null || host.getUsername().trim().isEmpty()) {
                return Result.badRequest("用户名不能为空");
            }
            
            Host created = hostService.createHost(host);
            return Result.success("主机创建成功", created);
        } catch (IllegalArgumentException e) {
            return Result.badRequest(e.getMessage());
        } catch (Exception e) {
            log.error("创建主机失败", e);
            return Result.error("创建主机失败: " + e.getMessage());
        }
    }
    
    /**
     * 更新主机
     */
    @Mapping(value = "/{id}", method = MethodType.PUT)
    public Result<Host> updateHost(@Path Long id, @Body Host host) {
        try {
            host.setId(id);
            
            // 基本参数验证
            if (host.getName() == null || host.getName().trim().isEmpty()) {
                return Result.badRequest("主机名称不能为空");
            }
            if (host.getHostIp() == null || host.getHostIp().trim().isEmpty()) {
                return Result.badRequest("主机IP不能为空");
            }
            if (host.getUsername() == null || host.getUsername().trim().isEmpty()) {
                return Result.badRequest("用户名不能为空");
            }
            
            Host updated = hostService.updateHost(host);
            return Result.success("主机更新成功", updated);
        } catch (IllegalArgumentException e) {
            return Result.badRequest(e.getMessage());
        } catch (Exception e) {
            log.error("更新主机失败", e);
            return Result.error("更新主机失败: " + e.getMessage());
        }
    }
    
    /**
     * 删除主机
     */
    @Mapping(value = "/{id}", method = MethodType.DELETE)
    public Result<Void> deleteHost(@Path Long id) {
        try {
            boolean deleted = hostService.deleteHost(id);
            if (deleted) {
                return Result.success("主机删除成功");
            } else {
                return Result.notFound("主机不存在");
            }
        } catch (Exception e) {
            log.error("删除主机失败", e);
            return Result.error("删除主机失败: " + e.getMessage());
        }
    }
    
    /**
     * 测试主机连接
     */
    @Mapping(value = "/{id}/test", method = MethodType.POST)
    public Result<Boolean> testConnection(@Path Long id) {
        try {
            boolean connected = hostService.testConnection(id);
            String message = connected ? "连接成功" : "连接失败";
            return Result.success(message, connected);
        } catch (Exception e) {
            log.error("测试主机连接失败", e);
            return Result.error("测试主机连接失败: " + e.getMessage());
        }
    }
    
    /**
     * 获取在线主机列表
     */
    @Mapping(value = "/online", method = MethodType.GET)
    public Result<List<Host>> getOnlineHosts() {
        try {
            List<Host> hosts = hostService.getOnlineHosts();
            return Result.success(hosts);
        } catch (Exception e) {
            log.error("查询在线主机失败", e);
            return Result.error("查询在线主机失败: " + e.getMessage());
        }
    }
    
    /**
     * 批量更新主机状态
     */
    @Mapping(value = "/status/update", method = MethodType.POST)
    public Result<Void> updateHostsStatus() {
        try {
            hostService.updateHostsStatus();
            return Result.success("已开始更新所有主机状态，请稍后查看结果");
        } catch (Exception e) {
            log.error("更新主机状态失败", e);
            return Result.error("更新主机状态失败: " + e.getMessage());
        }
    }
}