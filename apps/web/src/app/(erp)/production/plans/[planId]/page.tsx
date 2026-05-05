import { notFound } from "next/navigation";
import { ProductionPlanDetailPrototype } from "@/modules/production-planning/components/ProductionPlanDetailPrototype";
import { getBackendSession } from "@/shared/auth/serverSession";
import { hasPermission } from "@/shared/permissions/menu";

type ProductionPlanDetailPageProps = {
  params: Promise<{
    planId: string;
  }>;
};

export default async function ProductionPlanDetailPage({ params }: ProductionPlanDetailPageProps) {
  const { planId } = await params;
  const session = await getBackendSession();

  if (!session.isAuthenticated || !hasPermission(session.user, "production:view")) {
    notFound();
  }

  return <ProductionPlanDetailPrototype planId={planId} />;
}
