package application

import (
	"context"
	"errors"
	"fmt"
)

func (s *PostgresItemCatalog) EnsureSeed(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("database connection is required")
	}

	for _, item := range prototypeItems() {
		item = item.Clone()
		if _, err := s.Get(ctx, item.ID); err == nil {
			continue
		} else if !errors.Is(err, ErrItemNotFound) {
			return fmt.Errorf("check seed item %q: %w", item.ID, err)
		}
		if err := s.saveNewItem(ctx, item); err != nil {
			return fmt.Errorf("seed item %q: %w", item.ID, err)
		}
	}

	return nil
}

func (s *PostgresWarehouseLocationCatalog) EnsureSeed(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("database connection is required")
	}

	for _, warehouse := range prototypeWarehouses() {
		warehouse = warehouse.Clone()
		if _, err := s.GetWarehouse(ctx, warehouse.ID); err == nil {
			continue
		} else if !errors.Is(err, ErrWarehouseNotFound) {
			return fmt.Errorf("check seed warehouse %q: %w", warehouse.ID, err)
		}
		if err := s.insertWarehouse(ctx, warehouse); err != nil {
			return fmt.Errorf("seed warehouse %q: %w", warehouse.ID, err)
		}
	}

	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	for _, location := range prototypeLocations() {
		location = location.Clone()
		if _, err := s.GetLocation(ctx, location.ID); err == nil {
			continue
		} else if !errors.Is(err, ErrLocationNotFound) {
			return fmt.Errorf("check seed location %q: %w", location.ID, err)
		}
		persistedWarehouseID, _, err := s.findPersistedWarehouse(ctx, orgID, location.WarehouseID)
		if err != nil {
			return fmt.Errorf("resolve seed location warehouse %q: %w", location.ID, err)
		}
		if err := s.insertLocation(ctx, orgID, persistedWarehouseID, location); err != nil {
			return fmt.Errorf("seed location %q: %w", location.ID, err)
		}
	}

	return nil
}

func (s *PostgresPartyCatalog) EnsureSeed(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("database connection is required")
	}

	for _, supplier := range prototypeSuppliers() {
		supplier = supplier.Clone()
		if _, err := s.GetSupplier(ctx, supplier.ID); err == nil {
			continue
		} else if !errors.Is(err, ErrSupplierNotFound) {
			return fmt.Errorf("check seed supplier %q: %w", supplier.ID, err)
		}
		if err := s.insertSupplier(ctx, supplier); err != nil {
			return fmt.Errorf("seed supplier %q: %w", supplier.ID, err)
		}
	}

	for _, customer := range prototypeCustomers() {
		customer = customer.Clone()
		if _, err := s.GetCustomer(ctx, customer.ID); err == nil {
			continue
		} else if !errors.Is(err, ErrCustomerNotFound) {
			return fmt.Errorf("check seed customer %q: %w", customer.ID, err)
		}
		if err := s.insertCustomer(ctx, customer); err != nil {
			return fmt.Errorf("seed customer %q: %w", customer.ID, err)
		}
	}

	return nil
}

func (c *PostgresUOMCatalog) EnsureSeed(ctx context.Context) error {
	if c == nil || c.db == nil {
		return errors.New("database connection is required")
	}

	return c.ensureSeed(ctx)
}
