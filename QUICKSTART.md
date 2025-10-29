# 快速开始指南

## 🚀 5分钟快速部署

### 步骤1：启动平台

```bash
# 方式1：一键启动（推荐）
./start.sh

# 方式2：手动启动
docker-compose up -d
```

### 步骤2：访问平台

打开浏览器访问：http://localhost:8000

登录凭据：
- 用户名：`admin`
- 密码：`admin123`

### 步骤3：创建第一个项目

1. 点击"项目" → "+ 新建项目"
2. 填写项目信息（示例）：
   - **项目名称**: `my-app`
   - **Git地址**: `https://github.com/username/repo.git`
   - **开发语言**: `Java`
   - **部署路径**: `/opt/my-app`
   - **目标主机**: `192.168.1.100`
   - **SSH端口**: `22`
   - **主机用户名**: `root`
   - **重启命令**: `systemctl restart my-app`

3. 点击"保存"

### 步骤4：配置Git Webhook

#### GitHub配置示例：

1. 进入GitHub仓库 → Settings → Webhooks → Add webhook
2. **Payload URL**: `http://你的服务器IP:8000/api/webhook/github/1`
3. **Content type**: `application/json`
4. **Secret**: 填写项目配置中的Webhook密钥
5. **Events**: "Just the push event"
6. 点击"Add webhook"

#### GitLab配置示例：

1. 进入GitLab项目 → Settings → Webhooks
2. **URL**: `http://你的服务器IP:8000/api/webhook/gitlab/1`
3. **Secret token**: 填写项目配置中的Webhook密钥
4. **Trigger events**: "Push events"
5. 点击"Add webhook"

### 步骤5：测试部署

1. 进入"部署记录" → 点击"执行"
2. 切换到"实时日志"查看部署过程
3. 部署成功后，访问目标主机验证应用

## 📖 完整功能说明

### 项目配置示例

#### Java项目

```
项目名称: spring-boot-app
Git地址: https://github.com/user/spring-boot-app.git
开发语言: Java
构建命令: (留空，使用默认 mvn clean package)
部署路径: /opt/spring-boot
目标主机: 192.168.1.100
SSH端口: 22
主机用户名: root
重启命令: systemctl restart spring-boot-app
```

#### Python项目

```
项目名称: flask-api
Git地址: https://github.com/user/flask-api.git
开发语言: Python
构建命令: pip install -r requirements.txt
部署路径: /opt/flask-api
目标主机: 192.168.1.100
SSH端口: 22
主机用户名: ubuntu
重启命令: pm2 restart flask-api
```

#### Node.js项目

```
项目名称: react-frontend
Git地址: https://github.com/user/react-frontend.git
开发语言: Node
构建命令: npm install && npm run build
部署路径: /var/www/react
目标主机: 192.168.1.100
SSH端口: 22
主机用户名: www-data
重启命令: nginx -s reload
```

#### Go项目

```
项目名称: go-api
Git地址: https://github.com/user/go-api.git
开发语言: Go
构建命令: go mod download && go build -o api main.go
部署路径: /opt/go-api
目标主机: 192.168.1.100
SSH端口: 22
主机用户名: root
重启命令: systemctl restart go-api
```

## 🔧 常用操作

### 查看部署日志

1. 进入"部署记录"页面
2. 点击"查看日志"
3. 实时查看编译和部署过程

### 手动触发部署

1. 进入"部署记录"页面
2. 点击"执行"按钮
3. 查看任务提交结果

### 查看部署统计

进入"仪表盘"页面，查看：
- 总项目数
- 活跃项目数
- 今日部署数
- 成功率
- 平均耗时

### 测试项目连接

1. 进入"项目"页面
2. 点击"测试"按钮
3. 查看Docker和SSH连接测试结果

## 🐛 故障排查

### 编译失败

**原因1**: Git仓库访问失败
```bash
# 检查Git URL是否正确
# 检查网络是否可达
# 检查目标主机是否能访问Git仓库
```

**原因2**: 构建命令错误
```bash
# 检查构建命令是否正确
# 检查依赖是否完整
# 查看详细日志定位问题
```

### 部署失败

**原因1**: SSH连接失败
```bash
# 检查目标主机IP和端口
# 检查用户名和密码/密钥
# 检查防火墙设置
```

**原因2**: 权限不足
```bash
# 确保SSH用户有部署路径的写权限
# 检查sudo权限（如需要）
```

### Webhook不触发

**原因1**: URL配置错误
```bash
# 检查Webhook URL是否正确
# 确保服务器IP可从Git仓库访问
# 检查端口8000是否开放
```

**原因2**: Secret不匹配
```bash
# 检查Webhook密钥是否正确
# 检查GitHub/GitLab Secret配置
```

## 📝 最佳实践

### 1. 使用SSH密钥

强烈建议使用SSH密钥而非密码连接目标主机：

```bash
# 生成SSH密钥对
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# 将公钥复制到目标主机
ssh-copy-id user@target_host

# 在项目配置中使用密钥路径
ssh_key_path: /root/.ssh/id_rsa
```

### 2. 配置Webhook

为每个项目配置Webhook实现自动部署：
- 每次push代码自动触发部署
- 支持指定分支触发
- 支持推送特定标签触发

### 3. 监控和告警

启用Flower监控面板：http://localhost:5555
- 查看任务队列状态
- 查看Worker运行状态
- 查看任务执行历史

### 4. 日志管理

```bash
# 查看Web服务日志
docker-compose logs -f web

# 查看Worker日志
docker-compose logs -f worker

# 查看Redis日志
docker-compose logs -f redis

# 清理日志
docker-compose down && docker-compose up -d
```

## 🎯 下一步

- [ ] 配置HTTPS访问
- [ ] 添加邮件/钉钉通知
- [ ] 配置多环境部署（dev/staging/prod）
- [ ] 集成堡垒机
- [ ] 添加审批流程
- [ ] 配置自动回滚

## 💡 提示

1. **首次使用**：建议先用手动部署测试整个流程
2. **安全加固**：生产环境务必修改默认密码和使用HTTPS
3. **资源监控**：注意服务器磁盘空间，防止日志堆积
4. **备份策略**：定期备份数据库和重要配置文件

有问题？查看 [README.md](README.md) 或提交Issue。
