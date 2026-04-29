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

var ErrSubcontractPaymentMilestoneNotFound = errors.New("subcontract payment milestone not found")

type SubcontractPaymentMilestoneStore interface {
	Save(ctx context.Context, milestone productiondomain.SubcontractPaymentMilestone) error
	Get(ctx context.Context, id string) (productiondomain.SubcontractPaymentMilestone, error)
	ListBySubcontractOrder(ctx context.Context, subcontractOrderID string) ([]productiondomain.SubcontractPaymentMilestone, error)
}

type PrototypeSubcontractPaymentMilestoneStore struct {
	mu      sync.RWMutex
	records map[string]productiondomain.SubcontractPaymentMilestone
}

type SubcontractPaymentMilestoneService struct {
	clock func() time.Time
}

type BuildSubcontractDepositMilestoneInput struct {
	ID          string
	MilestoneNo string
	Order       productiondomain.SubcontractOrder
	Amount      string
	RecordedBy  string
	RecordedAt  time.Time
	ActorID     string
	Note        string
}

type BuildSubcontractFinalPaymentMilestoneInput struct {
	ID                  string
	MilestoneNo         string
	Order               productiondomain.SubcontractOrder
	Amount              string
	ReadyBy             string
	ReadyAt             time.Time
	BlockingClaims      []productiondomain.SubcontractFactoryClaim
	ApprovedExceptionID string
	ActorID             string
	Note                string
}

type SubcontractPaymentMilestoneBuildResult struct {
	Milestone      productiondomain.SubcontractPaymentMilestone
	UpdatedOrder   productiondomain.SubcontractOrder
	PreviousStatus productiondomain.SubcontractOrderStatus
	CurrentStatus  productiondomain.SubcontractOrderStatus
}

func NewSubcontractPaymentMilestoneService() SubcontractPaymentMilestoneService {
	return SubcontractPaymentMilestoneService{clock: func() time.Time { return time.Now().UTC() }}
}

func NewPrototypeSubcontractPaymentMilestoneStore() *PrototypeSubcontractPaymentMilestoneStore {
	return &PrototypeSubcontractPaymentMilestoneStore{records: make(map[string]productiondomain.SubcontractPaymentMilestone)}
}

func (s SubcontractPaymentMilestoneService) BuildDepositMilestone(
	ctx context.Context,
	input BuildSubcontractDepositMilestoneInput,
) (SubcontractPaymentMilestoneBuildResult, error) {
	_ = ctx
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return SubcontractPaymentMilestoneBuildResult{}, productiondomain.ErrSubcontractOrderTransitionActorRequired
	}
	amount, err := parseSubcontractPaymentMilestoneAmount(input.Amount)
	if err != nil {
		return SubcontractPaymentMilestoneBuildResult{}, err
	}

	now := s.now()
	recordedAt := input.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = now
	}
	recordedBy := firstNonBlankSubcontractOrder(input.RecordedBy, actorID)
	updatedOrder, err := input.Order.RecordDeposit(actorID, amount, recordedAt)
	if err != nil {
		return SubcontractPaymentMilestoneBuildResult{}, err
	}

	milestone, err := productiondomain.NewSubcontractPaymentMilestone(productiondomain.NewSubcontractPaymentMilestoneInput{
		ID:                 firstNonBlankSubcontractOrder(input.ID, newSubcontractPaymentMilestoneID(now)),
		OrgID:              input.Order.OrgID,
		MilestoneNo:        firstNonBlankSubcontractOrder(input.MilestoneNo, newSubcontractPaymentMilestoneNo(now)),
		SubcontractOrderID: input.Order.ID,
		SubcontractOrderNo: input.Order.OrderNo,
		FactoryID:          input.Order.FactoryID,
		FactoryCode:        input.Order.FactoryCode,
		FactoryName:        input.Order.FactoryName,
		Kind:               productiondomain.SubcontractPaymentMilestoneKindDeposit,
		Amount:             amount,
		CurrencyCode:       input.Order.CurrencyCode.String(),
		Note:               input.Note,
		CreatedAt:          now,
		CreatedBy:          actorID,
		UpdatedAt:          now,
		UpdatedBy:          actorID,
	})
	if err != nil {
		return SubcontractPaymentMilestoneBuildResult{}, err
	}
	milestone, err = milestone.Record(recordedBy, recordedAt)
	if err != nil {
		return SubcontractPaymentMilestoneBuildResult{}, err
	}

	return SubcontractPaymentMilestoneBuildResult{
		Milestone:      milestone,
		UpdatedOrder:   updatedOrder,
		PreviousStatus: input.Order.Status,
		CurrentStatus:  updatedOrder.Status,
	}, nil
}

