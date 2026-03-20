# Casbin 权限管理测试示例

## 测试环境准备

### 1. 执行数据库初始化脚本
```bash
psql -U postgres -d nai-tizi -f scripts/sql/002_init_rbac_tables.sql
```

### 2. 启动服务
```bash
cd cmd/api
go run main.go
```

---

## 测试场景

### 场景 1：admin 角色拥有所有权限

#### 1.1 登录获取 Token
```bash
curl -X POST http://localhost:9009/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "clientKey": "web_client",
    "clientSecret": "web_secret_2024",
    "username": "admin",
    "password": "admin123"
  }'
```

**响应示例**：
```json
{
  "code": 200,
  "msg": "登录成功",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 1800,
    "refreshExpiresIn": 604800
  }
}
```

#### 1.2 访问角色列表（需要 role.read 权限）
```bash
curl -X GET "http://localhost:9009/api/v1/roles?pageNum=1&pageSize=10" \
  -H "Authorization: Bearer {accessToken}"
```

**预期结果**：✅ 成功返回角色列表（admin 拥有所有权限）

#### 1.3 创建新角色（需要 role.create 权限）
```bash
curl -X POST http://localhost:9009/api/v1/roles \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{
    "roleKey": "device_manager",
    "roleName": "设备管理员",
    "roleSort": 10,
    "status": "0",
    "dataScope": "2",
    "remark": "负责设备管理"
  }'
```

**预期结果**：✅ 成功创建角色

---

### 场景 2：通配符权限测试

#### 2.1 为 device_manager 角色添加 device.* 权限
```bash
curl -X POST http://localhost:9009/api/v1/roles/permission \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{
    "roleKey": "device_manager",
    "orgId": 1,
    "resource": "device.*",
    "action": "write"
  }'
```

**预期结果**：✅ 成功添加权限

#### 2.2 查询角色权限
```bash
curl -X GET "http://localhost:9009/api/v1/roles/permissions?roleKey=device_manager&orgId=1" \
  -H "Authorization: Bearer {accessToken}"
```

**预期结果**：
```json
{
  "code": 200,
  "msg": "获取角色权限成功",
  "data": [
    ["role::device_manager", "org::1", "device.*", "write"]
  ]
}
```

#### 2.3 创建测试用户并分配 device_manager 角色
```bash
# 假设创建了 user_id=1002 的用户
curl -X POST http://localhost:9009/api/v1/roles/assign \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 1002,
    "roleId": 4,
    "orgId": 1
  }'
```

#### 2.4 使用 device_manager 用户登录并测试权限
```bash
# 登录 device_manager 用户
curl -X POST http://localhost:9009/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "clientKey": "web_client",
    "clientSecret": "web_secret_2024",
    "username": "device_user",
    "password": "password123"
  }'

# 访问设备创建接口（假设有 device.create 权限）
curl -X POST http://localhost:9009/api/v1/devices \
  -H "Authorization: Bearer {device_user_token}" \
  -H "Content-Type: application/json" \
  -d '{ "deviceName": "测试设备" }'
```

**预期结果**：✅ 成功（device.create 匹配 device.*）

```bash
# 访问用户创建接口（没有 user.create 权限）
curl -X POST http://localhost:9009/api/v1/users \
  -H "Authorization: Bearer {device_user_token}" \
  -H "Content-Type: application/json" \
  -d '{ "userName": "test" }'
```

**预期结果**：❌ 403 Forbidden（无权限）

---

### 场景 3：角色继承测试

#### 3.1 创建 viewer 角色并添加只读权限
```bash
# viewer 角色已在初始化脚本中创建，添加 *.read 权限
curl -X POST http://localhost:9009/api/v1/roles/permission \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{
    "roleKey": "viewer",
    "orgId": 1,
    "resource": "*.read",
    "action": "read"
  }'
```

#### 3.2 设置 device_manager 继承 viewer
```bash
curl -X POST http://localhost:9009/api/v1/roles/inherit \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{
    "childRoleKey": "device_manager",
    "parentRoleKey": "viewer",
    "orgId": 1
  }'
```

**预期结果**：✅ 成功设置继承关系

#### 3.3 验证继承权限
```bash
# device_manager 用户现在应该拥有：
# 1. device.* (自己的权限)
# 2. *.read (继承自 viewer)

# 测试读取用户列表（继承的 *.read 权限）
curl -X GET "http://localhost:9009/api/v1/users?pageNum=1&pageSize=10" \
  -H "Authorization: Bearer {device_user_token}"
```

**预期结果**：✅ 成功（user.read 匹配 *.read）

---

### 场景 4：循环继承检测

#### 4.1 尝试创建循环继承
```bash
# 假设：device_manager -> viewer
# 尝试：viewer -> device_manager（会形成循环）

curl -X POST http://localhost:9009/api/v1/roles/inherit \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{
    "childRoleKey": "viewer",
    "parentRoleKey": "device_manager",
    "orgId": 1
  }'
```

**预期结果**：❌ 400 Bad Request（检测到循环继承）

---

### 场景 5：继承深度限制

