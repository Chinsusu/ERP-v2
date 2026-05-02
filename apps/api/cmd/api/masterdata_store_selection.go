package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type masterDataRuntimeStores struct {
	items      itemCatalog
	uoms       uomCatalog
	warehouses warehouseLocationCatalog
	parties    partyCatalog
}

type masterDataSeedStore interface {
	EnsureSeed(context.Context) error
}

func newRuntimeMasterDataStores(
	cfg config.Config,
	auditLogStore audit.LogStore,
) (masterDataRuntimeStores, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return masterDataRuntimeStores{
			items:      masterdataapp.NewPrototypeItemCatalog(auditLogStore),
			uoms:       masterdataapp.NewPrototypeUOMCatalog(),
			warehouses: masterdataapp.NewPrototypeWarehouseLocationCatalog(auditLogStore),
			parties:    masterdataapp.NewPrototypePartyCatalog(auditLogStore),
		}, nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return masterDataRuntimeStores{}, nil, err
	}

	itemConfig := masterdataapp.PostgresItemCatalogConfig{}
	warehouseConfig := masterdataapp.PostgresWarehouseLocationCatalogConfig{}
	partyConfig := masterdataapp.PostgresPartyCatalogConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		itemConfig.DefaultOrgID = localAuditOrgID
		warehouseConfig.DefaultOrgID = localAuditOrgID
		partyConfig.DefaultOrgID = localAuditOrgID
	}

	return masterDataRuntimeStores{
		items:      masterdataapp.NewPostgresItemCatalog(db, auditLogStore, itemConfig),
		uoms:       masterdataapp.NewPostgresUOMCatalog(db),
		warehouses: masterdataapp.NewPostgresWarehouseLocationCatalog(db, auditLogStore, warehouseConfig),
		parties:    masterdataapp.NewPostgresPartyCatalog(db, auditLogStore, partyConfig),
	}, db.Close, nil
}

func seedRuntimeMasterDataStores(ctx context.Context, stores masterDataRuntimeStores) error {
	for _, entry := range []struct {
		name  string
		store any
	}{
		{name: "uom catalog", store: stores.uoms},
		{name: "item catalog", store: stores.items},
		{name: "warehouse catalog", store: stores.warehouses},
		{name: "party catalog", store: stores.parties},
	} {
		seeder, ok := entry.store.(masterDataSeedStore)
		if !ok {
			continue
		}
		if err := seeder.EnsureSeed(ctx); err != nil {
			return fmt.Errorf("seed %s: %w", entry.name, err)
		}
	}

	return nil
}
