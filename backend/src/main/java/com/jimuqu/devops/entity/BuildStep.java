package com.jimuqu.devops.entity;

import com.jimuqu.devops.common.entity.BaseEntity;
import com.jimuqu.devops.common.enums.BuildStatus;
import lombok.Data;
import lombok.EqualsAndHashCode;
import java.time.LocalDateTime;

/**
 * 构建步骤记录实体
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class BuildStep extends BaseEntity {
    
    /**
     * 关联构建ID
     */
    private Long buildId;
    
    /**
     * 步骤名称
     */
    private String stepName;
    
    /**
     * 步骤顺序
     */
    private Integer stepOrder;
    
    /**
     * 步骤状态
     */
    private BuildStatus status;
    
    /**
     * 开始时间
     */
    private LocalDateTime startTime;
    
    /**
     * 结束时间
     */
    private LocalDateTime endTime;
    
    /**
     * 执行时长(秒)
     */
    private Integer durationSeconds;
    
    /**
     * 日志内容
     */
    private String logContent;
    
    /**
     * 错误信息
     */
    private String errorMessage;
}