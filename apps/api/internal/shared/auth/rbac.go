package auth

import (
	"net/http"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type RoleKey string

const RoleCEO RoleKey = "CEO"
const RoleERPAdmin RoleKey = "ERP_ADMIN"
const RoleWarehouseStaff RoleKey = "WAREHOUSE_STAFF"
const RoleWarehouseLead RoleKey = "WAREHOUSE_LEAD"
const RoleQA RoleKey = "QA"
const RoleSalesOps RoleKey = "SALES_OPS"
const RoleProductionOps RoleKey = "PRODUCTION_OPS"

type PermissionKey string

const PermissionDashboardView PermissionKey = "dashboard:view"
const PermissionWarehouseView PermissionKey = "warehouse:view"
const PermissionInventoryView PermissionKey = "inventory:view"
const PermissionPurchaseView PermissionKey = "purchase:view"
const PermissionQCView PermissionKey = "qc:view"
const PermissionProductionView PermissionKey = "production:view"
const PermissionSalesView PermissionKey = "sales:view"
const PermissionShippingView PermissionKey = "shipping:view"
const PermissionReturnsView PermissionKey = "returns:view"
const PermissionMasterDataView PermissionKey = "master-data:view"
const PermissionApprovalsView PermissionKey = "approvals:view"
const PermissionAuditLogView PermissionKey = "audit-log:view"
const PermissionReportsView PermissionKey = "reports:view"
const PermissionSettingsView PermissionKey = "settings:view"
const PermissionRecordCreate PermissionKey = "record:create"
const PermissionRecordExport PermissionKey = "record:export"

type RoleDefinition struct {
	Key         RoleKey
	Name        string
	Permissions []PermissionKey
}

var roleOrder = []RoleKey{
	RoleCEO,
	RoleERPAdmin,
	RoleWarehouseStaff,
	RoleWarehouseLead,
	RoleQA,
	RoleSalesOps,
	RoleProductionOps,
}

var roleDisplayNames = map[RoleKey]string{
	RoleCEO:            "CEO",
	RoleERPAdmin:       "ERP Admin",
	RoleWarehouseStaff: "Warehouse Staff",
	RoleWarehouseLead:  "Warehouse Lead",
	RoleQA:             "QA",
	RoleSalesOps:       "Sales Ops",
	RoleProductionOps:  "Production Ops",
}

var rolePermissions = map[RoleKey][]PermissionKey{
	RoleCEO: {
		PermissionDashboardView,
		PermissionApprovalsView,
		PermissionAuditLogView,
		PermissionReportsView,
		PermissionRecordExport,
	},
	RoleERPAdmin: {
		PermissionDashboardView,
		PermissionWarehouseView,
		PermissionInventoryView,
		PermissionPurchaseView,
		PermissionQCView,
		PermissionProductionView,
		PermissionSalesView,
		PermissionShippingView,
		PermissionReturnsView,
		PermissionMasterDataView,
		PermissionApprovalsView,
		PermissionAuditLogView,
		PermissionReportsView,
		PermissionSettingsView,
		PermissionRecordCreate,
		PermissionRecordExport,
	},
	RoleWarehouseStaff: {
		PermissionDashboardView,
		PermissionWarehouseView,
		PermissionInventoryView,
		PermissionShippingView,
		PermissionReturnsView,
	},
	RoleWarehouseLead: {
		PermissionDashboardView,
		PermissionWarehouseView,
		PermissionInventoryView,
		PermissionShippingView,
		PermissionReturnsView,
		PermissionApprovalsView,
		PermissionReportsView,
		PermissionRecordCreate,
		PermissionRecordExport,
	},
	RoleQA: {
		PermissionDashboardView,
		PermissionInventoryView,
		PermissionQCView,
		PermissionProductionView,
		PermissionReturnsView,
		PermissionReportsView,
		PermissionRecordCreate,
	},
	RoleSalesOps: {
		PermissionDashboardView,
		PermissionSalesView,
		PermissionShippingView,
		PermissionReturnsView,
		PermissionMasterDataView,
		PermissionReportsView,
		PermissionRecordCreate,
	},
	RoleProductionOps: {
		PermissionDashboardView,
		PermissionInventoryView,
		PermissionQCView,
		PermissionProductionView,
		PermissionReportsView,
		PermissionRecordCreate,
	},
}

func RoleCatalog() []RoleDefinition {
	roles := make([]RoleDefinition, 0, len(roleOrder))
	for _, role := range roleOrder {
		roles = append(roles, RoleDefinition{
			Key:         role,
			Name:        RoleDisplayName(role),
			Permissions: PermissionsForRole(role),
		})
	}

	return roles
}

func RoleDisplayName(role RoleKey) string {
	name, ok := roleDisplayNames[role]
	if !ok {
		return string(role)
	}

	return name
}

func PermissionsForRole(role RoleKey) []PermissionKey {
	permissions := rolePermissions[role]
	return append([]PermissionKey(nil), permissions...)
}

func HasPermission(principal Principal, permission PermissionKey) bool {
	for _, candidate := range principal.Permissions {
		if candidate == permission {
			return true
		}
	}

	return false
}

func RequireBearerPermission(cfg MockConfig, permission PermissionKey, next http.Handler) http.Handler {
	return RequireBearerToken(cfg, RequirePermission(permission, next))
}

func RequirePermission(permission PermissionKey, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(
				w,
				r,
				http.StatusUnauthorized,
				response.ErrorCodeUnauthorized,
				"Authentication required",
				nil,
			)
			return
		}

		if !HasPermission(principal, permission) {
			response.WriteError(
				w,
				r,
				http.StatusForbidden,
				response.ErrorCodeForbidden,
				"Permission denied",
				map[string]any{"permission": string(permission)},
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
