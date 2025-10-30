"""
主机管理API路由 - SSH主机的增删改查
"""
from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime
import paramiko
import time

from app.database import get_db
from app.models.host import Host
from app.schemas import HostCreate, HostUpdate, HostResponse, HostListResponse, HostTestConnection
from app.auth import get_current_active_user

router = APIRouter(prefix="/api/hosts", tags=["hosts"])


@router.post("/", response_model=HostResponse, status_code=status.HTTP_201_CREATED)
async def create_host(
    host: HostCreate,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    创建新主机
    """
    # 检查主机名是否已存在
    db_host = db.query(Host).filter(Host.name == host.name).first()
    if db_host:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="主机名称已存在"
        )

    # 检查主机地址和端口组合是否已存在
    existing_host = db.query(Host).filter(
        Host.host == host.host,
        Host.port == host.port,
        Host.username == host.username
    ).first()
    if existing_host:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="该主机连接配置已存在"
        )

    # 创建主机
    db_host = Host(**host.dict())
    db_host.status = "未连接"
    db.add(db_host)
    db.commit()
    db.refresh(db_host)

    return db_host


@router.get("/", response_model=HostListResponse)
async def list_hosts(
    page: int = Query(1, ge=1, description="页码"),
    size: int = Query(10, ge=1, le=100, description="每页数量"),
    search: Optional[str] = Query(None, description="搜索关键词"),
    group: Optional[str] = Query(None, description="主机分组"),
    status: Optional[str] = Query(None, description="连接状态"),
    is_active: Optional[bool] = Query(None, description="是否启用"),
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    获取主机列表
    """
    query = db.query(Host)

    # 应用筛选条件
    if search:
        search_term = f"%{search}%"
        query = query.filter(
            Host.name.ilike(search_term) |
            Host.host.ilike(search_term) |
            Host.description.ilike(search_term) |
            Host.username.ilike(search_term)
        )

    if group:
        query = query.filter(Host.group == group)

    if status:
        query = query.filter(Host.status == status)

    if is_active is not None:
        query = query.filter(Host.is_active == is_active)

    # 计算总数
    total = query.count()

    # 分页
    offset = (page - 1) * size
    hosts = query.order_by(Host.created_at.desc()).offset(offset).limit(size).all()

    pages = (total + size - 1) // size

    return HostListResponse(
        hosts=hosts,
        total=total,
        page=page,
        size=size,
        pages=pages
    )


@router.get("/{host_id}", response_model=HostResponse)
async def get_host(
    host_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    获取单个主机详情
    """
    host = db.query(Host).filter(Host.id == host_id).first()
    if not host:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="主机不存在"
        )

    return host


@router.put("/{host_id}", response_model=HostResponse)
async def update_host(
    host_id: int,
    host_update: HostUpdate,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    更新主机信息
    """
    db_host = db.query(Host).filter(Host.id == host_id).first()
    if not db_host:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="主机不存在"
        )

    # 更新主机字段
    update_data = host_update.dict(exclude_unset=True)

    # 检查主机名是否重复
    if "name" in update_data:
        existing_host = db.query(Host).filter(
            Host.name == update_data["name"],
            Host.id != host_id
        ).first()
        if existing_host:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="主机名称已存在"
            )

    # 检查连接配置是否重复
    if all(key in update_data for key in ["host", "port", "username"]):
        existing_host = db.query(Host).filter(
            Host.host == update_data["host"],
            Host.port == update_data["port"],
            Host.username == update_data["username"],
            Host.id != host_id
        ).first()
        if existing_host:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="该主机连接配置已存在"
            )

    for field, value in update_data.items():
        setattr(db_host, field, value)

    db_host.updated_at = datetime.utcnow()
    db.commit()
    db.refresh(db_host)

    return db_host


