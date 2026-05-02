package main

import (
	"database/sql"
	"strings"

	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeCarrierManifestStore(
	cfg config.Config,
) (shippingapp.CarrierManifestStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return shippingapp.NewPrototypeCarrierManifestStore(), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := shippingapp.PostgresCarrierManifestStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return shippingapp.NewPostgresCarrierManifestStore(db, storeConfig), db.Close, nil
}

func newRuntimePickTaskStore(
	cfg config.Config,
	tasks ...shippingdomain.PickTask,
) (shippingapp.PickTaskStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return shippingapp.NewPrototypePickTaskStore(tasks...), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := shippingapp.PostgresPickTaskStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return shippingapp.NewPostgresPickTaskStore(db, storeConfig), db.Close, nil
}

func newRuntimePackTaskStore(
	cfg config.Config,
	tasks ...shippingdomain.PackTask,
) (shippingapp.PackTaskStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return shippingapp.NewPrototypePackTaskStore(tasks...), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := shippingapp.PostgresPackTaskStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return shippingapp.NewPostgresPackTaskStore(db, storeConfig), db.Close, nil
}
