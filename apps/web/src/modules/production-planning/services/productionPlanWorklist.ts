import { formatProductionPlanQuantity } from "./productionPlanService";
import type { ProductionPlan } from "../types";

type WorkTaskTone = "normal" | "success" | "warning" | "danger" | "info";

export type ProductionPlanWorkTask = {
  id: string;
  step: number;
  title: string;
  statusLabel: string;
  statusTone: WorkTaskTone;
  detail: string;
  action?: {
    label: string;
    href?: string;
    disabled?: boolean;
  };
};

export function buildProductionPlanWorklist(plan: ProductionPlan): ProductionPlanWorkTask[] {
  const lineCount = plan.lines.length;
  const shortageLineCount = plan.lines.filter((line) => line.needsPurchase || Number(line.shortageQty) > 0).length;
  const purchaseLineCount = plan.purchaseRequestDraft.lines.length;
  const quantityLabel = formatProductionPlanQuantity(plan.plannedQty, plan.uomCode);

  return [
    {
      id: "production-plan",
      step: 1,
      title: "Kế hoạch sản xuất",
      statusLabel: "Đã tạo",
      statusTone: "success",
      detail: `${plan.planNo} - ${plan.outputSku} - ${quantityLabel}; công thức ${plan.formulaCode} - ${plan.formulaVersion}.`
    },
    {
      id: "material-demand",
      step: 2,
      title: "Nhu cầu vật tư",
      statusLabel: shortageLineCount > 0 ? `Thiếu ${shortageLineCount} dòng vật tư` : "Đủ vật tư",
      statusTone: shortageLineCount > 0 ? "warning" : "success",
      detail:
        shortageLineCount > 0
          ? `${shortageLineCount}/${lineCount} dòng vật tư cần mua thêm.`
          : `${lineCount} dòng vật tư đã đủ tồn khả dụng.`
    },
    {
      id: "purchase-order",
      step: 3,
      title: "PO vật tư thiếu",
      statusLabel: purchaseLineCount > 0 ? "Cần theo dõi PO" : "Không cần PO",
      statusTone: purchaseLineCount > 0 ? "warning" : "success",
      detail:
        purchaseLineCount > 0
          ? `${purchaseLineCount} dòng đề nghị mua nháp từ ${plan.planNo}; mở Mua hàng để tạo hoặc kiểm tra PO đã tạo.`
          : "Kế hoạch không phát sinh đề nghị mua từ MRP.",
      action:
        purchaseLineCount > 0
          ? {
              label: "Mở mua hàng",
              href: "/purchase",
              disabled: false
            }
          : undefined
    },
    {
      id: "receiving",
      step: 4,
      title: "Nhập kho vật tư",
      statusLabel: purchaseLineCount > 0 ? "Chờ PO/Nhập kho" : "Không chờ nhập mua",
      statusTone: purchaseLineCount > 0 ? "info" : "success",
      detail:
        purchaseLineCount > 0
          ? "Theo dõi PO, lịch giao và phiếu nhập kho ở module Mua hàng/Nhập kho."
          : "Không có vật tư thiếu cần mua thêm cho kế hoạch này.",
      action:
        purchaseLineCount > 0
          ? {
              label: "Mở nhập kho",
              href: "/receiving",
              disabled: false
            }
          : undefined
    },
    {
      id: "inbound-qc",
      step: 5,
      title: "QC vật tư nhập",
      statusLabel: purchaseLineCount > 0 ? "Chờ QC vật tư nhập" : "Không chờ QC vật tư mua",
      statusTone: purchaseLineCount > 0 ? "info" : "success",
      detail:
        purchaseLineCount > 0
          ? "Vật tư nhập mua cần đi qua QC nếu mặt hàng yêu cầu kiểm soát."
          : "Không có dòng vật tư mua thêm cần QC từ kế hoạch này.",
      action:
        purchaseLineCount > 0
          ? {
              label: "Mở QC",
              href: "/qc",
              disabled: false
            }
          : undefined
    },
    {
      id: "subcontract-readiness",
      step: 6,
      title: "Sẵn sàng gia công",
      statusLabel: shortageLineCount > 0 ? "Chưa sẵn sàng" : "Sẵn sàng",
      statusTone: shortageLineCount > 0 ? "warning" : "success",
      detail:
        shortageLineCount > 0
          ? `Còn ${shortageLineCount} dòng thiếu vật tư; chưa nên tạo lệnh gia công.`
          : "Vật tư đã đủ theo nhu cầu kế hoạch, có thể chuyển sang gia công.",
      action:
        shortageLineCount > 0
          ? {
              label: "Chờ đủ vật tư",
              disabled: true
            }
          : {
              label: "Mở gia công",
              href: "/subcontract",
              disabled: false
            }
    },
    {
      id: "subcontract-order",
      step: 7,
      title: "Lệnh gia công",
      statusLabel: shortageLineCount > 0 ? "Chờ đủ vật tư" : "Chưa có dữ liệu lệnh gia công",
      statusTone: shortageLineCount > 0 ? "warning" : "info",
      detail:
        shortageLineCount > 0
          ? "Tạo lệnh gia công sau khi vật tư thiếu đã được mua, nhập kho và QC đạt."
          : "Mở module Gia công để tạo hoặc theo dõi lệnh sản xuất từ kế hoạch này.",
      action: {
        label: shortageLineCount > 0 ? "Chờ bước 6" : "Mở gia công",
        href: shortageLineCount > 0 ? undefined : "/subcontract",
        disabled: shortageLineCount > 0
      }
    }
  ];
}
