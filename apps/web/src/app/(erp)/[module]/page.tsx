import { notFound } from "next/navigation";
import { AuditLogPrototype } from "@/modules/audit/components/AuditLogPrototype";
import { AvailableStockPrototype } from "@/modules/inventory/components/AvailableStockPrototype";
import { ProductMasterDataPrototype } from "@/modules/masterdata/components/ProductMasterDataPrototype";
import { ReturnReceivingPrototype } from "@/modules/returns/components/ReturnReceivingPrototype";
import { CarrierManifestPrototype } from "@/modules/shipping/components/CarrierManifestPrototype";
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
    return <ProductMasterDataPrototype />;
  }

  if (module === "warehouse") {
    return <WarehouseDailyBoard />;
  }

  if (module === "shipping") {
    return <CarrierManifestPrototype />;
  }

  if (module === "returns") {
    return <ReturnReceivingPrototype />;
  }

  if (module === "subcontract") {
    return <SubcontractOrderPrototype />;
  }

  if (module === "audit-log") {
    return <AuditLogPrototype />;
  }

  return <ModulePlaceholder item={item} user={session.user} />;
}
