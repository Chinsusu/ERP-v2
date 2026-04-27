import { ApiError, apiGet, apiGetRaw, apiPatch, apiPost } from "../../../shared/api/client";
import type { components, operations } from "../../../shared/api/generated/schema";
import type {
  CustomerMasterDataInput,
  CustomerMasterDataItem,
  CustomerMasterDataQuery,
  CustomerStatus,
  CustomerType,
  PartyMasterDataSummary,
  SupplierGroup,
  SupplierMasterDataInput,
  SupplierMasterDataItem,
  SupplierMasterDataQuery,
  SupplierStatus
} from "../types";

type SupplierApiItem = components["schemas"]["SupplierListItem"];
type SupplierApiQuery = operations["listSuppliers"]["parameters"]["query"];
type SupplierApiCreateRequest = components["schemas"]["CreateSupplierRequest"];
type SupplierApiUpdateRequest = components["schemas"]["UpdateSupplierRequest"];
type SupplierApiStatusRequest = components["schemas"]["ChangeSupplierStatusRequest"];
type CustomerApiItem = components["schemas"]["CustomerListItem"];
type CustomerApiQuery = operations["listCustomers"]["parameters"]["query"];
type CustomerApiCreateRequest = components["schemas"]["CreateCustomerRequest"];
type CustomerApiUpdateRequest = components["schemas"]["UpdateCustomerRequest"];
type CustomerApiStatusRequest = components["schemas"]["ChangeCustomerStatusRequest"];

const defaultAccessToken = "local-dev-access-token";

export const supplierGroupOptions: { label: string; value: SupplierGroup }[] = [
  { label: "Raw material", value: "raw_material" },
  { label: "Packaging", value: "packaging" },
  { label: "Service", value: "service" },
  { label: "Logistics", value: "logistics" },
  { label: "Outsource", value: "outsource" }
];

export const supplierStatusOptions: { label: string; value: SupplierStatus }[] = [
  { label: "Draft", value: "draft" },
  { label: "Active", value: "active" },
  { label: "Inactive", value: "inactive" },
  { label: "Blacklisted", value: "blacklisted" }
];

export const customerTypeOptions: { label: string; value: CustomerType }[] = [
  { label: "Distributor", value: "distributor" },
  { label: "Dealer", value: "dealer" },
  { label: "Retail customer", value: "retail_customer" },
  { label: "Marketplace", value: "marketplace" },
  { label: "Internal store", value: "internal_store" }
];

export const customerStatusOptions: { label: string; value: CustomerStatus }[] = [
  { label: "Draft", value: "draft" },
  { label: "Active", value: "active" },
  { label: "Inactive", value: "inactive" },
  { label: "Blocked", value: "blocked" }
];

export const emptySupplierInput: SupplierMasterDataInput = {
  supplierCode: "",
  supplierName: "",
  supplierGroup: "raw_material",
  contactName: "",
  phone: "",
  email: "",
  taxCode: "",
  address: "",
  paymentTerms: "NET30",
  leadTimeDays: 0,
  moq: 0,
  qualityScore: 0,
  deliveryScore: 0,
  status: "draft"
};

export const emptyCustomerInput: CustomerMasterDataInput = {
  customerCode: "",
  customerName: "",
  customerType: "dealer",
  channelCode: "DEALER",
  priceListCode: "PL-DEALER-2026",
  discountGroup: "",
  creditLimit: 0,
  paymentTerms: "NET15",
  contactName: "",
  phone: "",
  email: "",
  taxCode: "",
  address: "",
  status: "draft"
};

