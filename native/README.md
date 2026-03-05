# Nai-Tizi - 极简 Go Web 脚手架

[![Go Version](https://img.shields.io/badge/Go-1.25.6+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Vue 3](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Nai-Tizi** 是一款专注于**开发体验**的极简 Go Web 脚手架。它拒绝过度封装，保留了 Gin 的原生体验，同时集成了企业级开发必备的核心组件（JWT、Casbin、GORM、Zap）。

**核心理念：** 简单、透明

## ✨ 核心特性

- **极简架构**：标准的 Controller-Service-Model 分层，依赖注入清晰明了。
- **完备认证**：开箱即用的双 Token (Access/Refresh) 认证 + Casbin RBAC 权限控制。
- **企业级基建**：集成 GORM(PostgreSQL)、Redis、Zap 日志、Prometheus 监控。
- **现代化前端**：配套 Vue 3 + TypeScript + Ant Design Vue 管理后台，支持动态路由。
- **云原生就绪**：提供 Dockerfile、Docker Compose 及 K8s 部署清单。

## 🚀 30秒快速开始

1. **克隆项目**
   ```bash
   git clone git@github.com:force-c/nai-tizi.git
   ```

2. **配置运行**
   ```bash
   cd nai-tizi
   go mod download
   cp cmd/api/conf.dev.yaml cmd/api/conf.prod.yaml
   
   # 运行服务
   make run
   ```
   > 默认监听端口: 9009 | Swagger 文档: http://localhost:9009/swagger/index.html

3. **启动前端**
   ```bash
   cd web && pnpm install && pnpm dev
   ```
   > 访问地址: http://localhost:3000 (admin / admin123)

## 📖 项目结构

```
nai-tizi/
├── cmd/api/                # 入口与配置
├── internal/
│   ├── controller/         # 接口层 (参数解析/响应)
│   ├── service/            # 业务层 (核心逻辑/事务)
│   ├── domain/             # 领域层 (Model/DO/DTO)
│   ├── infrastructure/     # 基础层 (DB/Redis/MQ/S3)
│   └── router/             # 路由注册
├── web/                    # 前端源码 (Vue 3)
├── dockerfile/             # 容器化构建
└── k8s/                    # Kubernetes 部署
```

## � 文档支持

详细文档请查阅 `docs/` 目录：
- [开发规范](docs/01-规范/开发规范指南.md)
- [API 文档](docs/04-API文档/Swagger文档使用指南.md)
- [部署说明](docs/05-部署运维/部署说明.md)

## 🤝 参与贡献

欢迎提交 PR 或 Issue。
仓库地址：[github.com/force-c/nai-tizi](https://github.com/force-c/nai-tizi)

---
**License**: MIT
