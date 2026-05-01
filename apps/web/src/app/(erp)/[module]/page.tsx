import { notFound } from "next/navigation";
import { AuditLogPrototype } from "@/modules/audit/components/AuditLogPrototype";
import { FinanceReceivablesPrototype } from "@/modules/finance/components/FinanceReceivablesPrototype";
import { AvailableStockPrototype } from "@/modules/inventory/components/AvailableStockPrototype";
import { MasterDataPrototype } from "@/modules/masterdata/components/MasterDataPrototype";
import { PurchaseOrderPrototype } from "@/modules/purchase/components/PurchaseOrderPrototype";
import { InboundQCPrototype } from "@/modules/qc/components/InboundQCPrototype";
import { WarehouseReceivingPrototype } from "@/modules/receiving/components/WarehouseReceivingPrototype";
import { ReportingDashboard } from "@/modules/reporting/components/ReportingDashboard";
import { ReturnReceivingPrototype } from "@/modules/returns/components/ReturnReceivingPrototype";
import { SalesOrderPrototype } from "@/modules/sales/components/SalesOrderPrototype";
import { ShippingOperationsPrototype } from "@/modules/shipping/components/ShippingOperationsPrototype";
import { SubcontractOrderPrototype } from "@/modules/subcontract/components/SubcontractOrderPrototype";
import WarehouseDailyBoard from "@/modules/warehouse/components/WarehouseDailyBoard";
import { getMockSession } from "@/shared/auth/mockSession";
import { ModulePlaceholder } from "@/shared/layouts/ModulePlaceholder";
import { appMenuGroups, canAccessMenuItem } from "@/shared/permissions/menu";

type ERPModulePageProps = {
  params: Promise<{
    module: string;
  }>;
};

const menuItems = appMenuGroups.flatMap((group) => group.items);

export default async function ERPModulePage({ params }: ERPModulePageProps) {
  const { module } = await params;
  const session = await getMockSession();
  const item = menuItems.find((candidate) => candidate.href === `/${module}`);

  if (!session.isAuthenticated || !item || !canAccessMenuItem(session.user, item)) {
    notFound();
  }

  if (module === "inventory") {
    return <AvailableStockPrototype />;
  }

  if (module === "master-data") {
    return <MasterDataPrototype />;
  }

  if (module === "warehouse") {
    return <WarehouseDailyBoard />;
  }

  if (module === "receiving") {
    return <WarehouseReceivingPrototype />;
  }

  if (module === "purchase") {
    return <PurchaseOrderPrototype />;
  }

  if (module === "qc") {
    return <InboundQCPrototype />;
  }

  if (module === "shipping") {
    return <ShippingOperationsPrototype />;
  }

  if (module === "returns") {
    return <ReturnReceivingPrototype />;
  }

  if (module === "subcontract") {
    return <SubcontractOrderPrototype />;
  }

  if (module === "sales") {
    return <SalesOrderPrototype />;
  }

  if (module === "finance") {
    return <FinanceReceivablesPrototype />;
  }

  if (module === "audit-log") {
    return <AuditLogPrototype />;
  }

  if (module === "reports") {
    return <ReportingDashboard user={session.user} />;
  }

  return <ModulePlaceholder item={item} user={session.user} />;
}
