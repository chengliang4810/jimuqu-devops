# Jimuqu DevOps 平台运行指南

## 项目概述

基于Solon框架开发的Docker容器化DevOps平台，核心功能包括：
- 主机管理：SSH主机的增删改查和连接测试
- 模板管理：预置和自定义部署模板
- 应用管理：Git仓库集成和构建配置
- 流水线执行：Docker容器化构建和部署
- Webhook支持：自动触发构建
- 通知服务：构建结果通知

## 环境要求

### 基础环境
- JDK 17+
- Maven 3.6+
- MySQL 8.0+
- Docker Engine (用于容器化构建)

### 可选环境
- Redis (缓存，当前使用内存缓存)
- Nginx (反向代理)

## 快速启动

### 1. 数据库初始化

```sql
-- 执行数据库初始化脚本
source backend/src/main/resources/db/init.sql;
```

### 2. 配置文件设置

编辑 `backend/src/main/resources/app.yml`：

```yaml
solon.dataSources.main:
  url: jdbc:mysql://localhost:3306/jimuqu_devops?useUnicode=true&characterEncoding=utf8&useSSL=false&serverTimezone=Asia/Shanghai
  username: your_username
  password: your_password

docker:
  host: unix:///var/run/docker.sock  # Linux/Mac
  # host: tcp://localhost:2375      # Windows Docker Desktop
```

### 3. 编译运行

```bash
cd backend
mvn clean package
java -jar target/devops.jar
```

应用启动后访问：http://localhost:8080

## API文档

### 主机管理 API

| 接口 | 方法 | 描述 |
|------|------|------|
| GET /api/hosts | GET | 分页查询主机列表 |
| POST /api/hosts | POST | 创建主机 |
| PUT /api/hosts/{id} | PUT | 更新主机 |
| DELETE /api/hosts/{id} | DELETE | 删除主机 |
| POST /api/hosts/{id}/test | POST | 测试主机连接 |
| GET /api/hosts/online | GET | 获取在线主机 |

### 应用管理 API

| 接口 | 方法 | 描述 |
|------|------|------|
| POST /api/applications/{id}/build | POST | 手动触发构建 |
| POST /api/applications/{id}/webhook | POST | Webhook触发构建 |
| GET /api/applications/{id}/status | GET | 查询构建状态 |

## 使用示例

### 1. 创建主机

```bash
curl -X POST http://localhost:8080/api/hosts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "生产服务器",
    "hostIp": "192.168.1.100",
    "port": 22,
    "username": "root",
    "password": "your_password",
    "description": "生产环境服务器"
  }'
```

### 2. 测试主机连接

```bash
curl -X POST http://localhost:8080/api/hosts/1/test
```

### 3. 触发应用构建

```bash
curl -X POST http://localhost:8080/api/applications/1/build
```

## 流水线配置示例

### Spring Boot 项目流水线

```yaml
steps:
  - name: "Git代码拉取"
    image: "alpine/git:latest"
    commands:
      - "git clone ${GIT_REPO_URL} ."
      - "git checkout ${GIT_BRANCH}"
    
  - name: "Maven编译打包"
    image: "maven:3.8-openjdk-11"
    commands:
      - "mvn clean package -DskipTests=true"
    
  - name: "Docker镜像构建"
    image: "docker:20.10"
    commands:
      - "docker build -t ${APP_NAME}:${BUILD_NUMBER} ."
    
  - name: "启动应用容器"
    image: "docker:20.10"
    commands:
      - "docker stop ${APP_NAME} || true"
      - "docker rm ${APP_NAME} || true"
      - "docker run -d --name ${APP_NAME} -p ${APP_PORT}:8080 ${APP_NAME}:${BUILD_NUMBER}"
```

### Vue 前端项目流水线

```yaml
steps:
  - name: "Git代码拉取"
    image: "alpine/git:latest"
    commands:
      - "git clone ${GIT_REPO_URL} ."
      - "git checkout ${GIT_BRANCH}"
    
  - name: "Node环境构建"
    image: "node:16-alpine"
    commands:
      - "npm install"
      - "npm run build"
    
  - name: "Nginx镜像打包"
    image: "docker:20.10"
    commands:
      - "docker build -t ${APP_NAME}:${BUILD_NUMBER} -f Dockerfile.nginx ."
    
  - name: "启动Nginx容器"
    image: "docker:20.10"
    commands:
      - "docker stop ${APP_NAME} || true"
      - "docker rm ${APP_NAME} || true"
      - "docker run -d --name ${APP_NAME} -p ${APP_PORT}:80 ${APP_NAME}:${BUILD_NUMBER}"
```

## Webhook配置

### Gitee配置示例

1. 在Gitee项目设置中添加Webhook
2. URL: `http://your-domain.com/api/applications/{id}/webhook`
3. 选择推送事件
4. 设置密钥（可选）

## 故障排除

### 常见问题

1. **Docker连接失败**
   - 检查Docker守护进程是否运行
   - 确认Docker API地址配置正确
   - Linux系统检查用户权限

2. **SSH连接超时**
   - 检查主机IP和端口是否正确
   - 确认防火墙设置
   - 验证SSH服务状态

3. **构建失败**
   - 查看构建日志
   - 检查Docker镜像是否可用
   - 确认工作空间权限

### 日志查看

应用日志位置：`logs/app.log`

构建日志位置：`/tmp/jimuqu-devops/workspace/{app-name}/{build-number}/`

## 扩展开发

### 添加新的构建模板

1. 在数据库中添加模板记录
2. 配置模板步骤
3. 测试模板功能

### 集成新的通知方式

1. 实现通知接口
2. 添加配置选项
3. 注册通知服务

## 安全建议

1. 使用HTTPS协议
2. 配置防火墙规则
3. 定期更新依赖
4. 使用私钥认证替代密码
5. 设置Webhook签名验证

## 性能优化

1. 配置数据库连接池
2. 使用Redis缓存
3. 优化Docker镜像大小
4. 并行执行构建步骤