export const prototypeSuppliers: SupplierMasterDataItem[] = [
  {
    id: "sup-rm-bioactive",
    supplierCode: "SUP-RM-BIO",
    supplierName: "BioActive Raw Materials",
    supplierGroup: "raw_material",
    contactName: "Nguyen Van An",
    phone: "+84901234501",
    email: "purchasing@bioactive.example",
    taxCode: "0312345001",
    address: "Binh Duong raw material hub",
    paymentTerms: "NET30",
    leadTimeDays: 12,
    moq: 50,
    qualityScore: 94,
    deliveryScore: 91,
    status: "active",
    createdAt: "2026-04-26T11:00:00Z",
    updatedAt: "2026-04-26T11:00:00Z"
  },
  {
    id: "sup-pkg-vina",
    supplierCode: "SUP-PKG-VINA",
    supplierName: "Vina Packaging Solutions",
    supplierGroup: "packaging",
    contactName: "Tran Thi Binh",
    phone: "+84901234502",
    email: "sales@vinapack.example",
    taxCode: "0312345002",
    address: "Long An packaging park",
    paymentTerms: "NET45",
    leadTimeDays: 8,
    moq: 1000,
    qualityScore: 89,
    deliveryScore: 88,
    status: "active",
    createdAt: "2026-04-26T11:10:00Z",
    updatedAt: "2026-04-26T11:10:00Z"
  },
  {
    id: "sup-log-fastgo",
    supplierCode: "SUP-LOG-FASTGO",
    supplierName: "FastGo Logistics",
    supplierGroup: "logistics",
    contactName: "Le Minh Chau",
    phone: "+84901234503",
    email: "ops@fastgo.example",
    taxCode: "0312345003",
    address: "Ho Chi Minh logistics center",
    paymentTerms: "NET15",
    leadTimeDays: 2,
    moq: 0,
    qualityScore: 87,
    deliveryScore: 93,
    status: "active",
    createdAt: "2026-04-26T11:20:00Z",
    updatedAt: "2026-04-26T11:20:00Z"
  },
  {
    id: "sup-out-lotus",
    supplierCode: "SUP-OUT-LOTUS",
    supplierName: "Lotus Filling Partner",
    supplierGroup: "outsource",
    contactName: "Pham Quoc Duy",
    phone: "+84901234504",
    email: "qa@lotusfill.example",
    taxCode: "0312345004",
    address: "Dong Nai outsource site",
    paymentTerms: "NET30",
    leadTimeDays: 15,
    moq: 500,
    qualityScore: 82,
    deliveryScore: 80,
    status: "inactive",
    createdAt: "2026-04-26T11:30:00Z",
    updatedAt: "2026-04-26T11:30:00Z"
  }
];

export const prototypeCustomers: CustomerMasterDataItem[] = [
  {
    id: "cus-dl-minh-anh",
    customerCode: "CUS-DL-MINHANH",
    customerName: "Minh Anh Distributor",
    customerType: "distributor",
    channelCode: "B2B",
    priceListCode: "PL-B2B-2026",
    discountGroup: "tier_1",
    creditLimit: 500000000,
    paymentTerms: "NET30",
    contactName: "Do Minh Anh",
    phone: "+84909888111",
    email: "orders@minhanh.example",
    taxCode: "0315678001",
    address: "District 7, Ho Chi Minh City",
    status: "active",
    createdAt: "2026-04-26T12:00:00Z",
    updatedAt: "2026-04-26T12:00:00Z"
  },
  {
    id: "cus-dealer-linh-chi",
    customerCode: "CUS-DL-LINHCHI",
    customerName: "Linh Chi Dealer",
    customerType: "dealer",
    channelCode: "DEALER",
    priceListCode: "PL-DEALER-2026",
    discountGroup: "tier_2",
    creditLimit: 150000000,
    paymentTerms: "NET15",
    contactName: "Nguyen Linh Chi",
    phone: "+84909888222",
    email: "buyer@linhchi.example",
    taxCode: "0315678002",
    address: "Thu Duc City",
    status: "active",
    createdAt: "2026-04-26T12:10:00Z",
    updatedAt: "2026-04-26T12:10:00Z"
  },
  {
    id: "cus-mp-shopee",
    customerCode: "CUS-MP-SHOPEE",
    customerName: "Shopee Marketplace",
    customerType: "marketplace",
    channelCode: "MP",
    priceListCode: "PL-MP-2026",
    discountGroup: "marketplace",
    creditLimit: 0,
    paymentTerms: "PREPAID",
    contactName: "Marketplace Ops",
    phone: "+84909888333",
    email: "ops@marketplace.example",
    taxCode: "0315678003",
    address: "Marketplace fulfillment channel",
    status: "active",
    createdAt: "2026-04-26T12:20:00Z",
    updatedAt: "2026-04-26T12:20:00Z"
  },
  {
    id: "cus-internal-hcm-store",
    customerCode: "CUS-INT-HCMSTORE",
    customerName: "HCM Internal Store",
    customerType: "internal_store",
    channelCode: "INT",
    priceListCode: "PL-INT-2026",
    discountGroup: "internal",
    creditLimit: 0,
    paymentTerms: "INTERNAL",
    contactName: "Store Lead",
    phone: "+84909888444",
    email: "store-hcm@example.local",
    taxCode: "0315678004",
    address: "Ho Chi Minh flagship store",
    status: "draft",
    createdAt: "2026-04-26T12:30:00Z",
    updatedAt: "2026-04-26T12:30:00Z"
  }
];

