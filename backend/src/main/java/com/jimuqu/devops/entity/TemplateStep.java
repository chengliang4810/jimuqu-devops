package com.jimuqu.devops.entity;

import com.jimuqu.devops.common.entity.BaseEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;
import java.util.List;
import java.util.Map;

/**
 * 模板步骤实体
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class TemplateStep extends BaseEntity {
    
    /**
     * 关联模板ID
     */
    private Long templateId;
    
    /**
     * 步骤名称
     */
    private String name;
    
    /**
     * Docker镜像名
     */
    private String image;
    
    /**
     * 执行命令列表
     */
    private List<String> commands;
    
    /**
     * 执行顺序
     */
    private Integer stepOrder;
    
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
    private Map<String, String> environment;
}