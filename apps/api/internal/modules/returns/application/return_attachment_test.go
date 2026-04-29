package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestUploadReturnAttachmentRecordsMetadataAndAudit(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	inspectReturnForAttachment(t, store, auditStore)
	objectStore := NewInMemoryReturnAttachmentObjectStore()
	service := NewPrototypeUploadReturnAttachmentAt(
		store,
		auditStore,
		time.Date(2026, 4, 26, 11, 20, 0, 0, time.UTC),
	).WithObjectStore(objectStore)

	result, err := service.Execute(context.Background(), UploadReturnAttachmentInput{
		ReceiptID:     "rr-260426-0001",
		InspectionID:  "inspect-rr-260426-0001-intact",
		FileName:      "return-photo.png",
		MIMEType:      "image/png",
		FileSizeBytes: 2048,
		Content:       strings.NewReader(strings.Repeat("a", 2048)),
		Note:          "seal intact evidence",
		ActorID:       "user-return-inspector",
		RequestID:     "req-return-attachment",
	})
	if err != nil {
		t.Fatalf("upload return attachment: %v", err)
	}

	if result.Attachment.StorageBucket != domain.ReturnAttachmentStorageBucket {
		t.Fatalf("storage bucket = %q, want default return attachment bucket", result.Attachment.StorageBucket)
	}
	if result.Attachment.InspectionID != "inspect-rr-260426-0001-intact" ||
		result.Attachment.MIMEType != "image/png" ||
		result.Attachment.FileSizeBytes != 2048 {
		t.Fatalf("attachment = %+v, want inspection-linked png metadata", result.Attachment)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}
	object, ok := objectStore.Get(result.Attachment.StorageBucket, result.Attachment.StorageKey)
	if !ok || object.Size != 2048 || len(object.Bytes) != 2048 {
		t.Fatalf("stored object = %+v ok=%t, want 2048 bytes in object storage", object, ok)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "returns.inspection.attachment_uploaded"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].AfterData["inspection_id"] != "inspect-rr-260426-0001-intact" ||
		logs[0].AfterData["file_name"] != "return-photo.png" {
		t.Fatalf("audit after data = %+v, want inspection and file metadata", logs[0].AfterData)
	}
	if _, ok := logs[0].AfterData["file_content"]; ok {
		t.Fatalf("audit after data = %+v, must not contain file content", logs[0].AfterData)
	}
}

func TestUploadReturnAttachmentRejectsPendingInvalidAndMismatchedInputs(t *testing.T) {
	store := NewPrototypeReturnReceiptStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewPrototypeUploadReturnAttachmentAt(
		store,
		auditStore,
		time.Date(2026, 4, 26, 11, 25, 0, 0, time.UTC),
	)

	_, err := service.Execute(context.Background(), UploadReturnAttachmentInput{
		ReceiptID:     "rr-260426-0001",
		InspectionID:  "inspect-rr-260426-0001-intact",
		FileName:      "return-photo.png",
		MIMEType:      "image/png",
		FileSizeBytes: 2048,
		Content:       strings.NewReader(strings.Repeat("a", 2048)),
		ActorID:       "user-return-inspector",
	})
	if !errors.Is(err, ErrReturnAttachmentNotAllowed) {
		t.Fatalf("error = %v, want attachment not allowed while pending", err)
	}

	inspectReturnForAttachment(t, store, auditStore)
	_, err = service.Execute(context.Background(), UploadReturnAttachmentInput{
		ReceiptID:     "rr-260426-0001",
		InspectionID:  "missing-inspection",
		FileName:      "return-photo.png",
		MIMEType:      "image/png",
		FileSizeBytes: 2048,
		Content:       strings.NewReader(strings.Repeat("a", 2048)),
		ActorID:       "user-return-inspector",
	})
	if !errors.Is(err, ErrReturnInspectionNotFound) {
		t.Fatalf("error = %v, want inspection not found", err)
	}

	_, err = service.Execute(context.Background(), UploadReturnAttachmentInput{
		ReceiptID:     "rr-260426-0001",
		InspectionID:  "inspect-rr-260426-0001-intact",
		FileName:      "return-photo.exe",
		MIMEType:      "application/octet-stream",
		FileSizeBytes: 2048,
		Content:       strings.NewReader(strings.Repeat("a", 2048)),
		ActorID:       "user-return-inspector",
	})
	if !errors.Is(err, domain.ErrReturnAttachmentInvalidFileType) {
		t.Fatalf("error = %v, want invalid file type", err)
	}
}

func inspectReturnForAttachment(t *testing.T, store *PrototypeReturnReceiptStore, auditStore audit.LogStore) {
	t.Helper()

	_, err := NewPrototypeInspectReturnAt(
		store,
		auditStore,
		time.Date(2026, 4, 26, 11, 0, 0, 0, time.UTC),
	).Execute(context.Background(), InspectReturnInput{
		ReceiptID:   "rr-260426-0001",
		Condition:   "intact",
		Disposition: "reusable",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-inspect",
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}
}
