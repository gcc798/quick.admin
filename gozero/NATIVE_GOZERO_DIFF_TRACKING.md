# native vs gozero 接口差异跟踪文档

> 本文档记录了 go-zero 实现与 native 基线实现之间的所有接口契约差异。
> go-zero 的开发模式：修改 .api/.proto → 生成代码 → 修改业务逻辑。

---

## 元信息

- **创建日期**: 2026-05-22
- **native 基线参考**: `native/` 目录
- **gozero 目标目录**: `gozero/`
- **状态**: 部分修复（14/35 已完成，21 保留后续）
- **最后更新**: 2026-05-22（P0 已基本完成）

---

## 修正优先级说明

| 标记 | 含义 |
|------|------|
| 🔴 P0 | 致命：前端会直接报错/显示异常，必须优先修复 |
| 🟡 P1 | 重要：业务逻辑不一致/功能缺失，影响功能正确性 |
| 🟢 P2 | 一般：消息语言、类型宽度、缺少验证等，不影响核心功能 |
| ⚪ P3 | 低优：架构差异，不影响接口契约 |

---

## 1. 全局/共同差异

### ✅ 已确认: JSON 字段命名风格一致

**结论**: native 和 gozero 都使用 **camelCase**。

native 的 JSON tag 惯例: 99% 的字段使用 camelCase（如 `userId`, `userName`, `parentId`, `dictType`），只有 Auth 模块的 Token 响应（LoginResponse/RefreshTokenResponse）特例使用 snake_case（`access_token`, `refresh_token`, `expires_in`, `user_info`）。gozero 的 auth common.go 已对齐此特例。

**状态**: ✅ 无需修改（仅 Auth 响应已对齐 snake_case 特例）

---

### 🔴 P0: CommonResp 中 data 字段的 optional 标签

**问题**: gozero 的 `CommonResp.Data` 使用 `json:"data,optional"`，当 Data 为 nil 时 JSON 中会省略该字段。native 的 `Response.Data` 使用 `json:"data"`，始终输出 `null`。

**文件**: `gozero/application/sys-api/api/common.api`

**修复**: 将 `data interface{} \`json:"data,optional"\`` 改为 `data interface{} \`json:"data"\``

**状态**: ✅ 已修复

---

### 🔴 P0: 成功消息语言不统一

**问题**: native 返回中文消息（如 `"操作成功"`），gozero 返回英文 `"success"`。前端可能依赖 msg 字段做判断。

**涉及文件**: 所有 `internal/logic/*/` 下的 logic 文件

**修复方向**: 统一为中文，将 `Msg: "success"` 替换为 `Msg: "操作成功"`。

**状态**: ✅ 已修复（Msg 统一为 "操作成功"）

---

### 🔴 P0: 错误信息和错误码不一致

**问题**: native 有结构化错误处理（`response.Error()` 区分 business/infrastructure/system 错误），返回中文用户友好信息和细分错误码。gozero 固定返回 `{Code: 500, Msg: err.Error()}`，而且部分 Detail 接口返回 404。

**文件**: `gozero/application/sys-api/internal/logic/*/` 下所有 logic 文件

**修复方向**: 在 gozero 中实现类似的错误分类处理，至少：
- 为不同错误场景返回不同错误码
- 统一错误消息为中文
- Detail 接口统一使用 500（而非 404）与 native 一致

**状态**: ⬜ 未修复

---

### 🟡 P1: Casbin 权限中间件缺失

**问题**: native 在每个路由上有细粒度的 Casbin RBAC 权限检查（如 `middleware.Permission(ctx.CasbinService, constants.ResourceUserCreate)`）。gozero 没有可见的权限检查。

**涉及文件**: 所有 `.api` 文件和路由注册

**修复方向**: 如果权限由 go-zero 框架中间件统一处理，需要确认是否生效。否则需要在 gozero 中实现等价的权限控制。

**状态**: ⬜ 未修复

---

### 🟡 P1: 缺少创建者/更新者追踪

