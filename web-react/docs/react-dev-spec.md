# web-react 开发规范

## 1. 项目目标

`web-react` 是对现有 [web](/Users/guoc/dev/code_go/src/nai-tizi/web) 管理后台的 React 全量重写工程。

这份工程同时承担两个目标：

- 保证 React 版与现有后台的功能、路由、权限、主要交互路径保持等价。
- 作为学习 React 的工程，代码实现必须优先满足“可读、可学、可跟踪”，避免为了抽象而抽象。

本工程不是在旧 Vue 工程上做局部改造，而是一个新的前端实现。因此：

- 不延续第一次开发变动留下的前端兼容残留作为默认策略。
- 不以旧前端中的历史字段混用作为新的设计标准。
- 仅在后端契约无法完整表达返回结构时，才将旧前端作为补充参考。

## 2. 契约来源与优先级

React 工程开发时，接口真相来源按以下优先级执行，后级只能补充前级，不能覆盖前级。

### 2.1 第一优先级：`.api` 文件

主来源目录：

- [gozero/application/sys-api/api](/Users/guoc/dev/code_go/src/nai-tizi/gozero/application/sys-api/api)

用途：

- 定义 URL
- 定义 HTTP Method
- 定义 Path 参数
- 定义 Query 参数
- 定义 Body 请求结构

约束：

- 前端 API 模块必须直接按 `.api` 文件实现请求。
- 禁止根据旧前端代码“猜测”接口 URL。
- 禁止随意修改字段名以贴合旧页面习惯。

业务接口默认固定前缀：

- `/api/v1/*`

非 `/api` 例外接口必须单列处理：

- `/login`
- `/logout`
- `/auth/refresh`
- `/me`
- `/captcha/*`
- `/health*`

### 2.2 第二优先级：`internal/types/types.go`

参考文件：

- [gozero/application/sys-api/internal/types/types.go](/Users/guoc/dev/code_go/src/nai-tizi/gozero/application/sys-api/internal/types/types.go)

用途：

- 校验 `.api` 生成后的请求结构
- 二次确认字段名、tag、路径参数和表单参数定义

约束：

- 优先用于确认“请求结构”，不是最终响应结构的唯一依据。
- 当前端实现与 `.api` 理解出现偏差时，先对照该文件检查字段名和参数位置。

### 2.3 第三优先级：`internal/logic/**`

参考目录：

- [gozero/application/sys-api/internal/logic](/Users/guoc/dev/code_go/src/nai-tizi/gozero/application/sys-api/internal/logic)

用途：

- 确认 `CommonResp.data` 的真实返回值结构
- 判断分页数据、登录返回、详情返回、树结构返回的真实形态

已确认的后端返回模式：

#### 认证返回

登录与刷新返回以 [auth/common.go](/Users/guoc/dev/code_go/src/nai-tizi/gozero/application/sys-api/internal/logic/auth/common.go) 为准，核心字段为：

- `accessToken`
- `refreshToken`
- `expiresIn`
- `refreshExpiresIn`
- `userInfo`

其中 `userInfo` 已确认包含：

- `userId`
- `username`
- `nickname`
- `phonenumber`
- `email`
- `avatar`
- `userType`

#### 分页返回

分页类返回以 [commonutil/common.go](/Users/guoc/dev/code_go/src/nai-tizi/gozero/application/sys-api/internal/logic/commonutil/common.go) 中的 `PageData` 为准，固定字段为：

- `records`
- `total`
- `size`
- `current`
- `pages`

因此 React 版所有分页页面统一读取这一套字段，不再发明新的分页结构。

#### 树结构、详情、权限类返回

以下类型的接口通常直接把实际结果放进 `CommonResp.data`：

- 菜单树
- 组织树
- 详情对象
- 角色权限列表
- 角色菜单 ID 列表

实现时应优先查看对应 `logic` 中 `Data:` 的实际赋值方式，不要只根据名字推断。

### 2.4 第四优先级：现有 `web` 工程

参考目录：

