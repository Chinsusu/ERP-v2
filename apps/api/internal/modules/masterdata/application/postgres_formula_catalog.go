package application

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresFormulaCatalogConfig struct {
	DefaultOrgID string
	Clock        func() time.Time
}

type PostgresFormulaCatalog struct {
	db           *sql.DB
	auditLog     audit.LogStore
	defaultOrgID string
	clock        func() time.Time
}

func NewPostgresFormulaCatalog(db *sql.DB, auditLog audit.LogStore, cfg PostgresFormulaCatalogConfig) *PostgresFormulaCatalog {
	clock := cfg.Clock
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}

	return &PostgresFormulaCatalog{
		db:           db,
		auditLog:     auditLog,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
		clock:        clock,
	}
}

const selectPostgresFormulasSQL = `
SELECT
  formula.formula_ref,
  formula.formula_code,
  formula.finished_item_id::text,
  formula.finished_sku,
  formula.finished_item_name,
  formula.finished_item_type,
  formula.formula_version,
  formula.batch_qty::text,
  formula.batch_uom_code,
  formula.base_batch_qty::text,
  formula.base_batch_uom_code,
  formula.status,
  formula.approval_status,
  COALESCE(formula.effective_from::text, ''),
  COALESCE(formula.effective_to::text, ''),
  COALESCE(formula.note, ''),
  COALESCE(formula.approved_by_ref, ''),
  formula.approved_at,
  formula.created_at,
  formula.updated_at,
  formula.version
FROM mdm.item_formulas AS formula
WHERE formula.org_id = $1::uuid
ORDER BY formula.status, formula.finished_sku, formula.formula_version`

const selectPostgresFormulaSQL = `
SELECT
  formula.formula_ref,
  formula.formula_code,
  formula.finished_item_id::text,
  formula.finished_sku,
  formula.finished_item_name,
  formula.finished_item_type,
  formula.formula_version,
  formula.batch_qty::text,
  formula.batch_uom_code,
  formula.base_batch_qty::text,
  formula.base_batch_uom_code,
  formula.status,
  formula.approval_status,
  COALESCE(formula.effective_from::text, ''),
  COALESCE(formula.effective_to::text, ''),
  COALESCE(formula.note, ''),
  COALESCE(formula.approved_by_ref, ''),
  formula.approved_at,
  formula.created_at,
  formula.updated_at,
  formula.version
FROM mdm.item_formulas AS formula
WHERE formula.org_id = $1::uuid
  AND lower(formula.formula_ref) = lower($2)
LIMIT 1`

const selectPostgresFormulaLinesSQL = `
SELECT
  line.line_ref,
  line.line_no,
  COALESCE(line.component_item_ref, ''),
  COALESCE(line.component_sku, ''),
  COALESCE(line.component_name, ''),
  line.component_type,
  line.entered_qty::text,
  line.entered_uom_code,
  line.calc_qty::text,
  line.calc_uom_code,
  line.stock_base_qty::text,
  line.stock_base_uom_code,
  line.waste_percent::text,
  line.is_required,
  line.is_stock_managed,
  line.line_status,
  COALESCE(line.note, '')
FROM mdm.item_formula_lines AS line
JOIN mdm.item_formulas AS formula ON formula.id = line.formula_id
WHERE formula.org_id = $1::uuid
  AND lower(formula.formula_ref) = lower($2)
ORDER BY line.line_no`

const insertPostgresFormulaSQL = `
INSERT INTO mdm.item_formulas (
  org_id,
  formula_ref,
  formula_code,
  finished_item_ref,
  finished_item_id,
  finished_sku,
  finished_item_name,
  finished_item_type,
  formula_version,
  batch_qty,
  batch_uom_code,
  base_batch_qty,
  base_batch_uom_code,
  status,
  approval_status,
  effective_from,
  effective_to,
  note,
  approved_by_ref,
  approved_at,
  created_at,
  updated_at,
  version
) VALUES (
  $1::uuid,
  $2,
  $3,
  $4,
  $5::uuid,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16::date,
  $17::date,
  $18,
  $19,
  $20,
  $21,
  $22,
  $23
)
RETURNING id::text`

