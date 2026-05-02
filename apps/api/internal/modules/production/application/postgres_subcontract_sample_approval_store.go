package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
)

type PostgresSubcontractSampleApprovalStoreConfig struct {
	DefaultOrgID string
}

type PostgresSubcontractSampleApprovalStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSubcontractSampleApprovalQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSubcontractSampleApprovalHeader struct {
	PersistedID        string
	ID                 string
	OrgID              string
	SubcontractOrderID string
	SubcontractOrderNo string
	SampleCode         string
	FormulaVersion     string
	SpecVersion        string
	Status             string
	SubmittedBy        string
	SubmittedAt        time.Time
	DecisionBy         string
	DecisionAt         sql.NullTime
	DecisionReason     string
	StorageStatus      string
	Note               string
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
	Version            int
}

func NewPostgresSubcontractSampleApprovalStore(
	db *sql.DB,
	cfg PostgresSubcontractSampleApprovalStoreConfig,
) PostgresSubcontractSampleApprovalStore {
	return PostgresSubcontractSampleApprovalStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSubcontractSampleApprovalOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSubcontractSampleApprovalHeadersBaseSQL = `
SELECT
  sample.id::text,
  sample.sample_ref,
  sample.org_ref,
  sample.subcontract_order_ref,
  sample.subcontract_order_no,
  sample.sample_code,
  COALESCE(sample.formula_version, ''),
  COALESCE(sample.spec_version, ''),
  sample.status,
  sample.submitted_by_ref,
  sample.submitted_at,
  COALESCE(sample.decision_by_ref, ''),
  sample.decision_at,
  COALESCE(sample.decision_reason, ''),
  COALESCE(sample.storage_status, ''),
  COALESCE(sample.note, ''),
  sample.created_at,
  sample.created_by_ref,
  sample.updated_at,
  sample.updated_by_ref,
  sample.version
FROM subcontract.subcontract_sample_approvals AS sample`

const findSubcontractSampleApprovalHeaderSQL = selectSubcontractSampleApprovalHeadersBaseSQL + `
WHERE lower(sample.sample_ref) = lower($1)
   OR sample.id::text = $1
   OR lower(sample.sample_code) = lower($1)
LIMIT 1`

const findLatestSubcontractSampleApprovalByOrderSQL = selectSubcontractSampleApprovalHeadersBaseSQL + `
WHERE lower(sample.subcontract_order_ref) = lower($1)
   OR sample.subcontract_order_id::text = $1
   OR lower(sample.subcontract_order_no) = lower($1)
ORDER BY sample.submitted_at DESC, sample.created_at DESC, sample.sample_code DESC
LIMIT 1`

const selectSubcontractSampleApprovalEvidenceSQL = `
SELECT
  evidence.evidence_ref,
  evidence.evidence_type,
  COALESCE(evidence.file_name, ''),
  COALESCE(evidence.object_key, ''),
  COALESCE(evidence.external_url, ''),
  COALESCE(evidence.note, ''),
  evidence.created_at,
  COALESCE(evidence.created_by_ref, '')
FROM subcontract.subcontract_sample_approval_evidence AS evidence
WHERE evidence.sample_approval_id = $1::uuid
ORDER BY evidence.evidence_type, evidence.evidence_ref`

const upsertSubcontractSampleApprovalSQL = `
INSERT INTO subcontract.subcontract_sample_approvals (
  id,
  org_id,
  org_ref,
  sample_ref,
  subcontract_order_id,
  subcontract_order_ref,
  subcontract_order_no,
  sample_code,
  formula_version,
  spec_version,
  status,
  submitted_by_ref,
  submitted_at,
  decision_by_ref,
  decision_at,
  decision_reason,
  storage_status,
  note,
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
  (
    SELECT subcontract_order.id
    FROM subcontract.subcontract_orders AS subcontract_order
    WHERE subcontract_order.org_id = $1::uuid
      AND (
        subcontract_order.id::text = $4
        OR lower(subcontract_order.order_ref) = lower($4)
      )
    LIMIT 1
  ),
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
  $19,
  $20,
  $21
)
ON CONFLICT (org_id, sample_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  subcontract_order_id = EXCLUDED.subcontract_order_id,
  subcontract_order_ref = EXCLUDED.subcontract_order_ref,
  subcontract_order_no = EXCLUDED.subcontract_order_no,
  sample_code = EXCLUDED.sample_code,
  formula_version = EXCLUDED.formula_version,
  spec_version = EXCLUDED.spec_version,
  status = EXCLUDED.status,
  submitted_by_ref = EXCLUDED.submitted_by_ref,
  submitted_at = EXCLUDED.submitted_at,
  decision_by_ref = EXCLUDED.decision_by_ref,
  decision_at = EXCLUDED.decision_at,
  decision_reason = EXCLUDED.decision_reason,
  storage_status = EXCLUDED.storage_status,
  note = EXCLUDED.note,
  created_at = EXCLUDED.created_at,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const deleteSubcontractSampleApprovalEvidenceSQL = `
DELETE FROM subcontract.subcontract_sample_approval_evidence
WHERE sample_approval_id = $1::uuid`

const insertSubcontractSampleApprovalEvidenceSQL = `
INSERT INTO subcontract.subcontract_sample_approval_evidence (
  id,
  org_id,
  sample_approval_id,
  evidence_ref,
  evidence_type,
  file_name,
  object_key,
  external_url,
  note,
  created_at,
  created_by_ref
) VALUES (
  COALESCE(CASE WHEN NULLIF($1::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $1::uuid END, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $1,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10
)`

func (s PostgresSubcontractSampleApprovalStore) Save(
	ctx context.Context,
	sampleApproval productiondomain.SubcontractSampleApproval,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := sampleApproval.Validate(); err != nil {
		return err
	}
	return s.withTx(ctx, func(tx *sql.Tx) error {
		orgID, err := s.resolveOrgID(ctx, tx, sampleApproval.OrgID)
		if err != nil {
			return err
		}
		persistedID, err := upsertPostgresSubcontractSampleApproval(ctx, tx, orgID, sampleApproval)
		if err != nil {
			return err
		}
		return replacePostgresSubcontractSampleApprovalEvidence(ctx, tx, orgID, persistedID, sampleApproval)
	})
}

func (s PostgresSubcontractSampleApprovalStore) Get(
	ctx context.Context,
	id string,
) (productiondomain.SubcontractSampleApproval, error) {
	if s.db == nil {
		return productiondomain.SubcontractSampleApproval{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSubcontractSampleApprovalHeaderSQL, strings.TrimSpace(id))
	sampleApproval, err := scanPostgresSubcontractSampleApproval(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return productiondomain.SubcontractSampleApproval{}, ErrSubcontractSampleApprovalNotFound
	}
	if err != nil {
		return productiondomain.SubcontractSampleApproval{}, err
	}

	return sampleApproval, nil
}

func (s PostgresSubcontractSampleApprovalStore) GetLatestBySubcontractOrder(
	ctx context.Context,
	subcontractOrderID string,
) (productiondomain.SubcontractSampleApproval, error) {
	if s.db == nil {
		return productiondomain.SubcontractSampleApproval{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findLatestSubcontractSampleApprovalByOrderSQL, strings.TrimSpace(subcontractOrderID))
	sampleApproval, err := scanPostgresSubcontractSampleApproval(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return productiondomain.SubcontractSampleApproval{}, ErrSubcontractSampleApprovalNotFound
	}
	if err != nil {
		return productiondomain.SubcontractSampleApproval{}, err
	}

	return sampleApproval, nil
}

func (s PostgresSubcontractSampleApprovalStore) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	if tx, ok := postgresSubcontractTxFromContext(ctx); ok {
		return fn(tx)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin subcontract sample approval transaction: %w", err)
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
		return fmt.Errorf("commit subcontract sample approval transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSubcontractSampleApprovalStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSubcontractSampleApprovalQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSubcontractSampleApprovalUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSubcontractSampleApprovalOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSubcontractSampleApprovalUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve subcontract sample approval org %q: %w", orgRef, err)
		}
	}
	if isPostgresSubcontractSampleApprovalUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("subcontract sample approval org %q cannot be resolved", orgRef)
}

func scanPostgresSubcontractSampleApproval(
	ctx context.Context,
	queryer postgresSubcontractSampleApprovalQueryer,
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractSampleApproval, error) {
	header, err := scanPostgresSubcontractSampleApprovalHeader(row)
	if err != nil {
		return productiondomain.SubcontractSampleApproval{}, err
	}
	evidence, err := listPostgresSubcontractSampleApprovalEvidence(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractSampleApproval{}, err
	}

	return buildPostgresSubcontractSampleApproval(header, evidence)
}

func scanPostgresSubcontractSampleApprovalHeader(
	row interface{ Scan(dest ...any) error },
) (postgresSubcontractSampleApprovalHeader, error) {
	var header postgresSubcontractSampleApprovalHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.SubcontractOrderID,
		&header.SubcontractOrderNo,
		&header.SampleCode,
		&header.FormulaVersion,
		&header.SpecVersion,
		&header.Status,
		&header.SubmittedBy,
		&header.SubmittedAt,
		&header.DecisionBy,
		&header.DecisionAt,
		&header.DecisionReason,
		&header.StorageStatus,
		&header.Note,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
	)
	if err != nil {
		return postgresSubcontractSampleApprovalHeader{}, err
	}

	return header, nil
}

func listPostgresSubcontractSampleApprovalEvidence(
	ctx context.Context,
	queryer postgresSubcontractSampleApprovalQueryer,
	persistedID string,
) ([]productiondomain.SubcontractSampleApprovalEvidence, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractSampleApprovalEvidenceSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evidence := make([]productiondomain.SubcontractSampleApprovalEvidence, 0)
	for rows.Next() {
		item, err := scanPostgresSubcontractSampleApprovalEvidence(rows)
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

func scanPostgresSubcontractSampleApprovalEvidence(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractSampleApprovalEvidence, error) {
	var (
		id           string
		evidenceType string
		fileName     string
		objectKey    string
		externalURL  string
		note         string
		createdAt    time.Time
		createdBy    string
	)
	if err := row.Scan(&id, &evidenceType, &fileName, &objectKey, &externalURL, &note, &createdAt, &createdBy); err != nil {
		return productiondomain.SubcontractSampleApprovalEvidence{}, err
	}

	return productiondomain.NewSubcontractSampleApprovalEvidence(productiondomain.NewSubcontractSampleApprovalEvidenceInput{
		ID:           id,
		EvidenceType: evidenceType,
		FileName:     fileName,
		ObjectKey:    objectKey,
		ExternalURL:  externalURL,
		Note:         note,
		CreatedAt:    createdAt,
		CreatedBy:    createdBy,
	})
}

func buildPostgresSubcontractSampleApproval(
	header postgresSubcontractSampleApprovalHeader,
	evidence []productiondomain.SubcontractSampleApprovalEvidence,
) (productiondomain.SubcontractSampleApproval, error) {
	evidenceInputs := make([]productiondomain.NewSubcontractSampleApprovalEvidenceInput, 0, len(evidence))
	for _, item := range evidence {
		evidenceInputs = append(evidenceInputs, productiondomain.NewSubcontractSampleApprovalEvidenceInput{
			ID:           item.ID,
			EvidenceType: item.EvidenceType,
			FileName:     item.FileName,
			ObjectKey:    item.ObjectKey,
			ExternalURL:  item.ExternalURL,
			Note:         item.Note,
			CreatedAt:    item.CreatedAt,
			CreatedBy:    item.CreatedBy,
		})
	}
	sampleApproval, err := productiondomain.NewSubcontractSampleApproval(productiondomain.NewSubcontractSampleApprovalInput{
		ID:                 header.ID,
		OrgID:              header.OrgID,
		SubcontractOrderID: header.SubcontractOrderID,
		SubcontractOrderNo: header.SubcontractOrderNo,
		SampleCode:         header.SampleCode,
		FormulaVersion:     header.FormulaVersion,
		SpecVersion:        header.SpecVersion,
		Status:             productiondomain.SubcontractSampleApprovalStatus(header.Status),
		Evidence:           evidenceInputs,
		SubmittedBy:        header.SubmittedBy,
		SubmittedAt:        header.SubmittedAt,
		DecisionBy:         header.DecisionBy,
		DecisionAt:         nullablePostgresSubcontractSampleApprovalTimeValue(header.DecisionAt),
		DecisionReason:     header.DecisionReason,
		StorageStatus:      header.StorageStatus,
		Note:               header.Note,
		CreatedAt:          header.CreatedAt,
		CreatedBy:          header.CreatedBy,
		UpdatedAt:          header.UpdatedAt,
		UpdatedBy:          header.UpdatedBy,
	})
	if err != nil {
		return productiondomain.SubcontractSampleApproval{}, err
	}
	sampleApproval.Version = header.Version
	if err := sampleApproval.Validate(); err != nil {
		return productiondomain.SubcontractSampleApproval{}, err
	}

	return sampleApproval, nil
}

func upsertPostgresSubcontractSampleApproval(
	ctx context.Context,
	queryer postgresSubcontractSampleApprovalQueryer,
	orgID string,
	sampleApproval productiondomain.SubcontractSampleApproval,
) (string, error) {
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSubcontractSampleApprovalSQL,
		orgID,
		sampleApproval.ID,
		sampleApproval.OrgID,
		sampleApproval.SubcontractOrderID,
		sampleApproval.SubcontractOrderNo,
		sampleApproval.SampleCode,
		nullablePostgresSubcontractSampleApprovalText(sampleApproval.FormulaVersion),
		nullablePostgresSubcontractSampleApprovalText(sampleApproval.SpecVersion),
		string(sampleApproval.Status),
		sampleApproval.SubmittedBy,
		sampleApproval.SubmittedAt.UTC(),
		nullablePostgresSubcontractSampleApprovalText(sampleApproval.DecisionBy),
		nullablePostgresSubcontractSampleApprovalTime(sampleApproval.DecisionAt),
		nullablePostgresSubcontractSampleApprovalText(sampleApproval.DecisionReason),
		nullablePostgresSubcontractSampleApprovalText(sampleApproval.StorageStatus),
		nullablePostgresSubcontractSampleApprovalText(sampleApproval.Note),
		sampleApproval.CreatedAt.UTC(),
		sampleApproval.CreatedBy,
		sampleApproval.UpdatedAt.UTC(),
		sampleApproval.UpdatedBy,
		sampleApproval.Version,
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert subcontract sample approval %q: %w", sampleApproval.ID, err)
	}

	return persistedID, nil
}

func replacePostgresSubcontractSampleApprovalEvidence(
	ctx context.Context,
	queryer postgresSubcontractSampleApprovalQueryer,
	orgID string,
	persistedID string,
	sampleApproval productiondomain.SubcontractSampleApproval,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractSampleApprovalEvidenceSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract sample approval evidence: %w", err)
	}
	for _, evidence := range sampleApproval.Evidence {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractSampleApprovalEvidenceSQL,
			evidence.ID,
			orgID,
			persistedID,
			evidence.EvidenceType,
			nullablePostgresSubcontractSampleApprovalText(evidence.FileName),
			nullablePostgresSubcontractSampleApprovalText(evidence.ObjectKey),
			nullablePostgresSubcontractSampleApprovalText(evidence.ExternalURL),
			nullablePostgresSubcontractSampleApprovalText(evidence.Note),
			evidence.CreatedAt.UTC(),
			evidence.CreatedBy,
		); err != nil {
			return fmt.Errorf("insert subcontract sample approval evidence %q: %w", evidence.ID, err)
		}
	}

	return nil
}

func nullablePostgresSubcontractSampleApprovalText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresSubcontractSampleApprovalTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func nullablePostgresSubcontractSampleApprovalTimeValue(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func isPostgresSubcontractSampleApprovalUUIDText(value string) bool {
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
			if !isPostgresSubcontractSampleApprovalHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSubcontractSampleApprovalHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SubcontractSampleApprovalStore = PostgresSubcontractSampleApprovalStore{}
