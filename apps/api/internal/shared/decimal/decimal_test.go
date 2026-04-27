package decimal

import (
	"errors"
	"testing"
)

func TestParseFixedScaleDecimals(t *testing.T) {
	tests := []struct {
		name  string
		parse func(string) (Decimal, error)
		input string
		want  Decimal
	}{
		{name: "money", parse: ParseMoneyAmount, input: "1250000", want: "1250000.00"},
		{name: "unit price", parse: ParseUnitPrice, input: "125000.5", want: "125000.5000"},
		{name: "unit cost", parse: ParseUnitCost, input: "64000", want: "64000.000000"},
		{name: "quantity", parse: ParseQuantity, input: "10.5", want: "10.500000"},
		{name: "rate", parse: ParseRate, input: "8", want: "8.0000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parse(tt.input)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got != tt.want {
				t.Fatalf("decimal = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseDecimalRejectsUnsafeInputs(t *testing.T) {
	tests := []string{
		"1,250.00",
		"1.234.56",
		"12.345",
		"VND 100",
		".5",
	}

	for _, input := range tests {
		if _, err := ParseMoneyAmount(input); !errors.Is(err, ErrInvalidDecimal) {
			t.Fatalf("ParseMoneyAmount(%q) error = %v, want invalid decimal", input, err)
		}
	}
}

func TestRoundFixedScaleUsesHalfUpRules(t *testing.T) {
	tests := []struct {
		name  string
		round func(string) (Decimal, error)
		input string
		want  Decimal
	}{
		{name: "money down", round: RoundMoneyAmount, input: "1250.124", want: "1250.12"},
		{name: "money up", round: RoundMoneyAmount, input: "1250.125", want: "1250.13"},
		{name: "quantity", round: RoundQuantity, input: "10.1234567", want: "10.123457"},
		{name: "rate", round: RoundRate, input: "8.12345", want: "8.1235"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.round(tt.input)
			if err != nil {
				t.Fatalf("round: %v", err)
			}
			if got != tt.want {
				t.Fatalf("decimal = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecimalCodeNormalization(t *testing.T) {
	currency, err := NormalizeCurrencyCode("")
	if err != nil {
		t.Fatalf("currency: %v", err)
	}
	if currency != CurrencyVND {
		t.Fatalf("currency = %q, want VND", currency)
	}

	uom, err := NormalizeUOMCode(" kg ")
	if err != nil {
		t.Fatalf("uom: %v", err)
	}
	if uom != "KG" {
		t.Fatalf("uom = %q, want KG", uom)
	}
}
