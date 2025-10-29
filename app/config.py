"""
配置管理 - 统一管理所有配置参数
"""
import os
from typing import Optional
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """应用配置"""

    # 应用基础配置
    APP_NAME: str = "DevOps部署平台"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = False

    # 服务器配置
    HOST: str = "0.0.0.0"
    PORT: int = 8000

    # 管理员认证（通过环境变量配置）
    ADMIN_USERNAME: str = os.getenv("ADMIN_USERNAME", "admin")
    ADMIN_PASSWORD: str = os.getenv("ADMIN_PASSWORD", "admin123")

    # 数据库配置
    DATABASE_URL: str = os.getenv(
        "DATABASE_URL",
        "sqlite:///./devops.db"
    )

    # Redis配置
    REDIS_URL: str = os.getenv(
        "REDIS_URL",
        "redis://localhost:6379/0"
    )

    # Celery配置
    CELERY_BROKER_URL: str = os.getenv(
        "CELERY_BROKER_URL",
        "redis://localhost:6379/1"
    )
    CELERY_RESULT_BACKEND: str = os.getenv(
        "CELERY_RESULT_BACKEND",
        "redis://localhost:6379/2"
    )

    # GitHub/GitLab Webhook Secret（可选）
    GITHUB_WEBHOOK_SECRET: Optional[str] = os.getenv("GITHUB_WEBHOOK_SECRET")
    GITLAB_WEBHOOK_SECRET: Optional[str] = os.getenv("gitlab_WEBHOOK_SECRET")

    # Docker配置
    DOCKER_SOCKET: str = "/var/run/docker.sock"

    # 日志配置
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "INFO")
    LOG_FILE: str = os.getenv("LOG_FILE", "./logs/devops.log")

    class Config:
        env_file = ".env"


# 创建全局配置实例
settings = Settings()


# 创建必要目录
def create_directories():
    """创建应用运行时需要的目录"""
    directories = [
        "./logs",
        "./data",
        "./static/uploads",
    ]

    for directory in directories:
        os.makedirs(directory, exist_ok=True)


# 初始化时创建目录
create_directories()
