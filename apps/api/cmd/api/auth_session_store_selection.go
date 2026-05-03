package main

import (
	"database/sql"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeSessionManager(
	cfg config.Config,
	now func() time.Time,
) (*auth.SessionManager, func() error, error) {
	authConfig := auth.MockConfig{
		Email:       cfg.AuthMockEmail,
		Password:    cfg.AuthMockPassword,
		AccessToken: cfg.StaticAuthAccessToken(),
		Users:       sprint22UATLoginUsers(cfg),
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return auth.NewSessionManager(authConfig, now), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	sessionConfig := auth.PostgresSessionStoreConfig{}
	loginFailureConfig := auth.PostgresLoginFailureStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		sessionConfig.DefaultOrgID = localAuditOrgID
		loginFailureConfig.DefaultOrgID = localAuditOrgID
	}

	manager, err := auth.NewSessionManagerWithStores(
		authConfig,
		now,
		auth.NewPostgresSessionStore(db, sessionConfig),
		auth.NewPostgresLoginFailureStore(db, loginFailureConfig),
	)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, nil, closeErr
		}
		return nil, nil, err
	}

	return manager, db.Close, nil
}

func sprint22UATLoginUsers(cfg config.Config) []auth.LoginUser {
	if !config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		return nil
	}

	return []auth.LoginUser{
		{
			Email: "warehouse_user@example.local",
			Name:  "Warehouse User",
			Role:  auth.RoleWarehouseStaff,
		},
		{
			Email: "sales_user@example.local",
			Name:  "Sales User",
			Role:  auth.RoleSalesOps,
		},
		{
			Email: "qc_user@example.local",
			Name:  "QC User",
			Role:  auth.RoleQA,
		},
	}
}
