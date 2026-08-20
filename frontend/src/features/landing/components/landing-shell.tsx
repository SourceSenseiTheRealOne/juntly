type LandingShellProps = {
  eyebrow: string;
  tagline: string;
  heading: string;
  description: string;
  statusLabel: string;
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
  visionLinkLabel,
  visionTitle,
  visionDescription,
  footerLabel,
}: LandingShellProps) {
  return (
    <div className="flex min-h-screen flex-col overflow-hidden bg-canvas text-ink">
      <header className="relative z-10 border-b border-line/80 bg-canvas/90 backdrop-blur-sm">
        <div className="mx-auto flex w-full max-w-content items-center justify-between px-page py-header">
          <a
            href="#content"
            className="inline-flex min-h-touch items-center rounded-control text-2xl font-bold tracking-brand text-ink transition-colors outline-none hover:text-accent focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas"
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
            <div className="absolute top-[var(--offset-ambient-primary-block)] right-[var(--offset-ambient-primary-inline)] size-[var(--size-ambient-primary)] rounded-pill bg-accent-soft blur-3xl" />
            <div className="absolute bottom-[var(--offset-ambient-secondary-block)] left-[var(--offset-ambient-secondary-inline)] size-[var(--size-ambient-secondary)] rounded-pill bg-earth-soft blur-3xl" />
          </div>

          <div className="mx-auto grid min-h-hero w-full max-w-content items-center gap-content-gap px-page py-section lg:grid-cols-[var(--layout-hero-columns)]">
            <div className="max-w-copy">
              <p className="mb-5 text-sm font-semibold tracking-label text-accent uppercase">
                {tagline}
              </p>
              <h1 className="text-display leading-display font-semibold tracking-display text-balance text-ink">
                {heading}
              </h1>
              <p className="mt-7 max-w-lead text-lead leading-8 text-pretty text-muted">
                {description}
              </p>
              <div className="mt-10 flex flex-col items-start gap-5 sm:flex-row sm:items-center">
                <a
                  href="#vision"
                  className="inline-flex min-h-touch items-center justify-center rounded-pill bg-accent px-6 py-3 text-sm font-semibold text-inverse transition-[color,background-color,transform] outline-none hover:-translate-y-0.5 hover:bg-accent-hover focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas motion-reduce:transform-none"
                >
                  {visionLinkLabel}
                  <span className="ml-2" aria-hidden="true">
                    ↓
                  </span>
                </a>
                <span className="inline-flex min-h-touch items-center rounded-pill border border-line bg-surface px-5 text-sm font-medium text-muted shadow-surface">
                  <span
                    className="mr-3 size-status-dot rounded-pill bg-success"
                    aria-hidden="true"
                  />
                  {statusLabel}
                </span>
              </div>
            </div>

            <div className="relative mx-auto hidden aspect-square w-full max-w-orbit lg:block">
              <div className="absolute inset-0 rounded-pill border border-line bg-surface/70 shadow-elevated" />
              <div className="absolute inset-[var(--layout-orbit-ring-inset)] rounded-pill border border-accent/25 bg-accent-soft/70" />
              <div className="absolute inset-[var(--layout-orbit-core-inset)] grid place-items-center rounded-pill bg-accent text-5xl font-bold tracking-display text-inverse shadow-elevated">
                J
              </div>
              <span className="absolute top-[var(--layout-orbit-dot-top)] right-[var(--layout-orbit-dot-right)] size-orbit-dot-lg rounded-pill bg-earth" />
              <span className="absolute bottom-[var(--layout-orbit-dot-bottom)] left-[var(--layout-orbit-dot-left)] size-orbit-dot-sm rounded-pill bg-accent" />
            </div>
          </div>
        </section>

        <section id="vision" className="border-y border-line bg-surface">
          <div className="mx-auto grid w-full max-w-content gap-content-gap px-page py-section lg:grid-cols-[0.8fr_1.2fr]">
            <p className="text-sm font-semibold tracking-label text-accent uppercase">
              Juntly
            </p>
            <div>
              <h2 className="text-section-title font-semibold tracking-heading text-balance">
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
        <div className="mx-auto flex w-full max-w-content flex-col gap-3 px-page py-footer text-sm sm:flex-row sm:items-center sm:justify-between">
          <strong className="text-lg tracking-heading">Juntly.</strong>
          <p className="text-canvas/70">{footerLabel}</p>
        </div>
      </footer>
    </div>
  );
}
