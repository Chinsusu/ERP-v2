package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	purchaseapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/application"
	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	qcdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/domain"
	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	returnsdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/storage"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken      string       `json:"access_token"`
	RefreshToken     string       `json:"refresh_token"`
	TokenType        string       `json:"token_type"`
	ExpiresIn        int          `json:"expires_in"`
	RefreshExpiresIn int          `json:"refresh_expires_in"`
	ExpiresAt        string       `json:"expires_at"`
	User             userResponse `json:"user"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authPolicyResponse struct {
	PasswordMinLength              int  `json:"password_min_length"`
	PasswordRequiresLetter         bool `json:"password_requires_letter"`
	PasswordRequiresNumberOrSymbol bool `json:"password_requires_number_or_symbol"`
	CommonPasswordsBlocked         bool `json:"common_passwords_blocked"`
	MaxFailedAttempts              int  `json:"max_failed_attempts"`
	LockoutWindowSeconds           int  `json:"lockout_window_seconds"`
	LockoutDurationSeconds         int  `json:"lockout_duration_seconds"`
}

type userResponse struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type roleResponse struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type permissionResponse struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

type productResponse struct {
	ID               string `json:"id"`
	ItemCode         string `json:"item_code"`
	SKUCode          string `json:"sku_code"`
	Name             string `json:"name"`
	ItemType         string `json:"item_type"`
	ItemGroup        string `json:"item_group,omitempty"`
	BrandCode        string `json:"brand_code,omitempty"`
	UOMCode          string `json:"uom_code"`
	UOMBase          string `json:"uom_base"`
	UOMPurchase      string `json:"uom_purchase,omitempty"`
	UOMIssue         string `json:"uom_issue,omitempty"`
	LotControlled    bool   `json:"lot_controlled"`
	ExpiryControlled bool   `json:"expiry_controlled"`
	ShelfLifeDays    int    `json:"shelf_life_days,omitempty"`
	QCRequired       bool   `json:"qc_required"`
	Status           string `json:"status"`
	StandardCost     string `json:"standard_cost,omitempty"`
	IsSellable       bool   `json:"is_sellable"`
	IsPurchasable    bool   `json:"is_purchasable"`
	IsProducible     bool   `json:"is_producible"`
	SpecVersion      string `json:"spec_version,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	AuditLogID       string `json:"audit_log_id,omitempty"`
}

type productRequest struct {
	ItemCode         string `json:"item_code"`
	SKUCode          string `json:"sku_code"`
	Name             string `json:"name"`
	ItemType         string `json:"item_type"`
	ItemGroup        string `json:"item_group"`
	BrandCode        string `json:"brand_code"`
	UOMBase          string `json:"uom_base"`
	UOMPurchase      string `json:"uom_purchase"`
	UOMIssue         string `json:"uom_issue"`
	LotControlled    bool   `json:"lot_controlled"`
	ExpiryControlled bool   `json:"expiry_controlled"`
	ShelfLifeDays    int    `json:"shelf_life_days"`
	QCRequired       bool   `json:"qc_required"`
	Status           string `json:"status"`
	StandardCost     string `json:"standard_cost"`
	IsSellable       bool   `json:"is_sellable"`
	IsPurchasable    bool   `json:"is_purchasable"`
	IsProducible     bool   `json:"is_producible"`
	SpecVersion      string `json:"spec_version"`
}

type changeProductStatusRequest struct {
	Status string `json:"status"`
}

type warehouseResponse struct {
	ID              string `json:"id"`
	WarehouseCode   string `json:"warehouse_code"`
	WarehouseName   string `json:"warehouse_name"`
	WarehouseType   string `json:"warehouse_type"`
	SiteCode        string `json:"site_code"`
	Address         string `json:"address,omitempty"`
	AllowSaleIssue  bool   `json:"allow_sale_issue"`
	AllowProdIssue  bool   `json:"allow_prod_issue"`
	AllowQuarantine bool   `json:"allow_quarantine"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	AuditLogID      string `json:"audit_log_id,omitempty"`
}

type warehouseRequest struct {
	WarehouseCode   string `json:"warehouse_code"`
	WarehouseName   string `json:"warehouse_name"`
	WarehouseType   string `json:"warehouse_type"`
	SiteCode        string `json:"site_code"`
	Address         string `json:"address"`
	AllowSaleIssue  bool   `json:"allow_sale_issue"`
	AllowProdIssue  bool   `json:"allow_prod_issue"`
	AllowQuarantine bool   `json:"allow_quarantine"`
	Status          string `json:"status"`
}

type changeWarehouseStatusRequest struct {
	Status string `json:"status"`
}

type warehouseLocationResponse struct {
	ID            string `json:"id"`
	WarehouseID   string `json:"warehouse_id"`
	WarehouseCode string `json:"warehouse_code"`
	LocationCode  string `json:"location_code"`
	LocationName  string `json:"location_name"`
	LocationType  string `json:"location_type"`
	ZoneCode      string `json:"zone_code,omitempty"`
	AllowReceive  bool   `json:"allow_receive"`
	AllowPick     bool   `json:"allow_pick"`
	AllowStore    bool   `json:"allow_store"`
	IsDefault     bool   `json:"is_default"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	AuditLogID    string `json:"audit_log_id,omitempty"`
}

type warehouseLocationRequest struct {
	WarehouseID  string `json:"warehouse_id"`
	LocationCode string `json:"location_code"`
	LocationName string `json:"location_name"`
	LocationType string `json:"location_type"`
	ZoneCode     string `json:"zone_code"`
	AllowReceive bool   `json:"allow_receive"`
	AllowPick    bool   `json:"allow_pick"`
	AllowStore   bool   `json:"allow_store"`
	IsDefault    bool   `json:"is_default"`
	Status       string `json:"status"`
}

type changeWarehouseLocationStatusRequest struct {
	Status string `json:"status"`
}

type supplierResponse struct {
	ID            string `json:"id"`
	SupplierCode  string `json:"supplier_code"`
	SupplierName  string `json:"supplier_name"`
	SupplierGroup string `json:"supplier_group"`
	ContactName   string `json:"contact_name,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Email         string `json:"email,omitempty"`
	TaxCode       string `json:"tax_code,omitempty"`
	Address       string `json:"address,omitempty"`
	PaymentTerms  string `json:"payment_terms,omitempty"`
	LeadTimeDays  int    `json:"lead_time_days,omitempty"`
	MOQ           string `json:"moq,omitempty"`
	QualityScore  string `json:"quality_score,omitempty"`
	DeliveryScore string `json:"delivery_score,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	AuditLogID    string `json:"audit_log_id,omitempty"`
}

type supplierRequest struct {
	SupplierCode  string `json:"supplier_code"`
	SupplierName  string `json:"supplier_name"`
	SupplierGroup string `json:"supplier_group"`
	ContactName   string `json:"contact_name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	TaxCode       string `json:"tax_code"`
	Address       string `json:"address"`
	PaymentTerms  string `json:"payment_terms"`
	LeadTimeDays  int    `json:"lead_time_days"`
	MOQ           string `json:"moq"`
	QualityScore  string `json:"quality_score"`
	DeliveryScore string `json:"delivery_score"`
	Status        string `json:"status"`
}

type changeSupplierStatusRequest struct {
	Status string `json:"status"`
}

type customerResponse struct {
	ID            string `json:"id"`
	CustomerCode  string `json:"customer_code"`
	CustomerName  string `json:"customer_name"`
	CustomerType  string `json:"customer_type"`
	ChannelCode   string `json:"channel_code,omitempty"`
	PriceListCode string `json:"price_list_code,omitempty"`
	DiscountGroup string `json:"discount_group,omitempty"`
	CreditLimit   string `json:"credit_limit,omitempty"`
	PaymentTerms  string `json:"payment_terms,omitempty"`
	ContactName   string `json:"contact_name,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Email         string `json:"email,omitempty"`
	TaxCode       string `json:"tax_code,omitempty"`
	Address       string `json:"address,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	AuditLogID    string `json:"audit_log_id,omitempty"`
}

type customerRequest struct {
	CustomerCode  string `json:"customer_code"`
	CustomerName  string `json:"customer_name"`
	CustomerType  string `json:"customer_type"`
	ChannelCode   string `json:"channel_code"`
	PriceListCode string `json:"price_list_code"`
	DiscountGroup string `json:"discount_group"`
	CreditLimit   string `json:"credit_limit"`
	PaymentTerms  string `json:"payment_terms"`
	ContactName   string `json:"contact_name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	TaxCode       string `json:"tax_code"`
	Address       string `json:"address"`
	Status        string `json:"status"`
}

type changeCustomerStatusRequest struct {
	Status string `json:"status"`
}

type salesOrderLineRequest struct {
	ID                 string `json:"id"`
	LineNo             int    `json:"line_no"`
	ItemID             string `json:"item_id"`
	OrderedQty         string `json:"ordered_qty"`
	UOMCode            string `json:"uom_code"`
	UnitPrice          string `json:"unit_price"`
	CurrencyCode       string `json:"currency_code"`
	LineDiscountAmount string `json:"line_discount_amount"`
	BatchID            string `json:"batch_id"`
	BatchNo            string `json:"batch_no"`
}

type createSalesOrderRequest struct {
	ID           string                  `json:"id"`
	OrderNo      string                  `json:"order_no"`
	CustomerID   string                  `json:"customer_id"`
	Channel      string                  `json:"channel"`
	WarehouseID  string                  `json:"warehouse_id"`
	OrderDate    string                  `json:"order_date"`
	CurrencyCode string                  `json:"currency_code"`
	Note         string                  `json:"note"`
	Lines        []salesOrderLineRequest `json:"lines"`
}

type updateSalesOrderRequest struct {
	CustomerID      string                  `json:"customer_id"`
	Channel         string                  `json:"channel"`
	WarehouseID     string                  `json:"warehouse_id"`
	OrderDate       string                  `json:"order_date"`
	Note            string                  `json:"note"`
	ExpectedVersion int                     `json:"expected_version"`
	Lines           []salesOrderLineRequest `json:"lines"`
}

type salesOrderActionRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	Reason          string `json:"reason"`
	Note            string `json:"note"`
}

type salesOrderLineResponse struct {
	ID                 string `json:"id"`
	LineNo             int    `json:"line_no"`
	ItemID             string `json:"item_id"`
	SKUCode            string `json:"sku_code"`
	ItemName           string `json:"item_name"`
	OrderedQty         string `json:"ordered_qty"`
	UOMCode            string `json:"uom_code"`
	BaseOrderedQty     string `json:"base_ordered_qty"`
	BaseUOMCode        string `json:"base_uom_code"`
	ConversionFactor   string `json:"conversion_factor"`
	UnitPrice          string `json:"unit_price"`
	CurrencyCode       string `json:"currency_code"`
	LineDiscountAmount string `json:"line_discount_amount"`
	LineAmount         string `json:"line_amount"`
	ReservedQty        string `json:"reserved_qty"`
	ShippedQty         string `json:"shipped_qty"`
	BatchID            string `json:"batch_id,omitempty"`
	BatchNo            string `json:"batch_no,omitempty"`
}

type salesOrderListItemResponse struct {
	ID                string `json:"id"`
	OrderNo           string `json:"order_no"`
	CustomerID        string `json:"customer_id"`
	CustomerCode      string `json:"customer_code,omitempty"`
	CustomerName      string `json:"customer_name"`
	Channel           string `json:"channel"`
	WarehouseID       string `json:"warehouse_id,omitempty"`
	WarehouseCode     string `json:"warehouse_code,omitempty"`
	OrderDate         string `json:"order_date"`
	Status            string `json:"status"`
	CurrencyCode      string `json:"currency_code"`
	TotalAmount       string `json:"total_amount"`
	LineCount         int    `json:"line_count"`
	ReservedLineCount int    `json:"reserved_line_count"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	Version           int    `json:"version"`
}

type salesOrderResponse struct {
	ID                string                   `json:"id"`
	OrderNo           string                   `json:"order_no"`
	CustomerID        string                   `json:"customer_id"`
	CustomerCode      string                   `json:"customer_code,omitempty"`
	CustomerName      string                   `json:"customer_name"`
	Channel           string                   `json:"channel"`
	WarehouseID       string                   `json:"warehouse_id,omitempty"`
	WarehouseCode     string                   `json:"warehouse_code,omitempty"`
	OrderDate         string                   `json:"order_date"`
	Status            string                   `json:"status"`
	CurrencyCode      string                   `json:"currency_code"`
	SubtotalAmount    string                   `json:"subtotal_amount"`
	DiscountAmount    string                   `json:"discount_amount"`
	TaxAmount         string                   `json:"tax_amount"`
	ShippingFeeAmount string                   `json:"shipping_fee_amount"`
	NetAmount         string                   `json:"net_amount"`
	TotalAmount       string                   `json:"total_amount"`
	Note              string                   `json:"note,omitempty"`
	Lines             []salesOrderLineResponse `json:"lines"`
	AuditLogID        string                   `json:"audit_log_id,omitempty"`
	CreatedAt         string                   `json:"created_at"`
	UpdatedAt         string                   `json:"updated_at"`
	ConfirmedAt       string                   `json:"confirmed_at,omitempty"`
	CancelledAt       string                   `json:"cancelled_at,omitempty"`
	CancelReason      string                   `json:"cancel_reason,omitempty"`
	Version           int                      `json:"version"`
}

type salesOrderActionResultResponse struct {
	SalesOrder     salesOrderResponse `json:"sales_order"`
	PreviousStatus string             `json:"previous_status"`
	CurrentStatus  string             `json:"current_status"`
	AuditLogID     string             `json:"audit_log_id,omitempty"`
}

type warehouseFulfillmentMetricsResponse struct {
	WarehouseID           string `json:"warehouse_id,omitempty"`
	Date                  string `json:"date,omitempty"`
	ShiftCode             string `json:"shift_code,omitempty"`
	CarrierCode           string `json:"carrier_code,omitempty"`
	TotalOrders           int    `json:"total_orders"`
	NewOrders             int    `json:"new_orders"`
	ReservedOrders        int    `json:"reserved_orders"`
	PickingOrders         int    `json:"picking_orders"`
	PackedOrders          int    `json:"packed_orders"`
	WaitingHandoverOrders int    `json:"waiting_handover_orders"`
	MissingOrders         int    `json:"missing_orders"`
	HandoverOrders        int    `json:"handover_orders"`
	GeneratedAt           string `json:"generated_at"`
}

type warehouseInboundMetricsResponse struct {
	WarehouseID                string `json:"warehouse_id,omitempty"`
	Date                       string `json:"date,omitempty"`
	ShiftCode                  string `json:"shift_code,omitempty"`
	PurchaseOrdersIncoming     int    `json:"purchase_orders_incoming"`
	ReceivingPending           int    `json:"receiving_pending"`
	ReceivingDraft             int    `json:"receiving_draft"`
	ReceivingSubmitted         int    `json:"receiving_submitted"`
	ReceivingInspectReady      int    `json:"receiving_inspect_ready"`
	QCHold                     int    `json:"qc_hold"`
	QCFail                     int    `json:"qc_fail"`
	QCPass                     int    `json:"qc_pass"`
	QCPartial                  int    `json:"qc_partial"`
	SupplierRejections         int    `json:"supplier_rejections"`
	SupplierRejectionDraft     int    `json:"supplier_rejection_draft"`
	SupplierRejectionSubmitted int    `json:"supplier_rejection_submitted"`
	SupplierRejectionConfirmed int    `json:"supplier_rejection_confirmed"`
	SupplierRejectionCancelled int    `json:"supplier_rejection_cancelled"`
	GeneratedAt                string `json:"generated_at"`
}

type warehouseSubcontractMetricsResponse struct {
	WarehouseID             string `json:"warehouse_id,omitempty"`
	Date                    string `json:"date,omitempty"`
	ShiftCode               string `json:"shift_code,omitempty"`
	OpenOrders              int    `json:"open_orders"`
	MaterialIssuedOrders    int    `json:"material_issued_orders"`
	MaterialTransferCount   int    `json:"material_transfer_count"`
	SamplePending           int    `json:"sample_pending"`
	FactoryClaims           int    `json:"factory_claims"`
	FactoryClaimsOverdue    int    `json:"factory_claims_overdue"`
	FinalPaymentReadyOrders int    `json:"final_payment_ready_orders"`
	GeneratedAt             string `json:"generated_at"`
}

type availableStockResponse struct {
	WarehouseID      string `json:"warehouse_id"`
	WarehouseCode    string `json:"warehouse_code"`
	LocationID       string `json:"location_id,omitempty"`
	LocationCode     string `json:"location_code,omitempty"`
	SKU              string `json:"sku"`
	BatchID          string `json:"batch_id,omitempty"`
	BatchNo          string `json:"batch_no,omitempty"`
	BatchQCStatus    string `json:"batch_qc_status,omitempty"`
	BatchStatus      string `json:"batch_status,omitempty"`
	BatchExpiryDate  string `json:"batch_expiry_date,omitempty"`
	BaseUOMCode      string `json:"base_uom_code"`
	PhysicalQty      string `json:"physical_qty"`
	ReservedQty      string `json:"reserved_qty"`
	QCHoldQty        string `json:"qc_hold_qty"`
	DamagedQty       string `json:"damaged_qty"`
	ReturnPendingQty string `json:"return_pending_qty"`
	BlockedQty       string `json:"blocked_qty"`
	HoldQty          string `json:"hold_qty"`
	AvailableQty     string `json:"available_qty"`
}

type stockAdjustmentLineRequest struct {
	ID           string `json:"id"`
	ItemID       string `json:"item_id"`
	SKU          string `json:"sku"`
	BatchID      string `json:"batch_id"`
	BatchNo      string `json:"batch_no"`
	LocationID   string `json:"location_id"`
	LocationCode string `json:"location_code"`
	ExpectedQty  string `json:"expected_qty"`
	CountedQty   string `json:"counted_qty"`
	BaseUOMCode  string `json:"base_uom_code"`
	Reason       string `json:"reason"`
}

type createStockAdjustmentRequest struct {
	ID            string                       `json:"id"`
	AdjustmentNo  string                       `json:"adjustment_no"`
	OrgID         string                       `json:"org_id"`
	WarehouseID   string                       `json:"warehouse_id"`
	WarehouseCode string                       `json:"warehouse_code"`
	SourceType    string                       `json:"source_type"`
	SourceID      string                       `json:"source_id"`
	Reason        string                       `json:"reason"`
	Lines         []stockAdjustmentLineRequest `json:"lines"`
}

type stockCountLineRequest struct {
	ID           string `json:"id"`
	ItemID       string `json:"item_id"`
	SKU          string `json:"sku"`
	BatchID      string `json:"batch_id"`
	BatchNo      string `json:"batch_no"`
	LocationID   string `json:"location_id"`
	LocationCode string `json:"location_code"`
	ExpectedQty  string `json:"expected_qty"`
	BaseUOMCode  string `json:"base_uom_code"`
}

type createStockCountRequest struct {
	ID            string                  `json:"id"`
	CountNo       string                  `json:"count_no"`
	OrgID         string                  `json:"org_id"`
	WarehouseID   string                  `json:"warehouse_id"`
	WarehouseCode string                  `json:"warehouse_code"`
	Scope         string                  `json:"scope"`
	Lines         []stockCountLineRequest `json:"lines"`
}

type submitStockCountLineRequest struct {
	ID         string `json:"id"`
	SKU        string `json:"sku"`
	CountedQty string `json:"counted_qty"`
	Note       string `json:"note"`
}

type submitStockCountRequest struct {
	Lines []submitStockCountLineRequest `json:"lines"`
}

type stockCountLineResponse struct {
	ID           string `json:"id"`
	ItemID       string `json:"item_id,omitempty"`
	SKU          string `json:"sku"`
	BatchID      string `json:"batch_id,omitempty"`
	BatchNo      string `json:"batch_no,omitempty"`
	LocationID   string `json:"location_id,omitempty"`
	LocationCode string `json:"location_code,omitempty"`
	ExpectedQty  string `json:"expected_qty"`
	CountedQty   string `json:"counted_qty"`
	DeltaQty     string `json:"delta_qty"`
	BaseUOMCode  string `json:"base_uom_code"`
	Counted      bool   `json:"counted"`
	Note         string `json:"note,omitempty"`
}

type stockCountResponse struct {
	ID            string                   `json:"id"`
	CountNo       string                   `json:"count_no"`
	OrgID         string                   `json:"org_id"`
	WarehouseID   string                   `json:"warehouse_id"`
	WarehouseCode string                   `json:"warehouse_code,omitempty"`
	Scope         string                   `json:"scope"`
	Status        string                   `json:"status"`
	CreatedBy     string                   `json:"created_by"`
	SubmittedBy   string                   `json:"submitted_by,omitempty"`
	Lines         []stockCountLineResponse `json:"lines"`
	AuditLogID    string                   `json:"audit_log_id,omitempty"`
	CreatedAt     string                   `json:"created_at"`
	UpdatedAt     string                   `json:"updated_at"`
	SubmittedAt   string                   `json:"submitted_at,omitempty"`
}

type stockAdjustmentLineResponse struct {
	ID           string `json:"id"`
	ItemID       string `json:"item_id,omitempty"`
	SKU          string `json:"sku"`
	BatchID      string `json:"batch_id,omitempty"`
	BatchNo      string `json:"batch_no,omitempty"`
	LocationID   string `json:"location_id,omitempty"`
	LocationCode string `json:"location_code,omitempty"`
	ExpectedQty  string `json:"expected_qty"`
	CountedQty   string `json:"counted_qty"`
	DeltaQty     string `json:"delta_qty"`
	BaseUOMCode  string `json:"base_uom_code"`
	Reason       string `json:"reason,omitempty"`
}

type stockAdjustmentResponse struct {
	ID            string                        `json:"id"`
	AdjustmentNo  string                        `json:"adjustment_no"`
	OrgID         string                        `json:"org_id"`
	WarehouseID   string                        `json:"warehouse_id"`
	WarehouseCode string                        `json:"warehouse_code,omitempty"`
	SourceType    string                        `json:"source_type,omitempty"`
	SourceID      string                        `json:"source_id,omitempty"`
	Reason        string                        `json:"reason"`
	Status        string                        `json:"status"`
	RequestedBy   string                        `json:"requested_by"`
	SubmittedBy   string                        `json:"submitted_by,omitempty"`
	ApprovedBy    string                        `json:"approved_by,omitempty"`
	RejectedBy    string                        `json:"rejected_by,omitempty"`
	PostedBy      string                        `json:"posted_by,omitempty"`
	Lines         []stockAdjustmentLineResponse `json:"lines"`
	AuditLogID    string                        `json:"audit_log_id,omitempty"`
	CreatedAt     string                        `json:"created_at"`
	UpdatedAt     string                        `json:"updated_at"`
	SubmittedAt   string                        `json:"submitted_at,omitempty"`
	ApprovedAt    string                        `json:"approved_at,omitempty"`
	RejectedAt    string                        `json:"rejected_at,omitempty"`
	PostedAt      string                        `json:"posted_at,omitempty"`
}

type batchResponse struct {
	ID         string `json:"id"`
	OrgID      string `json:"org_id"`
	ItemID     string `json:"item_id"`
	SKU        string `json:"sku"`
	ItemName   string `json:"item_name"`
	BatchNo    string `json:"batch_no"`
	SupplierID string `json:"supplier_id,omitempty"`
	MfgDate    string `json:"mfg_date,omitempty"`
	ExpiryDate string `json:"expiry_date,omitempty"`
	QCStatus   string `json:"qc_status"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type batchQCTransitionRequest struct {
	QCStatus    string `json:"qc_status"`
	Reason      string `json:"reason"`
	BusinessRef string `json:"business_ref"`
}

type batchQCTransitionResponse struct {
	ID           string `json:"id"`
	BatchID      string `json:"batch_id"`
	BatchNo      string `json:"batch_no"`
	SKU          string `json:"sku"`
	FromQCStatus string `json:"from_qc_status"`
	ToQCStatus   string `json:"to_qc_status"`
	ActorID      string `json:"actor_id"`
	Reason       string `json:"reason"`
	BusinessRef  string `json:"business_ref"`
	AuditLogID   string `json:"audit_log_id"`
	CreatedAt    string `json:"created_at"`
}

type batchQCTransitionResultResponse struct {
	Batch      batchResponse             `json:"batch"`
	Transition batchQCTransitionResponse `json:"transition"`
}

type endOfDayReconciliationSummaryResponse struct {
	SystemQuantity     int64 `json:"system_quantity"`
	CountedQuantity    int64 `json:"counted_quantity"`
	VarianceQuantity   int64 `json:"variance_quantity"`
	VarianceCount      int   `json:"variance_count"`
	ChecklistTotal     int   `json:"checklist_total"`
	ChecklistCompleted int   `json:"checklist_completed"`
	ReadyToClose       bool  `json:"ready_to_close"`
}

type endOfDayReconciliationOperationsResponse struct {
	OrderCount             int `json:"order_count"`
	HandoverOrderCount     int `json:"handover_order_count"`
	ReturnOrderCount       int `json:"return_order_count"`
	StockMovementCount     int `json:"stock_movement_count"`
	StockCountSessionCount int `json:"stock_count_session_count"`
	PendingIssueCount      int `json:"pending_issue_count"`
}

type endOfDayReconciliationChecklistResponse struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Complete bool   `json:"complete"`
	Blocking bool   `json:"blocking"`
	Note     string `json:"note,omitempty"`
}

