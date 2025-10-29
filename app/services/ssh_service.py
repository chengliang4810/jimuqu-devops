"""
SSH部署服务 - 负责上传文件到目标主机并执行命令
"""
import os
import paramiko
import stat
import logging
from typing import Tuple, Optional
from pathlib import Path
import tarfile
import io

logger = logging.getLogger(__name__)


class SSHDeployService:
    """SSH部署服务"""

    def __init__(self):
        self.ssh_clients = {}  # 缓存SSH连接

    def deploy(
        self,
        local_dir: str,
        target_host: str,
        target_port: int,
        target_username: str,
        deploy_path: str,
        restart_command: Optional[str] = None,
        ssh_key_path: Optional[str] = None,
        ssh_password: Optional[str] = None,
        output_callback=None
    ) -> Tuple[bool, str, int]:
        """
        部署到目标主机

        Args:
            local_dir: 本地构建产物目录
            target_host: 目标主机
            target_port: SSH端口
            target_username: 目标主机用户名
            deploy_path: 部署路径
            restart_command: 重启命令
            ssh_key_path: SSH私钥路径
            ssh_password: SSH密码
            output_callback: 日志输出回调函数

        Returns:
            Tuple[是否成功, 日志, 部署耗时（秒）]
        """
        import time
        from datetime import datetime

        start_time = datetime.now()
        logs = []

        try:
            output_callback and output_callback("info", f"正在连接到 {target_host}:{target_port}")

            # 建立SSH连接
            ssh = self._get_ssh_connection(
                target_host, target_port, target_username, ssh_key_path, ssh_password
            )
            sftp = ssh.open_sftp()

            output_callback and output_callback("success", "SSH连接建立成功")

            # 1. 创建部署目录（如果不存在）
            output_callback and output_callback("info", f"检查部署目录: {deploy_path}")
            stdin, stdout, stderr = ssh.exec_command(f"mkdir -p {deploy_path}")
            stderr_output = stderr.read().decode()
            if stderr_output:
                raise Exception(f"创建目录失败: {stderr_output}")

            # 2. 备份当前版本（如果存在）
            backup_success = self._backup_current_version(
                ssh, sftp, deploy_path, output_callback
            )

            # 3. 上传新版本
            output_callback and output_callback("info", "开始上传构建产物")
            self._upload_build_artifacts(
                ssh, sftp, local_dir, deploy_path, output_callback
            )
            output_callback and output_callback("success", "文件上传完成")

            # 4. 设置文件权限
            self._set_permissions(ssh, deploy_path, output_callback)

            # 5. 执行重启命令
            if restart_command:
                output_callback and output_callback("info", f"执行重启命令: {restart_command}")
                stdin, stdout, stderr = ssh.exec_command(restart_command)
                stderr_output = stderr.read().decode()
                if stderr_output:
                    output_callback and output_callback("error", f"重启命令警告: {stderr_output}")

                # 等待应用启动
                output_callback and output_callback("info", "等待应用启动...")
                time.sleep(2)

            # 6. 清理旧备份（可选）
            if backup_success:
                output_callback and output_callback("info", "清理旧备份")

            # 计算部署时间
            from datetime import datetime
            deploy_time = int((datetime.now() - start_time).total_seconds())

            output_callback and output_callback("success", f"部署完成，耗时 {deploy_time} 秒")
            return True, "\n".join(logs), deploy_time

        except Exception as e:
            error_msg = f"部署失败: {str(e)}"
            logs.append(error_msg)
            logger.error(error_msg, exc_info=True)
            return False, "\n".join(logs), 0

        finally:
            # 清理连接
            if 'sftp' in locals():
                sftp.close()
            if 'ssh' in locals():
                ssh.close()

    def _get_ssh_connection(
        self,
        host: str,
        port: int,
        username: str,
        key_path: Optional[str] = None,
        password: Optional[str] = None
    ) -> paramiko.SSHClient:
        """建立SSH连接"""
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())

        try:
            if key_path and os.path.exists(key_path):
                # 使用私钥连接
                private_key = paramiko.RSAKey.from_private_key_file(key_path)
                ssh.connect(
                    hostname=host,
                    port=port,
                    username=username,
                    pkey=private_key,
                    timeout=10
                )
            elif password:
                # 使用密码连接
                ssh.connect(
                    hostname=host,
                    port=port,
                    username=username,
                    password=password,
                    timeout=10
                )
            else:
                raise ValueError("必须提供SSH私钥或密码")

            return ssh
        except Exception as e:
            raise Exception(f"SSH连接失败: {str(e)}")

    def _backup_current_version(
        self,
        ssh: paramiko.SSHClient,
        sftp: paramiko.SFTPClient,
        deploy_path: str,
        output_callback=None
    ) -> bool:
        """备份当前版本"""
        try:
            # 检查目录是否存在文件
            stdin, stdout, stderr = ssh.exec_command(f"ls -la {deploy_path}")
            result = stdout.read().decode()

            if not result.strip():
                return False  # 目录为空，无需备份

            # 创建备份
            from datetime import datetime
            backup_name = f"backup_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
            backup_path = f"{deploy_path}/{backup_name}"

            stdin, stdout, stderr = ssh.exec_command(f"cp -r {deploy_path} {backup_path}")
            stderr_output = stderr.read().decode()

            if stderr_output:
                output_callback and output_callback("error", f"备份失败: {stderr_output}")
                return False

            output_callback and output_callback("info", f"备份当前版本到: {backup_path}")
            return True

        except Exception as e:
            output_callback and output_callback("error", f"备份过程出错: {str(e)}")
            return False

    def _upload_build_artifacts(
        self,
        ssh: paramiko.SSHClient,
        sftp: paramiko.SFTPClient,
        local_dir: str,
        deploy_path: str,
        output_callback=None
    ):
        """上传构建产物"""
        # 创建临时压缩包
        temp_tar = io.BytesIO()

        with tarfile.open(fileobj=temp_tar, mode='w:gz') as tar:
            for root, dirs, files in os.walk(local_dir):
                for file in files:
                    local_path = os.path.join(root, file)
                    arcname = os.path.relpath(local_path, local_dir)
                    tar.add(local_path, arcname=arcname)

        temp_tar.seek(0)

        # 上传压缩包
        remote_tar = f"{deploy_path}/build.tar.gz"
        sftp.putfo(temp_tar, remote_tar)

        # 解压
        stdin, stdout, stderr = ssh.exec_command(f"cd {deploy_path} && tar -xzf build.tar.gz && rm build.tar.gz")
        stderr_output = stderr.read().decode()

        if stderr_output:
            raise Exception(f"解压失败: {stderr_output}")

    def _set_permissions(self, ssh: paramiko.SSHClient, deploy_path: str, output_callback=None):
        """设置文件权限"""
        # 设置755权限给目录
        stdin, stdout, stderr = ssh.exec_command(f"find {deploy_path} -type d -exec chmod 755 {{}} \\;")
        stdout.read()  # 等待完成

        # 设置644权限给文件
        stdin, stdout, stderr = ssh.exec_command(f"find {deploy_path} -type f -exec chmod 644 {{}} \\;")
        stdout.read()  # 等待完成

        # 可执行文件设置755
        stdin, stdout, stderr = ssh.exec_command(f"find {deploy_path} -name '*.sh' -exec chmod 755 {{}} \\;")
        stdout.read()  # 等待完成

        output_callback and output_callback("info", "文件权限设置完成")
