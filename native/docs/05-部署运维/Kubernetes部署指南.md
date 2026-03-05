# Kubernetes 部署指南

## 📋 概述

本文档提供完整的 Kubernetes (K8s) 部署方案，包括 Deployment、Service、ConfigMap、Secret、Ingress 和 HPA 配置。

---

## 🎯 Kubernetes 资源

### 已创建的配置文件

1. **k8s/deployment.yaml** - Deployment 配置
2. **k8s/service.yaml** - Service 配置
3. **k8s/configmap.yaml** - ConfigMap 配置
4. **k8s/secret.yaml.example** - Secret 配置模板
5. **k8s/ingress.yaml** - Ingress 配置
6. **k8s/hpa.yaml** - HPA 自动扩展配置

---

## 🚀 快速部署

### 前置要求

- Kubernetes 集群（v1.20+）
- kubectl 命令行工具
- Ingress Controller（如 nginx-ingress）
- cert-manager（用于 TLS 证书）

### 部署步骤

```bash
# 1. 创建命名空间
kubectl create namespace nai-tizi

# 2. 创建 Secret（从模板复制并修改）
cp k8s/secret.yaml.example k8s/secret.yaml
# 编辑 secret.yaml，填入真实密码
kubectl apply -f k8s/secret.yaml -n nai-tizi

# 3. 创建 ConfigMap
kubectl apply -f k8s/configmap.yaml -n nai-tizi

# 4. 部署应用
kubectl apply -f k8s/deployment.yaml -n nai-tizi

# 5. 创建 Service
kubectl apply -f k8s/service.yaml -n nai-tizi

# 6. 创建 Ingress
kubectl apply -f k8s/ingress.yaml -n nai-tizi

# 7. 创建 HPA
kubectl apply -f k8s/hpa.yaml -n nai-tizi
```

### 验证部署

```bash
# 查看 Pod 状态
kubectl get pods -n nai-tizi

# 查看 Service
kubectl get svc -n nai-tizi

# 查看 Ingress
kubectl get ingress -n nai-tizi

# 查看 HPA
kubectl get hpa -n nai-tizi

# 查看 Pod 日志
kubectl logs -f <pod-name> -n nai-tizi
```

---

## 📝 配置详解

### Deployment 配置

**文件：** `k8s/deployment.yaml`

**关键配置：**
- **副本数：** 2 个实例
- **资源限制：**
  - 请求：512Mi 内存，500m CPU
  - 限制：1Gi 内存，1000m CPU
- **健康检查：**
  - 存活探针：30秒后开始，每10秒检查一次
  - 就绪探针：10秒后开始，每5秒检查一次
- **环境变量：** 从 ConfigMap 和 Secret 注入

### Service 配置

**文件：** `k8s/service.yaml`

**服务类型：**
1. **nai-tizi-api** - 主服务（端口 80）
2. **nai-tizi-api-metrics** - Prometheus 指标服务（端口 8080）

### ConfigMap 配置

**文件：** `k8s/configmap.yaml`

**配置项：**
- 数据库连接信息
- Redis 连接信息
- 应用环境配置

### Secret 配置

**文件：** `k8s/secret.yaml.example`

**敏感信息：**
- 数据库用户名和密码
- JWT 密钥
- Redis 密码

**创建 Secret：**
```bash
# 从模板创建
cp k8s/secret.yaml.example k8s/secret.yaml

# 编辑并填入真实值
vim k8s/secret.yaml

# 应用配置
kubectl apply -f k8s/secret.yaml -n nai-tizi

# 删除本地文件（安全考虑）
rm k8s/secret.yaml
```

### Ingress 配置

**文件：** `k8s/ingress.yaml`

**功能：**
- HTTPS 访问（通过 cert-manager）
- 域名路由
- TLS 证书自动管理

**修改域名：**
```yaml
spec:
  tls:
  - hosts:
    - your-domain.com  # 修改为你的域名
  rules:
  - host: your-domain.com  # 修改为你的域名
```

### HPA 配置

**文件：** `k8s/hpa.yaml`

**自动扩展策略：**
- **最小副本数：** 2
- **最大副本数：** 10
- **扩展指标：**
  - CPU 使用率 > 70%
  - 内存使用率 > 80%
- **扩展行为：**
  - 缩容：5分钟稳定期，每次最多缩容50%
  - 扩容：立即扩容，每次最多扩容100%或2个Pod

---

## 🔧 常用操作

### 查看资源状态

