"""
管理员认证 - 基于环境变量的单用户认证
"""
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBasic, HTTPBasicCredentials
from app.config import settings

security = HTTPBasic()


def get_current_user(credentials: HTTPBasicCredentials = Depends(security)):
    """
    验证管理员用户名和密码
    """
    is_correct_username = credentials.username == settings.ADMIN_USERNAME
    is_correct_password = credentials.password == settings.ADMIN_PASSWORD

    if not (is_correct_username and is_correct_password):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="无效的凭据",
            headers={"WWW-Authenticate": "Basic"},
        )
    return {
        "username": credentials.username,
        "is_admin": True
    }


def get_current_active_user(current_user: dict = Depends(get_current_user)):
    """
    获取当前激活的用户
    """
    return current_user
