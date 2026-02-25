# SSO OIDC 登录认证系统 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现基于 OIDC 的 SSO 系统，支持自建用户与上游 IdP 代理，单体架构，符合 project_architecture.md 与 code_style.md。

**Architecture:** Clean Architecture 分层；Fosite 实现 IdP，go-oidc 实现 RP（上游联邦）；Gin HTTP；ent ORM；Handler → Service → Storage/Domain/Infra。

**Tech Stack:** Go 1.21+, Gin, ent, ory/fosite, coreos/go-oidc, zap, Viper, Cobra

**Design Doc:** `docs/plans/2025-02-25-sso-oidc-design.md`

---

## Phase 1: 项目脚手架

### Task 1: 初始化 Go 模块与 Cobra

**Files:**
- Create: `go.mod`
- Create: `cmd/root.go`
- Create: `cmd/svr.go`
- Create: `main.go`

**Step 1: 创建 go.mod**

```bash
cd /Users/qinzj/github/superpowers-demo
go mod init github.com/qinzj/superpowers-demo
```

在 go.mod 中设置 `go 1.21`。

**Step 2: 安装 Cobra 并初始化**

```bash
go install github.com/spf13/cobra-cli@latest
cobra-cli init --pkg-name github.com/qinzj/superpowers-demo -a "qinzj" -l mit
```

**Step 3: 添加 svr 子命令**

```bash
cobra-cli add svr
```

**Step 4: 调整 main.go 调用 root**

确保 main.go 调用 rootCmd.Execute()。

**Step 5: Commit**

```bash
git add go.mod go.sum main.go cmd/
git commit -m "chore: init Go module and Cobra CLI"
```

---

### Task 2: 配置与目录布局

**Files:**
- Create: `configs/settings.yaml`
- Create: `internal/server/svr.go`（占位）
- Create: `internal/router/router.go`（占位）
- Modify: `cmd/svr.go` 读取配置并启动

**Step 1: 创建 configs/settings.yaml**

```yaml
server:
  port: 8888
database:
  driver: sqlite3
  dsn: file:./data/sso.db?cache=shared&mode=rwc
log:
  level: info
  output: stdout
  file: ""
  max_size: 100
  max_age: 7
  max_backups: 3
oidc:
  issuer: http://localhost:8888
```

**Step 2: 创建目录与占位文件**

创建 `internal/server/svr.go`、`internal/router/router.go` 空实现；cmd/svr.go 使用 Viper 加载 configs/settings.yaml，打印配置后退出。

**Step 3: 添加依赖**

```bash
go get github.com/spf13/viper
go mod tidy
```

**Step 4: Commit**

```bash
git add configs/ internal/
git commit -m "chore: add config and directory layout"
```

---

## Phase 2: Domain 与 Storage

### Task 3: Domain 模型

**Files:**
- Create: `internal/domain/user.go`
- Create: `internal/domain/session.go`
- Create: `internal/domain/oauth2_client.go`
- Create: `internal/domain/idp_connector.go`

**Step 1: 定义 User**

```go
// internal/domain/user.go
package domain

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
```

**Step 2: 定义 Session、OAuth2Client、IdPConnector**

按设计文档定义；使用 `time.Time`；ID 为 string。

**Step 3: Commit**

```bash
git add internal/domain/
git commit -m "feat(domain): add User, Session, OAuth2Client, IdPConnector"
```

---

### Task 4: ent Schema

**Files:**
- Create: `ent/schema/user.go`
- Create: `ent/schema/session.go`
- Create: `ent/schema/oauth2client.go`
- Create: `ent/schema/idpconnector.go`

**Step 1: 安装 ent**

```bash
go get -u entgo.io/ent/cmd/ent
go run -mod=mod entgo.io/ent/cmd/ent new User Session OAuth2Client IdPConnector
```

**Step 2: 编写 schema**

参考 design doc 字段；User 含 username, email, password_hash；Session 含 user_id, token, expires_at；OAuth2Client 含 client_id, client_secret, redirect_uris；IdPConnector 含 issuer, client_id, client_secret。

**Step 3: 生成代码**

```bash
go generate ./ent
```

**Step 4: Commit**

```bash
git add ent/
git commit -m "feat(storage): add ent schema and generate"
```

---

## Phase 3: Infra 与 Service 基础

### Task 5: pkg/log

**Files:**
- Create: `pkg/log/log.go`

**Step 1: 实现 zap 封装**

支持 level、output（stdout/file）、file 时支持 rotation（max_size, max_age, max_backups）；提供 Logger 接口；从 config 构造。