let localSuppliers = cloneSuppliers(prototypeSuppliers);
let localCustomers = cloneCustomers(prototypeCustomers);

export async function getSuppliers(query: SupplierMasterDataQuery = {}): Promise<SupplierMasterDataItem[]> {
  try {
    const items = await apiGet("/suppliers", {
      accessToken: defaultAccessToken,
      query: toSupplierApiQuery(query)
    });

    return items.map(fromSupplierApiItem);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    return filterSuppliers(localSuppliers, query);
  }
}

export async function getSupplier(supplierId: string): Promise<SupplierMasterDataItem> {
  try {
    const item = await apiGetRaw<SupplierApiItem>(`/suppliers/${encodeURIComponent(supplierId)}`, {
      accessToken: defaultAccessToken
    });

    return fromSupplierApiItem(item);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    const item = localSuppliers.find((candidate) => candidate.id === supplierId);
    if (!item) {
      throw new Error("Supplier master data was not found");
    }

    return { ...item };
  }
}

export async function createSupplier(input: SupplierMasterDataInput): Promise<SupplierMasterDataItem> {
  const normalized = normalizeSupplierInput(input);
  validateSupplierInput(normalized);

  try {
    const item = await apiPost<SupplierApiItem, SupplierApiCreateRequest>("/suppliers", toSupplierApiRequest(normalized), {
      accessToken: defaultAccessToken
    });

    return fromSupplierApiItem(item);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    ensureUniqueSupplier(normalized);
    const now = new Date().toISOString();
    const item: SupplierMasterDataItem = {
      ...normalized,
      id: `sup-${normalized.supplierCode.toLowerCase().replaceAll("-", "_")}-${Date.now()}`,
      createdAt: now,
      updatedAt: now,
      auditLogId: `audit-local-supplier-create-${Date.now()}`
    };
    localSuppliers = sortSuppliers([...localSuppliers, item]);

    return { ...item };
  }
}

export async function updateSupplier(supplierId: string, input: SupplierMasterDataInput): Promise<SupplierMasterDataItem> {
  const normalized = normalizeSupplierInput(input);
  validateSupplierInput(normalized);

  try {
    const item = await apiPatch<SupplierApiItem, SupplierApiUpdateRequest>(
      `/suppliers/${encodeURIComponent(supplierId)}`,
      toSupplierApiRequest(normalized),
      {
        accessToken: defaultAccessToken
      }
    );

    return fromSupplierApiItem(item);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    ensureUniqueSupplier(normalized, supplierId);
    const current = localSuppliers.find((candidate) => candidate.id === supplierId);
    if (!current) {
      throw new Error("Supplier master data was not found");
    }
    const item: SupplierMasterDataItem = {
      ...current,
      ...normalized,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-supplier-update-${Date.now()}`
    };
    localSuppliers = sortSuppliers(localSuppliers.map((candidate) => (candidate.id === supplierId ? item : candidate)));

    return { ...item };
  }
}

export async function changeSupplierStatus(supplierId: string, status: SupplierStatus): Promise<SupplierMasterDataItem> {
  try {
    const item = await apiPatch<SupplierApiItem, SupplierApiStatusRequest>(
      `/suppliers/${encodeURIComponent(supplierId)}/status`,
      { status },
      {
        accessToken: defaultAccessToken
      }
    );

    return fromSupplierApiItem(item);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    const current = localSuppliers.find((candidate) => candidate.id === supplierId);
    if (!current) {
      throw new Error("Supplier master data was not found");
    }
    if (current.status === "blacklisted" && status === "active") {
      throw new Error("Supplier status transition is invalid");
    }
    const item: SupplierMasterDataItem = {
      ...current,
      status,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-supplier-status-${Date.now()}`
    };
    localSuppliers = sortSuppliers(localSuppliers.map((candidate) => (candidate.id === supplierId ? item : candidate)));

    return { ...item };
  }
}

