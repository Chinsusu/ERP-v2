package main

import (
	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func newRuntimeBatchCatalogStore(
	_ config.Config,
	auditLogStore audit.LogStore,
) (inventoryapp.BatchCatalogStore, func() error, error) {
	// S12-01-02 centralizes the runtime boundary; S12-01-03 promotes the DB path.
	return inventoryapp.NewPrototypeBatchCatalog(auditLogStore), nil, nil
}
