package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrFormulaNotFound = errors.New("formula not found")
var ErrDuplicateFormulaVersion = errors.New("formula version already exists")
var ErrFormulaParentItemNotFound = errors.New("formula parent item not found")
var ErrFormulaParentItemInactive = errors.New("formula parent item must be active")

type FormulaCatalog struct {
	mu       sync.RWMutex
	records  map[string]domain.Formula
	auditLog audit.LogStore
	clock    func() time.Time
}

type CreateFormulaInput struct {
	FormulaCode      string
	FinishedItemID   string
	FinishedSKU      string
	FinishedItemName string
	FinishedItemType string
	FormulaVersion   string
	BatchQty         decimal.Decimal
	BatchUOMCode     string
	BaseBatchQty     decimal.Decimal
	BaseBatchUOMCode string
	EffectiveFrom    string
	EffectiveTo      string
	Lines            []CreateFormulaLineInput
	Note             string
	ActorID          string
	RequestID        string
}

type CreateFormulaLineInput struct {
	LineNo           int
	ComponentItemID  string
	ComponentSKU     string
	ComponentName    string
	ComponentType    string
	EnteredQty       decimal.Decimal
	EnteredUOMCode   string
	CalcQty          decimal.Decimal
	CalcUOMCode      string
	StockBaseQty     decimal.Decimal
	StockBaseUOMCode string
	WastePercent     decimal.Decimal
	IsRequired       bool
	IsStockManaged   bool
	LineStatus       string
	Note             string
}

type ActivateFormulaInput struct {
	ID        string
	ActorID   string
	RequestID string
}

type CalculateFormulaRequirementInput struct {
	ID             string
	PlannedQty     decimal.Decimal
	PlannedUOMCode string
}

type FormulaResult struct {
	Formula    domain.Formula
	AuditLogID string
}

type FormulaRequirementResult struct {
	Formula      domain.Formula
	PlannedQty   decimal.Decimal
	PlannedUOM   decimal.UOMCode
	Requirements []domain.FormulaRequirement
}

func createFormulaInputForParent(input CreateFormulaInput, parent domain.Item) (CreateFormulaInput, error) {
	if strings.TrimSpace(parent.ID) == "" {
		return CreateFormulaInput{}, ErrFormulaParentItemNotFound
	}
	if parent.Status != domain.ItemStatusActive {
		return CreateFormulaInput{}, ErrFormulaParentItemInactive
	}
	if parent.Type != domain.ItemTypeFinishedGood && parent.Type != domain.ItemTypeSemiFinished {
		return CreateFormulaInput{}, domain.ErrFormulaInvalidFinishedItemType
	}

	input.FinishedItemID = strings.TrimSpace(parent.ID)
	input.FinishedSKU = domain.NormalizeSKUCode(parent.SKUCode)
	input.FinishedItemName = strings.TrimSpace(parent.Name)
	input.FinishedItemType = string(parent.Type)

	return input, nil
}

