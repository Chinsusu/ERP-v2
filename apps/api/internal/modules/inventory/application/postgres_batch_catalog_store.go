package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

type PostgresBatchCatalogStore struct {
	db       *sql.DB
	auditLog audit.LogStore
}

func NewPostgresBatchCatalogStore(
	db *sql.DB,
	auditLog audit.LogStore,
) PostgresBatchCatalogStore {
	return PostgresBatchCatalogStore{
		db:       db,
		auditLog: firstBatchAuditStore([]audit.LogStore{auditLog}),
	}
}

const selectPostgresBatchBaseSQL = `
SELECT
  batch.id::text,
  batch.org_id::text,
  COALESCE(batch.batch_ref, batch.id::text),
  COALESCE(batch.org_ref, batch.org_id::text),
  COALESCE(batch.item_ref, batch.item_id::text),
  item.sku,
  item.name,
  batch.batch_no,
  COALESCE(batch.supplier_ref, batch.supplier_id::text, ''),
  batch.mfg_date,
  batch.expiry_date,
  batch.qc_status,
  batch.status,
  batch.created_at,
  batch.updated_at
FROM inventory.batches AS batch
JOIN mdm.items AS item ON item.id = batch.item_id`

const selectPostgresBatchOrderSQL = `
ORDER BY item.sku,
  CASE WHEN batch.expiry_date IS NULL THEN 1 ELSE 0 END,
  batch.expiry_date,
  batch.batch_no`

const selectPostgresBatchForUpdateSQL = selectPostgresBatchBaseSQL + `
WHERE lower(COALESCE(batch.batch_ref, batch.id::text)) = lower($1)
   OR batch.id::text = $1
LIMIT 1
FOR UPDATE OF batch`

const updatePostgresBatchQCStatusSQL = `
UPDATE inventory.batches
SET qc_status = $2,
    updated_at = $3,
    updated_by = $4::uuid,
    updated_by_ref = $5,
    version = version + 1
WHERE id = $1::uuid`

const insertBatchQCAuditSQL = `
INSERT INTO audit.audit_logs (
  id,
  org_id,
  actor_id,
  action,
  entity_type,
  entity_id,
  request_id,
  before_data,
  after_data,
  metadata,
  created_at,
  log_ref,
  org_ref,
  actor_ref,
  entity_ref
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6::uuid,
  $7,
  $8::jsonb,
  $9::jsonb,
  $10::jsonb,
  $11,
  $12,
  $13,
  $14,
  $15
)`

func (s PostgresBatchCatalogStore) ListBatches(
	ctx context.Context,
	filter domain.BatchFilter,
) ([]domain.Batch, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}

	query, args := buildPostgresBatchListQuery(filter)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := make([]domain.Batch, 0)
	for rows.Next() {
		row, err := scanPostgresBatchRow(rows)
		if err != nil {
			return nil, err
		}
		batches = append(batches, row.batch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortBatches(batches)

	return batches, nil
}

func (s PostgresBatchCatalogStore) GetBatch(ctx context.Context, id string) (domain.Batch, error) {
	if s.db == nil {
		return domain.Batch{}, errors.New("database connection is required")
	}

	row := s.db.QueryRowContext(ctx, buildPostgresBatchLookupQuery(), strings.TrimSpace(id))
	batchRow, err := scanPostgresBatchRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Batch{}, ErrBatchNotFound
	}
	if err != nil {
		return domain.Batch{}, err
	}

	return batchRow.batch, nil
}