export async function getCustomers(query: CustomerMasterDataQuery = {}): Promise<CustomerMasterDataItem[]> {
  try {
    const items = await apiGet("/customers", {
      accessToken: defaultAccessToken,
      query: toCustomerApiQuery(query)
    });

    return items.map(fromCustomerApiItem);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    return filterCustomers(localCustomers, query);
  }
}

export async function getCustomer(customerId: string): Promise<CustomerMasterDataItem> {
  try {
    const item = await apiGetRaw<CustomerApiItem>(`/customers/${encodeURIComponent(customerId)}`, {
      accessToken: defaultAccessToken
    });

    return fromCustomerApiItem(item);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    const item = localCustomers.find((candidate) => candidate.id === customerId);
    if (!item) {
      throw new Error("Customer master data was not found");
    }

    return { ...item };
  }
}

export async function createCustomer(input: CustomerMasterDataInput): Promise<CustomerMasterDataItem> {
  const normalized = normalizeCustomerInput(input);
  validateCustomerInput(normalized);

  try {
    const item = await apiPost<CustomerApiItem, CustomerApiCreateRequest>("/customers", toCustomerApiRequest(normalized), {
      accessToken: defaultAccessToken
    });

    return fromCustomerApiItem(item);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    ensureUniqueCustomer(normalized);
    const now = new Date().toISOString();
    const item: CustomerMasterDataItem = {
      ...normalized,
      id: `cus-${normalized.customerCode.toLowerCase().replaceAll("-", "_")}-${Date.now()}`,
      createdAt: now,
      updatedAt: now,
      auditLogId: `audit-local-customer-create-${Date.now()}`
    };
    localCustomers = sortCustomers([...localCustomers, item]);

    return { ...item };
  }
}

export async function updateCustomer(customerId: string, input: CustomerMasterDataInput): Promise<CustomerMasterDataItem> {
  const normalized = normalizeCustomerInput(input);
  validateCustomerInput(normalized);

  try {
    const item = await apiPatch<CustomerApiItem, CustomerApiUpdateRequest>(
      `/customers/${encodeURIComponent(customerId)}`,
      toCustomerApiRequest(normalized),
      {
        accessToken: defaultAccessToken
      }
    );

    return fromCustomerApiItem(item);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    ensureUniqueCustomer(normalized, customerId);
    const current = localCustomers.find((candidate) => candidate.id === customerId);
    if (!current) {
      throw new Error("Customer master data was not found");
    }
    const item: CustomerMasterDataItem = {
      ...current,
      ...normalized,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-customer-update-${Date.now()}`
    };
    localCustomers = sortCustomers(localCustomers.map((candidate) => (candidate.id === customerId ? item : candidate)));

    return { ...item };
  }
}

export async function changeCustomerStatus(customerId: string, status: CustomerStatus): Promise<CustomerMasterDataItem> {
  try {
    const item = await apiPatch<CustomerApiItem, CustomerApiStatusRequest>(
      `/customers/${encodeURIComponent(customerId)}/status`,
      { status },
      {
        accessToken: defaultAccessToken
      }
    );

    return fromCustomerApiItem(item);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    const current = localCustomers.find((candidate) => candidate.id === customerId);
    if (!current) {
      throw new Error("Customer master data was not found");
    }
    if (current.status === "blocked" && status === "active") {
      throw new Error("Customer status transition is invalid");
    }
    const item: CustomerMasterDataItem = {
      ...current,
      status,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-customer-status-${Date.now()}`
    };
    localCustomers = sortCustomers(localCustomers.map((candidate) => (candidate.id === customerId ? item : candidate)));

    return { ...item };
  }
}

