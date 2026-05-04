package main

import (
	"database/sql"
	"strings"

	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeProductionPlanStore(
	cfg config.Config,
	auditLogStore audit.LogStore,
) (productionapp.ProductionPlanStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return productionapp.NewPrototypeProductionPlanStore(auditLogStore), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := productionapp.PostgresProductionPlanStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return productionapp.NewPostgresProductionPlanStore(db, auditLogStore, storeConfig), db.Close, nil
}
