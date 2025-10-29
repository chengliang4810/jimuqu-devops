"""
Celery任务 - 异步部署任务
"""
from celery import Celery
from app.config import settings
from app.database import SessionLocal
from app.services.deployment_service import DeploymentService
import logging

# 创建Celery实例
celery_app = Celery(
    "devops_deploy",
    broker=settings.CELERY_BROKER_URL,
    backend=settings.CELERY_RESULT_BACKEND,
    include=['worker.tasks']
)

# Celery配置
celery_app.conf.update(
    task_serializer='json',
    accept_content=['json'],
    result_serializer='json',
    timezone='UTC',
    enable_utc=True,
    # 任务过期时间（秒）
    result_expires=3600,
    # 任务路由
    task_routes={
        'worker.tasks.deploy_project': {'queue': 'deployments'},
    },
    # 任务重试配置
    task_acks_late=True,
    worker_prefetch_multiplier=1,
)

# 日志
logger = logging.getLogger(__name__)


@celery_app.task(bind=True)
def deploy_project(self, deployment_id: int):
    """
    部署项目任务

    Args:
        deployment_id: 部署ID
    """
    db = SessionLocal()
    deployment_service = DeploymentService()

    try:
        logger.info(f"开始执行部署任务: {deployment_id}")

        def log_callback(log_type: str, message: str, timestamp):
            """日志回调函数"""
            # 这里可以添加Redis pub/sub或其他机制来推送日志
            logger.info(f"[{log_type}] {message}")

        # 执行部署
        success = False
        try:
            # 同步执行部署（因为涉及大量I/O操作）
            import asyncio
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            success = loop.run_until_complete(
                deployment_service.execute_deployment(db, deployment_id, log_callback)
            )
            loop.close()
        except Exception as e:
            logger.error(f"部署执行异常: {str(e)}", exc_info=True)
            raise

        if success:
            logger.info(f"部署任务完成: {deployment_id}")
            return {"status": "success", "deployment_id": deployment_id}
        else:
            logger.error(f"部署任务失败: {deployment_id}")
            return {"status": "failed", "deployment_id": deployment_id}

    except Exception as e:
        logger.error(f"部署任务异常: {deployment_id}, error: {str(e)}", exc_info=True)
        # 标记任务失败
        db.query(Deployment).filter(Deployment.id == deployment_id).update({
            "status": Deployment.STATUS_FAILED,
            "error_message": str(e)
        })
        db.commit()

        # 重新抛出异常，Celery会记录任务为FAILURE
        raise

    finally:
        db.close()


@celery_app.task
def cleanup_old_logs():
    """
    清理旧日志的任务（可以设置定时任务执行）
    """
    logger.info("开始清理旧日志...")
    # 这里可以添加清理逻辑，例如删除30天前的部署日志
    pass


@celery_app.task(bind=True)
def test_connection(self, project_id: int):
    """
    测试项目连接（SSH、Docker等）

    Args:
        project_id: 项目ID
    """
    from app.models import Project
    db = SessionLocal()

    try:
        project = db.query(Project).filter(Project.id == project_id).first()
        if not project:
            raise ValueError(f"项目 {project_id} 不存在")

        # 测试Docker连接
        import docker
        docker_client = docker.from_env()
        docker_client.ping()

        # 测试SSH连接
        from app.services.ssh_service import SSHDeployService
        ssh_service = SSHDeployService()

        # 简单测试SSH连接
        ssh = ssh_service._get_ssh_connection(
            project.target_host,
            project.target_port,
            project.target_username,
            project.ssh_key_path,
            project.ssh_password
        )
        ssh.close()

        return {"status": "success", "message": "连接测试通过"}

    except Exception as e:
        logger.error(f"连接测试失败: {str(e)}", exc_info=True)
        return {"status": "failed", "message": str(e)}

    finally:
        db.close()


if __name__ == '__main__':
    celery_app.start()
