import { apiGet, apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type { components, operations } from "../../../shared/api/generated/schema";
import { decimalScales, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type {
  CashAllocationTargetType,
  CashTransaction,
  CashTransactionAllocation,
  CashTransactionDirection,
  CashTransactionQuery,
  CashTransactionStatus,
  CreateCashTransactionInput
} from "../types";

type CashTransactionApi = components["schemas"]["CashTransaction"];
type CashTransactionListItemApi = components["schemas"]["CashTransactionListItem"];
type CashTransactionAllocationApi = components["schemas"]["CashTransactionAllocation"];
type CreateCashTransactionApiRequest = components["schemas"]["CreateCashTransactionRequest"];
type CashTransactionListApiQuery = operations["listCashTransactions"]["parameters"]["query"];

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-30T10:30:00Z";

let prototypeCashTransactions = createPrototypeCashTransactions();

export const cashTransactionStatusOptions: { label: string; value: "" | CashTransactionStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Posted", value: "posted" },
  { label: "Void", value: "void" }
];

export const cashTransactionDirectionOptions: { label: string; value: "" | CashTransactionDirection }[] = [
  { label: "All directions", value: "" },
  { label: "Cash in", value: "cash_in" },
  { label: "Cash out", value: "cash_out" }
];

export const cashAllocationTargetOptions: { label: string; value: CashAllocationTargetType }[] = [
  { label: "Customer receivable", value: "customer_receivable" },
  { label: "Supplier payable", value: "supplier_payable" },
  { label: "COD remittance", value: "cod_remittance" },
  { label: "Payment request", value: "payment_request" },
  { label: "Manual adjustment", value: "manual_adjustment" }
];

export async function getCashTransactions(query: CashTransactionQuery = {}): Promise<CashTransaction[]> {
  try {
    const transactions = await apiGet("/cash-transactions", {
      accessToken: defaultAccessToken,
      query: toApiCashTransactionQuery(query)
    });

    return transactions.map(fromApiCashTransactionListItem);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeCashTransactions(query);
  }
}

export async function getCashTransaction(id: string): Promise<CashTransaction> {
  try {
    const transaction = await apiGetRaw<CashTransactionApi>(`/cash-transactions/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiCashTransaction(transaction);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return getPrototypeCashTransaction(id);
  }
}

export async function createCashTransaction(input: CreateCashTransactionInput): Promise<CashTransaction> {
  const payload = toApiCreateCashTransaction(input);
  try {
    const transaction = await apiPost<CashTransactionApi, CreateCashTransactionApiRequest>("/cash-transactions", payload, {
      accessToken: defaultAccessToken
    });

    return fromApiCashTransaction(transaction);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return createPrototypeCashTransaction(input);
  }
}

export function resetPrototypeCashTransactionsForTest() {
  prototypeCashTransactions = createPrototypeCashTransactions();
}

export function cashTransactionStatusTone(status: CashTransactionStatus): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "posted":
      return "success";
    case "void":
      return "danger";
    case "draft":
    default:
      return "warning";
  }
}

export function cashTransactionDirectionTone(direction: CashTransactionDirection): "success" | "warning" | "danger" | "info" | "normal" {
  return direction === "cash_in" ? "success" : "warning";
}

export function formatCashTransactionStatus(status: CashTransactionStatus) {
  return status.slice(0, 1).toUpperCase() + status.slice(1);
}

export function formatCashTransactionDirection(direction: CashTransactionDirection) {
  return direction === "cash_in" ? "Cash in" : "Cash out";
}

export function cashAllocationTargetAllowedForDirection(
  direction: CashTransactionDirection,
  targetType: CashAllocationTargetType
) {
  if (direction === "cash_in") {
    return ["customer_receivable", "cod_remittance", "manual_adjustment"].includes(targetType);
  }

  return ["supplier_payable", "payment_request", "manual_adjustment"].includes(targetType);
}


function toApiCashTransactionQuery(query: CashTransactionQuery): CashTransactionListApiQuery {
  return {
    q: query.search,
    status: query.status,
    direction: query.direction,
    counterparty_id: query.counterpartyId
  };
}

function toApiCreateCashTransaction(input: CreateCashTransactionInput): CreateCashTransactionApiRequest {
  return {
    id: input.id,
    transaction_no: input.transactionNo,
    direction: input.direction,
    business_date: input.businessDate,
    counterparty_id: input.counterpartyId,
    counterparty_name: input.counterpartyName.trim(),
    payment_method: input.paymentMethod.trim(),
    reference_no: input.referenceNo?.trim(),
    allocations: input.allocations.map((allocation) => ({
      id: allocation.id,
      target_type: allocation.targetType,
      target_id: allocation.targetId.trim(),
      target_no: allocation.targetNo.trim(),
      amount: normalizeDecimalInput(allocation.amount, decimalScales.money)
    })),
    total_amount: normalizeDecimalInput(input.totalAmount, decimalScales.money),
    currency_code: input.currencyCode,
    memo: input.memo?.trim()
  };
}

function fromApiCashTransactionListItem(item: CashTransactionListItemApi): CashTransaction {
  return {
    id: item.id,
    transactionNo: item.transaction_no,
    direction: item.direction,
    status: item.status,
    businessDate: item.business_date,
    counterpartyId: item.counterparty_id,
    counterpartyName: item.counterparty_name,
    paymentMethod: item.payment_method,
    referenceNo: item.reference_no,
    allocations: [],
    totalAmount: item.total_amount,
    currencyCode: item.currency_code,
    postedBy: item.posted_by,
    postedAt: item.posted_at,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    version: item.version
  };
}

function fromApiCashTransaction(transaction: CashTransactionApi): CashTransaction {
  return {
    id: transaction.id,
    orgId: transaction.org_id,
    transactionNo: transaction.transaction_no,
    direction: transaction.direction,
    status: transaction.status,
    businessDate: transaction.business_date,
    counterpartyId: transaction.counterparty_id,
    counterpartyName: transaction.counterparty_name,
    paymentMethod: transaction.payment_method,
    referenceNo: transaction.reference_no,
    allocations: transaction.allocations.map(fromApiCashTransactionAllocation),
    totalAmount: transaction.total_amount,
    currencyCode: transaction.currency_code,
    memo: transaction.memo,
    postedBy: transaction.posted_by,
    postedAt: transaction.posted_at,
    voidReason: transaction.void_reason,
    voidedBy: transaction.voided_by,
    voidedAt: transaction.voided_at,
    auditLogId: transaction.audit_log_id,
    createdAt: transaction.created_at,
    updatedAt: transaction.updated_at,
    version: transaction.version
  };
}

function fromApiCashTransactionAllocation(allocation: CashTransactionAllocationApi): CashTransactionAllocation {
  return {
    id: allocation.id,
    targetType: allocation.target_type,
    targetId: allocation.target_id,
    targetNo: allocation.target_no,
    amount: allocation.amount
  };
}

function filterPrototypeCashTransactions(query: CashTransactionQuery) {
  return prototypeCashTransactions
    .filter((transaction) => matchesCashTransactionQuery(transaction, query))
    .map(cloneCashTransaction);
}

function getPrototypeCashTransaction(id: string) {
  const transaction = prototypeCashTransactions.find((candidate) => candidate.id === id);
  if (!transaction) {
    throw new Error("Cash transaction not found");
  }

  return cloneCashTransaction(transaction);
}

function createPrototypeCashTransaction(input: CreateCashTransactionInput) {
  const transaction = createCashTransactionSeed({
    ...input,
    id: input.id ?? `cash-local-${Date.now()}`,
    transactionNo: input.transactionNo ?? nextPrototypeCashTransactionNo(input.direction),
    status: "posted",
    postedBy: "finance-user",
    postedAt: prototypeNow,
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 2
  });
  prototypeCashTransactions = [transaction, ...prototypeCashTransactions.filter((candidate) => candidate.id !== transaction.id)];

  return cloneCashTransaction(transaction);
}

function matchesCashTransactionQuery(transaction: CashTransaction, query: CashTransactionQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || transaction.status === query.status) &&
    (!query.direction || transaction.direction === query.direction) &&
    (!query.counterpartyId || transaction.counterpartyId === query.counterpartyId) &&
    (!search ||
      [
        transaction.transactionNo,
        transaction.counterpartyId,
        transaction.counterpartyName,
        transaction.referenceNo,
        ...transaction.allocations.flatMap((allocation) => [allocation.targetType, allocation.targetId, allocation.targetNo])
      ]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function createPrototypeCashTransactions(): CashTransaction[] {
  return [
    createCashTransactionSeed({
      id: "cash-in-260430-0001",
      transactionNo: "CASH-IN-260430-0001",
      direction: "cash_in",
      status: "posted",
      businessDate: "2026-04-30",
      counterpartyId: "carrier-ghn",
      counterpartyName: "GHN COD",
      paymentMethod: "bank_transfer",
      referenceNo: "BANK-COD-260430-0001",
      totalAmount: "1250000.00",
      currencyCode: "VND",
      postedBy: "finance-user",
      postedAt: prototypeNow,
      createdAt: prototypeNow,
      updatedAt: prototypeNow,
      version: 1,
      allocations: [
        {
          id: "cash-in-260430-0001-line-1",
          targetType: "customer_receivable",
          targetId: "ar-cod-260430-0001",
          targetNo: "AR-COD-260430-0001",
          amount: "1000000.00"
        },
        {
          id: "cash-in-260430-0001-line-2",
          targetType: "cod_remittance",
          targetId: "cod-remit-260430-0001",
          targetNo: "COD-REMIT-260430-0001",
          amount: "250000.00"
        }
      ]
    }),
    createCashTransactionSeed({
      id: "cash-out-260430-0002",
      transactionNo: "CASH-OUT-260430-0002",
      direction: "cash_out",
      status: "posted",
      businessDate: "2026-04-30",
      counterpartyId: "supplier-hcm-001",
      counterpartyName: "Nguyen Lieu HCM",
      paymentMethod: "bank_transfer",
      referenceNo: "BANK-AP-260430-0002",
      totalAmount: "4250000.00",
      currencyCode: "VND",
      postedBy: "finance-user",
      postedAt: prototypeNow,
      createdAt: prototypeNow,
      updatedAt: prototypeNow,
      version: 1,
      allocations: [
        {
          id: "cash-out-260430-0002-line-1",
          targetType: "supplier_payable",
          targetId: "ap-supplier-260430-0001",
          targetNo: "AP-SUP-260430-0001",
          amount: "4250000.00"
        }
      ]
    })
  ];
}

function createCashTransactionSeed(input: CreateCashTransactionInput & {
  id: string;
  transactionNo: string;
  status: CashTransactionStatus;
  postedBy?: string;
  postedAt?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
}): CashTransaction {
  const totalAmount = normalizeDecimalInput(input.totalAmount, decimalScales.money);
  const allocations = input.allocations.map((allocation) => ({
    ...allocation,
    amount: normalizeDecimalInput(allocation.amount, decimalScales.money)
  }));

  return {
    id: input.id,
    orgId: "org-my-pham",
    transactionNo: input.transactionNo,
    direction: input.direction,
    status: input.status,
    businessDate: input.businessDate,
    counterpartyId: input.counterpartyId,
    counterpartyName: input.counterpartyName,
    paymentMethod: input.paymentMethod,
    referenceNo: input.referenceNo,
    allocations,
    totalAmount,
    currencyCode: input.currencyCode,
    memo: input.memo,
    postedBy: input.postedBy,
    postedAt: input.postedAt,
    createdAt: input.createdAt,
    updatedAt: input.updatedAt,
    version: input.version
  };
}

function nextPrototypeCashTransactionNo(direction: CashTransactionDirection) {
  const prefix = direction === "cash_in" ? "CASH-IN" : "CASH-OUT";
  return `${prefix}-LOCAL-${String(prototypeCashTransactions.length + 1).padStart(4, "0")}`;
}

function cloneCashTransaction(transaction: CashTransaction): CashTransaction {
  return {
    ...transaction,
    allocations: transaction.allocations.map((allocation) => ({ ...allocation }))
  };
}
