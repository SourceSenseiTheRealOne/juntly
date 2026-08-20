import type { HealthStatusLabels } from "./health-status-indicator";
import { HealthStatusIndicator } from "./health-status-indicator";

type LandingShellProps = {
  eyebrow: string;
  tagline: string;
  heading: string;
  description: string;
  statusLabel: string;
  healthStatusLabels: HealthStatusLabels;
  visionLinkLabel: string;
  visionTitle: string;
  visionDescription: string;
  footerLabel: string;
};

export function LandingShell({
  eyebrow,
  tagline,
  heading,
  description,
  statusLabel,
  healthStatusLabels,
  visionLinkLabel,
  visionTitle,
  visionDescription,
  footerLabel,
}: LandingShellProps) {
  return (
    <div className="flex min-h-screen flex-col overflow-hidden bg-canvas text-ink">
      <header className="relative z-10 border-b border-line/80 bg-canvas/90 backdrop-blur-sm">
        <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-5 py-5 sm:px-8 lg:px-10">
          <a
            href="#content"
            className="rounded-sm text-2xl font-bold tracking-[-0.04em] text-ink transition-colors outline-none hover:text-accent focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas"
            aria-label="Juntly"
          >
            Juntly<span className="text-accent">.</span>
          </a>
          <p className="hidden text-sm font-medium text-muted sm:block">
            {eyebrow}
          </p>
        </div>
      </header>

      <main id="content" className="flex-1">
        <section className="relative isolate">
          <div className="absolute inset-0 -z-10 opacity-70" aria-hidden="true">
            <div className="absolute -top-36 right-[-9rem] h-96 w-96 rounded-full bg-accent-soft blur-3xl" />
            <div className="absolute bottom-[-12rem] left-[-10rem] h-80 w-80 rounded-full bg-earth-soft blur-3xl" />
          </div>

          <div className="mx-auto grid min-h-[70vh] w-full max-w-6xl items-center gap-14 px-5 py-20 sm:px-8 sm:py-28 lg:grid-cols-[1.35fr_0.65fr] lg:px-10 lg:py-32">
            <div className="max-w-3xl">
              <p className="mb-5 text-sm font-semibold tracking-[0.16em] text-accent uppercase">
                {tagline}
              </p>
              <h1 className="text-5xl leading-[0.98] font-semibold tracking-[-0.055em] text-balance text-ink sm:text-6xl lg:text-7xl">
                {heading}
              </h1>
              <p className="mt-7 max-w-2xl text-lg leading-8 text-pretty text-muted sm:text-xl">
                {description}
              </p>
              <div className="mt-10 flex flex-col items-start gap-5 sm:flex-row sm:items-center">
                <a
                  href="#vision"
                  className="inline-flex min-h-11 items-center justify-center rounded-full bg-ink px-6 py-3 text-sm font-semibold text-canvas transition-transform outline-none hover:-translate-y-0.5 focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas motion-reduce:transform-none"
                >
                  {visionLinkLabel}
                  <span className="ml-2" aria-hidden="true">
                    ↓
                  </span>
                </a>
                <span className="inline-flex min-h-11 items-center rounded-full border border-line bg-surface px-5 text-sm font-medium text-muted shadow-sm">
                  <span
                    className="mr-3 h-2.5 w-2.5 rounded-full bg-accent"
                    aria-hidden="true"
                  />
                  {statusLabel}
                </span>
                <HealthStatusIndicator labels={healthStatusLabels} />
              </div>
            </div>

            <div className="relative mx-auto hidden aspect-square w-full max-w-sm lg:block">
              <div className="absolute inset-0 rounded-full border border-line bg-surface/70 shadow-[0_32px_90px_rgba(53,45,35,0.10)]" />
              <div className="absolute inset-[16%] rounded-full border border-accent/25 bg-accent-soft/70" />
              <div className="absolute inset-[34%] grid place-items-center rounded-full bg-accent text-5xl font-bold tracking-[-0.08em] text-white shadow-xl">
                J
              </div>
              <span className="absolute top-[12%] right-[8%] h-5 w-5 rounded-full bg-earth" />
              <span className="absolute bottom-[18%] left-[6%] h-3 w-3 rounded-full bg-accent" />
            </div>
          </div>
        </section>

        <section id="vision" className="border-y border-line bg-surface">
          <div className="mx-auto grid w-full max-w-6xl gap-7 px-5 py-16 sm:px-8 sm:py-20 lg:grid-cols-[0.8fr_1.2fr] lg:px-10">
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

      <footer className="bg-ink text-canvas">
        <div className="mx-auto flex w-full max-w-6xl flex-col gap-3 px-5 py-8 text-sm sm:flex-row sm:items-center sm:justify-between sm:px-8 lg:px-10">
          <strong className="text-lg tracking-[-0.03em]">Juntly.</strong>
          <p className="text-canvas/70">{footerLabel}</p>
        </div>
      </footer>
    </div>
  );
}
