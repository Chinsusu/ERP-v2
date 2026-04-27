package domain

import (
	"errors"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrUOMInvalid = errors.New("uom is invalid")
var ErrUOMInactive = errors.New("uom is inactive")
var ErrUOMConversionMissing = errors.New("uom conversion is missing")
var ErrUOMConversionInactive = errors.New("uom conversion is inactive")

type UOMGroup string

const (
	UOMGroupMass    UOMGroup = "MASS"
	UOMGroupVolume  UOMGroup = "VOLUME"
	UOMGroupCount   UOMGroup = "COUNT"
	UOMGroupPack    UOMGroup = "PACK"
	UOMGroupService UOMGroup = "SERVICE"
)

type UOM struct {
	Code                decimal.UOMCode
	NameVI              string
	NameEN              string
	Group               UOMGroup
	DecimalScale        int
	AllowDecimal        bool
	IsGlobalConvertible bool
	IsActive            bool
	Description         string
}

type UOMConversionType string

const (
	UOMConversionGlobal       UOMConversionType = "GLOBAL"
	UOMConversionItemSpecific UOMConversionType = "ITEM_SPECIFIC"
	UOMConversionDirect       UOMConversionType = "DIRECT"
)

type UOMConversion struct {
	ID             string
	ItemID         string
	FromUOMCode    decimal.UOMCode
	ToUOMCode      decimal.UOMCode
	Factor         decimal.Decimal
	ConversionType UOMConversionType
	IsActive       bool
}

type UOMConversionError struct {
	Cause       error
	SKU         string
	ItemID      string
	FromUOMCode string
	ToUOMCode   string
}

func (e UOMConversionError) Error() string {
	if e.Cause == nil {
		return "uom conversion error"
	}

	return e.Cause.Error()
}

func (e UOMConversionError) Unwrap() error {
	return e.Cause
}

func (e UOMConversionError) Details() map[string]any {
	return map[string]any{
		"sku_code":      e.SKU,
		"item_id":       e.ItemID,
		"from_uom_code": e.FromUOMCode,
		"to_uom_code":   e.ToUOMCode,
		"base_uom_code": e.ToUOMCode,
	}
}

func NewUOM(code string, nameVI string, nameEN string, group UOMGroup, decimalScale int, allowDecimal bool, globalConvertible bool, active bool, description string) (UOM, error) {
	normalizedCode, err := decimal.NormalizeUOMCode(code)
	if err != nil {
		return UOM{}, ErrUOMInvalid
	}
	group = UOMGroup(strings.ToUpper(strings.TrimSpace(string(group))))
	if !IsValidUOMGroup(group) {
		return UOM{}, ErrUOMInvalid
	}
	if decimalScale < 0 || decimalScale > decimal.QuantityScale {
		return UOM{}, ErrUOMInvalid
	}
	if !allowDecimal && decimalScale != 0 {
		return UOM{}, ErrUOMInvalid
	}

	return UOM{
		Code:                normalizedCode,
		NameVI:              strings.TrimSpace(nameVI),
		NameEN:              strings.TrimSpace(nameEN),
		Group:               group,
		DecimalScale:        decimalScale,
		AllowDecimal:        allowDecimal,
		IsGlobalConvertible: globalConvertible,
		IsActive:            active,
		Description:         strings.TrimSpace(description),
	}, nil
}

func NewUOMConversion(id string, itemID string, fromUOMCode string, toUOMCode string, factor string, conversionType UOMConversionType, active bool) (UOMConversion, error) {
	itemID = strings.TrimSpace(itemID)
	from, err := decimal.NormalizeUOMCode(fromUOMCode)
	if err != nil {
		return UOMConversion{}, ErrUOMInvalid
	}
	to, err := decimal.NormalizeUOMCode(toUOMCode)
	if err != nil {
		return UOMConversion{}, ErrUOMInvalid
	}
	parsedFactor, err := decimal.ParseQuantity(factor)
	if err != nil || parsedFactor.IsNegative() || parsedFactor.IsZero() {
		return UOMConversion{}, ErrUOMInvalid
	}
	conversionType = UOMConversionType(strings.ToUpper(strings.TrimSpace(string(conversionType))))
	if conversionType == "" {
		conversionType = UOMConversionGlobal
	}
	if !IsValidUOMConversionType(conversionType) || conversionType == UOMConversionDirect {
		return UOMConversion{}, ErrUOMInvalid
	}
	if conversionType == UOMConversionGlobal && itemID != "" {
		return UOMConversion{}, ErrUOMInvalid
	}
	if conversionType == UOMConversionItemSpecific && itemID == "" {
		return UOMConversion{}, ErrUOMInvalid
	}

	return UOMConversion{
		ID:             strings.TrimSpace(id),
		ItemID:         itemID,
		FromUOMCode:    from,
		ToUOMCode:      to,
		Factor:         parsedFactor,
		ConversionType: conversionType,
		IsActive:       active,
	}, nil
}

func IsValidUOMGroup(group UOMGroup) bool {
	switch group {
	case UOMGroupMass, UOMGroupVolume, UOMGroupCount, UOMGroupPack, UOMGroupService:
		return true
	default:
		return false
	}
}

func IsValidUOMConversionType(conversionType UOMConversionType) bool {
	switch conversionType {
	case UOMConversionGlobal, UOMConversionItemSpecific, UOMConversionDirect:
		return true
	default:
		return false
	}
}
