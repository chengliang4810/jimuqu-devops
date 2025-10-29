"""
部署API路由 - 部署记录的管理和触发
"""
from fastapi import APIRouter, Depends, HTTPException, Query, BackgroundTasks
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime, timedelta

from app.database import get_db
from app.models import Project, Deployment
from app.schemas import DeploymentCreate, DeploymentResponse, DeploymentListResponse
from app.auth import get_current_active_user
from app.services.deployment_service import DeploymentService

router = APIRouter(prefix="/api/deployments", tags=["deployments"])


@router.post("/", response_model=DeploymentResponse)
async def create_deployment(
    deployment: DeploymentCreate,
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    手动创建部署任务
    """
    # 检查项目是否存在
    project = db.query(Project).filter(Project.id == deployment.project_id).first()
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")

    # 检查项目是否激活
    if not project.is_active:
        raise HTTPException(status_code=400, detail="项目未激活")

    deployment_service = DeploymentService()

    # 创建部署记录
    db_deployment = await deployment_service.create_deployment(
        db=db,
        project_id=deployment.project_id,
        commit_hash=deployment.commit_hash,
        commit_message=deployment.commit_message,
        author=deployment.author,
        branch=deployment.branch,
        triggered_by=deployment.triggered_by,
        webhook_payload=deployment.webhook_payload
    )

    return db_deployment


@router.post("/{deployment_id}/execute")
async def execute_deployment(
    deployment_id: int,
    background_tasks: BackgroundTasks,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    手动执行部署
    """
    # 获取部署记录
    deployment = db.query(Deployment).filter(Deployment.id == deployment_id).first()
    if not deployment:
        raise HTTPException(status_code=404, detail="部署记录不存在")

    # 检查部署状态
    if deployment.status in [Deployment.STATUS_RUNNING]:
        raise HTTPException(status_code=400, detail="部署正在进行中")

    # 提交Celery任务
    from worker.tasks import deploy_project
    task = deploy_project.delay(deployment_id)

    return {
        "task_id": task.id,
        "deployment_id": deployment_id,
        "status": "submitted",
        "message": "部署任务已提交到队列"
    }


@router.get("/", response_model=DeploymentListResponse)
async def list_deployments(
    project_id: int = Query(None, ge=1),
    skip: int = Query(0, ge=0),
    limit: int = Query(50, ge=1, le=100),
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    获取部署列表
    """
    deployment_service = DeploymentService()
    deployments, total = await deployment_service.get_deployments(
        db=db,
        project_id=project_id,
        limit=limit,
        offset=skip
    )

    pages = (total + limit - 1) // limit

    return {
        "deployments": deployments,
        "total": total,
        "page": skip // limit + 1,
        "size": limit,
        "pages": pages
    }


@router.get("/{deployment_id}", response_model=DeploymentResponse)
async def get_deployment(
    deployment_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    获取单个部署记录详情
    """
    deployment = db.query(Deployment).filter(Deployment.id == deployment_id).first()
    if not deployment:
        raise HTTPException(status_code=404, detail="部署记录不存在")

    return deployment


@router.get("/stats/dashboard")
async def get_dashboard_stats(
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    获取仪表盘统计数据
    """
    # 总项目数
    total_projects = db.query(Project).count()
    active_projects = db.query(Project).filter(Project.is_active == True).count()

    # 今日部署数
    today = datetime.utcnow().date()
    today_deployments = db.query(Deployment).filter(
        Deployment.start_time >= today
    ).count()

    # 成功率和平均耗时
    last_30_days = today - timedelta(days=30)
    deployments_30d = db.query(Deployment).filter(
        Deployment.start_time >= last_30_days
    ).all()

    success_count = sum(1 for d in deployments_30d if d.is_success)
    success_rate = (success_count / len(deployments_30d) * 100) if deployments_30d else 0

    successful_deployments = [d for d in deployments_30d if d.is_success and d.duration]
    avg_duration = sum(d.duration for d in successful_deployments) // len(successful_deployments) if successful_deployments else 0

    return {
        "total_projects": total_projects,
        "active_projects": active_projects,
        "today_deployments": today_deployments,
        "success_rate": round(success_rate, 2),
        "average_duration": avg_duration
    }
