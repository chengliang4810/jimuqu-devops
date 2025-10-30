"""
主机管理模型 - 定义SSH主机的数据库结构
用于管理自动化部署的目标主机
"""
from sqlalchemy import Column, Integer, String, Text, DateTime, Boolean
from sqlalchemy.ext.declarative import declarative_base
from datetime import datetime

Base = declarative_base()


class Host(Base):
    """主机模型 - SSH连接信息管理"""

    __tablename__ = "hosts"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, index=True, nullable=False, comment="主机名称")
    description = Column(Text, nullable=True, comment="主机描述")

    # SSH连接配置
    host = Column(String(100), nullable=False, comment="主机地址(IP或域名)")
    port = Column(Integer, default=22, nullable=False, comment="SSH端口")
    username = Column(String(50), nullable=False, comment="SSH用户名")

    # 认证方式：优先使用SSH密钥，其次使用密码
    ssh_key_path = Column(String(500), nullable=True, comment="SSH私钥路径")
    ssh_password = Column(String(200), nullable=True, comment="SSH密码(加密存储)")

    # 主机标签和分组
    tags = Column(String(200), nullable=True, comment="主机标签,逗号分隔")
    group = Column(String(50), nullable=True, comment="主机分组")

    # 连接状态
    status = Column(String(20), default="未连接", comment="连接状态: 未连接/正常/异常")
    last_connected_at = Column(DateTime, nullable=True, comment="最后连接时间")

    # 系统信息（可选，连接后自动获取）
    os_type = Column(String(50), nullable=True, comment="操作系统类型")
    os_version = Column(String(100), nullable=True, comment="操作系统版本")

    # 状态控制
    is_active = Column(Boolean, default=True, comment="是否启用")
    created_at = Column(DateTime, default=datetime.utcnow, comment="创建时间")
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, comment="更新时间")

    def __repr__(self):
        return f"<Host {self.name}@{self.host}:{self.port}>"
