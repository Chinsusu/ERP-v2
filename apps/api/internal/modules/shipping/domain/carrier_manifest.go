package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type CarrierManifestStatus string

const ManifestStatusDraft CarrierManifestStatus = "draft"
const ManifestStatusReady CarrierManifestStatus = "ready"
const ManifestStatusScanning CarrierManifestStatus = "scanning"
const ManifestStatusCompleted CarrierManifestStatus = "completed"
const ManifestStatusException CarrierManifestStatus = "exception"

var ErrManifestRequiredField = errors.New("carrier manifest required field is missing")
var ErrManifestDuplicateShipment = errors.New("shipment already exists in carrier manifest")
var ErrManifestShipmentNotPacked = errors.New("shipment must be packed before adding to carrier manifest")
var ErrManifestAlreadyCompleted = errors.New("carrier manifest is already completed")

type CarrierManifest struct {
	ID            string
	CarrierCode   string
	CarrierName   string
	WarehouseID   string
	WarehouseCode string
	Date          string
	HandoverBatch string
	StagingZone   string
	Status        CarrierManifestStatus
	Owner         string
	Lines         []CarrierManifestLine
	CreatedAt     time.Time
}

type CarrierManifestLine struct {
	ID          string
	ShipmentID  string
	OrderNo     string
	TrackingNo  string
	PackageCode string
	StagingZone string
	Scanned     bool
}

type PackedShipment struct {
	ID            string
	OrderNo       string
	TrackingNo    string
	CarrierCode   string
	CarrierName   string
	WarehouseID   string
	WarehouseCode string
	PackageCode   string
	StagingZone   string
	Packed        bool
}

type CarrierManifestSummary struct {
	ExpectedCount int
	ScannedCount  int
	MissingCount  int
}

type CarrierManifestFilter struct {
	WarehouseID string
	Date        string
	CarrierCode string
	Status      CarrierManifestStatus
}

type NewCarrierManifestInput struct {
	ID            string
	CarrierCode   string
	CarrierName   string
	WarehouseID   string
	WarehouseCode string
	Date          string
	HandoverBatch string
	StagingZone   string
	Owner         string
	CreatedAt     time.Time
}

func NewCarrierManifest(input NewCarrierManifestInput) (CarrierManifest, error) {
	manifest := CarrierManifest{
		ID:            strings.TrimSpace(input.ID),
		CarrierCode:   strings.ToUpper(strings.TrimSpace(input.CarrierCode)),
		CarrierName:   strings.TrimSpace(input.CarrierName),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		Date:          strings.TrimSpace(input.Date),
		HandoverBatch: strings.TrimSpace(input.HandoverBatch),
		StagingZone:   strings.TrimSpace(input.StagingZone),
		Status:        ManifestStatusDraft,
		Owner:         strings.TrimSpace(input.Owner),
		CreatedAt:     input.CreatedAt,
	}
	if manifest.CarrierCode == "" || manifest.WarehouseID == "" || manifest.Date == "" {
		return CarrierManifest{}, ErrManifestRequiredField
	}
	if manifest.CarrierName == "" {
		manifest.CarrierName = manifest.CarrierCode
	}
	if manifest.WarehouseCode == "" {
		manifest.WarehouseCode = manifest.WarehouseID
	}
	if manifest.HandoverBatch == "" {
		manifest.HandoverBatch = "day"
	}
	if manifest.StagingZone == "" {
		manifest.StagingZone = "handover"
	}
	if manifest.Owner == "" {
		manifest.Owner = "Warehouse Lead"
	}
	if manifest.CreatedAt.IsZero() {
		manifest.CreatedAt = time.Now().UTC()
	}
	if manifest.ID == "" {
		manifest.ID = fmt.Sprintf("manifest-%s-%s-%s", strings.ToLower(manifest.WarehouseCode), strings.ToLower(manifest.CarrierCode), strings.ReplaceAll(manifest.Date, "-", ""))
	}

	return manifest, nil
}

func NewCarrierManifestFilter(
	warehouseID string,
	date string,
	carrierCode string,
	status CarrierManifestStatus,
) CarrierManifestFilter {
	return CarrierManifestFilter{
		WarehouseID: strings.TrimSpace(warehouseID),
		Date:        strings.TrimSpace(date),
		CarrierCode: strings.ToUpper(strings.TrimSpace(carrierCode)),
		Status:      NormalizeManifestStatus(status),
	}
}

func NormalizeManifestStatus(status CarrierManifestStatus) CarrierManifestStatus {
	switch status {
	case ManifestStatusDraft, ManifestStatusReady, ManifestStatusScanning, ManifestStatusCompleted, ManifestStatusException:
		return status
	default:
		return ""
	}
}

func (m CarrierManifest) AddShipment(shipment PackedShipment) (CarrierManifest, error) {
	if m.Status == ManifestStatusCompleted {
		return CarrierManifest{}, ErrManifestAlreadyCompleted
	}
	if !shipment.Packed {
		return CarrierManifest{}, ErrManifestShipmentNotPacked
	}
	for _, line := range m.Lines {
		if line.ShipmentID == strings.TrimSpace(shipment.ID) {
			return CarrierManifest{}, ErrManifestDuplicateShipment
		}
	}

	next := m.Clone()
	next.Lines = append(next.Lines, CarrierManifestLine{
		ID:          fmt.Sprintf("line-%s", strings.TrimSpace(shipment.ID)),
		ShipmentID:  strings.TrimSpace(shipment.ID),
		OrderNo:     strings.TrimSpace(shipment.OrderNo),
		TrackingNo:  strings.TrimSpace(shipment.TrackingNo),
		PackageCode: strings.TrimSpace(shipment.PackageCode),
		StagingZone: strings.TrimSpace(shipment.StagingZone),
	})
	if next.Status == ManifestStatusDraft {
		next.Status = ManifestStatusReady
	}

	return next, nil
}

func (m CarrierManifest) Summary() CarrierManifestSummary {
	summary := CarrierManifestSummary{ExpectedCount: len(m.Lines)}
	for _, line := range m.Lines {
		if line.Scanned {
			summary.ScannedCount++
		}
	}
	summary.MissingCount = summary.ExpectedCount - summary.ScannedCount
	if summary.MissingCount < 0 {
		summary.MissingCount = 0
	}

	return summary
}

func (m CarrierManifest) Clone() CarrierManifest {
	clone := m
	clone.Lines = append([]CarrierManifestLine(nil), m.Lines...)
	return clone
}

func SortCarrierManifests(rows []CarrierManifest) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if left.Date != right.Date {
			return left.Date > right.Date
		}
		if left.WarehouseCode != right.WarehouseCode {
			return left.WarehouseCode < right.WarehouseCode
		}
		if left.CarrierCode != right.CarrierCode {
			return left.CarrierCode < right.CarrierCode
		}

		return left.HandoverBatch < right.HandoverBatch
	})
}