type endOfDayReconciliationLineResponse struct {
	ID               string `json:"id"`
	SKU              string `json:"sku"`
	BatchNo          string `json:"batch_no"`
	BinCode          string `json:"bin_code"`
	SystemQuantity   int64  `json:"system_quantity"`
	CountedQuantity  int64  `json:"counted_quantity"`
	VarianceQuantity int64  `json:"variance_quantity"`
	Reason           string `json:"reason,omitempty"`
	Owner            string `json:"owner"`
}

type endOfDayReconciliationResponse struct {
	ID            string                                    `json:"id"`
	WarehouseID   string                                    `json:"warehouse_id"`
	WarehouseCode string                                    `json:"warehouse_code"`
	Date          string                                    `json:"date"`
	ShiftCode     string                                    `json:"shift_code"`
	Status        string                                    `json:"status"`
	Owner         string                                    `json:"owner"`
	ClosedAt      string                                    `json:"closed_at,omitempty"`
	ClosedBy      string                                    `json:"closed_by,omitempty"`
	AuditLogID    string                                    `json:"audit_log_id,omitempty"`
	Summary       endOfDayReconciliationSummaryResponse     `json:"summary"`
	Operations    endOfDayReconciliationOperationsResponse  `json:"operations"`
	Checklist     []endOfDayReconciliationChecklistResponse `json:"checklist"`
	Lines         []endOfDayReconciliationLineResponse      `json:"lines"`
}

type closeReconciliationRequest struct {
	ExceptionNote string `json:"exception_note"`
}

type carrierManifestSummaryResponse struct {
	ExpectedCount int `json:"expected_count"`
	ScannedCount  int `json:"scanned_count"`
	MissingCount  int `json:"missing_count"`
}

type carrierManifestLineResponse struct {
	ID               string `json:"id"`
	ShipmentID       string `json:"shipment_id"`
	OrderNo          string `json:"order_no"`
	TrackingNo       string `json:"tracking_no"`
	PackageCode      string `json:"package_code"`
	StagingZone      string `json:"staging_zone"`
	HandoverZoneCode string `json:"handover_zone_code,omitempty"`
	HandoverBinCode  string `json:"handover_bin_code,omitempty"`
	Scanned          bool   `json:"scanned"`
}

type carrierManifestResponse struct {
	ID               string                         `json:"id"`
	CarrierCode      string                         `json:"carrier_code"`
	CarrierName      string                         `json:"carrier_name"`
	WarehouseID      string                         `json:"warehouse_id"`
	WarehouseCode    string                         `json:"warehouse_code"`
	Date             string                         `json:"date"`
	HandoverBatch    string                         `json:"handover_batch"`
	StagingZone      string                         `json:"staging_zone"`
	HandoverZoneCode string                         `json:"handover_zone_code,omitempty"`
	HandoverBinCode  string                         `json:"handover_bin_code,omitempty"`
	Status           string                         `json:"status"`
	Owner            string                         `json:"owner"`
	AuditLogID       string                         `json:"audit_log_id,omitempty"`
	Summary          carrierManifestSummaryResponse `json:"summary"`
	Lines            []carrierManifestLineResponse  `json:"lines"`
	MissingLines     []carrierManifestLineResponse  `json:"missing_lines"`
	CreatedAt        string                         `json:"created_at,omitempty"`
}

type createCarrierManifestRequest struct {
	ID               string `json:"id"`
	CarrierCode      string `json:"carrier_code"`
	CarrierName      string `json:"carrier_name"`
	WarehouseID      string `json:"warehouse_id"`
	WarehouseCode    string `json:"warehouse_code"`
	Date             string `json:"date"`
	HandoverBatch    string `json:"handover_batch"`
	StagingZone      string `json:"staging_zone"`
	HandoverZoneID   string `json:"handover_zone_id"`
	HandoverZoneCode string `json:"handover_zone_code"`
	HandoverBinID    string `json:"handover_bin_id"`
	HandoverBinCode  string `json:"handover_bin_code"`
	Owner            string `json:"owner"`
}

type addShipmentToManifestRequest struct {
	ShipmentID string `json:"shipment_id"`
}

type cancelCarrierManifestRequest struct {
	Reason string `json:"reason"`
}

type verifyCarrierManifestScanRequest struct {
	Code      string `json:"code"`
	StationID string `json:"station_id"`
	DeviceID  string `json:"device_id"`
	Source    string `json:"source"`
}

type carrierManifestScanEventResponse struct {
	ID                 string `json:"id"`
	ManifestID         string `json:"manifest_id"`
	ExpectedManifestID string `json:"expected_manifest_id,omitempty"`
	Code               string `json:"code"`
	ResultCode         string `json:"result_code"`
	Severity           string `json:"severity"`
	Message            string `json:"message"`
	ShipmentID         string `json:"shipment_id,omitempty"`
	OrderNo            string `json:"order_no,omitempty"`
	TrackingNo         string `json:"tracking_no,omitempty"`
	ActorID            string `json:"actor_id"`
	StationID          string `json:"station_id"`
	DeviceID           string `json:"device_id,omitempty"`
	Source             string `json:"source"`
	WarehouseID        string `json:"warehouse_id"`
	CarrierCode        string `json:"carrier_code"`
	CreatedAt          string `json:"created_at"`
}

type carrierManifestScanResponse struct {
	ResultCode         string                           `json:"result_code"`
	Severity           string                           `json:"severity"`
	Message            string                           `json:"message"`
	ExpectedManifestID string                           `json:"expected_manifest_id,omitempty"`
	Line               *carrierManifestLineResponse     `json:"line,omitempty"`
	ScanEvent          carrierManifestScanEventResponse `json:"scan_event"`
	Manifest           carrierManifestResponse          `json:"manifest"`
	AuditLogID         string                           `json:"audit_log_id,omitempty"`
}

type warehouseReceivingLineResponse struct {
	ID                  string `json:"id"`
	PurchaseOrderLineID string `json:"purchase_order_line_id,omitempty"`
	ItemID              string `json:"item_id"`
	SKU                 string `json:"sku"`
	ItemName            string `json:"item_name,omitempty"`
	BatchID             string `json:"batch_id,omitempty"`
	BatchNo             string `json:"batch_no,omitempty"`
	LotNo               string `json:"lot_no,omitempty"`
	ExpiryDate          string `json:"expiry_date,omitempty"`
	WarehouseID         string `json:"warehouse_id"`
	LocationID          string `json:"location_id"`
	Quantity            string `json:"quantity"`
	UOMCode             string `json:"uom_code"`
	BaseUOMCode         string `json:"base_uom_code"`
	PackagingStatus     string `json:"packaging_status"`
	QCStatus            string `json:"qc_status,omitempty"`
}

type warehouseReceivingStockMovementResponse struct {
	MovementNo      string `json:"movement_no"`
	MovementType    string `json:"movement_type"`
	ItemID          string `json:"item_id"`
	BatchID         string `json:"batch_id"`
	WarehouseID     string `json:"warehouse_id"`
	LocationID      string `json:"location_id"`
	Quantity        string `json:"quantity"`
	BaseUOMCode     string `json:"base_uom_code"`
	StockStatus     string `json:"stock_status"`
	SourceDocID     string `json:"source_doc_id"`
	SourceDocLineID string `json:"source_doc_line_id"`
}

type warehouseReceivingResponse struct {
	ID               string                                    `json:"id"`
	OrgID            string                                    `json:"org_id"`
	ReceiptNo        string                                    `json:"receipt_no"`
	WarehouseID      string                                    `json:"warehouse_id"`
	WarehouseCode    string                                    `json:"warehouse_code"`
	LocationID       string                                    `json:"location_id"`
	LocationCode     string                                    `json:"location_code"`
	ReferenceDocType string                                    `json:"reference_doc_type"`
	ReferenceDocID   string                                    `json:"reference_doc_id"`
	SupplierID       string                                    `json:"supplier_id,omitempty"`
	DeliveryNoteNo   string                                    `json:"delivery_note_no,omitempty"`
	Status           string                                    `json:"status"`
	Lines            []warehouseReceivingLineResponse          `json:"lines"`
	StockMovements   []warehouseReceivingStockMovementResponse `json:"stock_movements,omitempty"`
	CreatedBy        string                                    `json:"created_by"`
	SubmittedBy      string                                    `json:"submitted_by,omitempty"`
	InspectReadyBy   string                                    `json:"inspect_ready_by,omitempty"`
	PostedBy         string                                    `json:"posted_by,omitempty"`
	AuditLogID       string                                    `json:"audit_log_id,omitempty"`
	CreatedAt        string                                    `json:"created_at"`
	UpdatedAt        string                                    `json:"updated_at"`
	SubmittedAt      string                                    `json:"submitted_at,omitempty"`
	InspectReadyAt   string                                    `json:"inspect_ready_at,omitempty"`
	PostedAt         string                                    `json:"posted_at,omitempty"`
}

type createWarehouseReceivingLineRequest struct {
	ID                  string `json:"id"`
	PurchaseOrderLineID string `json:"purchase_order_line_id"`
	ItemID              string `json:"item_id"`
	SKU                 string `json:"sku"`
	ItemName            string `json:"item_name"`
	BatchID             string `json:"batch_id"`
	BatchNo             string `json:"batch_no"`
	LotNo               string `json:"lot_no"`
	ExpiryDate          string `json:"expiry_date"`
	Quantity            string `json:"quantity"`
	UOMCode             string `json:"uom_code"`
	BaseUOMCode         string `json:"base_uom_code"`
	PackagingStatus     string `json:"packaging_status"`
	QCStatus            string `json:"qc_status"`
}

type createWarehouseReceivingRequest struct {
	ID               string                                `json:"id"`
	OrgID            string                                `json:"org_id"`
	ReceiptNo        string                                `json:"receipt_no"`
	WarehouseID      string                                `json:"warehouse_id"`
	LocationID       string                                `json:"location_id"`
	ReferenceDocType string                                `json:"reference_doc_type"`
	ReferenceDocID   string                                `json:"reference_doc_id"`
	SupplierID       string                                `json:"supplier_id"`
	DeliveryNoteNo   string                                `json:"delivery_note_no"`
	Lines            []createWarehouseReceivingLineRequest `json:"lines"`
}

type returnReceiptLineResponse struct {
	ID          string `json:"id"`
	SKU         string `json:"sku"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	Condition   string `json:"condition"`
}

type returnStockMovementResponse struct {
	ID                string `json:"id"`
	MovementType      string `json:"movement_type"`
	SKU               string `json:"sku"`
	WarehouseID       string `json:"warehouse_id"`
	Quantity          int    `json:"quantity"`
	TargetStockStatus string `json:"target_stock_status"`
	SourceDocID       string `json:"source_doc_id"`
}

type returnReceiptResponse struct {
	ID                string                       `json:"id"`
	ReceiptNo         string                       `json:"receipt_no"`
	WarehouseID       string                       `json:"warehouse_id"`
	WarehouseCode     string                       `json:"warehouse_code"`
	Source            string                       `json:"source"`
	ReceivedBy        string                       `json:"received_by"`
	ReceivedAt        string                       `json:"received_at"`
	PackageCondition  string                       `json:"package_condition"`
	Status            string                       `json:"status"`
	Disposition       string                       `json:"disposition"`
	TargetLocation    string                       `json:"target_location"`
	OriginalOrderNo   string                       `json:"original_order_no,omitempty"`
	TrackingNo        string                       `json:"tracking_no,omitempty"`
	ReturnCode        string                       `json:"return_code,omitempty"`
	ScanCode          string                       `json:"scan_code"`
	CustomerName      string                       `json:"customer_name"`
	UnknownCase       bool                         `json:"unknown_case"`
	Lines             []returnReceiptLineResponse  `json:"lines"`
	StockMovement     *returnStockMovementResponse `json:"stock_movement,omitempty"`
	InvestigationNote string                       `json:"investigation_note,omitempty"`
	AuditLogID        string                       `json:"audit_log_id,omitempty"`
	CreatedAt         string                       `json:"created_at"`
}

type returnInspectionResponse struct {
	ID             string `json:"id"`
	ReceiptID      string `json:"receipt_id"`
	ReceiptNo      string `json:"receipt_no"`
	Condition      string `json:"condition"`
	Disposition    string `json:"disposition"`
	Status         string `json:"status"`
	TargetLocation string `json:"target_location"`
	RiskLevel      string `json:"risk_level"`
	InspectorID    string `json:"inspector_id"`
	Note           string `json:"note,omitempty"`
	EvidenceLabel  string `json:"evidence_label,omitempty"`
	AuditLogID     string `json:"audit_log_id,omitempty"`
	InspectedAt    string `json:"inspected_at"`
}

type returnDispositionActionResponse struct {
	ID                string `json:"id"`
	ReceiptID         string `json:"receipt_id"`
	ReceiptNo         string `json:"receipt_no"`
	Disposition       string `json:"disposition"`
	TargetLocation    string `json:"target_location"`
	TargetStockStatus string `json:"target_stock_status"`
	ActionCode        string `json:"action_code"`
	ActorID           string `json:"actor_id"`
	Note              string `json:"note,omitempty"`
	AuditLogID        string `json:"audit_log_id,omitempty"`
	DecidedAt         string `json:"decided_at"`
}

type returnAttachmentResponse struct {
	ID            string `json:"id"`
	ReceiptID     string `json:"receipt_id"`
	ReceiptNo     string `json:"receipt_no"`
	InspectionID  string `json:"inspection_id"`
	FileName      string `json:"file_name"`
	FileExt       string `json:"file_ext,omitempty"`
	MIMEType      string `json:"mime_type"`
	FileSizeBytes int64  `json:"file_size_bytes"`
	StorageBucket string `json:"storage_bucket"`
	StorageKey    string `json:"storage_key"`
	Status        string `json:"status"`
	UploadedBy    string `json:"uploaded_by"`
	Note          string `json:"note,omitempty"`
	AuditLogID    string `json:"audit_log_id,omitempty"`
	UploadedAt    string `json:"uploaded_at"`
}

type receiveReturnRequest struct {
	WarehouseID       string `json:"warehouse_id"`
	WarehouseCode     string `json:"warehouse_code"`
	Source            string `json:"source"`
	Code              string `json:"code"`
	PackageCondition  string `json:"package_condition"`
	Disposition       string `json:"disposition"`
	InvestigationNote string `json:"investigation_note"`
}

type inspectReturnRequest struct {
	Condition     string `json:"condition"`
	Disposition   string `json:"disposition"`
	Note          string `json:"note"`
	EvidenceLabel string `json:"evidence_label"`
}

type applyReturnDispositionRequest struct {
	Disposition string `json:"disposition"`
	Note        string `json:"note"`
}

type returnMasterDataResponse struct {
	Reasons      []returnReasonResponse      `json:"reasons"`
	Conditions   []returnConditionResponse   `json:"conditions"`
	Dispositions []returnDispositionResponse `json:"dispositions"`
}

type returnReasonResponse struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	SortOrder   int    `json:"sort_order"`
}

type returnConditionResponse struct {
	Code                 string `json:"code"`
	Label                string `json:"label"`
	Description          string `json:"description"`
	DefaultDisposition   string `json:"default_disposition"`
	InventoryDisposition string `json:"inventory_disposition"`
	RequiresQA           bool   `json:"requires_qa"`
	Active               bool   `json:"active"`
	SortOrder            int    `json:"sort_order"`
}

type returnDispositionResponse struct {
	Code                  string `json:"code"`
	Label                 string `json:"label"`
	Description           string `json:"description"`
	InventoryDisposition  string `json:"inventory_disposition"`
	TargetStockStatus     string `json:"target_stock_status"`
	TargetLocationType    string `json:"target_location_type"`
	CreatesAvailableStock bool   `json:"creates_available_stock"`
	RequiresApproval      bool   `json:"requires_approval"`
	Active                bool   `json:"active"`
	SortOrder             int    `json:"sort_order"`
}

type stockMovementRequest struct {
	MovementID       string `json:"movementId"`
	SKU              string `json:"sku"`
	WarehouseID      string `json:"warehouseId"`
	MovementType     string `json:"movementType"`
	Quantity         string `json:"quantity"`
	BaseUOMCode      string `json:"baseUomCode"`
	SourceQuantity   string `json:"sourceQuantity"`
	SourceUOMCode    string `json:"sourceUomCode"`
	ConversionFactor string `json:"conversionFactor"`
	Reason           string `json:"reason"`
}

type stockMovementResponse struct {
	MovementID       string `json:"movement_id"`
	Status           string `json:"status"`
	MovementQuantity string `json:"movement_qty"`
	BaseUOMCode      string `json:"base_uom_code"`
	SourceQuantity   string `json:"source_qty"`
	SourceUOMCode    string `json:"source_uom_code"`
	ConversionFactor string `json:"conversion_factor"`
}

type auditLogResponse struct {
	ID         string         `json:"id"`
	ActorID    string         `json:"actor_id"`
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	RequestID  string         `json:"request_id,omitempty"`
	BeforeData map[string]any `json:"before_data,omitempty"`
	AfterData  map[string]any `json:"after_data,omitempty"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  string         `json:"created_at"`
}

