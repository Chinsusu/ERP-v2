package domain

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrProductionPlanRequiredField = errors.New("production plan required field is missing")
var ErrProductionPlanInvalidStatus = errors.New("production plan status is invalid")
var ErrProductionPlanInvalidOutputType = errors.New("production plan output item type is invalid")
var ErrProductionPlanInvalidQuantity = errors.New("production plan quantity is invalid")
var ErrProductionPlanInvalidUOM = errors.New("production plan uom is invalid")
var ErrProductionPlanInvalidComponentType = errors.New("production plan component type is invalid")
var ErrProductionPlanInvalidShortage = errors.New("production plan shortage is invalid")
var ErrProductionPlanInvalidPurchaseRequestTransition = errors.New("purchase request status transition is invalid")

type ProductionPlanStatus string

const (
	ProductionPlanStatusDraft                       ProductionPlanStatus = "draft"
	ProductionPlanStatusPurchaseRequestDraftCreated ProductionPlanStatus = "purchase_request_draft_created"
	ProductionPlanStatusCancelled                   ProductionPlanStatus = "cancelled"
)

type PurchaseRequestDraftStatus string

const (
	PurchaseRequestDraftStatusDraft         PurchaseRequestDraftStatus = "draft"
	PurchaseRequestDraftStatusSubmitted     PurchaseRequestDraftStatus = "submitted"
	PurchaseRequestDraftStatusApproved      PurchaseRequestDraftStatus = "approved"
	PurchaseRequestDraftStatusConvertedToPO PurchaseRequestDraftStatus = "converted_to_po"
	PurchaseRequestDraftStatusCancelled     PurchaseRequestDraftStatus = "cancelled"
	PurchaseRequestDraftStatusRejected      PurchaseRequestDraftStatus = "rejected"
)

type ProductionPlan struct {
	ID                  string
	OrgID               string
	PlanNo              string
	OutputItemID        string
	OutputSKU           string
	OutputItemName      string
	OutputItemType      string
	PlannedQty          decimal.Decimal
	UOMCode             decimal.UOMCode
	FormulaID           string
	FormulaCode         string
	FormulaVersion      string
	FormulaBatchQty     decimal.Decimal
	FormulaBatchUOMCode decimal.UOMCode
	PlannedStartDate    string
	PlannedEndDate      string
	Status              ProductionPlanStatus
	Lines               []ProductionPlanLine
	PurchaseDraft       PurchaseRequestDraft
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	Version             int
	CancelReason        string
}

type ProductionPlanLine struct {
	ID                   string
	FormulaLineID        string
	LineNo               int
	ComponentItemID      string
	ComponentSKU         string
	ComponentName        string
	ComponentType        string
	FormulaQty           decimal.Decimal
	FormulaUOMCode       decimal.UOMCode
	RequiredQty          decimal.Decimal
	RequiredUOMCode      decimal.UOMCode
	RequiredStockBaseQty decimal.Decimal
	StockBaseUOMCode     decimal.UOMCode
	AvailableQty         decimal.Decimal
	ShortageQty          decimal.Decimal
	PurchaseDraftQty     decimal.Decimal
	PurchaseDraftUOMCode decimal.UOMCode
	IsStockManaged       bool
	NeedsPurchase        bool
	Note                 string
}

type PurchaseRequestDraft struct {
	ID                       string
	RequestNo                string
	SourceProductionPlanID   string
	SourceProductionPlanNo   string
	Status                   PurchaseRequestDraftStatus
	Lines                    []PurchaseRequestDraftLine
	CreatedAt                time.Time
	CreatedBy                string
	SubmittedAt              time.Time
	SubmittedBy              string
	ApprovedAt               time.Time
	ApprovedBy               string
	ConvertedAt              time.Time
	ConvertedBy              string
	ConvertedPurchaseOrderID string
	ConvertedPurchaseOrderNo string
	CancelledAt              time.Time
	CancelledBy              string
	RejectedAt               time.Time
	RejectedBy               string
	RejectReason             string
}

type PurchaseRequestDraftLine struct {
	ID                       string
	LineNo                   int
	SourceProductionPlanLine string
	ItemID                   string
	SKU                      string
	ItemName                 string
	RequestedQty             decimal.Decimal
	UOMCode                  decimal.UOMCode
	Note                     string
}

