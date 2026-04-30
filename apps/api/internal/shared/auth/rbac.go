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
const RolePurchaseOps RoleKey = "PURCHASE_OPS"
const RoleFinanceOps RoleKey = "FINANCE_OPS"
const RoleSalesOps RoleKey = "SALES_OPS"
const RoleProductionOps RoleKey = "PRODUCTION_OPS"

type PermissionKey string

const PermissionDashboardView PermissionKey = "dashboard:view"
const PermissionWarehouseView PermissionKey = "warehouse:view"
const PermissionInventoryView PermissionKey = "inventory:view"
const PermissionPurchaseView PermissionKey = "purchase:view"
const PermissionFinanceView PermissionKey = "finance:view"
const PermissionFinanceManage PermissionKey = "finance:manage"
const PermissionCODReconcile PermissionKey = "cod:reconcile"
const PermissionPaymentApprove PermissionKey = "payment:approve"
const PermissionQCView PermissionKey = "qc:view"
const PermissionQCDecision PermissionKey = "qc:decision"
const PermissionProductionView PermissionKey = "production:view"
const PermissionSubcontractView PermissionKey = "subcontract:view"
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

type PermissionDefinition struct {
	Key   PermissionKey
	Name  string
	Group string
}

var roleOrder = []RoleKey{
	RoleCEO,
	RoleERPAdmin,
	RoleWarehouseStaff,
	RoleWarehouseLead,
	RoleQA,
	RolePurchaseOps,
	RoleFinanceOps,
	RoleSalesOps,
	RoleProductionOps,
}

var permissionCatalog = []PermissionDefinition{
	{Key: PermissionDashboardView, Name: "Dashboard", Group: "overview"},
	{Key: PermissionWarehouseView, Name: "Warehouse", Group: "operations"},
	{Key: PermissionInventoryView, Name: "Inventory", Group: "operations"},
	{Key: PermissionPurchaseView, Name: "Purchase", Group: "operations"},
	{Key: PermissionFinanceView, Name: "Finance", Group: "control"},
	{Key: PermissionFinanceManage, Name: "Finance Manage", Group: "control"},
	{Key: PermissionCODReconcile, Name: "COD Reconcile", Group: "action"},
	{Key: PermissionPaymentApprove, Name: "Payment Approve", Group: "action"},
	{Key: PermissionQCView, Name: "QC", Group: "operations"},
	{Key: PermissionQCDecision, Name: "QC Decision", Group: "action"},
	{Key: PermissionProductionView, Name: "Production", Group: "operations"},
	{Key: PermissionSubcontractView, Name: "Subcontract", Group: "operations"},
	{Key: PermissionSalesView, Name: "Sales", Group: "operations"},
	{Key: PermissionShippingView, Name: "Shipping", Group: "operations"},
	{Key: PermissionReturnsView, Name: "Returns", Group: "operations"},
	{Key: PermissionMasterDataView, Name: "Master Data", Group: "data"},
	{Key: PermissionApprovalsView, Name: "Approvals", Group: "control"},
	{Key: PermissionAuditLogView, Name: "Audit Log", Group: "control"},
	{Key: PermissionReportsView, Name: "Reports", Group: "control"},
	{Key: PermissionSettingsView, Name: "Settings", Group: "control"},
	{Key: PermissionRecordCreate, Name: "Create Record", Group: "action"},
	{Key: PermissionRecordExport, Name: "Export Record", Group: "action"},
}

var roleDisplayNames = map[RoleKey]string{
	RoleCEO:            "CEO",
	RoleERPAdmin:       "ERP Admin",
	RoleWarehouseStaff: "Warehouse Staff",
	RoleWarehouseLead:  "Warehouse Lead",
	RoleQA:             "QA",
	RolePurchaseOps:    "Purchase Ops",
	RoleFinanceOps:     "Finance Ops",
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
		PermissionFinanceView,
		PermissionFinanceManage,
		PermissionCODReconcile,
		PermissionPaymentApprove,
		PermissionQCView,
		PermissionQCDecision,
		PermissionProductionView,
		PermissionSubcontractView,
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
		PermissionQCDecision,
		PermissionProductionView,
		PermissionSubcontractView,
		PermissionReturnsView,
		PermissionReportsView,
		PermissionRecordCreate,
	},
	RolePurchaseOps: {
		PermissionDashboardView,
		PermissionPurchaseView,
		PermissionMasterDataView,
		PermissionReportsView,
		PermissionRecordCreate,
		PermissionRecordExport,
	},
	RoleFinanceOps: {
		PermissionDashboardView,
		PermissionPurchaseView,
		PermissionFinanceView,
		PermissionFinanceManage,
		PermissionCODReconcile,
		PermissionPaymentApprove,
		PermissionReportsView,
		PermissionAuditLogView,
		PermissionRecordExport,
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
		PermissionSubcontractView,
		PermissionReportsView,
		PermissionRecordCreate,
	},
}

func PermissionCatalog() []PermissionDefinition {
	permissions := make([]PermissionDefinition, len(permissionCatalog))
	copy(permissions, permissionCatalog)
	return permissions
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

func RequireSessionPermission(sessions *SessionManager, permission PermissionKey, next http.Handler) http.Handler {
	return RequireSessionToken(sessions, RequirePermission(permission, next))
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
