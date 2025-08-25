package com.jimuqu.devops.entity;

import com.jimuqu.devops.common.entity.BaseEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;
import java.util.List;
import java.util.Map;

/**
 * 应用管理实体
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class Application extends BaseEntity {
    
    /**
     * 应用名称
     */
    private String name;
    
    /**
     * Git仓库地址
     */
    private String gitRepoUrl;
    
    /**
     * Git分支
     */
    private String gitBranch = "main";
    
    /**
     * Git用户名
     */
    private String gitUsername;
    
    /**
     * Git密码(加密存储)
     */
    private String gitPassword;
    
    /**
     * Git私钥
     */
    private String gitPrivateKey;
    
    /**
     * 关联模板ID
     */
    private Long templateId;
    
    /**
     * 目标主机ID列表
     */
    private List<Long> hostIds;
    
    /**
     * 环境变量
     */
    private Map<String, String> variables;
    
    /**
     * 通知地址
     */
    private String notificationUrl;
    
    /**
     * 通知令牌
     */
    private String notificationToken;
    
    /**
     * 是否自动触发
     */
    private Boolean autoTrigger = false;
    
    /**
     * Webhook密钥
     */
    private String webhookSecret;
    
    /**
     * 应用状态
     */
    private String status = "ACTIVE";
}