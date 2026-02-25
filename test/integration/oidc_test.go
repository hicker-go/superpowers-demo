// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/qinzj/superpowers-demo/ent"
	"github.com/qinzj/superpowers-demo/ent/enttest"
	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/infra/password"
	"github.com/qinzj/superpowers-demo/internal/router"
	"github.com/qinzj/superpowers-demo/internal/server/http/handler"
	"github.com/qinzj/superpowers-demo/internal/service/auth"
	"github.com/qinzj/superpowers-demo/internal/service/federation"
	"github.com/qinzj/superpowers-demo/internal/service/oidc"
	"github.com/qinzj/superpowers-demo/internal/service/user"
	"github.com/qinzj/superpowers-demo/internal/storage"
)

// testServer sets up an httptest server with full OIDC stack for integration tests.
// Uses in-memory SQLite, seeded OAuth2 client (sso-demo/secret), and OIDC routes.
func testServer(t *testing.T) (*httptest.Server, *ent.Client) {
	t.Helper()
	// NewEngine loads templates relative to cwd; ensure we run from module root.
	modRoot := findModuleRoot(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(modRoot); err != nil {
		t.Fatalf("chdir to module root: %v", err)
	}
	defer func() { _ = os.Chdir(orig) }()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")

	ctx := context.Background()
	if err := seedOAuth2Client(ctx, client); err != nil {
		t.Fatalf("seed OAuth2 client: %v", err)
	}

	issuer := "http://localhost:8888"
	oidcCfg, err := oidc.DefaultOIDCConfig(issuer)
	require.NoError(t, err)

	oidcStorage := oidc.NewFositeStorage(client)
	provider := oidc.NewOAuth2Provider(oidcCfg, oidcStorage)

	userRepo := storage.NewUserRepository(client)
	sessionRepo := storage.NewSessionRepository(client)
	idpConnRepo := storage.NewIdPConnectorRepository(client)
	userSvc := user.NewUserService(userRepo)
	authSvc := auth.NewAuthService(userRepo, sessionRepo)
	oidcAdapter := federation.NewOIDCClientAdapter()
	fedSvc := federation.NewFederationService(idpConnRepo, oidcAdapter, userRepo, authSvc)

	fedCfg := handler.FederationRouteConfig{
		Service: fedSvc,
		Issuer:  issuer,
	}

	engine := handler.NewEngine(nil)
	router.Setup(engine, &router.Config{
		OIDC: &handler.OIDCRouteConfig{
			Provider: provider,
			Issuer:   issuer,
			Auth:     authSvc,
		},
		Login: &handler.LoginRouteConfig{
			Auth:       authSvc,
			Federation: fedCfg,
		},
		Register: &handler.RegisterRouteConfig{
			UserService: userSvc,
		},
		Federation: &fedCfg,
	})

	srv := httptest.NewServer(engine)
	return srv, client
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func seedOAuth2Client(ctx context.Context, client *ent.Client) error {
	count, err := client.OAuth2Client.Query().Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	// Fosite expects bcrypt-hashed client secret for DefaultClient
	secretHash, err := password.Hash("secret")
	if err != nil {
		return err
	}
	_, err = client.OAuth2Client.Create().
		SetClientID("sso-demo").
		SetClientSecret(secretHash).
		SetRedirectUris([]string{"http://localhost:3000/callback"}).
		Save(ctx)
	return err
}

func TestOIDC_WellKnown(t *testing.T) {
	srv, db := testServer(t)
	defer srv.Close()
	defer db.Close()

	resp, err := srv.Client().Get(srv.URL + "/.well-known/openid-configuration")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "discovery endpoint should return 200")
	require.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	var doc map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&doc))

	require.Equal(t, "http://localhost:8888", doc["issuer"])
	require.Equal(t, "http://localhost:8888/authorize", doc["authorization_endpoint"])
	require.Equal(t, "http://localhost:8888/token", doc["token_endpoint"])
	require.Equal(t, "http://localhost:8888/userinfo", doc["userinfo_endpoint"])
	require.Equal(t, "http://localhost:8888/jwks.json", doc["jwks_uri"])

	scopes, ok := doc["scopes_supported"].([]interface{})
	require.True(t, ok)
	require.Contains(t, scopes, "openid")
	require.Contains(t, scopes, "profile")
}

func TestOIDC_Authorize_RedirectsToLogin(t *testing.T) {
	srv, db := testServer(t)
	defer srv.Close()
	defer db.Close()

	// No session; should redirect to /login with OAuth2 params preserved
	authURL := srv.URL + "/authorize?" + url.Values{
		"client_id":     []string{"sso-demo"},
		"redirect_uri": []string{"http://localhost:3000/callback"},
		"response_type": []string{"code"},
		"scope":         []string{"openid"},
		"state":         []string{"test-state-123"},
	}.Encode()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(authURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusFound, resp.StatusCode)
	loc := resp.Header.Get("Location")
	require.NotEmpty(t, loc, "should redirect")

	parsed, err := url.Parse(loc)
	require.NoError(t, err)
	require.Equal(t, "/login", parsed.Path)
	require.Equal(t, "sso-demo", parsed.Query().Get("client_id"))
	require.Equal(t, "http://localhost:3000/callback", parsed.Query().Get("redirect_uri"))
	require.Equal(t, "code", parsed.Query().Get("response_type"))
	require.Equal(t, "openid", parsed.Query().Get("scope"))
	require.Equal(t, "test-state-123", parsed.Query().Get("state"))
}

