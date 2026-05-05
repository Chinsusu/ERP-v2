package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
)

type SubcontractFactoryDispatchService struct {
	clock func() time.Time
}

type CreateFactoryDispatchInput struct {
	ID                 string
	DispatchNo         string
	SubcontractOrderID string
	ExpectedVersion    int
	ActorID            string
	RequestID          string
	Note               string
}

type FactoryDispatchActionInput struct {
	SubcontractOrderID string
	DispatchID         string
	ExpectedVersion    int
	ActorID            string
	RequestID          string
}

type MarkFactoryDispatchSentInput struct {
	SubcontractOrderID string
	DispatchID         string
	ExpectedVersion    int
	SentBy             string
	SentAt             time.Time
	ActorID            string
	RequestID          string
	Note               string
	Evidence           []FactoryDispatchEvidenceInput
}

type RecordFactoryDispatchResponseInput struct {
	SubcontractOrderID string
	DispatchID         string
	ExpectedVersion    int
	ResponseStatus     string
	ResponseBy         string
	RespondedAt        time.Time
	ResponseNote       string
	ActorID            string
	RequestID          string
}

type FactoryDispatchEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type BuildSubcontractFactoryDispatchInput struct {
	ID         string
	DispatchNo string
	Order      productiondomain.SubcontractOrder
	ActorID    string
	Note       string
}

type FactoryDispatchResult struct {
	SubcontractOrder productiondomain.SubcontractOrder
	Dispatch         productiondomain.SubcontractFactoryDispatch
	AuditLogID       string
}

type PrototypeSubcontractFactoryDispatchStore struct {
	mu      sync.RWMutex
	records map[string]productiondomain.SubcontractFactoryDispatch
}

func NewSubcontractFactoryDispatchService() SubcontractFactoryDispatchService {
	return SubcontractFactoryDispatchService{clock: func() time.Time { return time.Now().UTC() }}
}

func NewPrototypeSubcontractFactoryDispatchStore() *PrototypeSubcontractFactoryDispatchStore {
	return &PrototypeSubcontractFactoryDispatchStore{records: make(map[string]productiondomain.SubcontractFactoryDispatch)}
}

func (s SubcontractFactoryDispatchService) BuildFromOrder(
	_ context.Context,
	input BuildSubcontractFactoryDispatchInput,
) (productiondomain.SubcontractFactoryDispatch, error) {
	if strings.TrimSpace(input.ActorID) == "" {
		return productiondomain.SubcontractFactoryDispatch{}, productiondomain.ErrSubcontractFactoryDispatchRequiredField
	}
	if productiondomain.NormalizeSubcontractOrderStatus(input.Order.Status) != productiondomain.SubcontractOrderStatusApproved {
		return productiondomain.SubcontractFactoryDispatch{}, productiondomain.ErrSubcontractOrderInvalidTransition
	}
	now := s.now()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newSubcontractFactoryDispatchID(now)
	}
	dispatchNo := strings.TrimSpace(input.DispatchNo)
	if dispatchNo == "" {
		dispatchNo = newSubcontractFactoryDispatchNo(now)
	}

	lines := make([]productiondomain.NewSubcontractFactoryDispatchLineInput, 0, len(input.Order.MaterialLines))
	for index, line := range input.Order.MaterialLines {
		lineID := fmt.Sprintf("%s-line-%02d", id, index+1)
		lines = append(lines, productiondomain.NewSubcontractFactoryDispatchLineInput{
			ID:                  lineID,
			LineNo:              index + 1,
			OrderMaterialLineID: line.ID,
			ItemID:              line.ItemID,
			SKUCode:             line.SKUCode,
			ItemName:            line.ItemName,
			PlannedQty:          line.PlannedQty,
			UOMCode:             line.UOMCode.String(),
			LotTraceRequired:    line.LotTraceRequired,
			Note:                line.Note,
		})
	}

	return productiondomain.NewSubcontractFactoryDispatch(productiondomain.NewSubcontractFactoryDispatchInput{
		ID:                     id,
		OrgID:                  input.Order.OrgID,
		DispatchNo:             dispatchNo,
		SubcontractOrderID:     input.Order.ID,
		SubcontractOrderNo:     input.Order.OrderNo,
		SourceProductionPlanID: input.Order.SourceProductionPlanID,
		SourceProductionPlanNo: input.Order.SourceProductionPlanNo,
		FactoryID:              input.Order.FactoryID,
		FactoryCode:            input.Order.FactoryCode,
		FactoryName:            input.Order.FactoryName,
		FinishedItemID:         input.Order.FinishedItemID,
		FinishedSKUCode:        input.Order.FinishedSKUCode,
		FinishedItemName:       input.Order.FinishedItemName,
		PlannedQty:             input.Order.PlannedQty,
		UOMCode:                input.Order.UOMCode.String(),
		SpecSummary:            input.Order.SpecSummary,
		SampleRequired:         input.Order.SampleRequired,
		TargetStartDate:        input.Order.TargetStartDate,
		ExpectedReceiptDate:    input.Order.ExpectedReceiptDate,
		Lines:                  lines,
		Note:                   input.Note,
		CreatedAt:              now,
		CreatedBy:              input.ActorID,
		UpdatedAt:              now,
		UpdatedBy:              input.ActorID,
	})
}

