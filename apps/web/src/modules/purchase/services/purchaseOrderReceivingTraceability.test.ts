import { describe, expect, it } from "vitest";
import {
  buildPurchaseOrderReceiptRows,
  purchaseOrderReceiptListHref,
  purchaseOrderReceiptStatusLabel,
  purchaseOrderReceiptStatusTone
} from "./purchaseOrderReceivingTraceability";
import type { GoodsReceipt } from "../../receiving/types";
import type { PurchaseOrder, PurchaseOrderLine } from "../types";

describe("purchaseOrderReceivingTraceability", () => {
  it("builds receipt rows only for the selected purchase order", () => {
    const rows = buildPurchaseOrderReceiptRows(basePurchaseOrder, [
      { ...baseReceipt, id: "grn-other", referenceDocId: "po-other" },
      baseReceipt
    ]);

    expect(rows).toEqual([
      {
        id: "grn-po-001",
        receiptNo: "GRN-260505-001",
        status: "inspect_ready",
        statusLabel: "Sẵn sàng QC",
        statusTone: "warning",
        lineCount: 2,
        qcSummary: "Đạt 1 / Giữ 1",
        createdAt: "2026-05-05T12:00:00Z",
        postedAt: undefined,
        href: "/receiving?po_id=po-260505-249546&warehouse_id=wh-hcm-rm&status=inspect_ready#receiving-list"
      }
    ]);
  });

  it("formats receipt status for Vietnamese PO traceability", () => {
    expect(purchaseOrderReceiptStatusLabel("draft")).toBe("Nháp");
    expect(purchaseOrderReceiptStatusLabel("submitted")).toBe("Đã gửi");
    expect(purchaseOrderReceiptStatusLabel("inspect_ready")).toBe("Sẵn sàng QC");
    expect(purchaseOrderReceiptStatusLabel("posted")).toBe("Đã hạch toán");
    expect(purchaseOrderReceiptStatusTone("posted")).toBe("success");
  });

  it("builds a receiving list link scoped to the PO and warehouse", () => {
    expect(purchaseOrderReceiptListHref(basePurchaseOrder, "posted")).toBe(
      "/receiving?po_id=po-260505-249546&warehouse_id=wh-hcm-rm&status=posted#receiving-list"
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
  status: "approved",
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

const baseReceipt: GoodsReceipt = {
  id: "grn-po-001",
  orgId: "org-my-pham",
  receiptNo: "GRN-260505-001",
  warehouseId: "wh-hcm-rm",
  warehouseCode: "WH-HCM-RM",
  locationId: "loc-hcm-rm-recv-01",
  locationCode: "RM-RECV-01",
  referenceDocType: "purchase_order",
  referenceDocId: "po-260505-249546",
  supplierId: "sup-nguyen-ba",
  deliveryNoteNo: "DN-260505-001",
  status: "inspect_ready",
  lines: [
    {
      id: "grn-line-1",
      purchaseOrderLineId: "po-line-1",
      itemId: "item-aci-bha",
      sku: "ACI_BHA",
      itemName: "ACID SALICYLIC",
      warehouseId: "wh-hcm-rm",
      locationId: "loc-hcm-rm-recv-01",
      quantity: "0.050000",
      uomCode: "KG",
      baseUomCode: "KG",
      packagingStatus: "intact",
      qcStatus: "hold"
    },
    {
      id: "grn-line-2",
      purchaseOrderLineId: "po-line-1",
      itemId: "item-aci-bha",
      sku: "ACI_BHA",
      itemName: "ACID SALICYLIC",
      warehouseId: "wh-hcm-rm",
      locationId: "loc-hcm-rm-recv-01",
      quantity: "0.049900",
      uomCode: "KG",
      baseUomCode: "KG",
      packagingStatus: "intact",
      qcStatus: "pass"
    }
  ],
  createdBy: "user-warehouse",
  createdAt: "2026-05-05T12:00:00Z",
  updatedAt: "2026-05-05T12:30:00Z"
};
