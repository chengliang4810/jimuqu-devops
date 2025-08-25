package com.jimuqu.devops.common.enums;

/**
 * 主机状态枚举
 */
public enum HostStatus {
    ONLINE("在线"),
    OFFLINE("离线"),
    ERROR("错误");
    
    private final String description;
    
    HostStatus(String description) {
        this.description = description;
    }
    
    public String getDescription() {
        return description;
    }
}