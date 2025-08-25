package com.jimuqu.devops.entity;

import com.jimuqu.devops.common.entity.BaseEntity;
import com.jimuqu.devops.common.enums.BuildStatus;
import lombok.Data;
import lombok.EqualsAndHashCode;
import java.time.LocalDateTime;

/**
 * 构建记录实体
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class Build extends BaseEntity {
    
    /**
     * 关联应用ID
     */
    private Long applicationId;
    
    /**
     * 构建编号
     */
    private Integer buildNumber;
    
    /**
     * 构建状态
     */
    private BuildStatus status;
    
    /**
     * 触发方式
     */
    private String triggeredBy;
    
    /**
     * 触发用户
     */
    private String triggerUser;
    
    /**
     * Git提交哈希
     */
    private String gitCommit;
    
    /**
     * Git分支
     */
    private String gitBranch;
    
    /**
     * 开始时间
     */
    private LocalDateTime startTime;
    
    /**
     * 结束时间
     */
    private LocalDateTime endTime;
    
    /**
     * 构建时长(秒)
     */
    private Integer durationSeconds;
    
    /**
     * 日志文件路径
     */
    private String logFilePath;
    
    /**
     * 错误信息
     */
    private String errorMessage;
}