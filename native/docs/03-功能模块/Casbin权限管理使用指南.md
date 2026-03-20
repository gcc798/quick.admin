# Casbin RBAC 权限管理使用指南

## 目录
1. [系统概述](#系统概述)
2. [核心概念](#核心概念)
3. [权限模型](#权限模型)
4. [快速开始](#快速开始)
5. [API 使用示例](#api-使用示例)
6. [通配符权限](#通配符权限)
7. [角色继承](#角色继承)
8. [最佳实践](#最佳实践)

---

## 系统概述

本系统基于 Casbin 实现了完整的 RBAC（基于角色的访问控制）权限管理，支持以下特性：

- ✅ **多租户隔离**：不同组织的权限完全隔离
- ✅ **通配符权限**：支持 `user.*`、`*.read`、`*` 等通配符
- ✅ **角色继承**：角色可以继承多个父角色的权限
- ✅ **超级管理员**：admin 角色自动拥有所有权限
- ✅ **细粒度控制**：可精确到每个接口的权限控制
- ✅ **继承深度限制**：最多支持 3 层角色继承
- ✅ **循环继承检测**：自动检测并阻止循环继承

---

## 核心概念

### 1. 用户（User）
系统中的实际用户，通过 `user_id` 标识。

### 2. 角色（Role）
权限的集合，通过 `role_key` 标识（例如：`admin`、`user_manager`）。

### 3. 组织（Organization）
多租户的基本单位，通过 `org_id` 标识。用户在不同组织可以有不同的角色。

### 4. 权限（Permission）
由 **资源（Resource）** 和 **操作（Action）** 组成：
- 资源：例如 `user.create`、`device.read`
- 操作：例如 `read`、`write`、`delete`

### 5. 菜单（Menu）
前端菜单和按钮，每个菜单可以关联权限标识（`perms`）。

---

## 权限模型

### Casbin 模型配置
位置：`cmd/api/casbin_model.conf`

```ini
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && keyMatch2(r.act, p.act) || p.sub == "role::admin" && g(r.sub, p.sub, r.dom) && r.dom == p.dom
```

### 关键说明
1. **keyMatch2**：支持通配符匹配（`*`）
2. **admin 特权**：`p.sub == "role::admin"` 确保 admin 角色拥有所有权限
3. **多租户**：通过 `dom`（domain）实现组织隔离

---

## 快速开始

### 1. 初始化数据库
执行 SQL 脚本：
```bash
psql -U postgres -d nai-tizi -f scripts/sql/002_init_rbac_tables.sql
```

### 2. 默认数据
系统会自动创建：
- **默认组织**：ID=1，名称="默认组织"
- **默认角色**：
  - `admin`：超级管理员（拥有所有权限）
  - `user_manager`：用户管理员（拥有 `user.*` 权限）
  - `viewer`：访客（拥有 `*.read` 权限）
- **默认用户**：user_id=1 拥有 admin 角色

### 3. 测试权限
```bash
# 登录获取 Token
curl -X POST http://localhost:9009/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"clientKey":"web_client","clientSecret":"web_secret_2024","username":"admin","password":"admin123"}'

# 使用 Token 访问受保护的接口
curl -X GET http://localhost:9009/api/v1/roles \
  -H "Authorization: Bearer {your_access_token}"
```

---

## API 使用示例

### 1. 创建角色
```bash
POST /api/v1/roles
Authorization: Bearer {token}
Content-Type: application/json

{
  "roleKey": "device_manager",
  "roleName": "设备管理员",
  "roleSort": 10,
  "status": "0",
  "dataScope": "2",
  "remark": "负责设备管理"
}
```

### 2. 为角色添加权限
```bash
POST /api/v1/roles/permission
Authorization: Bearer {token}
Content-Type: application/json

{
  "roleKey": "device_manager",
  "orgId": 1,
  "resource": "device.*",
  "action": "write"
}
```

**说明**：`device.*` 表示设备模块的所有操作权限。

### 3. 为用户分配角色
```bash
POST /api/v1/roles/assign
Authorization: Bearer {token}
Content-Type: application/json

{
  "userId": 1001,
  "roleId": 2,
  "orgId": 1
}
```

### 4. 设置角色继承
```bash
POST /api/v1/roles/inherit
Authorization: Bearer {token}
Content-Type: application/json

{
  "childRoleKey": "device_manager",
  "parentRoleKey": "viewer",
  "orgId": 1
}
```

**说明**：`device_manager` 角色将继承 `viewer` 角色的所有权限。

### 5. 检查用户权限（代码中使用）
```go
// 在控制器或服务中
allowed, err := casbinService.CheckPermission(ctx, userId, orgId, "device.create", "write")
if !allowed {
    return errors.New("无权限")
}
```

---

## 通配符权限

### 支持的通配符模式

| 通配符 | 说明 | 示例 | 匹配范围 |
|--------|------|------|----------|
| `user.*` | 用户模块所有操作 | `user.create`, `user.read`, `user.update`, `user.delete` | 用户模块的增删改查 |
| `*.read` | 所有模块的读操作 | `user.read`, `device.read`, `role.read` | 所有模块的查询权限 |
| `*` | 所有权限 | 任意资源和操作 | 超级管理员权限 |

### 使用示例

#### 1. 为角色添加模块级权限
```bash
# 用户管理员：拥有用户模块所有权限
POST /api/v1/roles/permission
{
  "roleKey": "user_manager",
  "orgId": 1,
  "resource": "user.*",
  "action": "write"
}
```

#### 2. 为角色添加只读权限
```bash
# 访客：拥有所有模块的只读权限
POST /api/v1/roles/permission
{
  "roleKey": "viewer",
  "orgId": 1,
  "resource": "*.read",
  "action": "read"
}
```

#### 3. 为角色添加超级权限
```bash
# 超级管理员：拥有所有权限
POST /api/v1/roles/permission
{
  "roleKey": "admin",
  "orgId": 1,
  "resource": "*",
  "action": "*"
}
```

### 权限匹配规则
Casbin 使用 `keyMatch2` 函数进行通配符匹配：
- `user.create` 匹配 `user.*` ✅
- `device.read` 匹配 `*.read` ✅
- `anything` 匹配 `*` ✅
- `user.create` 不匹配 `device.*` ❌

---

## 角色继承

### 继承规则
1. **多继承**：一个角色可以继承多个父角色
2. **传递性**：A 继承 B，B 继承 C，则 A 拥有 C 的权限
3. **深度限制**：最多支持 3 层继承
4. **循环检测**：自动检测并阻止循环继承

### 继承示例

#### 场景：构建角色层级
```
admin (超级管理员)
  ├── manager (经理)
  │     ├── team_leader (组长)
  │     └── viewer (访客)
  └── viewer (访客)
```

#### 实现步骤
```bash
# 1. 创建角色
POST /api/v1/roles
{ "roleKey": "manager", "roleName": "经理" }

POST /api/v1/roles
{ "roleKey": "team_leader", "roleName": "组长" }

# 2. 设置继承关系
POST /api/v1/roles/inherit
{ "childRoleKey": "manager", "parentRoleKey": "viewer", "orgId": 1 }

POST /api/v1/roles/inherit
{ "childRoleKey": "team_leader", "parentRoleKey": "manager", "orgId": 1 }

# 3. 为父角色添加权限
POST /api/v1/roles/permission
{ "roleKey": "viewer", "orgId": 1, "resource": "*.read", "action": "read" }

POST /api/v1/roles/permission
{ "roleKey": "manager", "orgId": 1, "resource": "user.*", "action": "write" }
```

**结果**：
- `viewer`：拥有 `*.read` 权限
- `manager`：拥有 `*.read` + `user.*` 权限
- `team_leader`：拥有 `*.read` + `user.*` 权限（继承自 manager）

---

## 最佳实践

### 1. 权限设计原则
- **最小权限原则**：只授予必要的权限
- **模块化设计**：按业务模块划分权限（user、device、role 等）
- **使用通配符**：减少权限配置数量
- **角色分层**：基础角色 → 业务角色 → 管理角色

### 2. 推荐的权限命名规范
```
模块.操作
例如：
- user.create   # 创建用户
- user.read     # 查询用户
- user.update   # 更新用户
- user.delete   # 删除用户
- user.*        # 用户模块所有权限
```

### 3. 推荐的角色设计
```
1. admin          - 超级管理员（*）
2. org_admin      - 组织管理员（org.*）
3. user_manager   - 用户管理员（user.*）
4. device_manager - 设备管理员（device.*）
5. viewer         - 访客（*.read）
```

### 4. 中间件使用示例
```go
// 单个权限检查
router.GET("/users", 
    middleware.Auth(tokenManager, cfg),
    middleware.Permission(casbinService, "user.read", "read"),
    userController.ListUsers)

// 任意权限检查（满足其中一个即可）
router.GET("/dashboard", 
    middleware.Auth(tokenManager, cfg),
    middleware.PermissionAny(casbinService, []string{"user.read", "device.read"}, "read"),
    dashboardController.GetDashboard)

// 所有权限检查（必须全部满足）
router.POST("/critical-operation", 
    middleware.Auth(tokenManager, cfg),
    middleware.PermissionAll(casbinService, []string{"user.delete", "device.delete"}, "write"),
    criticalController.Execute)
```

### 5. 组织隔离
- 用户在不同组织可以有不同角色
- 权限检查时必须指定组织ID
- 通过请求头 `X-Org-Id` 或查询参数 `orgId` 传递组织ID

### 6. admin 角色特权
- admin 角色在 Casbin 模型中配置了特殊规则
- 无需为 admin 角色添加任何权限，自动拥有所有权限
- 适用于系统超级管理员

---

## 常见问题

### Q1: 如何让某个用户拥有所有权限？
**A**: 为用户分配 `admin` 角色即可。

### Q2: 如何实现"部门管理员"只能管理本部门用户？
**A**: 使用数据权限（`data_scope`）配合业务逻辑实现，Casbin 负责功能权限，数据权限在业务层处理。

### Q3: 权限修改后需要重启服务吗？
**A**: 不需要。Casbin 会自动同步到数据库，实时生效。

### Q4: 如何调试权限问题？
**A**: 
1. 开启 Casbin 日志（开发环境默认开启）
2. 查看 `casbin_rule` 表的策略数据
3. 使用 `GetRolesForUser` 和 `GetPermissionsForRole` API 查看用户的角色和权限

### Q5: 如何批量导入权限？
**A**: 直接操作 `casbin_rule` 表，然后调用 `ReloadPolicy` API 重新加载策略。

---

## 相关文件

- **Casbin 模型配置**：`cmd/api/casbin_model.conf`
- **数据库初始化脚本**：`scripts/sql/002_init_rbac_schema.sql` 和 `scripts/sql/003_init_rbac_data.sql`
- **权限中间件**：`internal/middleware/permission.go`
- **Casbin 服务**：`internal/service/casbin_service.go`
- **角色服务**：`internal/service/role_service.go`
- **角色控制器**：`internal/controller/role.go`
- **路由配置**：`internal/router/role.go`

---

## 总结

本系统提供了完整的 RBAC 权限管理方案，支持：
- ✅ 通配符权限（`user.*`、`*.read`、`*`）
- ✅ 角色继承（最多 3 层）
- ✅ 多租户隔离
- ✅ admin 超级管理员
- ✅ 细粒度权限控制

通过合理的权限设计和角色规划，可以满足各种复杂的业务场景。
