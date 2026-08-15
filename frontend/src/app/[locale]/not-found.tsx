import { getTranslations } from "next-intl/server";

import { Link } from "@/i18n/navigation";

export default async function NotFoundPage() {
  const t = await getTranslations("NotFound");

  return (
    <main className="grid min-h-screen place-items-center bg-canvas px-5 text-ink">
      <div className="max-w-lg text-center">
        <p className="text-sm font-semibold tracking-[0.16em] text-accent uppercase">
          404
        </p>
        <h1 className="mt-4 text-4xl font-semibold tracking-[-0.04em]">
          {t("title")}
        </h1>
        <p className="mt-4 text-lg leading-8 text-muted">{t("description")}</p>
        <Link
          href="/"
          className="mt-8 inline-flex min-h-11 items-center rounded-full bg-ink px-6 py-3 text-sm font-semibold text-canvas outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas"
        >
          {t("home")}
        </Link>
      </div>
    </main>
  );
}