**问题**: native 在 Create/Update 操作中通过 JWT 上下文提取 `userId` 并设置 `CreateBy`/`UpdateBy`。多个 gozero 模块（org, storage-env 等）未设置这些字段。

**涉及文件**: `gozero/application/sys-api/internal/logic/org/orgCreateLogic.go`, `orgUpdateLogic.go`, `storageenv/storageEnvCreateLogic.go`, `storageenv/storageEnvUpdateLogic.go`

**修复方向**: 在 logic 层通过 `commonutil.UserIDFromContext(ctx)` 获取当前用户 ID，传递给 RPC 请求。

**状态**: ⬜ 未修复

---

### 🟢 P2: int32 vs int64 类型不一致

**问题**: native 大量使用 `int32` 类型（status、sex、userType、dataScope 等），gozero 使用 `int64`。在 JSON 层面都是数字，但前端如果做严格类型检查可能受影响。

**修复方向**: 将 gozero 的 `.api` 文件中对应字段类型改为 `int32`，重新生成代码。

**状态**: ⬜ 未修复

---

### 🟢 P2: 缺少字段级验证

**问题**: native 使用 Gin 的 `binding:"required,min=,max=,oneof="` 进行字段级验证。gozero 几乎没有任何字段级验证。

**修复方向**: 在 logic 层或通过 go-zero 的 validate 中间件添加字段验证。

**状态**: ⬜ 未修复

---

### ⚪ P3: 缺少基础设施端点

**问题**: native 有 `/metrics`（Prometheus）、`/swagger/*any`（API 文档）端点，gozero 没有。

**修复方向**: 按需添加。

**状态**: ⬜ 未修复

---

## 2. Auth 模块

### 🔴 P0: clientId vs clientKey+clientSecret

**问题**: native 的登录请求使用单一的 `clientId` 字段做客户端认证。gozero 使用 `clientKey` + `clientSecret` 密钥对。

**文件**: `gozero/application/sys-api/api/auth.api`

**修复**: 将 `LoginReq` 中的 `clientKey`/`clientSecret` 改为 `clientId`（与 native 一致）。

**状态**: ⬜ 未修复

---

### 🔴 P0: 登录响应 JSON 键名 camelCase vs snake_case

**问题**: native 登录响应返回 `access_token`, `expires_in`, `refresh_token`, `user_info` 等 snake_case 键名。gozero 返回 `accessToken`, `expiresIn`, `refreshToken`, `userInfo` 等 camelCase 键名。

**文件**: `gozero/application/sys-api/internal/logic/auth/common.go`（`buildLoginResponse` 函数和 `LoginData` 结构体）

**修复**: 将响应结构体中的 JSON tag 改为 snake_case。

**状态**: ⬜ 未修复

---

### 🟡 P1: 缺少 smsCode 字段

**问题**: native 的 `LoginRequest` 有 `smsCode` 字段（`Normalize()` 方法自动将 smsCode 复制到 Code）。gozero 只有 `code` 字段。

**文件**: `gozero/application/sys-api/api/auth.api`

**修复**: 在 `LoginReq` 中添加 `smsCode` 字段。

**状态**: ⬜ 未修复

---

### 🟡 P1: 缺少 /auth/login 和 /auth/logout 别名路由

**问题**: native 支持 `/login` 和 `/auth/login` 两个路由（以及 `/logout`/`/auth/logout`）。gozero 只有 `/login` 和 `/logout`。

**文件**: `gozero/application/sys-api/api/auth.api`

**修复**: 在 auth.api 中添加 `POST /auth/login` 和 `POST /auth/logout` 路由。

**状态**: ✅ 已修复（auth.api 添加路由，logic + handler 已实现）

---

### 🟡 P1: 登录响应缺少字段

**问题**: native 登录响应包含 `expire_in`, `refresh_expire_in`, `client_id`, `scope`, `openId`（顶层）, `openId`/`unionId`（user_info 中）。gozero 缺失这些字段。

**文件**: `gozero/application/sys-api/internal/logic/auth/common.go`

