package main

import (
	"database/sql"
	"strings"

	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type subcontractRuntimeStores struct {
	orders                productionapp.SubcontractOrderStore
	materialTransfers     productionapp.SubcontractMaterialTransferStore
	sampleApprovals       productionapp.SubcontractSampleApprovalStore
	finishedGoodsReceipts productionapp.SubcontractFinishedGoodsReceiptStore
	factoryClaims         productionapp.SubcontractFactoryClaimStore
	paymentMilestones     productionapp.SubcontractPaymentMilestoneStore
	factoryDispatches     productionapp.SubcontractFactoryDispatchStore
}

func newRuntimeSubcontractStores(
	cfg config.Config,
	auditLogStore audit.LogStore,
) (subcontractRuntimeStores, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return subcontractRuntimeStores{
			orders:                productionapp.NewPrototypeSubcontractOrderStore(auditLogStore),
			materialTransfers:     productionapp.NewPrototypeSubcontractMaterialTransferStore(),
			sampleApprovals:       productionapp.NewPrototypeSubcontractSampleApprovalStore(),
			finishedGoodsReceipts: productionapp.NewPrototypeSubcontractFinishedGoodsReceiptStore(),
			factoryClaims:         productionapp.NewPrototypeSubcontractFactoryClaimStore(),
			paymentMilestones:     productionapp.NewPrototypeSubcontractPaymentMilestoneStore(),
			factoryDispatches:     productionapp.NewPrototypeSubcontractFactoryDispatchStore(),
		}, nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return subcontractRuntimeStores{}, nil, err
	}

	orderConfig := productionapp.PostgresSubcontractOrderStoreConfig{}
	materialTransferConfig := productionapp.PostgresSubcontractMaterialTransferStoreConfig{}
	sampleApprovalConfig := productionapp.PostgresSubcontractSampleApprovalStoreConfig{}
	finishedGoodsReceiptConfig := productionapp.PostgresSubcontractFinishedGoodsReceiptStoreConfig{}
	factoryClaimConfig := productionapp.PostgresSubcontractFactoryClaimStoreConfig{}
	paymentMilestoneConfig := productionapp.PostgresSubcontractPaymentMilestoneStoreConfig{}
	factoryDispatchConfig := productionapp.PostgresSubcontractFactoryDispatchStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		orderConfig.DefaultOrgID = localAuditOrgID
		materialTransferConfig.DefaultOrgID = localAuditOrgID
		sampleApprovalConfig.DefaultOrgID = localAuditOrgID
		finishedGoodsReceiptConfig.DefaultOrgID = localAuditOrgID
		factoryClaimConfig.DefaultOrgID = localAuditOrgID
		paymentMilestoneConfig.DefaultOrgID = localAuditOrgID
		factoryDispatchConfig.DefaultOrgID = localAuditOrgID
	}

	return subcontractRuntimeStores{
		orders: productionapp.NewPostgresSubcontractOrderStore(
			db,
			orderConfig,
		),
		materialTransfers: productionapp.NewPostgresSubcontractMaterialTransferStore(
			db,
			materialTransferConfig,
		),
		sampleApprovals: productionapp.NewPostgresSubcontractSampleApprovalStore(
			db,
			sampleApprovalConfig,
		),
		finishedGoodsReceipts: productionapp.NewPostgresSubcontractFinishedGoodsReceiptStore(
			db,
			finishedGoodsReceiptConfig,
		),
		factoryClaims: productionapp.NewPostgresSubcontractFactoryClaimStore(
			db,
			factoryClaimConfig,
		),
		paymentMilestones: productionapp.NewPostgresSubcontractPaymentMilestoneStore(
			db,
			paymentMilestoneConfig,
		),
		factoryDispatches: productionapp.NewPostgresSubcontractFactoryDispatchStore(
			db,
			factoryDispatchConfig,
		),
	}, db.Close, nil
}
