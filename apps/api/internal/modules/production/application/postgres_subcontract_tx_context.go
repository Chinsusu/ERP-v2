package application

import (
	"context"
	"database/sql"
)

type postgresSubcontractTxContextKey struct{}

func contextWithPostgresSubcontractTx(ctx context.Context, tx *sql.Tx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if tx == nil {
		return ctx
	}

	return context.WithValue(ctx, postgresSubcontractTxContextKey{}, tx)
}

func postgresSubcontractTxFromContext(ctx context.Context) (*sql.Tx, bool) {
	if ctx == nil {
		return nil, false
	}
	tx, ok := ctx.Value(postgresSubcontractTxContextKey{}).(*sql.Tx)

	return tx, ok && tx != nil
}
