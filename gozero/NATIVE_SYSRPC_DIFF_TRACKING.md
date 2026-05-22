# native vs go-zero sys-rpc 层差异跟踪文档

> 本文档记录 go-zero sys-rpc 层与 native service 层之间的核心业务逻辑差异。
> sys-rpc 是 go-zero 微服务架构的核心，承载所有业务逻辑。sys-api 只做路由转发。

---

## 元信息

- **创建日期**: 2026-05-22
- **native 基线参考**: `native/internal/service/` + `native/internal/domain/model/`
- **gozero 目标目录**: `gozero/application/sys-rpc/internal/logic/sysservice/`
- **状态**: 差异分析完成，11/23 P0 已修复，部分 P1 已修复

---

## 修正优先级说明

| 标记 | 含义 |
|------|------|
| 🔴 P0 | 致命：业务逻辑完全错误或缺失，影响核心功能正确性 |
| 🟡 P1 | 重要：业务行为不一致，可能导致数据或功能偏差 |
| 🟢 P2 | 一般：排序/类型/字段差异，影响较小 |
| ⚪ P3 | 架构差异：框架选择导致的实现差异，不影响接口契约 |

---

## 1. Auth 模块

### 🔴 P0-1: 仅支持 password 授权类型，缺失其他 4 种

**文件**: `sysservice/authLoginLogic.go`, `common_auth_menu.go`

**问题**: native 支持 5 种授权类型：`password`、`xcx`(小程序)、`wechat`(微信公众号)、`email`(邮箱验证码)、`sms`(短信验证码)。go-zero sys-rpc 只实现了 `password` 一种，其他所有类型均缺失。

**native 参考**: `native/internal/service/auth.go` (LoginByPassword, LoginByXCX, LoginByWechat, LoginBySMS, LoginByEmail)

**修复**: 实现缺失的 4 种授权类型逻辑。

**状态**: ⬜ 未修复


### 🔴 P0-2: 登录认证使用 clientKey+clientSecret，native 使用 clientId

**文件**: `sysservice/authLoginLogic.go` (DB 查询)

**问题**: sys-rpc 的 DB 查询需要 `client_key` 和 `client_secret` 两个字段来认证客户端。native 只需要 `clientId`（查找 `s_auth_client` 表中 `client_id = ?` 的记录）。这导致：
- sys-api 只传递了 `clientId`（没有 `clientSecret`），可能无法匹配到任何记录
- 认证字段语义不匹配（native 的 `s_auth_client` 表存储 `client_id`，不需要 secret）

**修复**: 修改 DB 查询为 `WHERE client_id = $1`，只需 clientId。

**状态**: ⬜ 未修复


### 🔴 P0-3: 缺失登录日志记录

**文件**: `sysservice/authLoginLogic.go`

**问题**: native 在每次登录时记录到 `s_login_log` 表（成功/失败状态、IP 地址、消息等）。go-zero sys-rpc 没有任何登录日志记录。

**native 参考**: `native/internal/service/login_log.go` 的 `RecordLogin` 方法

**修复**: 在 authLoginLogic 中添加登录日志记录。

**状态**: ⬜ 未修复


### 🟡 P1-1: 缺失暴力破解防护

**问题**: native 有基于 Redis 的登录失败次数统计和锁定机制。go-zero sys-rpc 完全缺失。

**修复**: 添加登录失败计数和临时锁定逻辑。

**状态**: ⬜ 未修复


### 🟡 P1-2: 缺失并发登录管理

**问题**: native 检查同一用户是否已有有效登录，可配置是否允许并发。go-zero 没有此逻辑。

**修复**: 按需实现并发登录控制。

**状态**: ⬜ 未修复


### 🟡 P1-3: JWT claims 不同

**问题**: go-zero 的 JWT token 中包含 `OrgID`、`Roles`、`Permissions` 字段。native 的 JWT 不包含这些。导致：
- 不同后端发出的 token 大小不同
- 一方的 token 不能被另一方解析

**修复**: 统一 JWT claims 结构（建议以 native 为准，从 claims 中移除额外字段）。

**状态**: ⬜ 未修复


### 🟡 P1-4: 登录响应缺失字段

