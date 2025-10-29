"""
Webhook接收器 - 处理Git仓库的推送事件，自动触发部署
"""
import hmac
import hashlib
from fastapi import APIRouter, HTTPException, Depends, Request
from sqlalchemy.orm import Session
from typing import Optional

from app.database import get_db
from app.models import Project
from app.schemas import WebhookRequest
from app.auth import get_current_active_user
from worker.tasks import deploy_project

router = APIRouter(prefix="/api/webhook", tags=["webhook"])


def verify_github_signature(
    payload: bytes,
    secret: str,
    signature: Optional[str]
) -> bool:
    """
    验证GitHub Webhook签名

    Args:
        payload: 请求体
        secret: Webhook Secret
        signature: GitHub signature header

    Returns:
        是否验证通过
    """
    if not signature:
        return False

    # GitHub签名前缀
    expected_signature = 'sha256='

    if not signature.startswith(expected_signature):
        return False

    signature = signature[len(expected_signature):]

    # 计算HMAC
    mac = hmac.new(
        secret.encode('utf-8'),
        msg=payload,
        digestmod=hashlib.sha256
    )
    computed_signature = mac.hexdigest()

    # 安全比较（防止时序攻击）
    return hmac.compare_digest(signature, computed_signature)


def verify_gitlab_token(
    token: Optional[str],
    secret: str
) -> bool:
    """
    验证GitLab Webhook Token

    Args:
        token: GitLab token header
        secret: Webhook Secret

    Returns:
        是否验证通过
    """
    if not token or not secret:
        return False

    return hmac.compare_digest(token, secret)


@router.post("/github/{project_id}")
async def github_webhook(
    project_id: int,
    request: Request,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    GitHub Webhook接收器

    支持的event: push
    """
    # 获取项目
    project = db.query(Project).filter(Project.id == project_id).first()
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")

    # 获取请求体和签名
    payload = await request.body()
    signature = request.headers.get("X-Hub-Signature-256")

    # 验证签名
    if project.webhook_secret:
        if not verify_github_signature(payload, project.webhook_secret, signature):
            raise HTTPException(status_code=401, detail="签名验证失败")

    import json
    try:
        webhook_data = json.loads(payload)
    except json.JSONDecodeError:
        raise HTTPException(status_code=400, detail="无效的JSON")

    # 检查是否为push事件
    event = request.headers.get("X-GitHub-Event")
    if event != "push":
        return {
            "status": "ignored",
            "message": f"不支持的事件类型: {event}"
        }

    # 提取Git信息
    ref = webhook_data.get("ref", "")
    repository = webhook_data.get("repository", {})
    commits = webhook_data.get("commits", [])
    pusher = webhook_data.get("pusher", {})

    # 获取分支名
    branch = ref.split("/")[-1] if ref else "main"

    # 如果是最新的commit
    if not commits:
        return {
            "status": "ignored",
            "message": "没有提交信息"
        }

    # 使用最新的提交
    latest_commit = commits[-1]
    commit_hash = latest_commit["id"]
    commit_message = latest_commit.get("message", "")
    author = latest_commit.get("author", {}).get("name", "")

    # 检查分支是否匹配（可选）
    # if project.branch and branch != project.branch:
    #     return {
    #         "status": "ignored",
    #         "message": f"分支不匹配: {branch} != {project.branch}"
    #     }

    # 创建部署任务
    from app.services.deployment_service import DeploymentService

    deployment_service = DeploymentService()
    deployment = await deployment_service.create_deployment(
        db=db,
        project_id=project_id,
        commit_hash=commit_hash,
        commit_message=commit_message,
        author=author,
        branch=branch,
        triggered_by="webhook",
        webhook_payload=webhook_data
    )

    # 提交Celery部署任务
    task = deploy_project.delay(deployment.id)

    return {
        "status": "success",
        "deployment_id": deployment.id,
        "task_id": task.id,
        "message": f"自动部署已触发: {commit_hash[:8]}"
    }


@router.post("/gitlab/{project_id}")
async def gitlab_webhook(
    project_id: int,
    request: Request,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    GitLab Webhook接收器

    支持的event: Push Hook
    """
    # 获取项目
    project = db.query(Project).filter(Project.id == project_id).first()
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")

    # 验证Token
    token = request.headers.get("X-Gitlab-Token")
    if project.webhook_secret:
        if not verify_gitlab_token(token, project.webhook_secret):
            raise HTTPException(status_code=401, detail="Token验证失败")

    import json
    try:
        webhook_data = await request.json()
    except json.JSONDecodeError:
        raise HTTPException(status_code=400, detail="无效的JSON")

    # 检查事件类型
    event = request.headers.get("X-Gitlab-Event")
    if event != "Push Hook":
        return {
            "status": "ignored",
            "message": f"不支持的事件类型: {event}"
        }

    # 提取Git信息
    ref = webhook_data.get("ref", "")
    commits = webhook_data.get("commits", [])
    repository = webhook_data.get("repository", {})
    user = webhook_data.get("user", {})

    # 获取分支名
    branch = ref.split("/")[-1] if ref else "main"

    # 如果有commit
    if not commits:
        return {
            "status": "ignored",
            "message": "没有提交信息"
        }

    # 使用最新的提交
    latest_commit = commits[-1]
    commit_hash = latest_commit["id"]
    commit_message = latest_commit.get("message", "")
    author = latest_commit.get("author", {}).get("name", user.get("name", ""))

    # 创建部署任务
    from app.services.deployment_service import DeploymentService

    deployment_service = DeploymentService()
    deployment = await deployment_service.create_deployment(
        db=db,
        project_id=project_id,
        commit_hash=commit_hash,
        commit_message=commit_message,
        author=author,
        branch=branch,
        triggered_by="webhook",
        webhook_payload=webhook_data
    )

    # 提交Celery部署任务
    task = deploy_project.delay(deployment.id)

    return {
        "status": "success",
        "deployment_id": deployment.id,
        "task_id": task.id,
        "message": f"自动部署已触发: {commit_hash[:8]}"
    }