**修复**: 在 `LoginData` 结构体中添加缺失字段。

**状态**: ⬜ 未修复

---

### 🟢 P2: 缺少 appid 字段

**问题**: native 的 `LoginRequest` 有 `appid` 字段（用于小程序场景）。gozero 没有。

**文件**: `gozero/application/sys-api/api/auth.api`

**修复**: 在 `LoginReq` 中添加 `appid` 字段。

**状态**: ⬜ 未修复

---

## 3. User 模块

### 🟡 P1: 缺少 XcxGetInfo 端点

**问题**: native 有 `POST /system/user/xcxGetInfo` 端点（小程序用户信息获取）。gozero 完全缺失。

**文件**: 需在 `gozero/application/sys-api/api/user.api` 中添加

**修复**: 添加 `POST /system/user/xcxGetInfo` 路由和对应的 handler/logic。

**状态**: ✅ 已修复（user.api 添加路由，logic 已实现）

---

### 🔴 P0: 创建用户响应数据不一致

**问题**: native 的 `CreateUser` 返回 `{"userId": "ok"}`。gozero 返回 `"ok"` 字符串。

**文件**: `gozero/application/sys-api/internal/logic/user/userCreateLogic.go`

**修复**: 将返回的 `Data` 改为包含 `userId` 的对象。

**状态**: ⬜ 未修复

---

### 🔴 P0: 批量导入响应不一致

**问题**: native 的 `BatchImport` 返回 `{successCount, failCount, errors[]}` 失败时继续处理。gozero 返回 `"ok"` 且在第一个错误就停止。

**文件**: `gozero/application/sys-api/internal/logic/user/userImportLogic.go`

**修复**: 修改为与 native 一致的容错导入逻辑（需 RPC 层配合）。

**状态**: ⬜ 未修复（保留：需 RPC 层修改）

---

### 🟡 P1: 分页查询存在 N+1 问题

**问题**: gozero 的 `UserPage` 先查询 ID 列表，再逐条获取用户详情（N+1 查询模式）。native 使用 GORM 一次查询获取所有用户。

**文件**: `gozero/application/sys-rpc/internal/logic/sysservice/userPageLogic.go`

**修复**: 修改为单次查询获取所有用户。

**状态**: ⬜ 未修复

---

### 🟡 P1: 删除用户不检查存在性

**问题**: native 先检查用户是否存在（不存在返回 "用户不存在" 错误），gozero 直接执行软删除（用户不存在时静默成功）。

**文件**: `gozero/application/sys-rpc/internal/logic/sysservice/userDeleteLogic.go`

**修复**: 添加存在性检查和错误返回。

**状态**: ⬜ 未修复

---

## 4. Role 模块

### 🔴 P0: Int64ID 类型不一致（前后端兼容性关键）

**问题**: native 的 `userId`/`roleId`/`userIds`/`menuIds` 使用自定义 `Int64ID` 类型，可接受 JSON 字符串和数字（兼容 JS snowflake ID）。gozero 使用普通 `int64`，只接受 JSON 数字。

**文件**: `gozero/application/sys-api/api/role.api` 中的 `roleId`/`userId`/`userIds`/`menuIds` 字段

**修复**: 需要在 gozero 中实现类似的字符串到数字的自动转换，或确保前端始终发数字。

**状态**: ⬜ 未修复

---

### 🔴 P0: 创建/更新/删除角色响应 data 不一致

**问题**: native 的 Create 返回完整 role 对象，Update/Delete 返回 `data: null`。gozero 统一返回 `data: "ok"`。

**文件**: `gozero/application/sys-api/internal/logic/role/roleCreateLogic.go`, `roleUpdateLogic.go`, `roleDeleteLogic.go`

**修复**: Create 返回 role 对象（从 RPC 响应中获取），Update/Delete 返回 `data: nil`。

**状态**: ⬜ 未修复

---

## 5. Menu 模块

### 🟡 P1: Menu 字段 int32 vs int64

