package com.jimuqu.devops.entity;

import com.jimuqu.devops.common.entity.BaseEntity;
import com.jimuqu.devops.common.enums.HostStatus;
import lombok.Data;
import lombok.EqualsAndHashCode;

/**
 * 主机管理实体
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class Host extends BaseEntity {
    
    /**
     * 主机名称
     */
    private String name;
    
    /**
     * 主机IP地址
     */
    private String hostIp;
    
    /**
     * SSH端口
     */
    private Integer port = 22;
    
    /**
     * SSH用户名
     */
    private String username;
    
    /**
     * SSH密码(加密存储)
     */
    private String password;
    
    /**
     * SSH私钥
     */
    private String privateKey;
    
    /**
     * 主机状态
     */
    private HostStatus status = HostStatus.OFFLINE;
    
    /**
     * 主机描述
     */
    private String description;
}