func (s SubcontractFactoryDispatchService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func (s SubcontractOrderService) ListFactoryDispatches(
	ctx context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractFactoryDispatch, error) {
	if s.factoryDispatchStore == nil {
		return nil, errors.New("subcontract factory dispatch store is required")
	}

	return s.factoryDispatchStore.ListBySubcontractOrder(ctx, subcontractOrderID)
}

func (s SubcontractOrderService) CreateFactoryDispatch(
	ctx context.Context,
	input CreateFactoryDispatchInput,
) (FactoryDispatchResult, error) {
	if err := s.ensureReadyForFactoryDispatch(); err != nil {
		return FactoryDispatchResult{}, err
	}
	if err := requireSubcontractOrderActor(input.ActorID); err != nil {
		return FactoryDispatchResult{}, err
	}

	var result FactoryDispatchResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		order, err := tx.GetForUpdate(txCtx, input.SubcontractOrderID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(input.SubcontractOrderID)})
		}
		if err := ensureSubcontractOrderExpectedVersion(order, input.ExpectedVersion); err != nil {
			return err
		}
		dispatch, err := s.factoryDispatchBuild.BuildFromOrder(txCtx, BuildSubcontractFactoryDispatchInput{
			ID:         input.ID,
			DispatchNo: input.DispatchNo,
			Order:      order,
			ActorID:    input.ActorID,
			Note:       input.Note,
		})
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": order.ID})
		}
		if err := s.factoryDispatchStore.Save(txCtx, dispatch); err != nil {
			return err
		}
		log, err := newSubcontractOrderAuditLog(
			input.ActorID,
			input.RequestID,
			subcontractFactoryDispatchCreatedAction,
			order,
			nil,
			subcontractFactoryDispatchAuditData(dispatch),
			dispatch.CreatedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = FactoryDispatchResult{SubcontractOrder: order, Dispatch: dispatch, AuditLogID: log.ID}

		return nil
	})
	if err != nil {
		return FactoryDispatchResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) MarkFactoryDispatchReady(
	ctx context.Context,
	input FactoryDispatchActionInput,
) (FactoryDispatchResult, error) {
	return s.updateFactoryDispatch(ctx, input.SubcontractOrderID, input.DispatchID, input.ExpectedVersion, input.ActorID, input.RequestID, subcontractFactoryDispatchReadyAction, func(_ productiondomain.SubcontractOrder, dispatch productiondomain.SubcontractFactoryDispatch) (productiondomain.SubcontractFactoryDispatch, productiondomain.SubcontractOrder, bool, error) {
		ready, err := dispatch.MarkReady(input.ActorID, s.now())
		return ready, productiondomain.SubcontractOrder{}, false, err
	})
}

func (s SubcontractOrderService) MarkFactoryDispatchSent(
	ctx context.Context,
	input MarkFactoryDispatchSentInput,
) (FactoryDispatchResult, error) {
	actorID := firstNonBlankSubcontractOrder(input.SentBy, input.ActorID)
	return s.updateFactoryDispatch(ctx, input.SubcontractOrderID, input.DispatchID, input.ExpectedVersion, actorID, input.RequestID, subcontractFactoryDispatchSentAction, func(_ productiondomain.SubcontractOrder, dispatch productiondomain.SubcontractFactoryDispatch) (productiondomain.SubcontractFactoryDispatch, productiondomain.SubcontractOrder, bool, error) {
		sent, err := dispatch.MarkSent(productiondomain.MarkSubcontractFactoryDispatchSentInput{
			SentBy:   actorID,
			SentAt:   input.SentAt,
			Evidence: factoryDispatchEvidenceInputs(input.Evidence),
			Note:     input.Note,
		})
		return sent, productiondomain.SubcontractOrder{}, false, err
	})
}