**问题**: native 的登录响应包含 `client_id`、`scope`、`openId`(顶层) 和 `openId`/`unionId`(user_info中)。go-zero 缺失这些。

**修复**: 在登录响应中补充缺失字段。

**状态**: ⬜ 未修复


---

## 2. User 模块

### 🔴 P0-4: 删除方式不一致：硬删除 vs 软删除

**文件**: `sysservice/userDeleteLogic.go`

**问题**: native 使用 **硬删除**（`DELETE FROM s_user WHERE id = ?`）。go-zero sys-rpc 使用 **软删除**（`UPDATE s_user SET deleted_at = now()`）。数据可靠性存在根本差异。

**修复**: 改为硬删除与 native 一致。

**状态**: ⬜ 未修复


### 🔴 P0-5: 批量导入容错机制缺失

**文件**: `sysservice/userImportLogic.go`

**问题**: native 导入时逐条处理，即使部分失败也继续，返回 `{successCount, failCount, errors[]}`。go-zero 遇到第一条错误就停止，且只返回简单的错误信息。

**修复**: 改为逐条容错处理，返回结构化统计数据。

**状态**: ⬜ 未修复


### 🟡 P1-5: 用户分页存在 N+1 查询

**文件**: `sysservice/userPageLogic.go`, `common_user.go`

**问题**: go-zero 的 `UserPage` 先查询 ID 列表，再逐条 `GetUserByID` 获取详情（N+1 模式）。native 使用一次 JOIN 查询获取所有用户。性能差距巨大。

**修复**: 改为单次 JOIN 查询，一次性返回所有字段。

**状态**: ⬜ 未修复


### 🟡 P1-6: 修改密码用户 ID 来源不同

**文件**: `sysservice/changePasswordLogic.go`

**问题**: native 只能修改**当前登录用户**的密码（从 JWT 中获取 userId）。go-zero 接受请求中的任意 `userId`，可能导致越权修改他人密码。

**修复**: 从 JWT context 获取当前用户 ID，而非信任请求参数。

**状态**: ⬜ 未修复


### 🟡 P1-7: 排序顺序不同

**问题**: native 用户列表排序为 `sort ASC, created_time DESC`。go-zero 排序为 `id DESC`。

**修复**: 改为 `sort ASC, created_time DESC`。

**状态**: ⬜ 未修复


### 🟢 P2-1: create_by/update_by NULL vs 0

**问题**: native 使用 `0`（零值）表示未设置。go-zero 使用 SQL `NULL` (`nullif(?, 0)` 转换)。

**修复**: 统一为 0（移除 nullif 转换）。

**状态**: ⬜ 未修复


---

## 3. Role/Permission 模块

### 🔴 P0-6: 权限存储后端完全不同：Casbin vs Redis

**文件**: `sysservice/common_api_permission.go`, `common_role_org.go`

**问题**: native 使用 Casbin enforcer + DB adapter 存储权限策略。go-zero 使用 Redis SETs 存储。这是根本性架构差异：
- Casbin 支持复杂的 RBAC 策略评估（role 继承、资源层级、条件匹配）
- Redis SETs 是简单的集合，无法表达复杂策略
- 两个系统对同一数据库表的理解不同

**影响**: 虽然 HTTP 接口相同，但权限执行的正确性可能不同。

**修复**: 这个差异是架构选择的产物。需要确保 Redis 存储方式能正确表达 Casbin 的所有策略类型，或者评估迁移到 Casbin。

**状态**: ⬜ 未修复（架构评估中）


### 🔴 P0-7: API 权限更新/删除时未同步 Redis

**文件**: `sysservice/apiPermissionUpdateLogic.go`, `apiPermissionDeleteLogic.go`

**问题**: go-zero 在更新或删除 API 权限节点时，只更新数据库表，**不会同步更新 Redis 权限缓存**。这意味着权限变更后，Redis 中的数据是过期的，用户可能保持旧的权限。

**native 参考**: native 使用 Casbin DB adapter，权限策略直接存储在数据库，Casbin enforcer 自动加载。

**修复**: 在 API 权限 CRUD 操作后，同步更新 Redis 中的权限数据。

**状态**: ⬜ 未修复


### 🔴 P0-8: 角色分配菜单时未联动更新 API 权限

**文件**: `sysservice/roleAssignMenusLogic.go`

