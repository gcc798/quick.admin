# Swagger 快速参考

## 🚀 快速命令

```bash
# 生成文档
make swagger

# 运行服务（自动生成文档）
make run

# 格式化注释
make swagger-fmt

# 编译项目
make build
```

## 📍 访问地址

| 服务 | 地址 |
|------|------|
| **Swagger UI** | http://localhost:9009/swagger/index.html |
| **JSON 文档** | http://localhost:9009/swagger/doc.json |
| **YAML 文档** | 文件：`docs/swagger/swagger.yaml` |

## 📥 导入 Apifox

### 方式1：URL 导入（推荐）
```
http://localhost:9009/swagger/doc.json
```

### 方式2：文件导入
```
docs/swagger/swagger.json
```

## 📝 注释模板

### 接口注释

```go
// FunctionName godoc
// @Summary      接口简述
// @Description  接口详细描述
// @Tags         分组名称
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        name path string true "路径参数"
// @Param        request body RequestType true "请求体"
// @Success      200 {object} Response{data=ResponseType} "成功"
// @Failure      400 {object} Response "参数错误"
// @Failure      401 {object} Response "未授权"
// @Router       /path [method]
func FunctionName(c *gin.Context) {
    // ...
}
```

### 结构体注释

```go
// TypeName 类型说明
// @Description 详细描述
type TypeName struct {
    Field1 string `json:"field1" example:"示例值"` // 字段说明
    Field2 int    `json:"field2" example:"123" minimum:"1"` // 带验证
    Field3 string `json:"field3" enums:"a,b,c"` // 枚举值
}
```

## 🏷️ 常用标签

| 标签 | 用途 | 示例 |
|------|------|------|
| `@Summary` | 简短描述 | `@Summary 用户登录` |
| `@Description` | 详细描述 | `@Description 支持多种登录方式` |
| `@Tags` | 分组 | `@Tags 认证` |
| `@Param` | 参数 | `@Param id path int true "用户ID"` |
| `@Success` | 成功响应 | `@Success 200 {object} Response` |
| `@Failure` | 失败响应 | `@Failure 400 {object} Response` |
| `@Router` | 路由 | `@Router /users/{id} [get]` |
| `@Security` | 认证 | `@Security Bearer` |

## 🔧 字段标签

| 标签 | 说明 | 示例 |
|------|------|------|
| `example` | 示例值 | `example:"admin"` |
| `enums` | 枚举 | `enums:"a,b,c"` |
| `minimum` | 最小值 | `minimum:"1"` |
| `maximum` | 最大值 | `maximum:"100"` |
| `minLength` | 最小长度 | `minLength:"6"` |
| `maxLength` | 最大长度 | `maxLength:"20"` |
| `format` | 格式 | `format:"email"` |

## 📦 参数类型

| 位置 | 说明 | 示例 |
|------|------|------|
| `path` | 路径参数 | `/users/{id}` |
| `query` | 查询参数 | `/users?name=admin` |
| `header` | 请求头 | `Authorization: Bearer xxx` |
| `body` | 请求体 | JSON 数据 |
| `formData` | 表单 | `multipart/form-data` |

## 🎯 响应示例

```go
// @Success 200 {object} Response{data=UserInfo} "成功"
// @Success 200 {object} PageResponse{rows=[]UserInfo} "分页"
// @Success 200 {object} Response{data=string} "返回字符串"
// @Success 200 {object} Response{data=int} "返回数字"
// @Success 200 {object} Response{data=[]string} "返回数组"
```

## ⚠️ 常见错误

### 1. 文档没更新
```bash
# 解决：重新生成
make swagger
```

### 2. 结构体字段不显示
```bash
# 解决：使用完整参数
swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal
```

### 3. Swagger UI 404
```go
// 解决：检查导入
import _ "github.com/gcc798/quick.admin/docs/swagger"
```

## 📚 完整示例

```go
// Login godoc
// @Summary      用户登录
// @Description  支持密码、邮箱、微信小程序等多种登录方式
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body request.LoginRequest true "登录请求参数"
// @Success      200 {object} response.Response{data=response.LoginResponse} "登录成功，返回 Token"
// @Failure      400 {object} response.Response "参数错误"
// @Failure      401 {object} response.Response "认证失败"
// @Router       /login [post]
func (h *authController) Login(c *gin.Context) {
    var req request.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.FailCode(c, response.CodeInvalidParam, "参数错误")
        return
    }
    // ... 业务逻辑
    response.Success(c, loginResponse)
}
```

## 🔗 相关链接

- **Swagger UI**: http://localhost:9009/swagger/index.html
- **详细文档**: [Swagger文档使用指南.md](./Swagger文档使用指南.md)
- **swaggo GitHub**: https://github.com/swaggo/swag
- **OpenAPI 规范**: https://spec.openapis.org/oas/v3.0.0
