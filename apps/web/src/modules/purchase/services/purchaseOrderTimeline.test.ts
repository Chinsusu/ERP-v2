import { describe, expect, it } from "vitest";
import {
  buildPurchaseOrderTimeline,
  purchaseOrderReceivingHref,
  purchaseOrderSourcePlanNo,
  remainingPurchaseLineQuantity
} from "./purchaseOrderTimeline";
import type { PurchaseOrder, PurchaseOrderLine } from "../types";

describe("purchaseOrderTimeline", () => {
  it("builds an approved PO timeline with receiving as the current step", () => {
    const timeline = buildPurchaseOrderTimeline({
      ...basePurchaseOrder,
      status: "approved",
      submittedAt: "2026-05-05T09:00:00Z",
      approvedAt: "2026-05-05T10:00:00Z"
    });

    expect(timeline.map((item) => [item.id, item.status])).toEqual([
      ["created", "complete"],
      ["submitted", "complete"],
      ["approved", "complete"],
      ["receiving", "current"],
      ["closed", "pending"]
    ]);
    expect(timeline[2].occurredAt).toBe("2026-05-05T10:00:00Z");
    expect(timeline.find((item) => item.id === "receiving")?.action).toEqual({
      label: "Mở nhập hàng",
      href: "/receiving?po_id=po-260505-249546&warehouse_id=wh-hcm-rm#receiving-draft",
      disabled: false
    });
  });

  it("shows cancelled and rejected terminal states without marking downstream steps complete", () => {
    const cancelled = buildPurchaseOrderTimeline({
      ...basePurchaseOrder,
      status: "cancelled",
      submittedAt: "2026-05-05T09:00:00Z",
      cancelledAt: "2026-05-05T10:00:00Z",
      cancelReason: "Supplier delay"
    });
    const rejected = buildPurchaseOrderTimeline({
      ...basePurchaseOrder,
      status: "rejected",
      submittedAt: "2026-05-05T09:00:00Z",
      rejectedAt: "2026-05-05T10:00:00Z",
      rejectReason: "Wrong supplier"
    });

    expect(cancelled.map((item) => [item.id, item.status])).toEqual([
      ["created", "complete"],
      ["submitted", "complete"],
      ["approved", "blocked"],
      ["receiving", "blocked"],
      ["closed", "blocked"],
      ["cancelled", "complete"]
    ]);
    expect(cancelled.at(-1)).toMatchObject({
      id: "cancelled",
      description: "Supplier delay"
    });
    expect(rejected.at(-1)).toMatchObject({
      id: "rejected",
      status: "complete",
      description: "Wrong supplier"
    });
  });

  it("calculates remaining line quantity with six decimal places", () => {
    expect(
      remainingPurchaseLineQuantity({
        ...basePurchaseOrderLine,
        orderedQty: "12.500000",
        receivedQty: "2.125000"
      })
    ).toBe("10.375000");
  });

  it("does not return negative remaining quantity when a line is over-received", () => {
    expect(
      remainingPurchaseLineQuantity({
        ...basePurchaseOrderLine,
        orderedQty: "1.000000",
        receivedQty: "2.000000"
      })
    ).toBe("0.000000");
  });

  it("extracts the source production plan number from a PO note", () => {
    expect(
      purchaseOrderSourcePlanNo({
        ...basePurchaseOrder,
        note: "Tao tu ke hoach san xuat PP-260505-968033 / PR-DRAFT-260505-968033"
      })
    ).toBe("PP-260505-968033");
  });

  it("builds a receiving deep link for the PO warehouse context", () => {
    expect(purchaseOrderReceivingHref(basePurchaseOrder)).toBe(
      "/receiving?po_id=po-260505-249546&warehouse_id=wh-hcm-rm#receiving-draft"
    );
  });
});

const basePurchaseOrderLine: PurchaseOrderLine = {
  id: "po-line-1",
  lineNo: 1,
  itemId: "item-aci-bha",
  skuCode: "ACI_BHA",
  itemName: "ACID SALICYLIC",
  orderedQty: "1.000000",
  receivedQty: "0.000000",
  uomCode: "KG",
  baseOrderedQty: "1.000000",
  baseReceivedQty: "0.000000",
  baseUomCode: "KG",
  conversionFactor: "1.000000",
  unitPrice: "0.0000",
  currencyCode: "VND",
  lineAmount: "0.00",
  expectedDate: "2026-05-06"
};

const basePurchaseOrder: PurchaseOrder = {
  id: "po-260505-249546",
  poNo: "PO-260505-249546",
  supplierId: "sup-nguyen-ba",
  supplierCode: "SUP-NGUYEN-BA",
  supplierName: "Nguyen Ba",
  warehouseId: "wh-hcm-rm",
  warehouseCode: "WH-HCM-RM",
  expectedDate: "2026-05-06",
  status: "draft",
  currencyCode: "VND",
  subtotalAmount: "0.00",
  totalAmount: "0.00",
  lines: [basePurchaseOrderLine],
  lineCount: 1,
  receivedLineCount: 0,
  createdAt: "2026-05-05T08:00:00Z",
  updatedAt: "2026-05-05T08:00:00Z",
  version: 1
};
