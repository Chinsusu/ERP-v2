package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type warehouseRouteHandlers struct {
	endOfDayReconciliations      routeHandler
	closeEndOfDayReconciliation  routeHandler
	dailyBoardFulfillmentMetrics routeHandler
	dailyBoardInboundMetrics     routeHandler
	dailyBoardSubcontractMetrics routeHandler
}

func registerWarehouseRoutes(routes routeGroup, handlers warehouseRouteHandlers) {
	routes.permission("/api/v1/warehouse/end-of-day-reconciliations", auth.PermissionWarehouseView, handlers.endOfDayReconciliations)
	routes.permission("/api/v1/warehouse/end-of-day-reconciliations/{reconciliation_id}/close", auth.PermissionRecordCreate, handlers.closeEndOfDayReconciliation)
	routes.permission("/api/v1/warehouse/daily-board/fulfillment-metrics", auth.PermissionWarehouseView, handlers.dailyBoardFulfillmentMetrics)
	routes.permission("/api/v1/warehouse/daily-board/inbound-metrics", auth.PermissionWarehouseView, handlers.dailyBoardInboundMetrics)
	routes.permission("/api/v1/warehouse/daily-board/subcontract-metrics", auth.PermissionWarehouseView, handlers.dailyBoardSubcontractMetrics)
}