type NewProductionPlanDocumentInput struct {
	ID                  string
	OrgID               string
	PlanNo              string
	OutputItemID        string
	OutputSKU           string
	OutputItemName      string
	OutputItemType      string
	PlannedQty          decimal.Decimal
	UOMCode             string
	FormulaID           string
	FormulaCode         string
	FormulaVersion      string
	FormulaBatchQty     decimal.Decimal
	FormulaBatchUOMCode string
	PlannedStartDate    string
	PlannedEndDate      string
	Lines               []NewProductionPlanLineInput
	PurchaseDraftID     string
	PurchaseDraftNo     string
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
}

type NewProductionPlanLineInput struct {
	ID                   string
	FormulaLineID        string
	LineNo               int
	ComponentItemID      string
	ComponentSKU         string
	ComponentName        string
	ComponentType        string
	FormulaQty           decimal.Decimal
	FormulaUOMCode       string
	RequiredQty          decimal.Decimal
	RequiredUOMCode      string
	RequiredStockBaseQty decimal.Decimal
	StockBaseUOMCode     string
	AvailableQty         decimal.Decimal
	ShortageQty          decimal.Decimal
	PurchaseDraftQty     decimal.Decimal
	PurchaseDraftUOMCode string
	IsStockManaged       bool
	NeedsPurchase        bool
	Note                 string
}

func NewProductionPlanDocument(input NewProductionPlanDocumentInput) (ProductionPlan, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	updatedBy := strings.TrimSpace(input.UpdatedBy)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(input.CreatedBy)
	}
	plannedQty, err := normalizeProductionPlanPositiveQuantity(input.PlannedQty)
	if err != nil {
		return ProductionPlan{}, err
	}
	formulaBatchQty, err := normalizeProductionPlanPositiveQuantity(input.FormulaBatchQty)
	if err != nil {
		return ProductionPlan{}, err
	}
	uomCode, err := normalizeProductionPlanUOM(input.UOMCode)
	if err != nil {
		return ProductionPlan{}, err
	}
	formulaBatchUOMCode, err := normalizeProductionPlanUOM(input.FormulaBatchUOMCode)
	if err != nil {
		return ProductionPlan{}, err
	}

	plan := ProductionPlan{
		ID:                  strings.TrimSpace(input.ID),
		OrgID:               strings.TrimSpace(input.OrgID),
		PlanNo:              strings.ToUpper(strings.TrimSpace(input.PlanNo)),
		OutputItemID:        strings.TrimSpace(input.OutputItemID),
		OutputSKU:           strings.ToUpper(strings.TrimSpace(input.OutputSKU)),
		OutputItemName:      strings.TrimSpace(input.OutputItemName),
		OutputItemType:      normalizeProductionOutputType(input.OutputItemType),
		PlannedQty:          plannedQty,
		UOMCode:             uomCode,
		FormulaID:           strings.TrimSpace(input.FormulaID),
		FormulaCode:         strings.ToUpper(strings.TrimSpace(input.FormulaCode)),
		FormulaVersion:      strings.TrimSpace(input.FormulaVersion),
		FormulaBatchQty:     formulaBatchQty,
		FormulaBatchUOMCode: formulaBatchUOMCode,
		PlannedStartDate:    strings.TrimSpace(input.PlannedStartDate),
		PlannedEndDate:      strings.TrimSpace(input.PlannedEndDate),
		Status:              ProductionPlanStatusDraft,
		Lines:               make([]ProductionPlanLine, 0, len(input.Lines)),
		CreatedAt:           createdAt.UTC(),
		CreatedBy:           strings.TrimSpace(input.CreatedBy),
		UpdatedAt:           updatedAt.UTC(),
		UpdatedBy:           updatedBy,
		Version:             1,
	}

	for index, lineInput := range input.Lines {
		if lineInput.LineNo == 0 {
			lineInput.LineNo = index + 1
		}
		line, err := NewProductionPlanLine(lineInput)
		if err != nil {
			return ProductionPlan{}, err
		}
		plan.Lines = append(plan.Lines, line)
	}
	SortProductionPlanLines(plan.Lines)
	plan.PurchaseDraft = newPurchaseRequestDraftForPlan(plan, input.PurchaseDraftID, input.PurchaseDraftNo)
	if len(plan.PurchaseDraft.Lines) > 0 {
		plan.Status = ProductionPlanStatusPurchaseRequestDraftCreated
	}

	if err := plan.Validate(); err != nil {
		return ProductionPlan{}, err
	}

	return plan, nil
}

