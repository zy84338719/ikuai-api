# AGENTS.md — ikuai-api

> Go SDK for the iKuai v4.0 REST API (routers running iKuai OS 4.x).
> 当前版本：`v1.2.0` · module：`github.com/zy84338719/ikuai-api`

本仓遵循 [oekit](https://oekit.github.io) 的工程规范：分层契约、错误处理红线、SDK 包日志规则、生命周期签名、零外部依赖门禁。

## 本仓包含

`service/`（生成的 typed client）、`session/`（凭据持久化）、`output/`（响应渲染）、`codegen/`（生成器）、`skills/`（面向 Agent 的指南）。根包是 HTTP client 与契约数据（`v4_catalog.go` 端点目录 + `v4_schemas.go` body 契约）。

## 硬规则

### 依赖

- `go.mod` **只允许 `github.com/oekit/*`**。任何其他外部模块一律禁止。
- 强制点：CI `Dependency Gate` job 与本地 `make zero-dep`。
- 需要三方能力时：拆独立仓，或改用标准库实现，**不进本仓**。

### 错误处理

- 一律使用 `github.com/oekit/pkg/errors`：`errors.Wrap(err, "ctx", kv...)`（nil-safe）、`errors.Cause`。
- `APIError` **内嵌 `*oekerrors.BizError`**，iKuai 码即 BizCode，`StatusFor` 推导 HTTPCode。
- **类型同一性红线**：同一进程内只允许一套 errors 包。`APIError.Unwrap()` 返回内嵌 BizError，oekit 的 `errors.AsBizError` 必须能识别本仓错误——改 `Unwrap` 等于打断 `mw.Fail`。
- SDK 本地校验错误用 **4xxx 业务码段**（`errCodeValidation = 4900`），与路由器返回的码区分。
- 所有 error 必须处理，禁止 `_` 忽略（`resp.Body.Close()` 等纯资源释放除外）。

### 日志

- 包级 logger **必须** `logx.Delegating("ikuai-api")` 懒解析（见 `logger.go`）。
  禁止 `logx.Default()`、`logx.WithService` 在包级声明——service name 属于消费方，且 init 期求值看不到 main 里的 `SetDefault`。
- kv 结构化风格为默认；`Ctx*` 方法自动携带 trace_id；禁止 printf 风格。
- `SetLogLevel` 控制本 SDK 事件的级别下限。

### 并发安全

- 禁止裸 `go func`：用 `github.com/oekit/concurrency/async` 的 GoSafe 系列（panic 被 recover）。
  当前本仓无并发原语，改动时注意。
- 共享状态 mutex / atomic 保护。

### 生命周期

- `Close(ctx context.Context) error` + `sync.Once` 幂等（见 `Client.Close`）。
- 新增可关闭的资源遵循同签名。

### Option 与命名

- 函数选项模式 `XxxOption`（**禁止裸 `Option`**，多包 co-import 冲突）。
  本仓既有 `ClientOption`、`ErrorOption`、`RenderOptions`。
- 导出符号 PascalCase + godoc 注释；**包文档放 doc.go**。

### 契约数据的唯一真相源

这是本仓区别于普通 SDK 的地方，改动前务必读 `skills/SKILL.md`：

- `v4_catalog.go` — 157 个端点的路径、HTTP 动词、查询方言（`Query`）。路径已真机逐条验证，文件头记录了被证伪的几种写法。
- `v4_schemas.go` — 73 个端点的 create 默认值、update 白名单、必填字段。
- 两处都由 `codegen` 读取生成 `service/`。**改契约改这两处，然后 `make generate`**，不要手改生成代码。
- 固件的不一致之处是**刻意复现**的（`/auth/users` 的 `src_addr` 是裸 map、`load-balance` 的 `weight` 是字符串等）。归一化会让写入失败，测试锁住了这些行为。

### 查询方言（最容易改错的地方）

**已真机验证**（勿凭推测推翻）：分页大小参数是 `limit`，`page_size` 会被静默忽略并返回全部；
`order` 是方向、`order_by` 是字段，全局统一——曾有"security 组排序反转"的说法，真机证明那是 CLI 的
flag 映射怪癖，反转配对返回空列表。

三种响应包裹都要认：`data`、`results`、以及 `<group>_data` + `<group>_total`
（对象组与 DHCP 静态列表走这种，成员不在数组里）。

`V4Schemas` 的白名单已逐字段对真机响应核对过：固件返回但白名单没有的字段分两类——
可写的已补齐（漏掉会让 read-modify-write 静默清空设置），只读的保留在白名单外
（硬件读数与计算状态，发回去会被拒）。改动白名单前先按同样方式复核。

改路径或方言前先读 `v4_catalog.go` 文件头，那里记录了每条被探针证伪的写法（VPN 的 `/services`
后缀、`/ip6-objects`、根级 AC 端点）。

### 测试

- 契约类改动必须补测试：新增查询方言、新增 schema、改变排序键语义，都要有断言。
- `go test -race ./...` 必须全绿；`make zero-dep` 必须通过。

## 常用命令

```bash
make test        # go test ./...
make check       # fmt + vet + test
make generate    # 从 v4_catalog.go / v4_schemas.go 重新生成 service/
make zero-dep    # 依赖门禁：只允许 github.com/oekit/*
```

## 已知缺口

- legacy 大写方言与 `/auth/online-users` 的 mixed 方言无法在本机验证：`downstream`/`cameras`/`switch`
  都是空表，两种拼写返回同样的空结果，区分不出来。仍按已记录的契约生成。
- `/vpn/ikev2/services` 在被验证的固件上 404，属于功能缺失而非路径错误。


## 文档

- `skills/SKILL.md` — 能力地图、查询方言表、错误码
- `skills/monitoring.md` — 各监控端点该用哪个查询方法
- `skills/writing.md` — 写操作约定、body 契约、固件不一致清单
- `skills/rendering.md` — 输出格式化
