package com.jimuqu.devops;

import com.jimuqu.devops.entity.Host;
import com.jimuqu.devops.service.HostService;
import com.jimuqu.devops.service.PipelineExecutionService;
import com.jimuqu.devops.common.enums.HostStatus;
import org.junit.jupiter.api.Test;
import org.noear.solon.annotation.Inject;
import org.noear.solon.test.SolonTest;
import static org.junit.jupiter.api.Assertions.*;

/**
 * DevOps平台功能测试
 */
@SolonTest(App.class)
public class DevOpsPlatformTest {
    
    @Inject
    private HostService hostService;
    
    @Inject
    private PipelineExecutionService pipelineExecutionService;
    
    @Test
    public void testHostManagement() {
        // 测试创建主机
        Host host = new Host();
        host.setName("测试主机");
        host.setHostIp("192.168.1.100");
        host.setUsername("root");
        host.setPassword("password123");
        host.setDescription("用于测试的主机");
        
        Host created = hostService.createHost(host);
        assertNotNull(created.getId());
        assertEquals("测试主机", created.getName());
        assertEquals(HostStatus.OFFLINE, created.getStatus()); // 初始状态为离线
        
        // 测试查询主机
        var hostOpt = hostService.getHostById(created.getId());
        assertTrue(hostOpt.isPresent());
        assertEquals("测试主机", hostOpt.get().getName());
        
        // 测试更新主机
        created.setDescription("更新后的描述");
        Host updated = hostService.updateHost(created);
        assertEquals("更新后的描述", updated.getDescription());
        
        // 测试删除主机
        boolean deleted = hostService.deleteHost(created.getId());
        assertTrue(deleted);
        
        var deletedHost = hostService.getHostById(created.getId());
        assertFalse(deletedHost.isPresent());
    }
    
    @Test
    public void testPipelineExecution() {
        // 测试流水线执行
        PipelineExecutionService.BuildResult result = 
            pipelineExecutionService.execute(1L, "MANUAL");
        
        assertNotNull(result);
        // 注意：由于没有真实的Docker环境，这个测试可能会失败
        // 在生产环境中应该配置好Docker环境
    }
    
    @Test
    public void testHostNameValidation() {
        // 测试主机名重复验证
        Host host1 = new Host();
        host1.setName("重复名称主机");
        host1.setHostIp("192.168.1.101");
        host1.setUsername("root");
        host1.setPassword("password123");
        hostService.createHost(host1);
        
        Host host2 = new Host();
        host2.setName("重复名称主机");
        host2.setHostIp("192.168.1.102");
        host2.setUsername("root");
        host2.setPassword("password123");
        
        // 应该抛出异常
        assertThrows(IllegalArgumentException.class, () -> {
            hostService.createHost(host2);
        });
        
        // 清理
        hostService.deleteHost(host1.getId());
    }
}