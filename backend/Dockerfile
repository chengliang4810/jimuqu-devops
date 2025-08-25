# 多阶段构建 - 编译阶段
FROM maven:3.8-openjdk-17 AS builder

# 设置工作目录
WORKDIR /app

# 复制pom.xml文件
COPY pom.xml .

# 下载依赖（利用Docker缓存）
RUN mvn dependency:go-offline -B

# 复制源代码
COPY src ./src

# 编译打包应用
RUN mvn clean package -DskipTests=true -B

# 运行阶段
FROM openjdk:17-jdk-slim

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 安装curl和其他必要工具
RUN apt-get update && \
    apt-get install -y curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# 创建应用用户
RUN groupadd -r appuser && useradd -r -g appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制jar文件
COPY --from=builder /app/target/devops.jar app.jar

# 创建日志目录
RUN mkdir -p /app/logs && \
    mkdir -p /tmp/jimuqu-devops/workspace && \
    chown -R appuser:appuser /app && \
    chown -R appuser:appuser /tmp/jimuqu-devops

# 切换到应用用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
CMD ["java", \
     "-Djava.security.egd=file:/dev/./urandom", \
     "-Dserver.port=8080", \
     "-Xms256m", \
     "-Xmx512m", \
     "-jar", \
     "app.jar"]