**Step 2: 添加依赖**

```bash
go get go.uber.org/zap
```

**Step 3: Commit**

```bash
git add pkg/log/
git commit -m "feat(log): add zap logger with rotation"
```

---

### Task 6: infra/password

**Files:**
- Create: `internal/infra/password/password.go`

**Step 1: 写 failing test**

```go
func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("secret")
	require.NoError(t, err)
	require.True(t, Verify("secret", hash))
	require.False(t, Verify("wrong", hash))
}
```

**Step 2: Run test（应失败）**

```bash
go test ./internal/infra/password/... -v
```

**Step 3: 实现 Hash、Verify（bcrypt）**

**Step 4: Run test（应通过）**

**Step 5: Commit**

```bash
git add internal/infra/password/
git commit -m "feat(infra): add password hashing"
```

---

### Task 7: UserService 与 UserRepository

**Files:**
- Create: `internal/service/user/user.go`
- Create: `internal/service/user/repository.go`（接口定义）
- Create: `internal/storage/user_repository.go`（ent 实现）

**Step 1: 在 service/user 定义 UserRepository 接口**

```go
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	ByUsername(ctx context.Context, username string) (*domain.User, error)
}
```

**Step 2: 写 CreateUser failing test**

**Step 3: 实现 UserService.Create、storage 实现**

**Step 4: 测试通过后 Commit**

```bash
git add internal/service/user/ internal/storage/
git commit -m "feat(service): add UserService and UserRepository"
```

---

## Phase 4: OIDC IdP（Fosite）

### Task 8: Fosite 配置与 Storage 适配

**Files:**
- Create: `internal/service/oidc/fosite_storage.go`
- Create: `internal/service/oidc/config.go`

**Step 1: 添加 Fosite 依赖**

```bash
go get github.com/ory/fosite
```

**Step 2: 实现 fosite.Storage 接口**

实现 GetClient、CreateAuthorizeCodeSession、GetAuthorizeCodeSession 等；使用 ent 的 OAuth2Client、Session（或内存/暂存）。先实现最小集使 /authorize 可跑通。

**Step 3: 配置 Fosite OAuth2/OIDC provider**

签发 JWT、设置 issuer、配置 scopes。

**Step 4: Commit**

```bash
git add internal/service/oidc/
git commit -m "feat(oidc): add Fosite storage and config"
```

---

### Task 9: OIDC Handler 与 Router

**Files:**
- Create: `internal/server/http/handler/oidc.go`
- Create: `internal/server/http/handler/http.go`（Gin Engine 构造）
- Modify: `internal/router/router.go` 注册 OIDC 路由

**Step 1: 注册路由**

- GET /.well-known/openid-configuration
- GET /authorize
- POST /token
- GET /userinfo

**Step 2: Handler 调用 Fosite**

authorize 返回 302；token 返回 JSON；userinfo 返回 JSON。Handler 只做绑定与调用，不含业务逻辑。

**Step 3: 集成到 svr 命令**

cmd/svr 启动 Gin，监听 config 中的 port。

**Step 4: 手动验证**

```bash
curl http://localhost:8888/.well-known/openid-configuration
```

**Step 5: Commit**

```bash
git add internal/server/ internal/router/ cmd/
git commit -m "feat(oidc): add OIDC endpoints and handler"
```

---

## Phase 5: 登录与自建用户

### Task 10: AuthService

**Files:**
- Create: `internal/service/auth/auth.go`
- Create: `internal/service/dto/auth.go`

**Step 1: 定义 ValidateCredentials**

接收 username、password；调用 UserRepository.ByUsername；用 infra/password.Verify 校验；返回 User 或 error。

**Step 2: 写 failing test**

**Step 3: 实现**

**Step 4: Commit**

```bash
git add internal/service/auth/ internal/service/dto/
git commit -m "feat(auth): add AuthService.ValidateCredentials"
```

---

### Task 11: 登录页与 Session

**Files:**
- Create: `internal/server/http/handler/login.go`
- Create: `internal/service/auth/session.go`
- Modify: `internal/router/router.go`

**Step 1: Session 创建与校验**

AuthService.CreateSession、GetSession；Session 存 storage（或 Redis，先 DB）。

**Step 2: 登录页 GET /login**

渲染 HTML；redirect_uri、client_id、state 等从 query 传入并回传。

**Step 3: 登录页 POST /login**

解析表单；AuthService.ValidateCredentials；成功则 CreateSession，写 cookie，redirect 回 /authorize。

**Step 4: /authorize 校验 Session**

