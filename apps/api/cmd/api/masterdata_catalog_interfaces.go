package main

import (
	"context"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type itemCatalog interface {
	List(context.Context, masterdatadomain.ItemFilter) ([]masterdatadomain.Item, response.Pagination, error)
	Get(context.Context, string) (masterdatadomain.Item, error)
	Create(context.Context, masterdataapp.CreateItemInput) (masterdataapp.ItemResult, error)
	Update(context.Context, masterdataapp.UpdateItemInput) (masterdataapp.ItemResult, error)
	ChangeStatus(context.Context, masterdataapp.ChangeItemStatusInput) (masterdataapp.ItemResult, error)
}

type formulaCatalog interface {
	List(context.Context, masterdatadomain.FormulaFilter) ([]masterdatadomain.Formula, error)
	Get(context.Context, string) (masterdatadomain.Formula, error)
	Create(context.Context, masterdataapp.CreateFormulaInput) (masterdataapp.FormulaResult, error)
	Activate(context.Context, masterdataapp.ActivateFormulaInput) (masterdataapp.FormulaResult, error)
	CalculateRequirement(context.Context, masterdataapp.CalculateFormulaRequirementInput) (masterdataapp.FormulaRequirementResult, error)
}

type warehouseLocationCatalog interface {
	ListWarehouses(context.Context, masterdatadomain.WarehouseFilter) ([]masterdatadomain.Warehouse, response.Pagination, error)
	GetWarehouse(context.Context, string) (masterdatadomain.Warehouse, error)
	CreateWarehouse(context.Context, masterdataapp.CreateWarehouseInput) (masterdataapp.WarehouseResult, error)
	UpdateWarehouse(context.Context, masterdataapp.UpdateWarehouseInput) (masterdataapp.WarehouseResult, error)
	ChangeWarehouseStatus(context.Context, masterdataapp.ChangeWarehouseStatusInput) (masterdataapp.WarehouseResult, error)
	ListLocations(context.Context, masterdatadomain.LocationFilter) ([]masterdatadomain.Location, response.Pagination, error)
	GetLocation(context.Context, string) (masterdatadomain.Location, error)
	CreateLocation(context.Context, masterdataapp.CreateLocationInput) (masterdataapp.LocationResult, error)
	UpdateLocation(context.Context, masterdataapp.UpdateLocationInput) (masterdataapp.LocationResult, error)
	ChangeLocationStatus(context.Context, masterdataapp.ChangeLocationStatusInput) (masterdataapp.LocationResult, error)
}

type partyCatalog interface {
	ListSuppliers(context.Context, masterdatadomain.SupplierFilter) ([]masterdatadomain.Supplier, response.Pagination, error)
	GetSupplier(context.Context, string) (masterdatadomain.Supplier, error)
	CreateSupplier(context.Context, masterdataapp.CreateSupplierInput) (masterdataapp.SupplierResult, error)
	UpdateSupplier(context.Context, masterdataapp.UpdateSupplierInput) (masterdataapp.SupplierResult, error)
	ChangeSupplierStatus(context.Context, masterdataapp.ChangeSupplierStatusInput) (masterdataapp.SupplierResult, error)
	ListCustomers(context.Context, masterdatadomain.CustomerFilter) ([]masterdatadomain.Customer, response.Pagination, error)
	GetCustomer(context.Context, string) (masterdatadomain.Customer, error)
	CreateCustomer(context.Context, masterdataapp.CreateCustomerInput) (masterdataapp.CustomerResult, error)
	UpdateCustomer(context.Context, masterdataapp.UpdateCustomerInput) (masterdataapp.CustomerResult, error)
	ChangeCustomerStatus(context.Context, masterdataapp.ChangeCustomerStatusInput) (masterdataapp.CustomerResult, error)
}

type uomCatalog interface {
	ConvertToBase(context.Context, masterdataapp.ConvertToBaseInput) (masterdataapp.ConvertToBaseResult, error)
}