func TestOIDC_Token_ErrorResponseFormat(t *testing.T) {
	srv, db := testServer(t)
	defer srv.Close()
	defer db.Close()

	// POST /token with invalid request (missing grant_type) - verify JSON error format
	form := url.Values{}
	form.Set("client_id", "sso-demo")
	form.Set("client_secret", "secret")
	form.Set("redirect_uri", "http://localhost:3000/callback")
	// no grant_type, no code

	resp, err := srv.Client().Post(srv.URL+"/token", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Contains(t, body, "error")
	require.Contains(t, body, "error_description")
	require.NotEmpty(t, body["error"])
	require.NotEmpty(t, body["error_description"])
}

func TestOIDC_FullFlow_AuthorizationCode(t *testing.T) {
	srv, db := testServer(t)
	defer srv.Close()
	defer db.Close()

	ctx := context.Background()
	userRepo := storage.NewUserRepository(db)

	// Create test user
	hash, err := password.Hash("testpass123")
	require.NoError(t, err)
	u := &domain.User{
		Username:     "oidctest",
		Email:        "oidctest@example.com",
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}
	require.NoError(t, userRepo.Create(ctx, u))

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	jar := &testCookieJar{}
	client.Jar = jar

	// Step 1: POST /login to create session
	loginForm := url.Values{}
	loginForm.Set("username", "oidctest")
	loginForm.Set("password", "testpass123")
	loginForm.Set("client_id", "sso-demo")
	loginForm.Set("redirect_uri", "http://localhost:3000/callback")
	loginForm.Set("response_type", "code")
	loginForm.Set("scope", "openid")
	loginForm.Set("state", "flow-state")

	loginReq, err := http.NewRequest(http.MethodPost, srv.URL+"/login", strings.NewReader(loginForm.Encode()))
	require.NoError(t, err)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	loginResp, err := client.Do(loginReq)
	require.NoError(t, err)
	loginResp.Body.Close()
	require.Equal(t, http.StatusFound, loginResp.StatusCode, "login should redirect to /authorize")
	jar.Capture(loginResp)

	// Step 2: GET /authorize (with session cookie) - should redirect to client redirect_uri with code
	authURL := srv.URL + "/authorize?" + url.Values{
		"client_id":     []string{"sso-demo"},
		"redirect_uri": []string{"http://localhost:3000/callback"},
		"response_type": []string{"code"},
		"scope":         []string{"openid"},
		"state":         []string{"flow-state"},
	}.Encode()

	authReq, err := http.NewRequest(http.MethodGet, authURL, nil)
	require.NoError(t, err)
	jar.Inject(authReq)

	authResp, err := client.Do(authReq)
	require.NoError(t, err)
	authResp.Body.Close()

	require.True(t, authResp.StatusCode == http.StatusFound || authResp.StatusCode == http.StatusSeeOther,
		"expected 302 or 303 redirect, got %d", authResp.StatusCode)
	loc := authResp.Header.Get("Location")
	require.NotEmpty(t, loc)
	require.Contains(t, loc, "code=")
	require.Contains(t, loc, "state=flow-state")

	parsed, err := url.Parse(loc)
	require.NoError(t, err)
	code := parsed.Query().Get("code")
	require.NotEmpty(t, code)

	// Step 3: POST /token to exchange code for tokens (use Basic auth for client)
	tokenForm := url.Values{}
	tokenForm.Set("grant_type", "authorization_code")
	tokenForm.Set("code", code)
	tokenForm.Set("redirect_uri", "http://localhost:3000/callback")

	tokenReq, err := http.NewRequest(http.MethodPost, srv.URL+"/token", strings.NewReader(tokenForm.Encode()))
	require.NoError(t, err)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.SetBasicAuth("sso-demo", "secret")

	tokenResp, err := srv.Client().Do(tokenReq)
	require.NoError(t, err)
	defer tokenResp.Body.Close()

	var tokenBody map[string]interface{}
	require.NoError(t, json.NewDecoder(tokenResp.Body).Decode(&tokenBody))
	require.Equal(t, http.StatusOK, tokenResp.StatusCode, "token exchange failed: %+v", tokenBody)
	require.Contains(t, tokenResp.Header.Get("Content-Type"), "application/json")
	require.Contains(t, tokenBody, "access_token")
	require.NotEmpty(t, tokenBody["access_token"])
	require.Equal(t, "bearer", strings.ToLower(fmt.Sprint(tokenBody["token_type"])))
	require.Contains(t, tokenBody, "expires_in")
}

// testCookieJar captures Set-Cookie from responses and injects into requests.
type testCookieJar struct {
	cookies []*http.Cookie
	host    string
}

func (j *testCookieJar) Capture(resp *http.Response) {
	if u := resp.Request.URL; u != nil {
		j.host = u.Host
	}
	j.cookies = resp.Cookies()
}

func (j *testCookieJar) Inject(req *http.Request) {
	for _, c := range j.cookies {
		req.AddCookie(c)
	}
}

func (j *testCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.host = u.Host
	j.cookies = cookies
}

func (j *testCookieJar) Cookies(u *url.URL) []*http.Cookie {
	if u.Host != j.host {
		return nil
	}
	return j.cookies
}
