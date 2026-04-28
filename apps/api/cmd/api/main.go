package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
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
	ID          string `json:"id"`
	ShipmentID  string `json:"shipment_id"`
	OrderNo     string `json:"order_no"`
	TrackingNo  string `json:"tracking_no"`
	PackageCode string `json:"package_code"`
	StagingZone string `json:"staging_zone"`
	Scanned     bool   `json:"scanned"`
}

type carrierManifestResponse struct {
	ID            string                         `json:"id"`
	CarrierCode   string                         `json:"carrier_code"`
	CarrierName   string                         `json:"carrier_name"`
	WarehouseID   string                         `json:"warehouse_id"`
	WarehouseCode string                         `json:"warehouse_code"`
	Date          string                         `json:"date"`
	HandoverBatch string                         `json:"handover_batch"`
	StagingZone   string                         `json:"staging_zone"`
	Status        string                         `json:"status"`
	Owner         string                         `json:"owner"`
	AuditLogID    string                         `json:"audit_log_id,omitempty"`
	Summary       carrierManifestSummaryResponse `json:"summary"`
	Lines         []carrierManifestLineResponse  `json:"lines"`
	CreatedAt     string                         `json:"created_at,omitempty"`
}

type createCarrierManifestRequest struct {
	ID            string `json:"id"`
	CarrierCode   string `json:"carrier_code"`
	CarrierName   string `json:"carrier_name"`
	WarehouseID   string `json:"warehouse_id"`
	WarehouseCode string `json:"warehouse_code"`
	Date          string `json:"date"`
	HandoverBatch string `json:"handover_batch"`
	StagingZone   string `json:"staging_zone"`
	Owner         string `json:"owner"`
}

type addShipmentToManifestRequest struct {
	ShipmentID string `json:"shipment_id"`
}

