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
    <div className="market-page flex min-h-screen flex-col overflow-hidden">
      <main id="content" className="flex-1">
        <section className="relative isolate">
          <div
            className="absolute inset-0 -z-10 overflow-hidden"
            aria-hidden="true"
          >
            <div className="absolute -top-36 right-[-9rem] h-96 w-96 rounded-full bg-control blur-3xl" />
            <div className="absolute bottom-[-12rem] left-[-10rem] h-80 w-80 rounded-full bg-accent-soft blur-3xl" />
          </div>

          <div className="market-container grid min-h-[68vh] items-center gap-12 py-16 sm:py-20 lg:grid-cols-[1.2fr_0.8fr] lg:py-24">
            <div className="max-w-3xl">
              <p className="mb-5 text-xs font-bold tracking-[0.16em] text-accent uppercase">
                {tagline}
              </p>
              <h1 className="text-5xl leading-[0.98] font-bold tracking-[-0.06em] text-balance text-ink sm:text-6xl lg:text-7xl">
                {heading}
              </h1>
              <p className="mt-7 max-w-2xl text-lg leading-8 text-pretty text-muted sm:text-xl">
                {description}
              </p>
              <div className="mt-10 flex flex-col items-start gap-5 sm:flex-row sm:items-center">
                <a href="#vision" className="market-button">
                  {visionLinkLabel}
                  <span className="ml-2" aria-hidden="true">
                    ↓
                  </span>
                </a>
                <span className="inline-flex min-h-11 items-center rounded-xl border border-line bg-surface px-5 text-sm font-medium text-muted shadow-sm">
                  <span
                    className="mr-3 h-2.5 w-2.5 rounded-full bg-accent"
                    aria-hidden="true"
                  />
                  {statusLabel}
                </span>
                <HealthStatusIndicator labels={healthStatusLabels} />
              </div>
            </div>

            <div className="market-panel relative mx-auto hidden w-full max-w-md overflow-hidden p-5 lg:block">
              <div className="grid min-h-96 grid-cols-[0.8fr_1.2fr] gap-4 rounded-2xl bg-control p-4">
                <div className="rounded-xl bg-ink p-5 text-surface">
                  <p className="text-xs font-bold tracking-[0.14em] text-accent-soft uppercase">
                    Juntly
                  </p>
                  <p className="mt-20 text-3xl leading-none font-bold tracking-[-0.06em]">
                    {showcaseTitle}
                  </p>
                </div>
                <div className="space-y-3">
                  <div className="h-28 rounded-xl bg-surface" />
                  <div className="h-20 rounded-xl border border-line bg-surface p-4">
                    <div className="h-2 w-20 rounded-full bg-control-strong" />
                    <div className="mt-3 h-3 w-full rounded-full bg-control" />
                    <div className="mt-2 h-3 w-3/4 rounded-full bg-control" />
                  </div>
                  <div className="h-20 rounded-xl bg-accent-soft" />
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="vision" className="border-y border-line bg-surface">
          <div className="market-container grid gap-7 py-16 sm:py-20 lg:grid-cols-[0.8fr_1.2fr]">
            <p className="text-sm font-semibold tracking-[0.16em] text-accent uppercase">
              Juntly
            </p>
            <div>
              <h2 className="text-3xl font-semibold tracking-[-0.035em] text-balance sm:text-4xl">
                {visionTitle}
              </h2>
              <p className="mt-5 max-w-2xl text-lg leading-8 text-muted">
                {visionDescription}
              </p>
            </div>
          </div>
        </section>
      </main>

      <footer className="bg-ink text-surface">
        <div className="market-container flex flex-col gap-3 py-8 text-sm sm:flex-row sm:items-center sm:justify-between">
          <strong className="text-lg tracking-[-0.03em]">Juntly.</strong>
          <p className="text-surface/70">{footerLabel}</p>
        </div>
      </footer>
    </div>
  );
}
