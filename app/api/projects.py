"""
项目API路由 - 项目的增删改查
"""
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List

from app.database import get_db
from app.models import Project
from app.schemas import ProjectCreate, ProjectUpdate, ProjectResponse
from app.auth import get_current_active_user

router = APIRouter(prefix="/api/projects", tags=["projects"])


@router.post("/", response_model=ProjectResponse)
async def create_project(
    project: ProjectCreate,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    创建新项目
    """
    # 检查项目名是否已存在
    db_project = db.query(Project).filter(Project.name == project.name).first()
    if db_project:
        raise HTTPException(status_code=400, detail="项目名称已存在")

    # 创建项目
    db_project = Project(**project.dict())
    db.add(db_project)
    db.commit()
    db.refresh(db_project)

    return db_project


@router.get("/", response_model=List[ProjectResponse])
async def list_projects(
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=100),
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    获取项目列表
    """
    projects = db.query(Project).offset(skip).limit(limit).all()
    return projects


@router.get("/{project_id}", response_model=ProjectResponse)
async def get_project(
    project_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    获取单个项目详情
    """
    project = db.query(Project).filter(Project.id == project_id).first()
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")

    return project


@router.put("/{project_id}", response_model=ProjectResponse)
async def update_project(
    project_id: int,
    project_update: ProjectUpdate,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    更新项目信息
    """
    db_project = db.query(Project).filter(Project.id == project_id).first()
    if not db_project:
        raise HTTPException(status_code=404, detail="项目不存在")

    # 更新项目字段
    update_data = project_update.dict(exclude_unset=True)
    for field, value in update_data.items():
        setattr(db_project, field, value)

    db_project.updated_at = datetime.utcnow()
    db.commit()
    db.refresh(db_project)

    return db_project


@router.delete("/{project_id}")
async def delete_project(
    project_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    删除项目
    """
    db_project = db.query(Project).filter(Project.id == project_id).first()
    if not db_project:
        raise HTTPException(status_code=404, detail="项目不存在")

    db.delete(db_project)
    db.commit()

    return {"message": "项目已删除"}


@router.post("/{project_id}/test-connection")
async def test_project_connection(
    project_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    测试项目连接（Docker、SSH等）
    """
    from worker.tasks import test_connection

    # 异步执行连接测试
    task = test_connection.delay(project_id)

    return {
        "task_id": task.id,
        "status": "pending",
        "message": "连接测试已提交"
    }


from datetime import datetime
