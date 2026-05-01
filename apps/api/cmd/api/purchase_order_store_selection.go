package main

import (
	"database/sql"
	"strings"

	purchaseapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimePurchaseOrderStore(
	cfg config.Config,
	auditLogStore audit.LogStore,
) (purchaseapp.PurchaseOrderStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return purchaseapp.NewPrototypePurchaseOrderStore(auditLogStore), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := purchaseapp.PostgresPurchaseOrderStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return purchaseapp.NewPostgresPurchaseOrderStore(db, storeConfig), db.Close, nil
}
