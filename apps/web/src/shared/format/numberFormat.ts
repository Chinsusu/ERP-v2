export const erpNumberLocale = "vi-VN";
export const erpTimezone = "Asia/Ho_Chi_Minh";
export const erpCurrencyCode = "VND";

export const decimalScales = {
  money: 2,
  unitPrice: 4,
  unitCost: 6,
  quantity: 6,
  rate: 4
} as const;

export type DecimalScale = number;

export function normalizeDecimalInput(value: string | number | null | undefined, scale: DecimalScale): string {
  const normalized = normalizeDecimalSeparators(value);
  const rounded = roundDecimalString(normalized, scale);
  return padDecimalScale(rounded, scale);
}

export function normalizeCurrencyCode(value: string | null | undefined): string {
  const code = String(value ?? erpCurrencyCode)
    .trim()
    .toUpperCase();
  if (!/^[A-Z]{3}$/.test(code)) {
    throw new Error("Currency code is invalid");
  }

  return code;
}

export function normalizeUOMCode(value: string | null | undefined): string {
  const code = String(value ?? "")
    .trim()
    .toUpperCase();
  if (!/^[A-Z0-9_-]{1,20}$/.test(code)) {
    throw new Error("UOM code is invalid");
  }

  return code;
}

export function isNegativeDecimal(value: string | number | null | undefined) {
  return normalizeDecimalSeparators(value).startsWith("-");
}

export function formatMoney(value: string | number | null | undefined, currencyCode = erpCurrencyCode) {
  const decimal = normalizeDecimalInput(value, decimalScales.money);
  return new Intl.NumberFormat(erpNumberLocale, {
    style: "currency",
    currency: normalizeCurrencyCode(currencyCode),
    maximumFractionDigits: 0,
    minimumFractionDigits: 0
  })
    .format(Number(decimal))
    .replace(/\u00a0/g, " ");
}

export function formatQuantity(value: string | number | null | undefined, uomCode?: string) {
  const decimal = normalizeDecimalInput(value, decimalScales.quantity);
  const quantity = formatDecimal(decimal, decimalScales.quantity);
  const uom = uomCode ? ` ${normalizeUOMCode(uomCode)}` : "";

  return `${quantity}${uom}`;
}

export function formatRate(value: string | number | null | undefined) {
  const decimal = normalizeDecimalInput(value, decimalScales.rate);
  return `${formatDecimal(decimal, decimalScales.rate)}%`;
}

export function formatDateVI(value: string | Date) {
  return new Intl.DateTimeFormat(erpNumberLocale, {
    timeZone: erpTimezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit"
  }).format(new Date(value));
}

export function formatDateTimeVI(value: string | Date) {
  return new Intl.DateTimeFormat(erpNumberLocale, {
    timeZone: erpTimezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

function formatDecimal(value: string, maxFractionDigits: DecimalScale) {
  return new Intl.NumberFormat(erpNumberLocale, {
    minimumFractionDigits: 0,
    maximumFractionDigits: maxFractionDigits
  }).format(Number(value));
}

function normalizeDecimalSeparators(value: string | number | null | undefined) {
  let raw = String(value ?? "0")
    .trim()
    .replaceAll(" ", "")
    .replaceAll("₫", "");
  if (raw === "") {
    raw = "0";
  }

  const commaCount = count(raw, ",");
  const dotCount = count(raw, ".");
  if (commaCount > 1) {
    throw new Error("Decimal value is invalid");
  }
  if (commaCount === 1) {
    raw = raw.replaceAll(".", "").replace(",", ".");
  } else if (dotCount > 1) {
    raw = raw.replaceAll(".", "");
  }

  if (!/^-?[0-9]+(\.[0-9]+)?$/.test(raw)) {
    throw new Error("Decimal value is invalid");
  }

  return raw;
}

function roundDecimalString(value: string, scale: DecimalScale) {
  const negative = value.startsWith("-");
  const unsigned = negative ? value.slice(1) : value;
  const [integerPart, fraction = ""] = unsigned.split(".");
  if (fraction.length <= scale) {
    return `${negative && !isAllZero(integerPart + fraction) ? "-" : ""}${integerPart}${fraction ? `.${fraction}` : ""}`;
  }

  const kept = fraction.slice(0, scale);
  const roundUp = Number(fraction[scale]) >= 5;
  let digits = `${integerPart}${kept}`;
  if (roundUp) {
    digits = addOne(digits);
  }
  if (scale > 0 && digits.length <= scale) {
    digits = `${"0".repeat(scale - digits.length + 1)}${digits}`;
  }
  const roundedInteger = scale === 0 ? digits : digits.slice(0, -scale);
  const roundedFraction = scale === 0 ? "" : digits.slice(-scale);
  const result = `${roundedInteger}${roundedFraction ? `.${roundedFraction}` : ""}`;

  return `${negative && !isAllZero(result.replace(".", "")) ? "-" : ""}${result}`;
}

function padDecimalScale(value: string, scale: DecimalScale) {
  const [integerPart, fraction = ""] = value.split(".");
  if (scale === 0) {
    return integerPart;
  }

  return `${integerPart}.${fraction.padEnd(scale, "0")}`;
}

function addOne(value: string) {
  const digits = value.split("");
  let carry = 1;
  for (let index = digits.length - 1; index >= 0; index -= 1) {
    const next = Number(digits[index]) + carry;
    digits[index] = String(next % 10);
    carry = next >= 10 ? 1 : 0;
    if (carry === 0) {
      break;
    }
  }
  if (carry === 1) {
    digits.unshift("1");
  }

  return digits.join("");
}

function isAllZero(value: string) {
  return /^[0.]*$/.test(value);
}

function count(value: string, pattern: string) {
  return value.split(pattern).length - 1;
}