```bash
# 查看所有资源
kubectl get all -n nai-tizi

# 查看 Pod 详情
kubectl describe pod <pod-name> -n nai-tizi

# 查看 Pod 日志
kubectl logs -f <pod-name> -n nai-tizi

# 查看最近的事件
kubectl get events -n nai-tizi --sort-by='.lastTimestamp'
```

### 更新配置

```bash
# 更新 ConfigMap
kubectl apply -f k8s/configmap.yaml -n nai-tizi

# 更新 Secret
kubectl apply -f k8s/secret.yaml -n nai-tizi

# 重启 Pod 以应用新配置
kubectl rollout restart deployment/nai-tizi-api -n nai-tizi
```

### 滚动更新

```bash
# 更新镜像
kubectl set image deployment/nai-tizi-api api=ghcr.io/your-org/nai-tizi:v1.1.0 -n nai-tizi

# 查看滚动更新状态
kubectl rollout status deployment/nai-tizi-api -n nai-tizi

# 查看滚动更新历史
kubectl rollout history deployment/nai-tizi-api -n nai-tizi

# 回滚到上一个版本
kubectl rollout undo deployment/nai-tizi-api -n nai-tizi

# 回滚到指定版本
kubectl rollout undo deployment/nai-tizi-api --to-revision=2 -n nai-tizi
```

### 扩缩容

```bash
# 手动扩容
kubectl scale deployment/nai-tizi-api --replicas=5 -n nai-tizi

# 查看 HPA 状态
kubectl get hpa -n nai-tizi

# 查看 HPA 详情
kubectl describe hpa nai-tizi-api-hpa -n nai-tizi
```

### 调试

```bash
# 进入 Pod
kubectl exec -it <pod-name> -n nai-tizi -- sh

# 端口转发
kubectl port-forward <pod-name> 8080:8080 -n nai-tizi

# 查看资源使用
kubectl top pods -n nai-tizi
kubectl top nodes
```

---

## 📊 监控集成

### Prometheus 集成

在 Prometheus 配置中添加 Kubernetes 服务发现：

```yaml
scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
    - role: pod
      namespaces:
        names:
        - nai-tizi
    relabel_configs:
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
      action: keep
      regex: true
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
      action: replace
      target_label: __metrics_path__
      regex: (.+)
    - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
      action: replace
      regex: ([^:]+)(?::\d+)?;(\d+)
      replacement: $1:$2
      target_label: __address__
```

在 Deployment 中添加注解：

```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
```

---

## 🔒 安全最佳实践

### 1. 使用 RBAC

创建 ServiceAccount 和 Role：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nai-tizi-api
  namespace: nai-tizi
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: nai-tizi-api
  namespace: nai-tizi
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: nai-tizi-api
  namespace: nai-tizi
subjects:
- kind: ServiceAccount
  name: nai-tizi-api
roleRef:
  kind: Role
  name: nai-tizi-api
  apiGroup: rbac.authorization.k8s.io
```

### 2. 网络策略

限制 Pod 间通信：

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: nai-tizi-api-netpol
  namespace: nai-tizi
spec:
  podSelector:
    matchLabels:
      app: nai-tizi-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 5432  # PostgreSQL
    - protocol: TCP
      port: 6379  # Redis
```

### 3. Pod Security Standards

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: nai-tizi
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

---

## 🎯 生产环境清单

### 部署前检查

- [ ] 修改所有默认密码
- [ ] 配置正确的域名
- [ ] 配置 TLS 证书
- [ ] 设置资源限制
- [ ] 配置 HPA
- [ ] 配置监控和告警
- [ ] 配置日志收集
- [ ] 配置备份策略
- [ ] 测试滚动更新
- [ ] 测试回滚流程

### 监控指标

- [ ] Pod 健康状态
- [ ] CPU 和内存使用率
- [ ] 请求延迟和错误率
- [ ] HPA 扩缩容事件
- [ ] 存储使用情况

---

## 🔄 CI/CD 集成

### GitHub Actions 部署

在 `.github/workflows/cd.yml` 中添加 K8s 部署步骤：

```yaml
- name: Deploy to Kubernetes
  run: |
    kubectl set image deployment/nai-tizi-api \
      api=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }} \
      -n nai-tizi
    kubectl rollout status deployment/nai-tizi-api -n nai-tizi
```

---

## 📚 相关文档

- [Docker 部署指南](./Docker部署指南.md)
- [监控系统配置](./监控系统配置.md)
- [快速部署指南](./快速部署指南.md)

---

**文档版本：** v1.0  
**最后更新：** 2026-01-12  
**维护者：** 项目团队
