# 简化版 RBAC 方案

## 核心需求

### 用户管理
- ✅ 用户新增时可选择**多个角色**
- ✅ 用户只属于**一个组织**
- ✅ 简单直接，无需复杂的多组织关系

### 角色管理
- ✅ 新建角色时选择**菜单/API权限**
- ✅ 支持**显示顺序**配置
- ✅ 支持**启用/禁用**状态

### 组织管理
- ✅ 支持**显示顺序**配置
- ✅ 树形结构（公司-部门-小组）

---

## 简化的数据模型

### 1. 用户表 (s_user)
```sql
CREATE TABLE s_user (
    user_id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL,                 -- 所属组织（单一）
    user_name VARCHAR(50) NOT NULL UNIQUE,  -- 登录账号
    nick_name VARCHAR(50),                  -- 昵称
    email VARCHAR(100),
    phonenumber VARCHAR(20),
    password VARCHAR(255),
    avatar VARCHAR(500),
    status CHAR(1) DEFAULT '0',             -- 0正常 1停用
    user_type VARCHAR(20) DEFAULT 'system',
    open_id VARCHAR(100),
    union_id VARCHAR(100),
    login_ip VARCHAR(50),
    login_date BIGINT,
    remark VARCHAR(500),
    create_by BIGINT,
    update_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_user_org ON s_user(org_id);
```

**关键点：**
- `org_id` 直接在用户表中，一个用户只属于一个组织
- 移除 `dept_id`，统一使用 `org_id`