const insertPostgresFormulaLineSQL = `
INSERT INTO mdm.item_formula_lines (
  org_id,
  formula_id,
  line_ref,
  line_no,
  component_item_ref,
  component_sku,
  component_name,
  component_type,
  entered_qty,
  entered_uom_code,
  calc_qty,
  calc_uom_code,
  stock_base_qty,
  stock_base_uom_code,
  waste_percent,
  is_required,
  is_stock_managed,
  line_status,
  note
) VALUES (
  $1::uuid,
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19
)`

const selectPostgresFormulaDuplicateSQL = `
SELECT formula_ref
FROM mdm.item_formulas
WHERE org_id = $1::uuid
  AND finished_item_id = $2::uuid
  AND lower(formula_version) = lower($3)
LIMIT 1`

const selectPostgresFormulaParentItemSQL = `
SELECT
  item.id::text,
  COALESCE(item.item_code, item.sku),
  item.sku,
  item.name,
  item.item_type,
  CASE WHEN item.status = 'blocked' THEN 'inactive' ELSE item.status END,
  COALESCE(item.uom_base, unit.code, 'PCS')
FROM mdm.items AS item
LEFT JOIN mdm.units AS unit ON unit.id = item.base_unit_id
WHERE item.org_id = $1::uuid
  AND (
    item.id::text = $2
    OR lower(COALESCE(item.item_ref, item.id::text)) = lower($2)
    OR lower(item.sku) = lower($2)
    OR lower(COALESCE(item.item_code, '')) = lower($2)
  )
LIMIT 1`

const deactivatePostgresActiveFormulaSQL = `
UPDATE mdm.item_formulas
SET status = 'inactive',
    updated_at = $3,
    version = version + 1
WHERE org_id = $1::uuid
  AND finished_item_id = $2::uuid
  AND status = 'active'`

const activatePostgresFormulaSQL = `
UPDATE mdm.item_formulas
SET status = 'active',
    approval_status = 'approved',
    approved_by_ref = $3,
    approved_at = $4,
    updated_at = $4,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(formula_ref) = lower($2)`

