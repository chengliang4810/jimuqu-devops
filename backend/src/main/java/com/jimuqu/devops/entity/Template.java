package com.jimuqu.devops.entity;

import com.jimuqu.devops.common.entity.BaseEntity;
import com.jimuqu.devops.common.enums.TemplateType;
import lombok.Data;
import lombok.EqualsAndHashCode;

/**
 * 模板管理实体
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class Template extends BaseEntity {
    
    /**
     * 模板名称
     */
    private String name;
    
    /**
     * 模板类型
     */
    private TemplateType type;
    
    /**
     * 模板描述
     */
    private String description;
    
    /**
     * 是否为系统预置模板
     */
    private Boolean isSystem = false;
}