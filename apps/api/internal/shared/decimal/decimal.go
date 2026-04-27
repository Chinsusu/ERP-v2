package decimal

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

const (
	LocaleVI           = "vi-VN"
	TimezoneHoChiMinh  = "Asia/Ho_Chi_Minh"
	CurrencyVND        = CurrencyCode("VND")
	MoneyScale         = 2
	UnitPriceScale     = 4
	UnitCostScale      = 6
	QuantityScale      = 6
	RateScale          = 4
	moneyPrecision     = 18
	unitPricePrecision = 18
	unitCostPrecision  = 18
	quantityPrecision  = 18
	ratePrecision      = 9
	uomCodeMaxLength   = 20
	currencyCodeLength = 3
)

var (
	ErrInvalidDecimal      = errors.New("decimal value is invalid")
	ErrDecimalOutOfRange   = errors.New("decimal value is out of range")
	ErrInvalidCurrencyCode = errors.New("currency code is invalid")
	ErrInvalidUOMCode      = errors.New("uom code is invalid")
)

type Decimal string
type CurrencyCode string
type UOMCode string

func ParseMoneyAmount(value string) (Decimal, error) {
	return parseFixedScale(value, moneyPrecision, MoneyScale)
}

func ParseUnitPrice(value string) (Decimal, error) {
	return parseFixedScale(value, unitPricePrecision, UnitPriceScale)
}

func ParseUnitCost(value string) (Decimal, error) {
	return parseFixedScale(value, unitCostPrecision, UnitCostScale)
}

func ParseQuantity(value string) (Decimal, error) {
	return parseFixedScale(value, quantityPrecision, QuantityScale)
}

func ParseRate(value string) (Decimal, error) {
	return parseFixedScale(value, ratePrecision, RateScale)
}

func RoundMoneyAmount(value string) (Decimal, error) {
	return roundFixedScale(value, moneyPrecision, MoneyScale)
}

func RoundUnitPrice(value string) (Decimal, error) {
	return roundFixedScale(value, unitPricePrecision, UnitPriceScale)
}

func RoundUnitCost(value string) (Decimal, error) {
	return roundFixedScale(value, unitCostPrecision, UnitCostScale)
}

func RoundQuantity(value string) (Decimal, error) {
	return roundFixedScale(value, quantityPrecision, QuantityScale)
}

func RoundRate(value string) (Decimal, error) {
	return roundFixedScale(value, ratePrecision, RateScale)
}

func MultiplyQuantityByFactor(quantity Decimal, factor Decimal) (Decimal, error) {
	left, err := scaledInt(quantity.String(), QuantityScale)
	if err != nil {
		return "", err
	}
	right, err := scaledInt(factor.String(), QuantityScale)
	if err != nil {
		return "", err
	}

	product := new(big.Int).Mul(left, right)
	return scaledProductToDecimal(product, QuantityScale*2, QuantityScale, quantityPrecision)
}

func AddQuantity(left Decimal, right Decimal) (Decimal, error) {
	leftValue, err := scaledInt(left.String(), QuantityScale)
	if err != nil {
		return "", err
	}
	rightValue, err := scaledInt(right.String(), QuantityScale)
	if err != nil {
		return "", err
	}

	return scaledIntToDecimal(leftValue.Add(leftValue, rightValue), QuantityScale, quantityPrecision)
}

func SubtractQuantity(left Decimal, right Decimal) (Decimal, error) {
	leftValue, err := scaledInt(left.String(), QuantityScale)
	if err != nil {
		return "", err
	}
	rightValue, err := scaledInt(right.String(), QuantityScale)
	if err != nil {
		return "", err
	}

	return scaledIntToDecimal(leftValue.Sub(leftValue, rightValue), QuantityScale, quantityPrecision)
}

func MustMoneyAmount(value string) Decimal {
	return must(ParseMoneyAmount(value))
}

func MustUnitPrice(value string) Decimal {
	return must(ParseUnitPrice(value))
}

func MustUnitCost(value string) Decimal {
	return must(ParseUnitCost(value))
}

func MustQuantity(value string) Decimal {
	return must(ParseQuantity(value))
}

func MustRate(value string) Decimal {
	return must(ParseRate(value))
}

func (d Decimal) String() string {
	return string(d)
}

func (c CurrencyCode) String() string {
	return string(c)
}

func (c UOMCode) String() string {
	return string(c)
}

func (d Decimal) IsNegative() bool {
	return strings.HasPrefix(string(d), "-") && !d.IsZero()
}

func (d Decimal) IsZero() bool {
	value := strings.TrimPrefix(string(d), "-")
	value = strings.ReplaceAll(value, ".", "")
	value = strings.TrimLeft(value, "0")

	return value == ""
}

func NormalizeCurrencyCode(value string) (CurrencyCode, error) {
	code := strings.ToUpper(strings.TrimSpace(value))
	if code == "" {
		code = string(CurrencyVND)
	}
	if len(code) != currencyCodeLength {
		return "", ErrInvalidCurrencyCode
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return "", ErrInvalidCurrencyCode
		}
	}

	return CurrencyCode(code), nil
}

func MustCurrencyCode(value string) CurrencyCode {
	code, err := NormalizeCurrencyCode(value)
	if err != nil {
		panic(err)
	}

	return code
}

func NormalizeUOMCode(value string) (UOMCode, error) {
	code := strings.ToUpper(strings.TrimSpace(value))
	if code == "" || len(code) > uomCodeMaxLength {
		return "", ErrInvalidUOMCode
	}
	for _, r := range code {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return "", ErrInvalidUOMCode
	}

	return UOMCode(code), nil
}

func MustUOMCode(value string) UOMCode {
	code, err := NormalizeUOMCode(value)
	if err != nil {
		panic(err)
	}

	return code
}

