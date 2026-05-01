package main

import (
	"context"
	"database/sql"
	"strings"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeStockMovementStore(cfg config.Config) (inventoryapp.StockMovementStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return inventoryapp.NewInMemoryStockMovementStore(), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	return runtimeStockMovementStore{
		postgres: inventoryapp.NewPostgresStockMovementStore(db),
		memory:   inventoryapp.NewInMemoryStockMovementStore(),
	}, db.Close, nil
}

type runtimeStockMovementStore struct {
	postgres inventoryapp.StockMovementStore
	memory   inventoryapp.StockMovementStore
}

func (s runtimeStockMovementStore) Record(ctx context.Context, movement inventorydomain.StockMovement) error {
	if postgresCompatibleStockMovement(movement) {
		return s.postgres.Record(ctx, movement)
	}

	return s.memory.Record(ctx, movement)
}

func postgresCompatibleStockMovement(movement inventorydomain.StockMovement) bool {
	return isUUID(movement.OrgID) &&
		isUUID(movement.ItemID) &&
		isUUID(movement.WarehouseID) &&
		isUUID(movement.SourceDocID)
}

func isUUID(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 36 {
		return false
	}
	for index, char := range value {
		switch index {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !isHex(char) {
				return false
			}
		}
	}

	return true
}

func isHex(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}
