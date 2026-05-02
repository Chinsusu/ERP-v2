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
	_ config.Config,
	tasks ...shippingdomain.PickTask,
) (shippingapp.PickTaskStore, func() error, error) {
	return shippingapp.NewPrototypePickTaskStore(tasks...), nil, nil
}

func newRuntimePackTaskStore(
	_ config.Config,
	tasks ...shippingdomain.PackTask,
) (shippingapp.PackTaskStore, func() error, error) {
	return shippingapp.NewPrototypePackTaskStore(tasks...), nil, nil
}
