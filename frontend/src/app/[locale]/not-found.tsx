import { getTranslations } from "next-intl/server";

import { Link } from "@/i18n/navigation";

export default async function NotFoundPage() {
  const t = await getTranslations("NotFound");

  return (
    <main className="grid min-h-screen place-items-center bg-canvas px-page text-ink">
      <div className="max-w-lg text-center">
        <p className="text-sm font-semibold tracking-label text-accent uppercase">
          404
        </p>
        <h1 className="mt-4 text-4xl font-semibold tracking-brand">
          {t("title")}
        </h1>
        <p className="mt-4 text-lg leading-8 text-muted">{t("description")}</p>
        <Link
          href="/"
          className="mt-8 inline-flex min-h-touch items-center rounded-pill bg-accent px-6 py-3 text-sm font-semibold text-inverse transition-colors outline-none hover:bg-accent-hover focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas"
        >
          {t("home")}
        </Link>
      </div>
    </main>
  );
}
