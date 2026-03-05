# 接口 RESTful 设计规范

## 1. 核心原则

本项目遵循 RESTful 风格设计 API 接口：

1. **HTTP 方法语义化**
   - `GET`: 查询资源（列表或详情）
   - `POST`: 创建资源
   - `PUT`: 更新资源（全量或部分）
   - `DELETE`: 删除资源

2. **资源路径名词化**
   - 使用名词表示资源（如 `user`, `role`, `org`），避免使用动词（如 `getUser`）。
   - 路径层级体现资源关系。

3. **ID 在路径中**
   - 操作特定资源时，ID 应作为路径参数：`/resource/:id`。

## 2. 接口设计模式

### 2.1 基础 CRUD

| 操作 | HTTP 方法 | 路径示例 | 说明 |
|------|----------|----------|------|
| 分页查询 | `GET` | `/api/v1/user` | 使用查询参数控制分页 |
| 查询详情 | `GET` | `/api/v1/user/:id` | |
| 创建资源 | `POST` | `/api/v1/user` | |
| 更新资源 | `PUT` | `/api/v1/user/:id` | |
| 删除资源 | `DELETE` | `/api/v1/user/:id` | |

### 2.2 批量操作

| 操作 | HTTP 方法 | 路径示例 | 说明 |
|------|----------|----------|------|
| 批量删除 | `DELETE` | `/api/v1/user/batch` | ID列表放在 Body 中 |
| 批量导入 | `POST` | `/api/v1/user/import` | |

## 3. 现有模块路由概览

### 用户管理 (User)
- `GET /api/v1/user` (分页)
- `GET /api/v1/user/:id` (详情)
- `POST /api/v1/user` (创建)
- `PUT /api/v1/user/:id` (更新)
- `DELETE /api/v1/user/:id` (删除)

### 角色管理 (Role)
- `GET /api/v1/role` (分页)
- `GET /api/v1/role/all` (获取所有)
