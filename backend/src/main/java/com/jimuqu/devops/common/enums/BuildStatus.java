package com.jimuqu.devops.common.enums;

/**
 * 构建状态枚举
 */
public enum BuildStatus {
    PENDING("等待中"),
    RUNNING("运行中"),
    SUCCESS("成功"),
    FAILED("失败"),
    CANCELLED("已取消");
    
    private final String description;
    
    BuildStatus(String description) {
        this.description = description;
    }
    
    public String getDescription() {
        return description;
    }
}

/**
 * 触发方式枚举
 */
enum TriggerType {
    MANUAL("手动触发"),
    WEBHOOK("Webhook触发");
    
    private final String description;
    
    TriggerType(String description) {
        this.description = description;
    }
    
    public String getDescription() {
        return description;
    }
}

/**
 * 应用状态枚举
 */
enum ApplicationStatus {
    ACTIVE("激活"),
    INACTIVE("非激活");
    
    private final String description;
    
    ApplicationStatus(String description) {
        this.description = description;
    }
    
    public String getDescription() {
        return description;
    }
}