- [web/src/api](/Users/guoc/dev/code_go/src/nai-tizi/web/src/api)
- [web/src/views](/Users/guoc/dev/code_go/src/nai-tizi/web/src/views)
- [web/src/types](/Users/guoc/dev/code_go/src/nai-tizi/web/src/types)

用途：

- 当前三层无法完整判断返回值结构时，补充页面实际使用方式
- 辅助判断已有页面如何消费返回数据

约束：

- 只能补充，不能覆盖后端源码。
- 如果旧前端字段和后端真实返回不一致，以后端为准。
- 旧前端中的历史兼容逻辑，默认不迁移。

### 2.5 统一优先级结论

当 `.api`、`internal/types`、`internal/logic`、现有 `web` 四者不一致时，统一按以下顺序裁决：

1. `.api`
2. `internal/types`
3. `internal/logic` 中的真实返回组装
4. 现有 `web` 工程参考

## 3. 技术栈与目录约束

React 工程固定采用以下技术栈：

- React
- TypeScript
- Vite
- Ant Design
- React Router
- Zustand
- Axios

如无明确必要，不引入额外全局状态框架或表单引擎。

建议目录结构：

- `src/app`
- `src/router`
- `src/store`
- `src/api`
- `src/features`
- `src/components`
- `src/pages`
- `src/styles`
- `src/types`
- `docs`

约束：

- 每个目录职责单一，避免页面、请求、状态、公共组件互相混放。
- 公共组件放 `src/components`，业务私有组件优先放到对应 `features` 或 `pages` 下。
- 页面目录命名和功能模块命名尽量与现有后台保持对应，便于对照学习和迁移。

## 4. 页面迁移范围

React 版第一阶段即按“全量等价迁移”执行，不做核心页裁剪。

必须覆盖的功能范围：

- 登录
- 验证码
- 动态菜单路由
- 按钮权限
- 主题切换
- 仪表盘
- 用户管理
- 角色管理
- 菜单管理
- 组织管理
- 字典管理
- 配置管理
- 存储环境管理
- 文件管理
- 登录日志
- 操作日志
- `403`
- `404`
- `500`

实现要求：

- 尽量保持页面结构、查询项、表格列、操作按钮、弹窗流程一致。
- 不要求逐像素复刻 Vue 组件内部细节。
- 但不能丢失核心能力，尤其是：
  - 动态路由
  - 权限控制
  - 树表展开收起
  - Monaco 编辑器
  - 文件上传、预览、下载

## 5. 接口开发规范

### 5.1 API 模块拆分规则

每个后端模块对应一个前端 API 文件，命名尽量与后端一致，例如：

- `src/api/auth.ts`
- `src/api/user.ts`
- `src/api/role.ts`
- `src/api/menu.ts`
- `src/api/org.ts`
- `src/api/dict.ts`
- `src/api/config.ts`
- `src/api/storageenv.ts`
- `src/api/attachment.ts`
- `src/api/loginlog.ts`
- `src/api/operlog.ts`

### 5.2 请求实现规则

- URL 必须直接参考 `.api`。
- Method 必须直接参考 `.api`。
- Path 参数、Query 参数、Body 字段名必须直接参考 `.api` 与 `internal/types`。
- 不允许为了复用旧页面代码而篡改接口字段名。

### 5.3 统一响应处理规则

前端请求层统一处理以下内容：

- `CommonResp<T>`
- 分页结果展开
- 401 失效处理
- Token 刷新
- 常规错误提示

基础类型应固定定义为：

```ts
export interface CommonResp<T = unknown> {
  code: number;
  msg: string;
  data?: T;
}

export interface PageData<T = unknown> {
  records: T[];
  total: number;
  size: number;
  current: number;
  pages: number;
}

export interface AuthLoginData {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  refreshExpiresIn: number;
  userInfo: {
    userId: number;
    username: string;
    nickname: string;
    phonenumber?: string;
    email?: string;
    avatar?: string;
    userType: number | string;
  };
}
```

### 5.4 返回值建模规则

- 页面实体类型优先按真实返回值建模。
- 不强行继承旧 Vue 工程里的历史命名。
- 如果某个返回值无法仅通过 `.api` 和 `internal/types` 理解，就继续参考后端 `logic` 返回代码。
- 如果仍不能准确判断，再参考现有 `web` 的返回值定义和页面用法。