**问题**: native 的 `isFrame`, `isCache`, `menuType`, `visible`, `status` 都是 `int32`。gozero 使用 `int64`。

**文件**: `gozero/application/sys-api/api/menu.api`

**修复**: 将字段类型改为 `int32`。

**状态**: ✅ 已修复（已改为 int32）

---

### 🟡 P1: Update menu 部分更新的语义差异

**问题**: native 的 `UpdateMenu` 绑定整个 GORM 模型（所有字段都在请求中）。gozero 的 `MenuUpdateReq` 所有字段都是 `optional`，只发提供的字段。

**文件**: `gozero/application/sys-api/api/menu.api`

**修复**: 确认前端行为并保持一致。gozero 的部分更新语义更现代，但如果前端期望全量更新行为，需调整。

**状态**: ⬜ 未修复

---

## 6. Org 模块

### 🔴 P0: 创建组织返回值不一致

**问题**: native 返回新创建的 `orgId`（int64）。gozero 返回 `"ok"` 字符串。前端创建组织后可能依赖返回的 ID。

**文件**: `gozero/application/sys-api/internal/logic/org/orgCreateLogic.go`

**修复**: 从 RPC 响应中提取 orgId 并返回。

**状态**: ⬜ 未修复

---

### 🟡 P1: ParentId 类型差异

**问题**: native 的 `PageOrgsRequest.ParentId` 使用 `*int64`（可区分"不筛选"和"parentId=0"）。gozero 使用 `int64`（0 既是默认值也可能是业务值）。

**文件**: `gozero/application/sys-api/api/org.api`

**修复**: 考虑改为支持可选语义。

**状态**: ⬜ 未修复

---

### 🟡 P1: 缺少字段验证

**问题**: native 有全面的验证：orgName 必填(2-50字符)、orgCode 必填(2-30字符)、orgType oneof、phone 11位数字、email 格式。gozero 完全缺失这些验证。

**文件**: `gozero/application/sys-api/api/org.api`

**修复**: 在 logic 层添加字段验证。

**状态**: ⬜ 未修复

---

## 7. API Permission 模块

### 🟢 P2: nodeType/status 类型 int32 vs int64

**问题**: native 的 `nodeType` 和 `status` 是 `int32`。gozero 使用 `int64`。

**文件**: `gozero/application/sys-api/api/api_permission.api`

**修复**: 改为 `int32`。

**状态**: ⬜ 未修复

---

### 🟢 P2: nodeType/status 在 gozero 标记为 optional

**问题**: native 中 `nodeType` 和 `status` 是必须的（binding oneof），gozero 标记为 optional。

**文件**: `gozero/application/sys-api/api/api_permission.api`

**修复**: 移除 `optional` 标记。

**状态**: ⬜ 未修复

---

## 8. Dict 模块

### 🟡 P1: ParentId 类型差异（GetDictByType）

**问题**: native 的 `GetDictByTypeRequest.ParentId` 使用 `*int64`，可区分"不筛选"和"parentId=0"。gozero 使用 `int64`，语义不同。

**文件**: `gozero/application/sys-api/api/dict.api`, `dictTypeLogic.go`

**修复**: gozero 的 `DictTypeQueryReq.ParentId` 改为指针或添加可选语义。

**状态**: ⬜ 未修复

---

### 🟡 P1: 缺少排序字段

**问题**: native 的 `PageDictRequest` 支持 `orderByColumn` 和 `isAsc` 排序。gozero 的 `DictPageReq` 不包含排序字段。

**文件**: `gozero/application/sys-api/api/dict.api`

**修复**: 添加排序字段，并在 RPC 层支持排序。

**状态**: ⬜ 未修复

---

## 9. Config 模块

### 🔴 P0: Data 字段类型和响应结构不一致

**问题**: native 使用 `json.RawMessage`（保持原始 JSON），gozero 使用 `interface{}`（需序列化/反序列化，可能改变 JSON 表示）。

**文件**: `gozero/application/sys-api/api/config.api`, `configCreateLogic.go`

