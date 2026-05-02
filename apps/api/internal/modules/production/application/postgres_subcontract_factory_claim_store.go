package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresSubcontractFactoryClaimStoreConfig struct {
	DefaultOrgID string
}

type PostgresSubcontractFactoryClaimStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSubcontractFactoryClaimQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSubcontractFactoryClaimHeader struct {
	PersistedID        string
	ID                 string
	OrgID              string
	ClaimNo            string
	SubcontractOrderID string
	SubcontractOrderNo string
	FactoryID          string
	FactoryCode        string
	FactoryName        string
	ReceiptID          string
	ReceiptNo          string
	ReasonCode         string
	Reason             string
	Severity           string
	Status             string
	AffectedQty        string
	UOMCode            string
	BaseAffectedQty    string
	BaseUOMCode        string
	OwnerID            string
	OpenedBy           string
	OpenedAt           time.Time
	DueAt              time.Time
	AcknowledgedBy     string
	AcknowledgedAt     sql.NullTime
	ResolvedBy         string
	ResolvedAt         sql.NullTime
	ResolutionNote     string
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
	Version            int
}

func NewPostgresSubcontractFactoryClaimStore(
	db *sql.DB,
	cfg PostgresSubcontractFactoryClaimStoreConfig,
) PostgresSubcontractFactoryClaimStore {
	return PostgresSubcontractFactoryClaimStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSubcontractFactoryClaimOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSubcontractFactoryClaimHeadersBaseSQL = `
SELECT
  claim.id::text,
  claim.claim_ref,
  claim.org_ref,
  claim.claim_no,
  claim.subcontract_order_ref,
  claim.subcontract_order_no,
  claim.factory_ref,
  COALESCE(claim.factory_code, ''),
  claim.factory_name,
  COALESCE(claim.receipt_ref, ''),
  COALESCE(claim.receipt_no, ''),
  COALESCE(claim.reason_code, ''),
  claim.reason,
  claim.severity,
  claim.status,
  claim.affected_qty::text,
  claim.uom_code,
  claim.base_affected_qty::text,
  claim.base_uom_code,
  claim.owner_ref,
  claim.opened_by_ref,
  claim.opened_at,
  claim.due_at,
  COALESCE(claim.acknowledged_by_ref, ''),
  claim.acknowledged_at,
  COALESCE(claim.resolved_by_ref, ''),
  claim.resolved_at,
  COALESCE(claim.resolution_note, ''),
  claim.created_at,
  claim.created_by_ref,
  claim.updated_at,
  claim.updated_by_ref,
  claim.version
FROM subcontract.subcontract_factory_claims AS claim`

const findSubcontractFactoryClaimHeaderSQL = selectSubcontractFactoryClaimHeadersBaseSQL + `
WHERE lower(claim.claim_ref) = lower($1)
   OR claim.id::text = $1
   OR lower(claim.claim_no) = lower($1)
LIMIT 1`

const selectSubcontractFactoryClaimsByOrderSQL = selectSubcontractFactoryClaimHeadersBaseSQL + `
WHERE lower(claim.subcontract_order_ref) = lower($1)
   OR claim.subcontract_order_id::text = $1
   OR lower(claim.subcontract_order_no) = lower($1)
ORDER BY claim.opened_at DESC, claim.claim_no DESC`

const selectSubcontractFactoryClaimEvidenceSQL = `
SELECT
  evidence.evidence_ref,
  evidence.evidence_type,
  COALESCE(evidence.file_name, ''),
  COALESCE(evidence.object_key, ''),
  COALESCE(evidence.external_url, ''),
  COALESCE(evidence.note, '')
FROM subcontract.subcontract_factory_claim_evidence AS evidence
WHERE evidence.factory_claim_id = $1::uuid
ORDER BY evidence.evidence_type, evidence.evidence_ref`

const upsertSubcontractFactoryClaimSQL = `
INSERT INTO subcontract.subcontract_factory_claims (
  id,
  org_id,
  org_ref,
  claim_ref,
  claim_no,
  subcontract_order_id,
  subcontract_order_ref,
  subcontract_order_no,
  factory_ref,
  factory_code,
  factory_name,
  receipt_ref,
  receipt_no,
  reason_code,
  reason,
  severity,
  status,
  affected_qty,
  uom_code,
  base_affected_qty,
  base_uom_code,
  owner_ref,
  opened_by_ref,
  opened_at,
  due_at,
  acknowledged_by_ref,
  acknowledged_at,
  resolved_by_ref,
  resolved_at,
  resolution_note,
  created_at,
  created_by_ref,
  updated_at,
  updated_by_ref,
  version
) VALUES (
  COALESCE(CASE WHEN NULLIF($2::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $2::uuid END, gen_random_uuid()),
  $1::uuid,
  $3,
  $2,
  $4,
  (
    SELECT subcontract_order.id
    FROM subcontract.subcontract_orders AS subcontract_order
    WHERE subcontract_order.org_id = $1::uuid
      AND (
        subcontract_order.id::text = $5
        OR lower(subcontract_order.order_ref) = lower($5)
      )
    LIMIT 1
  ),
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
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26,
  $27,
  $28,
  $29,
  $30,
  $31,
  $32,
  $33
)
ON CONFLICT (org_id, claim_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  claim_no = EXCLUDED.claim_no,
  subcontract_order_id = EXCLUDED.subcontract_order_id,
  subcontract_order_ref = EXCLUDED.subcontract_order_ref,
  subcontract_order_no = EXCLUDED.subcontract_order_no,
  factory_ref = EXCLUDED.factory_ref,
  factory_code = EXCLUDED.factory_code,
  factory_name = EXCLUDED.factory_name,
  receipt_ref = EXCLUDED.receipt_ref,
  receipt_no = EXCLUDED.receipt_no,
  reason_code = EXCLUDED.reason_code,
  reason = EXCLUDED.reason,
  severity = EXCLUDED.severity,
  status = EXCLUDED.status,
  affected_qty = EXCLUDED.affected_qty,
  uom_code = EXCLUDED.uom_code,
  base_affected_qty = EXCLUDED.base_affected_qty,
  base_uom_code = EXCLUDED.base_uom_code,
  owner_ref = EXCLUDED.owner_ref,
  opened_by_ref = EXCLUDED.opened_by_ref,
  opened_at = EXCLUDED.opened_at,
  due_at = EXCLUDED.due_at,
  acknowledged_by_ref = EXCLUDED.acknowledged_by_ref,
  acknowledged_at = EXCLUDED.acknowledged_at,
  resolved_by_ref = EXCLUDED.resolved_by_ref,
  resolved_at = EXCLUDED.resolved_at,
  resolution_note = EXCLUDED.resolution_note,
  created_at = EXCLUDED.created_at,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const deleteSubcontractFactoryClaimEvidenceSQL = `
DELETE FROM subcontract.subcontract_factory_claim_evidence
WHERE factory_claim_id = $1::uuid`

const insertSubcontractFactoryClaimEvidenceSQL = `
INSERT INTO subcontract.subcontract_factory_claim_evidence (
  id,
  org_id,
  factory_claim_id,
  evidence_ref,
  evidence_type,
  file_name,
  object_key,
  external_url,
  note
) VALUES (
  COALESCE(CASE WHEN NULLIF($1::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $1::uuid END, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $1,
  $4,
  $5,
  $6,
  $7,
  $8
)`

func (s PostgresSubcontractFactoryClaimStore) Save(
	ctx context.Context,
	claim productiondomain.SubcontractFactoryClaim,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := claim.Validate(); err != nil {
		return err
	}
	return s.withTx(ctx, func(tx *sql.Tx) error {
		orgID, err := s.resolveOrgID(ctx, tx, claim.OrgID)
		if err != nil {
			return err
		}
		persistedID, err := upsertPostgresSubcontractFactoryClaim(ctx, tx, orgID, claim)
		if err != nil {
			return err
		}
		return replacePostgresSubcontractFactoryClaimEvidence(ctx, tx, orgID, persistedID, claim)
	})
}

func (s PostgresSubcontractFactoryClaimStore) Get(
	ctx context.Context,
	id string,
) (productiondomain.SubcontractFactoryClaim, error) {
	if s.db == nil {
		return productiondomain.SubcontractFactoryClaim{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSubcontractFactoryClaimHeaderSQL, strings.TrimSpace(id))
	claim, err := scanPostgresSubcontractFactoryClaim(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return productiondomain.SubcontractFactoryClaim{}, ErrSubcontractFactoryClaimNotFound
	}
	if err != nil {
		return productiondomain.SubcontractFactoryClaim{}, err
	}

	return claim, nil
}

func (s PostgresSubcontractFactoryClaimStore) ListBySubcontractOrder(
	ctx context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractFactoryClaim, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSubcontractFactoryClaimsByOrderSQL, strings.TrimSpace(subcontractOrderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := make([]productiondomain.SubcontractFactoryClaim, 0)
	for rows.Next() {
		claim, err := scanPostgresSubcontractFactoryClaim(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return claims, nil
}

func (s PostgresSubcontractFactoryClaimStore) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin subcontract factory claim transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subcontract factory claim transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSubcontractFactoryClaimStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSubcontractFactoryClaimQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSubcontractFactoryClaimUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSubcontractFactoryClaimOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSubcontractFactoryClaimUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve subcontract factory claim org %q: %w", orgRef, err)
		}
	}
	if isPostgresSubcontractFactoryClaimUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("subcontract factory claim org %q cannot be resolved", orgRef)
}

func scanPostgresSubcontractFactoryClaim(
	ctx context.Context,
	queryer postgresSubcontractFactoryClaimQueryer,
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractFactoryClaim, error) {
	header, err := scanPostgresSubcontractFactoryClaimHeader(row)
	if err != nil {
		return productiondomain.SubcontractFactoryClaim{}, err
	}
	evidence, err := listPostgresSubcontractFactoryClaimEvidence(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractFactoryClaim{}, err
	}

	return buildPostgresSubcontractFactoryClaim(header, evidence)
}

func scanPostgresSubcontractFactoryClaimHeader(
	row interface{ Scan(dest ...any) error },
) (postgresSubcontractFactoryClaimHeader, error) {
	var header postgresSubcontractFactoryClaimHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.ClaimNo,
		&header.SubcontractOrderID,
		&header.SubcontractOrderNo,
		&header.FactoryID,
		&header.FactoryCode,
		&header.FactoryName,
		&header.ReceiptID,
		&header.ReceiptNo,
		&header.ReasonCode,
		&header.Reason,
		&header.Severity,
		&header.Status,
		&header.AffectedQty,
		&header.UOMCode,
		&header.BaseAffectedQty,
		&header.BaseUOMCode,
		&header.OwnerID,
		&header.OpenedBy,
		&header.OpenedAt,
		&header.DueAt,
		&header.AcknowledgedBy,
		&header.AcknowledgedAt,
		&header.ResolvedBy,
		&header.ResolvedAt,
		&header.ResolutionNote,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
	)
	if err != nil {
		return postgresSubcontractFactoryClaimHeader{}, err
	}

	return header, nil
}

func listPostgresSubcontractFactoryClaimEvidence(
	ctx context.Context,
	queryer postgresSubcontractFactoryClaimQueryer,
	persistedID string,
) ([]productiondomain.SubcontractFactoryClaimEvidence, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractFactoryClaimEvidenceSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evidence := make([]productiondomain.SubcontractFactoryClaimEvidence, 0)
	for rows.Next() {
		item, err := scanPostgresSubcontractFactoryClaimEvidence(rows)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return evidence, nil
}

func scanPostgresSubcontractFactoryClaimEvidence(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractFactoryClaimEvidence, error) {
	var (
		id           string
		evidenceType string
		fileName     string
		objectKey    string
		externalURL  string
		note         string
	)
	if err := row.Scan(&id, &evidenceType, &fileName, &objectKey, &externalURL, &note); err != nil {
		return productiondomain.SubcontractFactoryClaimEvidence{}, err
	}

	return productiondomain.NewSubcontractFactoryClaimEvidence(productiondomain.NewSubcontractFactoryClaimEvidenceInput{
		ID:           id,
		EvidenceType: evidenceType,
		FileName:     fileName,
		ObjectKey:    objectKey,
		ExternalURL:  externalURL,
		Note:         note,
	})
}

func buildPostgresSubcontractFactoryClaim(
	header postgresSubcontractFactoryClaimHeader,
	evidence []productiondomain.SubcontractFactoryClaimEvidence,
) (productiondomain.SubcontractFactoryClaim, error) {
	affectedQty, err := decimal.ParseQuantity(header.AffectedQty)
	if err != nil {
		return productiondomain.SubcontractFactoryClaim{}, err
	}
	baseAffectedQty, err := decimal.ParseQuantity(header.BaseAffectedQty)
	if err != nil {
		return productiondomain.SubcontractFactoryClaim{}, err
	}
	evidenceInputs := make([]productiondomain.NewSubcontractFactoryClaimEvidenceInput, 0, len(evidence))
	for _, item := range evidence {
		evidenceInputs = append(evidenceInputs, productiondomain.NewSubcontractFactoryClaimEvidenceInput{
			ID:           item.ID,
			EvidenceType: item.EvidenceType,
			FileName:     item.FileName,
			ObjectKey:    item.ObjectKey,
			ExternalURL:  item.ExternalURL,
			Note:         item.Note,
		})
	}

	claim, err := productiondomain.NewSubcontractFactoryClaim(productiondomain.NewSubcontractFactoryClaimInput{
		ID:                 header.ID,
		OrgID:              header.OrgID,
		ClaimNo:            header.ClaimNo,
		SubcontractOrderID: header.SubcontractOrderID,
		SubcontractOrderNo: header.SubcontractOrderNo,
		FactoryID:          header.FactoryID,
		FactoryCode:        header.FactoryCode,
		FactoryName:        header.FactoryName,
		ReceiptID:          header.ReceiptID,
		ReceiptNo:          header.ReceiptNo,
		ReasonCode:         header.ReasonCode,
		Reason:             header.Reason,
		Severity:           header.Severity,
		Status:             productiondomain.SubcontractFactoryClaimStatus(header.Status),
		AffectedQty:        affectedQty,
		UOMCode:            header.UOMCode,
		BaseAffectedQty:    baseAffectedQty,
		BaseUOMCode:        header.BaseUOMCode,
		Evidence:           evidenceInputs,
		OwnerID:            header.OwnerID,
		OpenedBy:           header.OpenedBy,
		OpenedAt:           header.OpenedAt,
		DueAt:              header.DueAt,
		CreatedAt:          header.CreatedAt,
		CreatedBy:          header.CreatedBy,
		UpdatedAt:          header.UpdatedAt,
		UpdatedBy:          header.UpdatedBy,
	})
	if err != nil {
		return productiondomain.SubcontractFactoryClaim{}, err
	}
	claim.AcknowledgedBy = header.AcknowledgedBy
	claim.AcknowledgedAt = nullablePostgresSubcontractFactoryClaimTimeValue(header.AcknowledgedAt)
	claim.ResolvedBy = header.ResolvedBy
	claim.ResolvedAt = nullablePostgresSubcontractFactoryClaimTimeValue(header.ResolvedAt)
	claim.ResolutionNote = header.ResolutionNote
	claim.Version = header.Version
	if err := claim.Validate(); err != nil {
		return productiondomain.SubcontractFactoryClaim{}, err
	}

	return claim, nil
}

func upsertPostgresSubcontractFactoryClaim(
	ctx context.Context,
	queryer postgresSubcontractFactoryClaimQueryer,
	orgID string,
	claim productiondomain.SubcontractFactoryClaim,
) (string, error) {
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSubcontractFactoryClaimSQL,
		orgID,
		claim.ID,
		claim.OrgID,
		claim.ClaimNo,
		claim.SubcontractOrderID,
		claim.SubcontractOrderNo,
		claim.FactoryID,
		nullablePostgresSubcontractFactoryClaimText(claim.FactoryCode),
		claim.FactoryName,
		nullablePostgresSubcontractFactoryClaimText(claim.ReceiptID),
		nullablePostgresSubcontractFactoryClaimText(claim.ReceiptNo),
		nullablePostgresSubcontractFactoryClaimText(claim.ReasonCode),
		claim.Reason,
		claim.Severity,
		string(claim.Status),
		claim.AffectedQty.String(),
		claim.UOMCode.String(),
		claim.BaseAffectedQty.String(),
		claim.BaseUOMCode.String(),
		claim.OwnerID,
		claim.OpenedBy,
		claim.OpenedAt.UTC(),
		claim.DueAt.UTC(),
		nullablePostgresSubcontractFactoryClaimText(claim.AcknowledgedBy),
		nullablePostgresSubcontractFactoryClaimTime(claim.AcknowledgedAt),
		nullablePostgresSubcontractFactoryClaimText(claim.ResolvedBy),
		nullablePostgresSubcontractFactoryClaimTime(claim.ResolvedAt),
		nullablePostgresSubcontractFactoryClaimText(claim.ResolutionNote),
		claim.CreatedAt.UTC(),
		claim.CreatedBy,
		claim.UpdatedAt.UTC(),
		claim.UpdatedBy,
		claim.Version,
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert subcontract factory claim %q: %w", claim.ID, err)
	}

	return persistedID, nil
}

func replacePostgresSubcontractFactoryClaimEvidence(
	ctx context.Context,
	queryer postgresSubcontractFactoryClaimQueryer,
	orgID string,
	persistedID string,
	claim productiondomain.SubcontractFactoryClaim,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractFactoryClaimEvidenceSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract factory claim evidence: %w", err)
	}
	for _, evidence := range claim.Evidence {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractFactoryClaimEvidenceSQL,
			evidence.ID,
			orgID,
			persistedID,
			evidence.EvidenceType,
			nullablePostgresSubcontractFactoryClaimText(evidence.FileName),
			nullablePostgresSubcontractFactoryClaimText(evidence.ObjectKey),
			nullablePostgresSubcontractFactoryClaimText(evidence.ExternalURL),
			nullablePostgresSubcontractFactoryClaimText(evidence.Note),
		); err != nil {
			return fmt.Errorf("insert subcontract factory claim evidence %q: %w", evidence.ID, err)
		}
	}

	return nil
}

func nullablePostgresSubcontractFactoryClaimText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresSubcontractFactoryClaimTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func nullablePostgresSubcontractFactoryClaimTimeValue(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func isPostgresSubcontractFactoryClaimUUIDText(value string) bool {
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
			if !isPostgresSubcontractFactoryClaimHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSubcontractFactoryClaimHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SubcontractFactoryClaimStore = PostgresSubcontractFactoryClaimStore{}
