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
    <main className="market-page grid place-items-center px-5">
      <div className="market-panel max-w-lg p-8 text-center">
        <p className="text-sm font-semibold tracking-[0.16em] text-accent uppercase">
          Vila
        </p>
        <h1 className="mt-4 text-4xl font-semibold tracking-[-0.04em]">
          {t("title")}
        </h1>
        <p className="mt-4 text-lg leading-8 text-muted">{t("description")}</p>
        <button type="button" onClick={reset} className="market-button mt-8">
          {t("retry")}
        </button>
      </div>
    </main>
  );
}
