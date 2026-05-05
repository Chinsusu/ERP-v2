package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

type PostgresWarehouseDocumentStoreConfig struct {
	DefaultOrgID string
}

type PostgresStockTransferStore struct {
	db           *sql.DB
	defaultOrgID string
}

type PostgresWarehouseIssueStore struct {
	db           *sql.DB
	defaultOrgID string
}

func NewPostgresStockTransferStore(
	db *sql.DB,
	cfg PostgresWarehouseDocumentStoreConfig,
) PostgresStockTransferStore {
	return PostgresStockTransferStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

func NewPostgresWarehouseIssueStore(
	db *sql.DB,
	cfg PostgresWarehouseDocumentStoreConfig,
) PostgresWarehouseIssueStore {
	return PostgresWarehouseIssueStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const upsertPostgresStockTransferSQL = `
INSERT INTO inventory.stock_transfer_documents (
  org_id,
  transfer_ref,
  transfer_no,
  source_warehouse_ref,
  destination_warehouse_ref,
  status,
  transfer_payload,
  created_at,
  updated_at
) VALUES (
  $1::uuid,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7::jsonb,
  $8,
  $9
)
ON CONFLICT ON CONSTRAINT uq_stock_transfer_documents_org_ref
DO UPDATE SET
  transfer_no = EXCLUDED.transfer_no,
  source_warehouse_ref = EXCLUDED.source_warehouse_ref,
  destination_warehouse_ref = EXCLUDED.destination_warehouse_ref,
  status = EXCLUDED.status,
  transfer_payload = EXCLUDED.transfer_payload,
  updated_at = EXCLUDED.updated_at`

const selectPostgresStockTransferPayloadSQL = `
SELECT transfer_payload
FROM inventory.stock_transfer_documents
WHERE org_id = $1::uuid
  AND (lower(transfer_ref) = lower($2) OR lower(transfer_no) = lower($2))
LIMIT 1`

const selectPostgresStockTransferPayloadsSQL = `
SELECT transfer_payload
FROM inventory.stock_transfer_documents
WHERE org_id = $1::uuid
ORDER BY created_at DESC, transfer_no DESC`

const upsertPostgresWarehouseIssueSQL = `
INSERT INTO inventory.warehouse_issue_documents (
  org_id,
  issue_ref,
  issue_no,
  warehouse_ref,
  destination_type,
  destination_name,
  status,
  issue_payload,
  created_at,
  updated_at
) VALUES (
  $1::uuid,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8::jsonb,
  $9,
  $10
)
ON CONFLICT ON CONSTRAINT uq_warehouse_issue_documents_org_ref
DO UPDATE SET
  issue_no = EXCLUDED.issue_no,
  warehouse_ref = EXCLUDED.warehouse_ref,
  destination_type = EXCLUDED.destination_type,
  destination_name = EXCLUDED.destination_name,
  status = EXCLUDED.status,
  issue_payload = EXCLUDED.issue_payload,
  updated_at = EXCLUDED.updated_at`

const selectPostgresWarehouseIssuePayloadSQL = `
SELECT issue_payload
FROM inventory.warehouse_issue_documents
WHERE org_id = $1::uuid
  AND (lower(issue_ref) = lower($2) OR lower(issue_no) = lower($2))
LIMIT 1`

const selectPostgresWarehouseIssuePayloadsSQL = `
SELECT issue_payload
FROM inventory.warehouse_issue_documents
WHERE org_id = $1::uuid
ORDER BY created_at DESC, issue_no DESC`

func (s PostgresStockTransferStore) ListStockTransfers(ctx context.Context) ([]domain.StockTransfer, error) {
	if s.db == nil {
		return nil, errors.New("stock transfer postgres store is required")
	}
	orgID, err := resolveWarehouseDocumentOrgID(ctx, s.db, s.defaultOrgID, "")
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, selectPostgresStockTransferPayloadsSQL, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transfers := make([]domain.StockTransfer, 0)
	for rows.Next() {
		transfer, err := scanPostgresStockTransferPayload(rows)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortStockTransfers(transfers)

	return transfers, nil
}

func (s PostgresStockTransferStore) FindStockTransferByID(ctx context.Context, id string) (domain.StockTransfer, error) {
	if s.db == nil {
		return domain.StockTransfer{}, errors.New("stock transfer postgres store is required")
	}
	orgID, err := resolveWarehouseDocumentOrgID(ctx, s.db, s.defaultOrgID, "")
	if err != nil {
		return domain.StockTransfer{}, err
	}
	transfer, err := scanPostgresStockTransferPayload(
		s.db.QueryRowContext(ctx, selectPostgresStockTransferPayloadSQL, orgID, strings.TrimSpace(id)),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.StockTransfer{}, ErrStockTransferNotFound
		}
		return domain.StockTransfer{}, err
	}

	return transfer, nil
}

func (s PostgresStockTransferStore) SaveStockTransfer(ctx context.Context, transfer domain.StockTransfer) error {
	if s.db == nil {
		return errors.New("stock transfer postgres store is required")
	}
	if err := transfer.Validate(); err != nil {
		return err
	}
	orgID, err := resolveWarehouseDocumentOrgID(ctx, s.db, s.defaultOrgID, transfer.OrgID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(transfer)
	if err != nil {
		return fmt.Errorf("encode stock transfer payload: %w", err)
	}
	_, err = s.db.ExecContext(
		ctx,
		upsertPostgresStockTransferSQL,
		orgID,
		transfer.ID,
		transfer.TransferNo,
		warehouseDocumentRef(transfer.SourceWarehouseID, transfer.SourceWarehouseCode),
		warehouseDocumentRef(transfer.DestinationWarehouseID, transfer.DestinationWarehouseCode),
		string(transfer.Status),
		string(payload),
		transfer.CreatedAt.UTC(),
		transfer.UpdatedAt.UTC(),
	)

	return err
}

func (s PostgresWarehouseIssueStore) ListWarehouseIssues(ctx context.Context) ([]domain.WarehouseIssue, error) {
	if s.db == nil {
		return nil, errors.New("warehouse issue postgres store is required")
	}
	orgID, err := resolveWarehouseDocumentOrgID(ctx, s.db, s.defaultOrgID, "")
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, selectPostgresWarehouseIssuePayloadsSQL, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]domain.WarehouseIssue, 0)
	for rows.Next() {
		issue, err := scanPostgresWarehouseIssuePayload(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortWarehouseIssues(issues)

	return issues, nil
}

func (s PostgresWarehouseIssueStore) FindWarehouseIssueByID(ctx context.Context, id string) (domain.WarehouseIssue, error) {
	if s.db == nil {
		return domain.WarehouseIssue{}, errors.New("warehouse issue postgres store is required")
	}
	orgID, err := resolveWarehouseDocumentOrgID(ctx, s.db, s.defaultOrgID, "")
	if err != nil {
		return domain.WarehouseIssue{}, err
	}
	issue, err := scanPostgresWarehouseIssuePayload(
		s.db.QueryRowContext(ctx, selectPostgresWarehouseIssuePayloadSQL, orgID, strings.TrimSpace(id)),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.WarehouseIssue{}, ErrWarehouseIssueNotFound
		}
		return domain.WarehouseIssue{}, err
	}

	return issue, nil
}

func (s PostgresWarehouseIssueStore) SaveWarehouseIssue(ctx context.Context, issue domain.WarehouseIssue) error {
	if s.db == nil {
		return errors.New("warehouse issue postgres store is required")
	}
	if err := issue.Validate(); err != nil {
		return err
	}
	orgID, err := resolveWarehouseDocumentOrgID(ctx, s.db, s.defaultOrgID, issue.OrgID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("encode warehouse issue payload: %w", err)
	}
	_, err = s.db.ExecContext(
		ctx,
		upsertPostgresWarehouseIssueSQL,
		orgID,
		issue.ID,
		issue.IssueNo,
		warehouseDocumentRef(issue.WarehouseID, issue.WarehouseCode),
		issue.DestinationType,
		issue.DestinationName,
		string(issue.Status),
		string(payload),
		issue.CreatedAt.UTC(),
		issue.UpdatedAt.UTC(),
	)

	return err
}

type warehouseDocumentPayloadScanner interface {
	Scan(dest ...any) error
}

func scanPostgresStockTransferPayload(scanner warehouseDocumentPayloadScanner) (domain.StockTransfer, error) {
	var raw []byte
	if err := scanner.Scan(&raw); err != nil {
		return domain.StockTransfer{}, err
	}
	var transfer domain.StockTransfer
	if err := json.Unmarshal(raw, &transfer); err != nil {
		return domain.StockTransfer{}, err
	}
	if err := transfer.Validate(); err != nil {
		return domain.StockTransfer{}, err
	}

	return transfer, nil
}

func scanPostgresWarehouseIssuePayload(scanner warehouseDocumentPayloadScanner) (domain.WarehouseIssue, error) {
	var raw []byte
	if err := scanner.Scan(&raw); err != nil {
		return domain.WarehouseIssue{}, err
	}
	var issue domain.WarehouseIssue
	if err := json.Unmarshal(raw, &issue); err != nil {
		return domain.WarehouseIssue{}, err
	}
	if err := issue.Validate(); err != nil {
		return domain.WarehouseIssue{}, err
	}

	return issue, nil
}

func resolveWarehouseDocumentOrgID(ctx context.Context, db *sql.DB, defaultOrgID string, orgRef string) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isWarehouseDocumentUUIDText(orgRef) {
		return orgRef, nil
	}
	if isWarehouseDocumentUUIDText(defaultOrgID) {
		return defaultOrgID, nil
	}
	if orgRef == "" {
		orgRef = "org-my-pham"
	}
	if orgRef != "" {
		var orgID string
		err := db.QueryRowContext(ctx, `SELECT id::text FROM core.organizations WHERE code = $1 LIMIT 1`, orgRef).Scan(&orgID)
		if err == nil && isWarehouseDocumentUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}

	return "", fmt.Errorf("warehouse document org %q cannot be resolved", orgRef)
}

func warehouseDocumentRef(id string, code string) string {
	if trimmed := strings.TrimSpace(id); trimmed != "" {
		return trimmed
	}

	return strings.TrimSpace(code)
}

func isWarehouseDocumentUUIDText(value string) bool {
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
