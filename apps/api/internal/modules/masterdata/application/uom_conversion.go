package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type UOMCatalog struct {
	mu          sync.RWMutex
	uoms        map[decimal.UOMCode]domain.UOM
	conversions []domain.UOMConversion
}

type ConvertToBaseInput struct {
	ItemID      string
	SKU         string
	Quantity    decimal.Decimal
	FromUOMCode string
	BaseUOMCode string
}

type ConvertToBaseResult struct {
	Quantity          decimal.Decimal
	SourceUOMCode     decimal.UOMCode
	BaseQuantity      decimal.Decimal
	BaseUOMCode       decimal.UOMCode
	ConversionFactor  decimal.Decimal
	ConversionType    domain.UOMConversionType
	ConversionItemID  string
	IsBasePassthrough bool
}

func NewPrototypeUOMCatalog() *UOMCatalog {
	catalog := &UOMCatalog{
		uoms:        make(map[decimal.UOMCode]domain.UOM),
		conversions: make([]domain.UOMConversion, 0, 12),
	}
	for _, uom := range phase1UOMs() {
		catalog.uoms[uom.Code] = uom
	}
	catalog.conversions = append(catalog.conversions, phase1Conversions()...)

	return catalog
}

func (c *UOMCatalog) ConvertToBase(_ context.Context, input ConvertToBaseInput) (ConvertToBaseResult, error) {
	if c == nil {
		return ConvertToBaseResult{}, errors.New("uom catalog is required")
	}

	quantity, err := decimal.ParseQuantity(input.Quantity.String())
	if err != nil || quantity.IsNegative() || quantity.IsZero() {
		return ConvertToBaseResult{}, domain.ErrUOMInvalid
	}
	fromUOM, err := decimal.NormalizeUOMCode(input.FromUOMCode)
	if err != nil {
		return ConvertToBaseResult{}, domain.ErrUOMInvalid
	}
	baseUOM, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return ConvertToBaseResult{}, domain.ErrUOMInvalid
	}
	if err := c.ensureActiveUOM(fromUOM); err != nil {
		return ConvertToBaseResult{}, c.conversionError(err, input, fromUOM, baseUOM)
	}
	if err := c.ensureActiveUOM(baseUOM); err != nil {
		return ConvertToBaseResult{}, c.conversionError(err, input, fromUOM, baseUOM)
	}

	if fromUOM == baseUOM {
		return ConvertToBaseResult{
			Quantity:          quantity,
			SourceUOMCode:     fromUOM,
			BaseQuantity:      quantity,
			BaseUOMCode:       baseUOM,
			ConversionFactor:  decimal.MustQuantity("1"),
			ConversionType:    domain.UOMConversionDirect,
			IsBasePassthrough: true,
		}, nil
	}

	conversion, err := c.findConversion(strings.TrimSpace(input.ItemID), fromUOM, baseUOM)
	if err != nil {
		return ConvertToBaseResult{}, c.conversionError(err, input, fromUOM, baseUOM)
	}
	baseQuantity, err := decimal.MultiplyQuantityByFactor(quantity, conversion.Factor)
	if err != nil {
		return ConvertToBaseResult{}, err
	}

	return ConvertToBaseResult{
		Quantity:         quantity,
		SourceUOMCode:    fromUOM,
		BaseQuantity:     baseQuantity,
		BaseUOMCode:      baseUOM,
		ConversionFactor: conversion.Factor,
		ConversionType:   conversion.ConversionType,
		ConversionItemID: conversion.ItemID,
	}, nil
}

func (c *UOMCatalog) UpsertConversion(conversion domain.UOMConversion) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for index, existing := range c.conversions {
		if existing.ItemID == conversion.ItemID && existing.FromUOMCode == conversion.FromUOMCode && existing.ToUOMCode == conversion.ToUOMCode {
			c.conversions[index] = conversion
			return
		}
	}
	c.conversions = append(c.conversions, conversion)
}

func (c *UOMCatalog) ensureActiveUOM(code decimal.UOMCode) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	uom, ok := c.uoms[code]
	if !ok {
		return domain.ErrUOMInvalid
	}
	if !uom.IsActive {
		return domain.ErrUOMInactive
	}

	return nil
}

func (c *UOMCatalog) findConversion(itemID string, fromUOM decimal.UOMCode, toUOM decimal.UOMCode) (domain.UOMConversion, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var inactive *domain.UOMConversion
	for _, conversion := range c.conversions {
		if conversion.ItemID != itemID || conversion.FromUOMCode != fromUOM || conversion.ToUOMCode != toUOM {
			continue
		}
		if conversion.IsActive {
			return conversion, nil
		}
		candidate := conversion
		inactive = &candidate
	}
	if inactive != nil {
		return domain.UOMConversion{}, domain.ErrUOMConversionInactive
	}

	for _, conversion := range c.conversions {
		if conversion.ItemID != "" || conversion.FromUOMCode != fromUOM || conversion.ToUOMCode != toUOM {
			continue
		}
		if conversion.IsActive {
			return conversion, nil
		}
		candidate := conversion
		inactive = &candidate
	}
	if inactive != nil {
		return domain.UOMConversion{}, domain.ErrUOMConversionInactive
	}

	return domain.UOMConversion{}, domain.ErrUOMConversionMissing
}

func (c *UOMCatalog) conversionError(cause error, input ConvertToBaseInput, fromUOM decimal.UOMCode, toUOM decimal.UOMCode) error {
	return domain.UOMConversionError{
		Cause:       cause,
		SKU:         strings.TrimSpace(input.SKU),
		ItemID:      strings.TrimSpace(input.ItemID),
		FromUOMCode: fromUOM.String(),
		ToUOMCode:   toUOM.String(),
	}
}

