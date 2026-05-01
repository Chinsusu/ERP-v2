package main

import (
	"database/sql"
	"strings"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeBatchCatalogStore(
	cfg config.Config,
	auditLogStore audit.LogStore,
) (inventoryapp.BatchCatalogStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return inventoryapp.NewPrototypeBatchCatalog(auditLogStore), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	return inventoryapp.NewPostgresBatchCatalogStore(db, auditLogStore), db.Close, nil
}
