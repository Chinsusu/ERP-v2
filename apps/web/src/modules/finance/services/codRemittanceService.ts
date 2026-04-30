import { ApiError, apiGet, apiGetRaw, apiPost } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import { decimalScales, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import type {
  CODDiscrepancy,
  CODDiscrepancyStatus,
  CODDiscrepancyType,
  CODLineMatchStatus,
  CODRemittance,
  CODRemittanceActionResult,
  CODRemittanceDiscrepancyInput,
  CODRemittanceLine,
  CODRemittanceQuery,
  CODRemittanceStatus
} from "../types";

type CODRemittanceApi = components["schemas"]["CODRemittance"];
type CODRemittanceListItemApi = components["schemas"]["CODRemittanceListItem"];
type CODRemittanceLineApi = components["schemas"]["CODRemittanceLine"];
type CODDiscrepancyApi = components["schemas"]["CODDiscrepancy"];
type CODRemittanceActionApiResult = components["schemas"]["CODRemittanceActionResult"];
type CODRemittanceDiscrepancyApiRequest = components["schemas"]["CODRemittanceDiscrepancyRequest"];
type CODRemittanceListApiQuery = operations["listCODRemittances"]["parameters"]["query"];

const defaultAccessToken = "local-dev-access-token";
const prototypeNow = "2026-04-30T09:00:00Z";

let prototypeCODRemittances = createPrototypeCODRemittances();

export const codRemittanceStatusOptions: { label: string; value: "" | CODRemittanceStatus }[] = [
  { label: "All statuses", value: "" },
  { label: "Draft", value: "draft" },
  { label: "Matching", value: "matching" },
  { label: "Discrepancy", value: "discrepancy" },
  { label: "Submitted", value: "submitted" },
  { label: "Approved", value: "approved" },
  { label: "Closed", value: "closed" },
  { label: "Void", value: "void" }
];

export async function getCODRemittances(query: CODRemittanceQuery = {}): Promise<CODRemittance[]> {
  try {
    const remittances = await apiGet("/cod-remittances", {
      accessToken: defaultAccessToken,
      query: toApiCODRemittanceQuery(query)
    });

    return remittances.map(fromApiCODRemittanceListItem);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return filterPrototypeCODRemittances(query);
  }
}

export async function getCODRemittance(id: string): Promise<CODRemittance> {
  try {
    const remittance = await apiGetRaw<CODRemittanceApi>(`/cod-remittances/${encodeURIComponent(id)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiCODRemittance(remittance);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return getPrototypeCODRemittance(id);
  }
}

export async function matchCODRemittance(id: string): Promise<CODRemittanceActionResult> {
  return runCODRemittanceAction(id, "match");
}

export async function submitCODRemittance(id: string): Promise<CODRemittanceActionResult> {
  return runCODRemittanceAction(id, "submit");
}

export async function approveCODRemittance(id: string): Promise<CODRemittanceActionResult> {
  return runCODRemittanceAction(id, "approve");
}

export async function closeCODRemittance(id: string): Promise<CODRemittanceActionResult> {
  return runCODRemittanceAction(id, "close");
}

export async function recordCODRemittanceDiscrepancy(
  id: string,
  input: CODRemittanceDiscrepancyInput
): Promise<CODRemittanceActionResult> {
  try {
    const result = await apiPost<CODRemittanceActionApiResult, CODRemittanceDiscrepancyApiRequest>(
      `/cod-remittances/${encodeURIComponent(id)}/record-discrepancy`,
      toApiDiscrepancyInput(input),
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypeCODRemittance(id, "record-discrepancy", input);
  }
}

export function resetPrototypeCODRemittancesForTest() {
  prototypeCODRemittances = createPrototypeCODRemittances();
}

export function codRemittanceStatusTone(
  status: CODRemittanceStatus
): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "closed":
      return "success";
    case "submitted":
    case "approved":
      return "info";
    case "discrepancy":
      return "warning";
    case "void":
      return "danger";
    case "draft":
    case "matching":
    default:
      return "normal";
  }
}

export function codLineMatchStatusTone(
  status: CODLineMatchStatus
): "success" | "warning" | "danger" | "info" | "normal" {
  switch (status) {
    case "matched":
      return "success";
    case "over_paid":
      return "info";
    case "short_paid":
    default:
      return "warning";
  }
}

export function formatCODStatus(status: CODRemittanceStatus | CODLineMatchStatus | CODDiscrepancyType | CODDiscrepancyStatus) {
  return status
    .split("_")
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}

export function canMatchCODRemittance(remittance: CODRemittance | null) {
  return Boolean(remittance && remittance.status === "draft" && Number(remittance.discrepancyAmount) === 0);
}

export function canRecordCODDiscrepancy(remittance: CODRemittance | null) {
  return Boolean(remittance && ["draft", "matching", "discrepancy"].includes(remittance.status));
}

export function canSubmitCODRemittance(remittance: CODRemittance | null) {
  return Boolean(
    remittance &&
      ["matching", "discrepancy"].includes(remittance.status) &&
      (!hasCODDiscrepancy(remittance) || hasTraceForEveryDiscrepantLine(remittance))
  );
}

export function canApproveCODRemittance(remittance: CODRemittance | null) {
  return Boolean(remittance && remittance.status === "submitted");
}

export function canCloseCODRemittance(remittance: CODRemittance | null) {
  return Boolean(remittance && remittance.status === "approved");
}

async function runCODRemittanceAction(
  id: string,
  action: "match" | "submit" | "approve" | "close"
): Promise<CODRemittanceActionResult> {
  try {
    const result = await apiPost<CODRemittanceActionApiResult, Record<string, never>>(
      `/cod-remittances/${encodeURIComponent(id)}/${action}`,
      {},
      { accessToken: defaultAccessToken }
    );

    return fromApiActionResult(result);
  } catch (cause) {
    if (!shouldUsePrototypeFallback(cause)) {
      throw cause;
    }

    return transitionPrototypeCODRemittance(id, action);
  }
}

function shouldUsePrototypeFallback(reason: unknown) {
  if (reason instanceof ApiError) {
    return false;
  }

  return !(reason instanceof Error && reason.message.startsWith("API request failed:"));
}

function toApiCODRemittanceQuery(query: CODRemittanceQuery): CODRemittanceListApiQuery {
  return {
    q: query.search,
    status: query.status,
    carrier_id: query.carrierId
  };
}

function fromApiCODRemittanceListItem(item: CODRemittanceListItemApi): CODRemittance {
  return {
    id: item.id,
    remittanceNo: item.remittance_no,
    carrierId: item.carrier_id,
    carrierCode: item.carrier_code,
    carrierName: item.carrier_name,
    status: item.status,
    businessDate: item.business_date,
    expectedAmount: item.expected_amount,
    remittedAmount: item.remitted_amount,
    discrepancyAmount: item.discrepancy_amount,
    currencyCode: item.currency_code,
    lines: [],
    discrepancies: [],
    lineCount: item.line_count,
    discrepancyCount: item.discrepancy_count,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    version: item.version
  };
}

function fromApiCODRemittance(remittance: CODRemittanceApi): CODRemittance {
  const lines = remittance.lines.map(fromApiCODRemittanceLine);
  const discrepancies = remittance.discrepancies.map(fromApiCODDiscrepancy);

  return {
    id: remittance.id,
    orgId: remittance.org_id,
    remittanceNo: remittance.remittance_no,
    carrierId: remittance.carrier_id,
    carrierCode: remittance.carrier_code,
    carrierName: remittance.carrier_name,
    status: remittance.status,
    businessDate: remittance.business_date,
    expectedAmount: remittance.expected_amount,
    remittedAmount: remittance.remitted_amount,
    discrepancyAmount: remittance.discrepancy_amount,
    currencyCode: remittance.currency_code,
    lines,
    discrepancies,
    lineCount: lines.length,
    discrepancyCount: discrepancies.length,
    auditLogId: remittance.audit_log_id,
    submittedBy: remittance.submitted_by,
    submittedAt: remittance.submitted_at,
    approvedBy: remittance.approved_by,
    approvedAt: remittance.approved_at,
    closedBy: remittance.closed_by,
    closedAt: remittance.closed_at,
    createdAt: remittance.created_at,
    updatedAt: remittance.updated_at,
    version: remittance.version
  };
}

function fromApiCODRemittanceLine(line: CODRemittanceLineApi): CODRemittanceLine {
  return {
    id: line.id,
    receivableId: line.receivable_id,
    receivableNo: line.receivable_no,
    shipmentId: line.shipment_id,
    trackingNo: line.tracking_no,
    customerName: line.customer_name,
    expectedAmount: line.expected_amount,
    remittedAmount: line.remitted_amount,
    discrepancyAmount: line.discrepancy_amount,
    matchStatus: line.match_status
  };
}

function fromApiCODDiscrepancy(discrepancy: CODDiscrepancyApi): CODDiscrepancy {
  return {
    id: discrepancy.id,
    lineId: discrepancy.line_id,
    receivableId: discrepancy.receivable_id,
    type: discrepancy.type,
    status: discrepancy.status,
    amount: discrepancy.amount,
    reason: discrepancy.reason,
    ownerId: discrepancy.owner_id,
    recordedBy: discrepancy.recorded_by,
    recordedAt: discrepancy.recorded_at,
    resolvedBy: discrepancy.resolved_by,
    resolvedAt: discrepancy.resolved_at,
    resolution: discrepancy.resolution
  };
}

function fromApiActionResult(result: CODRemittanceActionApiResult): CODRemittanceActionResult {
  return {
    codRemittance: fromApiCODRemittance(result.cod_remittance),
    previousStatus: result.previous_status,
    currentStatus: result.current_status,
    auditLogId: result.audit_log_id
  };
}

function toApiDiscrepancyInput(input: CODRemittanceDiscrepancyInput): CODRemittanceDiscrepancyApiRequest {
  return {
    id: input.id,
    line_id: input.lineId,
    type: input.type,
    status: input.status,
    reason: input.reason,
    owner_id: input.ownerId
  };
}

function filterPrototypeCODRemittances(query: CODRemittanceQuery) {
  return prototypeCODRemittances
    .filter((remittance) => matchesCODRemittanceQuery(remittance, query))
    .map(cloneCODRemittance);
}

function getPrototypeCODRemittance(id: string) {
  const remittance = prototypeCODRemittances.find((candidate) => candidate.id === id);
  if (!remittance) {
    throw new Error("COD remittance not found");
  }

  return cloneCODRemittance(remittance);
}

function transitionPrototypeCODRemittance(
  id: string,
  action: "match" | "record-discrepancy" | "submit" | "approve" | "close",
  discrepancyInput?: CODRemittanceDiscrepancyInput
): CODRemittanceActionResult {
  const current = getPrototypeCODRemittance(id);
  const previousStatus = current.status;
  let next = cloneCODRemittance(current);

  if (action === "match") {
    if (!canMatchCODRemittance(next)) {
      throw new Error("Only clean draft COD remittances can be matched");
    }
    next = { ...next, status: "matching" };
  } else if (action === "record-discrepancy") {
    if (!discrepancyInput || !canRecordCODDiscrepancy(next)) {
      throw new Error("COD discrepancy cannot be recorded for this status");
    }
    next = recordPrototypeDiscrepancy(next, discrepancyInput);
  } else if (action === "submit") {
    if (!canSubmitCODRemittance(next)) {
      throw new Error("COD remittance must be matched or traced before submit");
    }
    next = { ...next, status: "submitted", submittedBy: "finance-user", submittedAt: prototypeNow };
  } else if (action === "approve") {
    if (!canApproveCODRemittance(next)) {
      throw new Error("Only submitted COD remittances can be approved");
    }
    next = { ...next, status: "approved", approvedBy: "finance-manager", approvedAt: prototypeNow };
  } else {
    if (!canCloseCODRemittance(next)) {
      throw new Error("Only approved COD remittances can be closed");
    }
    next = { ...next, status: "closed", closedBy: "finance-manager", closedAt: prototypeNow };
  }

  next = {
    ...next,
    updatedAt: prototypeNow,
    version: next.version + 1,
    auditLogId: `audit-cod-${action}-${next.id}`,
    lineCount: next.lines.length,
    discrepancyCount: next.discrepancies.length
  };
  prototypeCODRemittances = [next, ...prototypeCODRemittances.filter((remittance) => remittance.id !== next.id)];

  return {
    codRemittance: cloneCODRemittance(next),
    previousStatus,
    currentStatus: next.status,
    auditLogId: next.auditLogId
  };
}

function recordPrototypeDiscrepancy(
  remittance: CODRemittance,
  input: CODRemittanceDiscrepancyInput
): CODRemittance {
  const line = remittance.lines.find((candidate) => candidate.id === input.lineId);
  if (!line || Number(line.discrepancyAmount) === 0) {
    throw new Error("COD discrepancy line is invalid");
  }
  const discrepancy: CODDiscrepancy = {
    id: input.id ?? `${line.id}-discrepancy`,
    lineId: line.id,
    receivableId: line.receivableId,
    type: input.type ?? defaultDiscrepancyType(line.discrepancyAmount),
    status: input.status ?? "open",
    amount: line.discrepancyAmount,
    reason: input.reason.trim() || "Carrier remittance mismatch",
    ownerId: input.ownerId.trim() || "finance-user",
    recordedBy: "finance-user",
    recordedAt: prototypeNow
  };
  const discrepancies = [
    discrepancy,
    ...remittance.discrepancies.filter((candidate) => candidate.id !== discrepancy.id)
  ];

  return {
    ...remittance,
    status: "discrepancy",
    discrepancies,
    discrepancyCount: discrepancies.length
  };
}

function matchesCODRemittanceQuery(remittance: CODRemittance, query: CODRemittanceQuery) {
  const search = query.search?.trim().toLowerCase();
  return (
    (!query.status || remittance.status === query.status) &&
    (!query.carrierId || remittance.carrierId === query.carrierId) &&
    (!search ||
      [
        remittance.remittanceNo,
        remittance.carrierCode,
        remittance.carrierName,
        ...remittance.lines.flatMap((line) => [line.receivableNo, line.trackingNo, line.customerName])
      ]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)))
  );
}

function hasCODDiscrepancy(remittance: CODRemittance) {
  return remittance.lines.some((line) => Number(line.discrepancyAmount) !== 0);
}

function hasTraceForEveryDiscrepantLine(remittance: CODRemittance) {
  const tracedLineIds = new Set(remittance.discrepancies.map((discrepancy) => discrepancy.lineId));

  return remittance.lines.every((line) => Number(line.discrepancyAmount) === 0 || tracedLineIds.has(line.id));
}

function defaultDiscrepancyType(amount: string): CODDiscrepancyType {
  return Number(amount) < 0 ? "short_paid" : "over_paid";
}

function createPrototypeCODRemittances(): CODRemittance[] {
  return [
    createCODRemittanceSeed({
      id: "cod-remit-260430-0001",
      remittanceNo: "COD-GHN-260430-0001",
      carrierId: "carrier-ghn",
      carrierCode: "GHN",
      carrierName: "GHN Express",
      status: "draft",
      businessDate: "2026-04-30",
      lines: [
        {
          id: "cod-remit-260430-0001-line-1",
          receivableId: "ar-cod-260430-0001",
          receivableNo: "AR-COD-260430-0001",
          shipmentId: "shipment-cod-260430-0001",
          trackingNo: "GHN260430001",
          customerName: "My Pham HCM Retail",
          expectedAmount: "1250000.00",
          remittedAmount: "1200000.00"
        },
        {
          id: "cod-remit-260430-0001-line-2",
          receivableId: "ar-cod-260430-0002",
          receivableNo: "AR-COD-260430-0002",
          shipmentId: "shipment-cod-260430-0002",
          trackingNo: "GHN260430002",
          customerName: "Marketplace COD",
          expectedAmount: "750000.00",
          remittedAmount: "750000.00"
        }
      ]
    }),
    createCODRemittanceSeed({
      id: "cod-remit-260430-0002",
      remittanceNo: "COD-SPX-260430-0002",
      carrierId: "carrier-spx",
      carrierCode: "SPX",
      carrierName: "Shopee Express",
      status: "matching",
      businessDate: "2026-04-30",
      lines: [
        {
          id: "cod-remit-260430-0002-line-1",
          receivableId: "ar-cod-260430-0003",
          receivableNo: "AR-COD-260430-0003",
          shipmentId: "shipment-cod-260430-0003",
          trackingNo: "SPX260430122",
          customerName: "Marketplace COD",
          expectedAmount: "2480000.00",
          remittedAmount: "2480000.00"
        }
      ]
    })
  ];
}

function createCODRemittanceSeed(input: {
  id: string;
  remittanceNo: string;
  carrierId: string;
  carrierCode: string;
  carrierName: string;
  status: CODRemittanceStatus;
  businessDate: string;
  lines: Omit<CODRemittanceLine, "discrepancyAmount" | "matchStatus">[];
}): CODRemittance {
  const lines = input.lines.map((line) => {
    const expectedAmount = normalizeDecimalInput(line.expectedAmount, decimalScales.money);
    const remittedAmount = normalizeDecimalInput(line.remittedAmount, decimalScales.money);
    const discrepancyAmount = normalizeDecimalInput(Number(remittedAmount) - Number(expectedAmount), decimalScales.money);

    return {
      ...line,
      expectedAmount,
      remittedAmount,
      discrepancyAmount,
      matchStatus: matchStatusForAmount(discrepancyAmount)
    };
  });
  const expectedAmount = sumMoney(lines.map((line) => line.expectedAmount));
  const remittedAmount = sumMoney(lines.map((line) => line.remittedAmount));
  const discrepancyAmount = normalizeDecimalInput(Number(remittedAmount) - Number(expectedAmount), decimalScales.money);

  return {
    id: input.id,
    orgId: "org-my-pham",
    remittanceNo: input.remittanceNo,
    carrierId: input.carrierId,
    carrierCode: input.carrierCode,
    carrierName: input.carrierName,
    status: input.status,
    businessDate: input.businessDate,
    expectedAmount,
    remittedAmount,
    discrepancyAmount,
    currencyCode: "VND",
    lines,
    discrepancies: [],
    lineCount: lines.length,
    discrepancyCount: 0,
    createdAt: prototypeNow,
    updatedAt: prototypeNow,
    version: 1
  };
}

function matchStatusForAmount(amount: string): CODLineMatchStatus {
  const value = Number(amount);
  if (value < 0) {
    return "short_paid";
  }
  if (value > 0) {
    return "over_paid";
  }

  return "matched";
}

function sumMoney(values: string[]) {
  return normalizeDecimalInput(
    values.reduce((total, value) => total + Number(value), 0),
    decimalScales.money
  );
}

function cloneCODRemittance(remittance: CODRemittance): CODRemittance {
  return {
    ...remittance,
    lines: remittance.lines.map((line) => ({ ...line })),
    discrepancies: remittance.discrepancies.map((discrepancy) => ({ ...discrepancy }))
  };
}
