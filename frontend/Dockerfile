# 多阶段构建 - 构建阶段
FROM node:20-alpine AS builder

# 设置pnpm
RUN npm install -g pnpm

# 设置工作目录
WORKDIR /app

# 复制package文件
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# 复制packages目录
COPY packages ./packages

# 安装依赖
RUN pnpm install

# 复制源代码
COPY . .

# 构建生产版本
RUN pnpm build

# 生产阶段 - Nginx
FROM nginx:alpine

# 设置时区
ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone

# 复制构建产物到nginx目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制nginx配置文件
COPY docker/nginx/default.conf /etc/nginx/conf.d/default.conf

# 创建nginx用户目录
RUN mkdir -p /var/cache/nginx/client_temp && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /usr/share/nginx/html

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]