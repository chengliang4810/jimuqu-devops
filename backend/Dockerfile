# 使用官方Go镜像作为基础镜像
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -o devops-app .

# 使用scratch作为最终镜像
FROM scratch

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/devops-app /devops-app

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["/devops-app"]