func (s SubcontractOrderService) RecordFactoryDispatchResponse(
	ctx context.Context,
	input RecordFactoryDispatchResponseInput,
) (FactoryDispatchResult, error) {
	return s.updateFactoryDispatch(ctx, input.SubcontractOrderID, input.DispatchID, input.ExpectedVersion, input.ActorID, input.RequestID, subcontractFactoryDispatchResponseAction, func(order productiondomain.SubcontractOrder, dispatch productiondomain.SubcontractFactoryDispatch) (productiondomain.SubcontractFactoryDispatch, productiondomain.SubcontractOrder, bool, error) {
		responseStatus := productiondomain.NormalizeSubcontractFactoryDispatchStatus(productiondomain.SubcontractFactoryDispatchStatus(input.ResponseStatus))
		responded, err := dispatch.RecordResponse(productiondomain.RecordSubcontractFactoryDispatchResponseInput{
			ResponseStatus: responseStatus,
			ResponseBy:     firstNonBlankSubcontractOrder(input.ResponseBy, input.ActorID),
			RespondedAt:    input.RespondedAt,
			ResponseNote:   input.ResponseNote,
		})
		if err != nil {
			return productiondomain.SubcontractFactoryDispatch{}, productiondomain.SubcontractOrder{}, false, err
		}
		if responseStatus != productiondomain.SubcontractFactoryDispatchStatusConfirmed {
			return responded, productiondomain.SubcontractOrder{}, false, nil
		}

		confirmed, err := order.ConfirmFactory(firstNonBlankSubcontractOrder(input.ResponseBy, input.ActorID), responded.RespondedAt)
		if err != nil {
			return productiondomain.SubcontractFactoryDispatch{}, productiondomain.SubcontractOrder{}, false, err
		}
		return responded, confirmed, true, nil
	})
}

func (s SubcontractOrderService) updateFactoryDispatch(
	ctx context.Context,
	subcontractOrderID string,
	dispatchID string,
	expectedVersion int,
	actorID string,
	requestID string,
	auditAction string,
	update func(productiondomain.SubcontractOrder, productiondomain.SubcontractFactoryDispatch) (productiondomain.SubcontractFactoryDispatch, productiondomain.SubcontractOrder, bool, error),
) (FactoryDispatchResult, error) {
	if err := s.ensureReadyForFactoryDispatch(); err != nil {
		return FactoryDispatchResult{}, err
	}
	if err := requireSubcontractOrderActor(actorID); err != nil {
		return FactoryDispatchResult{}, err
	}

	var result FactoryDispatchResult
	err := s.store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		order, err := tx.GetForUpdate(txCtx, subcontractOrderID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"subcontract_order_id": strings.TrimSpace(subcontractOrderID)})
		}
		dispatch, err := s.factoryDispatchStore.Get(txCtx, dispatchID)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"factory_dispatch_id": strings.TrimSpace(dispatchID)})
		}
		if dispatch.SubcontractOrderID != order.ID {
			return MapSubcontractOrderError(ErrSubcontractFactoryDispatchNotFound, map[string]any{
				"factory_dispatch_id":  strings.TrimSpace(dispatchID),
				"subcontract_order_id": order.ID,
			})
		}
		if err := ensureFactoryDispatchExpectedVersion(dispatch, expectedVersion); err != nil {
			return err
		}
		updatedDispatch, updatedOrder, shouldSaveOrder, err := update(order, dispatch)
		if err != nil {
			return MapSubcontractOrderError(err, map[string]any{"factory_dispatch_id": dispatch.ID})
		}
		if shouldSaveOrder {
			if err := tx.Save(txCtx, updatedOrder); err != nil {
				return err
			}
			order = updatedOrder
			if updatedDispatch.Status == productiondomain.SubcontractFactoryDispatchStatusConfirmed {
				auditAction = subcontractFactoryDispatchConfirmedAction
			}
		}
		if err := s.factoryDispatchStore.Save(txCtx, updatedDispatch); err != nil {
			return err
		}
		log, err := newSubcontractOrderAuditLog(
			actorID,
			requestID,
			auditAction,
			order,
			nil,
			subcontractFactoryDispatchAuditData(updatedDispatch),
			updatedDispatch.UpdatedAt,
		)
		if err != nil {
			return err
		}
		if err := tx.RecordAudit(txCtx, log); err != nil {
			return err
		}
		result = FactoryDispatchResult{SubcontractOrder: order, Dispatch: updatedDispatch, AuditLogID: log.ID}

		return nil
	})
	if err != nil {
		return FactoryDispatchResult{}, err
	}

	return result, nil
}