**问题**: native 的 `AssignMenus` 在更新角色菜单关联后，会**同步重建该角色的 Casbin 策略**（通过 `m_role_menu` 中的 `menu_id` 关联到 `s_api_permission` 的 `api_path`）。go-zero 只更新 `m_role_menu` 表，不更新 Redis 权限缓存。

**修复**: 在 `AssignMenus` 后同步重建该角色的 Redis 权限数据。

**状态**: ⬜ 未修复


### 🔴 P0-9: 用户-角色分配时未更新权限缓存

**文件**: `sysservice/roleAssignUsersLogic.go`

**问题**: native 的 `AssignRoleUsers` 在更新用户角色关联后，会同步调用 `SyncPermissionsForUser` 重新计算用户权限。go-zero 只更新 `m_user_role` 表，不更新 Redis。

**修复**: 在 `AssignRoleUsers` 后同步更新该用户的权限缓存。

**状态**: ⬜ 未修复


### 🟡 P1-8: Role Detail 缺少 update_by/updated_time

**问题**: native 返回角色的完整信息（包括 createBy/updateBy/createdTime/updatedTime）。go-zero 的 role detail 缺少 `update_by` 和 `updated_time`。

**修复**: 在 role 查询和响应中补充缺失字段。

**状态**: ⬜ 未修复


### 🟡 P1-9: Role Create 缺少 create_by

**问题**: go-zero 的 role INSERT 语句不包含 `create_by` 字段。

**修复**: 在 INSERT 中添加 `create_by`。

**状态**: ⬜ 未修复


---

## 4. Menu/Org 模块

### 🔴 P0-10: 用户菜单树缺失祖先节点补充

**文件**: `sysservice/common_auth_menu.go` (`getUserMenus` 函数)

**问题**: native 的菜单树构建包含 `withAncestorMenus()` 步骤，递归加载所有祖先节点（parent、grandparent等），确保即使用户只有深层权限，菜单树的结构也是完整的。go-zero **完全没有此逻辑**，只返回用户直接有权访问的菜单。

**影响**: 如果用户角色只分配了深层菜单（如按钮），go-zero 的菜单树中该按钮将没有父节点（目录/菜单），前端树形结构会断裂。

**修复**: 实现 `withAncestorMenus` 等价的祖先节点补全逻辑。

**状态**: ⬜ 未修复


### 🔴 P0-11: 组织更新不能将 status 从 1 改回 0

**文件**: native `service/org.go` (UpdateOrg)

**问题**: native 的更新逻辑中有 `if req.Status != 0` 条件 —— 当 status 为 0 时不会更新（即无法将禁用的组织重新启用）。go-zero 无此限制。**这是 native 的 bug**，应保留 go-zero 的行为。

**修复**: 无需修改（go-zero 的行为正确）。

**状态**: ✅ 无需修改（native 存在此 bug）


### 🔴 P0-12: 组织分页 orgCode 匹配方式不同

**文件**: `sysservice/orgPageLogic.go`

**问题**: native 的 `PageOrgs` 对 `orgCode` 使用**精确匹配**（`WHERE org_code = ?`）。go-zero 使用**模糊匹配**（`WHERE org_code LIKE %?%`）。两种行为完全不同。

**修复**: 改为精确匹配与 native 一致。

**状态**: ⬜ 未修复


### 🟡 P1-10: 组织分页不支持 parentId=0 筛选

**问题**: native 可通过 `*int64` 指针支持 `parentId=0`（查询根组织）。go-zero 的 `int64` 参数通过 `if in.ParentId > 0` 判断，无法筛选 root 组织。

**修复**: 修改条件支持 parentId=0 的场景。

**状态**: ⬜ 未修复


### 🟡 P1-11: 排序 key 不同

**问题**: native 菜单/组织排序为 `sort ASC, created_time DESC`。go-zero 为 `sort ASC, id ASC`。

**修复**: 改为 `sort ASC, created_time DESC`。

**状态**: ⬜ 未修复


### 🟢 P2-2: 组织新建 create_by/update_by 用 NULL 而非 0

**问题**: native 插入 0，go-zero 插入 NULL。

**修复**: 与 User 模块一致处理。

**状态**: ⬜ 未修复（同 P2-1）


---

