package com.jimuqu.devops.service;

import com.jimuqu.devops.entity.Template;
import com.jimuqu.devops.entity.TemplateStep;
import com.jimuqu.devops.common.dto.PageResult;
import com.jimuqu.devops.common.enums.TemplateType;

import java.util.List;
import java.util.Optional;

/**
 * 模板管理服务接口
 */
public interface TemplateService {
    
    /**
     * 创建模板
     */
    Template createTemplate(Template template, List<TemplateStep> steps);
    
    /**
     * 更新模板
     */
    Template updateTemplate(Template template, List<TemplateStep> steps);
    
    /**
     * 删除模板
     */
    boolean deleteTemplate(Long id);
    
    /**
     * 根据ID查询模板
     */
    Optional<Template> getTemplateById(Long id);
    
    /**
     * 分页查询模板列表
     */
    PageResult<Template> getTemplatePage(int page, int size, TemplateType type, String keyword);
    
    /**
     * 获取模板步骤
     */
    List<TemplateStep> getTemplateSteps(Long templateId);
    
    /**
     * 获取系统预置模板
     */
    List<Template> getSystemTemplates();
    
    /**
     * 复制模板
     */
    Template copyTemplate(Long id, String newName);
}