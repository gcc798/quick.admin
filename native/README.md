# quick.admin

## 项目说明

这是一个极简脚手架仓库，用来并行放置 `quick.admin` 的多套后端实现和前端工程，便于保留业务基线、承载不同框架版本以及统一开展开发与联调。

当前仓库更偏向工程骨架与实现对照集合，重点是：

- 保留 `native` 作为业务基线
- 提供 `gozero`、`kratos` 两套重写实现
- 提供 `web` 前端工程作为脚手架组成部分

当前主要目录：

- `native/`
  - 原始后端实现
  - 作为业务基线和对照参考
- `gozero/`
  - 基于 go-zero 的后端重写版本
- `kratos/`
  - 基于 Kratos 的后端重写版本
- `web/`
  - 前端工程

## 仓库结构

```text
quick.admin/
├── native/
├── gozero/
├── kratos/
├── web/
├── LICENSE
└── README.md
```

## 各子工程职责

### `native/`

原始业务后端。

特点：

- 作为业务语义基线
- 路由、参数、返回结构、错误语义都以它为重要参考

### `gozero/`

基于 go-zero 的重写版本。

特点：

- 保留了 `sys-api` / `sys-rpc` 分层
- 主要用于和原始实现做框架迁移对照

### `kratos/`

基于 Kratos 的重写版本。

特点：

- 当前采用 monorepo 结构
- 主要服务位于：
  - `kratos/application/sys-api`
  - `kratos/application/sys-rpc`
- 共享 proto 位于：
  - `kratos/api/system/v1`
- 共享基础能力位于：
  - `kratos/pkg`
- 详细文档位于：
  - `kratos/docs`

### `web/`

前端工程。

特点：

- 对接后端接口
- 在重写过程中，尽量保持零改动或最小改动适配后端

## 当前约定

- `native/` 作为业务基线
- `gozero/` 和 `kratos/` 是两套独立重写实现
- 前端联调时，需要明确当前对接的是哪一套后端
- 如果对比接口契约、行为或返回结构，优先参考 `native/`

## 常见入口

### 启动或开发 Kratos 版本

目录：

- [/Users/guoc/dev/code_go/src/quick.admin/kratos](/Users/guoc/dev/code_go/src/quick.admin/kratos)

常用命令：

```bash
cd kratos
make conf
make proto-all
make wire
make ent
make test
make build-all
```

### 启动或开发 go-zero 版本

目录：

- [/Users/guoc/dev/code_go/src/quick.admin/gozero](/Users/guoc/dev/code_go/src/quick.admin/gozero)

### 开发前端

目录：

- [/Users/guoc/dev/code_go/src/quick.admin/web](/Users/guoc/dev/code_go/src/quick.admin/web)

## 说明

如果后续继续扩展新的后端实现或新的微服务，建议继续保持：

- 仓库根目录按实现或子系统分目录
- 各实现内部再按自身框架规范组织
- 公共契约、文档和基础设施说明尽量写在对应子工程内
