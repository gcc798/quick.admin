# EFK 日志聚合指南

## 📋 概述

本文档提供完整的 EFK Stack（Elasticsearch、Fluent Bit、Kibana）日志聚合解决方案。EFK 是 ELK 的轻量级替代方案，使用 Fluent Bit 替代 Logstash 和 Filebeat，显著降低资源消耗。

---

## 🎯 EFK Stack 架构

```
应用日志 → Fluent Bit → Elasticsearch → Kibana
```

**组件说明：**
- **Fluent Bit**: 轻量级日志收集和处理器（C 语言编写，内存占用 ~450KB）
- **Elasticsearch**: 分布式搜索和分析引擎，存储日志数据
- **Kibana**: 可视化平台，提供日志查询和分析界面

**相比 ELK 的优势：**
- 内存占用减少 ~300-400MB（移除 Logstash JVM）
- 启动时间从 ~60s 降至 ~10s
- 配置更简洁，单一配置文件
- 更适合容器化环境
- 性能更高，资源消耗更低

---

## 📦 配置文件

1. **docker-compose.efk.yml** - EFK Stack 服务配置
2. **monitoring/fluent-bit/fluent-bit.conf** - Fluent Bit 配置

---

## 🚀 快速开始

### 前置要求

- Docker 和 Docker Compose
- 至少 2GB 可用内存
- 应用网络已创建

### 启动 EFK Stack

```bash
# 1. 确保应用网络存在
docker network create app-network

# 2. 启动 EFK Stack
docker-compose -f docker-compose.efk.yml up -d

# 3. 等待服务就绪（约 30 秒）
docker-compose -f docker-compose.efk.yml ps

# 4. 验证服务
curl http://localhost:9200  # Elasticsearch
curl http://localhost:5601  # Kibana
```

### 访问服务

- **Elasticsearch**: http://localhost:9200
- **Kibana**: http://localhost:5601

---

## 📝 配置详解

### Fluent Bit 配置

**文件**: `monitoring/fluent-bit/fluent-bit.conf`

**功能**:
- 监控应用日志文件 (`/var/log/app/*.log`)
- 监控 Docker 容器日志
- 自动解析 JSON 格式日志
- 添加服务和环境元数据
- 直接输出到 Elasticsearch

**关键配置**:
```ini
[INPUT]
    Name              tail
    Path              /var/log/app/*.log
    Parser            json
    Tag               app.*

[FILTER]
    Name                record_modifier
    Match               app.*
    Record service      nai-tizi-api
    Record environment  production

[OUTPUT]
    Name            es
    Host            elasticsearch
    Port            9200
    Logstash_Format On
    Logstash_Prefix nai-tizi
```

### Elasticsearch 配置

**Docker Compose 配置**:
```yaml
elasticsearch:
  environment:
    - discovery.type=single-node
    - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    - xpack.security.enabled=false
```

**索引模式**: `nai-tizi-YYYY.MM.dd`

---

## 🔧 使用指南

### 查看日志

#### 1. 使用 Kibana

访问 http://localhost:5601

**首次配置**:
1. 打开 Kibana
2. 进入 Management → Stack Management → Index Patterns
3. 创建索引模式：`nai-tizi-*`
4. 选择时间字段：`@timestamp`
5. 进入 Discover 查看日志

**常用查询**:
```
# 查询特定服务的日志
service: "nai-tizi-api"

# 查询错误日志
level: "error"

# 查询特定时间范围
@timestamp: [now-1h TO now]

# 组合查询
service: "nai-tizi-api" AND level: "error"
```

#### 2. 使用 Elasticsearch API

```bash
# 查询所有索引
curl http://localhost:9200/_cat/indices?v

# 搜索日志
curl -X GET "http://localhost:9200/nai-tizi-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "level": "error"
    }
  },
  "size": 10
}'

# 查询最近的日志
curl -X GET "http://localhost:9200/nai-tizi-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match_all": {}
  },
  "sort": [
    {
      "@timestamp": {
        "order": "desc"
      }
    }
  ],
  "size": 10
}'
```

### 创建 Kibana Dashboard

1. 进入 Kibana → Dashboard
2. 创建新 Dashboard
3. 添加可视化组件：
   - **日志数量趋势**: 时间序列图
   - **日志级别分布**: 饼图
   - **错误日志列表**: 数据表
   - **请求延迟**: 直方图

---

## 📊 日志格式

### 应用日志格式

应用应输出 JSON 格式的日志：

```json
{
  "timestamp": "2026-01-16T14:30:00+08:00",
  "level": "info",
  "service": "nai-tizi-api",
  "message": "HTTP request processed",
  "method": "GET",
  "path": "/api/v1/users",
  "status": 200,
  "duration": 45,
  "client_ip": "192.168.1.100"
}
```

