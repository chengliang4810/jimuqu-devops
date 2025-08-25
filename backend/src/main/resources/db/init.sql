-- jimuqu-devops 数据库初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS jimuqu_devops DEFAULT CHARACTER SET utf8mb4;
USE jimuqu_devops;

-- 主机管理表
CREATE TABLE IF NOT EXISTS hosts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主机ID',
    name VARCHAR(100) NOT NULL COMMENT '主机名称',
    host_ip VARCHAR(50) NOT NULL COMMENT '主机IP地址',
    port INT DEFAULT 22 COMMENT 'SSH端口',
    username VARCHAR(50) NOT NULL COMMENT 'SSH用户名',
    password VARCHAR(255) COMMENT 'SSH密码(加密存储)',
    private_key TEXT COMMENT 'SSH私钥',
    status VARCHAR(20) DEFAULT 'OFFLINE' COMMENT '主机状态：ONLINE/OFFLINE/ERROR',
    description TEXT COMMENT '主机描述',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_status (status),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主机管理表';

-- 模板管理表
CREATE TABLE IF NOT EXISTS templates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '模板ID',
    name VARCHAR(100) NOT NULL COMMENT '模板名称',
    type VARCHAR(50) NOT NULL COMMENT '模板类型：SPRINGBOOT/VUE/REACT/STATIC等',
    description TEXT COMMENT '模板描述',
    is_system BOOLEAN DEFAULT FALSE COMMENT '是否为系统预置模板',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_name (name),
    INDEX idx_type (type),
    INDEX idx_is_system (is_system)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模板管理表';