实现原则：

- 请求类型尽量贴近后端契约。
- 页面展示类型可以做少量、明确、可追踪的转换。
- 禁止扩散“全局字段兼容魔法”。

## 6. 路由、权限与状态规范

### 6.1 路由

- 路由路径与现有后台保持一致。
- 登录页、错误页与业务页拆分清楚。
- 动态菜单路由仍然由菜单树转换生成。
- 刷新页面后要能恢复已加载的动态路由。

### 6.2 权限

- 按钮权限继续基于权限码控制。
- 路由权限与按钮权限都需要保留。
- 权限判断逻辑集中放在统一模块中，不允许散落在各页面重复实现。

### 6.3 状态

建议集中维护以下状态：

- `auth`
- `permission`
- `app`
- `theme`

约束：

- 登录态不能由页面局部状态自行维护。
- 权限数据不能在多个页面重复拉取和各自缓存。
- 主题与布局状态应统一管理，避免页面自行决定全局样式行为。

## 7. React 实现风格

这是学习工程，代码风格以“直白、稳定、可理解”为优先。

### 7.1 优先采用的实现方式

- 组件职责明确
- 数据流单向清晰
- 副作用集中到 hooks 或请求层
- 公共能力适度抽离

### 7.2 建议保留的通用封装

- `BasicTable`
- `BasicForm`
- `BasicModal`
- `PermissionGate`
- `AppLayout`

这些封装的目标是减少重复，而不是隐藏逻辑。

### 7.3 明确避免的实现方式

- 过度泛型化的工厂函数
- 层层包裹、难以追踪的数据流
- 为“高级感”引入过深 hooks 链
- 把简单页面抽象成难学的元编程模式

判断标准很简单：

- 如果一段代码不利于学习和排错，就不要为了“优雅”保留它。

## 8. 中文注释规范

本工程必须有良好的中文注释。

注释目标：

- 说明“为什么这样写”
- 说明“这里依赖哪个后端约束”
- 说明“这里有哪些副作用或边界条件”

### 8.1 必须重点写注释的位置

- 请求拦截器
- 401 刷新 Token 流程
- 动态路由生成
- 权限判断
- 分页适配
- 树结构转换
- Monaco 编辑器数据处理
- 文件预览与下载分支

### 8.2 不要写的注释

- 逐行翻译式注释
- “给变量赋值”这种无信息量注释
- 简单 JSX 结构说明
- 普通 CRUD 按钮点击这类显而易见的动作说明

注释原则：

- 注释应该帮助未来的你快速看懂代码。
- 注释不是为了增加字数，而是为了减少误解。

## 9. 明确不做的事项

以下内容不纳入当前 React 重写工程的默认范围：

- 不写前端单元测试
- 不为了兼容旧前端残留而扩散新的兼容层
- 不修改后端接口定义来迎合前端实现
- 不为了学习目的引入与项目无关的新框架
- 不做过度抽象、过度封装、过度元编程

## 10. 验收清单

React 工程至少完成以下手工验收后，才视为达到可交付状态：

### 10.1 认证与全局能力

- 登录成功
- 退出登录成功
- Token 过期后自动刷新
- 刷新失败后正确回到登录页
- 页面刷新后能恢复动态路由
- 按钮权限隐藏正确

### 10.2 业务页面主流程

每个业务页至少验证一次以下主流程中的核心路径：

- 列表加载
- 查询
- 新增
- 编辑
- 删除

### 10.3 必测复杂功能

- 菜单树展开与收起
- 组织树展开与收起
- 文件上传
- 文件预览
- 文件下载
- Monaco 编辑器显示与编辑

## 11. 实施原则总结

可以把这份规范归纳成三条开发原则：

1. 先看后端契约，再写前端代码。
2. 先保证清晰可学，再考虑抽象复用。
3. 先保证功能等价，再考虑局部优化。

只要实现过程中出现分歧，默认按这三条原则回到最稳妥的方案。
