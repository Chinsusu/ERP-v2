import { notFound } from "next/navigation";
import { PurchaseRequestDetailPrototype } from "@/modules/purchase/components/PurchaseRequestDetailPrototype";
import { getBackendSession } from "@/shared/auth/serverSession";
import { hasPermission } from "@/shared/permissions/menu";

type PurchaseRequestDetailPageProps = {
  params: Promise<{
    requestId: string;
  }>;
};

export default async function PurchaseRequestDetailPage({ params }: PurchaseRequestDetailPageProps) {
  const { requestId } = await params;
  const session = await getBackendSession();

  if (!session.isAuthenticated || !hasPermission(session.user, "purchase:view")) {
    notFound();
  }

  return <PurchaseRequestDetailPrototype requestId={requestId} />;
}