func NewPrototypeFormulaCatalog(auditLog audit.LogStore) *FormulaCatalog {
	return &FormulaCatalog{
		records:  make(map[string]domain.Formula),
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func NewPrototypeFormulaCatalogAt(auditLog audit.LogStore, now time.Time) *FormulaCatalog {
	catalog := NewPrototypeFormulaCatalog(auditLog)
	catalog.clock = func() time.Time { return now.UTC() }

	return catalog
}

func (c *FormulaCatalog) List(_ context.Context, filter domain.FormulaFilter) ([]domain.Formula, error) {
	if c == nil {
		return nil, errors.New("formula catalog is required")
	}
	if filter.Status != "" && !domain.IsValidFormulaStatus(filter.Status) {
		return nil, domain.ErrFormulaInvalidStatus
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	formulas := make([]domain.Formula, 0, len(c.records))
	for _, formula := range c.records {
		if filter.Matches(formula) {
			formulas = append(formulas, formula.Clone())
		}
	}
	sortFormulas(formulas)

	return formulas, nil
}

func (c *FormulaCatalog) Get(_ context.Context, id string) (domain.Formula, error) {
	if c == nil {
		return domain.Formula{}, errors.New("formula catalog is required")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	formula, ok := c.records[strings.TrimSpace(id)]
	if !ok {
		return domain.Formula{}, ErrFormulaNotFound
	}

	return formula.Clone(), nil
}

func (c *FormulaCatalog) Create(ctx context.Context, input CreateFormulaInput) (FormulaResult, error) {
	if c == nil {
		return FormulaResult{}, errors.New("formula catalog is required")
	}
	if c.auditLog == nil {
		return FormulaResult{}, errors.New("audit log store is required")
	}

	now := c.clock().UTC()
	formula, err := domain.NewFormula(domain.NewFormulaInput{
		ID:               newFormulaID(input.FormulaCode, input.FormulaVersion, now),
		FormulaCode:      input.FormulaCode,
		FinishedItemID:   input.FinishedItemID,
		FinishedSKU:      input.FinishedSKU,
		FinishedItemName: input.FinishedItemName,
		FinishedItemType: domain.ItemType(input.FinishedItemType),
		FormulaVersion:   input.FormulaVersion,
		BatchQty:         input.BatchQty,
		BatchUOMCode:     input.BatchUOMCode,
		BaseBatchQty:     input.BaseBatchQty,
		BaseBatchUOMCode: input.BaseBatchUOMCode,
		Status:           domain.FormulaStatusDraft,
		ApprovalStatus:   domain.FormulaApprovalDraft,
		EffectiveFrom:    input.EffectiveFrom,
		EffectiveTo:      input.EffectiveTo,
		Lines:            formulaLineInputs(input.FormulaCode, input.FormulaVersion, input.Lines, now),
		Note:             input.Note,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		return FormulaResult{}, err
	}

	c.mu.Lock()
	if err := c.ensureUniqueVersionLocked(formula, ""); err != nil {
		c.mu.Unlock()
		return FormulaResult{}, err
	}
	c.records[formula.ID] = formula.Clone()
	c.mu.Unlock()

	log, err := newFormulaAuditLog(input.ActorID, input.RequestID, "masterdata.formula.created", formula, nil, formulaToAuditMap(formula), now)
	if err != nil {
		return FormulaResult{}, err
	}
	if err := c.auditLog.Record(ctx, log); err != nil {
		return FormulaResult{}, err
	}

	return FormulaResult{Formula: formula, AuditLogID: log.ID}, nil
}

func (c *FormulaCatalog) Activate(ctx context.Context, input ActivateFormulaInput) (FormulaResult, error) {
	if c == nil {
		return FormulaResult{}, errors.New("formula catalog is required")
	}
	if c.auditLog == nil {
		return FormulaResult{}, errors.New("audit log store is required")
	}

	now := c.clock().UTC()
	c.mu.Lock()
	current, ok := c.records[strings.TrimSpace(input.ID)]
	if !ok {
		c.mu.Unlock()
		return FormulaResult{}, ErrFormulaNotFound
	}
	if err := current.ValidateForActivation(); err != nil {
		c.mu.Unlock()
		return FormulaResult{}, err
	}
	before := current.Clone()
	for id, candidate := range c.records {
		if id == current.ID || !strings.EqualFold(candidate.FinishedItemID, current.FinishedItemID) || candidate.Status != domain.FormulaStatusActive {
			continue
		}
		candidate.Status = domain.FormulaStatusInactive
		candidate.UpdatedAt = now
		candidate.Version += 1
		c.records[id] = candidate.Clone()
	}
	current.Status = domain.FormulaStatusActive
	current.ApprovalStatus = domain.FormulaApprovalApproved
	current.UpdatedAt = now
	current.ApprovedBy = strings.TrimSpace(input.ActorID)
	current.ApprovedAt = now
	current.Version += 1
	c.records[current.ID] = current.Clone()
	c.mu.Unlock()

	log, err := newFormulaAuditLog(input.ActorID, input.RequestID, "masterdata.formula.activated", current, formulaToAuditMap(before), formulaToAuditMap(current), now)
	if err != nil {
		return FormulaResult{}, err
	}
	if err := c.auditLog.Record(ctx, log); err != nil {
		return FormulaResult{}, err
	}

	return FormulaResult{Formula: current, AuditLogID: log.ID}, nil
}

func (c *FormulaCatalog) CalculateRequirement(ctx context.Context, input CalculateFormulaRequirementInput) (FormulaRequirementResult, error) {
	formula, err := c.Get(ctx, input.ID)
	if err != nil {
		return FormulaRequirementResult{}, err
	}
	plannedQty, err := decimal.ParseQuantity(input.PlannedQty.String())
	if err != nil {
		return FormulaRequirementResult{}, domain.ErrFormulaInvalidQuantity
	}
	plannedUOM, err := decimal.NormalizeUOMCode(input.PlannedUOMCode)
	if err != nil {
		return FormulaRequirementResult{}, domain.ErrFormulaInvalidUOM
	}
	requirements, err := formula.CalculateRequirement(plannedQty, plannedUOM.String())
	if err != nil {
		return FormulaRequirementResult{}, err
	}

	return FormulaRequirementResult{
		Formula:      formula,
		PlannedQty:   plannedQty,
		PlannedUOM:   plannedUOM,
		Requirements: requirements,
	}, nil
}

func (c *FormulaCatalog) ensureUniqueVersionLocked(formula domain.Formula, currentID string) error {
	for _, existing := range c.records {
		if currentID != "" && existing.ID == currentID {
			continue
		}
		if strings.EqualFold(existing.FinishedItemID, formula.FinishedItemID) && strings.EqualFold(existing.FormulaVersion, formula.FormulaVersion) {
			return ErrDuplicateFormulaVersion
		}
	}

	return nil
}

func formulaLineInputs(formulaCode string, formulaVersion string, inputs []CreateFormulaLineInput, now time.Time) []domain.NewFormulaLineInput {
	lines := make([]domain.NewFormulaLineInput, 0, len(inputs))
	for _, input := range inputs {
		lineStatus := domain.NormalizeFormulaLineStatus(domain.FormulaLineStatus(input.LineStatus))
		if lineStatus == "" {
			lineStatus = domain.FormulaLineStatusActive
		}
		lines = append(lines, domain.NewFormulaLineInput{
			ID:               newFormulaLineID(formulaCode, formulaVersion, input.LineNo, now),
			LineNo:           input.LineNo,
			ComponentItemID:  input.ComponentItemID,
			ComponentSKU:     input.ComponentSKU,
			ComponentName:    input.ComponentName,
			ComponentType:    domain.FormulaComponentType(input.ComponentType),
			EnteredQty:       input.EnteredQty,
			EnteredUOMCode:   input.EnteredUOMCode,
			CalcQty:          input.CalcQty,
			CalcUOMCode:      input.CalcUOMCode,
			StockBaseQty:     input.StockBaseQty,
			StockBaseUOMCode: input.StockBaseUOMCode,
			WastePercent:     input.WastePercent,
			IsRequired:       input.IsRequired,
			IsStockManaged:   input.IsStockManaged,
			LineStatus:       lineStatus,
			Note:             input.Note,
		})
	}

	return lines
}

func newFormulaID(formulaCode string, formulaVersion string, now time.Time) string {
	return fmt.Sprintf("formula-%s-%s-%d", slugFormulaPart(formulaCode), slugFormulaPart(formulaVersion), now.UnixNano())
}

func newFormulaLineID(formulaCode string, formulaVersion string, lineNo int, now time.Time) string {
	return fmt.Sprintf("formula-line-%s-%s-%03d-%d", slugFormulaPart(formulaCode), slugFormulaPart(formulaVersion), lineNo, now.UnixNano())
}

func slugFormulaPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "/", "-")
	if value == "" {
		return "draft"
	}

	return value
}

func sortFormulas(formulas []domain.Formula) {
	sort.SliceStable(formulas, func(i, j int) bool {
		leftStatus := formulaStatusRank(formulas[i].Status)
		rightStatus := formulaStatusRank(formulas[j].Status)
		if leftStatus != rightStatus {
			return leftStatus < rightStatus
		}
		if !strings.EqualFold(formulas[i].FinishedSKU, formulas[j].FinishedSKU) {
			return formulas[i].FinishedSKU < formulas[j].FinishedSKU
		}

		return formulas[i].FormulaVersion < formulas[j].FormulaVersion
	})
}

func formulaStatusRank(status domain.FormulaStatus) int {
	switch status {
	case domain.FormulaStatusActive:
		return 0
	case domain.FormulaStatusDraft:
		return 1
	case domain.FormulaStatusInactive:
		return 2
	case domain.FormulaStatusArchived:
		return 3
	default:
		return 4
	}
}

func newFormulaAuditLog(actorID string, requestID string, action string, formula domain.Formula, before map[string]any, after map[string]any, occurredAt time.Time) (audit.Log, error) {
	if strings.TrimSpace(actorID) == "" {
		actorID = "system"
	}
	if strings.TrimSpace(requestID) == "" {
		requestID = fmt.Sprintf("formula-%d", occurredAt.UnixNano())
	}

	return audit.NewLog(audit.NewLogInput{
		ID:         fmt.Sprintf("audit-formula-%d", occurredAt.UnixNano()),
		ActorID:    actorID,
		Action:     action,
		EntityType: "masterdata.formula",
		EntityID:   formula.ID,
		RequestID:  requestID,
		BeforeData: before,
		AfterData:  after,
		CreatedAt:  occurredAt,
	})
}

func formulaToAuditMap(formula domain.Formula) map[string]any {
	return map[string]any{
		"formula_code":       formula.FormulaCode,
		"finished_item_id":   formula.FinishedItemID,
		"finished_sku":       formula.FinishedSKU,
		"formula_version":    formula.FormulaVersion,
		"batch_qty":          formula.BatchQty.String(),
		"batch_uom_code":     formula.BatchUOMCode.String(),
		"status":             string(formula.Status),
		"approval_status":    string(formula.ApprovalStatus),
		"formula_line_count": len(formula.Lines),
	}
}