func (s PostgresBatchCatalogStore) ChangeQCStatus(
	ctx context.Context,
	input ChangeBatchQCStatusInput,
) (ChangeBatchQCStatusResult, error) {
	if s.db == nil {
		return ChangeBatchQCStatusResult{}, errors.New("database connection is required")
	}
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return ChangeBatchQCStatusResult{}, ErrBatchTransitionActorRequired
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return ChangeBatchQCStatusResult{}, ErrBatchTransitionReasonRequired
	}
	changedAt := input.ChangedAt
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return ChangeBatchQCStatusResult{}, fmt.Errorf("begin batch qc transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	batchRow, err := scanPostgresBatchRow(tx.QueryRowContext(ctx, selectPostgresBatchForUpdateSQL, strings.TrimSpace(input.BatchID)))
	if errors.Is(err, sql.ErrNoRows) {
		return ChangeBatchQCStatusResult{}, ErrBatchNotFound
	}
	if err != nil {
		return ChangeBatchQCStatusResult{}, err
	}

	updated, err := batchRow.batch.ChangeQCStatus(input.NextStatus, changedAt)
	if err != nil {
		return ChangeBatchQCStatusResult{}, err
	}
	businessRef := strings.TrimSpace(input.BusinessRef)
	if businessRef == "" {
		businessRef = batchRow.batch.ID
	}
	log, err := newBatchQCTransitionAuditLog(batchRow.batch, updated, actorID, reason, businessRef, input.RequestID, changedAt)
	if err != nil {
		return ChangeBatchQCStatusResult{}, err
	}
	if err := insertBatchQCAudit(ctx, tx, batchRow.persistedOrgID, log); err != nil {
		return ChangeBatchQCStatusResult{}, err
	}
	if err := updatePostgresBatchQCStatus(ctx, tx, batchRow.persistedID, updated, actorID); err != nil {
		return ChangeBatchQCStatusResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return ChangeBatchQCStatusResult{}, fmt.Errorf("commit batch qc transaction: %w", err)
	}
	committed = true

	return ChangeBatchQCStatusResult{
		Batch:      updated,
		Transition: batchQCTransitionFromAudit(log),
		AuditLogID: log.ID,
	}, nil
}

func (s PostgresBatchCatalogStore) ListQCTransitions(
	ctx context.Context,
	batchID string,
) ([]domain.BatchQCTransition, error) {
	batch, err := s.GetBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}

	logs, err := s.auditLog.List(ctx, audit.Query{
		Action:     batchQCTransitionAction,
		EntityType: batchQCTransitionEntityType,
		EntityID:   batch.ID,
		Limit:      100,
	})
	if err != nil {
		return nil, err
	}

	transitions := make([]domain.BatchQCTransition, 0, len(logs))
	for _, log := range logs {
		transitions = append(transitions, batchQCTransitionFromAudit(log))
	}

	return transitions, nil
}

func buildPostgresBatchListQuery(filter domain.BatchFilter) (string, []any) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)

	if sku := strings.ToUpper(strings.TrimSpace(filter.SKU)); sku != "" {
		args = append(args, sku)
		clauses = append(clauses, fmt.Sprintf("upper(item.sku) = $%d", len(args)))
	}
	if qcStatus := domain.NormalizeQCStatus(filter.QCStatus); qcStatus != "" {
		args = append(args, string(qcStatus))
		clauses = append(clauses, fmt.Sprintf("batch.qc_status = $%d", len(args)))
	}
	if status := domain.NormalizeBatchStatus(filter.Status); status != "" {
		args = append(args, string(status))
		clauses = append(clauses, fmt.Sprintf("batch.status = $%d", len(args)))
	}

	query := selectPostgresBatchBaseSQL
	if len(clauses) > 0 {
		query += "\nWHERE " + strings.Join(clauses, "\n  AND ")
	}
	query += selectPostgresBatchOrderSQL

	return query, args
}

func buildPostgresBatchLookupQuery() string {
	return selectPostgresBatchBaseSQL + `
WHERE lower(COALESCE(batch.batch_ref, batch.id::text)) = lower($1)
   OR batch.id::text = $1
LIMIT 1`
}

type postgresBatchRow struct {
	persistedID    string
	persistedOrgID string
	batch          domain.Batch
}

type postgresBatchRowScanner interface {
	Scan(dest ...any) error
}