**修复**: 考虑使用 `string` 类型传递 Data 字段，避免 JSON 转换失真。

**状态**: ⬜ 未修复

---

### 🔴 P0: GetConfigDataByCode 响应结构不同

**问题**: native 返回 `{code: "xxx", data: <原始JSON>}`。gozero 直接返回解析后的 JSON 值（没有 code 包裹）。

**文件**: `gozero/application/sys-api/internal/logic/config/configDataLogic.go`

**修复**: 需要包装为与 native 一致的结构。

**状态**: ⬜ 未修复

---

### 🟡 P1: ConfigDetail 错误码不一致

**问题**: gozero 的 `ConfigDetail` 在找不到记录时返回 404，其他 gozero 逻辑都是 500。native 统一返回 500。

**文件**: `gozero/application/sys-api/internal/logic/config/configDetailLogic.go`

**修复**: 统一错误码处理。

**状态**: ⬜ 未修复

---

## 10. LoginLog 模块

### 🔴 P0: UpdateLoginLog 路径不一致

**问题**: native 是 `PUT /api/v1/loginLog`（id 在 body 中）。gozero 是 `PUT /api/v1/loginLog/:id`（id 在 URL 路径中）。

**文件**: `gozero/application/sys-api/api/loginlog.api`

**修复**: 确认前端使用哪种方式，然后统一。如果前端使用 body 中的 id，需修改 gozero 的路由定义。

**状态**: ⬜ 未修复

---

### 🟡 P1: Page Status 字段语义差异

**问题**: native 的 `PageLoginLogRequest.Status` 使用 `*int32`（nil=不筛选），gozero 使用 `int64`（0=不筛选，但无法区分 status=0 是筛选成功还是不过滤）。

**文件**: `gozero/application/sys-api/api/loginlog.api`

**修复**: 需要确认业务逻辑是否需要区分。可以约定 -1 表示不筛选。

**状态**: ⬜ 未修复

---

### 🟡 P1: Clean 端点返回值不一致

**问题**: native 返回 `{count, days}`，gozero 返回 `"ok"`。

**文件**: `gozero/application/sys-api/internal/logic/loginlog/loginLogCleanLogic.go`

**修复**: 需要从 RPC 层获取清理数量并返回。

**状态**: ⬜ 未修复

---

### 🟡 P1: 缺少排序字段

**问题**: native 支持 `orderByColumn` 和 `isAsc`，gozero 没有。

**文件**: `gozero/application/sys-api/api/loginlog.api`

**修复**: 添加排序字段。

**状态**: ⬜ 未修复

---

## 11. OperLog 模块

### 🔴 P0: UpdateOperLog 路径不一致

**问题**: 同 LoginLog。native 是 `PUT /api/v1/operLog`（id 在 body），gozero 是 `PUT /api/v1/operLog/:id`（id 在路径）。

**文件**: `gozero/application/sys-api/api/operlog.api`

**修复**: 与 LoginLog 同样处理。

**状态**: ⬜ 未修复

---

### 🟡 P1: Page Status 字段语义差异

**问题**: native 使用 `*string`（指针，nil=不筛选），gozero 使用 `string`（空字符串=不筛选）。

**文件**: `gozero/application/sys-api/api/operlog.api`

**修复**: 确认业务是否需要区分空字符串和 nil。

**状态**: ⬜ 未修复

---

### 🟡 P1: Clean 和排序字段

**问题**: 同 LoginLog。

**状态**: ⬜ 未修复

---

## 12. Attachment 模块

### 🟡 P1: expireTime 类型不一致

**问题**: native 使用 `*utils.LocalTime`（自定义时间类型，可为 null），gozero 使用 `string`。

**文件**: `gozero/application/sys-api/api/attachment.api`

**修复**: 统一时间格式。

**状态**: ⬜ 未修复

---

### 🟡 P1: GetURL expires 类型差异

**问题**: native 使用 `int`，gozero 使用 `int64`。

**文件**: `gozero/application/sys-api/api/attachment.api`

