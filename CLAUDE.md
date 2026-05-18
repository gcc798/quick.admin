# CLAUDE.md

本文件供 Claude Code 快速理解仓库。核心信息与 `AGENTS.md` 保持一致。

## 项目是什么

`nai-tizi` 是一个后台系统脚手架工程。本质是一套后端业务逻辑，分别用三套 Go 后端实现承载：

- `native/`：开发者从零手搓的原生 Go 实现，是业务基线。
- `kratos/`：基于 Kratos 开源微服务框架的实现。
- `gozero/`：基于 go-zero 开源微服务框架的实现。

前端以 `web-react/` 为当前推荐工程。它通过统一 HTTP 契约与后端交互，目标是同一套前端无缝对接三套后端实现。`web/` 是历史 Vue 前端，严禁在后续开发中读取或参考，目录后续可能移除。

## 最重要的理解顺序

1. 先看 `native/`，理解业务语义、接口契约、请求响应结构和错误处理。
2. 再看当前要修改的目标实现：`kratos/` 或 `gozero/`。
3. 如果涉及前端联调，再看 `web-react/src`。

不要把三套后端理解为三套不同业务。它们是同一业务能力在不同框架下的实现，并且应该对 `web-react` 暴露同一套 HTTP 契约。

## 目录速览

```text
nai-tizi/
├── native/      # 手搓原生 Go 后端，业务基线
├── kratos/      # Kratos 微服务框架实现
├── gozero/      # go-zero 微服务框架实现
├── web-react/   # 当前推荐 React 前端
├── web/         # 历史 Vue 前端，严禁后续读取或参考，可能移除
├── README.md
├── AGENTS.md
└── CLAUDE.md
```

## native 后端

`native/` 是最重要的参考实现。

常用入口：

- `native/cmd/api/main.go`：HTTP API 启动入口。
- `native/internal/router/`：路由。
- `native/internal/controller/`：控制器。
- `native/internal/service/`：业务逻辑。
- `native/internal/domain/model/`：GORM 模型。
- `native/internal/domain/request/`：请求结构。
- `native/internal/domain/response/`：响应结构。
- `native/pkg/`：基础能力。
- `native/docs/swagger/`：Swagger 生成包，编译会用到。

常用命令：

```bash
cd native
go test ./...
make swagger
```

## Kratos 后端

`kratos/` 是基于 Kratos 的微服务框架版本。

关键结构：

- `kratos/api/system/v1/`：proto 契约。
- `kratos/application/sys-api/`：HTTP API 服务。
- `kratos/application/sys-rpc/`：gRPC 服务、业务逻辑和数据访问。
- `kratos/application/sys-rpc/ent/schema/`：Ent schema。
- `kratos/pkg/`：通用能力。

常用命令：

```bash
cd kratos
make conf
make proto
make ent
make wire
go test ./...
```

注意：Kratos 的 `.pb.go` 文件包含 protobuf descriptor。改 proto package、go_package 或 module 路径后，要用生成命令重建，不要只做文本替换。

## go-zero 后端

`gozero/` 是基于 go-zero 的微服务框架版本。

关键结构：

- `gozero/application/sys-api/`：对外 API 服务。
- `gozero/application/sys-rpc/`：内部 RPC 服务。
- `gozero/application/sys-api/sys.api`：API 描述文件。

常用命令：

```bash
cd gozero
go test ./...
```

## web-react 前端

`web-react/` 是当前推荐前端。后端改造必须尽量保持它无缝可用。

前端契约原则：

- `web-react` 只面向一套 HTTP 接口语义。
- 前端只要能和 `native` 正常交互，`kratos` 和 `gozero` 也应该能被同一套前端调用。
- 不应为了开发不同后端，让前端写兼容逻辑、分支判断或框架特化适配。
- 如果某套后端不能被同一套前端调用，优先修该后端的 HTTP 契约。

常用命令：

```bash
cd web-react
pnpm install
pnpm dev
pnpm build
```

如果接口契约有变，检查 `web-react/src` 中的 API 调用、类型定义和页面使用处。

## 开发约束

- 业务行为和 HTTP 契约以 `native/` 为基线。
- `kratos/` 和 `gozero/` 要实现同样业务，不要引入不同语义或不同前端契约。
- 三套 Go 后端是独立子工程，各自运行测试。
- `web-react/` 不应为某个后端版本做特殊兼容。
- 不要删除 `native/docs/swagger/`，除非同步调整 Swagger 生成和导入逻辑。
- 修改生成代码时优先运行框架生成命令。

## 快速定位

- 查接口行为：`native/internal/router`、`native/internal/controller`、`native/internal/service`。
- 查数据模型：`native/internal/domain/model`。
- 查 Kratos 契约：`kratos/api/system/v1`。
- 查 Kratos 核心实现：`kratos/application/sys-rpc/internal`。
- 查 go-zero API/RPC：`gozero/application/sys-api`、`gozero/application/sys-rpc`。
- 查当前前端：`web-react/src`。
