# 字符串ID转换中间件使用指南

## 概述

字符串ID转换中间件（StringIDConverter）用于解决前端传递字符串类型ID时后端无法解析的问题。这是因为JavaScript中的大数（超过 `Number.MAX_SAFE_INTEGER = 9007199254740991`）会丢失精度，因此前端通常将大数ID作为字符串传递。

## 问题背景

### JavaScript 大数精度问题

```javascript
// JavaScript 中的问题
const id = 9007199254740992; // 超过 MAX_SAFE_INTEGER
console.log(id === 9007199254740993); // true (错误！应该是 false)

// 解决方案：使用字符串
const idStr = "9007199254740992";
```

### 后端解析问题

当前端发送字符串ID时，Gin的JSON绑定会失败：

```json
// 前端发送
{
  "ids": ["1234567890123456789", "9876543210987654321"]
}

// 后端期望
type Request struct {
  IDs []int64 `json:"ids"`
}

// 结果：绑定失败，因为字符串无法直接转换为 int64
```

## 中间件功能

StringIDConverter 中间件自动处理以下场景：

### 1. 路径参数中的ID

```go
// 路由定义
r.GET("/api/v1/user/:id", handler)

// 前端请求
GET /api/v1/user/1234567890123456789

// 中间件自动转换并存储到上下文
c.Set("parsed_id", int64(1234567890123456789))
```

### 2. JSON请求体中的ID字段

中间件会自动识别并转换以下ID相关字段：

- 单个ID字段：`id`, `userId`, `deptId`, `roleId`, `menuId`, `configId`, `dictId`, `parentId`, `orgId`, `storageId`, `attachmentId`, `clientId`, `createBy`, `updateBy`
- ID数组字段：`ids`, `userIds`, `deptIds`, `roleIds`, `menuIds`, `configIds`, `dictIds`, `parentIds`, `orgIds`, `storageIds`, `attachmentIds`, `clientIds`

#### 示例 1：单个ID字段

```json
// 前端发送
{
  "userId": "1234567890123456789",
  "deptId": "9876543210987654321"
}

// 中间件自动转换为
{
  "userId": 1234567890123456789,
  "deptId": 9876543210987654321
}

// 后端正常绑定
type Request struct {
  UserId int64 `json:"userId"`
  DeptId int64 `json:"deptId"`
}
```

#### 示例 2：ID数组字段

```json
// 前端发送
{
  "ids": ["1234567890123456789", "9876543210987654321"]
}

// 中间件自动转换为
{
  "ids": [1234567890123456789, 9876543210987654321]
}

// 后端正常绑定
type Request struct {
  IDs []int64 `json:"ids"`
}
```

#### 示例 3：嵌套对象

```json
// 前端发送
{
  "user": {
    "id": "1234567890123456789",
    "deptId": "9876543210987654321"
  },
  "roles": [
    {"roleId": "111111111111111111"},
    {"roleId": "222222222222222222"}
  ]
}

// 中间件递归转换所有ID字段
{
  "user": {
    "id": 1234567890123456789,
    "deptId": 9876543210987654321
  },
  "roles": [
    {"roleId": 111111111111111111},
    {"roleId": 222222222222222222}
  ]
}
```

## 使用方式

### 1. 全局中间件（推荐）

在 `cmd/api/main.go` 中已经配置为全局中间件：

```go
// 添加字符串ID转换中间件（处理前端传递的字符串ID）
r.Use(middleware.StringIDConverter())
```

这样所有路由都会自动应用此中间件。

### 2. 路由组中间件

如果只想在特定路由组中使用：

```go
api := r.Group("/api/v1")
api.Use(middleware.StringIDConverter())
{
    api.POST("/user", handler)
    api.PUT("/user/:id", handler)
}
```

### 3. 单个路由中间件

如果只想在特定路由中使用：

```go
r.POST("/api/v1/user", middleware.StringIDConverter(), handler)
```

## 控制器中的使用

### 获取路径参数ID

```go
func (h *UserController) GetById(c *gin.Context) {
    // 方式1：使用工具函数（推荐）
    userId, err := utils.ParseInt64Param(c, "id", "required")
    if err != nil {
        response.FailCode(c, response.CodeInvalidParam, err.Error())
        return
    }
    
    // 方式2：从上下文获取已转换的ID
    if parsedId, exists := c.Get("parsed_id"); exists {
        userId := parsedId.(int64)
        // 使用 userId
    }
}
```

### JSON请求体绑定

