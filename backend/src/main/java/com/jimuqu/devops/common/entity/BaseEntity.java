package com.jimuqu.devops.common.entity;

import lombok.Data;
import java.time.LocalDateTime;

/**
 * 基础实体类
 */
@Data
public abstract class BaseEntity {
    
    /**
     * 主键ID
     */
    private Long id;
    
    /**
     * 创建时间
     */
    private LocalDateTime createTime;
    
    /**
     * 更新时间
     */
    private LocalDateTime updateTime;
}