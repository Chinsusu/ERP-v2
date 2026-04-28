package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type CarrierStatus string

const CarrierStatusActive CarrierStatus = "active"
const CarrierStatusInactive CarrierStatus = "inactive"

var ErrCarrierRequiredField = errors.New("carrier required field is missing")
var ErrCarrierInvalidStatus = errors.New("carrier status is invalid")

type Carrier struct {
	ID           string
	Code         string
	Name         string
	HandoverZone string
	Status       CarrierStatus
	SLAProfile   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type NewCarrierInput struct {
	ID           string
	Code         string
	Name         string
	HandoverZone string
	Status       CarrierStatus
	SLAProfile   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CarrierFilter struct {
	Search string
	Status CarrierStatus
}

func NewCarrier(input NewCarrierInput) (Carrier, error) {
	now := time.Now().UTC()
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	carrier := Carrier{
		ID:           strings.TrimSpace(input.ID),
		Code:         strings.ToUpper(strings.TrimSpace(input.Code)),
		Name:         strings.TrimSpace(input.Name),
		HandoverZone: strings.TrimSpace(input.HandoverZone),
		Status:       NormalizeCarrierStatus(input.Status),
		SLAProfile:   strings.TrimSpace(input.SLAProfile),
		CreatedAt:    createdAt.UTC(),
		UpdatedAt:    updatedAt.UTC(),
	}
	if carrier.ID == "" && carrier.Code != "" {
		carrier.ID = fmt.Sprintf("carrier-%s", strings.ToLower(carrier.Code))
	}
	if carrier.Status == "" && strings.TrimSpace(string(input.Status)) == "" {
		carrier.Status = CarrierStatusActive
	}
	if carrier.SLAProfile == "" {
		carrier.SLAProfile = "standard"
	}
	if err := carrier.Validate(); err != nil {
		return Carrier{}, err
	}

	return carrier, nil
}

func NormalizeCarrierStatus(status CarrierStatus) CarrierStatus {
	switch CarrierStatus(strings.ToLower(strings.TrimSpace(string(status)))) {
	case CarrierStatusActive:
		return CarrierStatusActive
	case CarrierStatusInactive:
		return CarrierStatusInactive
	default:
		return ""
	}
}

func (c Carrier) Validate() error {
	if strings.TrimSpace(c.ID) == "" ||
		strings.TrimSpace(c.Code) == "" ||
		strings.TrimSpace(c.Name) == "" ||
		strings.TrimSpace(c.HandoverZone) == "" ||
		strings.TrimSpace(c.SLAProfile) == "" ||
		c.CreatedAt.IsZero() ||
		c.UpdatedAt.IsZero() {
		return ErrCarrierRequiredField
	}
	if NormalizeCarrierStatus(c.Status) == "" {
		return ErrCarrierInvalidStatus
	}

	return nil
}

func (c Carrier) IsActive() bool {
	return c.Status == CarrierStatusActive
}

func (c Carrier) Clone() Carrier {
	return c
}

func NewCarrierFilter(search string, status CarrierStatus) CarrierFilter {
	return CarrierFilter{
		Search: strings.TrimSpace(search),
		Status: NormalizeCarrierStatus(status),
	}
}

func SortCarriers(carriers []Carrier) {
	sort.SliceStable(carriers, func(i, j int) bool {
		return carriers[i].Code < carriers[j].Code
	})
}
