package com.jimuqu.devops.controller;

import com.jimuqu.devops.service.PipelineExecutionService;
import com.jimuqu.devops.common.dto.Result;
import lombok.extern.slf4j.Slf4j;
import org.noear.solon.annotation.*;
import org.noear.solon.core.handle.MethodType;

import java.util.concurrent.CompletableFuture;

/**
 * 应用管理控制器
 */
@Slf4j
@Controller
@Mapping("/api/applications")
public class ApplicationController {
    
    @Inject
    private PipelineExecutionService pipelineExecutionService;
    
    /**
     * 手动触发构建
     */
    @Mapping(value = "/{id}/build", method = MethodType.POST)
    public Result<String> triggerBuild(@Path Long id) {
        try {
            log.info("手动触发构建: applicationId={}", id);
            
            // 异步执行构建
            CompletableFuture<PipelineExecutionService.BuildResult> future = 
                pipelineExecutionService.executeAsync(id, "MANUAL");
            
            // 这里可以返回构建ID，前端可以通过构建ID查询构建状态
            return Result.success("构建已开始，请查看构建日志获取详细信息", "build-" + System.currentTimeMillis());
            
        } catch (Exception e) {
            log.error("触发构建失败", e);
            return Result.error("触发构建失败: " + e.getMessage());
        }
    }
    
    /**
     * Webhook触发构建
     */
    @Mapping(value = "/{id}/webhook", method = MethodType.POST)
    public Result<String> webhookTrigger(@Path Long id, @Body String payload) {
        try {
            log.info("Webhook触发构建: applicationId={}, payload={}", id, payload);
            
            // 验证Webhook签名（生产环境需要实现）
            // validateWebhookSignature(payload, signature);
            
            CompletableFuture<PipelineExecutionService.BuildResult> future = 
                pipelineExecutionService.executeAsync(id, "WEBHOOK");
            
            return Result.success("Webhook构建已开始", "build-" + System.currentTimeMillis());
            
        } catch (Exception e) {
            log.error("Webhook触发构建失败", e);
            return Result.error("Webhook触发构建失败: " + e.getMessage());
        }
    }
    
    /**
     * 获取应用构建状态
     */
    @Mapping(value = "/{id}/status", method = MethodType.GET)
    public Result<String> getBuildStatus(@Path Long id) {
        try {
            // 这里应该从数据库查询最新的构建状态
            return Result.success("SUCCESS"); // 模拟返回
        } catch (Exception e) {
            log.error("查询构建状态失败", e);
            return Result.error("查询构建状态失败: " + e.getMessage());
        }
    }
}