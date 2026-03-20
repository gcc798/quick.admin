# Swagger 注释规范

## 核心原则

**只注释业务特定的错误场景，避免冗余的通用错误码注释**

---

## 规范说明

### 1. 应该注释的错误码

✅ **业务特定的错误**
- 特定业务场景的错误（如"短信功能未启用"、"用户名密码错误"）
- 非标准的状态码使用
- 需要特别说明的错误场景
- 400 错误如果需要说明具体参数问题

✅ **示例：**
```go
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "认证失败"
// @Failure 403 {object} response.Response "短信验证码功能未启用"
// @Failure 404 {object} response.Response "用户不存在"
```

### 2. 不需要注释的错误码

❌ **通用的错误码**
- 通用的 401（未授权）
- 通用的 403（无权限）
- 通用的 500（服务器错误）
- 标准的 400（参数错误）- 除非需要说明具体参数

❌ **避免的冗余注释：**
```go
// ❌ 不推荐
// @Failure 401 {object} response.Response "未登录或 Token 无效"
// @Failure 403 {object} response.Response "无权限访问"
// @Failure 500 {object} response.Response "服务器内部错误"

// ✅ 推荐（只在全局文档说明）
```

---

## 全局错误码说明

在 Swagger 配置中统一说明通用错误码：

```go
// @title           项目 API 文档
// @version         1.0
// @description     API 接口文档
// @description     
// @description     通用错误码说明：
// @description     - 400: 请求参数错误
// @description     - 401: 未授权或认证失败
// @description     - 403: 无权限访问
// @description     - 404: 资源不存在
// @description     - 500: 服务器内部错误
```

---

## 标准注释模板

### 模板 1：基础 CRUD 接口

```go
// CreateXXX 创建XXX
// @Summary 创建XXX
// @Description 创建新的XXX
// @Tags XXX管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param body body request.CreateXXXRequest true "XXX信息"
// @Success 200 {object} response.Response{data=response.XXXResponse}
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/xxx [post]
```

### 模板 2：查询接口

```go
// GetXXX 获取XXX详情
// @Summary 获取XXX详情
// @Description 根据ID获取XXX详情
// @Tags XXX管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path int true "XXX ID"
// @Success 200 {object} response.Response{data=response.XXXResponse}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "XXX不存在"
// @Router /api/v1/xxx/{id} [get]
```

### 模板 3：列表接口

```go
// ListXXX 获取XXX列表
// @Summary 获取XXX列表
// @Description 分页获取XXX列表
// @Tags XXX管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param pageNum query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResponse}
// @Router /api/v1/xxx [get]
```

### 模板 4：删除接口

```go
// DeleteXXX 删除XXX
// @Summary 删除XXX
// @Description 删除指定XXX
// @Tags XXX管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param id path int true "XXX ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/xxx/{id} [delete]
```

### 模板 5：特殊业务接口

```go
// SendSmsCode 发送短信验证码
// @Summary 发送短信验证码
// @Description 向指定手机号发送短信验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body request.SendSmsCodeRequest true "手机号"
// @Success 200 {object} response.Response{data=object{message=string}}
// @Failure 400 {object} response.Response "手机号格式错误"
// @Failure 403 {object} response.Response "短信验证码功能未启用"
// @Router /auth/sms [post]
```

---

## 实施建议

### 新接口开发

1. 使用标准模板
2. 只添加业务特定的错误码注释
3. 保持注释简洁明了

### 现有接口优化

1. 移除通用错误码注释（401/403/500）
2. 保留业务特定错误码注释
3. 统一注释格式

---

## 优化前后对比

### 优化前（冗余）

```go
// Login godoc
// @Summary      用户登录
// @Description  支持多种登录方式
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body request.LoginRequest true "登录请求参数"
// @Success      200 {object} response.Response{data=response.LoginResponse} "登录成功，返回 AccessToken 和 RefreshToken"
// @Failure      400 {object} response.Response "参数错误"
// @Failure      401 {object} response.Response "认证失败：用户名密码错误、验证码错误、客户端认证失败等"
// @Failure      500 {object} response.Response "服务器内部错误"
// @Router       /login [post]
```

### 优化后（简洁）

```go
// Login godoc
// @Summary      用户登录
// @Description  支持多种登录方式：密码登录(password)、邮箱验证码(email)、微信小程序(xcx)
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body request.LoginRequest true "登录请求参数"
// @Success      200 {object} response.Response{data=response.LoginResponse}
// @Failure      400 {object} response.Response "参数错误"
// @Failure      401 {object} response.Response "认证失败"
// @Router       /login [post]
```

---

## 优化效果

### 代码简化

- 每个接口减少 2-4 行冗余注释
- 文档更简洁易读
- 维护成本降低

### 可维护性提升

- 通用错误码在全局统一说明
- 修改通用错误格式只需改一处
- 业务错误码更突出

---

## 检查清单

在编写或审查 Swagger 注释时，使用以下检查清单：

- [ ] 是否包含 @Summary 和 @Description
- [ ] 是否包含 @Tags 分类
- [ ] 是否包含 @Accept 和 @Produce
- [ ] 是否包含必要的 @Param 参数说明
- [ ] 是否包含 @Success 成功响应
- [ ] 是否只包含业务特定的 @Failure 错误
- [ ] 是否移除了通用错误码（401/403/500）
- [ ] 是否包含正确的 @Router 路由

---

**制定时间：** 2024-12-20  
**适用范围：** 所有 API 接口  
**执行优先级：** 高

## 相关文档

- [代码规范](./代码规范.md) - 项目代码规范
- [开发规范指南](./开发规范指南.md) - 开发规范指南