func (s SubcontractPaymentMilestoneService) BuildFinalPaymentMilestone(
	ctx context.Context,
	input BuildSubcontractFinalPaymentMilestoneInput,
) (SubcontractPaymentMilestoneBuildResult, error) {
	_ = ctx
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return SubcontractPaymentMilestoneBuildResult{}, productiondomain.ErrSubcontractOrderTransitionActorRequired
	}
	approvedExceptionID := strings.TrimSpace(input.ApprovedExceptionID)
	if hasBlockingSubcontractFactoryClaim(input.BlockingClaims) && approvedExceptionID == "" {
		return SubcontractPaymentMilestoneBuildResult{}, productiondomain.ErrSubcontractPaymentMilestoneBlocked
	}
	amount, err := parseSubcontractFinalPaymentMilestoneAmount(input.Amount, input.Order.EstimatedCostAmount)
	if err != nil {
		return SubcontractPaymentMilestoneBuildResult{}, err
	}

	now := s.now()
	readyAt := input.ReadyAt
	if readyAt.IsZero() {
		readyAt = now
	}
	readyBy := firstNonBlankSubcontractOrder(input.ReadyBy, actorID)
	updatedOrder, err := input.Order.MarkFinalPaymentReady(actorID, readyAt)
	if err != nil {
		return SubcontractPaymentMilestoneBuildResult{}, err
	}

	milestone, err := productiondomain.NewSubcontractPaymentMilestone(productiondomain.NewSubcontractPaymentMilestoneInput{
		ID:                  firstNonBlankSubcontractOrder(input.ID, newSubcontractPaymentMilestoneID(now)),
		OrgID:               input.Order.OrgID,
		MilestoneNo:         firstNonBlankSubcontractOrder(input.MilestoneNo, newSubcontractPaymentMilestoneNo(now)),
		SubcontractOrderID:  input.Order.ID,
		SubcontractOrderNo:  input.Order.OrderNo,
		FactoryID:           input.Order.FactoryID,
		FactoryCode:         input.Order.FactoryCode,
		FactoryName:         input.Order.FactoryName,
		Kind:                productiondomain.SubcontractPaymentMilestoneKindFinalPayment,
		Amount:              amount,
		CurrencyCode:        input.Order.CurrencyCode.String(),
		Note:                input.Note,
		ApprovedExceptionID: approvedExceptionID,
		CreatedAt:           now,
		CreatedBy:           actorID,
		UpdatedAt:           now,
		UpdatedBy:           actorID,
	})
	if err != nil {
		return SubcontractPaymentMilestoneBuildResult{}, err
	}
	milestone, err = milestone.MarkReady(readyBy, readyAt)
	if err != nil {
		return SubcontractPaymentMilestoneBuildResult{}, err
	}

	return SubcontractPaymentMilestoneBuildResult{
		Milestone:      milestone,
		UpdatedOrder:   updatedOrder,
		PreviousStatus: input.Order.Status,
		CurrentStatus:  updatedOrder.Status,
	}, nil
}

func (s SubcontractPaymentMilestoneService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func parseSubcontractPaymentMilestoneAmount(value string) (decimal.Decimal, error) {
	amount, err := decimal.ParseMoneyAmount(value)
	if err != nil || amount.IsNegative() || amount.IsZero() {
		return "", productiondomain.ErrSubcontractPaymentMilestoneInvalidAmount
	}

	return amount, nil
}

func parseSubcontractFinalPaymentMilestoneAmount(value string, defaultAmount decimal.Decimal) (decimal.Decimal, error) {
	if strings.TrimSpace(value) != "" {
		return parseSubcontractPaymentMilestoneAmount(value)
	}

	return parseSubcontractPaymentMilestoneAmount(defaultAmount.String())
}

func hasBlockingSubcontractFactoryClaim(claims []productiondomain.SubcontractFactoryClaim) bool {
	for _, claim := range claims {
		if claim.BlocksFinalPayment() {
			return true
		}
	}

	return false
}

func newSubcontractPaymentMilestoneID(now time.Time) string {
	return fmt.Sprintf("spm-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newSubcontractPaymentMilestoneNo(now time.Time) string {
	return fmt.Sprintf("SPM-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func (s *PrototypeSubcontractPaymentMilestoneStore) Save(
	_ context.Context,
	milestone productiondomain.SubcontractPaymentMilestone,
) error {
	if s == nil {
		return errors.New("subcontract payment milestone store is required")
	}
	if err := milestone.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[milestone.ID] = milestone.Clone()

	return nil
}

func (s *PrototypeSubcontractPaymentMilestoneStore) Get(
	_ context.Context,
	id string,
) (productiondomain.SubcontractPaymentMilestone, error) {
	if s == nil {
		return productiondomain.SubcontractPaymentMilestone{}, errors.New("subcontract payment milestone store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	milestone, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return productiondomain.SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneNotFound
	}

	return milestone.Clone(), nil
}

func (s *PrototypeSubcontractPaymentMilestoneStore) ListBySubcontractOrder(
	_ context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractPaymentMilestone, error) {
	if s == nil {
		return nil, errors.New("subcontract payment milestone store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	milestones := make([]productiondomain.SubcontractPaymentMilestone, 0)
	for _, milestone := range s.records {
		if milestone.SubcontractOrderID == strings.TrimSpace(subcontractOrderID) {
			milestones = append(milestones, milestone.Clone())
		}
	}

	return milestones, nil
}

func (s *PrototypeSubcontractPaymentMilestoneStore) Count() int {
	if s == nil {
		return 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.records)
}
