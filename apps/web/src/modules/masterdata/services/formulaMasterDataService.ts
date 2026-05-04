import { apiGetRaw, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import { decimalScales, normalizeDecimalInput, normalizeUOMCode } from "../../../shared/format/numberFormat";
import type {
  FormulaApprovalStatus,
  FormulaComponentType,
  FormulaLineInput,
  FormulaLineItem,
  FormulaLineStatus,
  FormulaMasterDataInput,
  FormulaMasterDataItem,
  FormulaMasterDataQuery,
  FormulaMasterDataSummary,
  FormulaRequirementInput,
  FormulaRequirementLine,
  FormulaRequirementPreview,
  FormulaStatus,
  ProductMasterDataItem,
  ProductType
} from "../types";

type FormulaApiItem = {
  id: string;
  formula_code: string;
  finished_item_id: string;
  finished_sku: string;
  finished_item_name: string;
  finished_item_type: ProductType;
  formula_version: string;
  batch_qty: string;
  batch_uom_code: string;
  base_batch_qty: string;
  base_batch_uom_code: string;
  status: FormulaStatus;
  approval_status: FormulaApprovalStatus;
  effective_from?: string;
  effective_to?: string;
  lines: FormulaApiLine[];
  note?: string;
  created_at: string;
  updated_at: string;
  approved_by?: string;
  approved_at?: string;
  version: number;
  audit_log_id?: string;
};

type FormulaApiLine = {
  id: string;
  line_no: number;
  component_item_id?: string;
  component_sku: string;
  component_name: string;
  component_type: FormulaComponentType;
  entered_qty: string;
  entered_uom_code: string;
  calc_qty: string;
  calc_uom_code: string;
  stock_base_qty: string;
  stock_base_uom_code: string;
  waste_percent: string;
  is_required: boolean;
  is_stock_managed: boolean;
  line_status: FormulaLineStatus;
  note?: string;
};

type FormulaApiRequirement = {
  formula_id: string;
  formula_code: string;
  finished_sku: string;
  planned_qty: string;
  planned_uom_code: string;
  requirements: FormulaApiRequirementLine[];
};

type FormulaApiRequirementLine = {
  formula_line_id: string;
  line_no: number;
  component_item_id?: string;
  component_sku: string;
  component_name: string;
  component_type: FormulaComponentType;
  required_calc_qty: string;
  calc_uom_code: string;
  required_stock_base_qty: string;
  stock_base_uom_code: string;
  is_stock_managed: boolean;
};

const defaultAccessToken = "local-dev-access-token";
const quantityScale = decimalScales.quantity;

export const formulaStatusOptions: { label: string; value: FormulaStatus }[] = [
  { label: "Draft", value: "draft" },
  { label: "Active", value: "active" },
  { label: "Inactive", value: "inactive" },
  { label: "Archived", value: "archived" }
];

export const formulaComponentTypeOptions: { label: string; value: FormulaComponentType }[] = [
  { label: "Nguyên liệu", value: "raw_material" },
  { label: "Hương liệu", value: "fragrance" },
  { label: "Bao bì", value: "packaging" },
  { label: "Bán thành phẩm", value: "semi_finished" },
  { label: "Dịch vụ", value: "service" }
];

export const emptyFormulaInput: FormulaMasterDataInput = {
  formulaCode: "",
  finishedItemId: "",
  finishedSku: "",
  finishedItemName: "",
  finishedItemType: "finished_good",
  formulaVersion: "v1",
  batchQty: "1.000000",
  batchUomCode: "PCS",
  baseBatchQty: "1.000000",
  baseBatchUomCode: "PCS",
  effectiveFrom: "",
  effectiveTo: "",
  lines: [
    {
      lineNo: 1,
      componentItemId: "",
      componentSku: "",
      componentName: "",
      componentType: "raw_material",
      enteredQty: "0.000000",
      enteredUomCode: "KG",
      calcQty: "0.000000",
      calcUomCode: "G",
      stockBaseQty: "0.000000",
      stockBaseUomCode: "KG",
      wastePercent: "0.0000",
      isRequired: true,
      isStockManaged: true,
      lineStatus: "active",
      note: ""
    }
  ],
  note: ""
};

let localFormulas: FormulaMasterDataItem[] = [];

export async function getFormulas(query: FormulaMasterDataQuery = {}): Promise<FormulaMasterDataItem[]> {
  try {
    const formulas = await apiGetRaw<FormulaApiItem[]>(`/formulas${formulaQueryString(query)}`, {
      accessToken: defaultAccessToken
    });

    return formulas.map(fromApiFormula);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return filterFormulas(localFormulas, query).map(cloneFormula);
  }
}

export async function createFormula(input: FormulaMasterDataInput): Promise<FormulaMasterDataItem> {
  const normalized = normalizeFormulaInput(input);
  validateFormulaInput(normalized);

  try {
    const formula = await apiPost<FormulaApiItem, FormulaApiItemRequest>("/formulas", toApiFormulaRequest(normalized), {
      accessToken: defaultAccessToken
    });

    return fromApiFormula(formula);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    ensureUniqueFormulaVersion(normalized);
    const now = new Date().toISOString();
    const formula: FormulaMasterDataItem = {
      ...normalized,
      id: `formula-${normalized.formulaCode.toLowerCase()}-${normalized.formulaVersion.toLowerCase()}-${Date.now()}`,
      status: "draft",
      approvalStatus: "draft",
      lines: normalized.lines.map((line, index) => ({
        ...line,
        id: `formula-line-${normalized.formulaCode.toLowerCase()}-${index + 1}-${Date.now()}`
      })),
      createdAt: now,
      updatedAt: now,
      version: 1,
      auditLogId: `audit-local-formula-create-${Date.now()}`
    };
    localFormulas = sortFormulas([...localFormulas, formula]);

    return cloneFormula(formula);
  }
}

export async function activateFormula(formulaId: string): Promise<FormulaMasterDataItem> {
  try {
    const formula = await apiPost<FormulaApiItem, Record<string, never>>(
      `/formulas/${encodeURIComponent(formulaId)}/activate`,
      {},
      { accessToken: defaultAccessToken }
    );

    return fromApiFormula(formula);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    const current = localFormulas.find((formula) => formula.id === formulaId);
    if (!current) {
      throw new Error("Formula master data was not found");
    }
    const now = new Date().toISOString();
    localFormulas = localFormulas.map((formula) => {
      if (formula.id === formulaId) {
        return { ...formula, status: "active", approvalStatus: "approved", updatedAt: now, version: formula.version + 1 };
      }
      if (formula.finishedItemId === current.finishedItemId && formula.status === "active") {
        return { ...formula, status: "inactive", updatedAt: now, version: formula.version + 1 };
      }
      return formula;
    });

    return cloneFormula(localFormulas.find((formula) => formula.id === formulaId)!);
  }
}

export async function calculateFormulaRequirement(
  formulaId: string,
  input: FormulaRequirementInput
): Promise<FormulaRequirementPreview> {
  const planned = {
    plannedQty: normalizeDecimalInput(input.plannedQty, quantityScale),
    plannedUomCode: normalizeUOMCode(input.plannedUomCode)
  };
  try {
    const preview = await apiPost<FormulaApiRequirement, { planned_qty: string; planned_uom_code: string }>(
      `/formulas/${encodeURIComponent(formulaId)}/calculate-requirement`,
      {
        planned_qty: planned.plannedQty,
        planned_uom_code: planned.plannedUomCode
      },
      { accessToken: defaultAccessToken }
    );

    return fromApiRequirement(preview);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    const formula = localFormulas.find((candidate) => candidate.id === formulaId);
    if (!formula) {
      throw new Error("Formula master data was not found");
    }

    return {
      formulaId: formula.id,
      formulaCode: formula.formulaCode,
      finishedSku: formula.finishedSku,
      plannedQty: planned.plannedQty,
      plannedUomCode: planned.plannedUomCode,
      requirements: formula.lines
        .filter((line) => line.lineStatus === "active")
        .map((line) => ({
          formulaLineId: line.id,
          lineNo: line.lineNo,
          componentItemId: line.componentItemId,
          componentSku: line.componentSku,
          componentName: line.componentName,
          componentType: line.componentType,
          requiredCalcQty: scaleQuantity(line.calcQty, planned.plannedQty, formula.batchQty),
          calcUomCode: line.calcUomCode,
          requiredStockBaseQty: scaleQuantity(line.stockBaseQty, planned.plannedQty, formula.batchQty),
          stockBaseUomCode: line.stockBaseUomCode,
          isStockManaged: line.isStockManaged
        }))
    };
  }
}

export function normalizeFormulaInput(input: FormulaMasterDataInput): FormulaMasterDataInput {
  return {
    ...input,
    formulaCode: normalizeCode(input.formulaCode),
    finishedItemId: input.finishedItemId.trim(),
    finishedSku: normalizeCode(input.finishedSku),
    finishedItemName: input.finishedItemName.trim(),
    formulaVersion: input.formulaVersion.trim(),
    batchQty: normalizeDecimalInput(input.batchQty, quantityScale),
    batchUomCode: normalizeUOMCode(input.batchUomCode),
    baseBatchQty: normalizeDecimalInput(input.baseBatchQty, quantityScale),
    baseBatchUomCode: normalizeUOMCode(input.baseBatchUomCode),
    effectiveFrom: input.effectiveFrom.trim(),
    effectiveTo: input.effectiveTo.trim(),
    lines: input.lines.map(normalizeFormulaLineInput),
    note: input.note.trim()
  };
}

export function formulaInputForParentItem(input: FormulaMasterDataInput, parent: Pick<ProductMasterDataItem, "id" | "skuCode" | "name" | "itemType" | "status">): FormulaMasterDataInput {
  if (parent.status !== "active") {
    throw new Error("Formula parent item must be active");
  }
  if (parent.itemType !== "finished_good" && parent.itemType !== "semi_finished") {
    throw new Error("Formula parent item must be a finished good or semi finished item");
  }

  return {
    ...input,
    finishedItemId: parent.id,
    finishedSku: parent.skuCode,
    finishedItemName: parent.name,
    finishedItemType: parent.itemType
  };
}

export function summarizeFormulas(items: FormulaMasterDataItem[]): FormulaMasterDataSummary {
  return {
    total: items.length,
    active: items.filter((item) => item.status === "active").length,
    draft: items.filter((item) => item.status === "draft").length,
    lines: items.reduce((total, item) => total + item.lines.length, 0)
  };
}

export function formatFormulaQuantity(value: string, uomCode: string) {
  const normalizedUom = normalizeUOMCode(uomCode);
  const normalized = normalizeDecimalInput(value, quantityScale);
  if (normalizedUom === "KG") {
    const kilograms = Number(normalized);
    if (Math.abs(kilograms) > 0 && Math.abs(kilograms) < 0.001) {
      return `${formatLocalNumber(kilograms * 1_000_000)} mg`;
    }
    if (Math.abs(kilograms) < 1) {
      return `${formatLocalNumber(kilograms * 1_000)} g`;
    }
    return `${formatLocalNumber(kilograms)} kg`;
  }
  if (normalizedUom === "G") {
    const grams = Number(normalized);
    if (Math.abs(grams) > 0 && Math.abs(grams) < 1) {
      return `${formatLocalNumber(grams * 1_000)} mg`;
    }
    return `${formatLocalNumber(grams)} g`;
  }
  if (normalizedUom === "MG") {
    return `${formatLocalNumber(Number(normalized))} mg`;
  }

  return `${formatLocalNumber(Number(normalized))} ${normalizedUom}`;
}

export function resetPrototypeFormulaMasterData() {
  localFormulas = [];
}

type FormulaApiItemRequest = {
  formula_code: string;
  finished_item_id: string;
  finished_sku: string;
  finished_item_name: string;
  finished_item_type: ProductType;
  formula_version: string;
  batch_qty: string;
  batch_uom_code: string;
  base_batch_qty: string;
  base_batch_uom_code: string;
  effective_from: string;
  effective_to: string;
  lines: FormulaApiLineRequest[];
  note: string;
};

type FormulaApiLineRequest = {
  line_no: number;
  component_item_id: string;
  component_sku: string;
  component_name: string;
  component_type: FormulaComponentType;
  entered_qty: string;
  entered_uom_code: string;
  calc_qty: string;
  calc_uom_code: string;
  stock_base_qty: string;
  stock_base_uom_code: string;
  waste_percent: string;
  is_required: boolean;
  is_stock_managed: boolean;
  line_status: FormulaLineStatus;
  note: string;
};

function normalizeFormulaLineInput(line: FormulaLineInput, index: number): FormulaLineInput {
  const componentSku = normalizeCode(line.componentSku);

  return {
    ...line,
    lineNo: line.lineNo > 0 ? line.lineNo : index + 1,
    componentItemId: line.componentItemId.trim() || componentSku,
    componentSku,
    componentName: line.componentName.trim(),
    enteredQty: normalizeDecimalInput(line.enteredQty, quantityScale),
    enteredUomCode: normalizeUOMCode(line.enteredUomCode),
    calcQty: normalizeDecimalInput(line.calcQty, quantityScale),
    calcUomCode: normalizeUOMCode(line.calcUomCode),
    stockBaseQty: normalizeDecimalInput(line.stockBaseQty, quantityScale),
    stockBaseUomCode: normalizeUOMCode(line.stockBaseUomCode),
    wastePercent: normalizeDecimalInput(line.wastePercent, decimalScales.rate),
    lineStatus: line.lineStatus || "active",
    note: line.note.trim()
  };
}

function validateFormulaInput(input: FormulaMasterDataInput) {
  if (
    !input.formulaCode ||
    !input.finishedItemId ||
    !input.finishedSku ||
    !input.finishedItemName ||
    !input.formulaVersion ||
    input.lines.length === 0
  ) {
    throw new Error("Formula required fields are missing");
  }
  if (!["finished_good", "semi_finished"].includes(input.finishedItemType)) {
    throw new Error("Formula parent item must be a finished good or semi finished item");
  }
  if (input.batchQty === "0.000000" || input.baseBatchQty === "0.000000") {
    throw new Error("Formula batch quantity must be greater than zero");
  }
  input.lines.forEach((line) => {
    if (!line.componentSku || !line.componentName) {
      throw new Error("Formula line required fields are missing");
    }
    if (line.isRequired && line.enteredQty === "0.000000" && line.calcQty === "0.000000" && line.stockBaseQty === "0.000000") {
      throw new Error("Required formula lines need a positive quantity");
    }
  });
}

function toApiFormulaRequest(input: FormulaMasterDataInput): FormulaApiItemRequest {
  return {
    formula_code: input.formulaCode,
    finished_item_id: input.finishedItemId,
    finished_sku: input.finishedSku,
    finished_item_name: input.finishedItemName,
    finished_item_type: input.finishedItemType,
    formula_version: input.formulaVersion,
    batch_qty: input.batchQty,
    batch_uom_code: input.batchUomCode,
    base_batch_qty: input.baseBatchQty,
    base_batch_uom_code: input.baseBatchUomCode,
    effective_from: input.effectiveFrom,
    effective_to: input.effectiveTo,
    lines: input.lines.map((line) => ({
      line_no: line.lineNo,
      component_item_id: line.componentItemId,
      component_sku: line.componentSku,
      component_name: line.componentName,
      component_type: line.componentType,
      entered_qty: line.enteredQty,
      entered_uom_code: line.enteredUomCode,
      calc_qty: line.calcQty,
      calc_uom_code: line.calcUomCode,
      stock_base_qty: line.stockBaseQty,
      stock_base_uom_code: line.stockBaseUomCode,
      waste_percent: line.wastePercent,
      is_required: line.isRequired,
      is_stock_managed: line.isStockManaged,
      line_status: line.lineStatus,
      note: line.note
    })),
    note: input.note
  };
}

function fromApiFormula(item: FormulaApiItem): FormulaMasterDataItem {
  return {
    id: item.id,
    formulaCode: item.formula_code,
    finishedItemId: item.finished_item_id,
    finishedSku: item.finished_sku,
    finishedItemName: item.finished_item_name,
    finishedItemType: item.finished_item_type,
    formulaVersion: item.formula_version,
    batchQty: item.batch_qty,
    batchUomCode: item.batch_uom_code,
    baseBatchQty: item.base_batch_qty,
    baseBatchUomCode: item.base_batch_uom_code,
    status: item.status,
    approvalStatus: item.approval_status,
    effectiveFrom: item.effective_from,
    effectiveTo: item.effective_to,
    lines: item.lines.map(fromApiFormulaLine),
    note: item.note,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    approvedBy: item.approved_by,
    approvedAt: item.approved_at,
    version: item.version,
    auditLogId: item.audit_log_id
  };
}

function fromApiFormulaLine(line: FormulaApiLine): FormulaLineItem {
  return {
    id: line.id,
    lineNo: line.line_no,
    componentItemId: line.component_item_id,
    componentSku: line.component_sku,
    componentName: line.component_name,
    componentType: line.component_type,
    enteredQty: line.entered_qty,
    enteredUomCode: line.entered_uom_code,
    calcQty: line.calc_qty,
    calcUomCode: line.calc_uom_code,
    stockBaseQty: line.stock_base_qty,
    stockBaseUomCode: line.stock_base_uom_code,
    wastePercent: line.waste_percent,
    isRequired: line.is_required,
    isStockManaged: line.is_stock_managed,
    lineStatus: line.line_status,
    note: line.note
  };
}

function fromApiRequirement(preview: FormulaApiRequirement): FormulaRequirementPreview {
  return {
    formulaId: preview.formula_id,
    formulaCode: preview.formula_code,
    finishedSku: preview.finished_sku,
    plannedQty: preview.planned_qty,
    plannedUomCode: preview.planned_uom_code,
    requirements: preview.requirements.map(fromApiRequirementLine)
  };
}

function fromApiRequirementLine(line: FormulaApiRequirementLine): FormulaRequirementLine {
  return {
    formulaLineId: line.formula_line_id,
    lineNo: line.line_no,
    componentItemId: line.component_item_id,
    componentSku: line.component_sku,
    componentName: line.component_name,
    componentType: line.component_type,
    requiredCalcQty: line.required_calc_qty,
    calcUomCode: line.calc_uom_code,
    requiredStockBaseQty: line.required_stock_base_qty,
    stockBaseUomCode: line.stock_base_uom_code,
    isStockManaged: line.is_stock_managed
  };
}

function filterFormulas(items: FormulaMasterDataItem[], query: FormulaMasterDataQuery) {
  const search = query.search?.trim().toLowerCase() ?? "";
  return sortFormulas(
    items.filter((item) => {
      if (query.status && item.status !== query.status) {
        return false;
      }
      if (query.finishedItemId && item.finishedItemId !== query.finishedItemId) {
        return false;
      }
      if (!search) {
        return true;
      }

      return [item.formulaCode, item.finishedSku, item.finishedItemName, item.formulaVersion]
        .join(" ")
        .toLowerCase()
        .includes(search);
    })
  );
}

function ensureUniqueFormulaVersion(input: FormulaMasterDataInput) {
  if (
    localFormulas.some(
      (formula) => formula.finishedItemId === input.finishedItemId && formula.formulaVersion.toLowerCase() === input.formulaVersion.toLowerCase()
    )
  ) {
    throw new Error("Formula version already exists");
  }
}

function cloneFormula(formula: FormulaMasterDataItem): FormulaMasterDataItem {
  return {
    ...formula,
    lines: formula.lines.map((line) => ({ ...line }))
  };
}

function sortFormulas(items: FormulaMasterDataItem[]) {
  return [...items].sort((left, right) => {
    const status = formulaStatusRank(left.status) - formulaStatusRank(right.status);
    if (status !== 0) {
      return status;
    }
    return `${left.finishedSku}-${left.formulaVersion}`.localeCompare(`${right.finishedSku}-${right.formulaVersion}`);
  });
}

function formulaStatusRank(status: FormulaStatus) {
  return { active: 0, draft: 1, inactive: 2, archived: 3 }[status] ?? 4;
}

function formulaQueryString(query: FormulaMasterDataQuery) {
  const params = new URLSearchParams();
  if (query.search) {
    params.set("q", query.search);
  }
  if (query.status) {
    params.set("status", query.status);
  }
  if (query.finishedItemId) {
    params.set("finished_item_id", query.finishedItemId);
  }
  const value = params.toString();
  return value ? `?${value}` : "";
}

function normalizeCode(value: string) {
  return value.trim().toUpperCase();
}

function formatLocalNumber(value: number) {
  return new Intl.NumberFormat("vi-VN", {
    maximumFractionDigits: 6,
    minimumFractionDigits: 0
  }).format(value);
}

function scaleQuantity(value: string, plannedQty: string, batchQty: string) {
  const valueScaled = scaledQuantity(value);
  const plannedScaled = scaledQuantity(plannedQty);
  const batchScaled = scaledQuantity(batchQty);
  if (batchScaled <= BigInt(0)) {
    throw new Error("Formula batch quantity must be greater than zero");
  }

  const numerator = valueScaled * plannedScaled;
  const quotient = (numerator + batchScaled / BigInt(2)) / batchScaled;
  return scaledToDecimal(quotient);
}

function scaledQuantity(value: string) {
  const normalized = normalizeDecimalInput(value, quantityScale);
  const negative = normalized.startsWith("-");
  const digits = normalized.replace("-", "").replace(".", "");
  const scaled = BigInt(digits);
  return negative ? -scaled : scaled;
}

function scaledToDecimal(value: bigint) {
  const negative = value < BigInt(0);
  const unsigned = String(negative ? -value : value).padStart(quantityScale + 1, "0");
  const integer = unsigned.slice(0, -quantityScale);
  const fraction = unsigned.slice(-quantityScale);
  return `${negative ? "-" : ""}${integer}.${fraction}`;
}
