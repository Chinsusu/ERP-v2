package main

import (
	"database/sql"
	"strings"

	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type runtimeReturnReceiptStore interface {
	returnsapp.ReturnReceiptStore
	returnsapp.ReturnInspectionStore
	returnsapp.ReturnDispositionStore
	returnsapp.ReturnAttachmentStore
}

func newRuntimeReturnReceiptStore(cfg config.Config) (runtimeReturnReceiptStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return returnsapp.NewPrototypeReturnReceiptStore(), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := returnsapp.PostgresReturnReceiptStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return returnsapp.NewPostgresReturnReceiptStore(db, storeConfig), db.Close, nil
}