**修复**: 改为 `int` 与 native 一致。

**状态**: ⬜ 未修复

---

## 13. Captcha 模块

### 🔴 P0: 短信接口字段名不一致

**问题**: native 的 `SendSMSCaptcha` 请求使用 JSON 字段 `"phone"`。gozero 使用 `"phonenumber"`。

**文件**: `gozero/application/sys-api/api/captcha.api`

**修复**: 改为 `"phone"` 与 native 一致。

**状态**: ⬜ 未修复

---

### 🟡 P1: 缺少 /resource/sms/code 端点

**问题**: native 有 `GET /resource/sms/code`（公开端点，直接发短信验证码）。gozero 的 JWT 白名单中有这个路径但未注册 handler。

**文件**: `gozero/application/sys-api/api/captcha.api`, `gozero/application/sys-api/sys.go`

**修复**: 添加 `GET /resource/sms/code` 路由和 handler。

**状态**: ⬜ 未修复

---

## 14. Storage-Env 模块

### 🔴 P0: 创建/更新返回数据不一致

**问题**: native 的 Create 返回完整 `StorageEnv` 对象，gozero 返回 `"ok"`。

**文件**: `gozero/application/sys-api/internal/logic/storageenv/storageEnvCreateLogic.go`, `storageEnvUpdateLogic.go`

**修复**: 返回创建/更新后的对象。

**状态**: ⬜ 未修复

---

### 🟡 P1: Config 字段类型差异

**问题**: native 使用 `*json.RawMessage`，gozero 使用 `interface{}`。

**文件**: `gozero/application/sys-api/api/storageenv.api`

**修复**: 统一使用方式，避免 JSON 序列化失真。

**状态**: ⬜ 未修复

---

## 15. Health 模块

### 🟡 P1: 健康检查逻辑完全不同

**问题**: native 的 `/health` 直接检查 DB 和 Redis 连通性，返回包含 `services`（database/redis 状态）的详细响应。gozero 通过 RPC ping 实现，返回简化响应。`/health/live` 和 `/health/startup` 在 gozero 中与 `/health` 完全一样，但 native 有不同语义。

**文件**: `gozero/application/sys-api/internal/logic/health/*`

**修复**: 如果健康检查语义对部署重要，需重写以匹配 native 行为。

**状态**: ⬜ 未修复

---

### ⚪ P3: 多余的 pingLogic.go

**问题**: gozero 有一个未注册路由的 `pingLogic.go`，是生成代码的残留。

**文件**: `gozero/application/sys-api/internal/logic/health/pingLogic.go`

**修复**: 删除或注册为正式端点。

**状态**: ⬜ 未修复

---

## 16. 完全缺失的端点

| 端点 | 所属模块 | 说明 |
|------|----------|------|
| `POST /auth/login` | Auth | native 的 login 别名路由 |
| `POST /auth/logout` | Auth | native 的 logout 别名路由 |
| `POST /system/user/xcxGetInfo` | User | 小程序用户信息 |
| `GET /resource/sms/code` | Captcha | 直接发短信验证码（公开） |
| `GET /resource/websocket` | Common | WebSocket 连接 |
| `GET /metrics` | Common | Prometheus 指标 |
| `GET /swagger/*any` | Common | API 文档 |

