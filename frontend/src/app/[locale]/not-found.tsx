import { getTranslations } from "next-intl/server";

import { Link } from "@/i18n/navigation";

export default async function NotFoundPage() {
  const t = await getTranslations("NotFound");

  return (
    <main className="market-page grid place-items-center px-5">
      <div className="market-panel max-w-lg p-8 text-center">
        <p className="text-sm font-semibold tracking-[0.16em] text-accent uppercase">
          404
        </p>
        <h1 className="mt-4 text-4xl font-semibold tracking-[-0.04em]">
          {t("title")}
        </h1>
        <p className="mt-4 text-lg leading-8 text-muted">{t("description")}</p>
        <Link href="/" className="market-button mt-8">
          {t("home")}
        </Link>
      </div>
    </main>
  );
}
