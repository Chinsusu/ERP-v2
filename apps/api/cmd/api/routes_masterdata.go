package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type masterDataRouteHandlers struct {
	products                routeHandler
	productDetail           routeHandler
	productStatus           routeHandler
	warehouses              routeHandler
	warehouseDetail         routeHandler
	warehouseStatus         routeHandler
	warehouseLocations      routeHandler
	warehouseLocationDetail routeHandler
	warehouseLocationStatus routeHandler
	suppliers               routeHandler
	supplierDetail          routeHandler
	supplierStatus          routeHandler
	customers               routeHandler
	customerDetail          routeHandler
	customerStatus          routeHandler
}

func registerMasterDataRoutes(routes routeGroup, handlers masterDataRouteHandlers) {
	routes.token("/api/v1/products", handlers.products)
	routes.token("/api/v1/products/{product_id}", handlers.productDetail)
	routes.permission("/api/v1/products/{product_id}/status", auth.PermissionRecordCreate, handlers.productStatus)
	routes.token("/api/v1/warehouses", handlers.warehouses)
	routes.token("/api/v1/warehouses/{warehouse_id}", handlers.warehouseDetail)
	routes.permission("/api/v1/warehouses/{warehouse_id}/status", auth.PermissionRecordCreate, handlers.warehouseStatus)
	routes.token("/api/v1/warehouse-locations", handlers.warehouseLocations)
	routes.token("/api/v1/warehouse-locations/{location_id}", handlers.warehouseLocationDetail)
	routes.permission("/api/v1/warehouse-locations/{location_id}/status", auth.PermissionRecordCreate, handlers.warehouseLocationStatus)
	routes.token("/api/v1/suppliers", handlers.suppliers)
	routes.token("/api/v1/suppliers/{supplier_id}", handlers.supplierDetail)
	routes.permission("/api/v1/suppliers/{supplier_id}/status", auth.PermissionRecordCreate, handlers.supplierStatus)
	routes.token("/api/v1/customers", handlers.customers)
	routes.token("/api/v1/customers/{customer_id}", handlers.customerDetail)
	routes.permission("/api/v1/customers/{customer_id}/status", auth.PermissionRecordCreate, handlers.customerStatus)
}