func (s *PostgresFormulaCatalog) List(ctx context.Context, filter domain.FormulaFilter) ([]domain.Formula, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("database connection is required")
	}
	if filter.Status != "" && !domain.IsValidFormulaStatus(filter.Status) {
		return nil, domain.ErrFormulaInvalidStatus
	}
	orgID, err := s.resolveOrgID()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, selectPostgresFormulasSQL, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	formulas := make([]domain.Formula, 0)
	for rows.Next() {
		formula, err := s.scanFormulaWithLines(ctx, orgID, rows)
		if err != nil {
			return nil, err
		}
		if filter.Matches(formula) {
			formulas = append(formulas, formula)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortFormulas(formulas)

	return formulas, nil
}

func (s *PostgresFormulaCatalog) Get(ctx context.Context, id string) (domain.Formula, error) {
	if s == nil || s.db == nil {
		return domain.Formula{}, errors.New("database connection is required")
	}
	orgID, err := s.resolveOrgID()
	if err != nil {
		return domain.Formula{}, err
	}
	formula, err := s.scanFormulaWithLines(ctx, orgID, s.db.QueryRowContext(ctx, selectPostgresFormulaSQL, orgID, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Formula{}, ErrFormulaNotFound
	}
	if err != nil {
		return domain.Formula{}, err
	}

	return formula, nil
}

func (s *PostgresFormulaCatalog) Create(ctx context.Context, input CreateFormulaInput) (FormulaResult, error) {
	if s == nil || s.db == nil {
		return FormulaResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return FormulaResult{}, errors.New("audit log store is required")
	}
	orgID, err := s.resolveOrgID()
	if err != nil {
		return FormulaResult{}, err
	}
	parent, err := s.getFormulaParentItem(ctx, orgID, input.FinishedItemID)
	if err != nil {
		return FormulaResult{}, err
	}
	input, err = createFormulaInputForParent(input, parent)
	if err != nil {
		return FormulaResult{}, err
	}
	now := s.clock().UTC()
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
	if err := s.saveNewFormula(ctx, orgID, formula); err != nil {
		return FormulaResult{}, err
	}

	log, err := newFormulaAuditLog(input.ActorID, input.RequestID, "masterdata.formula.created", formula, nil, formulaToAuditMap(formula), now)
	if err != nil {
		return FormulaResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return FormulaResult{}, err
	}

	return FormulaResult{Formula: formula, AuditLogID: log.ID}, nil
}

func (s *PostgresFormulaCatalog) Activate(ctx context.Context, input ActivateFormulaInput) (FormulaResult, error) {
	if s == nil || s.db == nil {
		return FormulaResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return FormulaResult{}, errors.New("audit log store is required")
	}
	current, err := s.Get(ctx, input.ID)
	if err != nil {
		return FormulaResult{}, err
	}
	if err := current.ValidateForActivation(); err != nil {
		return FormulaResult{}, err
	}
	before := current.Clone()
	now := s.clock().UTC()
	orgID, err := s.resolveOrgID()
	if err != nil {
		return FormulaResult{}, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return FormulaResult{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if _, err := tx.ExecContext(ctx, deactivatePostgresActiveFormulaSQL, orgID, current.FinishedItemID, now); err != nil {
		return FormulaResult{}, err
	}
	result, err := tx.ExecContext(ctx, activatePostgresFormulaSQL, orgID, current.ID, strings.TrimSpace(input.ActorID), now)
	if err != nil {
		return FormulaResult{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return FormulaResult{}, err
	}
	if affected == 0 {
		return FormulaResult{}, ErrFormulaNotFound
	}
	if err := tx.Commit(); err != nil {
		return FormulaResult{}, err
	}
	committed = true

	current.Status = domain.FormulaStatusActive
	current.ApprovalStatus = domain.FormulaApprovalApproved
	current.ApprovedBy = strings.TrimSpace(input.ActorID)
	current.ApprovedAt = now
	current.UpdatedAt = now
	current.Version += 1
	log, err := newFormulaAuditLog(input.ActorID, input.RequestID, "masterdata.formula.activated", current, formulaToAuditMap(before), formulaToAuditMap(current), now)
	if err != nil {
		return FormulaResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return FormulaResult{}, err
	}

	return FormulaResult{Formula: current, AuditLogID: log.ID}, nil
}

func (s *PostgresFormulaCatalog) CalculateRequirement(ctx context.Context, input CalculateFormulaRequirementInput) (FormulaRequirementResult, error) {
	formula, err := s.Get(ctx, input.ID)
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

	return FormulaRequirementResult{Formula: formula, PlannedQty: plannedQty, PlannedUOM: plannedUOM, Requirements: requirements}, nil
}

func (s *PostgresFormulaCatalog) saveNewFormula(ctx context.Context, orgID string, formula domain.Formula) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	duplicate := ""
	if err := tx.QueryRowContext(ctx, selectPostgresFormulaDuplicateSQL, orgID, formula.FinishedItemID, formula.FormulaVersion).Scan(&duplicate); err == nil {
		return ErrDuplicateFormulaVersion
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var formulaUUID string
	if err := tx.QueryRowContext(ctx, insertPostgresFormulaSQL,
		orgID,
		formula.ID,
		formula.FormulaCode,
		formula.FinishedItemID,
		formula.FinishedItemID,
		formula.FinishedSKU,
		formula.FinishedItemName,
		string(formula.FinishedItemType),
		formula.FormulaVersion,
		formula.BatchQty.String(),
		formula.BatchUOMCode.String(),
		formula.BaseBatchQty.String(),
		formula.BaseBatchUOMCode.String(),
		string(formula.Status),
		string(formula.ApprovalStatus),
		nullablePostgresItemText(formula.EffectiveFrom),
		nullablePostgresItemText(formula.EffectiveTo),
		nullablePostgresItemText(formula.Note),
		nullablePostgresItemText(formula.ApprovedBy),
		nullablePostgresFormulaTime(formula.ApprovedAt),
		formula.CreatedAt,
		formula.UpdatedAt,
		formula.Version,
	).Scan(&formulaUUID); err != nil {
		return err
	}
	for _, line := range formula.Lines {
		if _, err := tx.ExecContext(ctx, insertPostgresFormulaLineSQL,
			orgID,
			formulaUUID,
			line.ID,
			line.LineNo,
			nullablePostgresItemText(line.ComponentItemID),
			nullablePostgresItemText(line.ComponentSKU),
			nullablePostgresItemText(line.ComponentName),
			string(line.ComponentType),
			line.EnteredQty.String(),
			line.EnteredUOMCode.String(),
			line.CalcQty.String(),
			line.CalcUOMCode.String(),
			line.StockBaseQty.String(),
			line.StockBaseUOMCode.String(),
			line.WastePercent.String(),
			line.IsRequired,
			line.IsStockManaged,
			string(line.LineStatus),
			nullablePostgresItemText(line.Note),
		); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	return nil
}

func (s *PostgresFormulaCatalog) getFormulaParentItem(ctx context.Context, orgID string, parentID string) (domain.Item, error) {
	if strings.TrimSpace(parentID) == "" {
		return domain.Item{}, ErrFormulaParentItemNotFound
	}
	var item domain.Item
	err := s.db.QueryRowContext(ctx, selectPostgresFormulaParentItemSQL, orgID, strings.TrimSpace(parentID)).Scan(
		&item.ID,
		&item.ItemCode,
		&item.SKUCode,
		&item.Name,
		&item.Type,
		&item.Status,
		&item.UOMBase,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, ErrFormulaParentItemNotFound
	}
	if err != nil {
		return domain.Item{}, err
	}

	return item, nil
}

func (s *PostgresFormulaCatalog) scanFormulaWithLines(ctx context.Context, orgID string, scanner interface{ Scan(dest ...any) error }) (domain.Formula, error) {
	var (
		header             postgresFormulaHeader
		batchQty           string
		baseBatchQty       string
		approvedAtNullable sql.NullTime
	)
	if err := scanner.Scan(
		&header.ID,
		&header.FormulaCode,
		&header.FinishedItemID,
		&header.FinishedSKU,
		&header.FinishedItemName,
		&header.FinishedItemType,
		&header.FormulaVersion,
		&batchQty,
		&header.BatchUOMCode,
		&baseBatchQty,
		&header.BaseBatchUOMCode,
		&header.Status,
		&header.ApprovalStatus,
		&header.EffectiveFrom,
		&header.EffectiveTo,
		&header.Note,
		&header.ApprovedBy,
		&approvedAtNullable,
		&header.CreatedAt,
		&header.UpdatedAt,
		&header.Version,
	); err != nil {
		return domain.Formula{}, err
	}
	if approvedAtNullable.Valid {
		header.ApprovedAt = approvedAtNullable.Time
	}
	lines, err := s.loadFormulaLines(ctx, orgID, header.ID)
	if err != nil {
		return domain.Formula{}, err
	}

	return domain.NewFormula(domain.NewFormulaInput{
		ID:               header.ID,
		FormulaCode:      header.FormulaCode,
		FinishedItemID:   header.FinishedItemID,
		FinishedSKU:      header.FinishedSKU,
		FinishedItemName: header.FinishedItemName,
		FinishedItemType: domain.ItemType(header.FinishedItemType),
		FormulaVersion:   header.FormulaVersion,
		BatchQty:         decimal.Decimal(batchQty),
		BatchUOMCode:     header.BatchUOMCode,
		BaseBatchQty:     decimal.Decimal(baseBatchQty),
		BaseBatchUOMCode: header.BaseBatchUOMCode,
		Status:           domain.FormulaStatus(header.Status),
		ApprovalStatus:   domain.FormulaApprovalStatus(header.ApprovalStatus),
		EffectiveFrom:    header.EffectiveFrom,
		EffectiveTo:      header.EffectiveTo,
		Lines:            lines,
		Note:             header.Note,
		CreatedAt:        header.CreatedAt,
		UpdatedAt:        header.UpdatedAt,
		ApprovedBy:       header.ApprovedBy,
		ApprovedAt:       header.ApprovedAt,
		Version:          header.Version,
	})
}

func (s *PostgresFormulaCatalog) loadFormulaLines(ctx context.Context, orgID string, formulaRef string) ([]domain.NewFormulaLineInput, error) {
	rows, err := s.db.QueryContext(ctx, selectPostgresFormulaLinesSQL, orgID, formulaRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.NewFormulaLineInput, 0)
	for rows.Next() {
		var line postgresFormulaLine
		if err := rows.Scan(
			&line.ID,
			&line.LineNo,
			&line.ComponentItemID,
			&line.ComponentSKU,
			&line.ComponentName,
			&line.ComponentType,
			&line.EnteredQty,
			&line.EnteredUOMCode,
			&line.CalcQty,
			&line.CalcUOMCode,
			&line.StockBaseQty,
			&line.StockBaseUOMCode,
			&line.WastePercent,
			&line.IsRequired,
			&line.IsStockManaged,
			&line.LineStatus,
			&line.Note,
		); err != nil {
			return nil, err
		}
		lines = append(lines, domain.NewFormulaLineInput{
			ID:               line.ID,
			LineNo:           line.LineNo,
			ComponentItemID:  line.ComponentItemID,
			ComponentSKU:     line.ComponentSKU,
			ComponentName:    line.ComponentName,
			ComponentType:    domain.FormulaComponentType(line.ComponentType),
			EnteredQty:       decimal.Decimal(line.EnteredQty),
			EnteredUOMCode:   line.EnteredUOMCode,
			CalcQty:          decimal.Decimal(line.CalcQty),
			CalcUOMCode:      line.CalcUOMCode,
			StockBaseQty:     decimal.Decimal(line.StockBaseQty),
			StockBaseUOMCode: line.StockBaseUOMCode,
			WastePercent:     decimal.Decimal(line.WastePercent),
			IsRequired:       line.IsRequired,
			IsStockManaged:   line.IsStockManaged,
			LineStatus:       domain.FormulaLineStatus(line.LineStatus),
			Note:             line.Note,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func (s *PostgresFormulaCatalog) resolveOrgID() (string, error) {
	if isPostgresItemUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", errors.New("formula catalog default org id is required")
}

func nullablePostgresFormulaTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

type postgresFormulaHeader struct {
	ID               string
	FormulaCode      string
	FinishedItemID   string
	FinishedSKU      string
	FinishedItemName string
	FinishedItemType string
	FormulaVersion   string
	BatchUOMCode     string
	BaseBatchUOMCode string
	Status           string
	ApprovalStatus   string
	EffectiveFrom    string
	EffectiveTo      string
	Note             string
	ApprovedBy       string
	ApprovedAt       time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Version          int
}

type postgresFormulaLine struct {
	ID               string
	LineNo           int
	ComponentItemID  string
	ComponentSKU     string
	ComponentName    string
	ComponentType    string
	EnteredQty       string
	EnteredUOMCode   string
	CalcQty          string
	CalcUOMCode      string
	StockBaseQty     string
	StockBaseUOMCode string
	WastePercent     string
	IsRequired       bool
	IsStockManaged   bool
	LineStatus       string
	Note             string
}