@router.delete("/{host_id}")
async def delete_host(
    host_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    删除主机
    """
    db_host = db.query(Host).filter(Host.id == host_id).first()
    if not db_host:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="主机不存在"
        )

    # 检查是否有项目在使用此主机
    from app.models.project import Project
    projects_using_host = db.query(Project).filter(
        Project.target_host == db_host.host,
        Project.target_port == db_host.port,
        Project.target_username == db_host.username
    ).count()

    if projects_using_host > 0:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"有 {projects_using_host} 个项目正在使用此主机，无法删除"
        )

    db.delete(db_host)
    db.commit()

    return {"message": "主机已删除"}


@router.post("/{host_id}/test-connection", response_model=HostTestConnection)
async def test_host_connection(
    host_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    测试主机SSH连接
    """
    host = db.query(Host).filter(Host.id == host_id).first()
    if not host:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="主机不存在"
        )

    start_time = time.time()

    try:
        # 创建SSH客户端
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())

        # 准备认证信息
        auth_kwargs = {
            "hostname": host.host,
            "port": host.port,
            "username": host.username,
            "timeout": 10
        }

        # 优先使用SSH密钥
        if host.ssh_key_path:
            auth_kwargs["key_filename"] = host.ssh_key_path
        elif host.ssh_password:
            auth_kwargs["password"] = host.ssh_password
        else:
            raise ValueError("未配置SSH认证信息")

        # 尝试连接
        ssh.connect(**auth_kwargs)

        # 执行简单命令测试连接
        stdin, stdout, stderr = ssh.exec_command("uname -a && whoami")
        output = stdout.read().decode().strip()
        error = stderr.read().decode().strip()

        ssh.close()

        response_time = time.time() - start_time

        if error and not output:
            raise Exception(f"命令执行失败: {error}")

        # 解析系统信息
        os_info = {
            "uname": output,
            "user": None,
            "os_type": None,
            "os_version": None
        }

        if output:
            lines = output.split('\n')
            if len(lines) > 1:
                os_info["user"] = lines[-1].strip()

            # 解析操作系统信息
            uname_line = lines[0] if lines else ""
            if "Linux" in uname_line:
                os_info["os_type"] = "Linux"
                parts = uname_line.split()
                if len(parts) >= 3:
                    os_info["os_version"] = parts[2]
            elif "Darwin" in uname_line:
                os_info["os_type"] = "macOS"
                parts = uname_line.split()
                if len(parts) >= 3:
                    os_info["os_version"] = parts[2]
            elif "MINGW" in uname_line or "MSYS" in uname_line:
                os_info["os_type"] = "Windows"
                parts = uname_line.split()
                if len(parts) >= 3:
                    os_info["os_version"] = parts[2]

        # 更新主机连接状态
        host.status = "正常"
        host.last_connected_at = datetime.utcnow()
        host.os_type = os_info.get("os_type")
        host.os_version = os_info.get("os_version")
        db.commit()

        return HostTestConnection(
            success=True,
            message="连接成功",
            response_time=response_time,
            os_info=os_info
        )

    except Exception as e:
        # 更新主机连接状态
        host.status = "异常"
        host.last_connected_at = datetime.utcnow()
        db.commit()

        return HostTestConnection(
            success=False,
            message=f"连接失败: {str(e)}",
            response_time=time.time() - start_time
        )


@router.get("/{host_id}/groups")
async def get_host_groups(
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    获取所有主机分组
    """
    groups = db.query(Host.group).filter(
        Host.group.isnot(None),
        Host.group != ""
    ).distinct().all()

    group_list = [group[0] for group in groups if group[0]]
    return {"groups": sorted(group_list)}


@router.post("/{host_id}/toggle-status")
async def toggle_host_status(
    host_id: int,
    db: Session = Depends(get_db),
    current_user: dict = Depends(get_current_active_user)
):
    """
    切换主机启用/禁用状态
    """
    host = db.query(Host).filter(Host.id == host_id).first()
    if not host:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="主机不存在"
        )

    host.is_active = not host.is_active
    host.updated_at = datetime.utcnow()
    db.commit()

    return {
        "id": host.id,
        "is_active": host.is_active,
        "message": f"主机已{'启用' if host.is_active else '禁用'}"
    }