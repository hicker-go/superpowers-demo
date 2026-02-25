/*
Copyright © 2026 qinzj
*/

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver for ent
	"github.com/qinzj/superpowers-demo/ent"
	"github.com/qinzj/superpowers-demo/internal/infra/password"
	"github.com/qinzj/superpowers-demo/internal/router"
	"github.com/qinzj/superpowers-demo/internal/server/http/handler"
	"github.com/qinzj/superpowers-demo/internal/service/auth"
	"github.com/qinzj/superpowers-demo/internal/service/federation"
	"github.com/qinzj/superpowers-demo/internal/service/oidc"
	"github.com/qinzj/superpowers-demo/internal/service/user"
	"github.com/qinzj/superpowers-demo/internal/storage"
	"github.com/qinzj/superpowers-demo/pkg/log"
)

// config keys
const (
	keyServerPort     = "server.port"
	keyDatabaseDriver = "database.driver"
	keyDatabaseDSN    = "database.dsn"
	keyOIDCIssuer     = "oidc.issuer"
)

func init() {
	rootCmd.AddCommand(svrCmd)
}

var svrCmd = &cobra.Command{
	Use:   "svr",
	Short: "Start the SSO OIDC server",
	Long:  `Start the HTTP server for SSO OIDC login and authentication.`,
	RunE:  runSvr,
}

func runSvr(cmd *cobra.Command, args []string) error {
	v := viper.New()
	v.SetConfigFile("configs/settings.yaml")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var logCfg log.Config
	if err := v.UnmarshalKey("log", &logCfg); err != nil {
		return fmt.Errorf("unmarshal log config: %w", err)
	}
	if logCfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(logCfg.File), 0o750); err != nil {
			return fmt.Errorf("create log dir: %w", err)
		}
	}
	logger, err := log.New(&logCfg)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer logger.Sync()

	port := v.GetInt(keyServerPort)
	if port == 0 {
		port = 8888
	}
	driver := v.GetString(keyDatabaseDriver)
	if driver == "" {
		driver = "sqlite3"
	}
	dsn := v.GetString(keyDatabaseDSN)
	if dsn == "" {
		return fmt.Errorf("database.dsn is required")
	}

	logger.Info("starting server", zap.Int(keyServerPort, port), zap.String(keyDatabaseDriver, driver), zap.String(keyDatabaseDSN, dsn))
	issuer := v.GetString(keyOIDCIssuer)
	if issuer == "" {
		issuer = fmt.Sprintf("http://localhost:%d", port)
	}

	// Ensure data dir exists for sqlite (dsn format: file:./data/sso.db?params)
	if dir := dataDirFromDSN(dsn); dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create data dir: %w", err)
		}
	}

	client, err := ent.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	if err := seedOAuth2Client(ctx, client); err != nil {
		return fmt.Errorf("seed OAuth2 client: %w", err)
	}

	oidcCfg, err := oidc.DefaultOIDCConfig(issuer)
	if err != nil {
		return fmt.Errorf("init OIDC config: %w", err)
	}

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

	engine := handler.NewEngine(logger)
	router.Setup(engine, &router.Config{
		Health: &handler.HealthRouteConfig{
			Client: client,
		},
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
		Account: &handler.AccountRouteConfig{
			UserService: userSvc,
			Auth:        authSvc,
		},
		Federation: &fedCfg,
	})

	addr := fmt.Sprintf(":%d", port)
	cmd.Printf("Starting server on http://localhost%s\n", addr)
	return engine.Run(addr)
}

func dataDirFromDSN(dsn string) string {
	const filePrefix = "file:"
	if !strings.HasPrefix(dsn, filePrefix) {
		return ""
	}
	pathPart := strings.SplitN(dsn[len(filePrefix):], "?", 2)[0]
	return filepath.Dir(pathPart)
}

// seedOAuth2Client inserts a development OAuth2 client if none exist.
// Client ID: sso-demo, secret: secret, redirect_uri: http://localhost:3000/callback
func seedOAuth2Client(ctx context.Context, client *ent.Client) error {
	count, err := client.OAuth2Client.Query().Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
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
