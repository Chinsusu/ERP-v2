package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type productionRouteHandlers struct {
	productionPlans               routeHandler
	productionPlanDetail          routeHandler
	productionPlanWarehouseIssues routeHandler
}

func registerProductionRoutes(routes routeGroup, handlers productionRouteHandlers) {
	routes.permission("/api/v1/production-plans", auth.PermissionProductionView, handlers.productionPlans)
	routes.permission("/api/v1/production-plans/{production_plan_id}", auth.PermissionProductionView, handlers.productionPlanDetail)
	routes.permission("/api/v1/production-plans/{production_plan_id}/warehouse-issues", auth.PermissionProductionView, handlers.productionPlanWarehouseIssues)
}
