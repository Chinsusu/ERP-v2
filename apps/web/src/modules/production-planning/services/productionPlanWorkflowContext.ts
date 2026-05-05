import { formatProductionPlanQuantity } from "./productionPlanService";
import type { ProductionPlan } from "../types";

export const productionPlanWorkflowSteps = [
  {
    number: 1,
    label: "Chọn kế hoạch sản xuất",
    description: "Chọn kế hoạch đang xử lý trước khi tính vật tư và tạo chứng từ."
  },
  {
    number: 2,
    label: "Tính nhu cầu vật tư",
    description: "Hệ thống tính nguyên liệu, bao bì cần dùng theo công thức."
  },
  {
    number: 3,
    label: "Mở đề nghị mua",
    description: "Mở đề nghị mua sinh từ các dòng vật tư còn thiếu."
  },
  {
    number: 4,
    label: "Duyệt đề nghị mua",
    description: "Gửi duyệt và duyệt đề nghị mua trước khi tạo PO."
  },
  {
    number: 5,
    label: "Tạo PO",
    description: "Tạo PO từ đề nghị mua đã duyệt."
  },
  {
    number: 6,
    label: "Nhập kho/QC vật tư",
    description: "Theo dõi vật tư mua về kho và QC nếu cần kiểm soát."
  },
  {
    number: 7,
    label: "Xuất kho vật tư",
    description: "Tạo, duyệt và post phiếu xuất kho vật tư cho kế hoạch."
  },
  {
    number: 8,
    label: "Tạo lệnh gia công",
    description: "Chỉ tạo lệnh sau khi vật tư đã xuất đủ hoặc có waiver."
  }
] as const;

export type ProductionPlanWorkflowContext = {
  planLabel: string;
  outputLabel: string;
  quantityLabel: string;
  formulaLabel: string;
  shortageLineCount: number;
  issueReadyLineCount: number;
  issuedLineCount: number;
  purchaseLineCount: number;
  materialStatusLabel: string;
  materialStatusTone: "success" | "warning" | "info";
  purchaseTitle: string;
  purchaseSummary: string;
  purchaseButtonLabel: string;
  subcontractTitle: string;
  subcontractSummary: string;
  subcontractButtonLabel: string;
};

export function buildProductionPlanWorkflowContext(plan: ProductionPlan): ProductionPlanWorkflowContext {
  const quantityLabel = formatProductionPlanQuantity(plan.plannedQty, plan.uomCode);
  const stockLines = plan.lines.filter((line) => line.isStockManaged);
  const shortageLineCount = stockLines.filter((line) => line.needsPurchase || Number(line.shortageQty) > 0 || line.issueStatus === "shortage").length;
  const issueReadyLineCount = stockLines.filter((line) => line.issueStatus === "ready_to_issue" || line.issueStatus === "partially_issued").length;
  const issuedLineCount = stockLines.filter((line) => line.issueStatus === "issued" || line.issueStatus === "waived").length;
  const purchaseLineCount = plan.purchaseRequestDraft.lines.length;
  const allIssued = stockLines.length === 0 || issuedLineCount === stockLines.length;

  return {
    planLabel: `${plan.planNo} - ${plan.outputSku} - ${quantityLabel}`,
    outputLabel: `${plan.outputSku} - ${plan.outputItemName}`,
    quantityLabel,
    formulaLabel: `${plan.formulaCode} - ${plan.formulaVersion}`,
    shortageLineCount,
    issueReadyLineCount,
    issuedLineCount,
    purchaseLineCount,
    materialStatusLabel: materialStatusLabel(stockLines.length, shortageLineCount, issueReadyLineCount, issuedLineCount),
    materialStatusTone: allIssued ? "success" : issueReadyLineCount > 0 ? "info" : "warning",
    purchaseTitle: `Đề nghị mua từ ${plan.planNo}`,
    purchaseSummary:
      purchaseLineCount > 0
        ? `${purchaseLineCount} dòng vật tư cần mua cho ${plan.outputSku} - ${quantityLabel}; mở đề nghị mua để gửi duyệt và tạo PO.`
        : "Kế hoạch này không có dòng đề nghị mua.",
    purchaseButtonLabel: "Mở đề nghị mua",
    subcontractTitle: `Tạo lệnh gia công từ ${plan.planNo}`,
    subcontractSummary: allIssued
      ? `Vật tư đã đủ bằng chứng xuất kho để tạo lệnh gia công từ ${plan.planNo}.`
      : `Còn vật tư chưa xuất kho cho ${plan.planNo}; cần hoàn tất phiếu xuất trước khi tạo lệnh.`,
    subcontractButtonLabel: "Tạo lệnh gia công từ kế hoạch này"
  };
}

function materialStatusLabel(stockLineCount: number, shortageLineCount: number, issueReadyLineCount: number, issuedLineCount: number) {
  if (stockLineCount === 0) {
    return "Không cần xuất kho";
  }
  if (issuedLineCount === stockLineCount) {
    return "Đã xuất đủ vật tư";
  }
  if (shortageLineCount > 0) {
    return `Thiếu ${shortageLineCount} dòng vật tư`;
  }
  if (issueReadyLineCount > 0) {
    return `Sẵn sàng xuất ${issueReadyLineCount} dòng`;
  }

  return "Chờ phiếu xuất vật tư";
}
