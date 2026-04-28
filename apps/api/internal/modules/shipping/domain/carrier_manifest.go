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
const ManifestStatusHandedOver CarrierManifestStatus = "handed_over"
const ManifestStatusException CarrierManifestStatus = "exception"
const ManifestStatusCancelled CarrierManifestStatus = "cancelled"

var ErrManifestRequiredField = errors.New("carrier manifest required field is missing")
var ErrManifestDuplicateShipment = errors.New("shipment already exists in carrier manifest")
var ErrManifestShipmentNotFound = errors.New("shipment was not found in carrier manifest")
var ErrManifestShipmentNotPacked = errors.New("shipment must be packed before adding to carrier manifest")
var ErrManifestCarrierMismatch = errors.New("shipment carrier does not match carrier manifest")
var ErrManifestAlreadyCompleted = errors.New("carrier manifest is already completed")
var ErrManifestInvalidTransition = errors.New("carrier manifest status transition is invalid")
var ErrManifestScanCodeRequired = errors.New("manifest scan code is required")
var ErrManifestScanNotFound = errors.New("manifest scan code was not found")
var ErrManifestScanInvalidState = errors.New("manifest cannot accept scan in current state")
var ErrManifestScanDuplicate = errors.New("manifest line is already scanned")

type CarrierManifestScanResultCode string

const ScanResultMatched CarrierManifestScanResultCode = "MATCHED"
const ScanResultNotFound CarrierManifestScanResultCode = "NOT_FOUND"
const ScanResultInvalidState CarrierManifestScanResultCode = "INVALID_STATE"
const ScanResultManifestMismatch CarrierManifestScanResultCode = "MANIFEST_MISMATCH"
const ScanResultDuplicate CarrierManifestScanResultCode = "DUPLICATE_SCAN"

type CarrierManifest struct {
	ID               string
	CarrierCode      string
	CarrierName      string
	WarehouseID      string
	WarehouseCode    string
	Date             string
	HandoverBatch    string
	StagingZone      string
	HandoverZoneID   string
	HandoverZoneCode string
	HandoverBinID    string
	HandoverBinCode  string
	Status           CarrierManifestStatus
	Owner            string
	Lines            []CarrierManifestLine
	CreatedAt        time.Time
}

type CarrierManifestLine struct {
	ID               string
	ShipmentID       string
	OrderNo          string
	TrackingNo       string
	PackageCode      string
	StagingZone      string
	HandoverZoneID   string
	HandoverZoneCode string
	HandoverBinID    string
	HandoverBinCode  string
	Scanned          bool
}

type PackedShipment struct {
	ID               string
	OrderNo          string
	TrackingNo       string
	CarrierCode      string
	CarrierName      string
	WarehouseID      string
	WarehouseCode    string
	PackageCode      string
	StagingZone      string
	HandoverZoneID   string
	HandoverZoneCode string
	HandoverBinID    string
	HandoverBinCode  string
	Packed           bool
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
	ID               string
	CarrierCode      string
	CarrierName      string
	WarehouseID      string
	WarehouseCode    string
	Date             string
	HandoverBatch    string
	StagingZone      string
	HandoverZoneID   string
	HandoverZoneCode string
	HandoverBinID    string
	HandoverBinCode  string
	Owner            string
	CreatedAt        time.Time
}