func (s SubcontractOrderService) ensureReadyForFactoryDispatch() error {
	if s.store == nil {
		return errors.New("subcontract order store is required")
	}
	if s.factoryDispatchStore == nil {
		return errors.New("subcontract factory dispatch store is required")
	}
	if s.factoryDispatchBuild.clock == nil {
		s.factoryDispatchBuild = NewSubcontractFactoryDispatchService()
	}

	return nil
}

func ensureFactoryDispatchExpectedVersion(dispatch productiondomain.SubcontractFactoryDispatch, expectedVersion int) error {
	if expectedVersion > 0 && dispatch.Version != expectedVersion {
		return ErrSubcontractOrderVersionConflict
	}

	return nil
}

func factoryDispatchEvidenceInputs(inputs []FactoryDispatchEvidenceInput) []productiondomain.NewSubcontractFactoryDispatchEvidenceInput {
	evidence := make([]productiondomain.NewSubcontractFactoryDispatchEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, productiondomain.NewSubcontractFactoryDispatchEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func subcontractFactoryDispatchAuditData(dispatch productiondomain.SubcontractFactoryDispatch) map[string]any {
	return map[string]any{
		"factory_dispatch_id":     dispatch.ID,
		"factory_dispatch_no":     dispatch.DispatchNo,
		"factory_dispatch_status": string(dispatch.Status),
		"factory_response_note":   dispatch.FactoryResponseNote,
	}
}

func newSubcontractFactoryDispatchID(now time.Time) string {
	return fmt.Sprintf("fdp-%s-%06d", now.Format("060102"), now.Nanosecond()%1000000)
}

func newSubcontractFactoryDispatchNo(now time.Time) string {
	return fmt.Sprintf("FDP-%s-%06d", now.Format("060102"), now.Nanosecond()%1000000)
}

func (s *PrototypeSubcontractFactoryDispatchStore) Save(
	_ context.Context,
	dispatch productiondomain.SubcontractFactoryDispatch,
) error {
	if s == nil {
		return errors.New("subcontract factory dispatch store is required")
	}
	if err := dispatch.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[dispatch.ID] = dispatch.Clone()

	return nil
}

func (s *PrototypeSubcontractFactoryDispatchStore) Get(
	_ context.Context,
	id string,
) (productiondomain.SubcontractFactoryDispatch, error) {
	if s == nil {
		return productiondomain.SubcontractFactoryDispatch{}, errors.New("subcontract factory dispatch store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	dispatch, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return productiondomain.SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchNotFound
	}

	return dispatch.Clone(), nil
}

func (s *PrototypeSubcontractFactoryDispatchStore) ListBySubcontractOrder(
	_ context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractFactoryDispatch, error) {
	if s == nil {
		return nil, errors.New("subcontract factory dispatch store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]productiondomain.SubcontractFactoryDispatch, 0)
	for _, dispatch := range s.records {
		if strings.EqualFold(dispatch.SubcontractOrderID, strings.TrimSpace(subcontractOrderID)) {
			records = append(records, dispatch.Clone())
		}
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})

	return records, nil
}

func (s *PrototypeSubcontractFactoryDispatchStore) GetLatestBySubcontractOrder(
	ctx context.Context,
	subcontractOrderID string,
) (productiondomain.SubcontractFactoryDispatch, error) {
	records, err := s.ListBySubcontractOrder(ctx, subcontractOrderID)
	if err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}
	if len(records) == 0 {
		return productiondomain.SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchNotFound
	}

	return records[0], nil
}

func (s *PrototypeSubcontractFactoryDispatchStore) Count() int {
	if s == nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.records)
}
