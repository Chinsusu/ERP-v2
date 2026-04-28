package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrReturnInspectionNotFound = errors.New("return inspection not found")
var ErrReturnAttachmentNotAllowed = errors.New("return attachment upload is not allowed")

type ReturnAttachmentStore interface {
	FindReceiptByID(ctx context.Context, id string) (domain.ReturnReceipt, error)
	FindInspectionByID(ctx context.Context, id string) (domain.ReturnInspection, error)
	SaveAttachment(ctx context.Context, attachment domain.ReturnAttachment) error
}

type UploadReturnAttachment struct {
	store    ReturnAttachmentStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type UploadReturnAttachmentInput struct {
	ReceiptID     string
	InspectionID  string
	FileName      string
	MIMEType      string
	FileSizeBytes int64
	Note          string
	ActorID       string
	RequestID     string
}

type ReturnAttachmentResult struct {
	Attachment domain.ReturnAttachment
	AuditLogID string
}

func NewUploadReturnAttachment(store ReturnAttachmentStore, auditLog audit.LogStore) UploadReturnAttachment {
	return UploadReturnAttachment{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc UploadReturnAttachment) Execute(
	ctx context.Context,
	input UploadReturnAttachmentInput,
) (ReturnAttachmentResult, error) {
	if uc.store == nil {
		return ReturnAttachmentResult{}, errors.New("return attachment store is required")
	}
	if uc.auditLog == nil {
		return ReturnAttachmentResult{}, errors.New("audit log store is required")
	}

	receipt, err := uc.store.FindReceiptByID(ctx, input.ReceiptID)
	if err != nil {
		return ReturnAttachmentResult{}, err
	}
	if receipt.Status == domain.ReturnStatusPendingInspection {
		return ReturnAttachmentResult{}, ErrReturnAttachmentNotAllowed
	}

	inspection, err := uc.store.FindInspectionByID(ctx, input.InspectionID)
	if err != nil {
		return ReturnAttachmentResult{}, err
	}
	if inspection.ReceiptID != receipt.ID {
		return ReturnAttachmentResult{}, ErrReturnAttachmentNotAllowed
	}

	attachment, err := domain.NewReturnAttachment(domain.NewReturnAttachmentInput{
		ReceiptID:     receipt.ID,
		ReceiptNo:     receipt.ReceiptNo,
		InspectionID:  inspection.ID,
		FileName:      input.FileName,
		MIMEType:      input.MIMEType,
		FileSizeBytes: input.FileSizeBytes,
		UploadedBy:    input.ActorID,
		Note:          input.Note,
		UploadedAt:    uc.clock(),
	})
	if err != nil {
		return ReturnAttachmentResult{}, err
	}

	if err := uc.store.SaveAttachment(ctx, attachment); err != nil {
		return ReturnAttachmentResult{}, err
	}

	log, err := newReturnAttachmentAuditLog(input.ActorID, input.RequestID, attachment)
	if err != nil {
		return ReturnAttachmentResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return ReturnAttachmentResult{}, err
	}

	return ReturnAttachmentResult{Attachment: attachment, AuditLogID: log.ID}, nil
}

func newReturnAttachmentAuditLog(actorID string, requestID string, attachment domain.ReturnAttachment) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     "returns.inspection.attachment_uploaded",
		EntityType: "returns.return_attachment",
		EntityID:   attachment.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"receipt_id":      attachment.ReceiptID,
			"receipt_no":      attachment.ReceiptNo,
			"inspection_id":   attachment.InspectionID,
			"file_name":       attachment.FileName,
			"mime_type":       attachment.MIMEType,
			"file_size_bytes": attachment.FileSizeBytes,
			"storage_bucket":  attachment.StorageBucket,
			"storage_key":     attachment.StorageKey,
			"status":          attachment.Status,
		},
		Metadata: map[string]any{
			"source": "return attachment",
		},
		CreatedAt: attachment.UploadedAt,
	})
}

func NewPrototypeUploadReturnAttachmentAt(
	store ReturnAttachmentStore,
	auditLog audit.LogStore,
	now time.Time,
) UploadReturnAttachment {
	service := NewUploadReturnAttachment(store, auditLog)
	service.clock = func() time.Time { return now.UTC() }

	return service
}