export function summarizeParties(suppliers: SupplierMasterDataItem[], customers: CustomerMasterDataItem[]): PartyMasterDataSummary {
  return {
    suppliers: suppliers.length,
    activeSuppliers: suppliers.filter((supplier) => supplier.status === "active").length,
    customers: customers.length,
    activeCustomers: customers.filter((customer) => customer.status === "active").length
  };
}

export function partyStatusTone(status: SupplierStatus | CustomerStatus): "normal" | "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "active":
      return "success";
    case "draft":
      return "info";
    case "inactive":
      return "warning";
    case "blacklisted":
    case "blocked":
      return "danger";
    default:
      return "normal";
  }
}

export function supplierGroupLabel(group: SupplierGroup) {
  return supplierGroupOptions.find((option) => option.value === group)?.label ?? group;
}

export function supplierStatusLabel(status: SupplierStatus) {
  return supplierStatusOptions.find((option) => option.value === status)?.label ?? status;
}

export function customerTypeLabel(type: CustomerType) {
  return customerTypeOptions.find((option) => option.value === type)?.label ?? type;
}

export function customerStatusLabel(status: CustomerStatus) {
  return customerStatusOptions.find((option) => option.value === status)?.label ?? status;
}

export function toSupplierInput(item: SupplierMasterDataItem): SupplierMasterDataInput {
  return {
    supplierCode: item.supplierCode,
    supplierName: item.supplierName,
    supplierGroup: item.supplierGroup,
    contactName: item.contactName ?? "",
    phone: item.phone ?? "",
    email: item.email ?? "",
    taxCode: item.taxCode ?? "",
    address: item.address ?? "",
    paymentTerms: item.paymentTerms ?? "",
    leadTimeDays: item.leadTimeDays ?? 0,
    moq: item.moq ?? 0,
    qualityScore: item.qualityScore ?? 0,
    deliveryScore: item.deliveryScore ?? 0,
    status: item.status
  };
}

export function toCustomerInput(item: CustomerMasterDataItem): CustomerMasterDataInput {
  return {
    customerCode: item.customerCode,
    customerName: item.customerName,
    customerType: item.customerType,
    channelCode: item.channelCode ?? "",
    priceListCode: item.priceListCode ?? "",
    discountGroup: item.discountGroup ?? "",
    creditLimit: item.creditLimit ?? 0,
    paymentTerms: item.paymentTerms ?? "",
    contactName: item.contactName ?? "",
    phone: item.phone ?? "",
    email: item.email ?? "",
    taxCode: item.taxCode ?? "",
    address: item.address ?? "",
    status: item.status
  };
}

export function resetPrototypePartyMasterData() {
  localSuppliers = cloneSuppliers(prototypeSuppliers);
  localCustomers = cloneCustomers(prototypeCustomers);
}

function fromSupplierApiItem(item: SupplierApiItem): SupplierMasterDataItem {
  return {
    id: item.id,
    supplierCode: item.supplier_code,
    supplierName: item.supplier_name,
    supplierGroup: item.supplier_group,
    contactName: item.contact_name,
    phone: item.phone,
    email: item.email,
    taxCode: item.tax_code,
    address: item.address,
    paymentTerms: item.payment_terms,
    leadTimeDays: item.lead_time_days,
    moq: item.moq,
    qualityScore: item.quality_score,
    deliveryScore: item.delivery_score,
    status: item.status,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    auditLogId: item.audit_log_id
  };
}

function fromCustomerApiItem(item: CustomerApiItem): CustomerMasterDataItem {
  return {
    id: item.id,
    customerCode: item.customer_code,
    customerName: item.customer_name,
    customerType: item.customer_type,
    channelCode: item.channel_code,
    priceListCode: item.price_list_code,
    discountGroup: item.discount_group,
    creditLimit: item.credit_limit,
    paymentTerms: item.payment_terms,
    contactName: item.contact_name,
    phone: item.phone,
    email: item.email,
    taxCode: item.tax_code,
    address: item.address,
    status: item.status,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    auditLogId: item.audit_log_id
  };
}