type verifyCarrierManifestScanRequest struct {
	Code      string `json:"code"`
	StationID string `json:"station_id"`
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
	ID          string `json:"id"`
	ItemID      string `json:"item_id"`
	SKU         string `json:"sku"`
	ItemName    string `json:"item_name,omitempty"`
	BatchID     string `json:"batch_id,omitempty"`
	BatchNo     string `json:"batch_no,omitempty"`
	WarehouseID string `json:"warehouse_id"`
	LocationID  string `json:"location_id"`
	Quantity    string `json:"quantity"`
	BaseUOMCode string `json:"base_uom_code"`
	QCStatus    string `json:"qc_status,omitempty"`
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
	ID          string `json:"id"`
	ItemID      string `json:"item_id"`
	SKU         string `json:"sku"`
	ItemName    string `json:"item_name"`
	BatchID     string `json:"batch_id"`
	BatchNo     string `json:"batch_no"`
	Quantity    string `json:"quantity"`
	BaseUOMCode string `json:"base_uom_code"`
	QCStatus    string `json:"qc_status"`
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

type receiveReturnRequest struct {
	WarehouseID       string `json:"warehouse_id"`
	WarehouseCode     string `json:"warehouse_code"`
	Source            string `json:"source"`
	Code              string `json:"code"`
	PackageCondition  string `json:"package_condition"`
	Disposition       string `json:"disposition"`
	InvestigationNote string `json:"investigation_note"`
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
	auditLogStore := audit.NewPrototypeLogStore()
	batchCatalog := inventoryapp.NewPrototypeBatchCatalog(auditLogStore)
	itemCatalog := masterdataapp.NewPrototypeItemCatalog(auditLogStore)
	warehouseCatalog := masterdataapp.NewPrototypeWarehouseLocationCatalog(auditLogStore)
	partyCatalog := masterdataapp.NewPrototypePartyCatalog(auditLogStore)
	salesOrderStore := salesapp.NewPrototypeSalesOrderStore(auditLogStore)
	salesOrderService := salesapp.NewSalesOrderService(salesOrderStore, partyCatalog, itemCatalog, warehouseCatalog)
	stockMovementStore := inventoryapp.NewInMemoryStockMovementStore()
	warehouseReceivingStore := inventoryapp.NewPrototypeWarehouseReceivingStore()
	warehouseReceiving := inventoryapp.NewWarehouseReceivingService(
		warehouseReceivingStore,
		warehouseCatalog,
		batchCatalog,
		stockMovementStore,
		auditLogStore,
	)
	reconciliationStore := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	listEndOfDayReconciliations := inventoryapp.NewListEndOfDayReconciliations(reconciliationStore)
	closeEndOfDayReconciliation := inventoryapp.NewCloseEndOfDayReconciliation(reconciliationStore, auditLogStore)
	carrierManifestStore := shippingapp.NewPrototypeCarrierManifestStore()
	listCarrierManifests := shippingapp.NewListCarrierManifests(carrierManifestStore)
	createCarrierManifest := shippingapp.NewCreateCarrierManifest(carrierManifestStore, auditLogStore)
	addShipmentToCarrierManifest := shippingapp.NewAddShipmentToCarrierManifest(carrierManifestStore, auditLogStore)
	verifyCarrierManifestScan := shippingapp.NewVerifyCarrierManifestScan(carrierManifestStore, auditLogStore)
	returnReceiptStore := returnsapp.NewPrototypeReturnReceiptStore()
	listReturnReceipts := returnsapp.NewListReturnReceipts(returnReceiptStore)
	receiveReturn := returnsapp.NewReceiveReturn(returnReceiptStore, auditLogStore)

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
		"/api/v1/shipping/manifests/{manifest_id}/scan",
		auth.RequireSessionPermission(
			authSessions,
			auth.PermissionShippingView,
			http.HandlerFunc(verifyCarrierManifestScanHandler(verifyCarrierManifestScan)),
		),
	)
	mux.Handle(
		"/api/v1/returns/receipts",
		auth.RequireSessionToken(
			authSessions,
			http.HandlerFunc(returnReceiptsHandler(listReturnReceipts, receiveReturn)),
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
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
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
				ID:            payload.ID,
				CarrierCode:   payload.CarrierCode,
				CarrierName:   payload.CarrierName,
				WarehouseID:   payload.WarehouseID,
				WarehouseCode: payload.WarehouseCode,
				Date:          payload.Date,
				HandoverBatch: payload.HandoverBatch,
				StagingZone:   payload.StagingZone,
				Owner:         payload.Owner,
				ActorID:       principal.UserID,
				RequestID:     response.RequestID(r),
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
		result, err := service.Execute(r.Context(), shippingapp.VerifyCarrierManifestScanInput{
			ManifestID: r.PathValue("manifest_id"),
			Code:       payload.Code,
			StationID:  payload.StationID,
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
			ID:          input.ID,
			ItemID:      input.ItemID,
			SKU:         input.SKU,
			ItemName:    input.ItemName,
			BatchID:     input.BatchID,
			BatchNo:     input.BatchNo,
			Quantity:    input.Quantity,
			BaseUOMCode: input.BaseUOMCode,
			QCStatus:    input.QCStatus,
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
			ID:          line.ID,
			ItemID:      line.ItemID,
			SKU:         line.SKU,
			ItemName:    line.ItemName,
			BatchID:     line.BatchID,
			BatchNo:     line.BatchNo,
			WarehouseID: line.WarehouseID,
			LocationID:  line.LocationID,
			Quantity:    line.Quantity.String(),
			BaseUOMCode: line.BaseUOMCode.String(),
			QCStatus:    string(line.QCStatus),
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
		ID:            manifest.ID,
		CarrierCode:   manifest.CarrierCode,
		CarrierName:   manifest.CarrierName,
		WarehouseID:   manifest.WarehouseID,
		WarehouseCode: manifest.WarehouseCode,
		Date:          manifest.Date,
		HandoverBatch: manifest.HandoverBatch,
		StagingZone:   manifest.StagingZone,
		Status:        string(manifest.Status),
		Owner:         manifest.Owner,
		AuditLogID:    auditLogID,
		Summary: carrierManifestSummaryResponse{
			ExpectedCount: summary.ExpectedCount,
			ScannedCount:  summary.ScannedCount,
			MissingCount:  summary.MissingCount,
		},
		Lines: make([]carrierManifestLineResponse, 0, len(manifest.Lines)),
	}
	if !manifest.CreatedAt.IsZero() {
		payload.CreatedAt = manifest.CreatedAt.UTC().Format(time.RFC3339)
	}
	for _, line := range manifest.Lines {
		payload.Lines = append(payload.Lines, carrierManifestLineResponse{
			ID:          line.ID,
			ShipmentID:  line.ShipmentID,
			OrderNo:     line.OrderNo,
			TrackingNo:  line.TrackingNo,
			PackageCode: line.PackageCode,
			StagingZone: line.StagingZone,
			Scanned:     line.Scanned,
		})
	}

	return payload
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
			WarehouseID:        result.Event.WarehouseID,
			CarrierCode:        result.Event.CarrierCode,
			CreatedAt:          result.Event.CreatedAt.UTC().Format(time.RFC3339),
		},
		Manifest:   newCarrierManifestResponse(result.Manifest, ""),
		AuditLogID: result.AuditLogID,
	}
	if result.Line != nil {
		payload.Line = &carrierManifestLineResponse{
			ID:          result.Line.ID,
			ShipmentID:  result.Line.ShipmentID,
			OrderNo:     result.Line.OrderNo,
			TrackingNo:  result.Line.TrackingNo,
			PackageCode: result.Line.PackageCode,
			StagingZone: result.Line.StagingZone,
			Scanned:     result.Line.Scanned,
		}
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
	case errors.Is(err, domain.ErrReceivingRequiredField),
		errors.Is(err, domain.ErrReceivingInvalidStatus),
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
			map[string]any{"required": "warehouse_id, location_id, reference_doc_type, reference_doc_id, lines, quantity, and base_uom_code"},
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
	switch {
	case errors.Is(err, shippingapp.ErrCarrierManifestNotFound), errors.Is(err, shippingapp.ErrPackedShipmentNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Carrier manifest resource not found", nil)
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
	case errors.Is(err, shippingdomain.ErrManifestDuplicateShipment):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Shipment already exists in carrier manifest", nil)
	case errors.Is(err, shippingdomain.ErrManifestAlreadyCompleted):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier manifest is already completed", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier manifest request could not be processed", nil)
	}
}

func writeReturnReceiptError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
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
	case errors.Is(err, returnsdomain.ErrReturnReceiptRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid return receiving payload",
			map[string]any{"required": "warehouse_id"},
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
