"""
Docker编译服务 - 负责代码编译和镜像构建
"""
import os
import shutil
import docker
import tempfile
from datetime import datetime
from pathlib import Path
from typing import Optional, Tuple
import logging

logger = logging.getLogger(__name__)


class DockerCompileService:
    """Docker编译服务"""

    # 语言和编译镜像映射
    LANGUAGE_IMAGES = {
        "java": "openjdk:11-jdk-slim",
        "python": "python:3.9-slim",
        "node": "node:16-alpine",
        "go": "golang:1.19-alpine",
    }

    # 默认构建命令
    BUILD_COMMANDS = {
        "java": "mvn clean package -DskipTests",
        "python": "python -m pip install -r requirements.txt",
        "node": "npm install && npm run build",
        "go": "go mod download && go build -o app main.go",
    }

    def __init__(self):
        self.client = docker.from_env()

    def compile_project(
        self,
        project_id: int,
        git_url: str,
        commit_hash: str,
        language: str,
        custom_build_command: Optional[str] = None,
        output_callback=None
    ) -> Tuple[bool, str, int]:
        """
        编译项目

        Args:
            project_id: 项目ID
            git_url: Git仓库地址
            commit_hash: 提交哈希
            language: 开发语言
            custom_build_command: 自定义构建命令
            output_callback: 日志输出回调函数

        Returns:
            Tuple[是否成功, 日志, 构建耗时（秒）]
        """
        start_time = datetime.now()
        logs = []
        temp_dir = None

        try:
            # 1. 创建临时工作目录
            temp_dir = tempfile.mkdtemp(prefix=f"project_{project_id}_")
            output_callback and output_callback("info", f"创建临时目录: {temp_dir}")

            # 2. 克隆代码
            output_callback and output_callback("info", f"正在克隆代码: {git_url}")
            repo_dir = self._clone_git_repo(git_url, commit_hash, temp_dir, output_callback)
            output_callback and output_callback("success", "代码克隆完成")

            # 3. 创建编译容器
            image = self.LANGUAGE_IMAGES.get(language.lower())
            if not image:
                raise ValueError(f"不支持的语言: {language}")

            output_callback and output_callback("info", f"使用编译镜像: {image}")

            # 4. 执行编译命令
            build_command = custom_build_command or self.BUILD_COMMANDS.get(language.lower())
            if not build_command:
                raise ValueError(f"未找到语言 {language} 的默认构建命令")

            output_callback and output_callback("info", f"执行构建命令: {build_command}")
            success, build_logs = self._run_build_container(
                image, repo_dir, build_command, output_callback
            )

            if not success:
                raise Exception(f"编译失败:\n{build_logs}")

            output_callback and output_callback("success", "项目编译成功")

            # 5. 计算构建时间
            build_time = int((datetime.now() - start_time).total_seconds())

            return True, "\n".join(logs), build_time

        except Exception as e:
            error_msg = f"编译失败: {str(e)}"
            logs.append(error_msg)
            logger.error(error_msg, exc_info=True)
            return False, "\n".join(logs), 0

        finally:
            # 清理临时目录
            if temp_dir and os.path.exists(temp_dir):
                shutil.rmtree(temp_dir, ignore_errors=True)
                output_callback and output_callback("info", "清理临时目录完成")

    def _clone_git_repo(self, git_url: str, commit_hash: str, work_dir: str, output_callback=None) -> str:
        """克隆Git仓库"""
        import git

        repo_dir = os.path.join(work_dir, "repo")
        os.makedirs(repo_dir, exist_ok=True)

        try:
            # 克隆仓库
            repo = git.Repo.clone_from(git_url, repo_dir)
            output_callback and output_callback("info", f"切换到提交: {commit_hash[:8]}")

            # 切换到指定提交
            repo.git.checkout(commit_hash)

            return repo_dir
        except Exception as e:
            raise Exception(f"Git操作失败: {str(e)}")

    def _run_build_container(
        self,
        image: str,
        repo_dir: str,
        build_command: str,
        output_callback=None
    ) -> Tuple[bool, str]:
        """运行编译容器"""
        logs = []

        def container_logs(line):
            msg = line.decode('utf-8').strip()
            logs.append(msg)
            output_callback and output_callback("info", msg)

        try:
            # 运行容器执行构建命令
            # 挂载repo目录到容器
            volumes = {
                repo_dir: {"bind": "/workspace", "mode": "rw"}
            }

            output_callback and output_callback("info", "启动编译容器...")
            container = self.client.containers.run(
                image,
                f"bash -c '{build_command}'",
                volumes=volumes,
                working_dir="/workspace",
                detach=True,
                mem_limit="1g",  # 限制内存
                cpu_period=100000,
                cpu_quota=50000,  # 限制50% CPU
            )

            # 实时获取日志
            output_callback and output_callback("info", "正在编译...")
            for line in container.logs(stream=True, stdout=True, stderr=True):
                container_logs(line)

            # 等待容器完成
            result = container.wait()
            exit_code = result.get("StatusCode", 0)

            # 清理容器
            container.remove(force=True)

            if exit_code != 0:
                return False, "\n".join(logs)

            return True, "\n".join(logs)

        except Exception as e:
            error_msg = f"容器执行失败: {str(e)}"
            logs.append(error_msg)
            return False, "\n".join(logs)
