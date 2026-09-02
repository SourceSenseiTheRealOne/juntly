import { getTranslations } from "next-intl/server";

export default async function LoadingPage() {
  const t = await getTranslations("Loading");

  return (
    <main
      className="market-page grid place-items-center"
      aria-live="polite"
      aria-busy="true"
    >
      <div className="market-panel flex items-center gap-3 p-5 text-sm font-medium text-muted">
        <span
          className="h-1.5 w-12 animate-pulse rounded-full bg-control-strong motion-reduce:animate-none"
          aria-hidden="true"
        />
        {t("label")}
      </div>
    </main>
  );
}
