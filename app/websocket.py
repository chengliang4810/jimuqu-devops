"""
WebSocket实时日志推送 - 部署过程中实时查看日志
"""
import json
from datetime import datetime
from typing import Dict, Set
import logging

from fastapi import WebSocket, WebSocketDisconnect
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from app.database import get_db
from app.auth import get_current_active_user

logger = logging.getLogger(__name__)

router = APIRouter()

# 管理WebSocket连接：deployment_id -> set of WebSocket connections
active_connections: Dict[int, Set[WebSocket]] = {}


class ConnectionManager:
    """WebSocket连接管理器"""

    async def connect(self, websocket: WebSocket, deployment_id: int):
        """建立WebSocket连接"""
        await websocket.accept()

        if deployment_id not in active_connections:
            active_connections[deployment_id] = set()

        active_connections[deployment_id].add(websocket)
        logger.info(f"WebSocket连接已建立: deployment_id={deployment_id}, "
                   f"total_connections={len(active_connections[deployment_id])}")

    def disconnect(self, websocket: WebSocket, deployment_id: int):
        """断开WebSocket连接"""
        if deployment_id in active_connections:
            active_connections[deployment_id].discard(websocket)

            # 如果没有连接了，删除该deployment的记录
            if not active_connections[deployment_id]:
                del active_connections[deployment_id]

        logger.info(f"WebSocket连接已断开: deployment_id={deployment_id}")

    async def send_log(self, deployment_id: int, log_type: str, message: str, timestamp: datetime):
        """发送日志到所有连接的客户端"""
        if deployment_id not in active_connections:
            return

        log_data = {
            "type": "log",
            "deployment_id": deployment_id,
            "log_type": log_type,  # info, success, error
            "message": message,
            "timestamp": timestamp.isoformat()
        }

        message_str = json.dumps(log_data, ensure_ascii=False)

        # 发送到所有连接的客户端
        disconnected_connections = set()
        for connection in active_connections[deployment_id]:
            try:
                await connection.send_text(message_str)
            except Exception as e:
                logger.error(f"发送日志失败: {e}")
                disconnected_connections.add(connection)

        # 清理断开的连接
        for connection in disconnected_connections:
            self.disconnect(connection, deployment_id)


# 创建全局连接管理器实例
manager = ConnectionManager()


@router.websocket("/ws/deployments/{deployment_id}")
async def websocket_endpoint(
    websocket: WebSocket,
    deployment_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    WebSocket端点 - 实时日志推送

    使用方法：
    1. 客户端连接到: ws://host:port/ws/deployments/{deployment_id}
    2. 认证: 使用HTTP Basic Auth（与API相同的用户名密码）
    3. 接收实时日志推送
    """
    await manager.connect(websocket, deployment_id)

    try:
        # 发送欢迎消息
        welcome_msg = {
            "type": "welcome",
            "deployment_id": deployment_id,
            "message": "WebSocket连接已建立，开始接收实时日志"
        }
        await websocket.send_text(json.dumps(welcome_msg, ensure_ascii=False))

        # 获取部署历史日志
        from app.models import Deployment
        deployment = db.query(Deployment).filter(Deployment.id == deployment_id).first()

        if deployment and deployment.logs:
            # 发送历史日志
            history_msg = {
                "type": "history",
                "deployment_id": deployment_id,
                "logs": deployment.logs,
                "status": deployment.status
            }
            await websocket.send_text(json.dumps(history_msg, ensure_ascii=False))

        # 保持连接
        while True:
            # 这里可以处理来自客户端的消息（例如心跳检测）
            await websocket.receive_text()

    except WebSocketDisconnect:
        manager.disconnect(websocket, deployment_id)
    except Exception as e:
        logger.error(f"WebSocket异常: {e}", exc_info=True)
        manager.disconnect(websocket, deployment_id)


async def send_deployment_log(
    deployment_id: int,
    log_type: str,
    message: str,
    timestamp: datetime = None
):
    """
    发送部署日志（供其他模块调用）

    Args:
        deployment_id: 部署ID
        log_type: 日志类型 (info, success, error)
        message: 日志消息
        timestamp: 时间戳（默认当前时间）
    """
    if timestamp is None:
        timestamp = datetime.utcnow()

    await manager.send_log(deployment_id, log_type, message, timestamp)