function toSupplierApiQuery(query: SupplierMasterDataQuery): SupplierApiQuery {
  return {
    q: query.search,
    status: query.status || undefined,
    supplier_group: query.supplierGroup || undefined,
    page: 1,
    page_size: 100
  };
}

function toCustomerApiQuery(query: CustomerMasterDataQuery): CustomerApiQuery {
  return {
    q: query.search,
    status: query.status || undefined,
    customer_type: query.customerType || undefined,
    page: 1,
    page_size: 100
  };
}

function toSupplierApiRequest(input: SupplierMasterDataInput): SupplierApiCreateRequest {
  return {
    supplier_code: input.supplierCode,
    supplier_name: input.supplierName,
    supplier_group: input.supplierGroup,
    contact_name: input.contactName || undefined,
    phone: input.phone || undefined,
    email: input.email || undefined,
    tax_code: input.taxCode || undefined,
    address: input.address || undefined,
    payment_terms: input.paymentTerms || undefined,
    lead_time_days: input.leadTimeDays,
    moq: input.moq,
    quality_score: input.qualityScore,
    delivery_score: input.deliveryScore,
    status: input.status
  };
}

function toCustomerApiRequest(input: CustomerMasterDataInput): CustomerApiCreateRequest {
  return {
    customer_code: input.customerCode,
    customer_name: input.customerName,
    customer_type: input.customerType,
    channel_code: input.channelCode || undefined,
    price_list_code: input.priceListCode || undefined,
    discount_group: input.discountGroup || undefined,
    credit_limit: input.creditLimit,
    payment_terms: input.paymentTerms || undefined,
    contact_name: input.contactName || undefined,
    phone: input.phone || undefined,
    email: input.email || undefined,
    tax_code: input.taxCode || undefined,
    address: input.address || undefined,
    status: input.status
  };
}

function normalizeSupplierInput(input: SupplierMasterDataInput): SupplierMasterDataInput {
  return {
    ...input,
    supplierCode: input.supplierCode.trim().toUpperCase(),
    supplierName: input.supplierName.trim(),
    contactName: input.contactName.trim(),
    phone: input.phone.trim(),
    email: input.email.trim().toLowerCase(),
    taxCode: input.taxCode.trim().toUpperCase(),
    address: input.address.trim(),
    paymentTerms: input.paymentTerms.trim().toUpperCase(),
    leadTimeDays: finiteNumber(input.leadTimeDays),
    moq: finiteNumber(input.moq),
    qualityScore: finiteNumber(input.qualityScore),
    deliveryScore: finiteNumber(input.deliveryScore)
  };
}

function normalizeCustomerInput(input: CustomerMasterDataInput): CustomerMasterDataInput {
  return {
    ...input,
    customerCode: input.customerCode.trim().toUpperCase(),
    customerName: input.customerName.trim(),
    channelCode: input.channelCode.trim().toUpperCase(),
    priceListCode: input.priceListCode.trim().toUpperCase(),
    discountGroup: input.discountGroup.trim(),
    creditLimit: finiteNumber(input.creditLimit),
    paymentTerms: input.paymentTerms.trim().toUpperCase(),
    contactName: input.contactName.trim(),
    phone: input.phone.trim(),
    email: input.email.trim().toLowerCase(),
    taxCode: input.taxCode.trim().toUpperCase(),
    address: input.address.trim()
  };
}

function validateSupplierInput(input: SupplierMasterDataInput) {
  const missing = [
    ["supplier code", input.supplierCode],
    ["supplier name", input.supplierName]
  ].filter(([, value]) => !String(value).trim());

  if (missing.length > 0) {
    throw new Error(`Missing required fields: ${missing.map(([label]) => label).join(", ")}`);
  }
  if (input.leadTimeDays < 0 || input.moq < 0 || input.qualityScore < 0 || input.deliveryScore < 0) {
    throw new Error("Supplier metrics cannot be negative");
  }
}

