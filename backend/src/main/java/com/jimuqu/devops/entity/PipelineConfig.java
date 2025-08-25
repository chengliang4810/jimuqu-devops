package com.jimuqu.devops.entity;

import com.jimuqu.devops.common.entity.BaseEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;
import java.util.List;

/**
 * 流水线配置实体
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class PipelineConfig extends BaseEntity {
    
    /**
     * 关联应用ID
     */
    private Long applicationId;
    
    /**
     * 流水线步骤列表
     */
    private List<PipelineStep> steps;
}

/**
 * 流水线步骤
 */
@Data
class PipelineStep {
    
    /**
     * 步骤名称
     */
    private String name;
    
    /**
     * Docker镜像
     */
    private String image;
    
    /**
     * 执行命令列表
     */
    private List<String> commands;
    
    /**
     * 执行顺序
     */
    private Integer order;
    
    /**
     * 出错是否继续
     */
    private Boolean continueOnError = false;
    
    /**
     * 工作目录
     */
    private String workDir = "/workspace";
    
    /**
     * 环境变量
     */
    private java.util.Map<String, String> environment;
}