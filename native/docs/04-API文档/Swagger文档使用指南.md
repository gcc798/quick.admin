# Swagger/OpenAPI 文档使用指南

## 概述

项目已集成 Swagger/OpenAPI 3.0 规范的 API 文档，支持在线查看和测试接口。

## 快速开始

### 1. 生成文档

```bash
# 方式1：使用 Makefile（推荐）
make swagger

# 方式2：直接使用 swag 命令
swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal
```

生成的文件：
```
docs/swagger/
├── docs.go         # Go 代码（自动导入）
├── swagger.json    # JSON 格式文档
└── swagger.yaml    # YAML 格式文档
```

### 2. 启动服务

```bash
# 方式1：使用 Makefile（会自动生成文档）
make run

# 方式2：直接运行
go run cmd/api/main.go
```

### 3. 访问文档

启动服务后，访问：

**Swagger UI**：http://localhost:9009/swagger/index.html

![Swagger UI 示例](https://swagger.io/swagger/media/Images/tools/SwaggerUI.png)

## 导入到 Apifox

### 方式1：URL 导入（推荐）

1. 启动本地服务
2. 在 Apifox 中选择"导入" → "URL/在线链接"
3. 输入：`http://localhost:9009/swagger/doc.json`
4. 点击"确认导入"

**优点**：支持自动同步更新

### 方式2：文件导入

1. 生成文档：`make swagger`
2. 在 Apifox 中选择"导入" → "数据导入"
3. 选择"OpenAPI/Swagger"
4. 上传 `docs/swagger/swagger.json` 或 `swagger.yaml`

### 方式3：自动同步（推荐用于团队）

1. 将服务部署到测试环境
2. 在 Apifox 中配置自动同步
3. URL：`https://your-test-server.com/swagger/doc.json`
4. 设置同步频率（如每天一次）

## Swagger 注释规范

### 总体配置（main.go）

```go
// @title           智控猫 API 文档
// @version         1.0
// @description     智控猫后端 RESTful API 接口文档
// @termsOfService  https://example.com/terms/

// @contact.name   技术支持
// @contact.url    https://example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:9009
// @BasePath  /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 格式: "Bearer {access_token}"

// @tag.name 认证
// @tag.description 用户认证相关接口
```

### 接口注释（Controller）

```go
// Login godoc
// @Summary      用户登录
// @Description  支持多种登录方式：密码登录、邮箱验证码、微信小程序
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body request.LoginRequest true "登录请求参数"
// @Success      200 {object} response.Response{data=response.LoginResponse} "登录成功"
// @Failure      400 {object} response.Response "参数错误"
// @Failure      401 {object} response.Response "认证失败"
// @Router       /login [post]
func (h *authController) Login(c *gin.Context) {
    // ...
}
```

### 结构体注释

```go
// LoginRequest 登录请求
// @Description 统一登录请求参数
type LoginRequest struct {
    ClientKey    string `json:"clientKey" binding:"required" example:"web-admin"` // 客户端Key
    ClientSecret string `json:"clientSecret" binding:"required" example:"web-secret-2024"` // 客户端密钥
    GrantType    string `json:"grantType" binding:"required" example:"password" enums:"password,email,xcx"` // 授权类型
    Username     string `json:"username" example:"admin"` // 用户名
    Password     string `json:"password" example:"admin123"` // 密码
}
```

## 常用注释标签

### 接口级别

| 标签 | 说明 | 示例 |
|------|------|------|
| `@Summary` | 接口简述 | `@Summary 用户登录` |
| `@Description` | 接口详细描述 | `@Description 支持多种登录方式` |
| `@Tags` | 接口分组 | `@Tags 认证` |
| `@Accept` | 接受的内容类型 | `@Accept json` |
| `@Produce` | 返回的内容类型 | `@Produce json` |
| `@Param` | 参数定义 | `@Param request body LoginRequest true "请求参数"` |
| `@Success` | 成功响应 | `@Success 200 {object} Response` |
| `@Failure` | 失败响应 | `@Failure 400 {object} Response` |
| `@Router` | 路由路径 | `@Router /login [post]` |
| `@Security` | 安全认证 | `@Security Bearer` |

### 字段级别

| 标签 | 说明 | 示例 |
|------|------|------|
| `example` | 示例值 | `example:"admin"` |
| `enums` | 枚举值 | `enums:"password,email,xcx"` |
| `minimum` | 最小值 | `minimum:"1"` |
| `maximum` | 最大值 | `maximum:"100"` |
| `minLength` | 最小长度 | `minLength:"6"` |
| `maxLength` | 最大长度 | `maxLength:"20"` |
| `format` | 格式 | `format:"email"` |

## 开发工作流

### 1. 添加新接口

```go
// CreateUser godoc
// @Summary      创建用户
// @Description  创建新用户账号
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body request.CreateUserRequest true "用户信息"
// @Success      200 {object} response.Response{data=response.UserInfo}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Router       /users [post]
func (h *userController) CreateUser(c *gin.Context) {
    // 实现代码
}
```

### 2. 重新生成文档

```bash
make swagger
```

### 3. 测试接口

访问 Swagger UI 测试新接口

### 4. 同步到 Apifox

如果配置了自动同步，Apifox 会自动更新；否则手动重新导入。

## 最佳实践

### 1. 注释规范

✅ **推荐**：
```go
// @Summary      用户登录
// @Description  支持密码、邮箱、微信小程序等多种登录方式
```

❌ **不推荐**：
```go
// @Summary login
// @Description login api
```

### 2. 示例值

✅ **推荐**：提供真实的示例值
```go
Username string `json:"username" example:"admin"`
```

❌ **不推荐**：使用占位符
```go
Username string `json:"username" example:"string"`
```

### 3. 错误响应

✅ **推荐**：详细说明错误场景
```go
// @Failure 400 {object} Response "参数错误：用户名或密码为空"
// @Failure 401 {object} Response "认证失败：用户名或密码错误"
```

❌ **不推荐**：笼统的错误说明
```go
// @Failure 400 {object} Response
```

### 4. 安全认证

需要认证的接口必须添加：
```go
// @Security Bearer
```

### 5. 参数验证

使用 `binding` 标签配合 Swagger 注释：
```go
type LoginRequest struct {
    Username string `json:"username" binding:"required" example:"admin"` // 必填
    Email    string `json:"email" binding:"omitempty,email" example:"admin@example.com"` // 可选，但必须是邮箱格式
}
```

## 常见问题

### Q1: 修改注释后文档没更新？

A: 需要重新生成文档：
```bash
make swagger
```

### Q2: Swagger UI 显示 404？

A: 检查：
1. 是否导入了 docs 包：`_ "github.com/gcc798/nai-tizi/docs/swagger"`
2. 是否注册了路由：`r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))`

### Q3: 结构体字段没有显示？

A: 确保：
1. 字段是导出的（首字母大写）
2. 添加了 `json` 标签
3. 使用了 `--parseDependency --parseInternal` 参数

### Q4: 如何隐藏某些接口？

A: 不添加 Swagger 注释即可，或者使用：
```go
// @Summary      内部接口
// @Description  此接口仅供内部使用
// @Tags         internal
```

### Q5: 如何自定义 Swagger UI 主题？

A: 可以通过配置 ginSwagger 中间件：
```go
url := ginSwagger.URL("http://localhost:9009/swagger/doc.json")
r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
```

## 团队协作

### 1. Git 提交规范

```bash
# 修改接口后，提交时包含文档
git add internal/controller/auth.go
git add docs/swagger/
git commit -m "feat: 添加用户登录接口"
```

### 2. Code Review 检查项

- [ ] 是否添加了 Swagger 注释
- [ ] 注释是否完整（Summary、Description、Param、Success、Failure）
- [ ] 示例值是否真实有效
- [ ] 是否重新生成了文档

### 3. CI/CD 集成

在 CI 流程中添加文档生成检查：

```yaml
# .github/workflows/ci.yml
- name: Generate Swagger Docs
  run: |
    make swagger
    git diff --exit-code docs/swagger/
```

## 参考资源

- **Swagger 官网**：https://swagger.io/
- **OpenAPI 规范**：https://spec.openapis.org/oas/v3.0.0
- **swaggo 文档**：https://github.com/swaggo/swag
- **Apifox 官网**：https://www.apifox.cn/

## 总结

✅ **已完成**：
- Swagger/OpenAPI 3.0 集成
- 认证接口文档完整
- Swagger UI 在线查看
- Apifox 导入支持
- Makefile 便捷命令

📝 **下一步**：
1. 为其他模块添加 Swagger 注释
2. 配置 Apifox 自动同步
3. 在 CI/CD 中集成文档检查
