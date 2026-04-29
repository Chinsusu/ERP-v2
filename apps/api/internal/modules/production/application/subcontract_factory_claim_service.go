package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSubcontractFactoryClaimNotFound = errors.New("subcontract factory claim not found")

type SubcontractFactoryClaimStore interface {
	Save(ctx context.Context, claim productiondomain.SubcontractFactoryClaim) error
	Get(ctx context.Context, id string) (productiondomain.SubcontractFactoryClaim, error)
	ListBySubcontractOrder(ctx context.Context, subcontractOrderID string) ([]productiondomain.SubcontractFactoryClaim, error)
}

type PrototypeSubcontractFactoryClaimStore struct {
	mu      sync.RWMutex
	records map[string]productiondomain.SubcontractFactoryClaim
}

type SubcontractFactoryClaimService struct {
	clock func() time.Time
}

type BuildSubcontractFactoryClaimInput struct {
	ID              string
	ClaimNo         string
	Order           productiondomain.SubcontractOrder
	ReceiptID       string
	ReceiptNo       string
	ReasonCode      string
	Reason          string
	Severity        string
	AffectedQty     string
	UOMCode         string
	BaseAffectedQty string
	BaseUOMCode     string
	Evidence        []BuildSubcontractFactoryClaimEvidenceInput
	OwnerID         string
	OpenedBy        string
	OpenedAt        time.Time
	ActorID         string
}

type BuildSubcontractFactoryClaimEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type SubcontractFactoryClaimBuildResult struct {
	Claim          productiondomain.SubcontractFactoryClaim
	UpdatedOrder   productiondomain.SubcontractOrder
	PreviousStatus productiondomain.SubcontractOrderStatus
	CurrentStatus  productiondomain.SubcontractOrderStatus
}

func NewSubcontractFactoryClaimService() SubcontractFactoryClaimService {
	return SubcontractFactoryClaimService{clock: func() time.Time { return time.Now().UTC() }}
}

func NewPrototypeSubcontractFactoryClaimStore() *PrototypeSubcontractFactoryClaimStore {
	return &PrototypeSubcontractFactoryClaimStore{records: make(map[string]productiondomain.SubcontractFactoryClaim)}
}

func (s SubcontractFactoryClaimService) BuildClaim(
	ctx context.Context,
	input BuildSubcontractFactoryClaimInput,
) (SubcontractFactoryClaimBuildResult, error) {
	claim, err := s.BuildClaimOnly(ctx, input)
	if err != nil {
		return SubcontractFactoryClaimBuildResult{}, err
	}
	actorID := strings.TrimSpace(input.ActorID)
	updatedOrder, err := input.Order.RejectFinishedGoodsWithFactoryIssue(actorID, claim.Reason, claim.OpenedAt)
	if err != nil {
		return SubcontractFactoryClaimBuildResult{}, err
	}

	return SubcontractFactoryClaimBuildResult{
		Claim:          claim,
		UpdatedOrder:   updatedOrder,
		PreviousStatus: input.Order.Status,
		CurrentStatus:  updatedOrder.Status,
	}, nil
}

func (s SubcontractFactoryClaimService) BuildClaimOnly(
	ctx context.Context,
	input BuildSubcontractFactoryClaimInput,
) (productiondomain.SubcontractFactoryClaim, error) {
	_ = ctx
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return productiondomain.SubcontractFactoryClaim{}, productiondomain.ErrSubcontractOrderTransitionActorRequired
	}

	now := s.now()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newSubcontractFactoryClaimID(now)
	}
	claimNo := strings.TrimSpace(input.ClaimNo)
	if claimNo == "" {
		claimNo = newSubcontractFactoryClaimNo(now)
	}
	openedAt := input.OpenedAt
	if openedAt.IsZero() {
		openedAt = now
	}
	claimWindowDays := input.Order.ClaimWindowDays
	if claimWindowDays == 0 {
		claimWindowDays = 7
	}
	if claimWindowDays < 3 || claimWindowDays > 7 {
		return productiondomain.SubcontractFactoryClaim{}, productiondomain.ErrSubcontractFactoryClaimInvalidSLA
	}

	affectedQty, baseAffectedQty, err := parseSubcontractFactoryClaimQuantities(input)
	if err != nil {
		return productiondomain.SubcontractFactoryClaim{}, err
	}
	return productiondomain.NewSubcontractFactoryClaim(productiondomain.NewSubcontractFactoryClaimInput{
		ID:                 id,
		OrgID:              input.Order.OrgID,
		ClaimNo:            claimNo,
		SubcontractOrderID: input.Order.ID,
		SubcontractOrderNo: input.Order.OrderNo,
		FactoryID:          input.Order.FactoryID,
		FactoryCode:        input.Order.FactoryCode,
		FactoryName:        input.Order.FactoryName,
		ReceiptID:          input.ReceiptID,
		ReceiptNo:          input.ReceiptNo,
		ReasonCode:         input.ReasonCode,
		Reason:             input.Reason,
		Severity:           input.Severity,
		AffectedQty:        affectedQty,
		UOMCode:            firstNonBlankSubcontractOrder(input.UOMCode, input.Order.UOMCode.String()),
		BaseAffectedQty:    baseAffectedQty,
		BaseUOMCode:        firstNonBlankSubcontractOrder(input.BaseUOMCode, input.Order.BaseUOMCode.String()),
		Evidence:           subcontractFactoryClaimEvidenceInputs(id, input.Evidence),
		OwnerID:            firstNonBlankSubcontractOrder(input.OwnerID, actorID),
		OpenedBy:           firstNonBlankSubcontractOrder(input.OpenedBy, actorID),
		OpenedAt:           openedAt,
		DueAt:              openedAt.AddDate(0, 0, claimWindowDays),
		CreatedAt:          now,
		CreatedBy:          actorID,
		UpdatedAt:          now,
		UpdatedBy:          actorID,
	})
}

