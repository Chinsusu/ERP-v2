package main

import (
	"database/sql"
	"strings"

	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeInboundQCInspectionStore(cfg config.Config) (qcapp.InboundQCInspectionStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return qcapp.NewPrototypeInboundQCInspectionStore(), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := qcapp.PostgresInboundQCInspectionStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return qcapp.NewPostgresInboundQCInspectionStore(db, storeConfig), db.Close, nil
}