```go
func (h *UserController) BatchDelete(c *gin.Context) {
    var req request.BatchDeleteUsersRequest
    // 中间件已经将字符串ID转换为数字，直接绑定即可
    if err := c.ShouldBindJSON(&req); err != nil {
        response.FailCode(c, response.CodeInvalidParam, "参数错误: "+err.Error())
        return
    }
    
    // req.IDs 已经是 []int64 类型
    if err := h.userService.BatchDelete(c.Request.Context(), req.IDs); err != nil {
        response.FailWithMsg(c, err.Error())
        return
    }
}
```

## 前端使用示例

### 1. 单个ID

```typescript
// 获取用户详情
export function getUserById(id: string) {
  return request.get(`/api/v1/user/${id}`);
}

// 调用
getUserById('1234567890123456789');
```

### 2. JSON请求体中的ID

```typescript
// 批量删除用户
export function batchDeleteUsers(ids: string[]) {
  return request.delete('/api/v1/user/batch', {
    data: { ids }
  });
}

// 调用
batchDeleteUsers(['1234567890123456789', '9876543210987654321']);
```

### 3. 更新请求

```typescript
// 更新用户
export function updateUser(id: string, data: UpdateUserRequest) {
  return request.put(`/api/v1/user/${id}`, {
    ...data,
    deptId: data.deptId.toString(), // 确保ID是字符串
  });
}
```

## 支持的ID字段列表

中间件会自动识别并转换以下字段：

### 单个ID字段
- `id`
- `userId`
- `deptId`
- `roleId`
- `menuId`
- `configId`
- `dictId`
- `parentId`
- `orgId`
- `storageId`
- `attachmentId`
- `clientId`
- `createBy`
- `updateBy`

### ID数组字段
- `ids`
- `userIds`
- `deptIds`
- `roleIds`
- `menuIds`
- `configIds`
- `dictIds`
- `parentIds`
- `orgIds`
- `storageIds`
- `attachmentIds`
- `clientIds`

## 扩展字段

如果需要支持新的ID字段，修改 `internal/middleware/string_id_converter.go` 中的 `isIDField` 函数：

```go
func isIDField(fieldName string) bool {
    idFields := []string{
        "id", "ids",
        "userId", "userIds",
        // ... 现有字段
        "newId", "newIds", // 添加新字段
    }
    
    for _, field := range idFields {
        if fieldName == field {
            return true
        }
    }
    return false
}
```

## 注意事项

1. **中间件顺序**：StringIDConverter 应该在 JSON 绑定之前执行，建议作为全局中间件尽早注册

2. **性能影响**：中间件会读取并解析请求体，对于大型请求可能有轻微性能影响

3. **错误处理**：如果字符串无法转换为有效的 int64，中间件会保持原值，由后续的参数验证处理错误

4. **兼容性**：中间件同时支持字符串和数字类型的ID，不会影响已经使用数字ID的接口

5. **递归转换**：中间件会递归处理嵌套对象和数组中的ID字段

## 测试示例

### 测试用例 1：路径参数

```bash
# 字符串ID
curl -X GET "http://localhost:9009/api/v1/user/1234567890123456789"

# 数字ID（仍然支持）
curl -X GET "http://localhost:9009/api/v1/user/123"
```

### 测试用例 2：JSON请求体

```bash
# 字符串ID数组
curl -X DELETE "http://localhost:9009/api/v1/user/batch" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["1234567890123456789", "9876543210987654321"]}'

# 混合类型（字符串和数字）
curl -X DELETE "http://localhost:9009/api/v1/user/batch" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["1234567890123456789", 123]}'
```

### 测试用例 3：嵌套对象

```bash
curl -X POST "http://localhost:9009/api/v1/user" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "1234567890123456789",
    "deptId": "9876543210987654321",
    "roles": [
      {"roleId": "111111111111111111"},
      {"roleId": "222222222222222222"}
    ]
  }'
```

## 总结

字符串ID转换中间件提供了一个透明的解决方案，使得前端可以安全地使用字符串传递大数ID，而后端仍然可以使用 int64 类型处理。这个中间件：

- ✅ 自动转换路径参数和JSON请求体中的ID字段
- ✅ 支持单个ID和ID数组
- ✅ 递归处理嵌套对象
- ✅ 向后兼容数字类型ID
- ✅ 零侵入，无需修改现有代码
- ✅ 易于扩展新的ID字段

通过使用这个中间件，可以彻底解决JavaScript大数精度丢失的问题，确保前后端ID传递的准确性。
