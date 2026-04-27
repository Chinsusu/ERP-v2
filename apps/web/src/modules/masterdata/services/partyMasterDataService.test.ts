import { beforeEach, describe, expect, it } from "vitest";
import {
  changeCustomerStatus,
  changeSupplierStatus,
  createCustomer,
  createSupplier,
  customerStatusLabel,
  customerTypeLabel,
  emptyCustomerInput,
  emptySupplierInput,
  getCustomers,
  getSuppliers,
  partyStatusTone,
  resetPrototypePartyMasterData,
  summarizeParties,
  supplierGroupLabel,
  supplierStatusLabel,
  updateCustomer,
  updateSupplier
} from "./partyMasterDataService";

describe("partyMasterDataService", () => {
  beforeEach(() => {
    resetPrototypePartyMasterData();
  });

  it("filters supplier master data by search, status, and group", async () => {
    await expect(
      getSuppliers({
        search: "bio",
        status: "active",
        supplierGroup: "raw_material"
      })
    ).resolves.toMatchObject([
      {
        supplierCode: "SUP-RM-BIO",
        status: "active",
        supplierGroup: "raw_material"
      }
    ]);
  });

  it("creates, updates, and changes supplier status in the local fallback store", async () => {
    const created = await createSupplier({
      ...emptySupplierInput,
      supplierCode: "sup-svc-lab",
      supplierName: "Lab Services Partner",
      supplierGroup: "service",
      email: "Lab@Partner.Example",
      taxCode: "0319999001",
      leadTimeDays: 5,
      qualityScore: 90,
      deliveryScore: 92
    });

    expect(created).toMatchObject({
      supplierCode: "SUP-SVC-LAB",
      email: "lab@partner.example",
      status: "draft"
    });
    expect(created.auditLogId).toContain("audit-local-supplier-create");

    const updated = await updateSupplier(created.id, {
      ...emptySupplierInput,
      supplierCode: "SUP-SVC-LAB",
      supplierName: "Lab Services Partner v2",
      supplierGroup: "service",
      status: "active"
    });
    expect(updated.supplierName).toBe("Lab Services Partner v2");

    const inactive = await changeSupplierStatus(created.id, "inactive");
    expect(inactive.status).toBe("inactive");
  });

  it("blocks duplicate supplier code and invalid supplier transition", async () => {
    await expect(
      createSupplier({
        ...emptySupplierInput,
        supplierCode: "SUP-RM-BIO",
        supplierName: "Duplicate Supplier"
      })
    ).rejects.toThrow("Supplier code already exists");

    const blacklisted = await changeSupplierStatus("sup-out-lotus", "blacklisted");
    expect(blacklisted.status).toBe("blacklisted");
    await expect(changeSupplierStatus("sup-out-lotus", "active")).rejects.toThrow("Supplier status transition is invalid");
  });

  it("filters customer master data by search, status, and type", async () => {
    await expect(
      getCustomers({
        search: "minh",
        status: "active",
        customerType: "distributor"
      })
    ).resolves.toMatchObject([
      {
        customerCode: "CUS-DL-MINHANH",
        status: "active",
        customerType: "distributor"
      }
    ]);
  });

  it("creates, updates, and changes customer status in the local fallback store", async () => {
    const created = await createCustomer({
      ...emptyCustomerInput,
      customerCode: "cus-dl-hanoi",
      customerName: "Ha Noi Dealer",
      channelCode: "dealer",
      priceListCode: "pl-dealer-2026",
      creditLimit: 200000000,
      email: "Buyer@HaNoiDealer.Example"
    });

    expect(created).toMatchObject({
      customerCode: "CUS-DL-HANOI",
      channelCode: "DEALER",
      email: "buyer@hanoidealer.example"
    });
    expect(created.auditLogId).toContain("audit-local-customer-create");

    const updated = await updateCustomer(created.id, {
      ...emptyCustomerInput,
      customerCode: "CUS-DL-HANOI",
      customerName: "Ha Noi Dealer v2",
      customerType: "dealer",
      status: "active"
    });
    expect(updated.customerName).toBe("Ha Noi Dealer v2");

    const inactive = await changeCustomerStatus(created.id, "inactive");
    expect(inactive.status).toBe("inactive");
  });

  it("blocks duplicate customer code and invalid customer transition", async () => {
    await expect(
      createCustomer({
        ...emptyCustomerInput,
        customerCode: "CUS-DL-MINHANH",
        customerName: "Duplicate Customer"
      })
    ).rejects.toThrow("Customer code already exists");

    const blocked = await changeCustomerStatus("cus-internal-hcm-store", "blocked");
    expect(blocked.status).toBe("blocked");
    await expect(changeCustomerStatus("cus-internal-hcm-store", "active")).rejects.toThrow("Customer status transition is invalid");
  });

  it("summarizes party master operating counts", async () => {
    const suppliers = await getSuppliers();
    const customers = await getCustomers();

    expect(summarizeParties(suppliers, customers)).toEqual({
      suppliers: 4,
      activeSuppliers: 3,
      customers: 4,
      activeCustomers: 3
    });
  });

  it("maps labels and tones", () => {
    expect(partyStatusTone("active")).toBe("success");
    expect(partyStatusTone("blacklisted")).toBe("danger");
    expect(supplierGroupLabel("raw_material")).toBe("Raw material");
    expect(supplierStatusLabel("draft")).toBe("Draft");
    expect(customerTypeLabel("internal_store")).toBe("Internal store");
    expect(customerStatusLabel("blocked")).toBe("Blocked");
  });
});