func main() {
	cfg := config.FromEnv()
	authConfig := auth.MockConfig{
		Email:       cfg.AuthMockEmail,
		Password:    cfg.AuthMockPassword,
		AccessToken: cfg.AuthMockAccessToken,
	}
	authSessions := auth.NewSessionManager(authConfig, time.Now)
	availableStockService := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	stockAdjustmentStore := inventoryapp.NewPrototypeStockAdjustmentStore()
	stockCountStore := inventoryapp.NewPrototypeStockCountStore()
	auditLogStore := audit.NewPrototypeLogStore()
	stockMovementStore := inventoryapp.NewInMemoryStockMovementStore()
	listStockAdjustments := inventoryapp.NewListStockAdjustments(stockAdjustmentStore)
	createStockAdjustment := inventoryapp.NewCreateStockAdjustment(stockAdjustmentStore, auditLogStore)
	transitionStockAdjustment := inventoryapp.NewTransitionStockAdjustment(stockAdjustmentStore, stockMovementStore, auditLogStore)
	listStockCounts := inventoryapp.NewListStockCounts(stockCountStore)
	createStockCount := inventoryapp.NewCreateStockCount(stockCountStore, auditLogStore)
	submitStockCount := inventoryapp.NewSubmitStockCount(stockCountStore, auditLogStore)
	batchCatalog := inventoryapp.NewPrototypeBatchCatalog(auditLogStore)
	itemCatalog := masterdataapp.NewPrototypeItemCatalog(auditLogStore)
	uomCatalog := masterdataapp.NewPrototypeUOMCatalog()
	warehouseCatalog := masterdataapp.NewPrototypeWarehouseLocationCatalog(auditLogStore)
	partyCatalog := masterdataapp.NewPrototypePartyCatalog(auditLogStore)
	purchaseOrderStore := purchaseapp.NewPrototypePurchaseOrderStore(auditLogStore)
	purchaseOrderService := purchaseapp.NewPurchaseOrderService(
		purchaseOrderStore,
		partyCatalog,
		itemCatalog,
		warehouseCatalog,
		purchaseOrderUOMConverterAdapter{catalog: uomCatalog},
	)
	subcontractOrderStore := productionapp.NewPrototypeSubcontractOrderStore(auditLogStore)
	subcontractMaterialTransferStore := productionapp.NewPrototypeSubcontractMaterialTransferStore()
	subcontractSampleApprovalStore := productionapp.NewPrototypeSubcontractSampleApprovalStore()
	subcontractFinishedGoodsReceiptStore := productionapp.NewPrototypeSubcontractFinishedGoodsReceiptStore()
	subcontractFactoryClaimStore := productionapp.NewPrototypeSubcontractFactoryClaimStore()
	subcontractPaymentMilestoneStore := productionapp.NewPrototypeSubcontractPaymentMilestoneStore()
	supplierPayableStore := financeapp.NewPrototypeSupplierPayableStore()
	supplierPayableService := financeapp.NewSupplierPayableService(supplierPayableStore, auditLogStore)
	subcontractOrderService := productionapp.NewSubcontractOrderService(
		subcontractOrderStore,
		partyCatalog,
		itemCatalog,
		subcontractOrderUOMConverterAdapter{catalog: uomCatalog},
	).
		WithMaterialIssueStores(subcontractMaterialTransferStore, stockMovementStore).
		WithSampleApprovalStore(subcontractSampleApprovalStore).
		WithFinishedGoodsReceiptStores(subcontractFinishedGoodsReceiptStore, stockMovementStore).
		WithFactoryClaimStore(subcontractFactoryClaimStore).
		WithPaymentMilestoneStore(subcontractPaymentMilestoneStore).
		WithSubcontractPayableCreator(subcontractSupplierPayableAdapter{service: supplierPayableService})
	salesOrderStore := salesapp.NewPrototypeSalesOrderStore(auditLogStore)
	salesOrderReservationStore := inventoryapp.NewPrototypeSalesOrderReservationStore(auditLogStore)
	salesOrderService := salesapp.NewSalesOrderService(salesOrderStore, partyCatalog, itemCatalog, warehouseCatalog).
		WithStockReserver(salesOrderReservationStore)
	customerReceivableStore := financeapp.NewPrototypeCustomerReceivableStore()
	customerReceivableService := financeapp.NewCustomerReceivableService(customerReceivableStore, auditLogStore)
	codRemittanceStore := financeapp.NewPrototypeCODRemittanceStore()
	codRemittanceService := financeapp.NewCODRemittanceService(codRemittanceStore, auditLogStore)
	cashTransactionStore := financeapp.NewPrototypeCashTransactionStore()
	cashTransactionService := financeapp.NewCashTransactionService(cashTransactionStore, auditLogStore)
	financeDashboardService := financeapp.NewFinanceDashboardService(
		customerReceivableStore,
		supplierPayableStore,
		codRemittanceStore,
		cashTransactionStore,
	)
	warehouseReceivingStore := inventoryapp.NewPrototypeWarehouseReceivingStore()
	warehouseReceiving := inventoryapp.NewWarehouseReceivingService(
		warehouseReceivingStore,
		warehouseCatalog,
		batchCatalog,
		stockMovementStore,
		auditLogStore,
	).WithPurchaseOrderReader(purchaseOrderService)
	inboundQCStore := qcapp.NewPrototypeInboundQCInspectionStore()
	inboundQCInspections := qcapp.NewInboundQCInspectionService(inboundQCStore, warehouseReceivingStore, auditLogStore).
		WithStockMovementRecorder(stockMovementStore).
		WithBatchQCStatusUpdater(inboundQCBatchQCStatusAdapter{catalog: batchCatalog})
	supplierRejectionStore := inventoryapp.NewPrototypeSupplierRejectionStore()
	listSupplierRejections := inventoryapp.NewListSupplierRejections(supplierRejectionStore)
	createSupplierRejection := inventoryapp.NewCreateSupplierRejection(supplierRejectionStore, auditLogStore)
	transitionSupplierRejection := inventoryapp.NewTransitionSupplierRejection(supplierRejectionStore, auditLogStore)
	reconciliationStore := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	listEndOfDayReconciliations := inventoryapp.NewListEndOfDayReconciliations(reconciliationStore)
	closeEndOfDayReconciliation := inventoryapp.NewCloseEndOfDayReconciliation(reconciliationStore, auditLogStore)
	carrierManifestStore := shippingapp.NewPrototypeCarrierManifestStore()
	listCarrierManifests := shippingapp.NewListCarrierManifests(carrierManifestStore)
	createCarrierManifest := shippingapp.NewCreateCarrierManifest(carrierManifestStore, auditLogStore)
	addShipmentToCarrierManifest := shippingapp.NewAddShipmentToCarrierManifest(carrierManifestStore, auditLogStore)
	removeShipmentFromCarrierManifest := shippingapp.NewRemoveShipmentFromCarrierManifest(carrierManifestStore, auditLogStore)
	markCarrierManifestReadyToScan := shippingapp.NewMarkCarrierManifestReadyToScan(carrierManifestStore, auditLogStore)
	cancelCarrierManifest := shippingapp.NewCancelCarrierManifest(carrierManifestStore, auditLogStore)
	reportCarrierManifestMissingOrders := shippingapp.NewReportCarrierManifestMissingOrders(carrierManifestStore, auditLogStore)
	confirmCarrierManifestHandover := shippingapp.NewConfirmCarrierManifestHandover(
		carrierManifestStore,
		auditLogStore,
		salesOrderHandoverAdapter{service: salesOrderService},
	)
	verifyCarrierManifestScan := shippingapp.NewVerifyCarrierManifestScan(carrierManifestStore, auditLogStore)
	pickTaskStore := shippingapp.NewPrototypePickTaskStore(mustPrototypePickTask())
	listPickTasks := shippingapp.NewListPickTasks(pickTaskStore)
	getPickTask := shippingapp.NewGetPickTask(pickTaskStore)
	startPickTask := shippingapp.NewStartPickTask(pickTaskStore, auditLogStore)
	confirmPickTaskLine := shippingapp.NewConfirmPickTaskLine(pickTaskStore, auditLogStore)
	completePickTask := shippingapp.NewCompletePickTask(pickTaskStore, auditLogStore)
	reportPickTaskException := shippingapp.NewReportPickTaskException(pickTaskStore, auditLogStore)
	packTaskStore := shippingapp.NewPrototypePackTaskStore(mustPrototypePackTask())
	listPackTasks := shippingapp.NewListPackTasks(packTaskStore)
	getPackTask := shippingapp.NewGetPackTask(packTaskStore)
	startPackTask := shippingapp.NewStartPackTask(packTaskStore, auditLogStore)
	confirmPackTask := shippingapp.NewConfirmPackTask(packTaskStore, auditLogStore, salesOrderPackerAdapter{service: salesOrderService})
	reportPackTaskException := shippingapp.NewReportPackTaskException(packTaskStore, auditLogStore)
	returnReceiptStore := returnsapp.NewPrototypeReturnReceiptStore()
	listReturnMasterData := returnsapp.NewListReturnMasterData()
	listReturnReceipts := returnsapp.NewListReturnReceipts(returnReceiptStore)
	receiveReturn := returnsapp.NewReceiveReturn(returnReceiptStore, auditLogStore)
	inspectReturn := returnsapp.NewInspectReturn(returnReceiptStore, auditLogStore)
	applyReturnDisposition := returnsapp.NewApplyReturnDisposition(returnReceiptStore, stockMovementStore, auditLogStore)
	attachmentObjectStore := storage.NewS3CompatibleObjectStore(storage.S3Config{
		Endpoint:     cfg.S3Endpoint,
		AccessKey:    cfg.S3AccessKey,
		SecretKey:    cfg.S3SecretKey,
		UseSSL:       cfg.S3UseSSL,
		UsePathStyle: cfg.S3UsePathStyle,
	})
	uploadReturnAttachment := returnsapp.NewUploadReturnAttachment(returnReceiptStore, auditLogStore).
		WithObjectStore(attachmentObjectStore).
		WithStorageBucket(cfg.S3Bucket)
	operationsDailySignals := operationsDailyRuntimeSignalSource{
		receivings:           warehouseReceiving,
		inboundQC:            inboundQCInspections,
		carrierManifests:     listCarrierManifests,
		pickTasks:            listPickTasks,
		returnReceipts:       listReturnReceipts,
		stockCounts:          listStockCounts,
		subcontractOrders:    subcontractOrderService,
		subcontractTransfers: subcontractMaterialTransferStore,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readinessHandler)
	mux.HandleFunc("/api/v1/health", healthHandler)
	mux.HandleFunc("/api/v1/ready", readinessHandler)
	mux.HandleFunc("/api/v1/auth/login", loginHandler(authSessions, auditLogStore))
	mux.HandleFunc("/api/v1/auth/mock-login", loginHandler(authSessions, auditLogStore))
	mux.HandleFunc("/api/v1/auth/refresh", refreshHandler(authSessions, auditLogStore))
	mux.HandleFunc("/api/v1/auth/policy", authPolicyHandler(authSessions))
	mux.Handle("/api/v1/me", auth.RequireSessionToken(authSessions, http.HandlerFunc(meHandler)))
	mux.Handle(
		"/api/v1/rbac/roles",
		auth.RequireSessionPermission(authSessions, auth.PermissionSettingsView, http.HandlerFunc(rbacRolesHandler)),
	)
	mux.Handle(
		"/api/v1/rbac/permissions",
		auth.RequireSessionPermission(authSessions, auth.PermissionSettingsView, http.HandlerFunc(rbacPermissionsHandler)),
	)
	mux.Handle(
		"/api/v1/audit-logs",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionAuditLogView,
			http.HandlerFunc(auditLogsHandler(auditLogStore)),
		),
	)
	mux.Handle(
		"/api/v1/products",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(productsHandler(itemCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/products/{product_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(productDetailHandler(itemCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/products/{product_id}/status",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(changeProductStatusHandler(itemCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/warehouses",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(warehousesHandler(warehouseCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/warehouses/{warehouse_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(warehouseDetailHandler(warehouseCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/warehouses/{warehouse_id}/status",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(changeWarehouseStatusHandler(warehouseCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse-locations",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(warehouseLocationsHandler(warehouseCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse-locations/{location_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(warehouseLocationDetailHandler(warehouseCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse-locations/{location_id}/status",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(changeWarehouseLocationStatusHandler(warehouseCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/suppliers",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(suppliersHandler(partyCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/suppliers/{supplier_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierDetailHandler(partyCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/suppliers/{supplier_id}/status",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(changeSupplierStatusHandler(partyCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/customers",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(customersHandler(partyCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/customers/{customer_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(customerDetailHandler(partyCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/customers/{customer_id}/status",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(changeCustomerStatusHandler(partyCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/sales-orders",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(salesOrdersHandler(salesOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/sales-orders/{sales_order_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(salesOrderDetailHandler(salesOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/sales-orders/{sales_order_id}/confirm",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(salesOrderConfirmHandler(salesOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/sales-orders/{sales_order_id}/cancel",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(salesOrderCancelHandler(salesOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/customer-receivables",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(customerReceivablesHandler(customerReceivableService)),
		),
	)
	mux.Handle(
		"/api/v1/customer-receivables/{customer_receivable_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(customerReceivableDetailHandler(customerReceivableService)),
		),
	)
	mux.Handle(
		"/api/v1/customer-receivables/{customer_receivable_id}/record-receipt",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(customerReceivableRecordReceiptHandler(customerReceivableService)),
		),
	)
	mux.Handle(
		"/api/v1/customer-receivables/{customer_receivable_id}/mark-disputed",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(customerReceivableMarkDisputedHandler(customerReceivableService)),
		),
	)
	mux.Handle(
		"/api/v1/customer-receivables/{customer_receivable_id}/void",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(customerReceivableVoidHandler(customerReceivableService)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-payables",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierPayablesHandler(supplierPayableService)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-payables/{supplier_payable_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierPayableDetailHandler(supplierPayableService)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-payables/{supplier_payable_id}/request-payment",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierPayableRequestPaymentHandler(supplierPayableService)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-payables/{supplier_payable_id}/approve-payment",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierPayableApprovePaymentHandler(supplierPayableService)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-payables/{supplier_payable_id}/reject-payment",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierPayableRejectPaymentHandler(supplierPayableService)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-payables/{supplier_payable_id}/record-payment",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierPayableRecordPaymentHandler(supplierPayableService)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-payables/{supplier_payable_id}/void",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierPayableVoidHandler(supplierPayableService)),
		),
	)
	mux.Handle(
		"/api/v1/cash-transactions",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(cashTransactionsHandler(cashTransactionService)),
		),
	)
	mux.Handle(
		"/api/v1/cash-transactions/{cash_transaction_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(cashTransactionDetailHandler(cashTransactionService)),
		),
	)
	mux.Handle(
		"/api/v1/finance/dashboard",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(financeDashboardHandler(financeDashboardService)),
		),
	)
	mux.Handle(
		"/api/v1/cod-remittances",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(codRemittancesHandler(codRemittanceService)),
		),
	)
	mux.Handle(
		"/api/v1/cod-remittances/{cod_remittance_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(codRemittanceDetailHandler(codRemittanceService)),
		),
	)
	mux.Handle(
		"/api/v1/cod-remittances/{cod_remittance_id}/match",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(codRemittanceMatchHandler(codRemittanceService)),
		),
	)
	mux.Handle(
		"/api/v1/cod-remittances/{cod_remittance_id}/record-discrepancy",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(codRemittanceDiscrepancyHandler(codRemittanceService)),
		),
	)
	mux.Handle(
		"/api/v1/cod-remittances/{cod_remittance_id}/submit",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(codRemittanceSubmitHandler(codRemittanceService)),
		),
	)
	mux.Handle(
		"/api/v1/cod-remittances/{cod_remittance_id}/approve",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(codRemittanceApproveHandler(codRemittanceService)),
		),
	)
	mux.Handle(
		"/api/v1/cod-remittances/{cod_remittance_id}/close",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(codRemittanceCloseHandler(codRemittanceService)),
		),
	)
	mux.Handle(
		"/api/v1/purchase-orders",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionPurchaseView,
			http.HandlerFunc(purchaseOrdersHandler(purchaseOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/purchase-orders/{purchase_order_id}",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionPurchaseView,
			http.HandlerFunc(purchaseOrderDetailHandler(purchaseOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/purchase-orders/{purchase_order_id}/submit",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(purchaseOrderSubmitHandler(purchaseOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/purchase-orders/{purchase_order_id}/approve",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(purchaseOrderApproveHandler(purchaseOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/purchase-orders/{purchase_order_id}/cancel",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(purchaseOrderCancelHandler(purchaseOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/purchase-orders/{purchase_order_id}/close",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(purchaseOrderCloseHandler(purchaseOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionSubcontractView,
			http.HandlerFunc(subcontractOrdersHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionSubcontractView,
			http.HandlerFunc(subcontractOrderDetailHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/submit",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderSubmitHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/approve",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderApproveHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/confirm-factory",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderConfirmFactoryHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/record-deposit",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderRecordDepositHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/issue-materials",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderIssueMaterialsHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/start-mass-production",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderStartMassProductionHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/receive-finished-goods",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderReceiveFinishedGoodsHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/report-factory-defect",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderReportFactoryDefectHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/accept",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderAcceptFinishedGoodsHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/partial-accept",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderPartialAcceptFinishedGoodsHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/mark-final-payment-ready",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderMarkFinalPaymentReadyHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/submit-sample",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderSubmitSampleHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/approve-sample",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderApproveSampleHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/reject-sample",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderRejectSampleHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/cancel",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderCancelHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/subcontract-orders/{subcontract_order_id}/close",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(subcontractOrderCloseHandler(subcontractOrderService)),
		),
	)
	mux.Handle(
		"/api/v1/inventory/stock-movements",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(stockMovementHandler(auditLogStore)),
		),
	)
	mux.Handle(
		"/api/v1/inventory/available-stock",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionInventoryView,
			http.HandlerFunc(availableStockHandler(availableStockService)),
		),
	)
	mux.Handle(
		"/api/v1/reports/inventory-snapshot",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inventorySnapshotReportHandler(availableStockService)),
		),
	)
	mux.Handle(
		"/api/v1/reports/inventory-snapshot/export.csv",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inventorySnapshotCSVExportHandler(availableStockService)),
		),
	)
	mux.Handle(
		"/api/v1/reports/operations-daily",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(operationsDailyReportHandler(operationsDailySignals)),
		),
	)
	mux.Handle(
		"/api/v1/reports/operations-daily/export.csv",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(operationsDailyCSVExportHandler(operationsDailySignals)),
		),
	)
	mux.Handle(
		"/api/v1/reports/finance-summary",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(financeSummaryReportHandler(
				customerReceivableStore,
				supplierPayableStore,
				codRemittanceStore,
				cashTransactionStore,
			)),
		),
	)
	mux.Handle(
		"/api/v1/reports/finance-summary/export.csv",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(financeSummaryCSVExportHandler(
				customerReceivableStore,
				supplierPayableStore,
				codRemittanceStore,
				cashTransactionStore,
			)),
		),
	)
	mux.Handle(
		"/api/v1/stock-adjustments",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(stockAdjustmentsHandler(listStockAdjustments, createStockAdjustment)),
		),
	)
	mux.Handle(
		"/api/v1/stock-adjustments/{stock_adjustment_id}/submit",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(stockAdjustmentActionHandler(transitionStockAdjustment, "submit")),
		),
	)
	mux.Handle(
		"/api/v1/stock-adjustments/{stock_adjustment_id}/approve",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(stockAdjustmentActionHandler(transitionStockAdjustment, "approve")),
		),
	)
	mux.Handle(
		"/api/v1/stock-adjustments/{stock_adjustment_id}/reject",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(stockAdjustmentActionHandler(transitionStockAdjustment, "reject")),
		),
	)
	mux.Handle(
		"/api/v1/stock-adjustments/{stock_adjustment_id}/post",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(stockAdjustmentActionHandler(transitionStockAdjustment, "post")),
		),
	)
	mux.Handle(
		"/api/v1/stock-counts",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(stockCountsHandler(listStockCounts, createStockCount)),
		),
	)
	mux.Handle(
		"/api/v1/stock-counts/{stock_count_id}/submit",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(stockCountSubmitHandler(submitStockCount)),
		),
	)
	mux.Handle(
		"/api/v1/inventory/batches",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionInventoryView,
			http.HandlerFunc(batchesHandler(batchCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/inventory/batches/{batch_id}/qc-transitions",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(batchQCTransitionsHandler(batchCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/inventory/batches/{batch_id}",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionInventoryView,
			http.HandlerFunc(batchDetailHandler(batchCatalog)),
		),
	)
	mux.Handle(
		"/api/v1/goods-receipts",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(goodsReceiptsHandler(warehouseReceiving)),
		),
	)
	mux.Handle(
		"/api/v1/goods-receipts/{receipt_id}/submit",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(submitGoodsReceiptHandler(warehouseReceiving)),
		),
	)
	mux.Handle(
		"/api/v1/goods-receipts/{receipt_id}/inspect-ready",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(markGoodsReceiptInspectReadyHandler(warehouseReceiving)),
		),
	)
	mux.Handle(
		"/api/v1/goods-receipts/{receipt_id}/post",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(postGoodsReceiptHandler(warehouseReceiving)),
		),
	)
	mux.Handle(
		"/api/v1/goods-receipts/{receipt_id}",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionWarehouseView,
			http.HandlerFunc(goodsReceiptDetailHandler(warehouseReceiving)),
		),
	)
	mux.Handle(
		"/api/v1/inbound-qc-inspections",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inboundQCInspectionsHandler(inboundQCInspections)),
		),
	)
	mux.Handle(
		"/api/v1/inbound-qc-inspections/{inspection_id}/start",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inboundQCInspectionStartHandler(inboundQCInspections)),
		),
	)
	mux.Handle(
		"/api/v1/inbound-qc-inspections/{inspection_id}/pass",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inboundQCInspectionPassHandler(inboundQCInspections)),
		),
	)
	mux.Handle(
		"/api/v1/inbound-qc-inspections/{inspection_id}/fail",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inboundQCInspectionFailHandler(inboundQCInspections)),
		),
	)
	mux.Handle(
		"/api/v1/inbound-qc-inspections/{inspection_id}/partial",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inboundQCInspectionPartialHandler(inboundQCInspections)),
		),
	)
	mux.Handle(
		"/api/v1/inbound-qc-inspections/{inspection_id}/hold",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inboundQCInspectionHoldHandler(inboundQCInspections)),
		),
	)
	mux.Handle(
		"/api/v1/inbound-qc-inspections/{inspection_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(inboundQCInspectionDetailHandler(inboundQCInspections)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-rejections",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(supplierRejectionsHandler(listSupplierRejections, createSupplierRejection)),
		),
	)
	mux.Handle(
		"/api/v1/supplier-rejections/{supplier_rejection_id}/submit",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(supplierRejectionActionHandler(transitionSupplierRejection, "submit")),
		),
	)
	mux.Handle(
		"/api/v1/supplier-rejections/{supplier_rejection_id}/confirm",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(supplierRejectionActionHandler(transitionSupplierRejection, "confirm")),
		),
	)
	mux.Handle(
		"/api/v1/supplier-rejections/{supplier_rejection_id}",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionWarehouseView,
			http.HandlerFunc(supplierRejectionDetailHandler(supplierRejectionStore)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse/end-of-day-reconciliations",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionWarehouseView,
			http.HandlerFunc(endOfDayReconciliationsHandler(listEndOfDayReconciliations)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse/end-of-day-reconciliations/{reconciliation_id}/close",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(closeEndOfDayReconciliationHandler(closeEndOfDayReconciliation)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse/daily-board/fulfillment-metrics",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionWarehouseView,
			http.HandlerFunc(warehouseDailyBoardFulfillmentMetricsHandler(salesOrderService, listCarrierManifests)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse/daily-board/inbound-metrics",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionWarehouseView,
			http.HandlerFunc(warehouseDailyBoardInboundMetricsHandler(
				purchaseOrderService,
				warehouseReceiving,
				inboundQCInspections,
				listSupplierRejections,
			)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse/daily-board/subcontract-metrics",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionWarehouseView,
			http.HandlerFunc(warehouseDailyBoardSubcontractMetricsHandler(
				subcontractOrderService,
				subcontractMaterialTransferStore,
				subcontractFactoryClaimStore,
				subcontractPaymentMilestoneStore,
			)),
		),
	)
	mux.Handle(
		"/api/v1/pick-tasks",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(pickTasksHandler(listPickTasks)),
		),
	)
	mux.Handle(
		"/api/v1/pick-tasks/{pick_task_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(pickTaskDetailHandler(getPickTask)),
		),
	)
	mux.Handle(
		"/api/v1/pick-tasks/{pick_task_id}/start",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(startPickTaskHandler(startPickTask)),
		),
	)
	mux.Handle(
		"/api/v1/pick-tasks/{pick_task_id}/confirm-line",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(confirmPickTaskLineHandler(confirmPickTaskLine)),
		),
	)
	mux.Handle(
		"/api/v1/pick-tasks/{pick_task_id}/complete",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(completePickTaskHandler(completePickTask)),
		),
	)
	mux.Handle(
		"/api/v1/pick-tasks/{pick_task_id}/exception",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(reportPickTaskExceptionHandler(reportPickTaskException)),
		),
	)
	mux.Handle(
		"/api/v1/pack-tasks",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(packTasksHandler(listPackTasks)),
		),
	)
	mux.Handle(
		"/api/v1/pack-tasks/{pack_task_id}",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(packTaskDetailHandler(getPackTask)),
		),
	)
	mux.Handle(
		"/api/v1/pack-tasks/{pack_task_id}/start",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(startPackTaskHandler(startPackTask)),
		),
	)
	mux.Handle(
		"/api/v1/pack-tasks/{pack_task_id}/confirm",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(confirmPackTaskHandler(confirmPackTask)),
		),
	)
	mux.Handle(
		"/api/v1/pack-tasks/{pack_task_id}/exception",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(reportPackTaskExceptionHandler(reportPackTaskException)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(carrierManifestsHandler(listCarrierManifests, createCarrierManifest)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/shipments",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(addShipmentToCarrierManifestHandler(addShipmentToCarrierManifest)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/shipments/{shipment_id}",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(removeShipmentFromCarrierManifestHandler(removeShipmentFromCarrierManifest)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/ready",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(markCarrierManifestReadyToScanHandler(markCarrierManifestReadyToScan)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/cancel",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(cancelCarrierManifestHandler(cancelCarrierManifest)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/exceptions",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(reportCarrierManifestMissingOrdersHandler(reportCarrierManifestMissingOrders)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/confirm-handover",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionRecordCreate,
			http.HandlerFunc(confirmCarrierManifestHandoverHandler(confirmCarrierManifestHandover)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/scan",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionShippingView,
			http.HandlerFunc(verifyCarrierManifestScanHandler(verifyCarrierManifestScan)),
		),
	)
	mux.Handle(
		"/api/v1/return-reasons",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(returnMasterDataHandler(listReturnMasterData)),
		),
	)
	mux.Handle(
		"/api/v1/returns/scan",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(returnScanHandler(receiveReturn)),
		),
	)
	mux.Handle(
		"/api/v1/returns/receipts",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(returnReceiptsHandler(listReturnReceipts, receiveReturn)),
		),
	)
	mux.Handle(
		"/api/v1/returns/{return_receipt_id}/inspect",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(returnInspectionHandler(inspectReturn)),
		),
	)
	mux.Handle(
		"/api/v1/returns/{return_receipt_id}/disposition",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(returnDispositionHandler(applyReturnDisposition)),
		),
	)
	mux.Handle(
		"/api/v1/returns/{return_receipt_id}/attachments",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(returnAttachmentHandler(uploadReturnAttachment)),
		),
	)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           accessLogMiddleware(mux, log.Default()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("api listening on :%s", cfg.AppPort)
	log.Fatal(server.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteSuccess(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   "api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteSuccess(w, r, http.StatusOK, healthResponse{
		Status:    "ready",
		Service:   "api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

type statusRecordingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (w *statusRecordingResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusRecordingResponseWriter) Write(body []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	written, err := w.ResponseWriter.Write(body)
	w.bytes += written
	return written, err
}

func accessLogMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecordingResponseWriter{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		statusCode := recorder.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		logger.Printf(
			"access method=%s path=%s status=%d bytes=%d duration_ms=%d remote=%s request_id=%s",
			r.Method,
			r.URL.Path,
			statusCode,
			recorder.bytes,
			time.Since(startedAt).Milliseconds(),
			r.RemoteAddr,
			r.Header.Get(response.HeaderRequestID),
		)
	})
}

func mockLoginHandler(authConfig auth.MockConfig) http.HandlerFunc {
	return loginHandler(auth.NewSessionManager(authConfig, time.Now))
}

func loginHandler(authSessions *auth.SessionManager, auditStores ...audit.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload loginRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid login payload",
				nil,
			)
			return
		}

		session, failure, ok := authSessions.Login(payload.Email, payload.Password)
		if !ok {
			recordAuthAudit(r, firstAuditStore(auditStores), "anonymous", "auth.login_failed", payload.Email, map[string]any{
				"reason": string(failure.Code),
			})
			details := map[string]any{"reason": string(failure.Code)}
			if !failure.LockedUntil.IsZero() {
				details["locked_until"] = failure.LockedUntil.Format(time.RFC3339)
			}
			response.WriteError(
				w,
				r,
				http.StatusUnauthorized,
				response.ErrorCodeUnauthorized,
				failure.Message,
				details,
			)
			return
		}

		recordAuthAudit(r, firstAuditStore(auditStores), session.Principal.UserID, "auth.login_succeeded", session.Principal.UserID, map[string]any{
			"email": session.Principal.Email,
			"role":  string(session.Principal.Role),
		})
		response.WriteSuccess(w, r, http.StatusOK, newLoginResponse(session, time.Now().UTC()))
	}
}

func refreshHandler(authSessions *auth.SessionManager, auditStores ...audit.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid refresh payload",
				nil,
			)
			return
		}

		session, ok := authSessions.Refresh(payload.RefreshToken)
		if !ok {
			recordAuthAudit(r, firstAuditStore(auditStores), "anonymous", "auth.refresh_failed", "unknown", nil)
			response.WriteError(
				w,
				r,
				http.StatusUnauthorized,
				response.ErrorCodeUnauthorized,
				"Invalid or expired refresh token",
				nil,
			)
			return
		}

		recordAuthAudit(r, firstAuditStore(auditStores), session.Principal.UserID, "auth.refresh_succeeded", session.Principal.UserID, map[string]any{
			"email": session.Principal.Email,
			"role":  string(session.Principal.Role),
		})
		response.WriteSuccess(w, r, http.StatusOK, newLoginResponse(session, time.Now().UTC()))
	}
}

func authPolicyHandler(authSessions *auth.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		passwordPolicy := authSessions.PasswordPolicy()
		lockoutPolicy := authSessions.LockoutPolicy()
		response.WriteSuccess(w, r, http.StatusOK, authPolicyResponse{
			PasswordMinLength:              passwordPolicy.MinLength,
			PasswordRequiresLetter:         passwordPolicy.RequireLetter,
			PasswordRequiresNumberOrSymbol: passwordPolicy.RequireNumberOrSymbol,
			CommonPasswordsBlocked:         passwordPolicy.CommonPasswordsBlocked,
			MaxFailedAttempts:              lockoutPolicy.MaxFailedAttempts,
			LockoutWindowSeconds:           int(lockoutPolicy.Window.Seconds()),
			LockoutDurationSeconds:         int(lockoutPolicy.Duration.Seconds()),
		})
	}
}

func firstAuditStore(stores []audit.LogStore) audit.LogStore {
	if len(stores) == 0 {
		return nil
	}
	return stores[0]
}

func recordAuthAudit(r *http.Request, store audit.LogStore, actorID string, action string, entityID string, metadata map[string]any) {
	if store == nil {
		return
	}

	log, err := audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    actorID,
		Action:     action,
		EntityType: "auth.session",
		EntityID:   strings.TrimSpace(entityID),
		RequestID:  response.RequestID(r),
		Metadata:   metadata,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return
	}

	_ = store.Record(r.Context(), log)
}

func newLoginResponse(session auth.Session, now time.Time) loginResponse {
	expiresIn := int(session.AccessExpiresAt.Sub(now).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}
	refreshExpiresIn := int(session.RefreshExpiresAt.Sub(now).Seconds())
	if refreshExpiresIn < 0 {
		refreshExpiresIn = 0
	}

	return loginResponse{
		AccessToken:      session.AccessToken,
		RefreshToken:     session.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
		ExpiresAt:        session.AccessExpiresAt.Format(time.RFC3339),
		User:             newUserResponse(session.Principal),
	}
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
		return
	}

	response.WriteSuccess(w, r, http.StatusOK, newUserResponse(principal))
}

func newUserResponse(principal auth.Principal) userResponse {
	return userResponse{
		ID:          principal.UserID,
		Email:       principal.Email,
		Name:        principal.Name,
		Role:        string(principal.Role),
		Permissions: permissionStrings(principal.Permissions),
	}
}

func rbacRolesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		return
	}

	roles := auth.RoleCatalog()
	payload := make([]roleResponse, 0, len(roles))
	for _, role := range roles {
		payload = append(payload, roleResponse{
			Key:         string(role.Key),
			Name:        role.Name,
			Permissions: permissionStrings(role.Permissions),
		})
	}

	response.WriteSuccess(w, r, http.StatusOK, payload)
}

func rbacPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		return
	}

	permissions := auth.PermissionCatalog()
	payload := make([]permissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		payload = append(payload, permissionResponse{
			Key:   string(permission.Key),
			Name:  permission.Name,
			Group: permission.Group,
		})
	}

	response.WriteSuccess(w, r, http.StatusOK, payload)
}

