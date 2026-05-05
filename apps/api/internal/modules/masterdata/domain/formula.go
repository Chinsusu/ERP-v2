package domain

import (
	"errors"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrFormulaRequiredField = errors.New("formula required field is missing")
var ErrFormulaInvalidStatus = errors.New("formula status is invalid")
var ErrFormulaInvalidApprovalStatus = errors.New("formula approval status is invalid")
var ErrFormulaInvalidFinishedItemType = errors.New("formula finished item type is invalid")
var ErrFormulaInvalidQuantity = errors.New("formula quantity is invalid")
var ErrFormulaInvalidLineQuantity = errors.New("formula line quantity is invalid")
var ErrFormulaInvalidComponentType = errors.New("formula component type is invalid")
var ErrFormulaInvalidLineStatus = errors.New("formula line status is invalid")
var ErrFormulaInvalidUOM = errors.New("formula uom is invalid")

type FormulaStatus string

const FormulaStatusDraft FormulaStatus = "draft"
const FormulaStatusActive FormulaStatus = "active"
const FormulaStatusInactive FormulaStatus = "inactive"
const FormulaStatusArchived FormulaStatus = "archived"

type FormulaApprovalStatus string

const FormulaApprovalDraft FormulaApprovalStatus = "draft"
const FormulaApprovalPending FormulaApprovalStatus = "pending_approval"
const FormulaApprovalApproved FormulaApprovalStatus = "approved"
const FormulaApprovalRejected FormulaApprovalStatus = "rejected"

type FormulaComponentType string

const FormulaComponentRawMaterial FormulaComponentType = "raw_material"
const FormulaComponentFragrance FormulaComponentType = "fragrance"
const FormulaComponentPackaging FormulaComponentType = "packaging"
const FormulaComponentSemiFinished FormulaComponentType = "semi_finished"
const FormulaComponentService FormulaComponentType = "service"

type FormulaLineStatus string

const FormulaLineStatusActive FormulaLineStatus = "active"
const FormulaLineStatusExcluded FormulaLineStatus = "excluded"
const FormulaLineStatusNeedsReview FormulaLineStatus = "needs_review"

type Formula struct {
	ID               string
	FormulaCode      string
	FinishedItemID   string
	FinishedSKU      string
	FinishedItemName string
	FinishedItemType ItemType
	FormulaVersion   string
	BatchQty         decimal.Decimal
	BatchUOMCode     decimal.UOMCode
	BaseBatchQty     decimal.Decimal
	BaseBatchUOMCode decimal.UOMCode
	Status           FormulaStatus
	ApprovalStatus   FormulaApprovalStatus
	EffectiveFrom    string
	EffectiveTo      string
	Lines            []FormulaLine
	Note             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ApprovedBy       string
	ApprovedAt       time.Time
	Version          int
}

type FormulaLine struct {
	ID               string
	LineNo           int
	ComponentItemID  string
	ComponentSKU     string
	ComponentName    string
	ComponentType    FormulaComponentType
	EnteredQty       decimal.Decimal
	EnteredUOMCode   decimal.UOMCode
	CalcQty          decimal.Decimal
	CalcUOMCode      decimal.UOMCode
	StockBaseQty     decimal.Decimal
	StockBaseUOMCode decimal.UOMCode
	WastePercent     decimal.Decimal
	IsRequired       bool
	IsStockManaged   bool
	LineStatus       FormulaLineStatus
	Note             string
}

type NewFormulaInput struct {
	ID               string
	FormulaCode      string
	FinishedItemID   string
	FinishedSKU      string
	FinishedItemName string
	FinishedItemType ItemType
	FormulaVersion   string
	BatchQty         decimal.Decimal
	BatchUOMCode     string
	BaseBatchQty     decimal.Decimal
	BaseBatchUOMCode string
	Status           FormulaStatus
	ApprovalStatus   FormulaApprovalStatus
	EffectiveFrom    string
	EffectiveTo      string
	Lines            []NewFormulaLineInput
	Note             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ApprovedBy       string
	ApprovedAt       time.Time
	Version          int
}

type NewFormulaLineInput struct {
	ID               string
	LineNo           int
	ComponentItemID  string
	ComponentSKU     string
	ComponentName    string
	ComponentType    FormulaComponentType
	EnteredQty       decimal.Decimal
	EnteredUOMCode   string
	CalcQty          decimal.Decimal
	CalcUOMCode      string
	StockBaseQty     decimal.Decimal
	StockBaseUOMCode string
	WastePercent     decimal.Decimal
	IsRequired       bool
	IsStockManaged   bool
	LineStatus       FormulaLineStatus
	Note             string
}

type FormulaRequirement struct {
	FormulaLineID        string
	LineNo               int
	ComponentItemID      string
	ComponentSKU         string
	ComponentName        string
	ComponentType        FormulaComponentType
	RequiredCalcQty      decimal.Decimal
	CalcUOMCode          decimal.UOMCode
	RequiredStockBaseQty decimal.Decimal
	StockBaseUOMCode     decimal.UOMCode
	IsStockManaged       bool
}

type FormulaFilter struct {
	FinishedItemID string
	Status         FormulaStatus
	Search         string
}

func (f FormulaFilter) Matches(formula Formula) bool {
	if strings.TrimSpace(f.FinishedItemID) != "" && !strings.EqualFold(formula.FinishedItemID, strings.TrimSpace(f.FinishedItemID)) {
		return false
	}
	if f.Status != "" && formula.Status != NormalizeFormulaStatus(f.Status) {
		return false
	}
	search := strings.ToLower(strings.TrimSpace(f.Search))
	if search == "" {
		return true
	}

	return strings.Contains(strings.ToLower(formula.FormulaCode), search) ||
		strings.Contains(strings.ToLower(formula.FinishedSKU), search) ||
		strings.Contains(strings.ToLower(formula.FinishedItemName), search) ||
		strings.Contains(strings.ToLower(formula.FormulaVersion), search)
}

func NewFormula(input NewFormulaInput) (Formula, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizeFormulaStatus(input.Status)
	if status == "" {
		status = FormulaStatusDraft
	}
	approvalStatus := NormalizeFormulaApprovalStatus(input.ApprovalStatus)
	if approvalStatus == "" {
		approvalStatus = FormulaApprovalDraft
	}
	batchQty, err := parseFormulaQuantity(input.BatchQty, false)
	if err != nil {
		return Formula{}, ErrFormulaInvalidQuantity
	}
	baseBatchQty, err := parseFormulaQuantity(input.BaseBatchQty, false)
	if err != nil {
		return Formula{}, ErrFormulaInvalidQuantity
	}
	batchUOM, err := normalizeFormulaUOM(input.BatchUOMCode)
	if err != nil {
		return Formula{}, err
	}
	baseBatchUOM, err := normalizeFormulaUOM(input.BaseBatchUOMCode)
	if err != nil {
		return Formula{}, err
	}

	formula := Formula{
		ID:               strings.TrimSpace(input.ID),
		FormulaCode:      NormalizeItemCode(input.FormulaCode),
		FinishedItemID:   strings.TrimSpace(input.FinishedItemID),
		FinishedSKU:      NormalizeSKUCode(input.FinishedSKU),
		FinishedItemName: strings.TrimSpace(input.FinishedItemName),
		FinishedItemType: NormalizeItemType(input.FinishedItemType),
		FormulaVersion:   strings.TrimSpace(input.FormulaVersion),
		BatchQty:         batchQty,
		BatchUOMCode:     batchUOM,
		BaseBatchQty:     baseBatchQty,
		BaseBatchUOMCode: baseBatchUOM,
		Status:           status,
		ApprovalStatus:   approvalStatus,
		EffectiveFrom:    strings.TrimSpace(input.EffectiveFrom),
		EffectiveTo:      strings.TrimSpace(input.EffectiveTo),
		Lines:            make([]FormulaLine, 0, len(input.Lines)),
		Note:             strings.TrimSpace(input.Note),
		CreatedAt:        createdAt.UTC(),
		UpdatedAt:        updatedAt.UTC(),
		ApprovedBy:       strings.TrimSpace(input.ApprovedBy),
		ApprovedAt:       input.ApprovedAt,
		Version:          input.Version,
	}
	if formula.Version <= 0 {
		formula.Version = 1
	}

	for _, lineInput := range input.Lines {
		line, err := NewFormulaLine(lineInput)
		if err != nil {
			return Formula{}, err
		}
		formula.Lines = append(formula.Lines, line)
	}
	SortFormulaLines(formula.Lines)

	if err := formula.Validate(); err != nil {
		return Formula{}, err
	}

	return formula, nil
}

func NewFormulaLine(input NewFormulaLineInput) (FormulaLine, error) {
	enteredQty, err := parseFormulaQuantity(input.EnteredQty, true)
	if err != nil {
		return FormulaLine{}, ErrFormulaInvalidLineQuantity
	}
	calcQty, err := parseFormulaQuantity(input.CalcQty, true)
	if err != nil {
		return FormulaLine{}, ErrFormulaInvalidLineQuantity
	}
	stockBaseQty, err := parseFormulaQuantity(input.StockBaseQty, true)
	if err != nil {
		return FormulaLine{}, ErrFormulaInvalidLineQuantity
	}
	wastePercent, err := decimal.ParseRate(input.WastePercent.String())
	if err != nil {
		return FormulaLine{}, ErrFormulaInvalidLineQuantity
	}
	if wastePercent.IsNegative() {
		return FormulaLine{}, ErrFormulaInvalidLineQuantity
	}
	enteredUOM, err := normalizeFormulaUOM(input.EnteredUOMCode)
	if err != nil {
		return FormulaLine{}, err
	}
	calcUOM, err := normalizeFormulaUOM(input.CalcUOMCode)
	if err != nil {
		return FormulaLine{}, err
	}
	stockBaseUOM, err := normalizeFormulaUOM(input.StockBaseUOMCode)
	if err != nil {
		return FormulaLine{}, err
	}
	componentType := NormalizeFormulaComponentType(input.ComponentType)
	if !IsValidFormulaComponentType(componentType) {
		return FormulaLine{}, ErrFormulaInvalidComponentType
	}
	lineStatus := NormalizeFormulaLineStatus(input.LineStatus)
	if lineStatus == "" {
		lineStatus = FormulaLineStatusActive
	}
	if !IsValidFormulaLineStatus(lineStatus) {
		return FormulaLine{}, ErrFormulaInvalidLineStatus
	}

	line := FormulaLine{
		ID:               strings.TrimSpace(input.ID),
		LineNo:           input.LineNo,
		ComponentItemID:  strings.TrimSpace(input.ComponentItemID),
		ComponentSKU:     NormalizeSKUCode(input.ComponentSKU),
		ComponentName:    strings.TrimSpace(input.ComponentName),
		ComponentType:    componentType,
		EnteredQty:       enteredQty,
		EnteredUOMCode:   enteredUOM,
		CalcQty:          calcQty,
		CalcUOMCode:      calcUOM,
		StockBaseQty:     stockBaseQty,
		StockBaseUOMCode: stockBaseUOM,
		WastePercent:     wastePercent,
		IsRequired:       input.IsRequired,
		IsStockManaged:   input.IsStockManaged,
		LineStatus:       lineStatus,
		Note:             strings.TrimSpace(input.Note),
	}
	if line.LineNo <= 0 {
		return FormulaLine{}, ErrFormulaRequiredField
	}

	return line, nil
}

func (f Formula) Validate() error {
	if strings.TrimSpace(f.ID) == "" ||
		strings.TrimSpace(f.FormulaCode) == "" ||
		strings.TrimSpace(f.FinishedItemID) == "" ||
		strings.TrimSpace(f.FinishedSKU) == "" ||
		strings.TrimSpace(f.FinishedItemName) == "" ||
		strings.TrimSpace(f.FormulaVersion) == "" ||
		len(f.Lines) == 0 {
		return ErrFormulaRequiredField
	}
	if f.FinishedItemType != ItemTypeFinishedGood && f.FinishedItemType != ItemTypeSemiFinished {
		return ErrFormulaInvalidFinishedItemType
	}
	if !IsValidFormulaStatus(f.Status) {
		return ErrFormulaInvalidStatus
	}
	if !IsValidFormulaApprovalStatus(f.ApprovalStatus) {
		return ErrFormulaInvalidApprovalStatus
	}
	if f.BatchQty.IsZero() || f.BatchQty.IsNegative() || f.BaseBatchQty.IsZero() || f.BaseBatchQty.IsNegative() {
		return ErrFormulaInvalidQuantity
	}
	for _, line := range f.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (f Formula) ValidateForActivation() error {
	if err := f.Validate(); err != nil {
		return err
	}
	for _, line := range f.Lines {
		if line.LineStatus == FormulaLineStatusExcluded {
			continue
		}
		if line.IsRequired && (line.EnteredQty.IsZero() || line.CalcQty.IsZero() || (line.IsStockManaged && line.StockBaseQty.IsZero())) {
			return ErrFormulaInvalidLineQuantity
		}
		if line.IsRequired && (strings.TrimSpace(line.ComponentItemID) == "" || strings.TrimSpace(line.ComponentSKU) == "" || strings.TrimSpace(line.ComponentName) == "") {
			return ErrFormulaRequiredField
		}
	}

	return nil
}

func (l FormulaLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" || l.LineNo <= 0 {
		return ErrFormulaRequiredField
	}
	if !IsValidFormulaComponentType(l.ComponentType) {
		return ErrFormulaInvalidComponentType
	}
	if !IsValidFormulaLineStatus(l.LineStatus) {
		return ErrFormulaInvalidLineStatus
	}
	if l.EnteredQty.IsNegative() || l.CalcQty.IsNegative() || l.StockBaseQty.IsNegative() || l.WastePercent.IsNegative() {
		return ErrFormulaInvalidLineQuantity
	}
	if l.LineStatus == FormulaLineStatusActive && strings.TrimSpace(l.ComponentSKU) == "" {
		return ErrFormulaRequiredField
	}

	return nil
}

func (f Formula) CalculateRequirement(plannedQty decimal.Decimal, plannedUOMCode string) ([]FormulaRequirement, error) {
	planned, err := parseFormulaQuantity(plannedQty, false)
	if err != nil {
		return nil, ErrFormulaInvalidQuantity
	}
	plannedUOM, err := normalizeFormulaUOM(plannedUOMCode)
	if err != nil {
		return nil, err
	}
	if plannedUOM != f.BatchUOMCode {
		return nil, ErrFormulaInvalidUOM
	}

	perUnitFormulaBasisQty := decimal.MustQuantity("1")
	requirements := make([]FormulaRequirement, 0, len(f.Lines))
	for _, line := range f.Lines {
		if line.LineStatus == FormulaLineStatusExcluded {
			continue
		}
		requiredCalcQty, err := scaleFormulaQuantity(line.CalcQty, planned, perUnitFormulaBasisQty)
		if err != nil {
			return nil, err
		}
		requiredStockBaseQty, err := scaleFormulaQuantity(line.StockBaseQty, planned, perUnitFormulaBasisQty)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, FormulaRequirement{
			FormulaLineID:        line.ID,
			LineNo:               line.LineNo,
			ComponentItemID:      line.ComponentItemID,
			ComponentSKU:         line.ComponentSKU,
			ComponentName:        line.ComponentName,
			ComponentType:        line.ComponentType,
			RequiredCalcQty:      requiredCalcQty,
			CalcUOMCode:          line.CalcUOMCode,
			RequiredStockBaseQty: requiredStockBaseQty,
			StockBaseUOMCode:     line.StockBaseUOMCode,
			IsStockManaged:       line.IsStockManaged,
		})
	}

	return requirements, nil
}

