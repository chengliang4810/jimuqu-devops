package com.jimuqu.devops.controller;

import com.jimuqu.devops.common.dto.Result;
import lombok.extern.slf4j.Slf4j;
import org.noear.solon.annotation.Controller;
import org.noear.solon.annotation.Mapping;
import org.noear.solon.core.handle.MethodType;

import java.util.HashMap;
import java.util.Map;

/**
 * 系统首页控制器 
 */
@Slf4j
@Controller
public class IndexController {
    
    /**
     * 系统首页
     */
    @Mapping("/")
    public String index() {
        return "欢迎使用Jimuqu DevOps平台！";
    }
    
    /**
     * API根路径
     */
    @Mapping("/api")
    public Result<Map<String, Object>> apiIndex() {
        Map<String, Object> info = new HashMap<>();
        info.put("name", "Jimuqu DevOps Platform");
        info.put("version", "1.0.0");
        info.put("description", "基于Docker容器的自动化部署平台");
        info.put("endpoints", Map.of(
            "hosts", "/api/hosts - 主机管理",
            "applications", "/api/applications - 应用管理",
            "builds", "/api/builds - 构建管理"
        ));
        
        return Result.success("Jimuqu DevOps API", info);
    }
    
    /**
     * 健康检查
     */
    @Mapping(value = "/health", method = MethodType.GET)
    public Result<String> health() {
        return Result.success("系统运行正常", "OK");
    }
}