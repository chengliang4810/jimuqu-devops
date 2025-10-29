"""
Pydantic模型 - API请求/响应模型
"""
from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from app.models.deployment import Deployment


# ===== 项目模型 =====
class ProjectBase(BaseModel):
    """项目基础模型"""
    name: str = Field(..., description="项目名称")
    description: Optional[str] = None
    git_url: str = Field(..., description="Git仓库地址")
    language: str = Field(..., description="开发语言：java, python, node, go")
    build_command: Optional[str] = Field(None, description="自定义构建命令")
    deploy_path: str = Field(..., description="部署路径")
    restart_command: Optional[str] = Field(None, description="重启命令")
    target_host: str = Field(..., description="目标主机")
    target_port: int = Field(22, description="SSH端口")
    target_username: str = Field(..., description="目标主机用户名")
    ssh_key_path: Optional[str] = Field(None, description="SSH私钥路径")
    ssh_password: Optional[str] = Field(None, description="SSH密码")
    webhook_secret: Optional[str] = Field(None, description="Webhook密钥")


class ProjectCreate(ProjectBase):
    """创建项目"""
    pass


class ProjectUpdate(BaseModel):
    """更新项目"""
    name: Optional[str] = None
    description: Optional[str] = None
    git_url: Optional[str] = None
    language: Optional[str] = None
    build_command: Optional[str] = None
    deploy_path: Optional[str] = None
    restart_command: Optional[str] = None
    target_host: Optional[str] = None
    target_port: Optional[int] = None
    target_username: Optional[str] = None
    ssh_key_path: Optional[str] = None
    ssh_password: Optional[str] = None
    webhook_secret: Optional[str] = None
    is_active: Optional[bool] = None


class ProjectResponse(ProjectBase):
    """项目响应模型"""
    id: int
    is_active: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


# ===== 部署模型 =====
class DeploymentBase(BaseModel):
    """部署基础模型"""
    project_id: int
    commit_hash: str
    commit_message: Optional[str] = None
    author: Optional[str] = None
    branch: str = "main"
    triggered_by: str = "manual"
    webhook_payload: Optional[dict] = None


class DeploymentCreate(DeploymentBase):
    """创建部署"""
    pass


class DeploymentResponse(DeploymentBase):
    """部署响应模型"""
    id: int
    status: str
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    duration: Optional[int] = None
    logs: Optional[str] = None
    error_message: Optional[str] = None
    build_time: Optional[int] = None
    deploy_time: Optional[int] = None

    class Config:
        from_attributes = True


class DeploymentListResponse(BaseModel):
    """部署列表响应"""
    deployments: List[DeploymentResponse]
    total: int
    page: int
    size: int
    pages: int


# ===== Webhook模型 =====
class WebhookRequest(BaseModel):
    """Webhook请求模型"""
    repository: dict
    ref: str
    commits: List[dict]
    pusher: dict


# ===== 实时日志模型 =====
class LogMessage(BaseModel):
    """日志消息模型"""
    type: str  # info, error, success
    message: str
    timestamp: datetime


# ===== 统计模型 =====
class DashboardStats(BaseModel):
    """仪表盘统计"""
    total_projects: int
    active_projects: int
    today_deployments: int
    success_rate: float
    average_duration: int