### 2. 组织表 (s_org) - 已存在，添加排序
```sql
CREATE TABLE s_org (
    org_id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT DEFAULT 0,
    ancestors VARCHAR(500),
    org_name VARCHAR(100) NOT NULL,
    org_code VARCHAR(100) UNIQUE,
    org_type VARCHAR(20) DEFAULT 'company',
    leader VARCHAR(50),
    phone VARCHAR(20),
    email VARCHAR(100),
    status CHAR(1) DEFAULT '0',             -- 0正常 1停用
    sort_order INT DEFAULT 0,               -- ✅ 显示顺序
    remark VARCHAR(500),
    create_by BIGINT,
    update_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

**关键点：**
- `sort_order` 用于控制显示顺序
- 保持树形结构

### 3. 角色表 (s_role) - 已存在，确认字段
```sql
CREATE TABLE s_role (
    role_id BIGSERIAL PRIMARY KEY,
    role_key VARCHAR(100) NOT NULL UNIQUE,
    role_name VARCHAR(100) NOT NULL,
    sort_order INT DEFAULT 0,               -- ✅ 显示顺序（原 role_sort）
    status CHAR(1) DEFAULT '0',             -- ✅ 0正常 1停用
    data_scope CHAR(1) DEFAULT '1',
    is_system BOOLEAN DEFAULT FALSE,
    remark VARCHAR(500),
    create_by BIGINT,
    update_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

**关键点：**
- `sort_order` 控制显示顺序
- `status` 控制启用/禁用（0正常 1停用）
- `is_system` 标记系统内置角色（不可删除）

### 4. 用户角色关系表 (s_user_role) - 简化版
```sql
CREATE TABLE s_user_role (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    create_by BIGINT,
    update_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)                -- ✅ 一个用户可以有多个角色
);

CREATE INDEX idx_user_role_user ON s_user_role(user_id);
CREATE INDEX idx_user_role_role ON s_user_role(role_id);
```

**关键点：**
- 移除 `org_id`（因为用户已经有固定组织）
- 支持一个用户拥有多个角色
- 简单的多对多关系

### 5. 角色菜单关系表 (s_role_menu) - 已存在
```sql
CREATE TABLE s_role_menu (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL,
    menu_id BIGINT NOT NULL,
    create_by BIGINT,
    update_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, menu_id)
);
```

**关键点：**
- 角色与菜单/API权限的关联
- 新建角色时配置权限

### 6. 菜单/权限表 (s_menu) - 已存在
```sql
CREATE TABLE s_menu (
    menu_id BIGSERIAL PRIMARY KEY,
    menu_name VARCHAR(100) NOT NULL,
    parent_id BIGINT DEFAULT 0,
    sort_order INT DEFAULT 0,               -- ✅ 显示顺序（原 order_num）
    path VARCHAR(200),
    component VARCHAR(200),
    menu_type CHAR(1) NOT NULL,             -- M目录 C菜单 F按钮/API
    visible CHAR(1) DEFAULT '0',
    status CHAR(1) DEFAULT '0',
    perms VARCHAR(200),                     -- 权限标识
    icon VARCHAR(100),
    remark VARCHAR(500),
    create_by BIGINT,
    update_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

---

## 简化的关系图

```
┌─────────────┐
│   s_user    │ 用户
│             │
│ - org_id    │ ← 直接关联组织（一对一）
└──────┬──────┘
       │
       │ 1:N
       ↓
┌─────────────┐
│ s_user_role │ 用户可以有多个角色
└──────┬──────┘
       │
       │ N:1
       ↓
┌─────────────┐
│   s_role    │ 角色
│             │
│ - status    │ ← 启用/禁用
│ - sort_order│ ← 显示顺序
└──────┬──────┘
       │
       │ 1:N
       ↓
┌─────────────┐
│ s_role_menu │ 角色权限
└──────┬──────┘
       │
       │ N:1
       ↓
┌─────────────┐
│   s_menu    │ 菜单/API权限
│             │
│ - sort_order│ ← 显示顺序
└─────────────┘

┌─────────────┐
│   s_user    │
│ - org_id ───┼──→ s_org (组织树)
└─────────────┘
```

---

## 业务场景示例

### 场景1：新增用户
```
1. 选择所属组织：技术部
2. 选择角色：[开发人员, 测试人员]
3. 保存

数据：
s_user:
| user_id | org_id | user_name |
|---------|--------|-----------|
| 1001    | 5      | zhangsan  |

s_user_role:
| user_id | role_id |
|---------|---------|
| 1001    | 10      | ← 开发人员
| 1001    | 11      | ← 测试人员
```

### 场景2：新建角色
```
1. 角色名称：项目经理
2. 角色标识：project_manager
3. 显示顺序：5
4. 状态：启用
5. 选择权限：
   ✓ 项目管理
   ✓ 任务管理
   ✓ 团队管理
6. 保存

数据：
s_role:
| role_id | role_key        | role_name | sort_order | status |
|---------|-----------------|-----------|------------|--------|
| 15      | project_manager | 项目经理   | 5          | 0      |

s_role_menu:
| role_id | menu_id |
|---------|---------|
| 15      | 100     | ← 项目管理
| 15      | 101     | ← 任务管理
| 15      | 102     | ← 团队管理
```

### 场景3：查询用户权限
```sql
-- 查询用户的所有角色
SELECT r.* 
FROM s_role r
INNER JOIN s_user_role ur ON r.role_id = ur.role_id
WHERE ur.user_id = 1001 AND r.status = '0'
ORDER BY r.sort_order ASC;

-- 查询用户的所有权限
SELECT DISTINCT m.*
FROM s_menu m
INNER JOIN s_role_menu rm ON m.menu_id = rm.menu_id
INNER JOIN s_user_role ur ON rm.role_id = ur.role_id
WHERE ur.user_id = 1001 AND m.status = '0'
ORDER BY m.sort_order ASC;
```

### 场景4：禁用角色
```sql
-- 禁用角色
UPDATE s_role SET status = '1' WHERE role_id = 15;

-- 结果：所有拥有该角色的用户将失去该角色的权限
-- 但用户角色关系保留，启用后权限自动恢复
```

---

## 与 Casbin 集成（简化版）

### Casbin 策略
```
# 用户角色关系（无需 org_id）
g, user::1001, role::developer
g, user::1001, role::tester

# 角色权限
p, role::developer, *, code.*, write
p, role::tester, *, test.*, write
```

### 权限检查
```go
// 检查用户权限（无需指定组织）
func CheckPermission(userId int64, resource, action string) (bool, error) {
    sub := fmt.Sprintf("user::%d", userId)
    return enforcer.Enforce(sub, resource, action)
}
```

---

## 核心优势

### 1. 简单直接
- ✅ 用户只属于一个组织，关系清晰
- ✅ 用户可以有多个角色，灵活够用
- ✅ 无需复杂的多组织关系表

### 2. 易于理解
- ✅ 数据模型简单，开发人员容易理解
- ✅ 业务逻辑清晰，维护成本低
- ✅ 用户界面直观，操作简单

### 3. 性能良好
- ✅ 查询简单，无需复杂 JOIN
- ✅ 索引优化容易
- ✅ 缓存策略简单

### 4. 满足需求
- ✅ 支持用户多角色
- ✅ 支持角色权限配置
- ✅ 支持显示顺序
- ✅ 支持启用/禁用

---

## 需要调整的内容

### 1. 表结构调整
- ✅ `s_user` 添加 `org_id` 字段，移除 `dept_id`
- ✅ `s_user_role` 移除 `org_id` 字段
- ✅ `s_role` 确认有 `sort_order` 和 `status` 字段
- ✅ `s_org` 添加 `sort_order` 字段
- ✅ `s_menu` 确认有 `sort_order` 字段

### 2. 所有表统一规范
- ✅ 表名：`sys_*` → `s_*`
- ✅ 模型：`Sys*` → `S*`
- ✅ 时间字段：统一使用 `created_at`, `updated_at`, `deleted_at`
- ✅ 移除 `gorm.Model`，避免 ID 冲突

### 3. 服务层调整
- ✅ 更新所有表名引用
- ✅ 简化权限检查逻辑（无需 org_id）
- ✅ 添加角色启用/禁用逻辑

---

## 总结

这个简化方案：
- **更简单**：用户-组织是一对一关系
- **够灵活**：用户可以有多个角色
- **易维护**：数据模型清晰，逻辑简单
- **满足需求**：完全符合您提出的核心功能

**是否符合您的预期？如果确认，我将开始执行具体的代码调整。**
