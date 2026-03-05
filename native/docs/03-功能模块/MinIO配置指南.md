# MinIO 配置指南

## 问题：无法访问 MinIO 中的文件

### 原因分析
1. **桶权限未设置为公开** - 默认情况下桶是私有的
2. **URL 地址问题** - `localhost:9000` 只能在服务器本地访问
3. **CORS 配置** - 浏览器跨域访问被阻止

## 解决方案

### 方案 1：设置桶为公开访问（推荐用于公开文件）

#### 1.1 通过 MinIO Console 设置
```bash
# 访问 MinIO Console
http://localhost:9001

# 登录后：
# 1. 进入 Buckets
# 2. 选择你的桶（nati-tizi）
# 3. 点击 "Manage" -> "Access Policy"
# 4. 选择 "Public" 或添加自定义策略
```

#### 1.2 通过 mc 命令行设置
```bash
# 安装 mc 客户端
brew install minio/stable/mc  # macOS
# 或
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc

# 配置 mc
mc alias set myminio http://localhost:9000 minioadmin minioadmin

# 设置桶为公开只读
mc anonymous set download myminio/nati-tizi

# 或设置为完全公开（读写）
mc anonymous set public myminio/nati-tizi

# 查看当前策略
mc anonymous get myminio/nati-tizi
```

#### 1.3 通过代码设置桶策略
```go
import (
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

// 设置桶策略为公开只读
policy := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {"AWS": ["*"]},
            "Action": ["s3:GetObject"],
            "Resource": ["arn:aws:s3:::nati-tizi/*"]
        }
    ]
}`

err := minioClient.SetBucketPolicy(ctx, "nati-tizi", policy)
```

### 方案 2：使用预签名 URL（推荐用于私有文件）

修改代码，使用临时 URL 而不是永久 URL：

```go
// 在 BindToBusiness 中
if req.IsPublic {
    // 生成 7 天有效期的临时 URL
    accessUrl, err = stor.GetURL(ctx, attachment.FileKey, 7*24*time.Hour)
    if err != nil {
        s.logger.Warn("获取访问 URL 失败", zap.Error(err))
    }
}
```

### 方案 3：配置外部访问地址

如果需要从外部访问，需要配置 MinIO 的外部地址：

#### 3.1 修改 MinIO 启动参数
```bash
# docker-compose.yml
environment:
  MINIO_ROOT_USER: minioadmin
  MINIO_ROOT_PASSWORD: minioadmin
  MINIO_SERVER_URL: http://your-domain.com:9000  # 外部访问地址
```

#### 3.2 修改存储配置
在数据库的 `s_storage_env` 表中，修改 config 字段：
```json
{
  "endpoint": "your-domain.com:9000",
  "accessKeyId": "minioadmin",
  "secretAccessKey": "minioadmin",
  "useSSL": false,
  "bucket": "nati-tizi",
  "region": "us-east-1"
}
```

### 方案 4：通过后端代理访问（最安全）

不直接暴露 MinIO URL，而是通过后端 API 代理访问：

```go
// 在 attachment controller 中已经有了
// GET /api/v1/attachments/:attachmentId/download
// 这个接口会：
// 1. 验证权限
// 2. 从 MinIO 读取文件
// 3. 返回文件流给前端
```

前端使用：
```javascript
// 不使用 accessUrl，而是使用下载接口
const downloadUrl = `/api/v1/attachments/${attachmentId}/download`;
```

## 推荐配置

### 开发环境
```bash
# 1. 设置桶为公开（方便测试）
mc anonymous set download myminio/nati-tizi

# 2. 验证
curl http://localhost:9000/nati-tizi/temp/20251227/xxx.png
```

### 生产环境
```bash
# 1. 保持桶私有
# 2. 使用预签名 URL（临时访问）
# 3. 或通过后端 API 代理访问
```

## 验证步骤

### 1. 检查 MinIO 是否运行
```bash
curl http://localhost:9000/minio/health/live
```

### 2. 检查桶是否存在
```bash
mc ls myminio/nati-tizi
```

### 3. 检查文件是否存在
```bash
mc ls myminio/nati-tizi/temp/20251227/
```

### 4. 测试文件访问
```bash
# 如果桶是公开的
curl http://localhost:9000/nati-tizi/temp/20251227/xxx.png -I

# 应该返回 200 OK
```

## 常见错误

### AccessDenied
```xml
<Error>
  <Code>AccessDenied</Code>
  <Message>Access Denied</Message>
</Error>
```
**解决**：设置桶为公开或使用预签名 URL

### NoSuchBucket
```xml
<Error>
  <Code>NoSuchBucket</Code>
  <Message>The specified bucket does not exist</Message>
</Error>
```
**解决**：检查桶名是否正确，或创建桶

### Connection Refused
```
curl: (7) Failed to connect to localhost port 9000
```
**解决**：检查 MinIO 是否运行
