"use client";

import { useTranslations } from "next-intl";
import { useEffect } from "react";

type ErrorPageProps = {
  error: Error & { digest?: string };
  reset: () => void;
};

export default function ErrorPage({ error, reset }: ErrorPageProps) {
  const t = useTranslations("Errors");

  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <main className="grid min-h-screen place-items-center bg-canvas px-page text-ink">
      <div className="max-w-lg text-center">
        <p className="text-sm font-semibold tracking-label text-accent uppercase">
          Juntly
        </p>
        <h1 className="mt-4 text-4xl font-semibold tracking-brand">
          {t("title")}
        </h1>
        <p className="mt-4 text-lg leading-8 text-muted">{t("description")}</p>
        <button
          type="button"
          onClick={reset}
          className="mt-8 min-h-touch rounded-pill bg-accent px-6 py-3 text-sm font-semibold text-inverse transition-colors outline-none hover:bg-accent-hover focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas"
        >
          {t("retry")}
        </button>
      </div>
    </main>
  );
}
