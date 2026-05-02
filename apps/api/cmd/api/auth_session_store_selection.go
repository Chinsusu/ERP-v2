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