### Go 应用日志配置

使用 Zap 日志库输出 JSON 格式：

```go
// 配置 JSON 编码器
encoderConfig := zap.NewProductionEncoderConfig()
encoderConfig.TimeKey = "timestamp"
encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

// 创建 logger
logger := zap.New(
    zapcore.NewCore(
        zapcore.NewJSONEncoder(encoderConfig),
        zapcore.AddSync(logFile),
        zapcore.InfoLevel,
    ),
)
```

---

## 🔍 故障排查

### 常见问题

**1. Elasticsearch 无法启动**

```bash
# 检查日志
docker-compose -f docker-compose.efk.yml logs elasticsearch

# 常见原因：内存不足
# 解决：增加 Docker 内存限制或减少 ES_JAVA_OPTS
```

**2. Fluent Bit 无法连接 Elasticsearch**

```bash
# 检查 Fluent Bit 日志
docker-compose -f docker-compose.efk.yml logs fluent-bit

# 检查 Elasticsearch 是否就绪
curl http://localhost:9200

# 重启 Fluent Bit
docker-compose -f docker-compose.efk.yml restart fluent-bit
```

**3. Kibana 中看不到日志**

```bash
# 检查索引是否存在
curl http://localhost:9200/_cat/indices?v

# 检查 Fluent Bit 是否在发送数据
docker-compose -f docker-compose.efk.yml logs fluent-bit | grep "elasticsearch"

# 重新创建索引模式
# 在 Kibana 中删除并重新创建索引模式
```

**4. 日志数据过多**

```bash
# 查看索引大小
curl http://localhost:9200/_cat/indices?v

# 删除旧索引
curl -X DELETE "http://localhost:9200/nai-tizi-2026.01.01"
```

---

## 🎯 生产环境优化

### 1. 性能优化

**Elasticsearch**:
```yaml
environment:
  - "ES_JAVA_OPTS=-Xms2g -Xmx2g"  # 增加内存
  - cluster.name=nai-tizi-cluster
  - bootstrap.memory_lock=true
```

**Fluent Bit**:
```ini
[SERVICE]
    Flush        1
    Log_Level    warn

[INPUT]
    Mem_Buf_Limit     10MB
```

### 2. 数据保留策略

**索引生命周期管理**:
```bash
# 创建 ILM 策略
curl -X PUT "http://localhost:9200/_ilm/policy/nai-tizi-policy" -H 'Content-Type: application/json' -d'
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover": {
            "max_size": "50GB",
            "max_age": "7d"
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}'
```

### 3. 安全配置

**启用 Elasticsearch 安全**:
```yaml
environment:
  - xpack.security.enabled=true
  - ELASTIC_PASSWORD=your_password
```

---

## 📈 监控 EFK Stack

### Elasticsearch 健康检查

```bash
# 集群健康
curl http://localhost:9200/_cluster/health?pretty

# 节点信息
curl http://localhost:9200/_nodes/stats?pretty

# 索引统计
curl http://localhost:9200/_stats?pretty
```

### Fluent Bit 监控

```bash
# 查看日志输出
docker-compose -f docker-compose.efk.yml logs -f fluent-bit

# 检查处理的记录数
docker-compose -f docker-compose.efk.yml logs fluent-bit | grep "records"
```

---

## 🔄 与其他系统集成

### 与 Prometheus 集成

使用 Elasticsearch Exporter 暴露指标：

```yaml
elasticsearch-exporter:
  image: quay.io/prometheuscommunity/elasticsearch-exporter:latest
  ports:
    - "9114:9114"
  command:
    - '--es.uri=http://elasticsearch:9200'
  networks:
    - app-network
```

### 与 Grafana 集成

在 Grafana 中添加 Elasticsearch 数据源：
1. Configuration → Data Sources → Add data source
2. 选择 Elasticsearch
3. URL: http://elasticsearch:9200
4. Index name: nai-tizi-*
5. Time field: @timestamp

---

## 📚 相关文档

- [Docker 部署指南](./Docker部署指南.md)
- [监控系统配置](./监控系统配置.md)
- [Kubernetes 部署指南](./Kubernetes部署指南.md)

---

## 🎓 最佳实践

1. **日志格式标准化**: 使用 JSON 格式，包含必要字段
2. **合理的日志级别**: 生产环境使用 INFO 及以上级别
3. **日志轮转**: 配置日志文件大小和保留策略
4. **索引管理**: 使用 ILM 自动管理索引生命周期
5. **性能监控**: 定期检查 Elasticsearch 性能指标
6. **备份策略**: 定期备份重要索引数据
7. **资源限制**: 为 Fluent Bit 设置合理的内存限制

---

**文档版本：** v1.0  
**最后更新：** 2026-01-16  
**维护者：** 项目团队