未登录则 redirect 到 /login。

**Step 5: Commit**

```bash
git add internal/server/http/handler/ internal/service/auth/ internal/router/
git commit -m "feat(login): add login page and session"
```

---

### Task 12: 用户注册

**Files:**
- Create: `internal/server/http/handler/register.go`
- Modify: `internal/service/user/user.go` 添加 Register
- Modify: `internal/router/router.go`

**Step 1: Register 逻辑**

校验 username 唯一、密码强度；Hash 后 Create。

**Step 2: 注册页 GET/POST /register**

DTO 带 binding；Handler 调用 UserService.Register。

**Step 3: Commit**

```bash
git add internal/server/http/handler/ internal/service/user/ internal/router/
git commit -m "feat(register): add user registration"
```

---

## Phase 6: 上游 IdP 代理

### Task 13: FederationService

**Files:**
- Create: `internal/service/federation/federation.go`
- Create: `internal/infra/oidc_client/client.go`

**Step 1: 添加 go-oidc**

```bash
go get github.com/coreos/go-oidc/v3/oidc
```

**Step 2: infra/oidc_client**

根据 IdPConnector 配置创建 oidc.Provider、oauth2.Config；实现 Exchange、UserInfo 等。

**Step 3: FederationService.LoginWithUpstream**

接收 connector_id、state、code；交换 token；拉取 userinfo；身份映射/链接到本系统 User；创建 Session。

**Step 4: 写 failing test**

**Step 5: 实现并 Commit**

```bash
git add internal/service/federation/ internal/infra/oidc_client/
git commit -m "feat(federation): add FederationService and OIDC client"
```

---

### Task 14: 上游 IdP 回调 Handler

**Files:**
- Create: `internal/server/http/handler/callback.go`
- Modify: `internal/router/router.go`

**Step 1: GET /auth/callback/{connector_id}**

解析 code、state；调用 FederationService.LoginWithUpstream；创建 Session 后 redirect 到应用。

**Step 2: 登录页增加「企业 SSO」入口**

根据 IdPConnector 列表渲染链接到 /auth/federation/{connector_id}（发起 OAuth 重定向）。

**Step 3: Commit**

```bash
git add internal/server/http/handler/ internal/router/
git commit -m "feat(federation): add callback handler and login entry"
```

---

## Phase 7: 健康检查与错误处理

### Task 15: 健康检查

**Files:**
- Create: `internal/server/http/handler/health.go`
- Modify: `internal/router/router.go`

**Step 1: GET /healthz**

返回 200，无依赖检查。

**Step 2: GET /ready**

检查 DB 连接；失败返回 503。

**Step 3: Commit**

```bash
git add internal/server/http/handler/ internal/router/
git commit -m "feat(health): add /healthz and /ready"
```

---

### Task 16: 统一错误响应

**Files:**
- Create: `internal/server/http/handler/errors.go`
- Modify: 各 handler 使用统一错误映射

**Step 1: 定义错误响应结构**

```go
type ErrorResp struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}
```

**Step 2: 中间件或辅助函数**

从 gin.Context 提取 request_id；根据 error 类型映射 HTTP 状态码；写入 JSON。

**Step 3: Commit**

```bash
git add internal/server/http/handler/
git commit -m "feat(errors): add unified error response"
```

---

## Phase 8: 测试与文档

### Task 17: 集成测试

**Files:**
- Create: `test/integration/oidc_test.go`

**Step 1: 编写 OIDC 流程集成测试**

使用 httptest 或实际 HTTP 调用；验证 /.well-known、/authorize 重定向、/token 响应格式。

**Step 2: Run**

```bash
go test ./test/integration/... -v
```

**Step 3: Commit**

```bash
git add test/
git commit -m "test: add OIDC integration tests"
```

---

### Task 18: 文档与 README

**Files:**
- Modify: `README.md`
- Create: `docs/design/sso-api.md`（API 概览）

**Step 1: README**

项目简介、如何运行、配置说明、OIDC 端点列表。

**Step 2: docs/design/sso-api.md**

简要 API 说明，引用 design doc。

**Step 3: Commit**

```bash
git add README.md docs/
git commit -m "docs: add README and API overview"
```

---

## Execution Handoff

计划已保存至 `docs/plans/2025-02-25-sso-oidc-implementation.md`。两种执行方式：

**1. Subagent-Driven（本会话）** — 按任务调度子 agent，任务间做 review，迭代更快  

**2. Parallel Session（单独会话）** — 在新会话中用 executing-plans，按检查点批量执行  

你更倾向哪种方式？
