package main

import (
	"database/sql"
	"strings"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeEndOfDayReconciliationStore(
	cfg config.Config,
) (inventoryapp.EndOfDayReconciliationStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return inventoryapp.NewPrototypeEndOfDayReconciliationStore(), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := inventoryapp.PostgresEndOfDayReconciliationStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return inventoryapp.NewPostgresEndOfDayReconciliationStore(db, storeConfig), db.Close, nil
}
