package main

import "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"

type platformRouteHandlers struct {
	health          routeHandler
	readiness       routeHandler
	login           routeHandler
	refresh         routeHandler
	logout          routeHandler
	policy          routeHandler
	me              routeHandler
	rbacRoles       routeHandler
	rbacPermissions routeHandler
	auditLogs       routeHandler
}

func registerPlatformRoutes(routes routeGroup, handlers platformRouteHandlers) {
	routes.public("/healthz", handlers.health)
	routes.public("/readyz", handlers.readiness)
	routes.public("/api/v1/health", handlers.health)
	routes.public("/api/v1/ready", handlers.readiness)
	routes.public("/api/v1/auth/login", handlers.login)
	routes.public("/api/v1/auth/mock-login", handlers.login)
	routes.public("/api/v1/auth/refresh", handlers.refresh)
	routes.public("/api/v1/auth/logout", handlers.logout)
	routes.public("/api/v1/auth/policy", handlers.policy)
	routes.token("/api/v1/me", handlers.me)
	routes.permission("/api/v1/rbac/roles", auth.PermissionSettingsView, handlers.rbacRoles)
	routes.permission("/api/v1/rbac/permissions", auth.PermissionSettingsView, handlers.rbacPermissions)
	routes.permission("/api/v1/audit-logs", auth.PermissionAuditLogView, handlers.auditLogs)
}
