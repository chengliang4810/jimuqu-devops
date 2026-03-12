FROM node:22-alpine AS frontend-builder
WORKDIR /app/web-next

COPY web-next/package.json web-next/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

COPY web-next/ ./
RUN NEXT_PUBLIC_APP_VERSION=docker pnpm build

FROM golang:1.25-alpine AS backend-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
COPY --from=frontend-builder /app/web-next/out ./web-next/out
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.22
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata git openssh-client docker-cli

COPY --from=backend-builder /out/server ./server
COPY --from=frontend-builder /app/web-next/out ./web-next/out
COPY README.md ./README.md

ENV APP_ADDR=:18080
ENV APP_DATA_DIR=/app/data
ENV APP_DB_DRIVER=sqlite
ENV APP_DB_SOURCE=/app/data/pipeline.db
ENV APP_WORKSPACE_DIR=/app/data/workspaces
ENV APP_SECRET=change-me-in-production

VOLUME ["/app/data"]
EXPOSE 18080

ENTRYPOINT ["./server"]