func NewProductionPlanLine(input NewProductionPlanLineInput) (ProductionPlanLine, error) {
	formulaQty, err := normalizeProductionPlanQuantity(input.FormulaQty, true)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	requiredQty, err := normalizeProductionPlanQuantity(input.RequiredQty, true)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	requiredStockBaseQty, err := normalizeProductionPlanQuantity(input.RequiredStockBaseQty, true)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	availableQty, err := normalizeProductionPlanQuantity(input.AvailableQty, true)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	shortageQty, err := normalizeProductionPlanQuantity(input.ShortageQty, true)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	purchaseDraftQty, err := normalizeProductionPlanQuantity(input.PurchaseDraftQty, true)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	formulaUOM, err := normalizeProductionPlanUOM(input.FormulaUOMCode)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	requiredUOM, err := normalizeProductionPlanUOM(input.RequiredUOMCode)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	stockBaseUOM, err := normalizeProductionPlanUOM(input.StockBaseUOMCode)
	if err != nil {
		return ProductionPlanLine{}, err
	}
	purchaseDraftUOMCode := strings.TrimSpace(input.PurchaseDraftUOMCode)
	if purchaseDraftUOMCode == "" {
		purchaseDraftUOMCode = stockBaseUOM.String()
	}
	purchaseDraftUOM, err := normalizeProductionPlanUOM(purchaseDraftUOMCode)
	if err != nil {
		return ProductionPlanLine{}, err
	}

	line := ProductionPlanLine{
		ID:                   strings.TrimSpace(input.ID),
		FormulaLineID:        strings.TrimSpace(input.FormulaLineID),
		LineNo:               input.LineNo,
		ComponentItemID:      strings.TrimSpace(input.ComponentItemID),
		ComponentSKU:         strings.ToUpper(strings.TrimSpace(input.ComponentSKU)),
		ComponentName:        strings.TrimSpace(input.ComponentName),
		ComponentType:        normalizeProductionComponentType(input.ComponentType),
		FormulaQty:           formulaQty,
		FormulaUOMCode:       formulaUOM,
		RequiredQty:          requiredQty,
		RequiredUOMCode:      requiredUOM,
		RequiredStockBaseQty: requiredStockBaseQty,
		StockBaseUOMCode:     stockBaseUOM,
		AvailableQty:         availableQty,
		ShortageQty:          shortageQty,
		PurchaseDraftQty:     purchaseDraftQty,
		PurchaseDraftUOMCode: purchaseDraftUOM,
		IsStockManaged:       input.IsStockManaged,
		NeedsPurchase:        input.NeedsPurchase,
		Note:                 strings.TrimSpace(input.Note),
	}
	if err := line.Validate(); err != nil {
		return ProductionPlanLine{}, err
	}

	return line, nil
}

