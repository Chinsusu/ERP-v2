import { describe, expect, it } from "vitest";
import {
  buildFactoryMaterialHandoverIssueInput,
  buildSubcontractFactoryMaterialHandover
} from "./subcontractFactoryMaterialHandover";
import type { SubcontractOrder } from "../types";

describe("subcontractFactoryMaterialHandover", () => {
  it("marks material handover ready after factory confirmation and paid deposit", () => {
    const handover = buildSubcontractFactoryMaterialHandover({
      ...baseOrder,
      status: "factory_confirmed",
      depositStatus: "paid",
      materialLines: [
        {
          ...baseOrder.materialLines[0],
          plannedQty: "1.500000",
          issuedQty: "0.500000"
        },
        {
          ...baseOrder.materialLines[1],
          plannedQty: "2.000000",
          issuedQty: "2.000000"
        }
      ]
    });

    expect(handover).toMatchObject({
      status: "ready",
      canIssue: true,
      totalLines: 2,
      completeLines: 1,
      pendingLines: 1
    });
    expect(handover.lines.map((line) => [line.id, line.remainingQty, line.status])).toEqual([
      ["sco-line-001", "1.000000", "ready"],
      ["sco-line-002", "0.000000", "complete"]
    ]);
  });

  it("blocks material handover until required deposit is recorded", () => {
    const handover = buildSubcontractFactoryMaterialHandover({
      ...baseOrder,
      status: "factory_confirmed",
      depositStatus: "pending"
    });

    expect(handover).toMatchObject({
      status: "blocked",
      canIssue: false,
      blockedReason: "Chờ ghi nhận đặt cọc trước khi bàn giao vật tư cho nhà máy."
    });
  });

  it("marks material handover complete when every line is fully issued", () => {
    const handover = buildSubcontractFactoryMaterialHandover({
      ...baseOrder,
      status: "materials_issued_to_factory",
      depositStatus: "paid",
      materialLines: baseOrder.materialLines.map((line) => ({
        ...line,
        issuedQty: line.plannedQty
      }))
    });

    expect(handover).toMatchObject({
      status: "complete",
      canIssue: false,
      completeLines: 2,
      pendingLines: 0
    });
  });

  it("builds an issue-materials payload only for lines that still need handover", () => {
    const order: SubcontractOrder = {
      ...baseOrder,
      status: "factory_confirmed",
      depositStatus: "paid",
      materialLines: [
        {
          ...baseOrder.materialLines[0],
          plannedQty: "1.500000",
          issuedQty: "0.500000"
        },
        {
          ...baseOrder.materialLines[1],
          plannedQty: "2.000000",
          issuedQty: "2.000000"
        }
      ]
    };

    const input = buildFactoryMaterialHandoverIssueInput({
      order,
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      handoverBy: "warehouse-user",
      receivedBy: "factory-receiver",
      receiverContact: "0900000000",
      vehicleNo: "51A-00000",
      note: "Bàn giao theo lệnh nhà máy",
      evidenceFileName: "handover.pdf",
      lineDrafts: {
        "sco-line-001": {
          issueQty: "1",
          batchNo: "RM-260506-A",
          sourceBinId: "BIN-A1"
        }
      }
    });

    expect(input).toMatchObject({
      order,
      sourceWarehouseId: "wh-hcm",
      sourceWarehouseCode: "HCM",
      handoverBy: "warehouse-user",
      receivedBy: "factory-receiver",
      receiverContact: "0900000000",
      vehicleNo: "51A-00000",
      note: "Bàn giao theo lệnh nhà máy",
      lines: [
        {
          orderMaterialLineId: "sco-line-001",
          issueQty: "1.000000",
          uomCode: "KG",
          batchNo: "RM-260506-A",
          sourceBinId: "BIN-A1"
        }
      ],
      evidence: [
        {
          evidenceType: "handover",
          fileName: "handover.pdf",
          note: "Biên bản bàn giao vật tư cho nhà máy"
        }
      ]
    });
  });
});

const baseOrder: SubcontractOrder = {
  id: "sco-001",
  orderNo: "SCO-260505-0001",
  factoryId: "factory-001",
  factoryCode: "FACTORY-001",
  factoryName: "Factory Partner",
  productId: "item-aah",
  sku: "AAH",
  productName: "Kem u phuc hoi AS A HABIT BIO 350GR",
  quantity: 999,
  uomCode: "PCS",
  receivedQty: "0.000000",
  acceptedQty: "0.000000",
  rejectedQty: "0.000000",
  sourceProductionPlanId: "plan-001",
  sourceProductionPlanNo: "PP-260505-0001",
  specVersion: "S23SMK260504200049",
  sampleRequired: true,
  expectedDeliveryDate: "2026-05-20",
  depositStatus: "pending",
  depositAmount: 1000000,
  finalPaymentStatus: "hold",
  status: "draft",
  createdBy: "Production Ops",
  createdAt: "2026-05-05T08:00:00Z",
  updatedAt: "2026-05-05T08:00:00Z",
  version: 1,
  estimatedCostAmount: "1000000.00",
  materialLines: [
    {
      id: "sco-line-001",
      itemId: "item-aci-bha",
      skuCode: "ACI_BHA",
      itemName: "ACID SALICYLIC",
      plannedQty: "1.000000",
      issuedQty: "0.000000",
      uomCode: "KG",
      unitCost: "0.000000",
      currencyCode: "VND",
      lineCostAmount: "0.00",
      lotTraceRequired: true
    },
    {
      id: "sco-line-002",
      itemId: "item-cpgc-01",
      skuCode: "CPGC-01",
      itemName: "Chai PET 100ml",
      plannedQty: "2.000000",
      issuedQty: "0.000000",
      uomCode: "PCS",
      unitCost: "0.000000",
      currencyCode: "VND",
      lineCostAmount: "0.00",
      lotTraceRequired: false
    }
  ],
  auditLogIds: []
};
