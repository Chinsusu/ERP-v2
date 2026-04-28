package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const ReturnAttachmentMaxFileSizeBytes int64 = 25 * 1024 * 1024
const ReturnAttachmentStorageBucket = "erp-return-attachments-dev"
const ReturnAttachmentStatusActive = "active"

var ErrReturnAttachmentRequiredField = errors.New("return attachment required field is missing")
var ErrReturnAttachmentInvalidFileType = errors.New("return attachment file type is invalid")
var ErrReturnAttachmentInvalidFileSize = errors.New("return attachment file size is invalid")

var attachmentFileNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type ReturnAttachment struct {
	ID            string
	ReceiptID     string
	ReceiptNo     string
	InspectionID  string
	FileName      string
	FileExt       string
	MIMEType      string
	FileSizeBytes int64
	StorageBucket string
	StorageKey    string
	Status        string
	UploadedBy    string
	Note          string
	UploadedAt    time.Time
}

type NewReturnAttachmentInput struct {
	ID            string
	ReceiptID     string
	ReceiptNo     string
	InspectionID  string
	FileName      string
	MIMEType      string
	FileSizeBytes int64
	UploadedBy    string
	Note          string
	UploadedAt    time.Time
}

func NewReturnAttachment(input NewReturnAttachmentInput) (ReturnAttachment, error) {
	receiptID := strings.TrimSpace(input.ReceiptID)
	inspectionID := strings.TrimSpace(input.InspectionID)
	uploadedBy := strings.TrimSpace(input.UploadedBy)
	fileName := safeAttachmentFileName(input.FileName)
	if receiptID == "" || inspectionID == "" || uploadedBy == "" || fileName == "" {
		return ReturnAttachment{}, ErrReturnAttachmentRequiredField
	}

	mimeType := normalizeReturnAttachmentMIMEType(input.MIMEType)
	if mimeType == "" {
		return ReturnAttachment{}, ErrReturnAttachmentInvalidFileType
	}
	if input.FileSizeBytes <= 0 || input.FileSizeBytes > ReturnAttachmentMaxFileSizeBytes {
		return ReturnAttachment{}, ErrReturnAttachmentInvalidFileSize
	}

	uploadedAt := input.UploadedAt
	if uploadedAt.IsZero() {
		uploadedAt = time.Now().UTC()
	}

	attachment := ReturnAttachment{
		ID:            strings.TrimSpace(input.ID),
		ReceiptID:     receiptID,
		ReceiptNo:     strings.TrimSpace(input.ReceiptNo),
		InspectionID:  inspectionID,
		FileName:      fileName,
		FileExt:       strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), "."),
		MIMEType:      mimeType,
		FileSizeBytes: input.FileSizeBytes,
		StorageBucket: ReturnAttachmentStorageBucket,
		StorageKey:    fmt.Sprintf("returns/%s/inspections/%s/%s", strings.ToLower(receiptID), strings.ToLower(inspectionID), fileName),
		Status:        ReturnAttachmentStatusActive,
		UploadedBy:    uploadedBy,
		Note:          strings.TrimSpace(input.Note),
		UploadedAt:    uploadedAt.UTC(),
	}
	if attachment.ID == "" {
		attachment.ID = fmt.Sprintf(
			"attach-%s-%s-%s",
			strings.ToLower(receiptID),
			strings.ToLower(inspectionID),
			strings.TrimSuffix(strings.ToLower(fileName), strings.ToLower(filepath.Ext(fileName))),
		)
	}

	return attachment, nil
}

func (attachment ReturnAttachment) Clone() ReturnAttachment {
	return attachment
}

func normalizeReturnAttachmentMIMEType(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/png", "image/webp", "video/mp4", "video/quicktime":
		return strings.ToLower(strings.TrimSpace(mimeType))
	default:
		return ""
	}
}

func safeAttachmentFileName(fileName string) string {
	baseName := filepath.Base(strings.TrimSpace(fileName))
	if baseName == "." || baseName == string(filepath.Separator) {
		return ""
	}

	return strings.Trim(attachmentFileNameSanitizer.ReplaceAllString(baseName, "-"), ".-_")
}