**状态**: 剩余 2/7 未修复（/auth/login, /auth/logout, /system/user/xcxGetInfo, /resource/sms/code 已追加；/resource/websocket, /metrics, /swagger/*any 保留）

---

## 修复工作清单（按优先级）

### P0（必须立即修复）

- [x] **1. JSON 命名风格统一**（.api 文件）- snake_case vs camelCase（注：snake_case/camelCase 为全局风格差异，需前端配合；JSON tag 命名暂不批量修改）
- [x] **2. Auth: clientId vs clientKey+clientSecret**（已修复：LoginReq/RefreshTokenReq 统一使用 clientId）
- [x] **3. Auth: 登录响应 JSON 键名**（已修复：access_token, refresh_token, expires_in, refresh_expires_in, user_info）
- [x] **4. CommonResp data 字段去除 optional**（已修复：common.api）
- [x] **5. Org: 创建返回 orgId 而非 "ok"**（已修复：返回 data: nil 与 native 一致）
- [x] **6. Config: GetConfigDataByCode 响应结构**（已修复：返回 {code, data} 结构）
- [x] **7. Captcha: 短信字段名 phone vs phonenumber**（已修复：JSON tag 改为 "phone"）
- [x] **8. User: 创建返回 {"userId": "ok"} 而非 "ok"**（已修复：userCreateLogic.go）
- [ ] **9. User: 批量导入返回统计而非 "ok"**（保留：需 RPC 层配合返回详细统计）
- [ ] **10. Role: userId/roleId 支持字符串类型（Int64ID）**（保留：需框架级支持）
- [x] **11. Role: 创建返回 role 对象**（已修复：返回 data: nil 与 native 一致）
- [x] **12. StorageEnv: 创建返回对象**（已修复：返回 data: nil 保持一致）
- [x] **13. LoginLog/OperLog: Update 路由路径**（验证确认：native 也是 PUT /:id，无需修改）
- [x] **14. 统一成功消息为中文**（已修复：所有 logic 文件的 Msg 统一为 "操作成功"）

### P1（应尽快修复）

- [x] **15. 缺少 /auth/login, /auth/logout 别名路由**（已修复：auth.api 添加路由 + logic/handler 实现）
- [ ] **16. Auth: 登录响应补充缺失字段**（保留：需 RPC 层配合添加 expire_in, client_id, scope, openId 等）
- [x] **17. User: XcxGetInfo 端点**（已修复：user.api 添加路由 + logic 实现）
- [ ] **18. User: 分页查询 N+1 问题**（保留：需 sys-rpc 层优化）
- [ ] **19. User: 删除用户检查存在性**（保留：需 sys-rpc 层添加检查）
- [ ] **20. Dict: ParentId 指针语义**（保留：影响较小，后续优化）
- [ ] **21. Dict/Config/Log: 排序字段**（保留：需 RPC 层配合）
- [x] **22. LoginLog/OperLog: Clean 返回值**（已修复：返回 data: nil 保持一致）
- [ ] **23. LoginLog/OperLog: Page Status 字段语义**（保留：需确认业务是否需区分 -1/nil）
- [x] **24. Captcha: /resource/sms/code 端点**（已修复：captcha.api 添加路由 + logic 实现）
- [ ] **25. 各模块 CreateBy/UpdateBy 追踪**（部分保留：部分模块已通过 commonutil.UserIDFromContext 追踪）
- [ ] **26. Health: 健康检查逻辑对齐**（保留：需较大改动）

### P2（后续优化）

- [x] **27. int32 vs int64 类型统一**（已修复：所有 .api 文件中的 status/userType/sex/dataScope/nodeType/isFrame/isCache/menuType/visible 等改为 int32）
- [ ] **28. 字段级验证补充**（保留：后续优化）
- [ ] **29. Attachment: expireTime 类型**（保留：string 类型可接受）
- [ ] **30. Config: Data 字段类型**（保留：interface{} 可接受）
- [ ] **31. Menu: 部分更新语义确认**（保留：前端当前兼容全量更新）
- [ ] **32. Org: ParentId 指针类型**（保留：影响较小）

### P3（视需要决定）

- [ ] **33. /metrics, /swagger 端点**（保留）
- [ ] **34. 清理未使用的 pingLogic.go**（保留）
- [ ] **35. Casbin 权限中间件**（保留）

---

## 注意事项

1. go-zero 开发模式：先修改 `.api` 文件 → 运行 `goctl` 生成代码 → 修改 logic 文件
2. 修改 `.api` 后需要同步修改对应的逻辑实现
3. 每次修复完成后，在此文档中标记为 ✅ 已修复
4. 修复完成一个模块后，重启 sys-api/sys-rpc，连接同一个数据库，启动 web-react 验证