func permissionStrings(permissions []auth.PermissionKey) []string {
	values := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		values = append(values, string(permission))
	}

	return values
}

func auditLogsHandler(store audit.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		logs, err := store.List(r.Context(), audit.Query{
			ActorID:    r.URL.Query().Get("actor_id"),
			Action:     r.URL.Query().Get("action"),
			EntityType: r.URL.Query().Get("entity_type"),
			EntityID:   r.URL.Query().Get("entity_id"),
			Limit:      queryInt(r, "limit"),
		})
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Audit logs could not be loaded",
				nil,
			)
			return
		}

		payload := make([]auditLogResponse, 0, len(logs))
		for _, log := range logs {
			payload = append(payload, newAuditLogResponse(log))
		}

		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func productsHandler(catalog *masterdataapp.ItemCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			filter := masterdatadomain.NewItemFilter(
				r.URL.Query().Get("q"),
				masterdatadomain.ItemStatus(r.URL.Query().Get("status")),
				masterdatadomain.ItemType(r.URL.Query().Get("item_type")),
				queryInt(r, "page"),
				queryInt(r, "page_size"),
			)
			items, pagination, err := catalog.List(r.Context(), filter)
			if err != nil {
				writeProductError(w, r, err)
				return
			}

			payload := make([]productResponse, 0, len(items))
			for _, item := range items {
				payload = append(payload, newProductResponse(item, ""))
			}
			response.WritePaginatedSuccess(w, r, http.StatusOK, payload, pagination)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload productRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid product master data payload",
					nil,
				)
				return
			}

			result, err := catalog.Create(r.Context(), masterdataapp.CreateItemInput{
				ItemCode:         payload.ItemCode,
				SKUCode:          payload.SKUCode,
				Name:             payload.Name,
				Type:             payload.ItemType,
				Group:            payload.ItemGroup,
				BrandCode:        payload.BrandCode,
				UOMBase:          payload.UOMBase,
				UOMPurchase:      payload.UOMPurchase,
				UOMIssue:         payload.UOMIssue,
				LotControlled:    payload.LotControlled,
				ExpiryControlled: payload.ExpiryControlled,
				ShelfLifeDays:    payload.ShelfLifeDays,
				QCRequired:       payload.QCRequired,
				Status:           payload.Status,
				StandardCost:     decimal.Decimal(payload.StandardCost),
				IsSellable:       payload.IsSellable,
				IsPurchasable:    payload.IsPurchasable,
				IsProducible:     payload.IsProducible,
				SpecVersion:      payload.SpecVersion,
				ActorID:          principal.UserID,
				RequestID:        response.RequestID(r),
			})
			if err != nil {
				writeProductError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newProductResponse(result.Item, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func productDetailHandler(catalog *masterdataapp.ItemCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			item, err := catalog.Get(r.Context(), r.PathValue("product_id"))
			if err != nil {
				writeProductError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newProductResponse(item, ""))
		case http.MethodPatch:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload productRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid product master data payload",
					nil,
				)
				return
			}

			result, err := catalog.Update(r.Context(), masterdataapp.UpdateItemInput{
				ID:               r.PathValue("product_id"),
				ItemCode:         payload.ItemCode,
				SKUCode:          payload.SKUCode,
				Name:             payload.Name,
				Type:             payload.ItemType,
				Group:            payload.ItemGroup,
				BrandCode:        payload.BrandCode,
				UOMBase:          payload.UOMBase,
				UOMPurchase:      payload.UOMPurchase,
				UOMIssue:         payload.UOMIssue,
				LotControlled:    payload.LotControlled,
				ExpiryControlled: payload.ExpiryControlled,
				ShelfLifeDays:    payload.ShelfLifeDays,
				QCRequired:       payload.QCRequired,
				Status:           payload.Status,
				StandardCost:     decimal.Decimal(payload.StandardCost),
				IsSellable:       payload.IsSellable,
				IsPurchasable:    payload.IsPurchasable,
				IsProducible:     payload.IsProducible,
				SpecVersion:      payload.SpecVersion,
				ActorID:          principal.UserID,
				RequestID:        response.RequestID(r),
			})
			if err != nil {
				writeProductError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newProductResponse(result.Item, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func changeProductStatusHandler(catalog *masterdataapp.ItemCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload changeProductStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid product status payload",
				nil,
			)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		result, err := catalog.ChangeStatus(r.Context(), masterdataapp.ChangeItemStatusInput{
			ID:        r.PathValue("product_id"),
			Status:    payload.Status,
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writeProductError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newProductResponse(result.Item, result.AuditLogID))
	}
}

func warehousesHandler(catalog *masterdataapp.WarehouseLocationCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			filter := masterdatadomain.NewWarehouseFilter(
				r.URL.Query().Get("q"),
				masterdatadomain.WarehouseStatus(r.URL.Query().Get("status")),
				masterdatadomain.WarehouseType(r.URL.Query().Get("warehouse_type")),
				queryInt(r, "page"),
				queryInt(r, "page_size"),
			)
			warehouses, pagination, err := catalog.ListWarehouses(r.Context(), filter)
			if err != nil {
				writeWarehouseError(w, r, err)
				return
			}

			payload := make([]warehouseResponse, 0, len(warehouses))
			for _, warehouse := range warehouses {
				payload = append(payload, newWarehouseResponse(warehouse, ""))
			}
			response.WritePaginatedSuccess(w, r, http.StatusOK, payload, pagination)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload warehouseRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid warehouse master data payload",
					nil,
				)
				return
			}

			result, err := catalog.CreateWarehouse(r.Context(), masterdataapp.CreateWarehouseInput{
				Code:            payload.WarehouseCode,
				Name:            payload.WarehouseName,
				Type:            payload.WarehouseType,
				SiteCode:        payload.SiteCode,
				Address:         payload.Address,
				AllowSaleIssue:  payload.AllowSaleIssue,
				AllowProdIssue:  payload.AllowProdIssue,
				AllowQuarantine: payload.AllowQuarantine,
				Status:          payload.Status,
				ActorID:         principal.UserID,
				RequestID:       response.RequestID(r),
			})
			if err != nil {
				writeWarehouseError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newWarehouseResponse(result.Warehouse, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func warehouseDetailHandler(catalog *masterdataapp.WarehouseLocationCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			warehouse, err := catalog.GetWarehouse(r.Context(), r.PathValue("warehouse_id"))
			if err != nil {
				writeWarehouseError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newWarehouseResponse(warehouse, ""))
		case http.MethodPatch:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload warehouseRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid warehouse master data payload",
					nil,
				)
				return
			}

			result, err := catalog.UpdateWarehouse(r.Context(), masterdataapp.UpdateWarehouseInput{
				ID:              r.PathValue("warehouse_id"),
				Code:            payload.WarehouseCode,
				Name:            payload.WarehouseName,
				Type:            payload.WarehouseType,
				SiteCode:        payload.SiteCode,
				Address:         payload.Address,
				AllowSaleIssue:  payload.AllowSaleIssue,
				AllowProdIssue:  payload.AllowProdIssue,
				AllowQuarantine: payload.AllowQuarantine,
				Status:          payload.Status,
				ActorID:         principal.UserID,
				RequestID:       response.RequestID(r),
			})
			if err != nil {
				writeWarehouseError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newWarehouseResponse(result.Warehouse, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func changeWarehouseStatusHandler(catalog *masterdataapp.WarehouseLocationCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload changeWarehouseStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid warehouse status payload",
				nil,
			)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		result, err := catalog.ChangeWarehouseStatus(r.Context(), masterdataapp.ChangeWarehouseStatusInput{
			ID:        r.PathValue("warehouse_id"),
			Status:    payload.Status,
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writeWarehouseError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newWarehouseResponse(result.Warehouse, result.AuditLogID))
	}
}

func warehouseLocationsHandler(catalog *masterdataapp.WarehouseLocationCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			filter := masterdatadomain.NewLocationFilter(
				r.URL.Query().Get("q"),
				r.URL.Query().Get("warehouse_id"),
				masterdatadomain.LocationStatus(r.URL.Query().Get("status")),
				masterdatadomain.LocationType(r.URL.Query().Get("location_type")),
				queryInt(r, "page"),
				queryInt(r, "page_size"),
			)
			locations, pagination, err := catalog.ListLocations(r.Context(), filter)
			if err != nil {
				writeWarehouseError(w, r, err)
				return
			}

			payload := make([]warehouseLocationResponse, 0, len(locations))
			for _, location := range locations {
				payload = append(payload, newWarehouseLocationResponse(location, ""))
			}
			response.WritePaginatedSuccess(w, r, http.StatusOK, payload, pagination)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload warehouseLocationRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid warehouse location payload",
					nil,
				)
				return
			}

			result, err := catalog.CreateLocation(r.Context(), masterdataapp.CreateLocationInput{
				WarehouseID:  payload.WarehouseID,
				Code:         payload.LocationCode,
				Name:         payload.LocationName,
				Type:         payload.LocationType,
				ZoneCode:     payload.ZoneCode,
				AllowReceive: payload.AllowReceive,
				AllowPick:    payload.AllowPick,
				AllowStore:   payload.AllowStore,
				IsDefault:    payload.IsDefault,
				Status:       payload.Status,
				ActorID:      principal.UserID,
				RequestID:    response.RequestID(r),
			})
			if err != nil {
				writeWarehouseError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newWarehouseLocationResponse(result.Location, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func warehouseLocationDetailHandler(catalog *masterdataapp.WarehouseLocationCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			location, err := catalog.GetLocation(r.Context(), r.PathValue("location_id"))
			if err != nil {
				writeWarehouseError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newWarehouseLocationResponse(location, ""))
		case http.MethodPatch:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload warehouseLocationRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid warehouse location payload",
					nil,
				)
				return
			}

			result, err := catalog.UpdateLocation(r.Context(), masterdataapp.UpdateLocationInput{
				ID:           r.PathValue("location_id"),
				WarehouseID:  payload.WarehouseID,
				Code:         payload.LocationCode,
				Name:         payload.LocationName,
				Type:         payload.LocationType,
				ZoneCode:     payload.ZoneCode,
				AllowReceive: payload.AllowReceive,
				AllowPick:    payload.AllowPick,
				AllowStore:   payload.AllowStore,
				IsDefault:    payload.IsDefault,
				Status:       payload.Status,
				ActorID:      principal.UserID,
				RequestID:    response.RequestID(r),
			})
			if err != nil {
				writeWarehouseError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newWarehouseLocationResponse(result.Location, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func changeWarehouseLocationStatusHandler(catalog *masterdataapp.WarehouseLocationCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload changeWarehouseLocationStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid warehouse location status payload",
				nil,
			)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		result, err := catalog.ChangeLocationStatus(r.Context(), masterdataapp.ChangeLocationStatusInput{
			ID:        r.PathValue("location_id"),
			Status:    payload.Status,
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writeWarehouseError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newWarehouseLocationResponse(result.Location, result.AuditLogID))
	}
}

func suppliersHandler(catalog *masterdataapp.PartyCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			filter := masterdatadomain.NewSupplierFilter(
				r.URL.Query().Get("q"),
				masterdatadomain.SupplierStatus(r.URL.Query().Get("status")),
				masterdatadomain.SupplierGroup(r.URL.Query().Get("supplier_group")),
				queryInt(r, "page"),
				queryInt(r, "page_size"),
			)
			suppliers, pagination, err := catalog.ListSuppliers(r.Context(), filter)
			if err != nil {
				writePartyError(w, r, err)
				return
			}

			payload := make([]supplierResponse, 0, len(suppliers))
			for _, supplier := range suppliers {
				payload = append(payload, newSupplierResponse(supplier, ""))
			}
			response.WritePaginatedSuccess(w, r, http.StatusOK, payload, pagination)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload supplierRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid supplier master data payload",
					nil,
				)
				return
			}

			result, err := catalog.CreateSupplier(r.Context(), masterdataapp.CreateSupplierInput{
				Code:          payload.SupplierCode,
				Name:          payload.SupplierName,
				Group:         payload.SupplierGroup,
				ContactName:   payload.ContactName,
				Phone:         payload.Phone,
				Email:         payload.Email,
				TaxCode:       payload.TaxCode,
				Address:       payload.Address,
				PaymentTerms:  payload.PaymentTerms,
				LeadTimeDays:  payload.LeadTimeDays,
				MOQ:           decimal.Decimal(payload.MOQ),
				QualityScore:  decimal.Decimal(payload.QualityScore),
				DeliveryScore: decimal.Decimal(payload.DeliveryScore),
				Status:        payload.Status,
				ActorID:       principal.UserID,
				RequestID:     response.RequestID(r),
			})
			if err != nil {
				writePartyError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newSupplierResponse(result.Supplier, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func supplierDetailHandler(catalog *masterdataapp.PartyCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			supplier, err := catalog.GetSupplier(r.Context(), r.PathValue("supplier_id"))
			if err != nil {
				writePartyError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newSupplierResponse(supplier, ""))
		case http.MethodPatch:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload supplierRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid supplier master data payload",
					nil,
				)
				return
			}

			result, err := catalog.UpdateSupplier(r.Context(), masterdataapp.UpdateSupplierInput{
				ID:            r.PathValue("supplier_id"),
				Code:          payload.SupplierCode,
				Name:          payload.SupplierName,
				Group:         payload.SupplierGroup,
				ContactName:   payload.ContactName,
				Phone:         payload.Phone,
				Email:         payload.Email,
				TaxCode:       payload.TaxCode,
				Address:       payload.Address,
				PaymentTerms:  payload.PaymentTerms,
				LeadTimeDays:  payload.LeadTimeDays,
				MOQ:           decimal.Decimal(payload.MOQ),
				QualityScore:  decimal.Decimal(payload.QualityScore),
				DeliveryScore: decimal.Decimal(payload.DeliveryScore),
				Status:        payload.Status,
				ActorID:       principal.UserID,
				RequestID:     response.RequestID(r),
			})
			if err != nil {
				writePartyError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newSupplierResponse(result.Supplier, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func changeSupplierStatusHandler(catalog *masterdataapp.PartyCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload changeSupplierStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid supplier status payload",
				nil,
			)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		result, err := catalog.ChangeSupplierStatus(r.Context(), masterdataapp.ChangeSupplierStatusInput{
			ID:        r.PathValue("supplier_id"),
			Status:    payload.Status,
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writePartyError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newSupplierResponse(result.Supplier, result.AuditLogID))
	}
}

func customersHandler(catalog *masterdataapp.PartyCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			filter := masterdatadomain.NewCustomerFilter(
				r.URL.Query().Get("q"),
				masterdatadomain.CustomerStatus(r.URL.Query().Get("status")),
				masterdatadomain.CustomerType(r.URL.Query().Get("customer_type")),
				queryInt(r, "page"),
				queryInt(r, "page_size"),
			)
			customers, pagination, err := catalog.ListCustomers(r.Context(), filter)
			if err != nil {
				writePartyError(w, r, err)
				return
			}

			payload := make([]customerResponse, 0, len(customers))
			for _, customer := range customers {
				payload = append(payload, newCustomerResponse(customer, ""))
			}
			response.WritePaginatedSuccess(w, r, http.StatusOK, payload, pagination)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload customerRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid customer master data payload",
					nil,
				)
				return
			}

			result, err := catalog.CreateCustomer(r.Context(), masterdataapp.CreateCustomerInput{
				Code:          payload.CustomerCode,
				Name:          payload.CustomerName,
				Type:          payload.CustomerType,
				ChannelCode:   payload.ChannelCode,
				PriceListCode: payload.PriceListCode,
				DiscountGroup: payload.DiscountGroup,
				CreditLimit:   decimal.Decimal(payload.CreditLimit),
				PaymentTerms:  payload.PaymentTerms,
				ContactName:   payload.ContactName,
				Phone:         payload.Phone,
				Email:         payload.Email,
				TaxCode:       payload.TaxCode,
				Address:       payload.Address,
				Status:        payload.Status,
				ActorID:       principal.UserID,
				RequestID:     response.RequestID(r),
			})
			if err != nil {
				writePartyError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newCustomerResponse(result.Customer, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func customerDetailHandler(catalog *masterdataapp.PartyCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionMasterDataView) {
				writePermissionDenied(w, r, auth.PermissionMasterDataView)
				return
			}
			customer, err := catalog.GetCustomer(r.Context(), r.PathValue("customer_id"))
			if err != nil {
				writePartyError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newCustomerResponse(customer, ""))
		case http.MethodPatch:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload customerRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid customer master data payload",
					nil,
				)
				return
			}

			result, err := catalog.UpdateCustomer(r.Context(), masterdataapp.UpdateCustomerInput{
				ID:            r.PathValue("customer_id"),
				Code:          payload.CustomerCode,
				Name:          payload.CustomerName,
				Type:          payload.CustomerType,
				ChannelCode:   payload.ChannelCode,
				PriceListCode: payload.PriceListCode,
				DiscountGroup: payload.DiscountGroup,
				CreditLimit:   decimal.Decimal(payload.CreditLimit),
				PaymentTerms:  payload.PaymentTerms,
				ContactName:   payload.ContactName,
				Phone:         payload.Phone,
				Email:         payload.Email,
				TaxCode:       payload.TaxCode,
				Address:       payload.Address,
				Status:        payload.Status,
				ActorID:       principal.UserID,
				RequestID:     response.RequestID(r),
			})
			if err != nil {
				writePartyError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, newCustomerResponse(result.Customer, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func changeCustomerStatusHandler(catalog *masterdataapp.PartyCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload changeCustomerStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid customer status payload",
				nil,
			)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		result, err := catalog.ChangeCustomerStatus(r.Context(), masterdataapp.ChangeCustomerStatusInput{
			ID:        r.PathValue("customer_id"),
			Status:    payload.Status,
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writePartyError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newCustomerResponse(result.Customer, result.AuditLogID))
	}
}

func salesOrdersHandler(service salesapp.SalesOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSalesView) {
			writePermissionDenied(w, r, auth.PermissionSalesView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			orders, err := service.ListSalesOrders(r.Context(), salesOrderFilterFromRequest(r))
			if err != nil {
				writeSalesOrderError(w, r, err)
				return
			}

			payload := make([]salesOrderListItemResponse, 0, len(orders))
			for _, order := range orders {
				payload = append(payload, newSalesOrderListItemResponse(order))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload createSalesOrderRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid sales order payload", nil)
				return
			}

			result, err := service.CreateSalesOrder(r.Context(), salesapp.CreateSalesOrderInput{
				ID:           payload.ID,
				OrderNo:      payload.OrderNo,
				CustomerID:   payload.CustomerID,
				Channel:      payload.Channel,
				WarehouseID:  payload.WarehouseID,
				OrderDate:    payload.OrderDate,
				CurrencyCode: payload.CurrencyCode,
				Note:         payload.Note,
				Lines:        salesOrderLineInputs(payload.Lines),
				ActorID:      principal.UserID,
				RequestID:    response.RequestID(r),
			})
			if err != nil {
				writeSalesOrderError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newSalesOrderResponse(result.SalesOrder, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func salesOrderDetailHandler(service salesapp.SalesOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSalesView) {
			writePermissionDenied(w, r, auth.PermissionSalesView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			order, err := service.GetSalesOrder(r.Context(), r.PathValue("sales_order_id"))
			if err != nil {
				writeSalesOrderError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusOK, newSalesOrderResponse(order, ""))
		case http.MethodPatch:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload updateSalesOrderRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid sales order payload", nil)
				return
			}

			result, err := service.UpdateSalesOrder(r.Context(), salesapp.UpdateSalesOrderInput{
				ID:              r.PathValue("sales_order_id"),
				CustomerID:      payload.CustomerID,
				Channel:         payload.Channel,
				WarehouseID:     payload.WarehouseID,
				OrderDate:       payload.OrderDate,
				Note:            payload.Note,
				Lines:           salesOrderLineInputs(payload.Lines),
				ExpectedVersion: payload.ExpectedVersion,
				ActorID:         principal.UserID,
				RequestID:       response.RequestID(r),
			})
			if err != nil {
				writeSalesOrderError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusOK, newSalesOrderResponse(result.SalesOrder, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func salesOrderConfirmHandler(service salesapp.SalesOrderService) http.HandlerFunc {
	return salesOrderActionHandler(service, "confirm")
}

func salesOrderCancelHandler(service salesapp.SalesOrderService) http.HandlerFunc {
	return salesOrderActionHandler(service, "cancel")
}

func salesOrderActionHandler(service salesapp.SalesOrderService, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSalesView) {
			writePermissionDenied(w, r, auth.PermissionSalesView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload salesOrderActionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				if !errors.Is(err, io.EOF) {
					response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid sales order action payload", nil)
					return
				}
			}
		}

		input := salesapp.SalesOrderActionInput{
			ID:              r.PathValue("sales_order_id"),
			ExpectedVersion: payload.ExpectedVersion,
			Reason:          payload.Reason,
			Note:            payload.Note,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		}
		var (
			result salesapp.SalesOrderActionResult
			err    error
		)
		switch action {
		case "confirm":
			result, err = service.ConfirmSalesOrder(r.Context(), input)
		case "cancel":
			result, err = service.CancelSalesOrder(r.Context(), input)
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writeSalesOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, salesOrderActionResultResponse{
			SalesOrder:     newSalesOrderResponse(result.SalesOrder, ""),
			PreviousStatus: string(result.PreviousStatus),
			CurrentStatus:  string(result.CurrentStatus),
			AuditLogID:     result.AuditLogID,
		})
	}
}

func stockMovementHandler(store audit.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload stockMovementRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid stock movement payload",
				nil,
			)
			return
		}
		if details := validateStockMovementPayload(payload); len(details) > 0 {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid stock movement payload",
				details,
			)
			return
		}

		if strings.EqualFold(strings.TrimSpace(payload.MovementType), "ADJUST") {
			if err := recordStockAdjustmentAudit(r, store, payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusConflict,
					response.ErrorCodeConflict,
					"Audit log could not be recorded",
					nil,
				)
				return
			}
		}

		movementQty, sourceQty, conversionFactor, baseUOMCode, sourceUOMCode := stockMovementContractValues(payload)
		response.WriteSuccess(w, r, http.StatusCreated, stockMovementResponse{
			MovementID:       strings.TrimSpace(payload.MovementID),
			Status:           "recorded",
			MovementQuantity: movementQty.String(),
			BaseUOMCode:      baseUOMCode.String(),
			SourceQuantity:   sourceQty.String(),
			SourceUOMCode:    sourceUOMCode.String(),
			ConversionFactor: conversionFactor.String(),
		})
	}
}

func availableStockHandler(service inventoryapp.ListAvailableStock) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		filter := domain.NewAvailableStockFilter(
			r.URL.Query().Get("warehouse_id"),
			r.URL.Query().Get("location_id"),
			r.URL.Query().Get("sku"),
			r.URL.Query().Get("batch_id"),
		)
		snapshots, err := service.Execute(r.Context(), filter)
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Available stock could not be calculated",
				nil,
			)
			return
		}

		payload := make([]availableStockResponse, 0, len(snapshots))
		for _, snapshot := range snapshots {
			payload = append(payload, newAvailableStockResponse(snapshot))
		}

		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func stockAdjustmentsHandler(
	listService inventoryapp.ListStockAdjustments,
	createService inventoryapp.CreateStockAdjustment,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionInventoryView) {
				writePermissionDenied(w, r, auth.PermissionInventoryView)
				return
			}
			rows, err := listService.Execute(r.Context())
			if err != nil {
				response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Stock adjustments could not be listed", nil)
				return
			}
			payload := make([]stockAdjustmentResponse, 0, len(rows))
			for _, row := range rows {
				payload = append(payload, newStockAdjustmentResponse(row, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			var payload createStockAdjustmentRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid stock adjustment payload",
					nil,
				)
				return
			}
			result, err := createService.Execute(r.Context(), inventoryapp.CreateStockAdjustmentInput{
				ID:            payload.ID,
				AdjustmentNo:  payload.AdjustmentNo,
				OrgID:         payload.OrgID,
				WarehouseID:   payload.WarehouseID,
				WarehouseCode: payload.WarehouseCode,
				SourceType:    payload.SourceType,
				SourceID:      payload.SourceID,
				Reason:        payload.Reason,
				RequestedBy:   principal.UserID,
				RequestID:     response.RequestID(r),
				Lines:         newCreateStockAdjustmentLines(payload.Lines),
			})
			if err != nil {
				writeStockAdjustmentError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newStockAdjustmentResponse(result.Adjustment, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func stockAdjustmentActionHandler(
	service inventoryapp.TransitionStockAdjustment,
	action string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		requiredPermission := auth.PermissionRecordCreate
		if action == "approve" || action == "reject" {
			requiredPermission = auth.PermissionApprovalsView
		}
		if !auth.HasPermission(principal, requiredPermission) {
			writePermissionDenied(w, r, requiredPermission)
			return
		}

		var result inventoryapp.StockAdjustmentResult
		var err error
		id := r.PathValue("stock_adjustment_id")
		switch action {
		case "submit":
			result, err = service.Submit(r.Context(), id, principal.UserID, response.RequestID(r))
		case "approve":
			result, err = service.Approve(r.Context(), id, principal.UserID, response.RequestID(r))
		case "reject":
			result, err = service.Reject(r.Context(), id, principal.UserID, response.RequestID(r))
		case "post":
			result, err = service.Post(r.Context(), id, principal.UserID, response.RequestID(r))
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writeStockAdjustmentError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newStockAdjustmentResponse(result.Adjustment, result.AuditLogID))
	}
}

func stockCountsHandler(
	listService inventoryapp.ListStockCounts,
	createService inventoryapp.CreateStockCount,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionInventoryView) {
				writePermissionDenied(w, r, auth.PermissionInventoryView)
				return
			}
			rows, err := listService.Execute(r.Context())
			if err != nil {
				response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Stock counts could not be listed", nil)
				return
			}
			payload := make([]stockCountResponse, 0, len(rows))
			for _, row := range rows {
				payload = append(payload, newStockCountResponse(row, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			var payload createStockCountRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid stock count payload", nil)
				return
			}
			result, err := createService.Execute(r.Context(), inventoryapp.CreateStockCountInput{
				ID:            payload.ID,
				CountNo:       payload.CountNo,
				OrgID:         payload.OrgID,
				WarehouseID:   payload.WarehouseID,
				WarehouseCode: payload.WarehouseCode,
				Scope:         payload.Scope,
				CreatedBy:     principal.UserID,
				RequestID:     response.RequestID(r),
				Lines:         newCreateStockCountLines(payload.Lines),
			})
			if err != nil {
				writeStockCountError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newStockCountResponse(result.Session, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func stockCountSubmitHandler(submitService inventoryapp.SubmitStockCount) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		var payload submitStockCountRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid stock count submit payload", nil)
			return
		}
		result, err := submitService.Execute(r.Context(), inventoryapp.SubmitStockCountInput{
			ID:          r.PathValue("stock_count_id"),
			SubmittedBy: principal.UserID,
			RequestID:   response.RequestID(r),
			Lines:       newSubmitStockCountLines(payload.Lines),
		})
		if err != nil {
			writeStockCountError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newStockCountResponse(result.Session, result.AuditLogID))
	}
}

func batchesHandler(catalog *inventoryapp.BatchCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		filter := domain.NewBatchFilter(
			r.URL.Query().Get("sku"),
			domain.QCStatus(r.URL.Query().Get("qc_status")),
			domain.BatchStatus(r.URL.Query().Get("status")),
		)
		batches, err := catalog.ListBatches(r.Context(), filter)
		if err != nil {
			response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Batches could not be loaded", nil)
			return
		}

		payload := make([]batchResponse, 0, len(batches))
		for _, batch := range batches {
			payload = append(payload, newBatchResponse(batch))
		}

		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func batchDetailHandler(catalog *inventoryapp.BatchCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		batch, err := catalog.GetBatch(r.Context(), r.PathValue("batch_id"))
		if err != nil {
			if errors.Is(err, inventoryapp.ErrBatchNotFound) {
				response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Batch not found", nil)
				return
			}
			response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Batch could not be loaded", nil)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newBatchResponse(batch))
	}
}

func batchQCTransitionsHandler(catalog *inventoryapp.BatchCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionInventoryView) {
				writePermissionDenied(w, r, auth.PermissionInventoryView)
				return
			}
			transitions, err := catalog.ListQCTransitions(r.Context(), r.PathValue("batch_id"))
			if err != nil {
				writeBatchQCTransitionError(w, r, err)
				return
			}

			payload := make([]batchQCTransitionResponse, 0, len(transitions))
			for _, transition := range transitions {
				payload = append(payload, newBatchQCTransitionResponse(transition))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionQCDecision) {
				writePermissionDenied(w, r, auth.PermissionQCDecision)
				return
			}

			var payload batchQCTransitionRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid batch QC transition payload",
					nil,
				)
				return
			}

			result, err := catalog.ChangeQCStatus(r.Context(), inventoryapp.ChangeBatchQCStatusInput{
				BatchID:     r.PathValue("batch_id"),
				NextStatus:  domain.QCStatus(payload.QCStatus),
				ActorID:     principal.UserID,
				Reason:      payload.Reason,
				BusinessRef: payload.BusinessRef,
				RequestID:   response.RequestID(r),
			})
			if err != nil {
				writeBatchQCTransitionError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusOK, batchQCTransitionResultResponse{
				Batch:      newBatchResponse(result.Batch),
				Transition: newBatchQCTransitionResponse(result.Transition),
			})
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func goodsReceiptsHandler(service inventoryapp.WarehouseReceivingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionWarehouseView) {
				writePermissionDenied(w, r, auth.PermissionWarehouseView)
				return
			}
			filter := domain.NewWarehouseReceivingFilter(
				r.URL.Query().Get("warehouse_id"),
				domain.WarehouseReceivingStatus(r.URL.Query().Get("status")),
			)
			receipts, err := service.ListWarehouseReceivings(r.Context(), filter)
			if err != nil {
				response.WriteError(
					w,
					r,
					http.StatusConflict,
					response.ErrorCodeConflict,
					"Goods receipts could not be loaded",
					nil,
				)
				return
			}

			payload := make([]warehouseReceivingResponse, 0, len(receipts))
			for _, receipt := range receipts {
				payload = append(payload, newWarehouseReceivingResponse(receipt, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}

			var payload createWarehouseReceivingRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid goods receipt payload",
					nil,
				)
				return
			}
			result, err := service.CreateWarehouseReceiving(r.Context(), inventoryapp.CreateWarehouseReceivingInput{
				ID:               payload.ID,
				OrgID:            payload.OrgID,
				ReceiptNo:        payload.ReceiptNo,
				WarehouseID:      payload.WarehouseID,
				LocationID:       payload.LocationID,
				ReferenceDocType: payload.ReferenceDocType,
				ReferenceDocID:   payload.ReferenceDocID,
				SupplierID:       payload.SupplierID,
				DeliveryNoteNo:   payload.DeliveryNoteNo,
				Lines:            newCreateWarehouseReceivingLines(payload.Lines),
				ActorID:          principal.UserID,
				RequestID:        response.RequestID(r),
			})
			if err != nil {
				writeWarehouseReceivingError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newWarehouseReceivingResponse(result.Receipt, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func goodsReceiptDetailHandler(service inventoryapp.WarehouseReceivingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		receipt, err := service.GetWarehouseReceiving(r.Context(), r.PathValue("receipt_id"))
		if err != nil {
			writeWarehouseReceivingError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newWarehouseReceivingResponse(receipt, ""))
	}
}

func submitGoodsReceiptHandler(service inventoryapp.WarehouseReceivingService) http.HandlerFunc {
	return goodsReceiptTransitionHandler(service, func(
		ctx context.Context,
		input inventoryapp.WarehouseReceivingTransitionInput,
	) (inventoryapp.WarehouseReceivingResult, error) {
		return service.SubmitWarehouseReceiving(ctx, input)
	})
}

func markGoodsReceiptInspectReadyHandler(service inventoryapp.WarehouseReceivingService) http.HandlerFunc {
	return goodsReceiptTransitionHandler(service, func(
		ctx context.Context,
		input inventoryapp.WarehouseReceivingTransitionInput,
	) (inventoryapp.WarehouseReceivingResult, error) {
		return service.MarkWarehouseReceivingInspectReady(ctx, input)
	})
}

func postGoodsReceiptHandler(service inventoryapp.WarehouseReceivingService) http.HandlerFunc {
	return goodsReceiptTransitionHandler(service, func(
		ctx context.Context,
		input inventoryapp.WarehouseReceivingTransitionInput,
	) (inventoryapp.WarehouseReceivingResult, error) {
		return service.PostWarehouseReceiving(ctx, input)
	})
}

func goodsReceiptTransitionHandler(
	_ inventoryapp.WarehouseReceivingService,
	apply func(context.Context, inventoryapp.WarehouseReceivingTransitionInput) (inventoryapp.WarehouseReceivingResult, error),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}
		result, err := apply(r.Context(), inventoryapp.WarehouseReceivingTransitionInput{
			ID:        r.PathValue("receipt_id"),
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writeWarehouseReceivingError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newWarehouseReceivingResponse(result.Receipt, result.AuditLogID))
	}
}

func endOfDayReconciliationsHandler(service inventoryapp.ListEndOfDayReconciliations) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		filter := domain.NewEndOfDayReconciliationFilter(
			r.URL.Query().Get("warehouse_id"),
			r.URL.Query().Get("date"),
			r.URL.Query().Get("shift_code"),
			domain.EndOfDayReconciliationStatus(r.URL.Query().Get("status")),
		)
		reconciliations, err := service.Execute(r.Context(), filter)
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"End-of-day reconciliations could not be loaded",
				nil,
			)
			return
		}

		payload := make([]endOfDayReconciliationResponse, 0, len(reconciliations))
		for _, reconciliation := range reconciliations {
			payload = append(payload, newEndOfDayReconciliationResponse(reconciliation, ""))
		}

		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func closeEndOfDayReconciliationHandler(service inventoryapp.CloseEndOfDayReconciliation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload closeReconciliationRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid close reconciliation payload",
				nil,
			)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		result, err := service.Execute(r.Context(), inventoryapp.CloseEndOfDayReconciliationInput{
			ID:            r.PathValue("reconciliation_id"),
			ActorID:       principal.UserID,
			RequestID:     response.RequestID(r),
			ExceptionNote: payload.ExceptionNote,
		})
		if err != nil {
			writeCloseReconciliationError(w, r, err)
			return
		}

		response.WriteSuccess(
			w,
			r,
			http.StatusOK,
			newEndOfDayReconciliationResponse(result.Reconciliation, result.AuditLogID),
		)
	}
}

func warehouseDailyBoardFulfillmentMetricsHandler(
	salesService salesapp.SalesOrderService,
	listManifests shippingapp.ListCarrierManifests,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionWarehouseView) {
			writePermissionDenied(w, r, auth.PermissionWarehouseView)
			return
		}

		warehouseID := strings.TrimSpace(r.URL.Query().Get("warehouse_id"))
		date := strings.TrimSpace(r.URL.Query().Get("date"))
		shiftCode := strings.TrimSpace(r.URL.Query().Get("shift_code"))
		carrierCode := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("carrier_code")))

		salesFilter := salesapp.SalesOrderFilter{WarehouseID: warehouseID}
		if date != "" {
			salesFilter.DateFrom = date
			salesFilter.DateTo = date
		}
		orders, err := salesService.ListSalesOrders(r.Context(), salesFilter)
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Daily board fulfillment metrics could not be loaded",
				nil,
			)
			return
		}

		manifests, err := listManifests.Execute(
			r.Context(),
			shippingdomain.NewCarrierManifestFilter(warehouseID, date, carrierCode, ""),
		)
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Daily board fulfillment metrics could not be loaded",
				nil,
			)
			return
		}

		payload := newWarehouseFulfillmentMetricsResponse(
			orders,
			manifests,
			warehouseID,
			date,
			shiftCode,
			carrierCode,
			time.Now().UTC(),
		)
		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

