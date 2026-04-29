package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrReturnInspectionNotFound = errors.New("return inspection not found")
var ErrReturnAttachmentNotAllowed = errors.New("return attachment upload is not allowed")
var ErrReturnAttachmentStorageUnavailable = errors.New("return attachment storage is unavailable")

type ReturnAttachmentStore interface {
	FindReceiptByID(ctx context.Context, id string) (domain.ReturnReceipt, error)
	FindInspectionByID(ctx context.Context, id string) (domain.ReturnInspection, error)
	SaveAttachment(ctx context.Context, attachment domain.ReturnAttachment) error
}

type ReturnAttachmentObjectStore interface {
	PutObject(ctx context.Context, bucket string, key string, contentType string, size int64, body io.Reader) error
}

type UploadReturnAttachment struct {
	store         ReturnAttachmentStore
	objectStore   ReturnAttachmentObjectStore
	storageBucket string
	auditLog      audit.LogStore
	clock         func() time.Time
}

type UploadReturnAttachmentInput struct {
	ReceiptID     string
	InspectionID  string
	FileName      string
	MIMEType      string
	FileSizeBytes int64
	Content       io.Reader
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
		store:         store,
		storageBucket: domain.ReturnAttachmentStorageBucket,
		auditLog:      auditLog,
		clock:         func() time.Time { return time.Now().UTC() },
	}
}

func (uc UploadReturnAttachment) WithObjectStore(store ReturnAttachmentObjectStore) UploadReturnAttachment {
	uc.objectStore = store

	return uc
}

func (uc UploadReturnAttachment) WithStorageBucket(bucket string) UploadReturnAttachment {
	if strings.TrimSpace(bucket) != "" {
		uc.storageBucket = strings.TrimSpace(bucket)
	}

	return uc
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
		StorageBucket: uc.storageBucket,
		UploadedBy:    input.ActorID,
		Note:          input.Note,
		UploadedAt:    uc.clock(),
	})
	if err != nil {
		return ReturnAttachmentResult{}, err
	}
	if uc.objectStore == nil || input.Content == nil {
		return ReturnAttachmentResult{}, ErrReturnAttachmentStorageUnavailable
	}
	if err := uc.objectStore.PutObject(
		ctx,
		attachment.StorageBucket,
		attachment.StorageKey,
		attachment.MIMEType,
		attachment.FileSizeBytes,
		input.Content,
	); err != nil {
		return ReturnAttachmentResult{}, fmt.Errorf("%w: %v", ErrReturnAttachmentStorageUnavailable, err)
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
	service.objectStore = NewInMemoryReturnAttachmentObjectStore()
	service.clock = func() time.Time { return now.UTC() }

	return service
}

type InMemoryReturnAttachmentObject struct {
	Bucket      string
	Key         string
	ContentType string
	Size        int64
	Bytes       []byte
}

type InMemoryReturnAttachmentObjectStore struct {
	mu      sync.RWMutex
	objects map[string]InMemoryReturnAttachmentObject
}

func NewInMemoryReturnAttachmentObjectStore() *InMemoryReturnAttachmentObjectStore {
	return &InMemoryReturnAttachmentObjectStore{objects: make(map[string]InMemoryReturnAttachmentObject)}
}

func (store *InMemoryReturnAttachmentObjectStore) PutObject(
	_ context.Context,
	bucket string,
	key string,
	contentType string,
	size int64,
	body io.Reader,
) error {
	if store == nil || body == nil || strings.TrimSpace(bucket) == "" || strings.TrimSpace(key) == "" || size <= 0 {
		return ErrReturnAttachmentStorageUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(body, size+1))
	if err != nil {
		return err
	}
	if int64(len(data)) != size {
		return ErrReturnAttachmentStorageUnavailable
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.objects[bucket+"/"+key] = InMemoryReturnAttachmentObject{
		Bucket:      bucket,
		Key:         key,
		ContentType: contentType,
		Size:        size,
		Bytes:       append([]byte(nil), data...),
	}

	return nil
}

func (store *InMemoryReturnAttachmentObjectStore) Get(
	bucket string,
	key string,
) (InMemoryReturnAttachmentObject, bool) {
	if store == nil {
		return InMemoryReturnAttachmentObject{}, false
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	object, ok := store.objects[bucket+"/"+key]
	if !ok {
		return InMemoryReturnAttachmentObject{}, false
	}
	object.Bytes = append([]byte(nil), object.Bytes...)

	return object, true
}

func (store *InMemoryReturnAttachmentObjectStore) Len() int {
	if store == nil {
		return 0
	}
	store.mu.RLock()
	defer store.mu.RUnlock()

	return len(store.objects)
}