## 5. Dict/Config 模块

### 🔴 P0-13: 字典数据过滤：status=0 vs deleted_at is null

**文件**: `sysservice/dictTypeLogic.go`, `dictLabelLogic.go`

**问题**: native 的字典查询使用 `WHERE status = 0`（只返回启用的字典）。go-zero 使用 `WHERE deleted_at is null`（返回所有未硬删除的字典，忽略 status）。这意味着：
- 被禁用的字典（status=1）在 native 中不可见，在 go-zero 中可见
- 字典标签查询返回的可能是已禁用的数据

**修复**: 将过滤条件从 `deleted_at is null` 改为 `status = 0`。

**状态**: ⬜ 未修复


### 🔴 P0-14: 字典分页未过滤 parentId

**文件**: `sysservice/dictPageLogic.go`

**问题**: native 的 `PageDict` 过滤 `parent_id = 0 OR parent_id IS NULL`（只展示根级字典）。go-zero 返回所有字典，忽略层级。

**修复**: 添加 parent_id=0 的过滤条件。

**状态**: ⬜ 未修复


### 🔴 P0-15: Config 唯一性校验：name vs code

**文件**: `sysservice/configCreateLogic.go`, `configUpdateLogic.go`

**问题**: native 检查**名称唯一性**（`CheckNameExists`），go-zero 检查**编码唯一性**（`configCodeExists`）。这是完全不同的业务逻辑。

**修复**: 改为 name 唯一性检查，或同时检查 name 和 code。

**状态**: ⬜ 未修复


### 🔴 P0-16: Config GetByCode 排序相反

**文件**: `sysservice/configCodeLogic.go`

**问题**: native 的 `FindByCode` 排序 `ORDER BY id ASC`。go-zero 排序 `ORDER BY id DESC`，完全相反。

**修复**: 改为 `ORDER BY id ASC`。

**状态**: ⬜ 未修复


### 🟡 P1-12: dict 删除硬删除 vs 软删除

**问题**: native 使用硬 DELETE。go-zero 使用软 UPDATE `deleted_at = now()`。

**修复**: 与 User 删除一致处理。

**状态**: ⬜ 未修复


---

## 6. LoginLog/OperLog 模块

### 🟡 P1-13: 日志清理返回值缺失

**文件**: `sysservice/loginLogCleanLogic.go`, `operLogCleanLogic.go`

**问题**: native 返回 `{count: <删除数量>, days: <输入天数>}`。go-zero 只返回 `Ack{Msg: "ok"}`。

**修复**: 返回清理数量。

**状态**: ⬜ 未修复


### 🟡 P1-14: 日志清理默认值 vs 拒绝无效值

**问题**: native 拒绝 `days <= 0` 返回错误。go-zero 默认为 30 天。

**修复**: 改为与 native 一致，拒绝非正数。

**状态**: ⬜ 未修复


### 🟡 P1-15: 日志清理时间计算方式不同

**问题**: native 使用 Go `time.Now().AddDate(0, 0, -days)`（day boundary）。go-zero 使用 PG `now() - 'N day'::interval`（exact interval）。清理范围和结果数量不同。

**修复**: 统一使用 Go 时间计算（day boundary）。

**状态**: ⬜ 未修复


### 🟢 P2-3: 页面排序 NULLS LAST

**问题**: go-zero 日志分页使用 `order by login_time desc NULLS LAST`。native 没有 NULLS LAST。

**修复**: 移除 `NULLS LAST`。

**状态**: ⬜ 未修复


---

## 7. Attachment 模块

### 🔴 P0-17: 存储后端架构差异：可插拔 vs 本地文件

**文件**: `sysservice/attachmentUploadFileLogic.go`, `common_attachment.go`

**问题**: native 使用抽象的 `storage.Manager` 支持可插拔存储后端（local/MinIO/S3/OSS）。go-zero 硬编码本地文件系统（`os.WriteFile`/`os.ReadFile`）。
- 架构差异，但接口契约层面应保持一致
- go-zero 的本地文件存储方式只适用于单机部署

**修复**: 评估是否需要实现存储抽象层，或确认当前功能满足需求。

**状态**: ⬜ 未修复（架构评估中）


### 🔴 P0-18: 文件下载：流式 vs 内存全缓冲

