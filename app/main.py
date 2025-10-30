"""
FastAPI主应用 - DevOps部署平台的入口
"""
from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.staticfiles import StaticFiles
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import HTTPBasic
import logging
from pathlib import Path
from datetime import datetime, timedelta
import jwt
from pydantic import BaseModel

from app.config import settings, create_directories
from app.database import get_db, create_tables
from app.auth import get_current_active_user
from app.api import projects, deployments, webhook, hosts
from app.websocket import router as websocket_router

# 创建应用
app = FastAPI(
    title=settings.APP_NAME,
    version=settings.APP_VERSION,
    description="基于FastAPI的DevOps自动化部署平台",
    docs_url="/docs" if settings.DEBUG else None,  # 生产环境关闭文档
    redoc_url="/redoc" if settings.DEBUG else None
)

# 允许跨域（开发环境）
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"] if settings.DEBUG else ["http://localhost"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 静态文件（前端页面）
static_dir = Path(__file__).parent.parent / "static"
if static_dir.exists():
    app.mount("/static", StaticFiles(directory=str(static_dir)), name="static")

# 包含API路由
app.include_router(projects.router)
app.include_router(deployments.router)
app.include_router(webhook.router)
app.include_router(hosts.router)
app.include_router(websocket_router)


# 登录模型
class LoginRequest(BaseModel):
    userName: str
    password: str


class LoginToken(BaseModel):
    token: str
    refreshToken: str


class UserInfo(BaseModel):
    userId: str
    userName: str
    roles: list
    buttons: list


# 登录API
@app.post("/auth/login")
async def login(request: LoginRequest):
    """用户登录"""
    # 验证用户名和密码
    if request.userName != settings.ADMIN_USERNAME or request.password != settings.ADMIN_PASSWORD:
        return JSONResponse(
            status_code=200,
            content={
                "code": "1001",
                "msg": "用户名或密码错误",
                "data": None
            }
        )

    # 生成token
    expire = datetime.utcnow() + timedelta(days=7)
    refresh_expire = datetime.utcnow() + timedelta(days=30)

    token_data = {
        "sub": request.userName,
        "exp": expire
    }
    refresh_token_data = {
        "sub": request.userName,
        "exp": refresh_expire,
        "type": "refresh"
    }

    # 使用简单密钥（生产环境应该使用更安全的方式）
    secret_key = "jimuqu-devops-secret-key"

    token = jwt.encode(token_data, secret_key, algorithm="HS256")
    refresh_token = jwt.encode(refresh_token_data, secret_key, algorithm="HS256")

    return JSONResponse(
        content={
            "code": "0000",
            "msg": "登录成功",
            "data": {
                "token": token,
                "refreshToken": refresh_token
            }
        }
    )


# 获取用户信息API
@app.get("/auth/getUserInfo")
async def get_user_info():
    """获取当前用户信息"""
    return JSONResponse(
        content={
            "code": "0000",
            "msg": "获取成功",
            "data": {
                "userId": "admin",
                "userName": settings.ADMIN_USERNAME,
                "roles": ["R_SUPER", "R_ADMIN"],
                "buttons": []
            }
        }
    )


# 挂载根路径到前端页面
@app.get("/", include_in_schema=False)
async def root():
    """根路径 - 返回前端页面"""
    from fastapi.responses import FileResponse
    index_file = static_dir / "index.html"
    if index_file.exists():
        return FileResponse(index_file)
    return JSONResponse({
        "message": "DevOps部署平台 API",
        "version": settings.APP_VERSION,
        "docs": "/docs"
    })


@app.get("/api/health", include_in_schema=False)
async def health_check():
    """健康检查端点"""
    return {
        "status": "healthy",
        "version": settings.APP_VERSION,
        "environment": "development" if settings.DEBUG else "production"
    }


@app.get("/api/me", include_in_schema=False)
async def get_current_user_info(current_user: dict = Depends(get_current_active_user)):
    """获取当前用户信息"""
    return current_user


# 异常处理
@app.exception_handler(Exception)
async def global_exception_handler(request, exc):
    """全局异常处理器"""
    logging.error(f"未处理的异常: {str(exc)}", exc_info=True)

    return JSONResponse(
        status_code=500,
        content={
            "error": "内部服务器错误",
            "message": "请联系管理员" if not settings.DEBUG else str(exc)
        }
    )


# 启动事件
@app.on_event("startup")
async def startup_event():
    """应用启动时执行"""
    # 创建数据库表
    create_tables()

    # 创建必要目录
    create_directories()

    # 初始化日志
    logging.basicConfig(
        level=getattr(logging, settings.LOG_LEVEL),
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.FileHandler(settings.LOG_FILE, encoding='utf-8'),
            logging.StreamHandler()
        ]
    )

    logging.info(f"{settings.APP_NAME} v{settings.APP_VERSION} 启动成功")
    logging.info(f"API文档: http://{settings.HOST}:{settings.PORT}/docs")


if __name__ == "__main__":
    import uvicorn

    # 启动应用
    uvicorn.run(
        "app.main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=settings.DEBUG,
        log_level=settings.LOG_LEVEL.lower()
    )