func (p ProductionPlan) Validate() error {
	if strings.TrimSpace(p.ID) == "" ||
		strings.TrimSpace(p.OrgID) == "" ||
		strings.TrimSpace(p.PlanNo) == "" ||
		strings.TrimSpace(p.OutputItemID) == "" ||
		strings.TrimSpace(p.OutputSKU) == "" ||
		strings.TrimSpace(p.OutputItemName) == "" ||
		strings.TrimSpace(p.FormulaID) == "" ||
		strings.TrimSpace(p.FormulaCode) == "" ||
		strings.TrimSpace(p.FormulaVersion) == "" ||
		strings.TrimSpace(p.CreatedBy) == "" ||
		len(p.Lines) == 0 {
		return ErrProductionPlanRequiredField
	}
	if !IsValidProductionPlanStatus(p.Status) {
		return ErrProductionPlanInvalidStatus
	}
	if !isValidProductionOutputType(p.OutputItemType) {
		return ErrProductionPlanInvalidOutputType
	}
	if p.PlannedQty.IsZero() || p.PlannedQty.IsNegative() || p.FormulaBatchQty.IsZero() || p.FormulaBatchQty.IsNegative() {
		return ErrProductionPlanInvalidQuantity
	}
	if p.UOMCode == "" || p.FormulaBatchUOMCode == "" {
		return ErrProductionPlanInvalidUOM
	}
	for _, line := range p.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
	}
	if len(p.PurchaseDraft.Lines) > 0 {
		if err := p.PurchaseDraft.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (l ProductionPlanLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		strings.TrimSpace(l.FormulaLineID) == "" ||
		l.LineNo <= 0 ||
		strings.TrimSpace(l.ComponentSKU) == "" ||
		strings.TrimSpace(l.ComponentName) == "" {
		return ErrProductionPlanRequiredField
	}
	if !isValidProductionComponentType(l.ComponentType) {
		return ErrProductionPlanInvalidComponentType
	}
	if l.FormulaQty.IsNegative() ||
		l.RequiredQty.IsNegative() ||
		l.RequiredStockBaseQty.IsNegative() ||
		l.AvailableQty.IsNegative() ||
		l.ShortageQty.IsNegative() ||
		l.PurchaseDraftQty.IsNegative() {
		return ErrProductionPlanInvalidQuantity
	}
	if l.FormulaUOMCode == "" || l.RequiredUOMCode == "" || l.StockBaseUOMCode == "" || l.PurchaseDraftUOMCode == "" {
		return ErrProductionPlanInvalidUOM
	}
	if !l.ShortageQty.IsZero() && l.NeedsPurchase && l.PurchaseDraftQty.IsZero() {
		return ErrProductionPlanInvalidShortage
	}
	if l.ShortageQty.IsZero() && (l.NeedsPurchase || !l.PurchaseDraftQty.IsZero()) {
		return ErrProductionPlanInvalidShortage
	}
	if l.NeedsPurchase && l.PurchaseDraftQty != l.ShortageQty {
		return ErrProductionPlanInvalidShortage
	}

	return nil
}

func (d PurchaseRequestDraft) Validate() error {
	if strings.TrimSpace(d.ID) == "" ||
		strings.TrimSpace(d.RequestNo) == "" ||
		strings.TrimSpace(d.SourceProductionPlanID) == "" ||
		strings.TrimSpace(d.SourceProductionPlanNo) == "" ||
		strings.TrimSpace(d.CreatedBy) == "" ||
		len(d.Lines) == 0 {
		return ErrProductionPlanRequiredField
	}
	if !IsValidPurchaseRequestDraftStatus(d.Status) {
		return ErrProductionPlanInvalidStatus
	}
	for _, line := range d.Lines {
		if strings.TrimSpace(line.ID) == "" ||
			line.LineNo <= 0 ||
			strings.TrimSpace(line.SourceProductionPlanLine) == "" ||
			strings.TrimSpace(line.SKU) == "" ||
			strings.TrimSpace(line.ItemName) == "" ||
			line.RequestedQty.IsZero() ||
			line.RequestedQty.IsNegative() ||
			line.UOMCode == "" {
			return ErrProductionPlanRequiredField
		}
	}
	if d.Status == PurchaseRequestDraftStatusConvertedToPO &&
		(strings.TrimSpace(d.ConvertedPurchaseOrderID) == "" || strings.TrimSpace(d.ConvertedPurchaseOrderNo) == "") {
		return ErrProductionPlanRequiredField
	}

	return nil
}

func (d PurchaseRequestDraft) Submit(actorID string, changedAt time.Time) (PurchaseRequestDraft, error) {
	return d.transitionTo(PurchaseRequestDraftStatusSubmitted, actorID, changedAt)
}

func (d PurchaseRequestDraft) Approve(actorID string, changedAt time.Time) (PurchaseRequestDraft, error) {
	return d.transitionTo(PurchaseRequestDraftStatusApproved, actorID, changedAt)
}

func (d PurchaseRequestDraft) MarkConvertedToPO(actorID string, changedAt time.Time, purchaseOrderID string, purchaseOrderNo string) (PurchaseRequestDraft, error) {
	actorID = strings.TrimSpace(actorID)
	purchaseOrderID = strings.TrimSpace(purchaseOrderID)
	purchaseOrderNo = strings.ToUpper(strings.TrimSpace(purchaseOrderNo))
	if actorID == "" || purchaseOrderID == "" || purchaseOrderNo == "" {
		return PurchaseRequestDraft{}, ErrProductionPlanRequiredField
	}
	from := NormalizePurchaseRequestDraftStatus(d.Status)
	to := PurchaseRequestDraftStatusConvertedToPO
	if !CanTransitionPurchaseRequestDraftStatus(from, to) {
		return PurchaseRequestDraft{}, ErrProductionPlanInvalidPurchaseRequestTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	converted := d.Clone()
	converted.Status = to
	converted.ConvertedPurchaseOrderID = purchaseOrderID
	converted.ConvertedPurchaseOrderNo = purchaseOrderNo
	converted.markPurchaseRequestDraftTransition(to, actorID, changedAt.UTC())
	if err := converted.Validate(); err != nil {
		return PurchaseRequestDraft{}, err
	}

	return converted, nil
}

func (d PurchaseRequestDraft) transitionTo(status PurchaseRequestDraftStatus, actorID string, changedAt time.Time) (PurchaseRequestDraft, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PurchaseRequestDraft{}, ErrProductionPlanRequiredField
	}
	from := NormalizePurchaseRequestDraftStatus(d.Status)
	to := NormalizePurchaseRequestDraftStatus(status)
	if !CanTransitionPurchaseRequestDraftStatus(from, to) {
		return PurchaseRequestDraft{}, ErrProductionPlanInvalidPurchaseRequestTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := d.Clone()
	updated.Status = to
	updated.markPurchaseRequestDraftTransition(to, actorID, changedAt.UTC())
	if err := updated.Validate(); err != nil {
		return PurchaseRequestDraft{}, err
	}

	return updated, nil
}

func (d *PurchaseRequestDraft) markPurchaseRequestDraftTransition(status PurchaseRequestDraftStatus, actorID string, changedAt time.Time) {
	switch status {
	case PurchaseRequestDraftStatusSubmitted:
		d.SubmittedAt = changedAt
		d.SubmittedBy = actorID
	case PurchaseRequestDraftStatusApproved:
		d.ApprovedAt = changedAt
		d.ApprovedBy = actorID
	case PurchaseRequestDraftStatusConvertedToPO:
		d.ConvertedAt = changedAt
		d.ConvertedBy = actorID
	case PurchaseRequestDraftStatusCancelled:
		d.CancelledAt = changedAt
		d.CancelledBy = actorID
	case PurchaseRequestDraftStatusRejected:
		d.RejectedAt = changedAt
		d.RejectedBy = actorID
	}
}

func (p ProductionPlan) Clone() ProductionPlan {
	clone := p
	clone.Lines = append([]ProductionPlanLine(nil), p.Lines...)
	clone.PurchaseDraft = p.PurchaseDraft.Clone()

	return clone
}

func (d PurchaseRequestDraft) Clone() PurchaseRequestDraft {
	clone := d
	clone.Lines = append([]PurchaseRequestDraftLine(nil), d.Lines...)

	return clone
}

func SortProductionPlanLines(lines []ProductionPlanLine) {
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].LineNo < lines[j].LineNo
	})
}

