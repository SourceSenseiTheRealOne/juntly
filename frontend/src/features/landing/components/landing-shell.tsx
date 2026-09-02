import Image from "next/image";

import type { HealthStatusLabels } from "./health-status-indicator";
import { HealthStatusIndicator } from "./health-status-indicator";

type LandingShellProps = {
  tagline: string;
  heading: string;
  description: string;
  statusLabel: string;
  healthStatusLabels: HealthStatusLabels;
  visionLinkLabel: string;
  visionTitle: string;
  visionDescription: string;
  showcaseTitle: string;
  footerLabel: string;
};

export function LandingShell({
  tagline,
  heading,
  description,
  statusLabel,
  healthStatusLabels,
  visionLinkLabel,
  visionTitle,
  visionDescription,
  showcaseTitle,
  footerLabel,
}: LandingShellProps) {
  return (
    <div className="market-page flex min-h-[100dvh] flex-col overflow-hidden">
      <main id="content" className="flex-1">
        <section className="border-b border-line bg-surface">
          <div className="market-container grid min-h-[calc(100dvh-4.5rem)] items-center gap-10 py-12 sm:py-16 lg:grid-cols-[0.92fr_1.08fr] lg:gap-16 lg:py-20">
            <div className="max-w-2xl lg:py-8">
              <p className="market-kicker mb-5">{tagline}</p>
              <h1 className="text-5xl leading-[0.98] font-bold tracking-[-0.065em] text-balance text-ink sm:text-6xl lg:text-7xl">
                {heading}
              </h1>
              <p className="mt-6 max-w-xl text-lg leading-8 text-pretty text-muted sm:text-xl">
                {description}
              </p>
              <div className="mt-8 flex flex-col items-start gap-4 sm:flex-row sm:items-center">
                <a href="#vision" className="market-button">
                  {visionLinkLabel}
                </a>
                <span className="inline-flex min-h-11 items-center rounded-xl border border-line bg-control px-4 text-sm font-medium text-muted">
                  {statusLabel}
                </span>
              </div>
              <div className="mt-4">
                <HealthStatusIndicator labels={healthStatusLabels} />
              </div>
            </div>

            <figure className="relative mx-auto w-full max-w-2xl lg:ml-auto">
              <div className="relative aspect-[4/3] overflow-hidden rounded-[1.75rem] border border-line bg-control shadow-[var(--shadow-float)] sm:aspect-[16/11]">
                <Image
                  alt=""
                  className="object-cover"
                  fill
                  priority
                  sizes="(max-width: 1024px) 100vw, 54vw"
                  src="/images/local-craft.jpg"
                />
              </div>
              <figcaption className="mt-4 border-b border-line pb-4 text-sm text-muted">
                <span className="font-semibold text-ink">{showcaseTitle}</span>
              </figcaption>
            </figure>
          </div>
        </section>

        <section id="vision" className="scroll-mt-24 bg-canvas">
          <div className="market-container grid gap-8 py-16 sm:py-24 lg:grid-cols-[0.7fr_1.3fr] lg:gap-16">
            <p className="max-w-xs text-sm leading-6 font-semibold text-muted">
              Juntly
            </p>
            <div className="max-w-3xl">
              <h2 className="text-3xl font-semibold tracking-[-0.045em] text-balance sm:text-5xl">
                {visionTitle}
              </h2>
              <p className="mt-6 max-w-2xl text-lg leading-8 text-muted">
                {visionDescription}
              </p>
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t border-line bg-surface text-ink">
        <div className="market-container flex flex-col gap-3 py-8 text-sm sm:flex-row sm:items-center sm:justify-between">
          <strong className="text-lg tracking-[-0.045em]">
            Juntly<span className="text-accent">.</span>
          </strong>
          <p className="text-muted">{footerLabel}</p>
        </div>
      </footer>
    </div>
  );
}