**文件**: `sysservice/attachmentDownloadLogic.go`

**问题**: native 返回 `io.ReadCloser`（流式，支持大文件）。go-zero 先 `os.ReadFile` 读入内存，再返回 `[]byte`（大文件可能导致 OOM）。

**修复**: 改为流式传输。

**状态**: ⬜ 未修复


### 🔴 P0-19: URL 生成：签名 URL vs base64 data URL

**文件**: `sysservice/attachmentUrlLogic.go`

**问题**: native 从存储后端生成**真实的签名 URL**（适用于云存储的场景）。go-zero 将文件内容 base64 编码后生成 **data: URL**，完全不适用于大文件，且带宽消耗巨大。

**影响**: 前端获取到 data: URL 后需要解码整个 base64 字符串，体验和资源消耗完全不可比。

**修复**: 返回可访问的 HTTP URL（本地存储时返回文件路径对应的路由 URL）。

**状态**: ⬜ 未修复


### 🔴 P0-20: 文件删除不清除物理文件

**文件**: `sysservice/attachmentDeleteLogic.go`

**问题**: native 的删除同时清理数据库记录和存储文件。go-zero 只软删除数据库记录（设置 `deleted_at`），**不删除磁盘上的实际文件**，造成文件泄露。

**修复**: 添加磁盘文件清理逻辑。

**状态**: ⬜ 未修复


### 🟡 P1-16: 清理过期附件功能缺失

**问题**: native 有 `CleanExpired()` 定期清理过期附件。go-zero 完全缺失。

**修复**: 实现过期附件清理。

**状态**: ⬜ 未修复


---

## 8. Captcha 模块

### 🔴 P0-21: 图形验证码安全性不足

**文件**: `sysservice/common_captcha.go`

**问题**: native 使用 `base64Captcha` 库生成带干扰线/变形/噪点的专业验证码。go-zero 在 SVG 中直接渲染**明文数字**，完全无防 OCR 干扰。安全性等同虚设。

**修复**: 引入 base64Captcha 库或实现等价的安全验证码生成。

**状态**: ⬜ 未修复


### 🔴 P0-22: 短信/邮件验证码不实际发送

**文件**: `sysservice/common_captcha.go`

**问题**: native 的短信/邮件验证码通过真实的第三方服务（SMS/email provider）发送到用户手机/邮箱。go-zero **只生成随机码存储在 Redis，从不发送**。用户永远收不到验证码，基于验证码的登录/验证完全无法工作。

**修复**: 集成真实的 SMS/email 发送服务。

**状态**: ⬜ 未修复


### 🔴 P0-23: 随机数生成算法不安全

**文件**: `sysservice/common_captcha.go` (`randomDigits` 函数)

**问题**: native 使用 `crypto/rand`（密码学安全）。go-zero 使用 `math/rand` 加 `time.Now().UnixNano()` 种子，可预测，不适用于安全场景。

**修复**: 改用 `crypto/rand`。

**状态**: ⬜ 未修复


### 🟡 P1-17: 验证码过期时间硬编码

**问题**: native 支持配置文件自定义过期时间。go-zero 硬编码 5 分钟。

**修复**: 从配置文件读取过期时间。

**状态**: ⬜ 未修复


### 🟡 P1-18: 短信/邮件验证码验证功能缺失

**问题**: go-zero 只实现了图形验证码的验证（`verifyImageCaptcha`），短信和邮件验证码的验证完全缺失。

**修复**: 实现短信和邮件验证码的验证逻辑。

**状态**: ⬜ 未修复


---

## 9. StorageEnv 模块

### 🟡 P1-19: 测试连接为假实现

**文件**: `sysservice/storageEnvTestLogic.go`

**问题**: native 执行实际的存储连接测试（按 storageType 分派）。go-zero 只检查数据库记录是否存在，不执行任何实际连接测试。

**修复**: 实现按 storage type 的真实连接测试。

**状态**: ⬜ 未修复


### 🟢 P2-4: 首个环境不自动设为默认

**问题**: native 创建第一个存储环境时自动设为 default。go-zero 不会。

**修复**: 添加首个环境自动默认的逻辑。

**状态**: ⬜ 未修复


---

## 10. 跨模块通用差异

### 🟡 P1-20: 分页排序字段和方向不支持

