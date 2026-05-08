import { describe, expect, it } from "vitest";
import { buildSupplierPayableFactoryCloseout } from "./supplierPayableFactoryCloseout";
import type { CashTransaction, SupplierInvoice, SupplierPayable } from "../types";

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
      { key: "payment-recording", status: "pending" },
      { key: "payment-voucher", status: "pending" }
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
      { key: "payment-recording", status: "pending" },
      { key: "payment-voucher", status: "pending" }
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
      { key: "payment-recording", status: "pending" },
      { key: "payment-voucher", status: "pending" }
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
      { key: "payment-recording", status: "current" },
      { key: "payment-voucher", status: "pending" }
    ]);
  });

  it("makes the voucher step current for paid factory AP without a posted cash-out voucher", () => {
    const closeout = buildSupplierPayableFactoryCloseout(
      paidFactoryFinalPaymentPayable,
      [matchedFactoryInvoice],
      false,
      [],
      false
    );

    expect(closeout).toMatchObject({
      summaryLabel: "Cần chứng từ chi",
      summaryTone: "warning",
      createVoucherHref:
        "/finance?cash_q=AP-SPM-S35-CLOSEOUT-FINAL&cash_direction=cash_out&cash_counterparty_id=factory-bd-002&cash_counterparty_name=Binh+Duong+Gia+Cong&cash_payment_method=bank_transfer&cash_reference_no=PAY-AP-SPM-S35-CLOSEOUT-FINAL&cash_amount=29000000.00&cash_target_type=supplier_payable&cash_target_id=ap-spm-s35-closeout-final&cash_target_no=AP-SPM-S35-CLOSEOUT-FINAL&cash_memo=Thanh+toan+AP+AP-SPM-S35-CLOSEOUT-FINAL%3B+cho+lenh+SCO-S35-CLOSEOUT%3B+hoa+don+INV-AP-SPM-S35-CLOSEOUT-FINAL%3B+moc+SPM-S35-CLOSEOUT-FINAL#cash-transactions"
    });
    expect(closeout?.steps.map(({ key, status }) => ({ key, status }))).toEqual([
      { key: "ap-created", status: "complete" },
      { key: "invoice-match", status: "complete" },
      { key: "payment-request", status: "complete" },
      { key: "payment-approval", status: "complete" },
      { key: "payment-recording", status: "complete" },
      { key: "payment-voucher", status: "current" }
    ]);
  });

  it("completes factory closeout when a posted cash-out voucher allocates to the AP", () => {
    const closeout = buildSupplierPayableFactoryCloseout(
      paidFactoryFinalPaymentPayable,
      [matchedFactoryInvoice],
      false,
      [factoryPaymentVoucher],
      false
    );

    expect(closeout).toMatchObject({
      summaryLabel: "Đã có chứng từ chi",
      summaryTone: "success",
      voucherNo: "CASH-OUT-S35-CLOSEOUT-FINAL",
      voucherReferenceNo: "BANK-S35-CLOSEOUT-FINAL",
      voucherHref: "/finance?cash_q=CASH-OUT-S35-CLOSEOUT-FINAL#cash-transactions"
    });
    expect(closeout?.steps.map(({ key, status }) => ({ key, status }))).toEqual([
      { key: "ap-created", status: "complete" },
      { key: "invoice-match", status: "complete" },
      { key: "payment-request", status: "complete" },
      { key: "payment-approval", status: "complete" },
      { key: "payment-recording", status: "complete" },
      { key: "payment-voucher", status: "complete" }
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

const paidFactoryFinalPaymentPayable: SupplierPayable = {
  ...factoryFinalPaymentPayable,
  status: "paid",
  paidAmount: "29000000.00",
  outstandingAmount: "0.00"
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

const factoryPaymentVoucher: CashTransaction = {
  id: "cash-s35-closeout-final",
  transactionNo: "CASH-OUT-S35-CLOSEOUT-FINAL",
  direction: "cash_out",
  status: "posted",
  businessDate: "2026-05-07",
  counterpartyId: "factory-bd-002",
  counterpartyName: "Binh Duong Gia Cong",
  paymentMethod: "bank_transfer",
  referenceNo: "BANK-S35-CLOSEOUT-FINAL",
  totalAmount: "29000000.00",
  currencyCode: "VND",
  allocations: [
    {
      id: "cash-s35-closeout-final-line-1",
      targetType: "supplier_payable",
      targetId: "ap-spm-s35-closeout-final",
      targetNo: "AP-SPM-S35-CLOSEOUT-FINAL",
      amount: "29000000.00"
    }
  ],
  postedBy: "finance-user",
  postedAt: "2026-05-07T09:00:00Z",
  createdAt: "2026-05-07T09:00:00Z",
  updatedAt: "2026-05-07T09:00:00Z",
  version: 1
};