func (f Formula) Clone() Formula {
	clone := f
	clone.Lines = append([]FormulaLine(nil), f.Lines...)

	return clone
}

func SortFormulaLines(lines []FormulaLine) {
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].LineNo < lines[j].LineNo
	})
}

func NormalizeFormulaStatus(status FormulaStatus) FormulaStatus {
	return FormulaStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidFormulaStatus(status FormulaStatus) bool {
	switch NormalizeFormulaStatus(status) {
	case FormulaStatusDraft, FormulaStatusActive, FormulaStatusInactive, FormulaStatusArchived:
		return true
	default:
		return false
	}
}

func NormalizeFormulaApprovalStatus(status FormulaApprovalStatus) FormulaApprovalStatus {
	return FormulaApprovalStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidFormulaApprovalStatus(status FormulaApprovalStatus) bool {
	switch NormalizeFormulaApprovalStatus(status) {
	case FormulaApprovalDraft, FormulaApprovalPending, FormulaApprovalApproved, FormulaApprovalRejected:
		return true
	default:
		return false
	}
}

func NormalizeFormulaComponentType(componentType FormulaComponentType) FormulaComponentType {
	return FormulaComponentType(strings.ToLower(strings.TrimSpace(string(componentType))))
}

func IsValidFormulaComponentType(componentType FormulaComponentType) bool {
	switch NormalizeFormulaComponentType(componentType) {
	case FormulaComponentRawMaterial, FormulaComponentFragrance, FormulaComponentPackaging, FormulaComponentSemiFinished, FormulaComponentService:
		return true
	default:
		return false
	}
}

func NormalizeFormulaLineStatus(status FormulaLineStatus) FormulaLineStatus {
	return FormulaLineStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidFormulaLineStatus(status FormulaLineStatus) bool {
	switch NormalizeFormulaLineStatus(status) {
	case FormulaLineStatusActive, FormulaLineStatusExcluded, FormulaLineStatusNeedsReview:
		return true
	default:
		return false
	}
}

func parseFormulaQuantity(value decimal.Decimal, allowZero bool) (decimal.Decimal, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil {
		return "", err
	}
	if quantity.IsNegative() || (!allowZero && quantity.IsZero()) {
		return "", ErrFormulaInvalidQuantity
	}

	return quantity, nil
}

func normalizeFormulaUOM(value string) (decimal.UOMCode, error) {
	code, err := decimal.NormalizeUOMCode(value)
	if err != nil {
		return "", ErrFormulaInvalidUOM
	}

	return code, nil
}

func scaleFormulaQuantity(quantity decimal.Decimal, planned decimal.Decimal, basis decimal.Decimal) (decimal.Decimal, error) {
	if basis.IsZero() || basis.IsNegative() || planned.IsZero() || planned.IsNegative() || quantity.IsNegative() {
		return "", ErrFormulaInvalidQuantity
	}
	quantityScaled, err := formulaQuantityScaledInt(quantity)
	if err != nil {
		return "", err
	}
	plannedScaled, err := formulaQuantityScaledInt(planned)
	if err != nil {
		return "", err
	}
	basisScaled, err := formulaQuantityScaledInt(basis)
	if err != nil {
		return "", err
	}

	numerator := new(big.Int).Mul(quantityScaled, plannedScaled)
	quotient, remainder := new(big.Int).QuoRem(numerator, basisScaled, new(big.Int))
	threshold := new(big.Int).Mul(new(big.Int).Abs(remainder), big.NewInt(2))
	if threshold.Cmp(basisScaled) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}

	return formulaScaledIntToQuantity(quotient)
}

func formulaQuantityScaledInt(value decimal.Decimal) (*big.Int, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil {
		return nil, err
	}
	digits := strings.ReplaceAll(quantity.String(), ".", "")
	scaled, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, ErrFormulaInvalidQuantity
	}

	return scaled, nil
}

func formulaScaledIntToQuantity(value *big.Int) (decimal.Decimal, error) {
	digits := value.String()
	if len(digits) <= decimal.QuantityScale {
		digits = strings.Repeat("0", decimal.QuantityScale-len(digits)+1) + digits
	}
	raw := digits[:len(digits)-decimal.QuantityScale] + "." + digits[len(digits)-decimal.QuantityScale:]

	return decimal.ParseQuantity(raw)
}
