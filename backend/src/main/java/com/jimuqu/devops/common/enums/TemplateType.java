package com.jimuqu.devops.common.enums;

/**
 * 模板类型枚举
 */
public enum TemplateType {
    SPRINGBOOT("Spring Boot项目"),
    VUE("Vue前端项目"),
    REACT("React前端项目"), 
    STATIC("静态网站"),
    CUSTOM("自定义");
    
    private final String description;
    
    TemplateType(String description) {
        this.description = description;
    }
    
    public String getDescription() {
        return description;
    }
}