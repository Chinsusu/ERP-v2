package main

type reportingRouteHandlers struct {
	inventorySnapshot    routeHandler
	inventorySnapshotCSV routeHandler
	operationsDaily      routeHandler
	operationsDailyCSV   routeHandler
	financeSummary       routeHandler
	financeSummaryCSV    routeHandler
}

func registerReportingRoutes(routes routeGroup, handlers reportingRouteHandlers) {
	routes.token("/api/v1/reports/inventory-snapshot", handlers.inventorySnapshot)
	routes.token("/api/v1/reports/inventory-snapshot/export.csv", handlers.inventorySnapshotCSV)
	routes.token("/api/v1/reports/operations-daily", handlers.operationsDaily)
	routes.token("/api/v1/reports/operations-daily/export.csv", handlers.operationsDailyCSV)
	routes.token("/api/v1/reports/finance-summary", handlers.financeSummary)
	routes.token("/api/v1/reports/finance-summary/export.csv", handlers.financeSummaryCSV)
}