type warehouseDailyBoardPurchaseOrderLister interface {
	ListPurchaseOrders(context.Context, purchaseapp.PurchaseOrderFilter) ([]purchasedomain.PurchaseOrder, error)
}

type warehouseDailyBoardReceivingLister interface {
	ListWarehouseReceivings(context.Context, domain.WarehouseReceivingFilter) ([]domain.WarehouseReceiving, error)
}

type warehouseDailyBoardInboundQCLister interface {
	ListInboundQCInspections(context.Context, qcapp.InboundQCInspectionFilter) ([]qcdomain.InboundQCInspection, error)
}

type warehouseDailyBoardSupplierRejectionLister interface {
	Execute(context.Context, domain.SupplierRejectionFilter) ([]domain.SupplierRejection, error)
}

type warehouseDailyBoardSubcontractOrderLister interface {
	ListSubcontractOrders(context.Context, productionapp.SubcontractOrderFilter) ([]productiondomain.SubcontractOrder, error)
}

type warehouseDailyBoardSubcontractMaterialTransferLister interface {
	ListBySubcontractOrder(context.Context, string) ([]productiondomain.SubcontractMaterialTransfer, error)
}

type warehouseDailyBoardSubcontractFactoryClaimLister interface {
	ListBySubcontractOrder(context.Context, string) ([]productiondomain.SubcontractFactoryClaim, error)
}

type warehouseDailyBoardSubcontractPaymentMilestoneLister interface {
	ListBySubcontractOrder(context.Context, string) ([]productiondomain.SubcontractPaymentMilestone, error)
}

func warehouseDailyBoardInboundMetricsHandler(
	purchaseOrders warehouseDailyBoardPurchaseOrderLister,
	receivings warehouseDailyBoardReceivingLister,
	inboundQC warehouseDailyBoardInboundQCLister,
	supplierRejections warehouseDailyBoardSupplierRejectionLister,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionWarehouseView) {
			writePermissionDenied(w, r, auth.PermissionWarehouseView)
			return
		}

		warehouseID := strings.TrimSpace(r.URL.Query().Get("warehouse_id"))
		date := strings.TrimSpace(r.URL.Query().Get("date"))
		shiftCode := strings.TrimSpace(r.URL.Query().Get("shift_code"))

		purchaseFilter := purchaseapp.PurchaseOrderFilter{WarehouseID: warehouseID}
		if date != "" {
			purchaseFilter.ExpectedFrom = date
			purchaseFilter.ExpectedTo = date
		}
		orders, err := purchaseOrders.ListPurchaseOrders(r.Context(), purchaseFilter)
		if err != nil {
			writeWarehouseInboundMetricsLoadError(w, r)
			return
		}

		receipts, err := receivings.ListWarehouseReceivings(
			r.Context(),
			domain.NewWarehouseReceivingFilter(warehouseID, ""),
		)
		if err != nil {
			writeWarehouseInboundMetricsLoadError(w, r)
			return
		}

		inspections, err := inboundQC.ListInboundQCInspections(
			r.Context(),
			qcapp.NewInboundQCInspectionFilter("", "", "", warehouseID),
		)
		if err != nil {
			writeWarehouseInboundMetricsLoadError(w, r)
			return
		}

		rejections, err := supplierRejections.Execute(
			r.Context(),
			domain.NewSupplierRejectionFilter("", warehouseID, ""),
		)
		if err != nil {
			writeWarehouseInboundMetricsLoadError(w, r)
			return
		}

		payload := newWarehouseInboundMetricsResponse(
			orders,
			receipts,
			inspections,
			rejections,
			warehouseID,
			date,
			shiftCode,
			time.Now().UTC(),
		)
		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func warehouseDailyBoardSubcontractMetricsHandler(
	orders warehouseDailyBoardSubcontractOrderLister,
	materialTransfers warehouseDailyBoardSubcontractMaterialTransferLister,
	factoryClaims warehouseDailyBoardSubcontractFactoryClaimLister,
	paymentMilestones warehouseDailyBoardSubcontractPaymentMilestoneLister,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionWarehouseView) {
			writePermissionDenied(w, r, auth.PermissionWarehouseView)
			return
		}

		warehouseID := strings.TrimSpace(r.URL.Query().Get("warehouse_id"))
		date := strings.TrimSpace(r.URL.Query().Get("date"))
		shiftCode := strings.TrimSpace(r.URL.Query().Get("shift_code"))

		orderRows, err := orders.ListSubcontractOrders(r.Context(), productionapp.SubcontractOrderFilter{})
		if err != nil {
			writeWarehouseSubcontractMetricsLoadError(w, r)
			return
		}
		payload, err := newWarehouseSubcontractMetricsResponse(
			r.Context(),
			orderRows,
			materialTransfers,
			factoryClaims,
			paymentMilestones,
			warehouseID,
			date,
			shiftCode,
			time.Now().UTC(),
		)
		if err != nil {
			writeWarehouseSubcontractMetricsLoadError(w, r)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func writeWarehouseInboundMetricsLoadError(w http.ResponseWriter, r *http.Request) {
	response.WriteError(
		w,
		r,
		http.StatusConflict,
		response.ErrorCodeConflict,
		"Daily board inbound metrics could not be loaded",
		nil,
	)
}

func writeWarehouseSubcontractMetricsLoadError(w http.ResponseWriter, r *http.Request) {
	response.WriteError(
		w,
		r,
		http.StatusConflict,
		response.ErrorCodeConflict,
		"Daily board subcontract metrics could not be loaded",
		nil,
	)
}

func carrierManifestsHandler(
	listService shippingapp.ListCarrierManifests,
	createService shippingapp.CreateCarrierManifest,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionShippingView) {
				writePermissionDenied(w, r, auth.PermissionShippingView)
				return
			}
			filter := shippingdomain.NewCarrierManifestFilter(
				r.URL.Query().Get("warehouse_id"),
				r.URL.Query().Get("date"),
				r.URL.Query().Get("carrier_code"),
				shippingdomain.CarrierManifestStatus(r.URL.Query().Get("status")),
			)
			manifests, err := listService.Execute(r.Context(), filter)
			if err != nil {
				response.WriteError(
					w,
					r,
					http.StatusConflict,
					response.ErrorCodeConflict,
					"Carrier manifests could not be loaded",
					nil,
				)
				return
			}

			payload := make([]carrierManifestResponse, 0, len(manifests))
			for _, manifest := range manifests {
				payload = append(payload, newCarrierManifestResponse(manifest, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionShippingView) {
				writePermissionDenied(w, r, auth.PermissionShippingView)
				return
			}
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			if !hasAnyRole(principal, auth.RoleWarehouseLead, auth.RoleERPAdmin) {
				writeRoleDenied(w, r, auth.RoleWarehouseLead, auth.RoleERPAdmin)
				return
			}
			var payload createCarrierManifestRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid carrier manifest payload",
					nil,
				)
				return
			}

			result, err := createService.Execute(r.Context(), shippingapp.CreateCarrierManifestInput{
				ID:               payload.ID,
				CarrierCode:      payload.CarrierCode,
				CarrierName:      payload.CarrierName,
				WarehouseID:      payload.WarehouseID,
				WarehouseCode:    payload.WarehouseCode,
				Date:             payload.Date,
				HandoverBatch:    payload.HandoverBatch,
				StagingZone:      payload.StagingZone,
				HandoverZoneID:   payload.HandoverZoneID,
				HandoverZoneCode: payload.HandoverZoneCode,
				HandoverBinID:    payload.HandoverBinID,
				HandoverBinCode:  payload.HandoverBinCode,
				Owner:            payload.Owner,
				ActorID:          principal.UserID,
				RequestID:        response.RequestID(r),
			})
			if err != nil {
				writeCarrierManifestError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newCarrierManifestResponse(result.Manifest, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func addShipmentToCarrierManifestHandler(service shippingapp.AddShipmentToCarrierManifest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload addShipmentToManifestRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid add shipment payload",
				nil,
			)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionShippingView) {
			writePermissionDenied(w, r, auth.PermissionShippingView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}
		if !hasAnyRole(principal, auth.RoleWarehouseLead, auth.RoleERPAdmin) {
			writeRoleDenied(w, r, auth.RoleWarehouseLead, auth.RoleERPAdmin)
			return
		}
		result, err := service.Execute(r.Context(), shippingapp.AddShipmentToCarrierManifestInput{
			ManifestID: r.PathValue("manifest_id"),
			ShipmentID: payload.ShipmentID,
			ActorID:    principal.UserID,
			RequestID:  response.RequestID(r),
		})
		if err != nil {
			writeCarrierManifestError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newCarrierManifestResponse(result.Manifest, result.AuditLogID))
	}
}

func removeShipmentFromCarrierManifestHandler(service shippingapp.RemoveShipmentFromCarrierManifest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionShippingView) {
			writePermissionDenied(w, r, auth.PermissionShippingView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}
		if !hasAnyRole(principal, auth.RoleWarehouseLead, auth.RoleERPAdmin) {
			writeRoleDenied(w, r, auth.RoleWarehouseLead, auth.RoleERPAdmin)
			return
		}
		result, err := service.Execute(r.Context(), shippingapp.RemoveShipmentFromCarrierManifestInput{
			ManifestID: r.PathValue("manifest_id"),
			ShipmentID: r.PathValue("shipment_id"),
			ActorID:    principal.UserID,
			RequestID:  response.RequestID(r),
		})
		if err != nil {
			writeCarrierManifestError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newCarrierManifestResponse(result.Manifest, result.AuditLogID))
	}
}

func markCarrierManifestReadyToScanHandler(service shippingapp.MarkCarrierManifestReadyToScan) http.HandlerFunc {
	return carrierManifestActionHandler(func(r *http.Request, actorID string, _ cancelCarrierManifestRequest) (shippingapp.CarrierManifestResult, error) {
		return service.Execute(r.Context(), shippingapp.CarrierManifestActionInput{
			ManifestID: r.PathValue("manifest_id"),
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
		})
	})
}

func cancelCarrierManifestHandler(service shippingapp.CancelCarrierManifest) http.HandlerFunc {
	return carrierManifestActionHandler(func(r *http.Request, actorID string, payload cancelCarrierManifestRequest) (shippingapp.CarrierManifestResult, error) {
		return service.Execute(r.Context(), shippingapp.CarrierManifestActionInput{
			ManifestID: r.PathValue("manifest_id"),
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
			Reason:     payload.Reason,
		})
	})
}

func reportCarrierManifestMissingOrdersHandler(service shippingapp.ReportCarrierManifestMissingOrders) http.HandlerFunc {
	return carrierManifestActionHandler(func(r *http.Request, actorID string, payload cancelCarrierManifestRequest) (shippingapp.CarrierManifestResult, error) {
		return service.Execute(r.Context(), shippingapp.CarrierManifestActionInput{
			ManifestID: r.PathValue("manifest_id"),
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
			Reason:     payload.Reason,
		})
	})
}

func confirmCarrierManifestHandoverHandler(service shippingapp.ConfirmCarrierManifestHandover) http.HandlerFunc {
	return carrierManifestActionHandler(func(r *http.Request, actorID string, payload cancelCarrierManifestRequest) (shippingapp.CarrierManifestResult, error) {
		return service.Execute(r.Context(), shippingapp.CarrierManifestActionInput{
			ManifestID: r.PathValue("manifest_id"),
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
			Reason:     payload.Reason,
		})
	})
}

type carrierManifestAction func(*http.Request, string, cancelCarrierManifestRequest) (shippingapp.CarrierManifestResult, error)

func carrierManifestActionHandler(action carrierManifestAction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload cancelCarrierManifestRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid carrier manifest action payload",
					nil,
				)
				return
			}
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionShippingView) {
			writePermissionDenied(w, r, auth.PermissionShippingView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}
		if !hasAnyRole(principal, auth.RoleWarehouseLead, auth.RoleERPAdmin) {
			writeRoleDenied(w, r, auth.RoleWarehouseLead, auth.RoleERPAdmin)
			return
		}
		result, err := action(r, principal.UserID, payload)
		if err != nil {
			writeCarrierManifestError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newCarrierManifestResponse(result.Manifest, result.AuditLogID))
	}
}

func verifyCarrierManifestScanHandler(service shippingapp.VerifyCarrierManifestScan) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload verifyCarrierManifestScanRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid carrier manifest scan payload",
				nil,
			)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionShippingView) {
			writePermissionDenied(w, r, auth.PermissionShippingView)
			return
		}
		if !hasAnyRole(principal, auth.RoleWarehouseStaff, auth.RoleWarehouseLead, auth.RoleERPAdmin) {
			writeRoleDenied(w, r, auth.RoleWarehouseStaff, auth.RoleWarehouseLead, auth.RoleERPAdmin)
			return
		}
		result, err := service.Execute(r.Context(), shippingapp.VerifyCarrierManifestScanInput{
			ManifestID: r.PathValue("manifest_id"),
			Code:       payload.Code,
			StationID:  payload.StationID,
			DeviceID:   payload.DeviceID,
			Source:     payload.Source,
			ActorID:    principal.UserID,
			RequestID:  response.RequestID(r),
		})
		if err != nil {
			writeCarrierManifestError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newCarrierManifestScanResponse(result))
	}
}

func returnMasterDataHandler(service returnsapp.ListReturnMasterData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionReturnsView) {
			writePermissionDenied(w, r, auth.PermissionReturnsView)
			return
		}

		data, err := service.Execute(r.Context())
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Return master data could not be loaded",
				nil,
			)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newReturnMasterDataResponse(data))
	}
}

func returnScanHandler(receiveService returnsapp.ReceiveReturn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		var payload receiveReturnRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid return scan payload",
				nil,
			)
			return
		}
		if strings.TrimSpace(payload.Disposition) == "" {
			payload.Disposition = string(returnsdomain.ReturnDispositionNeedsInspection)
		}

		result, err := receiveService.Execute(r.Context(), returnsapp.ReceiveReturnInput{
			WarehouseID:       payload.WarehouseID,
			WarehouseCode:     payload.WarehouseCode,
			Source:            payload.Source,
			ScanCode:          payload.Code,
			PackageCondition:  payload.PackageCondition,
			Disposition:       payload.Disposition,
			InvestigationNote: payload.InvestigationNote,
			ActorID:           principal.UserID,
			RequestID:         response.RequestID(r),
		})
		if err != nil {
			writeReturnReceiptError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusCreated, newReturnReceiptResponse(result.Receipt, result.AuditLogID))
	}
}

