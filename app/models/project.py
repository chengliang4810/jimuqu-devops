"""
项目模型 - 定义项目的数据库结构
"""
from sqlalchemy import Column, Integer, String, Text, DateTime, Boolean, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relationship
from datetime import datetime
from app.config import settings

Base = declarative_base()


class Project(Base):
    """项目模型"""

    __tablename__ = "projects"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, index=True, nullable=False)
    description = Column(Text, nullable=True)
    git_url = Column(String(500), nullable=False)

    # 部署配置
    language = Column(String(20), nullable=False)  # java, python, node, go
    build_command = Column(String(200), nullable=True)  # 自定义构建命令
    deploy_path = Column(String(200), nullable=False)  # 目标主机部署路径
    restart_command = Column(String(200), nullable=True)  # 重启命令

    # 目标主机配置
    target_host = Column(String(100), nullable=False)
    target_port = Column(Integer, default=22)
    target_username = Column(String(50), nullable=False)
    # SSH私钥路径（建议用密钥文件而非密码）
    ssh_key_path = Column(String(200), nullable=True)
    ssh_password = Column(String(100), nullable=True)

    # Webhook配置
    webhook_secret = Column(String(100), nullable=True)

    # 状态
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    # 关联部署记录
    deployments = relationship("Deployment", back_populates="project")

    def __repr__(self):
        return f"<Project {self.name}>"
