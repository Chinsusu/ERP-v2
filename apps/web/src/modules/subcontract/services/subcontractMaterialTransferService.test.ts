import { describe, expect, it } from "vitest";
import { prototypeSubcontractOrders } from "./subcontractOrderService";
import {
  createSubcontractMaterialTransfer,
  formatSubcontractAttachmentType,
  formatSubcontractTransferStatus,
  prototypeTransferLines,
  subcontractTransferStatusTone,
  summarizeSubcontractMaterialTransfers
} from "./subcontractMaterialTransferService";

describe("subcontractMaterialTransferService", () => {
  it("creates a material and packaging transfer with attachment placeholders and issue movements", () => {
    const [order] = prototypeSubcontractOrders;
    const transfer = createSubcontractMaterialTransfer({
      order,
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      signedHandover: true,
      lines: prototypeTransferLines
    });

    expect(transfer).toMatchObject({
      orderNo: "SUB-260426-0001",
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      factoryId: "factory-lotus",
      signedHandover: true,
      status: "SENT"
    });
    expect(transfer.attachmentPlaceholders).toEqual([
      { type: "COA", label: "COA", required: true, attached: false },
      { type: "MSDS", label: "MSDS", required: true, attached: false },
      { type: "LABEL", label: "Label", required: true, attached: false },
      { type: "VAT_INVOICE", label: "VAT invoice", required: true, attached: false }
    ]);
    expect(transfer.stockMovements).toHaveLength(2);
    expect(transfer.stockMovements[0]).toMatchObject({
      movementType: "SUBCONTRACT_ISSUE",
      itemCode: "RM-CREAM-BASE",
      batchNo: "RM-260426-A",
      targetLocation: "stock_in_subcontractor_hold:LOTUS"
    });
  });

  it("requires source warehouse and factory before creating the transfer document", () => {
    const [order] = prototypeSubcontractOrders;

    expect(() =>
      createSubcontractMaterialTransfer({
        order,
        sourceWarehouseId: "",
        sourceWarehouseCode: "HCM",
        signedHandover: true,
        lines: prototypeTransferLines
      })
    ).toThrow("Source warehouse is required");
  });

  it("requires batch or lot for lot controlled materials", () => {
    const [order] = prototypeSubcontractOrders;

    expect(() =>
      createSubcontractMaterialTransfer({
        order,
        sourceWarehouseId: "wh-hcm",
        sourceWarehouseCode: "HCM",
        signedHandover: true,
        lines: [
          {
            ...prototypeTransferLines[0],
            batchNo: undefined
          }
        ]
      })
    ).toThrow("RM-CREAM-BASE requires batch or lot before factory transfer");
  });

  it("blocks materials that have not passed QC", () => {
    const [order] = prototypeSubcontractOrders;

    expect(() =>
      createSubcontractMaterialTransfer({
        order,
        sourceWarehouseId: "wh-hcm",
        sourceWarehouseCode: "HCM",
        signedHandover: false,
        lines: [
          {
            ...prototypeTransferLines[0],
            qcStatus: "pending"
          }
        ]
      })
    ).toThrow("RM-CREAM-BASE must pass QC before factory transfer");
  });

  it("summarizes transfers and maps transfer UI labels", () => {
    const [order] = prototypeSubcontractOrders;
    const transfer = createSubcontractMaterialTransfer({
      order,
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      signedHandover: false,
      lines: prototypeTransferLines
    });

    expect(summarizeSubcontractMaterialTransfers([transfer])).toEqual({
      total: 1,
      signed: 0,
      movementCount: 2,
      attachmentPlaceholderCount: 4
    });
    expect(formatSubcontractTransferStatus("READY_TO_SEND")).toBe("Ready to send");
    expect(subcontractTransferStatusTone("SENT")).toBe("success");
    expect(formatSubcontractAttachmentType("VAT_INVOICE")).toBe("VAT invoice");
  });
});