func returnReceiptsHandler(
	listService returnsapp.ListReturnReceipts,
	receiveService returnsapp.ReceiveReturn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionReturnsView) {
				writePermissionDenied(w, r, auth.PermissionReturnsView)
				return
			}
			filter := returnsdomain.NewReturnReceiptFilter(
				r.URL.Query().Get("warehouse_id"),
				returnsdomain.ReturnReceiptStatus(r.URL.Query().Get("status")),
			)
			receipts, err := listService.Execute(r.Context(), filter)
			if err != nil {
				response.WriteError(
					w,
					r,
					http.StatusConflict,
					response.ErrorCodeConflict,
					"Return receipts could not be loaded",
					nil,
				)
				return
			}

			payload := make([]returnReceiptResponse, 0, len(receipts))
			for _, receipt := range receipts {
				payload = append(payload, newReturnReceiptResponse(receipt, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			var payload receiveReturnRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid return receiving payload",
					nil,
				)
				return
			}

			result, err := receiveService.Execute(r.Context(), returnsapp.ReceiveReturnInput{
				WarehouseID:       payload.WarehouseID,
				WarehouseCode:     payload.WarehouseCode,
				Source:            payload.Source,
				ScanCode:          payload.Code,
				PackageCondition:  payload.PackageCondition,
				Disposition:       payload.Disposition,
				InvestigationNote: payload.InvestigationNote,
				ActorID:           principal.UserID,
				RequestID:         response.RequestID(r),
			})
			if err != nil {
				writeReturnReceiptError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newReturnReceiptResponse(result.Receipt, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func returnInspectionHandler(inspectService returnsapp.InspectReturn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		var payload inspectReturnRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid return inspection payload",
				nil,
			)
			return
		}

		result, err := inspectService.Execute(r.Context(), returnsapp.InspectReturnInput{
			ReceiptID:     r.PathValue("return_receipt_id"),
			Condition:     payload.Condition,
			Disposition:   payload.Disposition,
			Note:          payload.Note,
			EvidenceLabel: payload.EvidenceLabel,
			ActorID:       principal.UserID,
			RequestID:     response.RequestID(r),
		})
		if err != nil {
			writeReturnReceiptError(w, r, err)
			return
		}

		response.WriteSuccess(
			w,
			r,
			http.StatusOK,
			newReturnInspectionResponse(result.Inspection, result.AuditLogID),
		)
	}
}

func returnDispositionHandler(applyService returnsapp.ApplyReturnDisposition) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		var payload applyReturnDispositionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid return disposition payload",
				nil,
			)
			return
		}

		result, err := applyService.Execute(r.Context(), returnsapp.ApplyReturnDispositionInput{
			ReceiptID:   r.PathValue("return_receipt_id"),
			Disposition: payload.Disposition,
			Note:        payload.Note,
			ActorID:     principal.UserID,
			RequestID:   response.RequestID(r),
		})
		if err != nil {
			writeReturnReceiptError(w, r, err)
			return
		}

		response.WriteSuccess(
			w,
			r,
			http.StatusOK,
			newReturnDispositionActionResponse(result.Action, result.AuditLogID),
		)
	}
}

func returnAttachmentHandler(uploadService returnsapp.UploadReturnAttachment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, returnsdomain.ReturnAttachmentMaxFileSizeBytes+(1<<20))
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid return attachment payload",
				map[string]any{"required": "file, inspection_id"},
			)
			return
		}
		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid return attachment payload",
				map[string]any{"required": "file, inspection_id"},
			)
			return
		}
		defer file.Close()

		mimeType := fileHeader.Header.Get("Content-Type")
		if mimeType == "" || mimeType == "application/octet-stream" {
			mimeType = mime.TypeByExtension(strings.ToLower(filepath.Ext(fileHeader.Filename)))
		}

		result, err := uploadService.Execute(r.Context(), returnsapp.UploadReturnAttachmentInput{
			ReceiptID:     r.PathValue("return_receipt_id"),
			InspectionID:  r.FormValue("inspection_id"),
			FileName:      fileHeader.Filename,
			MIMEType:      mimeType,
			FileSizeBytes: fileHeader.Size,
			Content:       file,
			Note:          r.FormValue("note"),
			ActorID:       principal.UserID,
			RequestID:     response.RequestID(r),
		})
		if err != nil {
			writeReturnReceiptError(w, r, err)
			return
		}

		response.WriteSuccess(
			w,
			r,
			http.StatusCreated,
			newReturnAttachmentResponse(result.Attachment, result.AuditLogID),
		)
	}
}

func newAvailableStockResponse(snapshot domain.AvailableStockSnapshot) availableStockResponse {
	return availableStockResponse{
		WarehouseID:      snapshot.WarehouseID,
		WarehouseCode:    snapshot.WarehouseCode,
		LocationID:       snapshot.LocationID,
		LocationCode:     snapshot.LocationCode,
		SKU:              snapshot.SKU,
		BatchID:          snapshot.BatchID,
		BatchNo:          snapshot.BatchNo,
		BatchQCStatus:    string(snapshot.BatchQCStatus),
		BatchStatus:      string(snapshot.BatchStatus),
		BatchExpiryDate:  dateString(snapshot.BatchExpiry),
		BaseUOMCode:      snapshot.BaseUOMCode.String(),
		PhysicalQty:      snapshot.PhysicalQty.String(),
		ReservedQty:      snapshot.ReservedQty.String(),
		QCHoldQty:        snapshot.QCHoldQty.String(),
		DamagedQty:       snapshot.DamagedQty.String(),
		ReturnPendingQty: snapshot.ReturnPendingQty.String(),
		BlockedQty:       snapshot.BlockedQty.String(),
		HoldQty:          snapshot.HoldQty.String(),
		AvailableQty:     snapshot.AvailableQty.String(),
	}
}

func newCreateStockAdjustmentLines(
	inputs []stockAdjustmentLineRequest,
) []inventoryapp.CreateStockAdjustmentLineInput {
	lines := make([]inventoryapp.CreateStockAdjustmentLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, inventoryapp.CreateStockAdjustmentLineInput{
			ID:           input.ID,
			ItemID:       input.ItemID,
			SKU:          input.SKU,
			BatchID:      input.BatchID,
			BatchNo:      input.BatchNo,
			LocationID:   input.LocationID,
			LocationCode: input.LocationCode,
			ExpectedQty:  input.ExpectedQty,
			CountedQty:   input.CountedQty,
			BaseUOMCode:  input.BaseUOMCode,
			Reason:       input.Reason,
		})
	}

	return lines
}

func newCreateStockCountLines(inputs []stockCountLineRequest) []inventoryapp.CreateStockCountLineInput {
	lines := make([]inventoryapp.CreateStockCountLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, inventoryapp.CreateStockCountLineInput{
			ID:           input.ID,
			ItemID:       input.ItemID,
			SKU:          input.SKU,
			BatchID:      input.BatchID,
			BatchNo:      input.BatchNo,
			LocationID:   input.LocationID,
			LocationCode: input.LocationCode,
			ExpectedQty:  input.ExpectedQty,
			BaseUOMCode:  input.BaseUOMCode,
		})
	}

	return lines
}

func newSubmitStockCountLines(inputs []submitStockCountLineRequest) []inventoryapp.SubmitStockCountLineInput {
	lines := make([]inventoryapp.SubmitStockCountLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, inventoryapp.SubmitStockCountLineInput{
			ID:         input.ID,
			SKU:        input.SKU,
			CountedQty: input.CountedQty,
			Note:       input.Note,
		})
	}

	return lines
}

func newStockCountResponse(session domain.StockCountSession, auditLogID string) stockCountResponse {
	payload := stockCountResponse{
		ID:            session.ID,
		CountNo:       session.CountNo,
		OrgID:         session.OrgID,
		WarehouseID:   session.WarehouseID,
		WarehouseCode: session.WarehouseCode,
		Scope:         session.Scope,
		Status:        string(session.Status),
		CreatedBy:     session.CreatedBy,
		SubmittedBy:   session.SubmittedBy,
		Lines:         make([]stockCountLineResponse, 0, len(session.Lines)),
		AuditLogID:    auditLogID,
		CreatedAt:     session.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     session.UpdatedAt.UTC().Format(time.RFC3339),
		SubmittedAt:   timeString(session.SubmittedAt),
	}
	for _, line := range session.Lines {
		payload.Lines = append(payload.Lines, stockCountLineResponse{
			ID:           line.ID,
			ItemID:       line.ItemID,
			SKU:          line.SKU,
			BatchID:      line.BatchID,
			BatchNo:      line.BatchNo,
			LocationID:   line.LocationID,
			LocationCode: line.LocationCode,
			ExpectedQty:  line.ExpectedQty.String(),
			CountedQty:   line.CountedQty.String(),
			DeltaQty:     line.DeltaQty.String(),
			BaseUOMCode:  line.BaseUOMCode.String(),
			Counted:      line.Counted,
			Note:         line.Note,
		})
	}

	return payload
}

func newStockAdjustmentResponse(
	adjustment domain.StockAdjustment,
	auditLogID string,
) stockAdjustmentResponse {
	payload := stockAdjustmentResponse{
		ID:            adjustment.ID,
		AdjustmentNo:  adjustment.AdjustmentNo,
		OrgID:         adjustment.OrgID,
		WarehouseID:   adjustment.WarehouseID,
		WarehouseCode: adjustment.WarehouseCode,
		SourceType:    adjustment.SourceType,
		SourceID:      adjustment.SourceID,
		Reason:        adjustment.Reason,
		Status:        string(adjustment.Status),
		RequestedBy:   adjustment.RequestedBy,
		SubmittedBy:   adjustment.SubmittedBy,
		ApprovedBy:    adjustment.ApprovedBy,
		RejectedBy:    adjustment.RejectedBy,
		PostedBy:      adjustment.PostedBy,
		Lines:         make([]stockAdjustmentLineResponse, 0, len(adjustment.Lines)),
		AuditLogID:    auditLogID,
		CreatedAt:     adjustment.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     adjustment.UpdatedAt.UTC().Format(time.RFC3339),
		SubmittedAt:   timeString(adjustment.SubmittedAt),
		ApprovedAt:    timeString(adjustment.ApprovedAt),
		RejectedAt:    timeString(adjustment.RejectedAt),
		PostedAt:      timeString(adjustment.PostedAt),
	}
	for _, line := range adjustment.Lines {
		payload.Lines = append(payload.Lines, stockAdjustmentLineResponse{
			ID:           line.ID,
			ItemID:       line.ItemID,
			SKU:          line.SKU,
			BatchID:      line.BatchID,
			BatchNo:      line.BatchNo,
			LocationID:   line.LocationID,
			LocationCode: line.LocationCode,
			ExpectedQty:  line.ExpectedQty.String(),
			CountedQty:   line.CountedQty.String(),
			DeltaQty:     line.DeltaQty.String(),
			BaseUOMCode:  line.BaseUOMCode.String(),
			Reason:       line.Reason,
		})
	}

	return payload
}

func newBatchResponse(batch domain.Batch) batchResponse {
	return batchResponse{
		ID:         batch.ID,
		OrgID:      batch.OrgID,
		ItemID:     batch.ItemID,
		SKU:        batch.SKU,
		ItemName:   batch.ItemName,
		BatchNo:    batch.BatchNo,
		SupplierID: batch.SupplierID,
		MfgDate:    dateString(batch.MfgDate),
		ExpiryDate: dateString(batch.ExpiryDate),
		QCStatus:   string(batch.QCStatus),
		Status:     string(batch.Status),
		CreatedAt:  batch.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  batch.UpdatedAt.Format(time.RFC3339),
	}
}

func newBatchQCTransitionResponse(transition domain.BatchQCTransition) batchQCTransitionResponse {
	return batchQCTransitionResponse{
		ID:           transition.ID,
		BatchID:      transition.BatchID,
		BatchNo:      transition.BatchNo,
		SKU:          transition.SKU,
		FromQCStatus: string(transition.FromQCStatus),
		ToQCStatus:   string(transition.ToQCStatus),
		ActorID:      transition.ActorID,
		Reason:       transition.Reason,
		BusinessRef:  transition.BusinessRef,
		AuditLogID:   transition.AuditLogID,
		CreatedAt:    transition.CreatedAt.Format(time.RFC3339),
	}
}

func newCreateWarehouseReceivingLines(
	inputs []createWarehouseReceivingLineRequest,
) []inventoryapp.CreateWarehouseReceivingLineInput {
	lines := make([]inventoryapp.CreateWarehouseReceivingLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, inventoryapp.CreateWarehouseReceivingLineInput{
			ID:                  input.ID,
			PurchaseOrderLineID: input.PurchaseOrderLineID,
			ItemID:              input.ItemID,
			SKU:                 input.SKU,
			ItemName:            input.ItemName,
			BatchID:             input.BatchID,
			BatchNo:             input.BatchNo,
			LotNo:               input.LotNo,
			ExpiryDate:          input.ExpiryDate,
			Quantity:            input.Quantity,
			UOMCode:             input.UOMCode,
			BaseUOMCode:         input.BaseUOMCode,
			PackagingStatus:     input.PackagingStatus,
			QCStatus:            input.QCStatus,
		})
	}

	return lines
}

func newWarehouseReceivingResponse(
	receipt domain.WarehouseReceiving,
	auditLogID string,
) warehouseReceivingResponse {
	payload := warehouseReceivingResponse{
		ID:               receipt.ID,
		OrgID:            receipt.OrgID,
		ReceiptNo:        receipt.ReceiptNo,
		WarehouseID:      receipt.WarehouseID,
		WarehouseCode:    receipt.WarehouseCode,
		LocationID:       receipt.LocationID,
		LocationCode:     receipt.LocationCode,
		ReferenceDocType: receipt.ReferenceDocType,
		ReferenceDocID:   receipt.ReferenceDocID,
		SupplierID:       receipt.SupplierID,
		DeliveryNoteNo:   receipt.DeliveryNoteNo,
		Status:           string(receipt.Status),
		Lines:            make([]warehouseReceivingLineResponse, 0, len(receipt.Lines)),
		StockMovements:   make([]warehouseReceivingStockMovementResponse, 0, len(receipt.StockMovements)),
		CreatedBy:        receipt.CreatedBy,
		SubmittedBy:      receipt.SubmittedBy,
		InspectReadyBy:   receipt.InspectReadyBy,
		PostedBy:         receipt.PostedBy,
		AuditLogID:       auditLogID,
		CreatedAt:        receipt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        receipt.UpdatedAt.Format(time.RFC3339),
		SubmittedAt:      timeString(receipt.SubmittedAt),
		InspectReadyAt:   timeString(receipt.InspectReadyAt),
		PostedAt:         timeString(receipt.PostedAt),
	}
	for _, line := range receipt.Lines {
		payload.Lines = append(payload.Lines, warehouseReceivingLineResponse{
			ID:                  line.ID,
			PurchaseOrderLineID: line.PurchaseOrderLineID,
			ItemID:              line.ItemID,
			SKU:                 line.SKU,
			ItemName:            line.ItemName,
			BatchID:             line.BatchID,
			BatchNo:             line.BatchNo,
			LotNo:               line.LotNo,
			ExpiryDate:          dateString(line.ExpiryDate),
			WarehouseID:         line.WarehouseID,
			LocationID:          line.LocationID,
			Quantity:            line.Quantity.String(),
			UOMCode:             line.UOMCode.String(),
			BaseUOMCode:         line.BaseUOMCode.String(),
			PackagingStatus:     string(line.PackagingStatus),
			QCStatus:            string(line.QCStatus),
		})
	}
	for _, movement := range receipt.StockMovements {
		payload.StockMovements = append(payload.StockMovements, warehouseReceivingStockMovementResponse{
			MovementNo:      movement.MovementNo,
			MovementType:    string(movement.MovementType),
			ItemID:          movement.ItemID,
			BatchID:         movement.BatchID,
			WarehouseID:     movement.WarehouseID,
			LocationID:      movement.BinID,
			Quantity:        movement.Quantity.String(),
			BaseUOMCode:     movement.BaseUOMCode.String(),
			StockStatus:     string(movement.StockStatus),
			SourceDocID:     movement.SourceDocID,
			SourceDocLineID: movement.SourceDocLineID,
		})
	}

	return payload
}

func timeString(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339)
}

func dateString(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format(time.DateOnly)
}

func newEndOfDayReconciliationResponse(
	reconciliation domain.EndOfDayReconciliation,
	auditLogID string,
) endOfDayReconciliationResponse {
	summary := reconciliation.Summary("")
	payload := endOfDayReconciliationResponse{
		ID:            reconciliation.ID,
		WarehouseID:   reconciliation.WarehouseID,
		WarehouseCode: reconciliation.WarehouseCode,
		Date:          reconciliation.Date,
		ShiftCode:     reconciliation.ShiftCode,
		Status:        string(reconciliation.Status),
		Owner:         reconciliation.Owner,
		AuditLogID:    auditLogID,
		Summary: endOfDayReconciliationSummaryResponse{
			SystemQuantity:     summary.SystemQuantity,
			CountedQuantity:    summary.CountedQuantity,
			VarianceQuantity:   summary.VarianceQuantity,
			VarianceCount:      summary.VarianceCount,
			ChecklistTotal:     summary.ChecklistTotal,
			ChecklistCompleted: summary.ChecklistCompleted,
			ReadyToClose:       summary.ReadyToClose,
		},
		Operations: endOfDayReconciliationOperationsResponse{
			OrderCount:             reconciliation.Operations.OrderCount,
			HandoverOrderCount:     reconciliation.Operations.HandoverOrderCount,
			ReturnOrderCount:       reconciliation.Operations.ReturnOrderCount,
			StockMovementCount:     reconciliation.Operations.StockMovementCount,
			StockCountSessionCount: reconciliation.Operations.StockCountSessionCount,
			PendingIssueCount:      reconciliation.Operations.PendingIssueCount,
		},
		Checklist: make([]endOfDayReconciliationChecklistResponse, 0, len(reconciliation.Checklist)),
		Lines:     make([]endOfDayReconciliationLineResponse, 0, len(reconciliation.Lines)),
	}
	if !reconciliation.ClosedAt.IsZero() {
		payload.ClosedAt = reconciliation.ClosedAt.UTC().Format(time.RFC3339)
	}
	if strings.TrimSpace(reconciliation.ClosedBy) != "" {
		payload.ClosedBy = strings.TrimSpace(reconciliation.ClosedBy)
	}
	for _, item := range reconciliation.Checklist {
		payload.Checklist = append(payload.Checklist, endOfDayReconciliationChecklistResponse{
			Key:      item.Key,
			Label:    item.Label,
			Complete: item.Complete,
			Blocking: item.Blocking,
			Note:     item.Note,
		})
	}
	for _, line := range reconciliation.Lines {
		payload.Lines = append(payload.Lines, endOfDayReconciliationLineResponse{
			ID:               line.ID,
			SKU:              line.SKU,
			BatchNo:          line.BatchNo,
			BinCode:          line.BinCode,
			SystemQuantity:   line.SystemQuantity,
			CountedQuantity:  line.CountedQuantity,
			VarianceQuantity: line.VarianceQuantity(),
			Reason:           line.Reason,
			Owner:            line.Owner,
		})
	}

	return payload
}

func newCarrierManifestResponse(manifest shippingdomain.CarrierManifest, auditLogID string) carrierManifestResponse {
	summary := manifest.Summary()
	payload := carrierManifestResponse{
		ID:               manifest.ID,
		CarrierCode:      manifest.CarrierCode,
		CarrierName:      manifest.CarrierName,
		WarehouseID:      manifest.WarehouseID,
		WarehouseCode:    manifest.WarehouseCode,
		Date:             manifest.Date,
		HandoverBatch:    manifest.HandoverBatch,
		StagingZone:      manifest.StagingZone,
		HandoverZoneCode: manifest.HandoverZoneCode,
		HandoverBinCode:  manifest.HandoverBinCode,
		Status:           string(manifest.Status),
		Owner:            manifest.Owner,
		AuditLogID:       auditLogID,
		Summary: carrierManifestSummaryResponse{
			ExpectedCount: summary.ExpectedCount,
			ScannedCount:  summary.ScannedCount,
			MissingCount:  summary.MissingCount,
		},
		Lines:        make([]carrierManifestLineResponse, 0, len(manifest.Lines)),
		MissingLines: make([]carrierManifestLineResponse, 0, summary.MissingCount),
	}
	if !manifest.CreatedAt.IsZero() {
		payload.CreatedAt = manifest.CreatedAt.UTC().Format(time.RFC3339)
	}
	for _, line := range manifest.Lines {
		lineResponse := newCarrierManifestLineResponse(line)
		payload.Lines = append(payload.Lines, lineResponse)
		if !line.Scanned {
			payload.MissingLines = append(payload.MissingLines, lineResponse)
		}
	}

	return payload
}

func newCarrierManifestLineResponse(line shippingdomain.CarrierManifestLine) carrierManifestLineResponse {
	return carrierManifestLineResponse{
		ID:               line.ID,
		ShipmentID:       line.ShipmentID,
		OrderNo:          line.OrderNo,
		TrackingNo:       line.TrackingNo,
		PackageCode:      line.PackageCode,
		StagingZone:      line.StagingZone,
		HandoverZoneCode: line.HandoverZoneCode,
		HandoverBinCode:  line.HandoverBinCode,
		Scanned:          line.Scanned,
	}
}

func newCarrierManifestScanResponse(result shippingapp.CarrierManifestScanResult) carrierManifestScanResponse {
	payload := carrierManifestScanResponse{
		ResultCode:         string(result.Code),
		Severity:           result.Severity,
		Message:            result.Message,
		ExpectedManifestID: result.ExpectedManifestID,
		ScanEvent: carrierManifestScanEventResponse{
			ID:                 result.Event.ID,
			ManifestID:         result.Event.ManifestID,
			ExpectedManifestID: result.Event.ExpectedManifestID,
			Code:               result.Event.Code,
			ResultCode:         string(result.Event.ResultCode),
			Severity:           result.Event.Severity,
			Message:            result.Event.Message,
			ShipmentID:         result.Event.ShipmentID,
			OrderNo:            result.Event.OrderNo,
			TrackingNo:         result.Event.TrackingNo,
			ActorID:            result.Event.ActorID,
			StationID:          result.Event.StationID,
			DeviceID:           result.Event.DeviceID,
			Source:             result.Event.Source,
			WarehouseID:        result.Event.WarehouseID,
			CarrierCode:        result.Event.CarrierCode,
			CreatedAt:          result.Event.CreatedAt.UTC().Format(time.RFC3339),
		},
		Manifest:   newCarrierManifestResponse(result.Manifest, ""),
		AuditLogID: result.AuditLogID,
	}
	if result.Line != nil {
		payload.Line = &carrierManifestLineResponse{
			ID:               result.Line.ID,
			ShipmentID:       result.Line.ShipmentID,
			OrderNo:          result.Line.OrderNo,
			TrackingNo:       result.Line.TrackingNo,
			PackageCode:      result.Line.PackageCode,
			StagingZone:      result.Line.StagingZone,
			HandoverZoneCode: result.Line.HandoverZoneCode,
			HandoverBinCode:  result.Line.HandoverBinCode,
			Scanned:          result.Line.Scanned,
		}
	}

	return payload
}

func newReturnMasterDataResponse(data returnsdomain.ReturnMasterData) returnMasterDataResponse {
	payload := returnMasterDataResponse{
		Reasons:      make([]returnReasonResponse, 0, len(data.Reasons)),
		Conditions:   make([]returnConditionResponse, 0, len(data.Conditions)),
		Dispositions: make([]returnDispositionResponse, 0, len(data.Dispositions)),
	}
	for _, reason := range data.Reasons {
		payload.Reasons = append(payload.Reasons, returnReasonResponse{
			Code:        reason.Code,
			Label:       reason.Label,
			Description: reason.Description,
			Active:      reason.Active,
			SortOrder:   reason.SortOrder,
		})
	}
	for _, condition := range data.Conditions {
		payload.Conditions = append(payload.Conditions, returnConditionResponse{
			Code:                 condition.Code,
			Label:                condition.Label,
			Description:          condition.Description,
			DefaultDisposition:   string(condition.DefaultDisposition),
			InventoryDisposition: condition.InventoryDisposition,
			RequiresQA:           condition.RequiresQA,
			Active:               condition.Active,
			SortOrder:            condition.SortOrder,
		})
	}
	for _, disposition := range data.Dispositions {
		payload.Dispositions = append(payload.Dispositions, returnDispositionResponse{
			Code:                  string(disposition.Code),
			Label:                 disposition.Label,
			Description:           disposition.Description,
			InventoryDisposition:  disposition.InventoryDisposition,
			TargetStockStatus:     disposition.TargetStockStatus,
			TargetLocationType:    disposition.TargetLocationType,
			CreatesAvailableStock: disposition.CreatesAvailableStock,
			RequiresApproval:      disposition.RequiresApproval,
			Active:                disposition.Active,
			SortOrder:             disposition.SortOrder,
		})
	}

	return payload
}

func newReturnReceiptResponse(receipt returnsdomain.ReturnReceipt, auditLogID string) returnReceiptResponse {
	payload := returnReceiptResponse{
		ID:                receipt.ID,
		ReceiptNo:         receipt.ReceiptNo,
		WarehouseID:       receipt.WarehouseID,
		WarehouseCode:     receipt.WarehouseCode,
		Source:            string(receipt.Source),
		ReceivedBy:        receipt.ReceivedBy,
		ReceivedAt:        receipt.ReceivedAt.UTC().Format(time.RFC3339),
		PackageCondition:  receipt.PackageCondition,
		Status:            string(receipt.Status),
		Disposition:       string(receipt.Disposition),
		TargetLocation:    receipt.TargetLocation,
		OriginalOrderNo:   receipt.OriginalOrderNo,
		TrackingNo:        receipt.TrackingNo,
		ReturnCode:        receipt.ReturnCode,
		ScanCode:          receipt.ScanCode,
		CustomerName:      receipt.CustomerName,
		UnknownCase:       receipt.UnknownCase,
		Lines:             make([]returnReceiptLineResponse, 0, len(receipt.Lines)),
		InvestigationNote: receipt.InvestigationNote,
		AuditLogID:        auditLogID,
		CreatedAt:         receipt.CreatedAt.UTC().Format(time.RFC3339),
	}
	for _, line := range receipt.Lines {
		payload.Lines = append(payload.Lines, returnReceiptLineResponse{
			ID:          line.ID,
			SKU:         line.SKU,
			ProductName: line.ProductName,
			Quantity:    line.Quantity,
			Condition:   line.Condition,
		})
	}
	if receipt.StockMovement != nil {
		payload.StockMovement = &returnStockMovementResponse{
			ID:                receipt.StockMovement.ID,
			MovementType:      receipt.StockMovement.MovementType,
			SKU:               receipt.StockMovement.SKU,
			WarehouseID:       receipt.StockMovement.WarehouseID,
			Quantity:          receipt.StockMovement.Quantity,
			TargetStockStatus: receipt.StockMovement.TargetStockStatus,
			SourceDocID:       receipt.StockMovement.SourceDocID,
		}
	}

	return payload
}

