import { apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import { decimalScales, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type {
  CreateSupplierInvoiceInput,
  FinanceSourceDocument,
  SupplierInvoice,
  SupplierInvoiceActionResult,
  SupplierInvoiceLine,
  SupplierInvoiceMatchStatus,
  SupplierInvoiceQuery,
  SupplierInvoiceStatus
} from "../types";

type SupplierInvoiceApi = {
  id: string;
  org_id?: string;
  invoice_no: string;
  supplier_id: string;
  supplier_code?: string;
  supplier_name: string;
  payable_id: string;
  payable_no: string;
  status: SupplierInvoiceStatus;
  match_status: SupplierInvoiceMatchStatus;
  source_document: FinanceSourceDocumentApi;
  lines?: SupplierInvoiceLineApi[];
  invoice_amount: string;
  expected_amount: string;
  variance_amount: string;
  currency_code: string;
  invoice_date: string;
  void_reason?: string;
  audit_log_id?: string;
  created_at: string;
  updated_at: string;
  version: number;
};

type SupplierInvoiceLineApi = {
  id: string;
  description: string;
  source_document: FinanceSourceDocumentApi;
  amount: string;
};

type FinanceSourceDocumentApi = {
  type: FinanceSourceDocument["type"];
  id?: string;
  no?: string;
};

type CreateSupplierInvoiceApiRequest = {
  id?: string;
  invoice_no?: string;
  supplier_id?: string;
  supplier_code?: string;
  supplier_name?: string;
  payable_id: string;
  invoice_date?: string;
  invoice_amount: string;
  currency_code: string;
};

type SupplierInvoiceActionApiResult = {
  supplier_invoice: SupplierInvoiceApi;
  previous_status: SupplierInvoiceStatus;
  current_status: SupplierInvoiceStatus;
  audit_log_id?: string;
};

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-05-05T10:00:00Z";

let prototypeSupplierInvoices = createPrototypeSupplierInvoices();

export const supplierInvoiceStatusOptions: { label: string; value: "" | SupplierInvoiceStatus }[] = [
  { label: "Tất cả trạng thái", value: "" },
  { label: "Đã khớp", value: "matched" },
  { label: "Lệch đối chiếu", value: "mismatch" },
  { label: "Nháp", value: "draft" },
  { label: "Đã hủy", value: "void" }
];

export async function getSupplierInvoices(query: SupplierInvoiceQuery = {}): Promise<SupplierInvoice[]> {
  try {
    const invoices = await apiGetRaw<SupplierInvoiceApi[]>(`/supplier-invoices${supplierInvoiceQueryString(query)}`, {
      accessToken: defaultAccessToken
    });

    return invoices.map(fromApiSupplierInvoice);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeSupplierInvoices(query);
  }
}

export async function getSupplierInvoice(id: string): Promise<SupplierInvoice> {
  try {
    const invoice = await apiGetRaw<SupplierInvoiceApi>(`/supplier-invoices/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiSupplierInvoice(invoice);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return getPrototypeSupplierInvoice(id);
  }
}

export async function createSupplierInvoice(input: CreateSupplierInvoiceInput): Promise<SupplierInvoice> {
  const body: CreateSupplierInvoiceApiRequest = {
    id: input.id,
    invoice_no: input.invoiceNo,
    supplier_id: input.supplierId,
    supplier_code: input.supplierCode,
    supplier_name: input.supplierName,
    payable_id: input.payableId,
    invoice_date: input.invoiceDate,
    invoice_amount: normalizeDecimalInput(input.invoiceAmount, decimalScales.money),
    currency_code: input.currencyCode
  };

  try {
    const invoice = await apiPost<SupplierInvoiceApi, CreateSupplierInvoiceApiRequest>("/supplier-invoices", body, {
      accessToken: defaultAccessToken
    });

    return fromApiSupplierInvoice(invoice);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeSupplierInvoice(body);
  }
}

export async function voidSupplierInvoice(id: string, reason: string): Promise<SupplierInvoiceActionResult> {
  try {
    const result = await apiPost<SupplierInvoiceActionApiResult, { reason: string }>(
      `/supplier-invoices/${encodeURIComponent(id)}/void`,
      { reason },
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return voidPrototypeSupplierInvoice(id, reason);
  }
}

export function resetPrototypeSupplierInvoicesForTest() {
  prototypeSupplierInvoices = createPrototypeSupplierInvoices();
}

export function supplierInvoiceStatusTone(
  status: SupplierInvoiceStatus
): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "matched":
      return "success";
    case "mismatch":
      return "warning";
    case "void":
      return "danger";
    case "draft":
    default:
      return "normal";
  }
}

export function formatSupplierInvoiceStatus(status: SupplierInvoiceStatus) {
  switch (status) {
    case "matched":
      return "Đã khớp";
    case "mismatch":
      return "Lệch đối chiếu";
    case "void":
      return "Đã hủy";
    case "draft":
    default:
      return "Nháp";
  }
}

function supplierInvoiceQueryString(query: SupplierInvoiceQuery) {
  const params = new URLSearchParams();
  if (query.search) {
    params.set("q", query.search);
  }
  if (query.status) {
    params.set("status", query.status);
  }
  if (query.supplierId) {
    params.set("supplier_id", query.supplierId);
  }
  if (query.payableId) {
    params.set("payable_id", query.payableId);
  }
  const value = params.toString();
  return value ? `?${value}` : "";
}

function fromApiSupplierInvoice(invoice: SupplierInvoiceApi): SupplierInvoice {
  return {
    id: invoice.id,
    orgId: invoice.org_id,
    invoiceNo: invoice.invoice_no,
    supplierId: invoice.supplier_id,
    supplierCode: invoice.supplier_code,
    supplierName: invoice.supplier_name,
    payableId: invoice.payable_id,
    payableNo: invoice.payable_no,
    status: invoice.status,
    matchStatus: invoice.match_status,
    sourceDocument: fromApiSourceDocument(invoice.source_document),
    lines: (invoice.lines ?? []).map(fromApiSupplierInvoiceLine),
    invoiceAmount: invoice.invoice_amount,
    expectedAmount: invoice.expected_amount,
    varianceAmount: invoice.variance_amount,
    currencyCode: invoice.currency_code,
    invoiceDate: invoice.invoice_date,
    voidReason: invoice.void_reason,
    auditLogId: invoice.audit_log_id,
    createdAt: invoice.created_at,
    updatedAt: invoice.updated_at,
    version: invoice.version
  };
}

function fromApiSupplierInvoiceLine(line: SupplierInvoiceLineApi): SupplierInvoiceLine {
  return {
    id: line.id,
    description: line.description,
    sourceDocument: fromApiSourceDocument(line.source_document),
    amount: line.amount
  };
}

function fromApiSourceDocument(source: FinanceSourceDocumentApi): FinanceSourceDocument {
  return {
    type: source.type,
    id: source.id,
    no: source.no
  };
}

function fromApiActionResult(result: SupplierInvoiceActionApiResult): SupplierInvoiceActionResult {
  return {
    supplierInvoice: fromApiSupplierInvoice(result.supplier_invoice),
    previousStatus: result.previous_status,
    currentStatus: result.current_status,
    auditLogId: result.audit_log_id
  };
}

function filterPrototypeSupplierInvoices(query: SupplierInvoiceQuery) {
  return prototypeSupplierInvoices
    .filter((invoice) => matchesSupplierInvoiceQuery(invoice, query))
    .map(cloneSupplierInvoice);
}

function getPrototypeSupplierInvoice(id: string) {
  const invoice = prototypeSupplierInvoices.find((candidate) => candidate.id === id);
  if (!invoice) {
    throw new Error("Supplier invoice not found");
  }

  return cloneSupplierInvoice(invoice);
}

function createPrototypeSupplierInvoice(input: CreateSupplierInvoiceApiRequest) {
  const expectedAmount = input.payable_id === "ap-supplier-260430-0001" ? "4250000.00" : input.invoice_amount;
  const invoiceAmount = normalizeDecimalInput(input.invoice_amount, decimalScales.money);
  const varianceAmount = normalizeDecimalInput(Number(invoiceAmount) - Number(expectedAmount), decimalScales.money);
  const status: SupplierInvoiceStatus = Number(varianceAmount) === 0 ? "matched" : "mismatch";
  const lines: SupplierInvoiceLine[] = [
    {
      id: "prototype-si-line-1",
      description: "Supplier invoice line",
      sourceDocument: { type: "purchase_order", id: "po-260430-0001", no: "PO-260430-0001" },
      amount: expectedAmount
    }
  ];
  if (Number(varianceAmount) !== 0) {
    lines.push({
      id: "prototype-si-variance",
      description: "Supplier invoice variance against AP",
      sourceDocument: { type: "supplier_payable", id: input.payable_id, no: input.payable_id },
      amount: varianceAmount
    });
  }
  const invoice = createSupplierInvoiceSeed({
    id: input.id || `si-${Date.now()}`,
    invoiceNo: input.invoice_no || `SI-${Date.now()}`,
    supplierId: input.supplier_id || "supplier-hcm-001",
    supplierCode: input.supplier_code || "SUP-HCM-001",
    supplierName: input.supplier_name || "Nguyen Lieu HCM",
    payableId: input.payable_id,
    payableNo: input.payable_id === "ap-supplier-260430-0001" ? "AP-SUP-260430-0001" : input.payable_id,
    status,
    matchStatus: status === "matched" ? "matched" : "mismatch",
    invoiceAmount,
    expectedAmount,
    varianceAmount,
    invoiceDate: input.invoice_date || "2026-05-05",
    sourceDocument: { type: "warehouse_receipt", id: "gr-260430-0001", no: "GR-260430-0001" },
    lines
  });
  prototypeSupplierInvoices = [invoice, ...prototypeSupplierInvoices.filter((candidate) => candidate.id !== invoice.id)];

  return cloneSupplierInvoice(invoice);
}

function voidPrototypeSupplierInvoice(id: string, reason: string): SupplierInvoiceActionResult {
  const current = getPrototypeSupplierInvoice(id);
  if (current.status === "void") {
    throw new Error("Supplier invoice is already void");
  }
  const next: SupplierInvoice = {
    ...current,
    status: "void",
    voidReason: reason,
    updatedAt: prototypeNow,
    version: current.version + 1,
    auditLogId: `audit-si-void-${current.id}`
  };
  prototypeSupplierInvoices = [next, ...prototypeSupplierInvoices.filter((invoice) => invoice.id !== id)];

  return {
    supplierInvoice: cloneSupplierInvoice(next),
    previousStatus: current.status,
    currentStatus: next.status,
    auditLogId: next.auditLogId
  };
}

function matchesSupplierInvoiceQuery(invoice: SupplierInvoice, query: SupplierInvoiceQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || invoice.status === query.status) &&
    (!query.supplierId || invoice.supplierId === query.supplierId) &&
    (!query.payableId || invoice.payableId === query.payableId) &&
    (!search ||
      [
        invoice.invoiceNo,
        invoice.supplierCode,
        invoice.supplierName,
        invoice.payableId,
        invoice.payableNo,
        invoice.sourceDocument.id,
        invoice.sourceDocument.no,
        ...invoice.lines.flatMap((line) => [line.description, line.sourceDocument.id, line.sourceDocument.no])
      ]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function createPrototypeSupplierInvoices(): SupplierInvoice[] {
  return [
    createSupplierInvoiceSeed({
      id: "si-supplier-260430-0001",
      invoiceNo: "INV-SUP-260430-0001",
      supplierId: "supplier-hcm-001",
      supplierCode: "SUP-HCM-001",
      supplierName: "Nguyen Lieu HCM",
      payableId: "ap-supplier-260430-0001",
      payableNo: "AP-SUP-260430-0001",
      status: "matched",
      matchStatus: "matched",
      invoiceAmount: "4250000.00",
      expectedAmount: "4250000.00",
      varianceAmount: "0.00",
      invoiceDate: "2026-05-05",
      sourceDocument: { type: "warehouse_receipt", id: "gr-260430-0001", no: "GR-260430-0001" },
      lines: [
        {
          id: "si-supplier-260430-0001-line-1",
          description: "Accepted raw material after inbound QC",
          sourceDocument: { type: "warehouse_receipt", id: "gr-260430-0001", no: "GR-260430-0001" },
          amount: "3000000.00"
        },
        {
          id: "si-supplier-260430-0001-line-2",
          description: "Accepted packaging after inbound QC",
          sourceDocument: { type: "purchase_order", id: "po-260430-0001", no: "PO-260430-0001" },
          amount: "1250000.00"
        }
      ]
    })
  ];
}

function createSupplierInvoiceSeed(input: {
  id: string;
  invoiceNo: string;
  supplierId: string;
  supplierCode: string;
  supplierName: string;
  payableId: string;
  payableNo: string;
  status: SupplierInvoiceStatus;
  matchStatus: SupplierInvoiceMatchStatus;
  invoiceAmount: string;
  expectedAmount: string;
  varianceAmount: string;
  invoiceDate: string;
  sourceDocument: FinanceSourceDocument;
  lines: SupplierInvoiceLine[];
}): SupplierInvoice {
  return {
    id: input.id,
    orgId: "org-my-pham",
    invoiceNo: input.invoiceNo,
    supplierId: input.supplierId,
    supplierCode: input.supplierCode,
    supplierName: input.supplierName,
    payableId: input.payableId,
    payableNo: input.payableNo,
    status: input.status,
    matchStatus: input.matchStatus,
    sourceDocument: input.sourceDocument,
    lines: input.lines,
    invoiceAmount: normalizeDecimalInput(input.invoiceAmount, decimalScales.money),
    expectedAmount: normalizeDecimalInput(input.expectedAmount, decimalScales.money),
    varianceAmount: normalizeDecimalInput(input.varianceAmount, decimalScales.money),
    currencyCode: "VND",
    invoiceDate: input.invoiceDate,
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 1
  };
}

function cloneSupplierInvoice(invoice: SupplierInvoice): SupplierInvoice {
  return {
    ...invoice,
    sourceDocument: { ...invoice.sourceDocument },
    lines: invoice.lines.map((line) => ({
      ...line,
      sourceDocument: { ...line.sourceDocument }
    }))
  };
}