func NewCarrierManifest(input NewCarrierManifestInput) (CarrierManifest, error) {
	manifest := CarrierManifest{
		ID:               strings.TrimSpace(input.ID),
		CarrierCode:      strings.ToUpper(strings.TrimSpace(input.CarrierCode)),
		CarrierName:      strings.TrimSpace(input.CarrierName),
		WarehouseID:      strings.TrimSpace(input.WarehouseID),
		WarehouseCode:    strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		Date:             strings.TrimSpace(input.Date),
		HandoverBatch:    strings.TrimSpace(input.HandoverBatch),
		StagingZone:      strings.TrimSpace(input.StagingZone),
		HandoverZoneID:   strings.TrimSpace(input.HandoverZoneID),
		HandoverZoneCode: normalizeManifestLocationCode(input.HandoverZoneCode),
		HandoverBinID:    strings.TrimSpace(input.HandoverBinID),
		HandoverBinCode:  normalizeManifestLocationCode(input.HandoverBinCode),
		Status:           ManifestStatusDraft,
		Owner:            strings.TrimSpace(input.Owner),
		CreatedAt:        input.CreatedAt,
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
	if manifest.HandoverZoneCode == "" {
		manifest.HandoverZoneCode = normalizeManifestLocationCode(manifest.StagingZone)
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
	case ManifestStatusDraft,
		ManifestStatusReady,
		ManifestStatusScanning,
		ManifestStatusCompleted,
		ManifestStatusHandedOver,
		ManifestStatusException,
		ManifestStatusCancelled:
		return status
	default:
		return ""
	}
}

func (m CarrierManifest) AddShipment(shipment PackedShipment) (CarrierManifest, error) {
	if isCarrierManifestClosed(m.Status) {
		return CarrierManifest{}, ErrManifestAlreadyCompleted
	}
	if m.Status != ManifestStatusDraft && m.Status != ManifestStatusReady {
		return CarrierManifest{}, ErrManifestInvalidTransition
	}
	if !shipment.Packed {
		return CarrierManifest{}, ErrManifestShipmentNotPacked
	}
	if shipmentCarrierCode := strings.ToUpper(strings.TrimSpace(shipment.CarrierCode)); shipmentCarrierCode != "" && shipmentCarrierCode != m.CarrierCode {
		return CarrierManifest{}, ErrManifestCarrierMismatch
	}
	for _, line := range m.Lines {
		if line.ShipmentID == strings.TrimSpace(shipment.ID) {
			return CarrierManifest{}, ErrManifestDuplicateShipment
		}
	}

	next := m.Clone()
	next.Lines = append(next.Lines, CarrierManifestLine{
		ID:               fmt.Sprintf("line-%s", strings.TrimSpace(shipment.ID)),
		ShipmentID:       strings.TrimSpace(shipment.ID),
		OrderNo:          strings.TrimSpace(shipment.OrderNo),
		TrackingNo:       strings.TrimSpace(shipment.TrackingNo),
		PackageCode:      strings.TrimSpace(shipment.PackageCode),
		StagingZone:      firstNonEmpty(shipment.StagingZone, m.StagingZone),
		HandoverZoneID:   firstNonEmpty(shipment.HandoverZoneID, m.HandoverZoneID),
		HandoverZoneCode: firstNonEmpty(normalizeManifestLocationCode(shipment.HandoverZoneCode), m.HandoverZoneCode),
		HandoverBinID:    firstNonEmpty(shipment.HandoverBinID, m.HandoverBinID),
		HandoverBinCode:  firstNonEmpty(normalizeManifestLocationCode(shipment.HandoverBinCode), m.HandoverBinCode),
	})

	return next, nil
}

func (m CarrierManifest) RemoveShipment(shipmentID string) (CarrierManifest, error) {
	shipmentID = strings.TrimSpace(shipmentID)
	if shipmentID == "" {
		return CarrierManifest{}, ErrManifestRequiredField
	}
	if isCarrierManifestClosed(m.Status) {
		return CarrierManifest{}, ErrManifestAlreadyCompleted
	}
	if m.Status != ManifestStatusDraft && m.Status != ManifestStatusReady {
		return CarrierManifest{}, ErrManifestInvalidTransition
	}

	next := m.Clone()
	lines := make([]CarrierManifestLine, 0, len(next.Lines))
	removed := false
	for _, line := range next.Lines {
		if strings.TrimSpace(line.ShipmentID) == shipmentID {
			removed = true
			continue
		}
		lines = append(lines, line)
	}
	if !removed {
		return CarrierManifest{}, ErrManifestShipmentNotFound
	}
	next.Lines = lines
	if len(next.Lines) == 0 {
		next.Status = ManifestStatusDraft
	}

	return next, nil
}

func (m CarrierManifest) MarkReadyToScan() (CarrierManifest, error) {
	if m.Status == ManifestStatusReady || m.Status == ManifestStatusScanning {
		return m.Clone(), nil
	}
	if m.Status != ManifestStatusDraft {
		return CarrierManifest{}, ErrManifestInvalidTransition
	}
	if len(m.Lines) == 0 {
		return CarrierManifest{}, ErrManifestRequiredField
	}

	next := m.Clone()
	next.Status = ManifestStatusReady

	return next, nil
}

func (m CarrierManifest) Cancel() (CarrierManifest, error) {
	if m.Status == ManifestStatusCancelled {
		return m.Clone(), nil
	}
	if m.Status == ManifestStatusCompleted || m.Status == ManifestStatusHandedOver {
		return CarrierManifest{}, ErrManifestAlreadyCompleted
	}

	next := m.Clone()
	next.Status = ManifestStatusCancelled

	return next, nil
}

func (m CarrierManifest) MarkLineScanned(code string) (CarrierManifest, CarrierManifestLine, error) {
	normalizedCode := NormalizeManifestScanCode(code)
	if normalizedCode == "" {
		return CarrierManifest{}, CarrierManifestLine{}, ErrManifestScanCodeRequired
	}
	if m.Status != ManifestStatusReady && m.Status != ManifestStatusScanning {
		return CarrierManifest{}, CarrierManifestLine{}, ErrManifestScanInvalidState
	}

	lineIndex, line, ok := m.FindLineByScanCode(normalizedCode)
	if !ok {
		return CarrierManifest{}, CarrierManifestLine{}, ErrManifestScanNotFound
	}
	if line.Scanned {
		return CarrierManifest{}, line, ErrManifestScanDuplicate
	}

	next := m.Clone()
	next.Lines[lineIndex].Scanned = true
	if next.Status == ManifestStatusReady {
		next.Status = ManifestStatusScanning
	}

	return next, next.Lines[lineIndex], nil
}

func isCarrierManifestClosed(status CarrierManifestStatus) bool {
	return status == ManifestStatusCompleted ||
		status == ManifestStatusHandedOver ||
		status == ManifestStatusCancelled
}

func (m CarrierManifest) FindLineByScanCode(code string) (int, CarrierManifestLine, bool) {
	normalizedCode := NormalizeManifestScanCode(code)
	if normalizedCode == "" {
		return -1, CarrierManifestLine{}, false
	}

	for index, line := range m.Lines {
		if line.MatchesScanCode(normalizedCode) {
			return index, line, true
		}
	}

	return -1, CarrierManifestLine{}, false
}

func (line CarrierManifestLine) MatchesScanCode(code string) bool {
	normalizedCode := NormalizeManifestScanCode(code)
	if normalizedCode == "" {
		return false
	}

	for _, candidate := range []string{line.OrderNo, line.TrackingNo, line.ShipmentID, line.PackageCode} {
		if NormalizeManifestScanCode(candidate) == normalizedCode {
			return true
		}
	}

	return false
}

func NormalizeManifestScanCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func normalizeManifestLocationCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
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