func newReturnInspectionResponse(
	inspection returnsdomain.ReturnInspection,
	auditLogID string,
) returnInspectionResponse {
	return returnInspectionResponse{
		ID:             inspection.ID,
		ReceiptID:      inspection.ReceiptID,
		ReceiptNo:      inspection.ReceiptNo,
		Condition:      string(inspection.Condition),
		Disposition:    string(inspection.Disposition),
		Status:         string(inspection.Status),
		TargetLocation: inspection.TargetLocation,
		RiskLevel:      inspection.RiskLevel,
		InspectorID:    inspection.InspectorID,
		Note:           inspection.Note,
		EvidenceLabel:  inspection.EvidenceLabel,
		AuditLogID:     auditLogID,
		InspectedAt:    inspection.InspectedAt.UTC().Format(time.RFC3339),
	}
}

func newReturnDispositionActionResponse(
	action returnsdomain.ReturnDispositionAction,
	auditLogID string,
) returnDispositionActionResponse {
	return returnDispositionActionResponse{
		ID:                action.ID,
		ReceiptID:         action.ReceiptID,
		ReceiptNo:         action.ReceiptNo,
		Disposition:       string(action.Disposition),
		TargetLocation:    action.TargetLocation,
		TargetStockStatus: action.TargetStockStatus,
		ActionCode:        action.ActionCode,
		ActorID:           action.ActorID,
		Note:              action.Note,
		AuditLogID:        auditLogID,
		DecidedAt:         action.DecidedAt.UTC().Format(time.RFC3339),
	}
}

func newReturnAttachmentResponse(
	attachment returnsdomain.ReturnAttachment,
	auditLogID string,
) returnAttachmentResponse {
	return returnAttachmentResponse{
		ID:            attachment.ID,
		ReceiptID:     attachment.ReceiptID,
		ReceiptNo:     attachment.ReceiptNo,
		InspectionID:  attachment.InspectionID,
		FileName:      attachment.FileName,
		FileExt:       attachment.FileExt,
		MIMEType:      attachment.MIMEType,
		FileSizeBytes: attachment.FileSizeBytes,
		StorageBucket: attachment.StorageBucket,
		StorageKey:    attachment.StorageKey,
		Status:        attachment.Status,
		UploadedBy:    attachment.UploadedBy,
		Note:          attachment.Note,
		AuditLogID:    auditLogID,
		UploadedAt:    attachment.UploadedAt.UTC().Format(time.RFC3339),
	}
}

func writeBatchQCTransitionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrBatchNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Batch not found", nil)
	case errors.Is(err, inventoryapp.ErrBatchTransitionActorRequired):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Batch QC transition actor is required",
			map[string]any{"required": "actor"},
		)
	case errors.Is(err, inventoryapp.ErrBatchTransitionReasonRequired):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Batch QC transition reason is required",
			map[string]any{"required": "reason"},
		)
	case errors.Is(err, domain.ErrBatchInvalidQCStatus):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Batch QC status is invalid",
			map[string]any{"allowed": "hold, pass, fail, quarantine, retest_required"},
		)
	case errors.Is(err, domain.ErrBatchInvalidQCTransition):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Batch QC status transition is not allowed",
			nil,
		)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Batch QC transition could not be processed", nil)
	}
}

func writeStockAdjustmentError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrStockAdjustmentNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Stock adjustment not found", nil)
	case errors.Is(err, domain.ErrStockAdjustmentRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Stock adjustment required field is missing",
			map[string]any{"required": "warehouse_id, reason, lines, sku, expected_qty, counted_qty, base_uom_code"},
		)
	case errors.Is(err, domain.ErrStockAdjustmentInvalidQuantity),
		errors.Is(err, decimal.ErrInvalidDecimal),
		errors.Is(err, decimal.ErrInvalidUOMCode):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Stock adjustment quantity or UOM is invalid",
			nil,
		)
	case errors.Is(err, domain.ErrStockAdjustmentNoVariance):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Stock adjustment requires at least one variance",
			nil,
		)
	case errors.Is(err, domain.ErrStockAdjustmentInvalidStatus):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Stock adjustment status transition is not allowed", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Stock adjustment could not be processed", nil)
	}
}

func writeStockCountError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrStockCountNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Stock count not found", nil)
	case errors.Is(err, domain.ErrStockCountRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Stock count required field is missing",
			map[string]any{"required": "warehouse_id, lines, sku, expected_qty, base_uom_code, counted_qty"},
		)
	case errors.Is(err, domain.ErrStockCountInvalidQuantity),
		errors.Is(err, decimal.ErrInvalidDecimal),
		errors.Is(err, decimal.ErrInvalidUOMCode):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Stock count quantity or UOM is invalid", nil)
	case errors.Is(err, domain.ErrStockCountInvalidStatus):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Stock count status transition is not allowed", nil)
	case errors.Is(err, domain.ErrStockCountLineNotFound):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Stock count line not found", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Stock count could not be processed", nil)
	}
}

func writeWarehouseReceivingError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrWarehouseReceivingNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Goods receipt not found", nil)
	case errors.Is(err, inventoryapp.ErrReceivingInvalidLocation):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Goods receipt location is invalid",
			map[string]any{"field": "location_id"},
		)
	case errors.Is(err, inventoryapp.ErrBatchNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Batch not found", nil)
	case errors.Is(err, inventoryapp.ErrReceivingBatchMismatch):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Goods receipt batch does not match the receiving line",
			nil,
		)
	case errors.Is(err, inventoryapp.ErrReceivingPurchaseOrderInvalidState):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Purchase order is not open for goods receiving",
			nil,
		)
	case errors.Is(err, inventoryapp.ErrReceivingPurchaseOrderMismatch),
		errors.Is(err, inventoryapp.ErrReceivingQuantityExceedsPurchaseOrder):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Goods receipt does not match the linked purchase order",
			nil,
		)
	case errors.Is(err, domain.ErrReceivingRequiredField),
		errors.Is(err, domain.ErrReceivingInvalidStatus),
		errors.Is(err, domain.ErrReceivingInvalidPackagingStatus),
		errors.Is(err, domain.ErrBatchInvalidQCStatus),
		errors.Is(err, decimal.ErrInvalidDecimal),
		errors.Is(err, decimal.ErrInvalidUOMCode),
		errors.Is(err, decimal.ErrDecimalOutOfRange):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid goods receipt payload",
			map[string]any{"required": "warehouse_id, location_id, reference_doc_type, reference_doc_id, supplier_id, delivery_note_no, lines, purchase_order_line_id, quantity, uom_code, base_uom_code, packaging_status, and expiry_date when batch/lot is present"},
		)
	case errors.Is(err, domain.ErrReceivingMissingBatchQCData):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Batch and QC data are required before posting goods receipt",
			map[string]any{"required": "batch_id and qc_status"},
		)
	case errors.Is(err, domain.ErrReceivingAlreadyPosted):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Goods receipt is already posted", nil)
	case errors.Is(err, domain.ErrReceivingInvalidTransition):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Goods receipt status transition is not allowed", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Goods receipt request could not be processed", nil)
	}
}

func writeCloseReconciliationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrEndOfDayReconciliationNotFound):
		response.WriteError(
			w,
			r,
			http.StatusNotFound,
			response.ErrorCodeNotFound,
			"End-of-day reconciliation not found",
			nil,
		)
	case errors.Is(err, domain.ErrReconciliationAlreadyClosed):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"End-of-day reconciliation is already closed",
			nil,
		)
	case errors.Is(err, domain.ErrReconciliationNeedsExceptionNote):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Exception note is required before closing this shift",
			map[string]any{"exception_note": "required"},
		)
	case errors.Is(err, domain.ErrReconciliationUnresolvedIssue):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Resolve return, manifest, adjustment, or pending issue before closing this shift",
			map[string]any{"unresolved_issue": "required"},
		)
	default:
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"End-of-day reconciliation could not be closed",
			nil,
		)
	}
}

func writeSalesOrderError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Sales order request could not be processed", nil)
}

func writeCarrierManifestError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	switch {
	case errors.Is(err, shippingapp.ErrCarrierManifestNotFound),
		errors.Is(err, shippingapp.ErrPackedShipmentNotFound),
		errors.Is(err, shippingdomain.ErrManifestShipmentNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Carrier manifest resource not found", nil)
	case errors.Is(err, shippingapp.ErrCarrierNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Carrier was not found", nil)
	case errors.Is(err, shippingdomain.ErrManifestRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid carrier manifest payload",
			map[string]any{"required": "carrier_code, warehouse_id, and date"},
		)
	case errors.Is(err, shippingdomain.ErrManifestScanCodeRequired):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Scan code is required",
			map[string]any{"required": "code"},
		)
	case errors.Is(err, shippingdomain.ErrManifestShipmentNotPacked):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Shipment must be packed before adding to manifest", nil)
	case errors.Is(err, shippingdomain.ErrManifestCarrierMismatch):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Shipment carrier does not match carrier manifest", nil)
	case errors.Is(err, shippingdomain.ErrManifestDuplicateShipment):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Shipment already exists in carrier manifest", nil)
	case errors.Is(err, shippingdomain.ErrManifestAlreadyCompleted):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier manifest is already completed", nil)
	case errors.Is(err, shippingdomain.ErrManifestNoMissingOrders):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier manifest has no missing orders", nil)
	case errors.Is(err, shippingdomain.ErrManifestMissingOrders):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Carrier manifest has missing orders",
			map[string]any{"missing_lines": "scan all expected manifest lines before confirming handover"},
		)
	case errors.Is(err, shippingdomain.ErrManifestInvalidTransition):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier manifest status transition is invalid", nil)
	case errors.Is(err, shippingapp.ErrCarrierInactive):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier is inactive", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier manifest request could not be processed", nil)
	}
}

func writeReturnReceiptError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, returnsapp.ErrReturnReceiptNotFound):
		response.WriteError(
			w,
			r,
			http.StatusNotFound,
			response.ErrorCodeNotFound,
			"Return receipt not found",
			nil,
		)
	case errors.Is(err, returnsdomain.ErrReturnReceiptScanCodeRequired):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Return scan code is required",
			map[string]any{"required": "code"},
		)
	case errors.Is(err, returnsdomain.ErrReturnReceiptInvalidDisposition):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Return disposition is invalid",
			map[string]any{"allowed": "reusable, not_reusable, needs_inspection"},
		)
	case errors.Is(err, returnsdomain.ErrReturnInspectionInvalidCondition):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Return inspection condition is invalid",
			map[string]any{"allowed": "intact, dented_box, seal_torn, used, damaged, missing_accessory"},
		)
	case errors.Is(err, returnsdomain.ErrReturnInspectionRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid return inspection payload",
			map[string]any{"required": "receipt_id, condition, disposition, and inspector"},
		)
	case errors.Is(err, returnsdomain.ErrReturnDispositionRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid return disposition payload",
			map[string]any{"required": "receipt_id, disposition, and actor"},
		)
	case errors.Is(err, returnsdomain.ErrReturnAttachmentRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid return attachment payload",
			map[string]any{"required": "receipt_id, inspection_id, file, and actor"},
		)
	case errors.Is(err, returnsdomain.ErrReturnAttachmentInvalidFileType):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Return attachment file type is invalid",
			map[string]any{"allowed": "image/jpeg, image/png, image/webp, video/mp4, video/quicktime"},
		)
	case errors.Is(err, returnsdomain.ErrReturnAttachmentInvalidFileSize):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Return attachment file size is invalid",
			map[string]any{"max_file_size_bytes": returnsdomain.ReturnAttachmentMaxFileSizeBytes},
		)
	case errors.Is(err, returnsdomain.ErrReturnReceiptRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid return receiving payload",
			map[string]any{"required": "warehouse_id"},
		)
	case errors.Is(err, returnsapp.ErrExpectedReturnOrderNotReturnable):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Order status is not eligible for return receiving",
			map[string]any{"allowed_order_status": "handed_over, delivered"},
		)
	case errors.Is(err, returnsapp.ErrReturnReceiptDuplicate):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Return receipt already exists for this scan",
			nil,
		)
	case errors.Is(err, returnsapp.ErrReturnReceiptNotInspectable):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Return receipt is not pending inspection",
			map[string]any{"required_status": "pending_inspection"},
		)
	case errors.Is(err, returnsapp.ErrReturnReceiptDispositionNotAllowed):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Return receipt must be inspected before disposition",
			map[string]any{"required_status": "inspected"},
		)
	case errors.Is(err, returnsapp.ErrReturnInspectionNotFound):
		response.WriteError(
			w,
			r,
			http.StatusNotFound,
			response.ErrorCodeNotFound,
			"Return inspection not found",
			nil,
		)
	case errors.Is(err, returnsapp.ErrReturnAttachmentNotAllowed):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Return attachment must be linked to an inspected return",
			map[string]any{"required_status": "inspected or dispositioned"},
		)
	case errors.Is(err, returnsapp.ErrReturnAttachmentStorageUnavailable):
		response.WriteError(
			w,
			r,
			http.StatusServiceUnavailable,
			response.ErrorCodeConflict,
			"Return attachment storage is unavailable",
			nil,
		)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Return receipt could not be processed", nil)
	}
}

func writeProductError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, masterdataapp.ErrItemNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Product master data was not found", nil)
	case errors.Is(err, masterdataapp.ErrDuplicateItemCode):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Item code already exists",
			map[string]any{"field": "item_code"},
		)
	case errors.Is(err, masterdataapp.ErrDuplicateSKUCode):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"SKU code already exists",
			map[string]any{"field": "sku_code"},
		)
	case errors.Is(err, masterdatadomain.ErrItemRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Product master data is missing required fields",
			map[string]any{"required": "item_code, sku_code, name, item_type, and uom_base"},
		)
	case errors.Is(err, masterdatadomain.ErrItemInvalidType):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Product item type is invalid",
			map[string]any{"allowed": "raw_material, packaging, semi_finished, finished_good, service"},
		)
	case errors.Is(err, masterdatadomain.ErrItemInvalidStatus):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Product status is invalid",
			map[string]any{"allowed": "draft, active, inactive, obsolete"},
		)
	case errors.Is(err, masterdatadomain.ErrItemInvalidShelfLife):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Shelf life days must be positive when expiry control is enabled",
			map[string]any{"field": "shelf_life_days"},
		)
	case errors.Is(err, masterdatadomain.ErrItemInvalidCost):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Standard cost cannot be negative",
			map[string]any{"field": "standard_cost"},
		)
	default:
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Product master data request could not be processed",
			nil,
		)
	}
}

func writeWarehouseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, masterdataapp.ErrWarehouseNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Warehouse master data was not found", nil)
	case errors.Is(err, masterdataapp.ErrLocationNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Warehouse location was not found", nil)
	case errors.Is(err, masterdataapp.ErrDuplicateWarehouseCode):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Warehouse code already exists",
			map[string]any{"field": "warehouse_code"},
		)
	case errors.Is(err, masterdataapp.ErrDuplicateLocationCode):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Location code already exists for this warehouse",
			map[string]any{"field": "location_code"},
		)
	case errors.Is(err, masterdataapp.ErrInvalidLocationWarehouse):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Warehouse location references an invalid warehouse",
			map[string]any{"field": "warehouse_id"},
		)
	case errors.Is(err, masterdataapp.ErrInactiveLocation):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Inactive warehouse location cannot be edited except by reactivating it",
			map[string]any{"field": "status"},
		)
	case errors.Is(err, masterdatadomain.ErrWarehouseRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Warehouse master data is missing required fields",
			map[string]any{"required": "warehouse_code, warehouse_name, warehouse_type, and site_code"},
		)
	case errors.Is(err, masterdatadomain.ErrWarehouseInvalidType):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Warehouse type is invalid",
			map[string]any{"allowed": "raw_material, packaging, semi_finished, finished_good, quarantine, sample, defect, retail_store"},
		)
	case errors.Is(err, masterdatadomain.ErrWarehouseInvalidStatus):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Warehouse status is invalid",
			map[string]any{"allowed": "active, inactive"},
		)
	case errors.Is(err, masterdatadomain.ErrLocationRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Warehouse location is missing required fields",
			map[string]any{"required": "warehouse_id, location_code, location_name, and location_type"},
		)
	case errors.Is(err, masterdatadomain.ErrLocationInvalidType):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Warehouse location type is invalid",
			map[string]any{"allowed": "receiving, qc_hold, storage, pick, pack, handover, return, lab, scrap"},
		)
	case errors.Is(err, masterdatadomain.ErrLocationInvalidStatus):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Warehouse location status is invalid",
			map[string]any{"allowed": "active, inactive"},
		)
	default:
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Warehouse master data request could not be processed",
			nil,
		)
	}
}

func writePartyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, masterdataapp.ErrSupplierNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Supplier master data was not found", nil)
	case errors.Is(err, masterdataapp.ErrCustomerNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Customer master data was not found", nil)
	case errors.Is(err, masterdataapp.ErrDuplicateSupplierCode):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Supplier code already exists",
			map[string]any{"field": "supplier_code"},
		)
	case errors.Is(err, masterdataapp.ErrDuplicateCustomerCode):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Customer code already exists",
			map[string]any{"field": "customer_code"},
		)
	case errors.Is(err, masterdatadomain.ErrSupplierRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Supplier master data is missing required fields",
			map[string]any{"required": "supplier_code, supplier_name, and supplier_group"},
		)
	case errors.Is(err, masterdatadomain.ErrSupplierInvalidGroup):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Supplier group is invalid",
			map[string]any{"allowed": "raw_material, packaging, service, logistics, outsource"},
		)
	case errors.Is(err, masterdatadomain.ErrSupplierInvalidStatus):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Supplier status is invalid",
			map[string]any{"allowed": "draft, active, inactive, blacklisted"},
		)
	case errors.Is(err, masterdatadomain.ErrSupplierInvalidMetric):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Supplier metrics cannot be negative",
			map[string]any{"fields": "lead_time_days, moq, quality_score, delivery_score"},
		)
	case errors.Is(err, masterdatadomain.ErrSupplierInvalidStatusTransition):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Supplier status transition is invalid",
			map[string]any{"field": "status"},
		)
	case errors.Is(err, masterdatadomain.ErrCustomerRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Customer master data is missing required fields",
			map[string]any{"required": "customer_code, customer_name, and customer_type"},
		)
	case errors.Is(err, masterdatadomain.ErrCustomerInvalidType):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Customer type is invalid",
			map[string]any{"allowed": "distributor, dealer, retail_customer, marketplace, internal_store"},
		)
	case errors.Is(err, masterdatadomain.ErrCustomerInvalidStatus):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Customer status is invalid",
			map[string]any{"allowed": "draft, active, inactive, blocked"},
		)
	case errors.Is(err, masterdatadomain.ErrCustomerInvalidCreditLimit):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Customer credit limit cannot be negative",
			map[string]any{"field": "credit_limit"},
		)
	case errors.Is(err, masterdatadomain.ErrCustomerInvalidStatusTransition):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Customer status transition is invalid",
			map[string]any{"field": "status"},
		)
	default:
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Party master data request could not be processed",
			nil,
		)
	}
}

func writePermissionDenied(w http.ResponseWriter, r *http.Request, permission auth.PermissionKey) {
	response.WriteError(
		w,
		r,
		http.StatusForbidden,
		response.ErrorCodeForbidden,
		"Permission denied",
		map[string]any{"permission": string(permission)},
	)
}

func writeRoleDenied(w http.ResponseWriter, r *http.Request, roles ...auth.RoleKey) {
	response.WriteError(
		w,
		r,
		http.StatusForbidden,
		response.ErrorCodeForbidden,
		"Permission denied",
		map[string]any{"roles": roleKeyStrings(roles)},
	)
}

func hasAnyRole(principal auth.Principal, roles ...auth.RoleKey) bool {
	for _, role := range roles {
		if principal.Role == role {
			return true
		}
	}

	return false
}

func roleKeyStrings(roles []auth.RoleKey) []string {
	values := make([]string, 0, len(roles))
	for _, role := range roles {
		values = append(values, string(role))
	}

	return values
}

func newProductResponse(item masterdatadomain.Item, auditLogID string) productResponse {
	return productResponse{
		ID:               item.ID,
		ItemCode:         item.ItemCode,
		SKUCode:          item.SKUCode,
		Name:             item.Name,
		ItemType:         string(item.Type),
		ItemGroup:        item.Group,
		BrandCode:        item.BrandCode,
		UOMCode:          item.UOMBase,
		UOMBase:          item.UOMBase,
		UOMPurchase:      item.UOMPurchase,
		UOMIssue:         item.UOMIssue,
		LotControlled:    item.LotControlled,
		ExpiryControlled: item.ExpiryControlled,
		ShelfLifeDays:    item.ShelfLifeDays,
		QCRequired:       item.QCRequired,
		Status:           string(item.Status),
		StandardCost:     item.StandardCost.String(),
		IsSellable:       item.IsSellable,
		IsPurchasable:    item.IsPurchasable,
		IsProducible:     item.IsProducible,
		SpecVersion:      item.SpecVersion,
		CreatedAt:        item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        item.UpdatedAt.Format(time.RFC3339),
		AuditLogID:       auditLogID,
	}
}

func newWarehouseResponse(warehouse masterdatadomain.Warehouse, auditLogID string) warehouseResponse {
	return warehouseResponse{
		ID:              warehouse.ID,
		WarehouseCode:   warehouse.Code,
		WarehouseName:   warehouse.Name,
		WarehouseType:   string(warehouse.Type),
		SiteCode:        warehouse.SiteCode,
		Address:         warehouse.Address,
		AllowSaleIssue:  warehouse.AllowSaleIssue,
		AllowProdIssue:  warehouse.AllowProdIssue,
		AllowQuarantine: warehouse.AllowQuarantine,
		Status:          string(warehouse.Status),
		CreatedAt:       warehouse.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       warehouse.UpdatedAt.Format(time.RFC3339),
		AuditLogID:      auditLogID,
	}
}

func newWarehouseLocationResponse(location masterdatadomain.Location, auditLogID string) warehouseLocationResponse {
	return warehouseLocationResponse{
		ID:            location.ID,
		WarehouseID:   location.WarehouseID,
		WarehouseCode: location.WarehouseCode,
		LocationCode:  location.Code,
		LocationName:  location.Name,
		LocationType:  string(location.Type),
		ZoneCode:      location.ZoneCode,
		AllowReceive:  location.AllowReceive,
		AllowPick:     location.AllowPick,
		AllowStore:    location.AllowStore,
		IsDefault:     location.IsDefault,
		Status:        string(location.Status),
		CreatedAt:     location.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     location.UpdatedAt.Format(time.RFC3339),
		AuditLogID:    auditLogID,
	}
}

func newSupplierResponse(supplier masterdatadomain.Supplier, auditLogID string) supplierResponse {
	return supplierResponse{
		ID:            supplier.ID,
		SupplierCode:  supplier.Code,
		SupplierName:  supplier.Name,
		SupplierGroup: string(supplier.Group),
		ContactName:   supplier.ContactName,
		Phone:         supplier.Phone,
		Email:         supplier.Email,
		TaxCode:       supplier.TaxCode,
		Address:       supplier.Address,
		PaymentTerms:  supplier.PaymentTerms,
		LeadTimeDays:  supplier.LeadTimeDays,
		MOQ:           supplier.MOQ.String(),
		QualityScore:  supplier.QualityScore.String(),
		DeliveryScore: supplier.DeliveryScore.String(),
		Status:        string(supplier.Status),
		CreatedAt:     supplier.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     supplier.UpdatedAt.Format(time.RFC3339),
		AuditLogID:    auditLogID,
	}
}

func newCustomerResponse(customer masterdatadomain.Customer, auditLogID string) customerResponse {
	return customerResponse{
		ID:            customer.ID,
		CustomerCode:  customer.Code,
		CustomerName:  customer.Name,
		CustomerType:  string(customer.Type),
		ChannelCode:   customer.ChannelCode,
		PriceListCode: customer.PriceListCode,
		DiscountGroup: customer.DiscountGroup,
		CreditLimit:   customer.CreditLimit.String(),
		PaymentTerms:  customer.PaymentTerms,
		ContactName:   customer.ContactName,
		Phone:         customer.Phone,
		Email:         customer.Email,
		TaxCode:       customer.TaxCode,
		Address:       customer.Address,
		Status:        string(customer.Status),
		CreatedAt:     customer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     customer.UpdatedAt.Format(time.RFC3339),
		AuditLogID:    auditLogID,
	}
}

