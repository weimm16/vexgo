# 第一阶段：构建前端
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package.json frontend/package-lock.json frontend/pnpm-lock.yaml ./

# 安装依赖（使用pnpm）
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# 复制前端源码
COPY frontend/ ./

# 创建 backend/public 目录（vite构建需要）
RUN mkdir -p backend/public

# 构建前端（输出到 backend/public/dist）
RUN pnpm run build

# 第二阶段：构建后端
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

# 安装必要的构建工具（包括gcc用于CGO）
RUN apk add --no-cache git ca-certificates gcc musl-dev

# 复制go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源码
COPY backend/ ./backend/

# 复制前端构建结果（从frontend-builder阶段的 backend/public/dist）
COPY --from=frontend-builder /app/backend/public/dist/ ./backend/public/dist/

# 构建后端
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o vexgo ./backend/main.go

# 第三阶段：运行镜像
FROM alpine:3.19

WORKDIR /app

# 安装必要的运行时依赖（用于SQLite）
RUN apk add --no-cache ca-certificates tzdata sqlite

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/vexgo ./

# 创建数据目录
RUN mkdir -p /app/data/media

# 暴露端口
EXPOSE 3001

# 设置环境变量
ENV ADDR=0.0.0.0
ENV PORT=3001
ENV DATA_DIR=/app/data

# 运行应用
CMD ["./vexgo"]