func parseFixedScale(value string, precision int, scale int) (Decimal, error) {
	parts, err := splitDecimal(value)
	if err != nil {
		return "", err
	}
	if len(parts.fracPart) > scale {
		return "", fmt.Errorf("%w: max scale %d", ErrInvalidDecimal, scale)
	}
	parts.fracPart = parts.fracPart + strings.Repeat("0", scale-len(parts.fracPart))

	return parts.toDecimal(precision, scale)
}

func roundFixedScale(value string, precision int, scale int) (Decimal, error) {
	parts, err := splitDecimal(value)
	if err != nil {
		return "", err
	}
	if len(parts.fracPart) <= scale {
		parts.fracPart = parts.fracPart + strings.Repeat("0", scale-len(parts.fracPart))
		return parts.toDecimal(precision, scale)
	}

	roundUp := parts.fracPart[scale] >= '5'
	keptFrac := parts.fracPart[:scale]
	number := parts.intPart + keptFrac
	if number == "" {
		number = "0"
	}
	intValue, ok := new(big.Int).SetString(number, 10)
	if !ok {
		return "", ErrInvalidDecimal
	}
	if roundUp {
		intValue.Add(intValue, big.NewInt(1))
	}

	digits := intValue.String()
	if scale > 0 && len(digits) <= scale {
		digits = strings.Repeat("0", scale-len(digits)+1) + digits
	}

	if scale == 0 {
		parts.intPart = digits
		parts.fracPart = ""
	} else {
		parts.intPart = digits[:len(digits)-scale]
		parts.fracPart = digits[len(digits)-scale:]
	}
	parts.intPart = trimLeadingZeros(parts.intPart)

	return parts.toDecimal(precision, scale)
}

type decimalParts struct {
	negative bool
	intPart  string
	fracPart string
}

func splitDecimal(value string) (decimalParts, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		raw = "0"
	}

	negative := false
	if strings.HasPrefix(raw, "-") {
		negative = true
		raw = strings.TrimPrefix(raw, "-")
	}
	if strings.HasPrefix(raw, "+") {
		raw = strings.TrimPrefix(raw, "+")
	}
	if raw == "" || strings.Count(raw, ".") > 1 {
		return decimalParts{}, ErrInvalidDecimal
	}

	segments := strings.Split(raw, ".")
	intPart := segments[0]
	fracPart := ""
	if len(segments) == 2 {
		fracPart = segments[1]
	}
	if intPart == "" {
		return decimalParts{}, ErrInvalidDecimal
	}
	if !digitsOnly(intPart) || !digitsOnly(fracPart) {
		return decimalParts{}, ErrInvalidDecimal
	}

	return decimalParts{
		negative: negative,
		intPart:  trimLeadingZeros(intPart),
		fracPart: fracPart,
	}, nil
}

func (p decimalParts) toDecimal(precision int, scale int) (Decimal, error) {
	intPart := trimLeadingZeros(p.intPart)
	if intPart == "" {
		intPart = "0"
	}
	if len(intPart) > precision-scale {
		return "", ErrDecimalOutOfRange
	}
	if len(p.fracPart) != scale {
		return "", ErrInvalidDecimal
	}
	value := intPart
	if scale > 0 {
		value += "." + p.fracPart
	}
	if p.negative && !allZero(intPart+p.fracPart) {
		value = "-" + value
	}

	return Decimal(value), nil
}

func digitsOnly(value string) bool {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func trimLeadingZeros(value string) string {
	value = strings.TrimLeft(value, "0")
	if value == "" {
		return "0"
	}

	return value
}

func allZero(value string) bool {
	for _, r := range value {
		if r != '0' {
			return false
		}
	}

	return true
}

func scaledInt(value string, scale int) (*big.Int, error) {
	parts, err := parseFixedScale(value, quantityPrecision, scale)
	if err != nil {
		return nil, err
	}
	digits := strings.ReplaceAll(parts.String(), ".", "")
	negative := strings.HasPrefix(digits, "-")
	digits = strings.TrimPrefix(digits, "-")
	intValue, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, ErrInvalidDecimal
	}
	if negative {
		intValue.Neg(intValue)
	}

	return intValue, nil
}

func scaledProductToDecimal(product *big.Int, productScale int, targetScale int, precision int) (Decimal, error) {
	if productScale < targetScale {
		return "", ErrInvalidDecimal
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(productScale-targetScale)), nil)
	quotient, remainder := new(big.Int).QuoRem(product, divisor, new(big.Int))
	absRemainder := new(big.Int).Abs(remainder)
	threshold := new(big.Int).Mul(absRemainder, big.NewInt(2))
	if threshold.Cmp(divisor) >= 0 {
		if product.Sign() < 0 {
			quotient.Sub(quotient, big.NewInt(1))
		} else {
			quotient.Add(quotient, big.NewInt(1))
		}
	}

	negative := quotient.Sign() < 0
	digits := new(big.Int).Abs(quotient).String()
	if targetScale > 0 && len(digits) <= targetScale {
		digits = strings.Repeat("0", targetScale-len(digits)+1) + digits
	}

	parts := decimalParts{negative: negative}
	if targetScale == 0 {
		parts.intPart = digits
	} else {
		parts.intPart = digits[:len(digits)-targetScale]
		parts.fracPart = digits[len(digits)-targetScale:]
	}

	return parts.toDecimal(precision, targetScale)
}

func scaledIntToDecimal(value *big.Int, scale int, precision int) (Decimal, error) {
	return scaledProductToDecimal(value, scale, scale, precision)
}

func must(value Decimal, err error) Decimal {
	if err != nil {
		panic(err)
	}

	return value
}