function validateCustomerInput(input: CustomerMasterDataInput) {
  const missing = [
    ["customer code", input.customerCode],
    ["customer name", input.customerName]
  ].filter(([, value]) => !String(value).trim());

  if (missing.length > 0) {
    throw new Error(`Missing required fields: ${missing.map(([label]) => label).join(", ")}`);
  }
  if (input.creditLimit < 0) {
    throw new Error("Customer credit limit cannot be negative");
  }
}

function ensureUniqueSupplier(input: SupplierMasterDataInput, currentId = "") {
  if (localSuppliers.some((item) => item.id !== currentId && item.supplierCode === input.supplierCode)) {
    throw new Error("Supplier code already exists");
  }
}

function ensureUniqueCustomer(input: CustomerMasterDataInput, currentId = "") {
  if (localCustomers.some((item) => item.id !== currentId && item.customerCode === input.customerCode)) {
    throw new Error("Customer code already exists");
  }
}

function filterSuppliers(items: SupplierMasterDataItem[], query: SupplierMasterDataQuery) {
  const search = query.search?.trim().toLowerCase();
  return sortSuppliers(
    items.filter((item) => {
      if (query.status && item.status !== query.status) {
        return false;
      }
      if (query.supplierGroup && item.supplierGroup !== query.supplierGroup) {
        return false;
      }
      if (!search) {
        return true;
      }

      return [
        item.supplierCode,
        item.supplierName,
        supplierGroupLabel(item.supplierGroup),
        item.contactName ?? "",
        item.phone ?? "",
        item.email ?? "",
        item.taxCode ?? "",
        item.address ?? "",
        item.paymentTerms ?? ""
      ].some((value) => value.toLowerCase().includes(search));
    })
  );
}

function filterCustomers(items: CustomerMasterDataItem[], query: CustomerMasterDataQuery) {
  const search = query.search?.trim().toLowerCase();
  return sortCustomers(
    items.filter((item) => {
      if (query.status && item.status !== query.status) {
        return false;
      }
      if (query.customerType && item.customerType !== query.customerType) {
        return false;
      }
      if (!search) {
        return true;
      }

      return [
        item.customerCode,
        item.customerName,
        customerTypeLabel(item.customerType),
        item.channelCode ?? "",
        item.priceListCode ?? "",
        item.discountGroup ?? "",
        item.paymentTerms ?? "",
        item.contactName ?? "",
        item.phone ?? "",
        item.email ?? "",
        item.taxCode ?? "",
        item.address ?? ""
      ].some((value) => value.toLowerCase().includes(search));
    })
  );
}

function sortSuppliers(items: SupplierMasterDataItem[]) {
  return [...items].sort((left, right) => {
    const statusDelta = partyStatusRank(left.status) - partyStatusRank(right.status);
    if (statusDelta !== 0) {
      return statusDelta;
    }
    if (left.supplierGroup !== right.supplierGroup) {
      return left.supplierGroup.localeCompare(right.supplierGroup);
    }

    return left.supplierCode.localeCompare(right.supplierCode);
  });
}

function sortCustomers(items: CustomerMasterDataItem[]) {
  return [...items].sort((left, right) => {
    const statusDelta = partyStatusRank(left.status) - partyStatusRank(right.status);
    if (statusDelta !== 0) {
      return statusDelta;
    }
    if (left.customerType !== right.customerType) {
      return left.customerType.localeCompare(right.customerType);
    }

    return left.customerCode.localeCompare(right.customerCode);
  });
}

function partyStatusRank(status: SupplierStatus | CustomerStatus) {
  switch (status) {
    case "active":
      return 0;
    case "draft":
      return 1;
    case "inactive":
      return 2;
    case "blacklisted":
    case "blocked":
      return 3;
    default:
      return 4;
  }
}

function finiteNumber(value: number) {
  return Number.isFinite(value) ? value : 0;
}

function cloneSuppliers(items: SupplierMasterDataItem[]) {
  return items.map((item) => ({ ...item }));
}

function cloneCustomers(items: CustomerMasterDataItem[]) {
  return items.map((item) => ({ ...item }));
}