func (s SubcontractFactoryClaimService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func parseSubcontractFactoryClaimQuantities(
	input BuildSubcontractFactoryClaimInput,
) (decimal.Decimal, decimal.Decimal, error) {
	affectedQty, err := decimal.ParseQuantity(input.AffectedQty)
	if err != nil || affectedQty.IsNegative() || affectedQty.IsZero() {
		return "", "", productiondomain.ErrSubcontractFactoryClaimInvalidQuantity
	}
	baseAffectedQty := decimal.Decimal("")
	if strings.TrimSpace(input.BaseAffectedQty) != "" {
		baseAffectedQty, err = decimal.ParseQuantity(input.BaseAffectedQty)
		if err != nil || baseAffectedQty.IsNegative() || baseAffectedQty.IsZero() {
			return "", "", productiondomain.ErrSubcontractFactoryClaimInvalidQuantity
		}
	} else if strings.EqualFold(
		firstNonBlankSubcontractOrder(input.UOMCode, input.Order.UOMCode.String()),
		firstNonBlankSubcontractOrder(input.BaseUOMCode, input.Order.BaseUOMCode.String()),
	) {
		baseAffectedQty = affectedQty
	} else {
		baseAffectedQty, err = decimal.MultiplyQuantityByFactor(affectedQty, input.Order.ConversionFactor)
		if err != nil {
			return "", "", productiondomain.ErrSubcontractFactoryClaimInvalidQuantity
		}
	}

	return affectedQty, baseAffectedQty, nil
}

func subcontractFactoryClaimEvidenceInputs(
	claimID string,
	inputs []BuildSubcontractFactoryClaimEvidenceInput,
) []productiondomain.NewSubcontractFactoryClaimEvidenceInput {
	evidence := make([]productiondomain.NewSubcontractFactoryClaimEvidenceInput, 0, len(inputs))
	for index, input := range inputs {
		id := strings.TrimSpace(input.ID)
		if id == "" {
			id = fmt.Sprintf("%s-evidence-%02d", strings.TrimSpace(claimID), index+1)
		}
		evidence = append(evidence, productiondomain.NewSubcontractFactoryClaimEvidenceInput{
			ID:           id,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func newSubcontractFactoryClaimID(now time.Time) string {
	return fmt.Sprintf("sfc-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newSubcontractFactoryClaimNo(now time.Time) string {
	return fmt.Sprintf("SFC-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func (s *PrototypeSubcontractFactoryClaimStore) Save(
	_ context.Context,
	claim productiondomain.SubcontractFactoryClaim,
) error {
	if s == nil {
		return fmt.Errorf("subcontract factory claim store is required")
	}
	if err := claim.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[claim.ID] = claim.Clone()

	return nil
}

func (s *PrototypeSubcontractFactoryClaimStore) Get(
	_ context.Context,
	id string,
) (productiondomain.SubcontractFactoryClaim, error) {
	if s == nil {
		return productiondomain.SubcontractFactoryClaim{}, fmt.Errorf("subcontract factory claim store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	claim, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return productiondomain.SubcontractFactoryClaim{}, ErrSubcontractFactoryClaimNotFound
	}

	return claim.Clone(), nil
}

func (s *PrototypeSubcontractFactoryClaimStore) ListBySubcontractOrder(
	_ context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractFactoryClaim, error) {
	if s == nil {
		return nil, fmt.Errorf("subcontract factory claim store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	claims := make([]productiondomain.SubcontractFactoryClaim, 0)
	for _, claim := range s.records {
		if claim.SubcontractOrderID == strings.TrimSpace(subcontractOrderID) {
			claims = append(claims, claim.Clone())
		}
	}

	return claims, nil
}

func (s *PrototypeSubcontractFactoryClaimStore) Count() int {
	if s == nil {
		return 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.records)
}
