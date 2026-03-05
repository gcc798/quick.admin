# Docker 部署指南

## 📋 概述

本文档提供完整的 Docker 容器化部署方案，包括 Dockerfile、docker-compose 配置和最佳实践。

---

## 🐳 Dockerfile

### 多阶段构建 Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git make

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/api /app/api

# 复制配置文件
COPY --from=builder /app/cmd/api/conf.prod.yaml /app/conf.yaml
COPY --from=builder /app/cmd/api/zaplogger.prod.yaml /app/zaplogger.yaml
COPY --from=builder /app/cmd/api/casbin_model.conf /app/casbin_model.conf

# 创建日志目录
RUN mkdir -p /app/logs && chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["/app/api"]
```

### .dockerignore

```
# Git
.git
.gitignore

# IDE
.vscode
.idea
*.swp
*.swo

# 构建产物
bin/
*.exe
*.dll
*.so
*.dylib

# 测试
*.test
*.out
coverage.txt

# 依赖
vendor/

# 日志
logs/
*.log

# 临时文件
tmp/
temp/

# 文档
docs/
README.md

# 前端
web/
node_modules/

# 环境配置
.env
.env.local
*.dev.yaml

# 其他
.DS_Store
Thumbs.db
```

---

## 🚀 Docker Compose

### 开发环境配置

```yaml
version: '3.8'

services:
  # 后端 API 服务
  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: nai-tizi-api
    ports:
      - "8080:8080"
    environment:
      - ENV=development
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=nai_tizi
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./logs:/app/logs
      - ./cmd/api/conf.dev.yaml:/app/conf.yaml
    networks:
      - app-network
    restart: unless-stopped

  # PostgreSQL 数据库
  postgres:
    image: postgres:15-alpine
    container_name: nai-tizi-postgres
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=nai_tizi
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./scripts/sql:/docker-entrypoint-initdb.d
    networks:
      - app-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # Redis 缓存
  redis:
    image: redis:7-alpine
    container_name: nai-tizi-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    restart: unless-stopped

  # MinIO 对象存储
  minio:
    image: minio/minio:latest
    container_name: nai-tizi-minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    volumes:
      - minio-data:/data
    command: server /data --console-address ":9001"
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
    restart: unless-stopped

volumes:
  postgres-data:
  redis-data:
  minio-data:

networks:
  app-network:
    driver: bridge
```

### 生产环境配置

```yaml
version: '3.8'

services:
  api:
    image: nai-tizi-api:latest
    container_name: nai-tizi-api-prod
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - DB_HOST=${DB_HOST}
      - DB_PORT=${DB_PORT}
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - REDIS_HOST=${REDIS_HOST}
      - REDIS_PORT=${REDIS_PORT}
    volumes:
      - ./logs:/app/logs
      - ./conf.prod.yaml:/app/conf.yaml
    networks:
      - app-network
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '1'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

networks:
  app-network:
    driver: bridge
```

---

## 📝 使用指南

### 构建镜像

```bash
# 构建镜像
docker build -t nai-tizi-api:latest .

# 构建并指定平台
docker build --platform linux/amd64 -t nai-tizi-api:latest .

# 查看镜像
docker images | grep nai-tizi
```

### 运行容器

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f api

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

### 进入容器

```bash
# 进入 API 容器
docker exec -it nai-tizi-api sh

# 进入 PostgreSQL 容器
docker exec -it nai-tizi-postgres psql -U postgres -d nai_tizi

# 进入 Redis 容器
docker exec -it nai-tizi-redis redis-cli
```

---

## 🔧 环境变量配置

### .env 文件示例

```bash
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=nai_tizi

# Redis 配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# MinIO 配置
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=your_minio_password

# JWT 配置
JWT_SECRET=your_jwt_secret_key

# 应用配置
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info
```

---

## 🎯 最佳实践

### 1. 镜像优化

- ✅ 使用多阶段构建减小镜像大小
- ✅ 使用 Alpine 基础镜像
- ✅ 合并 RUN 命令减少层数
- ✅ 使用 .dockerignore 排除不必要文件
- ✅ 使用非 root 用户运行应用

### 2. 安全配置

- ✅ 不在镜像中硬编码敏感信息
- ✅ 使用环境变量或 secrets 管理配置
- ✅ 定期更新基础镜像
- ✅ 扫描镜像漏洞
- ✅ 限制容器资源使用

### 3. 健康检查

```dockerfile
# HTTP 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# TCP 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD nc -z localhost 8080 || exit 1
```

### 4. 日志管理

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
    labels: "service,environment"
```

---

## 🚀 部署流程

### 1. 准备工作

```bash
# 克隆代码
git clone <repository-url>
cd nai-tizi

# 配置环境变量
cp .env.example .env
vim .env

# 准备配置文件
cp cmd/api/conf.dev.yaml cmd/api/conf.prod.yaml
vim cmd/api/conf.prod.yaml
```

### 2. 构建和启动

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 检查服务状态
docker-compose ps
docker-compose logs -f
```

### 3. 初始化数据库

```bash
# 进入数据库容器
docker exec -it nai-tizi-postgres psql -U postgres -d nai_tizi

# 或者从外部执行 SQL
docker exec -i nai-tizi-postgres psql -U postgres -d nai_tizi < scripts/sql/pgsql.sql
docker exec -i nai-tizi-postgres psql -U postgres -d nai_tizi < scripts/sql/insert.sql
```

### 4. 验证部署

```bash
# 检查健康状态
curl http://localhost:8080/health

# 检查 API 文档
curl http://localhost:8080/swagger/index.html

# 测试登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

## 🔍 故障排查

### 常见问题

**1. 容器无法启动**
```bash
# 查看容器日志
docker-compose logs api

# 检查容器状态
docker-compose ps

# 重启容器
docker-compose restart api
```

**2. 数据库连接失败**
```bash
# 检查数据库是否就绪
docker-compose logs postgres

# 测试数据库连接
docker exec -it nai-tizi-postgres psql -U postgres -d nai_tizi -c "SELECT 1"
```

**3. 端口冲突**
```bash
# 查看端口占用
lsof -i :8080

# 修改 docker-compose.yml 中的端口映射
ports:
  - "8081:8080"  # 改为其他端口
```

---

## 📊 监控集成

### Prometheus 配置

```yaml
# 在 docker-compose.yml 中添加
prometheus:
  image: prom/prometheus:latest
  container_name: prometheus
  ports:
    - "9090:9090"
  volumes:
    - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus-data:/prometheus
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
  networks:
    - app-network
```

### Grafana 配置

```yaml
grafana:
  image: grafana/grafana:latest
  container_name: grafana
  ports:
    - "3000:3000"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=admin
  volumes:
    - grafana-data:/var/lib/grafana
  networks:
    - app-network
```

---

## 📚 相关文档

- [Kubernetes 部署指南](./Kubernetes部署指南.md)
- [监控系统配置](./监控系统配置.md)
- [CI/CD 配置指南](./CICD配置指南.md)

---

**文档版本：** v1.0  
**最后更新：** 2026-01-12  
**维护者：** 项目团队
