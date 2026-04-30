import { ApiError, apiGet, apiGetRaw, apiPost } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import {
  decimalScales,
  formatDateVI,
  formatMoney,
  normalizeDecimalInput
} from "../../../shared/format/numberFormat";
import type {
  CreateCustomerReceivableInput,
  CustomerReceivable,
  CustomerReceivableActionResult,
  CustomerReceivableLine,
  CustomerReceivableQuery,
  CustomerReceivableStatus,
  FinanceSourceDocument
} from "../types";

type CustomerReceivableApi = components["schemas"]["CustomerReceivable"];
type CustomerReceivableListItemApi = components["schemas"]["CustomerReceivableListItem"];
type CustomerReceivableLineApi = components["schemas"]["CustomerReceivableLine"];
type CreateCustomerReceivableApiRequest = components["schemas"]["CreateCustomerReceivableRequest"];
type CustomerReceivableActionApiRequest = components["schemas"]["CustomerReceivableActionRequest"];
type CustomerReceivableActionApiResult = components["schemas"]["CustomerReceivableActionResult"];
type CustomerReceivableListApiQuery = operations["listCustomerReceivables"]["parameters"]["query"];

type CustomerOption = {
  label: string;
  value: string;
  code: string;
  name: string;
};

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-30T08:00:00Z";
let receivableSequence = 20;

export const customerReceivableStatusOptions: { label: string; value: "" | CustomerReceivableStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Open", value: "open" },
  { label: "Partially paid", value: "partially_paid" },
  { label: "Paid", value: "paid" },
  { label: "Disputed", value: "disputed" },
  { label: "Void", value: "void" }
];

export const financeCustomerOptions: CustomerOption[] = [
  { label: "KH-HCM-001 / My Pham HCM Retail", value: "customer-hcm-001", code: "KH-HCM-001", name: "My Pham HCM Retail" },
  { label: "KH-MKT-022 / Marketplace COD", value: "customer-marketplace-022", code: "KH-MKT-022", name: "Marketplace COD" },
  { label: "KH-HN-018 / My Pham Hanoi Wholesale", value: "customer-hn-018", code: "KH-HN-018", name: "My Pham Hanoi Wholesale" }
];

let prototypeCustomerReceivables = createPrototypeCustomerReceivables();

export async function getCustomerReceivables(query: CustomerReceivableQuery = {}): Promise<CustomerReceivable[]> {
  try {
    const receivables = await apiGet("/customer-receivables", {
      accessToken: defaultAccessToken,
      query: toApiCustomerReceivableQuery(query)
    });

    return receivables.map(fromApiCustomerReceivableListItem);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeCustomerReceivables(query);
  }
}