func salesOrderFilterFromRequest(r *http.Request) salesapp.SalesOrderFilter {
	statuses := make([]salesdomain.SalesOrderStatus, 0)
	for _, raw := range r.URL.Query()["status"] {
		for _, value := range strings.Split(raw, ",") {
			status := salesdomain.NormalizeSalesOrderStatus(salesdomain.SalesOrderStatus(value))
			if status != "" {
				statuses = append(statuses, status)
			}
		}
	}

	return salesapp.SalesOrderFilter{
		Search:      r.URL.Query().Get("q"),
		Statuses:    statuses,
		CustomerID:  r.URL.Query().Get("customer_id"),
		Channel:     r.URL.Query().Get("channel"),
		WarehouseID: r.URL.Query().Get("warehouse_id"),
		DateFrom:    r.URL.Query().Get("date_from"),
		DateTo:      r.URL.Query().Get("date_to"),
	}
}

func salesOrderLineInputs(lines []salesOrderLineRequest) []salesapp.SalesOrderLineInput {
	if lines == nil {
		return nil
	}

	inputs := make([]salesapp.SalesOrderLineInput, 0, len(lines))
	for _, line := range lines {
		inputs = append(inputs, salesapp.SalesOrderLineInput{
			ID:                 line.ID,
			LineNo:             line.LineNo,
			ItemID:             line.ItemID,
			OrderedQty:         line.OrderedQty,
			UOMCode:            line.UOMCode,
			UnitPrice:          line.UnitPrice,
			CurrencyCode:       line.CurrencyCode,
			LineDiscountAmount: line.LineDiscountAmount,
			BatchID:            line.BatchID,
			BatchNo:            line.BatchNo,
		})
	}

	return inputs
}

func newSalesOrderListItemResponse(order salesdomain.SalesOrder) salesOrderListItemResponse {
	return salesOrderListItemResponse{
		ID:                order.ID,
		OrderNo:           order.OrderNo,
		CustomerID:        order.CustomerID,
		CustomerCode:      order.CustomerCode,
		CustomerName:      order.CustomerName,
		Channel:           order.Channel,
		WarehouseID:       order.WarehouseID,
		WarehouseCode:     order.WarehouseCode,
		OrderDate:         order.OrderDate,
		Status:            string(order.Status),
		CurrencyCode:      order.CurrencyCode.String(),
		TotalAmount:       order.TotalAmount.String(),
		LineCount:         len(order.Lines),
		ReservedLineCount: countReservedSalesOrderLines(order),
		CreatedAt:         order.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         order.UpdatedAt.UTC().Format(time.RFC3339),
		Version:           order.Version,
	}
}

func newSalesOrderResponse(order salesdomain.SalesOrder, auditLogID string) salesOrderResponse {
	payload := salesOrderResponse{
		ID:                order.ID,
		OrderNo:           order.OrderNo,
		CustomerID:        order.CustomerID,
		CustomerCode:      order.CustomerCode,
		CustomerName:      order.CustomerName,
		Channel:           order.Channel,
		WarehouseID:       order.WarehouseID,
		WarehouseCode:     order.WarehouseCode,
		OrderDate:         order.OrderDate,
		Status:            string(order.Status),
		CurrencyCode:      order.CurrencyCode.String(),
		SubtotalAmount:    order.SubtotalAmount.String(),
		DiscountAmount:    order.DiscountAmount.String(),
		TaxAmount:         order.TaxAmount.String(),
		ShippingFeeAmount: order.ShippingFeeAmount.String(),
		NetAmount:         order.NetAmount.String(),
		TotalAmount:       order.TotalAmount.String(),
		Note:              order.Note,
		Lines:             make([]salesOrderLineResponse, 0, len(order.Lines)),
		AuditLogID:        auditLogID,
		CreatedAt:         order.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         order.UpdatedAt.UTC().Format(time.RFC3339),
		CancelReason:      order.CancelReason,
		Version:           order.Version,
	}
	if !order.ConfirmedAt.IsZero() {
		payload.ConfirmedAt = order.ConfirmedAt.UTC().Format(time.RFC3339)
	}
	if !order.CancelledAt.IsZero() {
		payload.CancelledAt = order.CancelledAt.UTC().Format(time.RFC3339)
	}
	for _, line := range order.Lines {
		payload.Lines = append(payload.Lines, salesOrderLineResponse{
			ID:                 line.ID,
			LineNo:             line.LineNo,
			ItemID:             line.ItemID,
			SKUCode:            line.SKUCode,
			ItemName:           line.ItemName,
			OrderedQty:         line.OrderedQty.String(),
			UOMCode:            line.UOMCode.String(),
			BaseOrderedQty:     line.BaseOrderedQty.String(),
			BaseUOMCode:        line.BaseUOMCode.String(),
			ConversionFactor:   line.ConversionFactor.String(),
			UnitPrice:          line.UnitPrice.String(),
			CurrencyCode:       line.CurrencyCode.String(),
			LineDiscountAmount: line.LineDiscountAmount.String(),
			LineAmount:         line.LineAmount.String(),
			ReservedQty:        line.ReservedQty.String(),
			ShippedQty:         line.ShippedQty.String(),
			BatchID:            line.BatchID,
			BatchNo:            line.BatchNo,
		})
	}

	return payload
}

func countReservedSalesOrderLines(order salesdomain.SalesOrder) int {
	count := 0
	for _, line := range order.Lines {
		if !line.ReservedQty.IsZero() {
			count++
		}
	}

	return count
}

func newWarehouseFulfillmentMetricsResponse(
	orders []salesdomain.SalesOrder,
	manifests []shippingdomain.CarrierManifest,
	warehouseID string,
	date string,
	shiftCode string,
	carrierCode string,
	generatedAt time.Time,
) warehouseFulfillmentMetricsResponse {
	manifestOrderNos := carrierManifestOrderNoSet(manifests)
	filterByCarrier := strings.TrimSpace(carrierCode) != ""
	missingOrderNos := carrierManifestMissingOrderNoSet(manifests)
	payload := warehouseFulfillmentMetricsResponse{
		WarehouseID:   strings.TrimSpace(warehouseID),
		Date:          strings.TrimSpace(date),
		ShiftCode:     strings.TrimSpace(shiftCode),
		CarrierCode:   strings.TrimSpace(carrierCode),
		GeneratedAt:   generatedAt.UTC().Format(time.RFC3339),
		MissingOrders: len(missingOrderNos),
	}

	for _, order := range orders {
		if filterByCarrier && !stringSetContains(manifestOrderNos, order.OrderNo) {
			continue
		}

		payload.TotalOrders++
		switch salesdomain.NormalizeSalesOrderStatus(order.Status) {
		case salesdomain.SalesOrderStatusDraft, salesdomain.SalesOrderStatusConfirmed:
			payload.NewOrders++
		case salesdomain.SalesOrderStatusReserved:
			payload.ReservedOrders++
		case salesdomain.SalesOrderStatusPicking, salesdomain.SalesOrderStatusPicked, salesdomain.SalesOrderStatusPacking:
			payload.PickingOrders++
		case salesdomain.SalesOrderStatusPacked:
			payload.PackedOrders++
		case salesdomain.SalesOrderStatusWaitingHandover:
			payload.WaitingHandoverOrders++
		case salesdomain.SalesOrderStatusHandedOver:
			payload.HandoverOrders++
		case salesdomain.SalesOrderStatusHandoverException:
			if orderNo := strings.TrimSpace(order.OrderNo); orderNo != "" {
				missingOrderNos[orderNo] = struct{}{}
			}
		}
	}
	payload.MissingOrders = len(missingOrderNos)

	return payload
}

func newWarehouseInboundMetricsResponse(
	orders []purchasedomain.PurchaseOrder,
	receipts []domain.WarehouseReceiving,
	inspections []qcdomain.InboundQCInspection,
	rejections []domain.SupplierRejection,
	warehouseID string,
	date string,
	shiftCode string,
	generatedAt time.Time,
) warehouseInboundMetricsResponse {
	payload := warehouseInboundMetricsResponse{
		WarehouseID: strings.TrimSpace(warehouseID),
		Date:        strings.TrimSpace(date),
		ShiftCode:   strings.TrimSpace(shiftCode),
		GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
	}

	for _, order := range orders {
		if !matchesWarehouseID(order.WarehouseID, warehouseID) || !matchesBusinessDateString(order.ExpectedDate, date) {
			continue
		}
		switch purchasedomain.NormalizePurchaseOrderStatus(order.Status) {
		case purchasedomain.PurchaseOrderStatusApproved, purchasedomain.PurchaseOrderStatusPartiallyReceived:
			payload.PurchaseOrdersIncoming++
		}
	}

	for _, receipt := range receipts {
		if !matchesWarehouseID(receipt.WarehouseID, warehouseID) || !matchesBusinessDateTime(receipt.CreatedAt, date) {
			continue
		}
		switch domain.NormalizeWarehouseReceivingStatus(receipt.Status) {
		case domain.WarehouseReceivingStatusDraft:
			payload.ReceivingDraft++
			payload.ReceivingPending++
		case domain.WarehouseReceivingStatusSubmitted:
			payload.ReceivingSubmitted++
			payload.ReceivingPending++
		case domain.WarehouseReceivingStatusInspectReady:
			payload.ReceivingInspectReady++
			payload.ReceivingPending++
		}
	}

	for _, inspection := range inspections {
		if !matchesWarehouseID(inspection.WarehouseID, warehouseID) || !matchesBusinessDateTime(inspection.CreatedAt, date) {
			continue
		}
		if qcdomain.NormalizeInboundQCInspectionStatus(inspection.Status) != qcdomain.InboundQCInspectionStatusCompleted {
			continue
		}
		switch qcdomain.NormalizeInboundQCResult(inspection.Result) {
		case qcdomain.InboundQCResultHold:
			payload.QCHold++
		case qcdomain.InboundQCResultFail:
			payload.QCFail++
		case qcdomain.InboundQCResultPass:
			payload.QCPass++
		case qcdomain.InboundQCResultPartial:
			payload.QCPartial++
		}
	}

	for _, rejection := range rejections {
		if !matchesWarehouseID(rejection.WarehouseID, warehouseID) || !matchesBusinessDateTime(rejection.CreatedAt, date) {
			continue
		}
		switch domain.NormalizeSupplierRejectionStatus(rejection.Status) {
		case domain.SupplierRejectionStatusDraft:
			payload.SupplierRejectionDraft++
			payload.SupplierRejections++
		case domain.SupplierRejectionStatusSubmitted:
			payload.SupplierRejectionSubmitted++
			payload.SupplierRejections++
		case domain.SupplierRejectionStatusConfirmed:
			payload.SupplierRejectionConfirmed++
			payload.SupplierRejections++
		case domain.SupplierRejectionStatusCancelled:
			payload.SupplierRejectionCancelled++
		}
	}

	return payload
}

func newWarehouseSubcontractMetricsResponse(
	ctx context.Context,
	orders []productiondomain.SubcontractOrder,
	materialTransfers warehouseDailyBoardSubcontractMaterialTransferLister,
	factoryClaims warehouseDailyBoardSubcontractFactoryClaimLister,
	paymentMilestones warehouseDailyBoardSubcontractPaymentMilestoneLister,
	warehouseID string,
	date string,
	shiftCode string,
	generatedAt time.Time,
) (warehouseSubcontractMetricsResponse, error) {
	if materialTransfers == nil || factoryClaims == nil || paymentMilestones == nil {
		return warehouseSubcontractMetricsResponse{}, errors.New("subcontract daily board dependencies are required")
	}

	payload := warehouseSubcontractMetricsResponse{
		WarehouseID: strings.TrimSpace(warehouseID),
		Date:        strings.TrimSpace(date),
		ShiftCode:   strings.TrimSpace(shiftCode),
		GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
	}
	materialIssuedOrderIDs := make(map[string]struct{})
	finalPaymentReadyOrderIDs := make(map[string]struct{})

	for _, order := range orders {
		if subcontractOrderOpenForDailyBoard(order, date) {
			payload.OpenOrders++
		}
		if subcontractOrderSamplePendingForDailyBoard(order, date) {
			payload.SamplePending++
		}
		if subcontractOrderFinalPaymentReadyOnDate(order, date) {
			finalPaymentReadyOrderIDs[order.ID] = struct{}{}
		}

		transfers, err := materialTransfers.ListBySubcontractOrder(ctx, order.ID)
		if err != nil {
			return warehouseSubcontractMetricsResponse{}, err
		}
		matchedTransfers := 0
		for _, transfer := range transfers {
			if !matchesWarehouseScopeID(transfer.SourceWarehouseID, warehouseID) ||
				!matchesBusinessDateTime(transfer.HandoverAt, date) {
				continue
			}
			matchedTransfers++
		}
		payload.MaterialTransferCount += matchedTransfers
		if matchedTransfers > 0 || (strings.TrimSpace(warehouseID) == "" && subcontractOrderMaterialIssuedOnDate(order, date)) {
			materialIssuedOrderIDs[order.ID] = struct{}{}
		}

		claims, err := factoryClaims.ListBySubcontractOrder(ctx, order.ID)
		if err != nil {
			return warehouseSubcontractMetricsResponse{}, err
		}
		for _, claim := range claims {
			if !claim.BlocksFinalPayment() || !matchesBusinessDateTimeOnOrBefore(claim.OpenedAt, date) {
				continue
			}
			payload.FactoryClaims++
			if claim.IsOverdue(generatedAt) {
				payload.FactoryClaimsOverdue++
			}
		}

		milestones, err := paymentMilestones.ListBySubcontractOrder(ctx, order.ID)
		if err != nil {
			return warehouseSubcontractMetricsResponse{}, err
		}
		for _, milestone := range milestones {
			if subcontractPaymentMilestoneFinalReadyOnDate(milestone, date) {
				finalPaymentReadyOrderIDs[order.ID] = struct{}{}
			}
		}
	}

	payload.MaterialIssuedOrders = len(materialIssuedOrderIDs)
	payload.FinalPaymentReadyOrders = len(finalPaymentReadyOrderIDs)

	return payload, nil
}

func matchesWarehouseID(rowWarehouseID string, filterWarehouseID string) bool {
	filterWarehouseID = strings.TrimSpace(filterWarehouseID)
	if filterWarehouseID == "" {
		return true
	}

	return strings.TrimSpace(rowWarehouseID) == filterWarehouseID
}

func matchesWarehouseScopeID(rowWarehouseID string, filterWarehouseID string) bool {
	filterWarehouseID = strings.TrimSpace(filterWarehouseID)
	if filterWarehouseID == "" {
		return true
	}
	rowWarehouseID = strings.TrimSpace(rowWarehouseID)

	return rowWarehouseID == filterWarehouseID || strings.HasPrefix(rowWarehouseID, filterWarehouseID+"-")
}

func matchesBusinessDateString(rowDate string, filterDate string) bool {
	filterDate = strings.TrimSpace(filterDate)
	if filterDate == "" {
		return true
	}

	return strings.TrimSpace(rowDate) == filterDate
}

func matchesBusinessDateTime(rowTime time.Time, filterDate string) bool {
	filterDate = strings.TrimSpace(filterDate)
	if filterDate == "" {
		return true
	}
	if rowTime.IsZero() {
		return false
	}

	return businessDate(rowTime) == filterDate
}

func matchesBusinessDateTimeOnOrBefore(rowTime time.Time, filterDate string) bool {
	filterDate = strings.TrimSpace(filterDate)
	if filterDate == "" {
		return true
	}
	if rowTime.IsZero() {
		return false
	}

	return businessDate(rowTime) <= filterDate
}

func matchesBusinessDateStringOnOrBefore(rowDate string, filterDate string) bool {
	filterDate = strings.TrimSpace(filterDate)
	if filterDate == "" {
		return true
	}
	rowDate = strings.TrimSpace(rowDate)
	if rowDate == "" {
		return true
	}

	return rowDate <= filterDate
}

func subcontractOrderOpenForDailyBoard(order productiondomain.SubcontractOrder, date string) bool {
	switch productiondomain.NormalizeSubcontractOrderStatus(order.Status) {
	case productiondomain.SubcontractOrderStatusClosed, productiondomain.SubcontractOrderStatusCancelled:
		return false
	default:
		return matchesBusinessDateStringOnOrBefore(order.ExpectedReceiptDate, date)
	}
}

func subcontractOrderMaterialIssuedOnDate(order productiondomain.SubcontractOrder, date string) bool {
	status := productiondomain.NormalizeSubcontractOrderStatus(order.Status)
	if !subcontractOrderStatusAtOrAfterMaterialsIssued(status) {
		return false
	}

	return matchesBusinessDateTime(order.MaterialsIssuedAt, date)
}

func subcontractOrderSamplePendingForDailyBoard(order productiondomain.SubcontractOrder, date string) bool {
	if !order.SampleRequired {
		return false
	}
	switch productiondomain.NormalizeSubcontractOrderStatus(order.Status) {
	case productiondomain.SubcontractOrderStatusMaterialsIssued:
		return matchesBusinessDateTimeOnOrBefore(firstNonZeroTime(order.MaterialsIssuedAt, order.UpdatedAt), date)
	case productiondomain.SubcontractOrderStatusSampleSubmitted:
		return matchesBusinessDateTimeOnOrBefore(firstNonZeroTime(order.SampleSubmittedAt, order.UpdatedAt), date)
	case productiondomain.SubcontractOrderStatusSampleRejected:
		return matchesBusinessDateTimeOnOrBefore(firstNonZeroTime(order.SampleRejectedAt, order.UpdatedAt), date)
	default:
		return false
	}
}

func subcontractOrderFinalPaymentReadyOnDate(order productiondomain.SubcontractOrder, date string) bool {
	if productiondomain.NormalizeSubcontractOrderStatus(order.Status) != productiondomain.SubcontractOrderStatusFinalPaymentReady {
		return false
	}

	return matchesBusinessDateTime(order.FinalPaymentReadyAt, date)
}

func subcontractPaymentMilestoneFinalReadyOnDate(
	milestone productiondomain.SubcontractPaymentMilestone,
	date string,
) bool {
	if productiondomain.NormalizeSubcontractPaymentMilestoneKind(milestone.Kind) != productiondomain.SubcontractPaymentMilestoneKindFinalPayment ||
		productiondomain.NormalizeSubcontractPaymentMilestoneStatus(milestone.Status) != productiondomain.SubcontractPaymentMilestoneStatusReady {
		return false
	}

	return matchesBusinessDateTime(milestone.ReadyAt, date)
}

func subcontractOrderStatusAtOrAfterMaterialsIssued(status productiondomain.SubcontractOrderStatus) bool {
	switch status {
	case productiondomain.SubcontractOrderStatusMaterialsIssued,
		productiondomain.SubcontractOrderStatusSampleSubmitted,
		productiondomain.SubcontractOrderStatusSampleApproved,
		productiondomain.SubcontractOrderStatusSampleRejected,
		productiondomain.SubcontractOrderStatusMassProductionStarted,
		productiondomain.SubcontractOrderStatusFinishedGoodsReceived,
		productiondomain.SubcontractOrderStatusQCInProgress,
		productiondomain.SubcontractOrderStatusAccepted,
		productiondomain.SubcontractOrderStatusRejectedFactoryIssue,
		productiondomain.SubcontractOrderStatusFinalPaymentReady,
		productiondomain.SubcontractOrderStatusClosed:
		return true
	default:
		return false
	}
}

func firstNonZeroTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value
		}
	}

	return time.Time{}
}

func businessDate(value time.Time) string {
	loc, err := time.LoadLocation(decimal.TimezoneHoChiMinh)
	if err != nil {
		loc = time.FixedZone(decimal.TimezoneHoChiMinh, 7*60*60)
	}

	return value.In(loc).Format("2006-01-02")
}

func carrierManifestOrderNoSet(manifests []shippingdomain.CarrierManifest) map[string]struct{} {
	orderNos := make(map[string]struct{})
	for _, manifest := range manifests {
		for _, line := range manifest.Lines {
			if orderNo := strings.TrimSpace(line.OrderNo); orderNo != "" {
				orderNos[orderNo] = struct{}{}
			}
		}
	}

	return orderNos
}

func carrierManifestMissingOrderNoSet(manifests []shippingdomain.CarrierManifest) map[string]struct{} {
	orderNos := make(map[string]struct{})
	for _, manifest := range manifests {
		for _, line := range manifest.MissingLines() {
			if orderNo := strings.TrimSpace(line.OrderNo); orderNo != "" {
				orderNos[orderNo] = struct{}{}
			}
		}
	}

	return orderNos
}

func stringSetContains(values map[string]struct{}, value string) bool {
	_, ok := values[strings.TrimSpace(value)]
	return ok
}

func newAuditLogResponse(log audit.Log) auditLogResponse {
	metadata := log.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	return auditLogResponse{
		ID:         log.ID,
		ActorID:    log.ActorID,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		RequestID:  log.RequestID,
		BeforeData: log.BeforeData,
		AfterData:  log.AfterData,
		Metadata:   metadata,
		CreatedAt:  log.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func validateStockMovementPayload(payload stockMovementRequest) map[string]any {
	details := make(map[string]any)
	if strings.TrimSpace(payload.MovementID) == "" {
		details["movementId"] = "required"
	}
	if strings.TrimSpace(payload.SKU) == "" {
		details["sku"] = "required"
	}
	if strings.TrimSpace(payload.WarehouseID) == "" {
		details["warehouseId"] = "required"
	}
	if strings.TrimSpace(payload.Reason) == "" {
		details["reason"] = "required"
	}
	quantity, err := decimal.ParseQuantity(payload.Quantity)
	if err != nil || quantity.IsNegative() || quantity.IsZero() {
		details["quantity"] = "must be positive"
	}
	if _, err := decimal.NormalizeUOMCode(payload.BaseUOMCode); err != nil {
		details["baseUomCode"] = "required"
	}
	if strings.TrimSpace(payload.SourceQuantity) != "" {
		sourceQuantity, err := decimal.ParseQuantity(payload.SourceQuantity)
		if err != nil || sourceQuantity.IsNegative() || sourceQuantity.IsZero() {
			details["sourceQuantity"] = "must be positive"
		}
	}
	if strings.TrimSpace(payload.SourceUOMCode) != "" {
		if _, err := decimal.NormalizeUOMCode(payload.SourceUOMCode); err != nil {
			details["sourceUomCode"] = "invalid"
		}
	}
	if strings.TrimSpace(payload.ConversionFactor) != "" {
		conversionFactor, err := decimal.ParseQuantity(payload.ConversionFactor)
		if err != nil || conversionFactor.IsNegative() || conversionFactor.IsZero() {
			details["conversionFactor"] = "must be positive"
		}
	}
	switch strings.ToUpper(strings.TrimSpace(payload.MovementType)) {
	case "RECEIVE", "ISSUE", "TRANSFER_IN", "ADJUST":
	default:
		details["movementType"] = "unsupported"
	}

	return details
}

func stockMovementContractValues(payload stockMovementRequest) (decimal.Decimal, decimal.Decimal, decimal.Decimal, decimal.UOMCode, decimal.UOMCode) {
	movementQty := decimal.MustQuantity(payload.Quantity)
	baseUOMCode := decimal.MustUOMCode(payload.BaseUOMCode)
	sourceQty := movementQty
	if strings.TrimSpace(payload.SourceQuantity) != "" {
		sourceQty = decimal.MustQuantity(payload.SourceQuantity)
	}
	sourceUOMCode := baseUOMCode
	if strings.TrimSpace(payload.SourceUOMCode) != "" {
		sourceUOMCode = decimal.MustUOMCode(payload.SourceUOMCode)
	}
	conversionFactor := decimal.MustQuantity("1")
	if strings.TrimSpace(payload.ConversionFactor) != "" {
		conversionFactor = decimal.MustQuantity(payload.ConversionFactor)
	}

	return movementQty, sourceQty, conversionFactor, baseUOMCode, sourceUOMCode
}

func recordStockAdjustmentAudit(r *http.Request, store audit.LogStore, payload stockMovementRequest) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return http.ErrNoCookie
	}
	movementQty, sourceQty, conversionFactor, baseUOMCode, sourceUOMCode := stockMovementContractValues(payload)

	log, err := audit.NewLog(audit.NewLogInput{
		ActorID:    principal.UserID,
		Action:     "inventory.stock_movement.adjusted",
		EntityType: "inventory.stock_movement",
		EntityID:   strings.TrimSpace(payload.MovementID),
		RequestID:  response.RequestID(r),
		AfterData: map[string]any{
			"movement_type":     strings.ToUpper(strings.TrimSpace(payload.MovementType)),
			"movement_qty":      movementQty.String(),
			"base_uom_code":     baseUOMCode.String(),
			"source_qty":        sourceQty.String(),
			"source_uom_code":   sourceUOMCode.String(),
			"conversion_factor": conversionFactor.String(),
			"warehouse_id":      strings.TrimSpace(payload.WarehouseID),
			"sku":               strings.ToUpper(strings.TrimSpace(payload.SKU)),
		},
		Metadata: map[string]any{
			"reason": strings.TrimSpace(payload.Reason),
			"source": "inventory stock movement",
		},
	})
	if err != nil {
		return err
	}

	return store.Record(r.Context(), log)
}

func queryInt(r *http.Request, key string) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return parsed
}

func requestWithStableID(r *http.Request) *http.Request {
	if strings.TrimSpace(r.Header.Get(response.HeaderRequestID)) != "" {
		return r
	}

	clone := r.Clone(r.Context())
	clone.Header.Set(response.HeaderRequestID, response.RequestID(r))
	return clone
}
