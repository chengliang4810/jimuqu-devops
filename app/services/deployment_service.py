"""
部署服务 - 整合编译和部署流程
"""
import os
import tempfile
import logging
from datetime import datetime
from sqlalchemy.orm import Session
from typing import Optional, Callable

from app.models import Project, Deployment
from app.services.docker_service import DockerCompileService
from app.services.ssh_service import SSHDeployService

logger = logging.getLogger(__name__)


class DeploymentService:
    """部署服务"""

    def __init__(self):
        self.docker_service = DockerCompileService()
        self.ssh_service = SSHDeployService()

    async def create_deployment(
        self,
        db: Session,
        project_id: int,
        commit_hash: str,
        commit_message: Optional[str] = None,
        author: Optional[str] = None,
        branch: str = "main",
        triggered_by: str = "manual",
        webhook_payload: Optional[dict] = None,
        log_callback: Optional[Callable] = None
    ) -> Deployment:
        """
        创建部署任务

        Args:
            db: 数据库会话
            project_id: 项目ID
            commit_hash: 提交哈希
            commit_message: 提交信息
            author: 提交作者
            branch: 分支
            triggered_by: 触发方式
            webhook_payload: Webhook载荷
            log_callback: 日志回调函数

        Returns:
            Deployment对象
        """
        # 获取项目信息
        project = db.query(Project).filter(Project.id == project_id).first()
        if not project:
            raise ValueError(f"项目 {project_id} 不存在")

        # 创建部署记录
        deployment = Deployment(
            project_id=project_id,
            commit_hash=commit_hash,
            commit_message=commit_message,
            author=author,
            branch=branch,
            triggered_by=triggered_by,
            webhook_payload=webhook_payload,
            status=Deployment.STATUS_PENDING,
            start_time=datetime.utcnow()
        )

        db.add(deployment)
        db.commit()
        db.refresh(deployment)

        return deployment

    async def execute_deployment(
        self,
        db: Session,
        deployment_id: int,
        log_callback: Optional[Callable] = None
    ) -> bool:
        """
        执行部署

        Args:
            db: 数据库会话
            deployment_id: 部署ID
            log_callback: 日志回调函数

        Returns:
            是否成功
        """
        # 获取部署记录
        deployment = db.query(Deployment).filter(Deployment.id == deployment_id).first()
        if not deployment:
            raise ValueError(f"部署 {deployment_id} 不存在")

        project = deployment.project
        all_logs = []

        try:
            # 更新状态为运行中
            deployment.status = Deployment.STATUS_RUNNING
            db.commit()

            def append_log(log_type: str, message: str):
                """追加日志"""
                timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                log_entry = f"[{timestamp}] [{log_type.upper()}] {message}"
                all_logs.append(log_entry)

                # 调用回调函数（用于WebSocket推送）
                if log_callback:
                    log_callback(log_type, message, datetime.now())

            append_log("info", f"开始部署项目: {project.name}")
            append_log("info", f"Git提交: {deployment.commit_hash[:8]}")
            append_log("info", f"目标主机: {project.target_host}")

            # 1. 编译阶段
            append_log("info", "=== 编译阶段 ===")
            compile_success, compile_logs, build_time = self.docker_service.compile_project(
                project_id=project.id,
                git_url=project.git_url,
                commit_hash=deployment.commit_hash,
                language=project.language,
                custom_build_command=project.build_command,
                output_callback=append_log
            )

            deployment.build_time = build_time

            if not compile_success:
                raise Exception("编译失败")

            # 2. 部署阶段
            append_log("info", "=== 部署阶段 ===")

            # 由于Docker编译容器结束后会删除构建产物，
            # 这里我们需要重新获取构建产物
            # 简化处理：假设构建产物在临时目录的repo子目录中
            temp_dir = tempfile.mkdtemp(prefix=f"project_{project.id}_")

            # 重新克隆代码到临时目录
            try:
                from git import Repo
                repo_dir = os.path.join(temp_dir, "repo")
                repo = Repo.clone_from(project.git_url, repo_dir)
                repo.git.checkout(deployment.commit_hash)

                deploy_success, deploy_logs, deploy_time = self.ssh_service.deploy(
                    local_dir=repo_dir,
                    target_host=project.target_host,
                    target_port=project.target_port,
                    target_username=project.target_username,
                    deploy_path=project.deploy_path,
                    restart_command=project.restart_command,
                    ssh_key_path=project.ssh_key_path,
                    ssh_password=project.ssh_password,
                    output_callback=append_log
                )

                deployment.deploy_time = deploy_time

                if not deploy_success:
                    raise Exception("部署失败")

            finally:
                # 清理临时目录
                import shutil
                if os.path.exists(temp_dir):
                    shutil.rmtree(temp_dir, ignore_errors=True)

            # 部署成功
            deployment.status = Deployment.STATUS_SUCCESS
            deployment.end_time = datetime.utcnow()
            deployment.duration = int((deployment.end_time - deployment.start_time).total_seconds())
            deployment.logs = "\n".join(all_logs)

            db.commit()

            append_log("success", f"部署成功完成！总耗时: {deployment.duration} 秒")
            return True

        except Exception as e:
            # 部署失败
            error_msg = str(e)
            all_logs.append(f"[ERROR] {error_msg}")

            deployment.status = Deployment.STATUS_FAILED
            deployment.end_time = datetime.utcnow()
            deployment.duration = int((deployment.end_time - deployment.start_time).total_seconds())
            deployment.logs = "\n".join(all_logs)
            deployment.error_message = error_msg

            db.commit()

            append_log("error", f"部署失败: {error_msg}")
            logger.error(f"部署失败: {error_msg}", exc_info=True)
            return False

    async def get_deployments(
        self,
        db: Session,
        project_id: Optional[int] = None,
        limit: int = 50,
        offset: int = 0
    ):
        """获取部署列表"""
        query = db.query(Deployment)

        if project_id:
            query = query.filter(Deployment.project_id == project_id)

        total = query.count()
        deployments = query.order_by(Deployment.start_time.desc()).offset(offset).limit(limit).all()

        return deployments, total

    async def get_deployment(self, db: Session, deployment_id: int) -> Optional[Deployment]:
        """获取单个部署记录"""
        return db.query(Deployment).filter(Deployment.id == deployment_id).first()

    async def get_project_stats(self, db: Session, project_id: int) -> dict:
        """获取项目统计信息"""
        # 总部署数
        total = db.query(Deployment).filter(Deployment.project_id == project_id).count()

        # 成功部署数
        success = db.query(Deployment).filter(
            Deployment.project_id == project_id,
            Deployment.status == Deployment.STATUS_SUCCESS
        ).count()

        # 失败部署数
        failed = db.query(Deployment).filter(
            Deployment.project_id == project_id,
            Deployment.status == Deployment.STATUS_FAILED
        ).count()

        # 平均耗时
        avg_duration = db.query(Deployment.duration).filter(
            Deployment.project_id == project_id,
            Deployment.status == Deployment.STATUS_SUCCESS
        ).all()

        if avg_duration:
            avg = sum([d[0] for d in avg_duration if d[0]]) / len(avg_duration)
        else:
            avg = 0

        return {
            "total": total,
            "success": success,
            "failed": failed,
            "success_rate": (success / total * 100) if total > 0 else 0,
            "average_duration": int(avg)
        }
