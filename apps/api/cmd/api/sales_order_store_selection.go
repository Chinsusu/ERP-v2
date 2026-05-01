package main

import (
	"database/sql"
	"strings"

	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeSalesOrderStore(
	cfg config.Config,
	auditLogStore audit.LogStore,
) (salesapp.SalesOrderStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return salesapp.NewPrototypeSalesOrderStore(auditLogStore), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := salesapp.PostgresSalesOrderStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return salesapp.NewPostgresSalesOrderStore(db, storeConfig), db.Close, nil
}
