import Link from "next/link";
import { t } from "@/shared/i18n";

export default function ERPNotFound() {
  return (
    <main className="erp-page erp-page--centered">
      <section className="erp-card erp-form-card">
        <h1 className="erp-form-title">{t("auth.errors.forbidden")}</h1>
        <Link className="erp-button erp-button--primary erp-button--full" href="/dashboard">
          {t("navigation.items.dashboard")}
        </Link>
      </section>
    </main>
  );
}