func scanPostgresBatchRow(scanner postgresBatchRowScanner) (postgresBatchRow, error) {
	var (
		row        postgresBatchRow
		mfgDate    sql.NullTime
		expiryDate sql.NullTime
		qcStatus   string
		status     string
	)
	if err := scanner.Scan(
		&row.persistedID,
		&row.persistedOrgID,
		&row.batch.ID,
		&row.batch.OrgID,
		&row.batch.ItemID,
		&row.batch.SKU,
		&row.batch.ItemName,
		&row.batch.BatchNo,
		&row.batch.SupplierID,
		&mfgDate,
		&expiryDate,
		&qcStatus,
		&status,
		&row.batch.CreatedAt,
		&row.batch.UpdatedAt,
	); err != nil {
		return postgresBatchRow{}, err
	}
	if mfgDate.Valid {
		row.batch.MfgDate = mfgDate.Time
	}
	if expiryDate.Valid {
		row.batch.ExpiryDate = expiryDate.Time
	}
	row.batch.QCStatus = domain.QCStatus(qcStatus)
	row.batch.Status = domain.BatchStatus(status)

	batch, err := domain.NewBatch(domain.NewBatchInput{
		ID:         row.batch.ID,
		OrgID:      row.batch.OrgID,
		ItemID:     row.batch.ItemID,
		SKU:        row.batch.SKU,
		ItemName:   row.batch.ItemName,
		BatchNo:    row.batch.BatchNo,
		SupplierID: row.batch.SupplierID,
		MfgDate:    row.batch.MfgDate,
		ExpiryDate: row.batch.ExpiryDate,
		QCStatus:   row.batch.QCStatus,
		Status:     row.batch.Status,
		CreatedAt:  row.batch.CreatedAt,
		UpdatedAt:  row.batch.UpdatedAt,
	})
	if err != nil {
		return postgresBatchRow{}, err
	}
	row.batch = batch

	return row, nil
}

func insertBatchQCAudit(ctx context.Context, tx *sql.Tx, orgID string, log audit.Log) error {
	beforeJSON, err := batchCatalogJSONMap(log.BeforeData)
	if err != nil {
		return fmt.Errorf("encode batch qc audit before_data: %w", err)
	}
	afterJSON, err := batchCatalogJSONMap(log.AfterData)
	if err != nil {
		return fmt.Errorf("encode batch qc audit after_data: %w", err)
	}
	metadataJSON, err := requiredBatchCatalogJSONMap(log.Metadata)
	if err != nil {
		return fmt.Errorf("encode batch qc audit metadata: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		insertBatchQCAuditSQL,
		nullableUUID(log.ID),
		orgID,
		nullableUUID(log.ActorID),
		log.Action,
		log.EntityType,
		nullableUUID(log.EntityID),
		nullableText(log.RequestID),
		beforeJSON,
		afterJSON,
		metadataJSON,
		log.CreatedAt.UTC(),
		nullableText(log.ID),
		nullableText(log.OrgID),
		nullableText(log.ActorID),
		nullableText(log.EntityID),
	)
	if err != nil {
		return fmt.Errorf("insert batch qc audit: %w", err)
	}

	return nil
}

func updatePostgresBatchQCStatus(
	ctx context.Context,
	tx *sql.Tx,
	persistedID string,
	batch domain.Batch,
	actorID string,
) error {
	result, err := tx.ExecContext(
		ctx,
		updatePostgresBatchQCStatusSQL,
		persistedID,
		string(batch.QCStatus),
		batch.UpdatedAt.UTC(),
		nullableUUID(actorID),
		nullableText(actorID),
	)
	if err != nil {
		return fmt.Errorf("update batch qc status: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update batch qc status affected rows: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("update batch qc status affected %d rows, want 1", affected)
	}

	return nil
}

func batchCatalogJSONMap(value map[string]any) (any, error) {
	if value == nil {
		return nil, nil
	}

	return requiredBatchCatalogJSONMap(value)
}

func requiredBatchCatalogJSONMap(value map[string]any) (string, error) {
	if value == nil {
		value = map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

var _ BatchCatalogStore = PostgresBatchCatalogStore{}
