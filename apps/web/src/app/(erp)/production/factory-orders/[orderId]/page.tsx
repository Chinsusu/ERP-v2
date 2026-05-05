import { notFound } from "next/navigation";
import { SubcontractOrderDetailPrototype } from "@/modules/subcontract/components/SubcontractOrderDetailPrototype";
import { getBackendSession } from "@/shared/auth/serverSession";
import { hasPermission } from "@/shared/permissions/menu";

type FactoryOrderDetailPageProps = {
  params: Promise<{
    orderId: string;
  }>;
};

export default async function FactoryOrderDetailPage({ params }: FactoryOrderDetailPageProps) {
  const { orderId } = await params;
  const session = await getBackendSession();

  if (!session.isAuthenticated || !hasPermission(session.user, "production:view")) {
    notFound();
  }

  return <SubcontractOrderDetailPrototype orderId={orderId} />;
}