func phase1UOMs() []domain.UOM {
	rows := []struct {
		code        string
		nameVI      string
		nameEN      string
		group       domain.UOMGroup
		decimal     bool
		convertible bool
		description string
	}{
		{code: "MG", nameVI: "Milligram", nameEN: "Milligram", group: domain.UOMGroupMass, decimal: true, convertible: true, description: "R&D and small BOM quantity"},
		{code: "G", nameVI: "Gram", nameEN: "Gram", group: domain.UOMGroupMass, decimal: true, convertible: true, description: "Base UOM for solid raw materials"},
		{code: "KG", nameVI: "Kilogram", nameEN: "Kilogram", group: domain.UOMGroupMass, decimal: true, convertible: true, description: "Purchase UOM for solid raw materials"},
		{code: "ML", nameVI: "Milliliter", nameEN: "Milliliter", group: domain.UOMGroupVolume, decimal: true, convertible: true, description: "Base UOM for liquid materials"},
		{code: "L", nameVI: "Liter", nameEN: "Liter", group: domain.UOMGroupVolume, decimal: true, convertible: true, description: "Purchase UOM for liquid materials"},
		{code: "PCS", nameVI: "Piece", nameEN: "Piece", group: domain.UOMGroupCount, description: "Count unit"},
		{code: "BOTTLE", nameVI: "Chai", nameEN: "Bottle", group: domain.UOMGroupPack, description: "Item-specific package UOM"},
		{code: "JAR", nameVI: "Hũ", nameEN: "Jar", group: domain.UOMGroupPack, description: "Item-specific package UOM"},
		{code: "TUBE", nameVI: "Tuýp", nameEN: "Tube", group: domain.UOMGroupPack, description: "Item-specific package UOM"},
		{code: "BOX", nameVI: "Hộp", nameEN: "Box", group: domain.UOMGroupPack, description: "Item-specific package UOM"},
		{code: "CARTON", nameVI: "Thùng", nameEN: "Carton", group: domain.UOMGroupPack, description: "Item-specific package UOM"},
		{code: "SET", nameVI: "Bộ/combo", nameEN: "Set", group: domain.UOMGroupPack, description: "BOM/combo UOM"},
		{code: "BAG", nameVI: "T\u00fai", nameEN: "Bag", group: domain.UOMGroupPack, description: "Item-specific package UOM"},
		{code: "ROLL", nameVI: "Cu\u1ed9n", nameEN: "Roll", group: domain.UOMGroupPack, description: "Item-specific package UOM"},
		{code: "CM", nameVI: "Centimet", nameEN: "Centimeter", group: domain.UOMGroupPack, decimal: true, description: "Length unit for roll or fabric packaging"},
		{code: "SERVICE", nameVI: "Dịch vụ", nameEN: "Service", group: domain.UOMGroupService, description: "Non-stock service"},
	}

	uoms := make([]domain.UOM, 0, len(rows))
	for _, row := range rows {
		decimalScale := 0
		if row.decimal {
			decimalScale = decimal.QuantityScale
		}
		uom, err := domain.NewUOM(row.code, row.nameVI, row.nameEN, row.group, decimalScale, row.decimal, row.convertible, true, row.description)
		if err != nil {
			panic(fmt.Sprintf("invalid phase 1 uom %s: %v", row.code, err))
		}
		uoms = append(uoms, uom)
	}

	return uoms
}

func phase1Conversions() []domain.UOMConversion {
	rows := []struct {
		id             string
		itemID         string
		fromUOMCode    string
		toUOMCode      string
		factor         string
		conversionType domain.UOMConversionType
	}{
		{id: "global-kg-g", fromUOMCode: "KG", toUOMCode: "G", factor: "1000", conversionType: domain.UOMConversionGlobal},
		{id: "global-g-kg", fromUOMCode: "G", toUOMCode: "KG", factor: "0.001", conversionType: domain.UOMConversionGlobal},
		{id: "global-mg-g", fromUOMCode: "MG", toUOMCode: "G", factor: "0.001", conversionType: domain.UOMConversionGlobal},
		{id: "global-g-mg", fromUOMCode: "G", toUOMCode: "MG", factor: "1000", conversionType: domain.UOMConversionGlobal},
		{id: "global-l-ml", fromUOMCode: "L", toUOMCode: "ML", factor: "1000", conversionType: domain.UOMConversionGlobal},
		{id: "global-ml-l", fromUOMCode: "ML", toUOMCode: "L", factor: "0.001", conversionType: domain.UOMConversionGlobal},
		{id: "item-serum-carton-pcs", itemID: "item-serum-30ml", fromUOMCode: "CARTON", toUOMCode: "PCS", factor: "48", conversionType: domain.UOMConversionItemSpecific},
		{id: "item-serum-box-pcs", itemID: "item-serum-30ml", fromUOMCode: "BOX", toUOMCode: "PCS", factor: "12", conversionType: domain.UOMConversionItemSpecific},
		{id: "item-mask-set-pcs", itemID: "item-mask-set", fromUOMCode: "SET", toUOMCode: "PCS", factor: "5", conversionType: domain.UOMConversionItemSpecific},
	}

	conversions := make([]domain.UOMConversion, 0, len(rows))
	for _, row := range rows {
		conversion, err := domain.NewUOMConversion(row.id, row.itemID, row.fromUOMCode, row.toUOMCode, row.factor, row.conversionType, true)
		if err != nil {
			panic(fmt.Sprintf("invalid phase 1 uom conversion %s: %v", row.id, err))
		}
		conversions = append(conversions, conversion)
	}

	return conversions
}