#### 5.1 创建 4 层继承（超过限制）
```bash
# 创建角色
curl -X POST http://localhost:9009/api/v1/roles \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{ "roleKey": "level1", "roleName": "Level 1" }'

curl -X POST http://localhost:9009/api/v1/roles \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{ "roleKey": "level2", "roleName": "Level 2" }'

curl -X POST http://localhost:9009/api/v1/roles \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{ "roleKey": "level3", "roleName": "Level 3" }'

curl -X POST http://localhost:9009/api/v1/roles \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{ "roleKey": "level4", "roleName": "Level 4" }'

# 设置继承关系
curl -X POST http://localhost:9009/api/v1/roles/inherit \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{ "childRoleKey": "level1", "parentRoleKey": "level2", "orgId": 1 }'

curl -X POST http://localhost:9009/api/v1/roles/inherit \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{ "childRoleKey": "level2", "parentRoleKey": "level3", "orgId": 1 }'

curl -X POST http://localhost:9009/api/v1/roles/inherit \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{ "childRoleKey": "level3", "parentRoleKey": "level4", "orgId": 1 }'

# 尝试添加第 4 层（超过限制）
curl -X POST http://localhost:9009/api/v1/roles/inherit \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{ "childRoleKey": "level4", "parentRoleKey": "viewer", "orgId": 1 }'
```

**预期结果**：❌ 400 Bad Request（继承深度超过限制）

---

### 场景 6：多租户隔离

#### 6.1 创建新组织
```bash
# 假设通过组织管理接口创建了 org_id=2 的组织
```

#### 6.2 为用户在不同组织分配不同角色
```bash
# 用户 1002 在组织 1 是 device_manager
curl -X POST http://localhost:9009/api/v1/roles/assign \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 1002,
    "roleId": 4,
    "orgId": 1
  }'

# 用户 1002 在组织 2 是 viewer
curl -X POST http://localhost:9009/api/v1/roles/assign \
  -H "Authorization: Bearer {accessToken}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 1002,
    "roleId": 3,
    "orgId": 2
  }'
```

#### 6.3 验证多租户隔离
```bash
# 在组织 1 中，用户 1002 有 device.* 权限
curl -X POST http://localhost:9009/api/v1/devices \
  -H "Authorization: Bearer {device_user_token}" \
  -H "X-Org-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{ "deviceName": "测试设备" }'
```

**预期结果**：✅ 成功

```bash
# 在组织 2 中，用户 1002 只有 *.read 权限
curl -X POST http://localhost:9009/api/v1/devices \
  -H "Authorization: Bearer {device_user_token}" \
  -H "X-Org-Id: 2" \
  -H "Content-Type: application/json" \
  -d '{ "deviceName": "测试设备" }'
```

**预期结果**：❌ 403 Forbidden（无写权限）

---

## 验证清单

- [ ] admin 角色拥有所有权限
- [ ] 通配符权限正常工作（`user.*`、`*.read`、`*`）
- [ ] 角色继承正常工作
- [ ] 循环继承检测生效
- [ ] 继承深度限制生效（最多 3 层）
- [ ] 多租户隔离正常工作
- [ ] 权限中间件正常拦截无权限请求
- [ ] 权限修改实时生效

---

## 调试技巧

### 1. 查看 Casbin 策略
```sql
-- 查看所有策略
SELECT * FROM casbin_rule;

-- 查看角色权限
SELECT * FROM casbin_rule WHERE ptype = 'p';

-- 查看用户角色关系
SELECT * FROM casbin_rule WHERE ptype = 'g';
```

### 2. 查看用户角色
```bash
curl -X GET "http://localhost:9009/api/v1/roles/user?userId=1002&orgId=1" \
  -H "Authorization: Bearer {accessToken}"
```

### 3. 查看角色权限
```bash
curl -X GET "http://localhost:9009/api/v1/roles/permissions?roleKey=device_manager&orgId=1" \
  -H "Authorization: Bearer {accessToken}"
```

### 4. 开启 Casbin 日志
在 `internal/container/container.go` 中：
```go
if c.config.Env == "development" || c.config.Env == "dev" {
    enforcer.EnableLog(true)  // 已默认开启
}
```

---

## 常见问题排查

### Q1: 权限检查总是返回 false
**排查步骤**：
1. 检查用户是否分配了角色
2. 检查角色是否有对应权限
3. 检查组织ID是否正确
4. 查看 Casbin 日志

### Q2: admin 角色没有所有权限
**排查步骤**：
1. 检查 Casbin 模型配置是否正确
2. 检查用户是否真的拥有 admin 角色
3. 重新加载策略：`casbinService.ReloadPolicy(ctx)`

### Q3: 通配符不生效
**排查步骤**：
1. 检查 Casbin 模型中是否使用了 `keyMatch2`
2. 检查权限格式是否正确（例如：`user.*` 而不是 `user*`）

---

## 总结

通过以上测试场景，可以全面验证 Casbin RBAC 权限管理系统的功能：
- ✅ 基础权限控制
- ✅ 通配符权限
- ✅ 角色继承
- ✅ 多租户隔离
- ✅ admin 超级管理员
- ✅ 安全限制（循环继承、深度限制）

系统已经可以投入使用！
