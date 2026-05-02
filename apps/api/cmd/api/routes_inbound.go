package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type inboundRouteHandlers struct {
	goodsReceipts            routeHandler
	goodsReceiptSubmit       routeHandler
	goodsReceiptInspectReady routeHandler
	goodsReceiptPost         routeHandler
	goodsReceiptDetail       routeHandler
	inboundQCInspections     routeHandler
	inboundQCStart           routeHandler
	inboundQCPass            routeHandler
	inboundQCFail            routeHandler
	inboundQCPartial         routeHandler
	inboundQCHold            routeHandler
	inboundQCDetail          routeHandler
	supplierRejections       routeHandler
	supplierRejectionSubmit  routeHandler
	supplierRejectionConfirm routeHandler
	supplierRejectionDetail  routeHandler
}

func registerInboundRoutes(routes routeGroup, handlers inboundRouteHandlers) {
	routes.token("/api/v1/goods-receipts", handlers.goodsReceipts)
	routes.permission("/api/v1/goods-receipts/{receipt_id}/submit", auth.PermissionRecordCreate, handlers.goodsReceiptSubmit)
	routes.permission("/api/v1/goods-receipts/{receipt_id}/inspect-ready", auth.PermissionRecordCreate, handlers.goodsReceiptInspectReady)
	routes.permission("/api/v1/goods-receipts/{receipt_id}/post", auth.PermissionRecordCreate, handlers.goodsReceiptPost)
	routes.permission("/api/v1/goods-receipts/{receipt_id}", auth.PermissionWarehouseView, handlers.goodsReceiptDetail)
	routes.token("/api/v1/inbound-qc-inspections", handlers.inboundQCInspections)
	routes.token("/api/v1/inbound-qc-inspections/{inspection_id}/start", handlers.inboundQCStart)
	routes.token("/api/v1/inbound-qc-inspections/{inspection_id}/pass", handlers.inboundQCPass)
	routes.token("/api/v1/inbound-qc-inspections/{inspection_id}/fail", handlers.inboundQCFail)
	routes.token("/api/v1/inbound-qc-inspections/{inspection_id}/partial", handlers.inboundQCPartial)
	routes.token("/api/v1/inbound-qc-inspections/{inspection_id}/hold", handlers.inboundQCHold)
	routes.token("/api/v1/inbound-qc-inspections/{inspection_id}", handlers.inboundQCDetail)
	routes.token("/api/v1/supplier-rejections", handlers.supplierRejections)
	routes.permission("/api/v1/supplier-rejections/{supplier_rejection_id}/submit", auth.PermissionRecordCreate, handlers.supplierRejectionSubmit)
	routes.permission("/api/v1/supplier-rejections/{supplier_rejection_id}/confirm", auth.PermissionRecordCreate, handlers.supplierRejectionConfirm)
	routes.permission("/api/v1/supplier-rejections/{supplier_rejection_id}", auth.PermissionWarehouseView, handlers.supplierRejectionDetail)
}
