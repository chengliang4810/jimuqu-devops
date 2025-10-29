"""
部署记录模型 - 存储每次部署的信息
"""
from sqlalchemy import Column, Integer, String, Text, DateTime, Boolean, ForeignKey, JSON
from sqlalchemy.orm import relationship
from datetime import datetime
from app.models.project import Base


class Deployment(Base):
    """部署记录模型"""

    __tablename__ = "deployments"

    id = Column(Integer, primary_key=True, index=True)
    project_id = Column(Integer, ForeignKey("projects.id"), nullable=False)

    # Git信息
    commit_hash = Column(String(40), nullable=False)
    commit_message = Column(String(500), nullable=True)
    author = Column(String(100), nullable=True)
    branch = Column(String(100), default="main")

    # 部署状态
    STATUS_PENDING = "pending"
    STATUS_RUNNING = "running"
    STATUS_SUCCESS = "success"
    STATUS_FAILED = "failed"

    status = Column(String(20), default=STATUS_PENDING)
    start_time = Column(DateTime, nullable=True)
    end_time = Column(DateTime, nullable=True)
    duration = Column(Integer, nullable=True)  # 秒

    # 部署结果
    logs = Column(Text, nullable=True)  # 实时日志
    error_message = Column(Text, nullable=True)

    # 部署统计
    build_time = Column(Integer, nullable=True)  # 构建耗时（秒）
    deploy_time = Column(Integer, nullable=True)  # 部署耗时（秒）

    # Webhook信息
    triggered_by = Column(String(100), default="manual")  # manual, webhook
    webhook_payload = Column(JSON, nullable=True)

    # 关联项目
    project = relationship("Project", back_populates="deployments")

    def __repr__(self):
        return f"<Deployment {self.project.name} - {self.status}>"

    @property
    def is_running(self):
        return self.status == self.STATUS_RUNNING

    @property
    def is_success(self):
        return self.status == self.STATUS_SUCCESS

    @property
    def is_failed(self):
        return self.status == self.STATUS_FAILED
