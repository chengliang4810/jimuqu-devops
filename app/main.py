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

from app.config import settings, create_directories
from app.database import get_db, create_tables
from app.auth import get_current_active_user
from app.api import projects, deployments, webhook
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
app.include_router(websocket_router)

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