func IsValidProductionPlanStatus(status ProductionPlanStatus) bool {
	switch NormalizeProductionPlanStatus(status) {
	case ProductionPlanStatusDraft, ProductionPlanStatusPurchaseRequestDraftCreated, ProductionPlanStatusCancelled:
		return true
	default:
		return false
	}
}

func IsValidPurchaseRequestDraftStatus(status PurchaseRequestDraftStatus) bool {
	switch NormalizePurchaseRequestDraftStatus(status) {
	case PurchaseRequestDraftStatusDraft,
		PurchaseRequestDraftStatusSubmitted,
		PurchaseRequestDraftStatusApproved,
		PurchaseRequestDraftStatusConvertedToPO,
		PurchaseRequestDraftStatusCancelled,
		PurchaseRequestDraftStatusRejected:
		return true
	default:
		return false
	}
}

func NormalizePurchaseRequestDraftStatus(status PurchaseRequestDraftStatus) PurchaseRequestDraftStatus {
	return PurchaseRequestDraftStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func CanTransitionPurchaseRequestDraftStatus(from PurchaseRequestDraftStatus, to PurchaseRequestDraftStatus) bool {
	from = NormalizePurchaseRequestDraftStatus(from)
	to = NormalizePurchaseRequestDraftStatus(to)
	switch from {
	case PurchaseRequestDraftStatusDraft:
		return to == PurchaseRequestDraftStatusSubmitted || to == PurchaseRequestDraftStatusCancelled
	case PurchaseRequestDraftStatusSubmitted:
		return to == PurchaseRequestDraftStatusApproved || to == PurchaseRequestDraftStatusRejected || to == PurchaseRequestDraftStatusCancelled
	case PurchaseRequestDraftStatusApproved:
		return to == PurchaseRequestDraftStatusConvertedToPO || to == PurchaseRequestDraftStatusCancelled
	default:
		return false
	}
}

func NormalizeProductionPlanStatus(status ProductionPlanStatus) ProductionPlanStatus {
	return ProductionPlanStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func newPurchaseRequestDraftForPlan(plan ProductionPlan, draftID string, draftNo string) PurchaseRequestDraft {
	lines := make([]PurchaseRequestDraftLine, 0)
	for _, line := range plan.Lines {
		if !line.NeedsPurchase {
			continue
		}
		lines = append(lines, PurchaseRequestDraftLine{
			ID:                       firstNonBlankProductionPlan("pr-line-"+line.ID, line.ID),
			LineNo:                   len(lines) + 1,
			SourceProductionPlanLine: line.ID,
			ItemID:                   line.ComponentItemID,
			SKU:                      line.ComponentSKU,
			ItemName:                 line.ComponentName,
			RequestedQty:             line.PurchaseDraftQty,
			UOMCode:                  line.PurchaseDraftUOMCode,
			Note:                     "Generated from production material shortage",
		})
	}
	if len(lines) == 0 {
		return PurchaseRequestDraft{}
	}
	id := strings.TrimSpace(draftID)
	if id == "" {
		id = "pr-draft-" + strings.ToLower(plan.ID)
	}
	requestNo := strings.ToUpper(strings.TrimSpace(draftNo))
	if requestNo == "" {
		requestNo = "PR-" + plan.PlanNo
	}

	return PurchaseRequestDraft{
		ID:                     id,
		RequestNo:              requestNo,
		SourceProductionPlanID: plan.ID,
		SourceProductionPlanNo: plan.PlanNo,
		Status:                 PurchaseRequestDraftStatusDraft,
		Lines:                  lines,
		CreatedAt:              plan.CreatedAt,
		CreatedBy:              plan.CreatedBy,
	}
}

func normalizeProductionPlanQuantity(value decimal.Decimal, allowZero bool) (decimal.Decimal, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil {
		return "", ErrProductionPlanInvalidQuantity
	}
	if quantity.IsNegative() || (!allowZero && quantity.IsZero()) {
		return "", ErrProductionPlanInvalidQuantity
	}

	return quantity, nil
}

func normalizeProductionPlanPositiveQuantity(value decimal.Decimal) (decimal.Decimal, error) {
	return normalizeProductionPlanQuantity(value, false)
}

func normalizeProductionPlanUOM(value string) (decimal.UOMCode, error) {
	code, err := decimal.NormalizeUOMCode(value)
	if err != nil {
		return "", ErrProductionPlanInvalidUOM
	}

	return code, nil
}

func normalizeProductionOutputType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isValidProductionOutputType(value string) bool {
	switch normalizeProductionOutputType(value) {
	case "finished_good", "semi_finished":
		return true
	default:
		return false
	}
}

func normalizeProductionComponentType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isValidProductionComponentType(value string) bool {
	switch normalizeProductionComponentType(value) {
	case "raw_material", "fragrance", "packaging", "semi_finished", "service":
		return true
	default:
		return false
	}
}

func firstNonBlankProductionPlan(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}