**问题**: native 的多个模块支持 `orderByColumn` 和 `isAsc` 动态排序。go-zero 使用固定的排序方式。

**涉及文件**: 所有 `*PageLogic.go` 文件

**修复**: 在分页接口中支持动态排序。

**状态**: ⬜ 未修复


### 🟡 P1-21: 所有 query 查询缺失 deleted_at is null

**问题**: go-zero 在大多数查询中使用 `deleted_at is null`（软删除模式），但 native 使用 `status = 0`（状态模式）。某些模块（如 dict、config、loginlog、operlog）native 不使用软删除。

**修复**: 根据各模块的 native 模式统一：user/role/menu/org 保持软删除，dict/attachment 改回硬删除，config/loginlog/operlog 保持硬删除。

**状态**: ⬜ 未修复


---

## 修复工作清单

### P0（立即修复）

- [ ] P0-1: Auth 支持 4 种缺失的授权类型 (xcx/wechat/email/sms)
- [x] P0-2: Auth ClientId 替代 ClientKey+ClientSecret
- [x] P0-3: Auth 添加登录日志记录
- [x] P0-4: User 删除改为硬删除
- [x] P0-5: User 批量导入改为容错模式
- [ ] P0-6: Role 权限后端统一评估（Casbin vs Redis）
- [ ] P0-7: API 权限 CRUD 后同步 Redis
- [ ] P0-8: 角色分配菜单后同步更新 API 权限
- [ ] P0-9: 用户-角色分配后更新权限缓存
- [ ] P0-10: Menu 用户菜单树补全祖先节点
- [x] P0-11: Org update status 问题（go-zero 正确，无需修改）
- [x] P0-12: Org 分页 orgCode 改为精确匹配
- [x] P0-13: Dict 过滤改为 status=0
- [x] P0-14: Dict 分页过滤 parent_id
- [x] P0-15: Config 唯一性改为 name
- [x] P0-16: Config GetByCode 排序改为 ASC
- [ ] P0-17: Attachment 存储后端（架构评估）
- [ ] P0-18: Attachment 下载改为流式
- [ ] P0-19: Attachment URL 改为 HTTP URL
- [x] P0-20: Attachment 删除物理文件
- [ ] P0-21: Captcha 图像安全性
- [ ] P0-22: Captcha SMS/email 实际发送
- [x] P0-23: Captcha 随机数用 crypto/rand

### P1（尽快修复）

- [ ] P1-1: Auth 暴力破解防护
- [ ] P1-2: Auth 并发登录管理
- [ ] P1-3: Auth JWT claims 统一
- [ ] P1-4: Auth 登录响应补充字段
- [x] P1-5: User 分页 N+1 优化
- [ ] P1-6: User 修改密码 userId 来源
- [x] P1-7: User 排序统一
- [ ] P1-8: Role detail 补充字段
- [ ] P1-9: Role create 补充 create_by
- [ ] P1-10: Org 分页支持 parentId=0
- [ ] P1-11: Menu/Org 排序统一
- [ ] P1-12: Dict 删除改为硬删除
- [x] P1-13: Log 清理返回数量
- [x] P1-14: Log 清理拒绝无效值
- [ ] P1-15: Log 清理时间计算方式
- [ ] P1-16: Attachment 清理过期功能
- [ ] P1-17: Captcha 过期时间可配置
- [ ] P1-18: Captcha SMS/email 验证
- [ ] P1-19: StorageEnv 测试连接实现
- [ ] P1-20: 分页动态排序支持
- [ ] P1-21: 删除策略统一

### P2（后续优化）

- [ ] P2-1: User role org create_by/update_by NULL→0
- [ ] P2-2: 同上
- [ ] P2-3: Log ORDER BY 移除 NULLS LAST
- [ ] P2-4: StorageEnv 首个环境自动 default

---

## 注意事项

1. go-zero 开发模式：先修改 `.proto` 文件 → 运行 protoc 生成代码 → 修改 logic 文件
2. sys-rpc 修改后需要重新编译并重启 sys-rpc 服务
3. 修改完成后需要验证 sys-api → sys-rpc 调用链路是否正常
4. 部分架构差异（Casbin vs Redis、存储抽象层）需要评估是否影响业务功能
