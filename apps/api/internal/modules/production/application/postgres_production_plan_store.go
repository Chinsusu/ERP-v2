package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

type PostgresProductionPlanStoreConfig struct {
	DefaultOrgID string
}

type PostgresProductionPlanStore struct {
	db           *sql.DB
	auditLog     audit.LogStore
	defaultOrgID string
}

func NewPostgresProductionPlanStore(db *sql.DB, auditLog audit.LogStore, cfg PostgresProductionPlanStoreConfig) *PostgresProductionPlanStore {
	return &PostgresProductionPlanStore{
		db:           db,
		auditLog:     auditLog,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const upsertPostgresProductionPlanSQL = `
INSERT INTO subcontract.production_plans (
  org_id,
  plan_ref,
  plan_no,
  output_item_ref,
  output_sku,
  output_item_name,
  formula_ref,
  formula_code,
  formula_version,
  status,
  plan_payload,
  created_at,
  updated_at,
  version
) VALUES (
  $1::uuid,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11::jsonb,
  $12,
  $13,
  $14
)
ON CONFLICT (org_id, plan_ref)
DO UPDATE SET
  plan_no = EXCLUDED.plan_no,
  output_item_ref = EXCLUDED.output_item_ref,
  output_sku = EXCLUDED.output_sku,
  output_item_name = EXCLUDED.output_item_name,
  formula_ref = EXCLUDED.formula_ref,
  formula_code = EXCLUDED.formula_code,
  formula_version = EXCLUDED.formula_version,
  status = EXCLUDED.status,
  plan_payload = EXCLUDED.plan_payload,
  updated_at = EXCLUDED.updated_at,
  version = EXCLUDED.version`

const upsertPostgresPurchaseRequestDraftSQL = `
INSERT INTO subcontract.purchase_request_drafts (
  org_id,
  draft_ref,
  request_no,
  source_plan_ref,
  status,
  draft_payload,
  created_at
) VALUES (
  $1::uuid,
  $2,
  $3,
  $4,
  $5,
  $6::jsonb,
  $7
)
ON CONFLICT (org_id, draft_ref)
DO UPDATE SET
  request_no = EXCLUDED.request_no,
  source_plan_ref = EXCLUDED.source_plan_ref,
  status = EXCLUDED.status,
  draft_payload = EXCLUDED.draft_payload`

const selectPostgresProductionPlanPayloadSQL = `
SELECT plan_payload
FROM subcontract.production_plans
WHERE org_id = $1::uuid
  AND lower(plan_ref) = lower($2)
LIMIT 1`

const selectPostgresProductionPlanPayloadsSQL = `
SELECT plan_payload
FROM subcontract.production_plans`

func (s *PostgresProductionPlanStore) List(ctx context.Context, filter ProductionPlanFilter) ([]productiondomain.ProductionPlan, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("production plan postgres store is required")
	}
	orgID, err := s.resolveOrgID(ctx, "")
	if err != nil {
		return nil, err
	}
	query, args := buildPostgresProductionPlanListQuery(orgID, filter)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]productiondomain.ProductionPlan, 0)
	for rows.Next() {
		plan, err := scanProductionPlanPayload(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func (s *PostgresProductionPlanStore) Get(ctx context.Context, id string) (productiondomain.ProductionPlan, error) {
	if s == nil || s.db == nil {
		return productiondomain.ProductionPlan{}, errors.New("production plan postgres store is required")
	}
	orgID, err := s.resolveOrgID(ctx, "")
	if err != nil {
		return productiondomain.ProductionPlan{}, err
	}
	row := s.db.QueryRowContext(ctx, selectPostgresProductionPlanPayloadSQL, orgID, strings.TrimSpace(id))
	plan, err := scanProductionPlanPayload(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return productiondomain.ProductionPlan{}, ErrProductionPlanNotFound
		}
		return productiondomain.ProductionPlan{}, err
	}

	return plan, nil
}

func (s *PostgresProductionPlanStore) Save(ctx context.Context, plan productiondomain.ProductionPlan) error {
	if s == nil || s.db == nil {
		return errors.New("production plan postgres store is required")
	}
	if err := plan.Validate(); err != nil {
		return err
	}
	orgID, err := s.resolveOrgID(ctx, plan.OrgID)
	if err != nil {
		return err
	}
	planPayload, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("encode production plan payload: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(
		ctx,
		upsertPostgresProductionPlanSQL,
		orgID,
		plan.ID,
		plan.PlanNo,
		plan.OutputItemID,
		plan.OutputSKU,
		plan.OutputItemName,
		plan.FormulaID,
		plan.FormulaCode,
		plan.FormulaVersion,
		string(plan.Status),
		string(planPayload),
		plan.CreatedAt.UTC(),
		plan.UpdatedAt.UTC(),
		plan.Version,
	); err != nil {
		return err
	}
	if len(plan.PurchaseDraft.Lines) > 0 {
		draftPayload, err := json.Marshal(plan.PurchaseDraft)
		if err != nil {
			return fmt.Errorf("encode purchase request draft payload: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			upsertPostgresPurchaseRequestDraftSQL,
			orgID,
			plan.PurchaseDraft.ID,
			plan.PurchaseDraft.RequestNo,
			plan.ID,
			string(plan.PurchaseDraft.Status),
			string(draftPayload),
			plan.PurchaseDraft.CreatedAt.UTC(),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresProductionPlanStore) RecordAudit(ctx context.Context, log audit.Log) error {
	if s == nil {
		return errors.New("production plan postgres store is required")
	}
	if s.auditLog == nil {
		return errors.New("audit log store is required")
	}

	return s.auditLog.Record(ctx, log)
}

func (s *PostgresProductionPlanStore) resolveOrgID(ctx context.Context, orgRef string) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isProductionPlanUUIDText(orgRef) {
		return orgRef, nil
	}
	if isProductionPlanUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}
	if orgRef == "" {
		orgRef = defaultProductionPlanOrgID
	}
	if orgRef != "" {
		var orgID string
		err := s.db.QueryRowContext(ctx, `SELECT id::text FROM core.organizations WHERE code = $1 LIMIT 1`, orgRef).Scan(&orgID)
		if err == nil && isProductionPlanUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}

	return "", fmt.Errorf("production plan org %q cannot be resolved", orgRef)
}

func buildPostgresProductionPlanListQuery(orgID string, filter ProductionPlanFilter) (string, []any) {
	clauses := []string{"org_id = $1::uuid"}
	args := []any{orgID}
	addClause := func(clause string, value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		args = append(args, trimmed)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	addClause("output_item_ref = $%d", filter.OutputItemID)
	if search := strings.TrimSpace(filter.Search); search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		clauses = append(clauses, fmt.Sprintf("(lower(plan_no) LIKE $%d OR lower(output_sku) LIKE $%d OR lower(output_item_name) LIKE $%d)", len(args), len(args), len(args)))
	}
	if len(filter.Statuses) > 0 {
		statusPlaceholders := make([]string, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			if normalized := productiondomain.NormalizeProductionPlanStatus(status); normalized != "" {
				args = append(args, string(normalized))
				statusPlaceholders = append(statusPlaceholders, fmt.Sprintf("$%d", len(args)))
			}
		}
		if len(statusPlaceholders) > 0 {
			clauses = append(clauses, "status IN ("+strings.Join(statusPlaceholders, ", ")+")")
		}
	}

	return selectPostgresProductionPlanPayloadsSQL + "\nWHERE " + strings.Join(clauses, "\n  AND ") + "\nORDER BY created_at DESC, plan_no DESC", args
}

type productionPlanPayloadScanner interface {
	Scan(dest ...any) error
}

func scanProductionPlanPayload(scanner productionPlanPayloadScanner) (productiondomain.ProductionPlan, error) {
	var raw []byte
	if err := scanner.Scan(&raw); err != nil {
		return productiondomain.ProductionPlan{}, err
	}
	var plan productiondomain.ProductionPlan
	if err := json.Unmarshal(raw, &plan); err != nil {
		return productiondomain.ProductionPlan{}, err
	}
	if err := plan.Validate(); err != nil {
		return productiondomain.ProductionPlan{}, err
	}

	return plan, nil
}

func isProductionPlanUUIDText(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 36 {
		return false
	}
	for index, char := range value {
		switch index {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
	}

	return true
}
