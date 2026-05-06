import { describe, expect, it } from "vitest";
import { buildSupplierPayableFactoryCloseout } from "./supplierPayableFactoryCloseout";
import type { SupplierInvoice, SupplierPayable } from "../types";

describe("supplierPayableFactoryCloseout", () => {
  it("builds a factory final-payment checklist from subcontract AP source evidence", () => {
    const closeout = buildSupplierPayableFactoryCloseout(factoryFinalPaymentPayable, [], false);

    expect(closeout).toMatchObject({
      payableNo: "AP-SPM-S35-CLOSEOUT-FINAL",
      factoryOrderNo: "SCO-S35-CLOSEOUT",
      productionHref: "/production/factory-orders/sco-s35-closeout#factory-claim-final-payment-closeout",
      summaryTone: "warning"
    });
    expect(closeout?.steps.map(({ key, status }) => ({ key, status }))).toEqual([
      { key: "ap-created", status: "complete" },
      { key: "invoice-match", status: "current" },
      { key: "payment-request", status: "blocked" },
      { key: "payment-approval", status: "pending" },
      { key: "payment-recording", status: "pending" }
    ]);
  });

  it("moves the current step to payment request after the supplier invoice matches the AP", () => {
    const closeout = buildSupplierPayableFactoryCloseout(
      factoryFinalPaymentPayable,
      [matchedFactoryInvoice],
      false
    );

    expect(closeout).toMatchObject({
      invoiceNo: "INV-AP-SPM-S35-CLOSEOUT-FINAL",
      summaryTone: "success"
    });
    expect(closeout?.steps.map(({ key, status }) => ({ key, status }))).toEqual([
      { key: "ap-created", status: "complete" },
      { key: "invoice-match", status: "complete" },
      { key: "payment-request", status: "current" },
      { key: "payment-approval", status: "pending" },
      { key: "payment-recording", status: "pending" }
    ]);
  });

  it("tracks the approval and payment recording stages from AP status", () => {
    expect(
      buildSupplierPayableFactoryCloseout(
        { ...factoryFinalPaymentPayable, status: "payment_requested" },
        [matchedFactoryInvoice],
        false
      )?.steps.map(({ key, status }) => ({ key, status }))
    ).toEqual([
      { key: "ap-created", status: "complete" },
      { key: "invoice-match", status: "complete" },
      { key: "payment-request", status: "complete" },
      { key: "payment-approval", status: "current" },
      { key: "payment-recording", status: "pending" }
    ]);

    expect(
      buildSupplierPayableFactoryCloseout(
        { ...factoryFinalPaymentPayable, status: "payment_approved" },
        [matchedFactoryInvoice],
        false
      )?.steps.map(({ key, status }) => ({ key, status }))
    ).toEqual([
      { key: "ap-created", status: "complete" },
      { key: "invoice-match", status: "complete" },
      { key: "payment-request", status: "complete" },
      { key: "payment-approval", status: "complete" },
      { key: "payment-recording", status: "current" }
    ]);
  });

  it("returns no closeout for non-factory AP documents", () => {
    expect(
      buildSupplierPayableFactoryCloseout(
        {
          ...factoryFinalPaymentPayable,
          sourceDocument: { type: "purchase_order", id: "po-1", no: "PO-1" },
          lines: [
            {
              id: "ap-line-po",
              description: "Supplier material invoice",
              sourceDocument: { type: "warehouse_receipt", id: "gr-1", no: "GR-1" },
              amount: "29000000.00"
            }
          ]
        },
        [],
        false
      )
    ).toBeNull();
  });
});

const factoryFinalPaymentPayable: SupplierPayable = {
  id: "ap-spm-s35-closeout-final",
  payableNo: "AP-SPM-S35-CLOSEOUT-FINAL",
  supplierId: "factory-bd-002",
  supplierCode: "FACT-BD-002",
  supplierName: "Binh Duong Gia Cong",
  status: "open",
  sourceDocument: {
    type: "subcontract_payment_milestone",
    id: "spm-s35-closeout-final",
    no: "SPM-S35-CLOSEOUT-FINAL"
  },
  lines: [
    {
      id: "ap-line-s35-closeout-order",
      description: "Factory final payment after accepted finished goods",
      sourceDocument: { type: "subcontract_order", id: "sco-s35-closeout", no: "SCO-S35-CLOSEOUT" },
      amount: "29000000.00"
    }
  ],
  totalAmount: "29000000.00",
  paidAmount: "0.00",
  outstandingAmount: "29000000.00",
  currencyCode: "VND",
  dueDate: "2026-05-14",
  createdAt: "2026-05-07T08:00:00Z",
  updatedAt: "2026-05-07T08:00:00Z",
  version: 1
};

const matchedFactoryInvoice: SupplierInvoice = {
  id: "si-s35-closeout",
  invoiceNo: "INV-AP-SPM-S35-CLOSEOUT-FINAL",
  supplierId: "factory-bd-002",
  supplierCode: "FACT-BD-002",
  supplierName: "Binh Duong Gia Cong",
  payableId: "ap-spm-s35-closeout-final",
  payableNo: "AP-SPM-S35-CLOSEOUT-FINAL",
  status: "matched",
  matchStatus: "matched",
  sourceDocument: {
    type: "subcontract_payment_milestone",
    id: "spm-s35-closeout-final",
    no: "SPM-S35-CLOSEOUT-FINAL"
  },
  lines: [
    {
      id: "si-line-s35-closeout-order",
      description: "Factory final payment after accepted finished goods",
      sourceDocument: { type: "subcontract_order", id: "sco-s35-closeout", no: "SCO-S35-CLOSEOUT" },
      amount: "29000000.00"
    }
  ],
  invoiceAmount: "29000000.00",
  expectedAmount: "29000000.00",
  varianceAmount: "0.00",
  currencyCode: "VND",
  invoiceDate: "2026-05-07",
  createdAt: "2026-05-07T08:30:00Z",
  updatedAt: "2026-05-07T08:30:00Z",
  version: 1
};
