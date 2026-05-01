package main

import (
	"database/sql"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const localAuditOrgID = "00000000-0000-4000-8000-000000000001"

func newRuntimeAuditLogStore(cfg config.Config) (audit.LogStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return audit.NewPrototypeLogStore(), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := audit.PostgresLogStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return audit.NewPostgresLogStore(db, storeConfig), db.Close, nil
}