export async function getCustomerReceivable(id: string): Promise<CustomerReceivable> {
  try {
    const receivable = await apiGetRaw<CustomerReceivableApi>(`/customer-receivables/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiCustomerReceivable(receivable);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return getPrototypeCustomerReceivable(id);
  }
}

export async function createCustomerReceivable(input: CreateCustomerReceivableInput): Promise<CustomerReceivable> {
  try {
    const receivable = await apiPost<CustomerReceivableApi, CreateCustomerReceivableApiRequest>(
      "/customer-receivables",
      toApiCreateInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiCustomerReceivable(receivable);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeCustomerReceivable(input);
  }
}

export async function recordCustomerReceivableReceipt(
  id: string,
  amount: string
): Promise<CustomerReceivableActionResult> {
  return runCustomerReceivableAction(id, "record-receipt", { amount: normalizeDecimalInput(amount, decimalScales.money) });
}

export async function markCustomerReceivableDisputed(
  id: string,
  reason: string
): Promise<CustomerReceivableActionResult> {
  return runCustomerReceivableAction(id, "mark-disputed", { reason });
}

export async function voidCustomerReceivable(id: string, reason: string): Promise<CustomerReceivableActionResult> {
  return runCustomerReceivableAction(id, "void", { reason });
}

export function resetPrototypeCustomerReceivablesForTest() {
  receivableSequence = 20;
  prototypeCustomerReceivables = createPrototypeCustomerReceivables();
}

export function customerReceivableStatusTone(
  status: CustomerReceivableStatus
): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "paid":
      return "success";
    case "partially_paid":
      return "info";
    case "disputed":
      return "warning";
    case "void":
      return "danger";
    case "draft":
      return "normal";
    case "open":
    default:
      return "warning";
  }
}

export function formatCustomerReceivableStatus(status: CustomerReceivableStatus) {
  return status
    .split("_")
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}

export function formatFinanceMoney(value: string | number, currencyCode = "VND") {
  return formatMoney(value, currencyCode);
}

export function formatFinanceDate(value?: string) {
  return value ? formatDateVI(value) : "-";
}

export function canRecordCustomerReceipt(receivable: CustomerReceivable) {
  return (
    Number(receivable.outstandingAmount) > 0 &&
    ["open", "partially_paid", "disputed"].includes(receivable.status)
  );
}

async function runCustomerReceivableAction(
  id: string,
  action: "record-receipt" | "mark-disputed" | "void",
  body: CustomerReceivableActionApiRequest
): Promise<CustomerReceivableActionResult> {
  try {
    const result = await apiPost<CustomerReceivableActionApiResult, CustomerReceivableActionApiRequest>(
      `/customer-receivables/${encodeURIComponent(id)}/${action}`,
      body,
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypeCustomerReceivable(id, action, body);
  }
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function toApiCustomerReceivableQuery(query: CustomerReceivableQuery): CustomerReceivableListApiQuery {
  return {
    q: query.search,
    status: query.status,
    customer_id: query.customerId
  };
}

function fromApiCustomerReceivableListItem(item: CustomerReceivableListItemApi): CustomerReceivable {
  return {
    id: item.id,
    receivableNo: item.receivable_no,
    customerId: item.customer_id,
    customerCode: item.customer_code,
    customerName: item.customer_name,
    status: item.status,
    lines: [],
    totalAmount: item.total_amount,
    paidAmount: item.paid_amount,
    outstandingAmount: item.outstanding_amount,
    currencyCode: item.currency_code,
    dueDate: item.due_date,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    version: item.version
  };
}

function fromApiCustomerReceivable(receivable: CustomerReceivableApi): CustomerReceivable {
  return {
    id: receivable.id,
    orgId: receivable.org_id,
    receivableNo: receivable.receivable_no,
    customerId: receivable.customer_id,
    customerCode: receivable.customer_code,
    customerName: receivable.customer_name,
    status: receivable.status,
    sourceDocument: fromApiSourceDocument(receivable.source_document),
    lines: receivable.lines.map(fromApiCustomerReceivableLine),
    totalAmount: receivable.total_amount,
    paidAmount: receivable.paid_amount,
    outstandingAmount: receivable.outstanding_amount,
    currencyCode: receivable.currency_code,
    dueDate: receivable.due_date,
    disputeReason: receivable.dispute_reason,
    voidReason: receivable.void_reason,
    auditLogId: receivable.audit_log_id,
    createdAt: receivable.created_at,
    updatedAt: receivable.updated_at,
    version: receivable.version
  };
}

function fromApiCustomerReceivableLine(line: CustomerReceivableLineApi): CustomerReceivableLine {
  return {
    id: line.id,
    description: line.description,
    sourceDocument: fromApiSourceDocument(line.source_document),
    amount: line.amount
  };
}

function fromApiSourceDocument(source: components["schemas"]["FinanceSourceDocument"]): FinanceSourceDocument {
  return {
    type: source.type,
    id: source.id,
    no: source.no
  };
}

function fromApiActionResult(result: CustomerReceivableActionApiResult): CustomerReceivableActionResult {
  return {
    customerReceivable: fromApiCustomerReceivable(result.customer_receivable),
    previousStatus: result.previous_status,
    currentStatus: result.current_status,
    auditLogId: result.audit_log_id
  };
}

function toApiCreateInput(input: CreateCustomerReceivableInput): CreateCustomerReceivableApiRequest {
  return {
    id: input.id,
    receivable_no: input.receivableNo,
    customer_id: input.customerId,
    customer_code: input.customerCode,
    customer_name: input.customerName,
    status: input.status,
    source_document: input.sourceDocument,
    lines: input.lines.map((line) => ({
      id: line.id,
      description: line.description,
      source_document: line.sourceDocument,
      amount: normalizeDecimalInput(line.amount, decimalScales.money)
    })),
    total_amount: normalizeDecimalInput(input.totalAmount, decimalScales.money),
    currency_code: input.currencyCode,
    due_date: input.dueDate
  };
}

function filterPrototypeCustomerReceivables(query: CustomerReceivableQuery) {
  return prototypeCustomerReceivables
    .filter((receivable) => matchesCustomerReceivableQuery(receivable, query))
    .map(cloneCustomerReceivable);
}

function getPrototypeCustomerReceivable(id: string) {
  const receivable = prototypeCustomerReceivables.find((candidate) => candidate.id === id);
  if (!receivable) {
    throw new Error("Customer receivable not found");
  }

  return cloneCustomerReceivable(receivable);
}

function createPrototypeCustomerReceivable(input: CreateCustomerReceivableInput): CustomerReceivable {
  if (input.lines.length === 0) {
    throw new Error("At least one AR line is required");
  }

  receivableSequence += 1;
  const id = input.id ?? `ar-ui-${receivableSequence}`;
  const totalAmount = normalizeDecimalInput(input.totalAmount, decimalScales.money);
  const receivable: CustomerReceivable = {
    id,
    orgId: "org-my-pham",
    receivableNo: input.receivableNo ?? `AR-260430-${String(receivableSequence).padStart(4, "0")}`,
    customerId: input.customerId,
    customerCode: input.customerCode,
    customerName: input.customerName,
    status: input.status ?? "open",
    sourceDocument: input.sourceDocument,
    lines: input.lines.map((line) => ({
      id: line.id,
      description: line.description,
      sourceDocument: line.sourceDocument,
      amount: normalizeDecimalInput(line.amount, decimalScales.money)
    })),
    totalAmount,
    paidAmount: "0.00",
    outstandingAmount: totalAmount,
    currencyCode: input.currencyCode,
    dueDate: input.dueDate,
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 1
  };

  prototypeCustomerReceivables = [receivable, ...prototypeCustomerReceivables];
  return cloneCustomerReceivable(receivable);
}

function transitionPrototypeCustomerReceivable(
  id: string,
  action: "record-receipt" | "mark-disputed" | "void",
  body: CustomerReceivableActionApiRequest
): CustomerReceivableActionResult {
  const current = getPrototypeCustomerReceivable(id);
  const previousStatus = current.status;
  let next = cloneCustomerReceivable(current);

  if (action === "record-receipt") {
    const amount = normalizeDecimalInput(body.amount, decimalScales.money);
    const amountValue = Number(amount);
    if (!canRecordCustomerReceipt(next) || amountValue <= 0 || amountValue > Number(next.outstandingAmount)) {
      throw new Error("Receipt amount is invalid for this AR");
    }

    const paidAmount = Number(next.paidAmount) + amountValue;
    const outstandingAmount = Math.max(0, Number(next.totalAmount) - paidAmount);
    next = {
      ...next,
      paidAmount: normalizeDecimalInput(paidAmount, decimalScales.money),
      outstandingAmount: normalizeDecimalInput(outstandingAmount, decimalScales.money),
      status: outstandingAmount === 0 ? "paid" : "partially_paid"
    };
  } else if (action === "mark-disputed") {
    if (!["open", "partially_paid"].includes(next.status)) {
      throw new Error("Only open AR can be disputed");
    }
    next = { ...next, status: "disputed", disputeReason: body.reason?.trim() || "Disputed by finance" };
  } else {
    if (next.status === "paid" || next.status === "void") {
      throw new Error("Paid or void AR cannot be voided");
    }
    next = { ...next, status: "void", outstandingAmount: "0.00", voidReason: body.reason?.trim() || "Voided by finance" };
  }

  next = {
    ...next,
    updatedAt: prototypeNow,
    version: next.version + 1,
    auditLogId: `audit-${action}-${next.id}`
  };
  prototypeCustomerReceivables = [
    next,
    ...prototypeCustomerReceivables.filter((receivable) => receivable.id !== next.id)
  ];

  return {
    customerReceivable: cloneCustomerReceivable(next),
    previousStatus,
    currentStatus: next.status,
    auditLogId: next.auditLogId
  };
}

function matchesCustomerReceivableQuery(receivable: CustomerReceivable, query: CustomerReceivableQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || receivable.status === query.status) &&
    (!query.customerId || receivable.customerId === query.customerId) &&
    (!search ||
      [receivable.receivableNo, receivable.customerCode, receivable.customerName, receivable.sourceDocument?.no]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function createPrototypeCustomerReceivables(): CustomerReceivable[] {
  return [
    createReceivableSeed({
      id: "ar-cod-260430-0001",
      receivableNo: "AR-COD-260430-0001",
      customerId: "customer-hcm-001",
      customerCode: "KH-HCM-001",
      customerName: "My Pham HCM Retail",
      status: "open",
      totalAmount: "1250000.00",
      paidAmount: "0.00",
      sourceDocument: { type: "shipment", id: "ship-hcm-260430-0001", no: "GHN260430001" },
      dueDate: "2026-05-01",
      lines: [
        {
          id: "ar-cod-260430-0001-line-1",
          description: "COD receivable for SO-260430-0001",
          sourceDocument: { type: "sales_order", id: "so-260430-0001", no: "SO-260430-0001" },
          amount: "1250000.00"
        }
      ]
    }),
    createReceivableSeed({
      id: "ar-cod-260430-0002",
      receivableNo: "AR-COD-260430-0002",
      customerId: "customer-marketplace-022",
      customerCode: "KH-MKT-022",
      customerName: "Marketplace COD",
      status: "partially_paid",
      totalAmount: "2480000.00",
      paidAmount: "1000000.00",
      sourceDocument: { type: "shipment", id: "ship-mkt-260430-0002", no: "SPX260430122" },
      dueDate: "2026-05-02",
      lines: [
        {
          id: "ar-cod-260430-0002-line-1",
          description: "COD receivable for marketplace batch",
          sourceDocument: { type: "sales_order", id: "so-mkt-260430-003", no: "SO-MKT-260430-003" },
          amount: "2480000.00"
        }
      ]
    }),
    createReceivableSeed({
      id: "ar-manual-260430-0003",
      receivableNo: "AR-MANUAL-260430-0003",
      customerId: "customer-hn-018",
      customerCode: "KH-HN-018",
      customerName: "My Pham Hanoi Wholesale",
      status: "disputed",
      totalAmount: "3200000.00",
      paidAmount: "0.00",
      sourceDocument: { type: "manual_adjustment", id: "manual-ar-260430-0003", no: "MANUAL-AR-260430-0003" },
      dueDate: "2026-05-03",
      disputeReason: "Customer disputed delivery shortage",
      lines: [
        {
          id: "ar-manual-260430-0003-line-1",
          description: "Manual AR for wholesale delivery shortage review",
          sourceDocument: { type: "sales_order", id: "so-hn-260429-007", no: "SO-HN-260429-007" },
          amount: "3200000.00"
        }
      ]
    })
  ];
}

function createReceivableSeed(input: {
  id: string;
  receivableNo: string;
  customerId: string;
  customerCode: string;
  customerName: string;
  status: CustomerReceivableStatus;
  totalAmount: string;
  paidAmount: string;
  sourceDocument: FinanceSourceDocument;
  dueDate: string;
  disputeReason?: string;
  lines: CustomerReceivableLine[];
}): CustomerReceivable {
  return {
    id: input.id,
    orgId: "org-my-pham",
    receivableNo: input.receivableNo,
    customerId: input.customerId,
    customerCode: input.customerCode,
    customerName: input.customerName,
    status: input.status,
    sourceDocument: input.sourceDocument,
    lines: input.lines,
    totalAmount: normalizeDecimalInput(input.totalAmount, decimalScales.money),
    paidAmount: normalizeDecimalInput(input.paidAmount, decimalScales.money),
    outstandingAmount: normalizeDecimalInput(Number(input.totalAmount) - Number(input.paidAmount), decimalScales.money),
    currencyCode: "VND",
    dueDate: input.dueDate,
    disputeReason: input.disputeReason,
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 1
  };
}

function cloneCustomerReceivable(receivable: CustomerReceivable): CustomerReceivable {
  return {
    ...receivable,
    sourceDocument: receivable.sourceDocument ? { ...receivable.sourceDocument } : undefined,
    lines: receivable.lines.map((line) => ({
      ...line,
      sourceDocument: { ...line.sourceDocument }
    }))
  };
}
