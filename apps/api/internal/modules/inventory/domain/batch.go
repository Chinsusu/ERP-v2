package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
)

var ErrBatchRequiredField = errors.New("batch required field is missing")
var ErrBatchInvalidQCStatus = errors.New("batch qc status is invalid")
var ErrBatchInvalidStatus = errors.New("batch status is invalid")
var ErrBatchInvalidExpiry = errors.New("batch expiry date is invalid")
var ErrBatchInvalidQCTransition = errors.New("batch qc status transition is invalid")

type QCStatus string
type BatchStatus string

const (
	QCStatusHold           QCStatus = "hold"
	QCStatusPass           QCStatus = "pass"
	QCStatusFail           QCStatus = "fail"
	QCStatusQuarantine     QCStatus = "quarantine"
	QCStatusRetestRequired QCStatus = "retest_required"

	BatchStatusActive   BatchStatus = "active"
	BatchStatusInactive BatchStatus = "inactive"
	BatchStatusBlocked  BatchStatus = "blocked"
)

type Batch struct {
	ID         string
	OrgID      string
	ItemID     string
	SKU        string
	ItemName   string
	BatchNo    string
	SupplierID string
	MfgDate    time.Time
	ExpiryDate time.Time
	QCStatus   QCStatus
	Status     BatchStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type NewBatchInput struct {
	ID         string
	OrgID      string
	ItemID     string
	SKU        string
	ItemName   string
	BatchNo    string
	SupplierID string
	MfgDate    time.Time
	ExpiryDate time.Time
	QCStatus   QCStatus
	Status     BatchStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type BatchFilter struct {
	SKU      string
	QCStatus QCStatus
	Status   BatchStatus
}

func NewBatch(input NewBatchInput) (Batch, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	qcStatus := NormalizeQCStatus(input.QCStatus)
	if qcStatus == "" {
		qcStatus = QCStatusHold
	}
	status := NormalizeBatchStatus(input.Status)
	if status == "" {
		status = BatchStatusActive
	}

	batch := Batch{
		ID:         strings.TrimSpace(input.ID),
		OrgID:      strings.TrimSpace(input.OrgID),
		ItemID:     strings.TrimSpace(input.ItemID),
		SKU:        strings.ToUpper(strings.TrimSpace(input.SKU)),
		ItemName:   strings.TrimSpace(input.ItemName),
		BatchNo:    NormalizeBatchNo(input.BatchNo),
		SupplierID: strings.TrimSpace(input.SupplierID),
		MfgDate:    dateOnly(input.MfgDate),
		ExpiryDate: dateOnly(input.ExpiryDate),
		QCStatus:   qcStatus,
		Status:     status,
		CreatedAt:  createdAt.UTC(),
		UpdatedAt:  updatedAt.UTC(),
	}
	if err := batch.Validate(); err != nil {
		return Batch{}, err
	}

	return batch, nil
}

func NewBatchFilter(sku string, qcStatus QCStatus, status BatchStatus) BatchFilter {
	return BatchFilter{
		SKU:      strings.ToUpper(strings.TrimSpace(sku)),
		QCStatus: NormalizeQCStatus(qcStatus),
		Status:   NormalizeBatchStatus(status),
	}
}

func (f BatchFilter) Matches(batch Batch) bool {
	if f.SKU != "" && batch.SKU != f.SKU {
		return false
	}
	if f.QCStatus != "" && batch.QCStatus != f.QCStatus {
		return false
	}
	if f.Status != "" && batch.Status != f.Status {
		return false
	}

	return true
}

func (b Batch) Validate() error {
	if strings.TrimSpace(b.ID) == "" ||
		strings.TrimSpace(b.OrgID) == "" ||
		strings.TrimSpace(b.ItemID) == "" ||
		strings.TrimSpace(b.SKU) == "" ||
		strings.TrimSpace(b.BatchNo) == "" {
		return ErrBatchRequiredField
	}
	if !IsValidQCStatus(b.QCStatus) {
		return ErrBatchInvalidQCStatus
	}
	if !IsValidBatchStatus(b.Status) {
		return ErrBatchInvalidStatus
	}
	if !b.MfgDate.IsZero() && !b.ExpiryDate.IsZero() && b.ExpiryDate.Before(b.MfgDate) {
		return ErrBatchInvalidExpiry
	}

	return nil
}

func (b Batch) ChangeQCStatus(next QCStatus, updatedAt time.Time) (Batch, error) {
	next = NormalizeQCStatus(next)
	if !IsValidQCStatus(next) {
		return Batch{}, ErrBatchInvalidQCStatus
	}
	if !CanTransitionQCStatus(b.QCStatus, next) {
		return Batch{}, ErrBatchInvalidQCTransition
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated := b
	updated.QCStatus = next
	updated.UpdatedAt = updatedAt.UTC()

	return updated, nil
}

func (b Batch) IsExpired(asOf time.Time) bool {
	if b.ExpiryDate.IsZero() {
		return false
	}
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	return b.ExpiryDate.Before(dateOnly(asOf))
}

func (b Batch) IsAvailableForInventory(asOf time.Time) bool {
	return b.Status == BatchStatusActive && b.QCStatus == QCStatusPass && !b.IsExpired(asOf)
}

func (b Batch) BlocksAvailableStock(asOf time.Time) bool {
	return !b.IsAvailableForInventory(asOf)
}

func CanTransitionQCStatus(current QCStatus, next QCStatus) bool {
	current = NormalizeQCStatus(current)
	next = NormalizeQCStatus(next)
	if current == next {
		return true
	}

	switch current {
	case QCStatusHold, QCStatusQuarantine, QCStatusRetestRequired:
		switch next {
		case QCStatusPass, QCStatusFail, QCStatusQuarantine, QCStatusRetestRequired:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func NormalizeQCStatus(value QCStatus) QCStatus {
	return QCStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeBatchStatus(value BatchStatus) BatchStatus {
	return BatchStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeBatchNo(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func IsValidQCStatus(value QCStatus) bool {
	switch NormalizeQCStatus(value) {
	case QCStatusHold, QCStatusPass, QCStatusFail, QCStatusQuarantine, QCStatusRetestRequired:
		return true
	default:
		return false
	}
}

func IsValidBatchStatus(value BatchStatus) bool {
	switch NormalizeBatchStatus(value) {
	case BatchStatusActive, BatchStatusInactive, BatchStatusBlocked:
		return true
	default:
		return false
	}
}

func SortBatches(batches []Batch) {
	sort.Slice(batches, func(i int, j int) bool {
		left := batches[i]
		right := batches[j]
		if left.SKU != right.SKU {
			return left.SKU < right.SKU
		}
		if left.ExpiryDate.IsZero() != right.ExpiryDate.IsZero() {
			return !left.ExpiryDate.IsZero()
		}
		if !left.ExpiryDate.Equal(right.ExpiryDate) {
			return left.ExpiryDate.Before(right.ExpiryDate)
		}

		return left.BatchNo < right.BatchNo
	})
}

func dateOnly(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}

	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}