-- 模板步骤表
CREATE TABLE IF NOT EXISTS template_steps (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '步骤ID',
    template_id BIGINT NOT NULL COMMENT '关联模板ID',
    name VARCHAR(100) NOT NULL COMMENT '步骤名称',
    image VARCHAR(200) NOT NULL COMMENT 'Docker镜像名',
    commands TEXT NOT NULL COMMENT '执行命令列表(JSON格式)',
    step_order INT NOT NULL COMMENT '执行顺序',
    continue_on_error BOOLEAN DEFAULT FALSE COMMENT '出错是否继续',
    work_dir VARCHAR(500) DEFAULT '/workspace' COMMENT '工作目录',
    environment TEXT COMMENT '环境变量(JSON格式)',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE,
    INDEX idx_template_id (template_id),
    INDEX idx_step_order (step_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模板步骤表';

-- 应用管理表
CREATE TABLE IF NOT EXISTS applications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '应用ID',
    name VARCHAR(100) NOT NULL COMMENT '应用名称',
    git_repo_url VARCHAR(500) NOT NULL COMMENT 'Git仓库地址',
    git_branch VARCHAR(100) DEFAULT 'main' COMMENT 'Git分支',
    git_username VARCHAR(100) COMMENT 'Git用户名',
    git_password VARCHAR(255) COMMENT 'Git密码(加密存储)',
    git_private_key TEXT COMMENT 'Git私钥',
    template_id BIGINT COMMENT '关联模板ID',
    host_ids TEXT COMMENT '目标主机ID列表(JSON格式)',
    variables TEXT COMMENT '环境变量(JSON格式)',
    notification_url VARCHAR(500) COMMENT '通知地址',
    notification_token VARCHAR(255) COMMENT '通知令牌',
    auto_trigger BOOLEAN DEFAULT FALSE COMMENT '是否自动触发',
    webhook_secret VARCHAR(255) COMMENT 'Webhook密钥',
    status VARCHAR(20) DEFAULT 'ACTIVE' COMMENT '应用状态：ACTIVE/INACTIVE',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_name (name),
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE SET NULL,
    INDEX idx_status (status),
    INDEX idx_auto_trigger (auto_trigger)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用管理表';

-- 流水线配置表
CREATE TABLE IF NOT EXISTS pipeline_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '配置ID',
    application_id BIGINT NOT NULL COMMENT '关联应用ID',
    steps TEXT NOT NULL COMMENT '流水线步骤配置(JSON格式)',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    INDEX idx_application_id (application_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水线配置表';

-- 构建记录表
CREATE TABLE IF NOT EXISTS builds (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '构建ID',
    application_id BIGINT NOT NULL COMMENT '关联应用ID',
    build_number INT NOT NULL COMMENT '构建编号',
    status VARCHAR(20) NOT NULL COMMENT '构建状态：PENDING/RUNNING/SUCCESS/FAILED/CANCELLED',
    triggered_by VARCHAR(50) COMMENT '触发方式：MANUAL/WEBHOOK',
    trigger_user VARCHAR(100) COMMENT '触发用户',
    git_commit VARCHAR(100) COMMENT 'Git提交哈希',
    git_branch VARCHAR(100) COMMENT 'Git分支',
    start_time DATETIME COMMENT '开始时间',
    end_time DATETIME COMMENT '结束时间',
    duration_seconds INT COMMENT '构建时长(秒)',
    log_file_path VARCHAR(500) COMMENT '日志文件路径',
    error_message TEXT COMMENT '错误信息',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    INDEX idx_application_id (application_id),
    INDEX idx_status (status),
    INDEX idx_build_number (application_id, build_number),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='构建记录表';

-- 构建步骤记录表
CREATE TABLE IF NOT EXISTS build_steps (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '步骤记录ID',
    build_id BIGINT NOT NULL COMMENT '关联构建ID',
    step_name VARCHAR(100) NOT NULL COMMENT '步骤名称',
    step_order INT NOT NULL COMMENT '步骤顺序',
    status VARCHAR(20) NOT NULL COMMENT '步骤状态：PENDING/RUNNING/SUCCESS/FAILED/SKIPPED',
    start_time DATETIME COMMENT '开始时间',
    end_time DATETIME COMMENT '结束时间',
    duration_seconds INT COMMENT '执行时长(秒)',
    log_content TEXT COMMENT '日志内容',
    error_message TEXT COMMENT '错误信息',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE,
    INDEX idx_build_id (build_id),
    INDEX idx_step_order (step_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='构建步骤记录表';

-- 统计数据表
CREATE TABLE IF NOT EXISTS build_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '统计ID',
    application_id BIGINT COMMENT '应用ID(NULL表示全局统计)',
    metric_date DATE NOT NULL COMMENT '统计日期',
    build_count INT DEFAULT 0 COMMENT '构建次数',
    success_count INT DEFAULT 0 COMMENT '成功次数',
    failed_count INT DEFAULT 0 COMMENT '失败次数',
    avg_duration_seconds INT DEFAULT 0 COMMENT '平均构建时长(秒)',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    UNIQUE KEY uk_app_date (application_id, metric_date),
    INDEX idx_metric_date (metric_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='构建统计表';

-- 通知记录表
CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '通知ID',
    build_id BIGINT NOT NULL COMMENT '关联构建ID',
    notification_type VARCHAR(50) NOT NULL COMMENT '通知类型：HTTP/EMAIL/WECHAT等',
    target VARCHAR(500) NOT NULL COMMENT '通知目标地址',
    payload TEXT NOT NULL COMMENT '通知内容',
    status VARCHAR(20) NOT NULL COMMENT '通知状态：PENDING/SUCCESS/FAILED',
    retry_count INT DEFAULT 0 COMMENT '重试次数',
    error_message TEXT COMMENT '错误信息',
    send_time DATETIME COMMENT '发送时间',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE,
    INDEX idx_build_id (build_id),
    INDEX idx_status (status),
    INDEX idx_send_time (send_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知记录表';

-- 插入系统预置模板数据
INSERT INTO templates (name, type, description, is_system) VALUES
('Spring Boot 项目', 'SPRINGBOOT', 'Maven + Spring Boot 项目的标准构建流程', TRUE),
('Vue 前端项目', 'VUE', 'Node.js + Vue 前端项目的标准构建流程', TRUE),
('React 前端项目', 'REACT', 'Node.js + React 前端项目的标准构建流程', TRUE),
('静态网站', 'STATIC', '静态网站的标准部署流程', TRUE);

-- 插入Spring Boot模板步骤
INSERT INTO template_steps (template_id, name, image, commands, step_order, continue_on_error, work_dir, environment) VALUES
(1, 'Git代码拉取', 'alpine/git:latest', '["git clone ${GIT_REPO_URL} .", "git checkout ${GIT_BRANCH}"]', 1, FALSE, '/workspace', '{}'),
(1, 'Maven编译打包', 'maven:3.8-openjdk-11', '["mvn clean package -DskipTests=true"]', 2, FALSE, '/workspace', '{}'),
(1, 'Docker镜像构建', 'docker:20.10', '["docker build -t ${APP_NAME}:${BUILD_NUMBER} ."]', 3, FALSE, '/workspace', '{}'),
(1, '启动应用容器', 'docker:20.10', '["docker stop ${APP_NAME} || true", "docker rm ${APP_NAME} || true", "docker run -d --name ${APP_NAME} -p ${APP_PORT}:8080 ${APP_NAME}:${BUILD_NUMBER}"]', 4, FALSE, '/workspace', '{}');

-- 插入Vue模板步骤
INSERT INTO template_steps (template_id, name, image, commands, step_order, continue_on_error, work_dir, environment) VALUES
(2, 'Git代码拉取', 'alpine/git:latest', '["git clone ${GIT_REPO_URL} .", "git checkout ${GIT_BRANCH}"]', 1, FALSE, '/workspace', '{}'),
(2, 'Node环境构建', 'node:16-alpine', '["npm install", "npm run build"]', 2, FALSE, '/workspace', '{}'),
(2, 'Nginx镜像打包', 'docker:20.10', '["docker build -t ${APP_NAME}:${BUILD_NUMBER} -f Dockerfile.nginx ."]', 3, FALSE, '/workspace', '{}'),
(2, '启动Nginx容器', 'docker:20.10', '["docker stop ${APP_NAME} || true", "docker rm ${APP_NAME} || true", "docker run -d --name ${APP_NAME} -p ${APP_PORT}:80 ${APP_NAME}:${BUILD_NUMBER}"]', 4, FALSE, '/workspace', '{}');