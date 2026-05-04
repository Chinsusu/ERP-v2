import { afterEach, describe, expect, it, vi } from "vitest";
import { createProductionPlan, createProductionPlans, formatProductionPlanQuantity, summarizeProductionPlans } from "./productionPlanService";

describe("productionPlanService", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("creates a production plan and maps material shortages into an internal purchase request draft", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({
          success: true,
          request_id: "req-production-plan",
          data: {
            id: "plan-001",
            org_id: "org-my-pham",
            plan_no: "PP-260504-0001",
            output_item_id: "item-xff-150",
            output_sku: "XFF",
            output_item_name: "Tinh chat buoi Fast & Furious 150ML",
            output_item_type: "finished_good",
            planned_qty: "162.000000",
            uom_code: "PCS",
            formula_id: "formula-xff-v1",
            formula_code: "XFF-150ML",
            formula_version: "v1",
            formula_batch_qty: "81.000000",
            formula_batch_uom_code: "PCS",
            status: "purchase_request_draft_created",
            lines: [
              {
                id: "pp-line-plan-001-001",
                formula_line_id: "formula-line-001",
                line_no: 1,
                component_item_id: "item-act-baicapil",
                component_sku: "ACT_BAICAPIL",
                component_name: "BAICAPIL",
                component_type: "raw_material",
                formula_qty: "1.000000",
                formula_uom_code: "G",
                required_qty: "2.000000",
                required_uom_code: "G",
                required_stock_base_qty: "0.002000",
                stock_base_uom_code: "KG",
                available_qty: "0.000500",
                shortage_qty: "0.001500",
                purchase_draft_qty: "0.001500",
                purchase_draft_uom_code: "KG",
                is_stock_managed: true,
                needs_purchase: true
              }
            ],
            purchase_request_draft: {
              id: "pr-draft-001",
              request_no: "PR-DRAFT-260504-0001",
              source_production_plan_id: "plan-001",
              source_production_plan_no: "PP-260504-0001",
              status: "draft",
              lines: [
                {
                  id: "pr-line-001",
                  line_no: 1,
                  source_production_plan_line_id: "pp-line-plan-001-001",
                  item_id: "item-act-baicapil",
                  sku: "ACT_BAICAPIL",
                  item_name: "BAICAPIL",
                  requested_qty: "0.001500",
                  uom_code: "KG"
                }
              ],
              created_at: "2026-05-04T03:00:00Z",
              created_by: "user-production"
            },
            audit_log_id: "audit-production-plan-001",
            created_at: "2026-05-04T03:00:00Z",
            created_by: "user-production",
            updated_at: "2026-05-04T03:00:00Z",
            updated_by: "user-production",
            version: 1
          }
        }),
        { status: 201, headers: { "Content-Type": "application/json" } }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const plan = await createProductionPlan({
      outputItemId: "item-xff-150",
      formulaId: "formula-xff-v1",
      plannedQty: "162",
      uomCode: "pcs",
      plannedStartDate: "2026-05-10",
      plannedEndDate: "2026-05-11"
    });

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/production-plans"),
      expect.objectContaining({
        method: "POST",
        body: expect.stringContaining('"planned_qty":"162.000000"')
      })
    );
    expect(plan.lines[0]).toMatchObject({
      componentSku: "ACT_BAICAPIL",
      requiredStockBaseQty: "0.002000",
      shortageQty: "0.001500",
      needsPurchase: true
    });
    expect(plan.purchaseRequestDraft.lines[0]).toMatchObject({
      sku: "ACT_BAICAPIL",
      requestedQty: "0.001500",
      uomCode: "KG"
    });
    expect(summarizeProductionPlans([plan])).toEqual({
      total: 1,
      draft: 0,
      shortageLines: 1,
      purchaseDraftLines: 1
    });
  });

  it("formats small metric quantities for readable material demand review", () => {
    expect(formatProductionPlanQuantity("0.000001", "KG")).toBe("1 mg");
    expect(formatProductionPlanQuantity("0.001500", "KG")).toBe("1,5 g");
    expect(formatProductionPlanQuantity("2.000000", "PCS")).toBe("2 PCS");
  });

  it("creates multiple production plans from one submission", async () => {
    const fetchMock = vi.fn(async (_url, init) => {
      const body = JSON.parse(String((init as RequestInit).body));
      const sku = body.output_item_id === "item-aah" ? "AAH" : "XFF";

      return new Response(
        JSON.stringify({
          success: true,
          request_id: `req-${sku.toLowerCase()}`,
          data: {
            id: `plan-${sku.toLowerCase()}`,
            org_id: "org-my-pham",
            plan_no: `PP-260504-${sku}`,
            output_item_id: body.output_item_id,
            output_sku: sku,
            output_item_name: sku,
            output_item_type: "finished_good",
            planned_qty: body.planned_qty,
            uom_code: body.uom_code,
            formula_id: body.formula_id,
            formula_code: sku,
            formula_version: "v1",
            formula_batch_qty: "1.000000",
            formula_batch_uom_code: body.uom_code,
            status: "purchase_request_draft_created",
            lines: [],
            purchase_request_draft: { lines: [] },
            created_at: "2026-05-04T03:00:00Z",
            created_by: "user-production",
            updated_at: "2026-05-04T03:00:00Z",
            updated_by: "user-production",
            version: 1
          }
        }),
        { status: 201, headers: { "Content-Type": "application/json" } }
      );
    });
    vi.stubGlobal("fetch", fetchMock);

    const plans = await createProductionPlans([
      {
        outputItemId: "item-aah",
        formulaId: "formula-aah-v1",
        plannedQty: "10",
        uomCode: "pcs"
      },
      {
        outputItemId: "item-xff",
        formulaId: "formula-xff-v1",
        plannedQty: "50",
        uomCode: "pcs"
      }
    ]);

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(fetchMock.mock.calls[0][1]).toEqual(
      expect.objectContaining({
        body: expect.stringContaining('"output_item_id":"item-aah"')
      })
    );
    expect(fetchMock.mock.calls[1][1]).toEqual(
      expect.objectContaining({
        body: expect.stringContaining('"output_item_id":"item-xff"')
      })
    );
    expect(plans.map((plan) => plan.outputSku)).toEqual(["AAH", "XFF"]);
  });

  it("rejects empty multi-plan submissions before calling the API", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(createProductionPlans([])).rejects.toThrow("At least one production plan line is required");
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
