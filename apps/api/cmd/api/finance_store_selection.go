package main

import (
	"database/sql"
	"strings"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type financeRuntimeStores struct {
	customerReceivables financeapp.CustomerReceivableStore
	supplierPayables    financeapp.SupplierPayableStore
	supplierInvoices    financeapp.SupplierInvoiceStore
	codRemittances      financeapp.CODRemittanceStore
	cashTransactions    financeapp.CashTransactionStore
}

func newRuntimeFinanceStores(cfg config.Config) (financeRuntimeStores, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return financeRuntimeStores{
			customerReceivables: financeapp.NewPrototypeCustomerReceivableStore(),
			supplierPayables:    financeapp.NewPrototypeSupplierPayableStore(),
			supplierInvoices:    financeapp.NewPrototypeSupplierInvoiceStore(),
			codRemittances:      financeapp.NewPrototypeCODRemittanceStore(),
			cashTransactions:    financeapp.NewPrototypeCashTransactionStore(),
		}, nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return financeRuntimeStores{}, nil, err
	}

	customerReceivableConfig := financeapp.PostgresCustomerReceivableStoreConfig{}
	supplierPayableConfig := financeapp.PostgresSupplierPayableStoreConfig{}
	supplierInvoiceConfig := financeapp.PostgresSupplierInvoiceStoreConfig{}
	codRemittanceConfig := financeapp.PostgresCODRemittanceStoreConfig{}
	cashTransactionConfig := financeapp.PostgresCashTransactionStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		customerReceivableConfig.DefaultOrgID = localAuditOrgID
		supplierPayableConfig.DefaultOrgID = localAuditOrgID
		supplierInvoiceConfig.DefaultOrgID = localAuditOrgID
		codRemittanceConfig.DefaultOrgID = localAuditOrgID
		cashTransactionConfig.DefaultOrgID = localAuditOrgID
	}

	return financeRuntimeStores{
		customerReceivables: financeapp.NewPostgresCustomerReceivableStore(db, customerReceivableConfig),
		supplierPayables:    financeapp.NewPostgresSupplierPayableStore(db, supplierPayableConfig),
		supplierInvoices:    financeapp.NewPostgresSupplierInvoiceStore(db, supplierInvoiceConfig),
		codRemittances:      financeapp.NewPostgresCODRemittanceStore(db, codRemittanceConfig),
		cashTransactions:    financeapp.NewPostgresCashTransactionStore(db, cashTransactionConfig),
	}, db.Close, nil
}
