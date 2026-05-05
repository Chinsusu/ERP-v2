import { notFound } from "next/navigation";
import { PurchaseOrderDetailPrototype } from "@/modules/purchase/components/PurchaseOrderDetailPrototype";
import { getBackendSession } from "@/shared/auth/serverSession";
import { hasPermission } from "@/shared/permissions/menu";

type PurchaseOrderDetailPageProps = {
  params: Promise<{
    poId: string;
  }>;
};

export default async function PurchaseOrderDetailPage({ params }: PurchaseOrderDetailPageProps) {
  const { poId } = await params;
  const session = await getBackendSession();

  if (!session.isAuthenticated || !hasPermission(session.user, "purchase:view")) {
    notFound();
  }

  return <PurchaseOrderDetailPrototype poId={poId} />;
}
