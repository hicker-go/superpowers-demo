# SSO OIDC 登录认证系统 设计文档

> 设计日期：2025-02-25
> 状态：已批准

## 1. 概述

基于 OIDC 协议的标准 SSO 登录认证系统，单体架构，同时扮演 IdP 与 RP。支持自建用户（数据库）与上游 IdP 代理（联邦登录）。

**技术选型：** Fosite（IdP）+ go-oidc（RP 与上游 IdP 通信）+ Gin + ent + zap

---

## 2. 架构与分层

### 2.1 目录结构（符合 project_architecture.md）

```
cmd/
  root.go, svr.go
configs/
  settings.yaml
docs/
  changelog/     # SDD、变更记录
  design/        # PRD、架构、API
internal/
  domain/        # User, Session, OAuth2Client, IdPConnector
  storage/       # ent 生成 (storage/databases/ent)
  server/
    http/
      handler/   # 仅调用 service
  infra/         # Redis、上游 IdP 客户端；仅被 service 使用
  router/        # 路由定义
  service/       # auth, user, oidc, federation, dto
pkg/
  log/           # zap，支持落盘与切割
```

### 2.2 依赖规则

| 规则 | 说明 |
|------|------|
| Handler 不直接调用 dao/storage/domain/infra | 仅调用 service |
| dao/storage/domain 仅被 service 使用 | 数据访问集中在 service |
| infra 为叶依赖 | 可被 service 使用，不反向依赖 internal |
| 接口定义在消费方 | service 定义接口，storage 实现 |
| 禁止循环依赖 | 包间依赖单向 |

---

## 3. 代码风格约束

实现时必须遵守 `.ai-context/rules/code_style.md` 全文约束，摘要如下。

### 3.1 Formatting

- `gofmt -s`，`goimports`
- 行宽 ≤ 120
- import 分组：标准库 → 第三方 → 本地

### 3.2 File Layout

- 顺序：package → import → const → var → type → func
- 方法紧邻 receiver 类型

### 3.3 Naming

- 包名：小写、单数、无下划线
- 文件名：小写，必要时下划线
- 缩写：ID、URL、HTTP、JSON、DB
- Receiver：1–2 字母，同类型一致
- 避免重复包名

### 3.4 DTO

- 显式 `json` 标签
- Gin：`binding:"required"`，`min=3,max=64`
- 可选字段 `omitempty`

### 3.5 Functions

- 动词命名（CreateUser、ValidateEmail）
- `error` 放在最后返回值
- 大类型/需修改用指针 receiver

### 3.6 Interfaces

- 按行为命名（Reader、Repository）
- 在消费方定义（service 层）

### 3.7 Errors

- `fmt.Errorf("...: %w", err)`
- Sentinel：`var ErrNotFound = errors.New("not found")`
- `errors.Is` / `errors.As`
- service/handler 不 panic

### 3.8 Context & Concurrency

- `context.Context` 作为首参，不存入 struct
- Goroutine 有明确退出路径
- 不复制含 Mutex 的值
- 使用 `defer` 做资源释放

---

## 4. 组件与职责

### 4.1 Handler（internal/server/http/handler/）

| 文件 | 职责 |
|------|------|
| oidc.go | /authorize, /token, /userinfo, /.well-known/openid-configuration |
| login.go | 登录页 GET/POST、登出 |
| register.go | 用户注册 |
| callback.go | 上游 IdP 回调 |

### 4.2 Service（internal/service/）

| 子包 | 职责 |
|------|------|
| auth | 登录校验、会话创建/吊销 |
| user | 用户 CRUD、密码哈希/校验、注册 |
| oidc | Fosite 封装、Storage 实现、用户信息注入 |
| federation | 上游 IdP OAuth 流程、身份映射与链接 |

### 4.3 Storage（internal/storage/databases/ent）

- ent 生成
- 模型：User、Session、OAuth2Client、IdPConnector

### 4.4 Infra（internal/infra）

- oidc_client：上游 IdP 的 go-oidc 客户端
- password：bcrypt 哈希

---

## 5. 数据流

### 5.1 自建用户 OIDC（Authorization Code）

1. RP → /authorize（client_id, redirect_uri, scope, state）
2. 未登录 → 登录页
3. 提交用户名/密码 → AuthService.Validate → 创建 Session
4. 重定向回 /authorize → Fosite 签发 code → redirect to RP
5. RP POST /token（code）→ Fosite 签发 access_token、id_token

### 5.2 上游 IdP 代理

1. 用户选择「企业 SSO」→ 以 RP 身份 redirect 到上游 IdP
2. 上游回调 → go-oidc 交换 token、拉取 userinfo
3. FederationService 身份映射/链接 → 创建本系统 User 与 Session
4. 重定向回应用 → 走自建用户后续流程

### 5.3 Token 刷新

- grant_type=refresh_token → Fosite 签发新 access_token、id_token

---

## 6. 错误处理与安全

- Handler：映射 error → HTTP 状态码与统一错误体
- Service：错误用 `%w` 包装，使用 sentinel
- 统一错误体：`{ code, message, request_id }`
- 健康检查：/healthz（存活）、/ready（就绪，含 DB）
- 安全：CSRF、限流、密码不落日志、Token 合理过期

---

## 7. 测试策略

- 单元测试：Service，mock storage/infra
- 集成测试：infra、DB、ent
- 契约测试：OIDC 端点、登录/注册 API
- TDD：先写 failing test 再实现

---

## 8. 技术栈

- HTTP：Gin
- ORM：ent
- 日志：zap（pkg/log，落盘与切割）
- OIDC：ory/fosite + coreos/go-oidc
- 配置：Viper + settings.yaml
