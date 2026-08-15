import { getTranslations } from "next-intl/server";

export default async function LoadingPage() {
  const t = await getTranslations("Loading");

  return (
    <main
      className="grid min-h-screen place-items-center bg-canvas text-ink"
      aria-live="polite"
      aria-busy="true"
    >
      <div className="flex items-center gap-3 text-sm font-medium text-muted">
        <span
          className="h-3 w-3 animate-pulse rounded-full bg-accent motion-reduce:animate-none"
          aria-hidden="true"
        />
        {t("label")}
      </div>
    </main>
  );
}
