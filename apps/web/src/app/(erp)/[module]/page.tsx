import { notFound } from "next/navigation";
import { mockSession } from "@/shared/auth/mockSession";
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
  const item = menuItems.find((candidate) => candidate.href === `/${module}`);

  if (!item || !canAccessMenuItem(mockSession.user, item)) {
    notFound();
  }

  return <ModulePlaceholder item={item} />;
}
