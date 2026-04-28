package application

import (
	"context"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
)

type ListReturnMasterData struct{}

func NewListReturnMasterData() ListReturnMasterData {
	return ListReturnMasterData{}
}

func (ListReturnMasterData) Execute(context.Context) (domain.ReturnMasterData, error) {
	return domain.PrototypeReturnMasterData(), nil
}
