import Image from "next/image";

type ContentBlock = { title: string; description: string };

type LandingShellProps = {
  tagline: string;
  heading: string;
  description: string;
  visionLinkLabel: string;
  discoverLinkLabel: string;
  discoverUrl: string;
  accountUrl: string;
  signUpUrl: string;
  visionTitle: string;
  visionDescription: string;
  showcaseTitle: string;
  howTitle: string;
  howDescription: string;
  discoverBlock: ContentBlock;
  compareBlock: ContentBlock;
  contactBlock: ContentBlock;
  customerBlock: ContentBlock & { action: string };
  providerBlock: ContentBlock & { action: string; imageAlt: string };
  trustTitle: string;
  trustDescription: string;
  privacyBlock: ContentBlock;
  localBlock: ContentBlock;
  reputationBlock: ContentBlock;
  closingTitle: string;
  closingDescription: string;
  closingAction: string;
  footerLabel: string;
};

export function LandingShell({
  tagline,
  heading,
  description,
  visionLinkLabel,
  discoverLinkLabel,
  discoverUrl,
  accountUrl,
  signUpUrl,
  visionTitle,
  visionDescription,
  showcaseTitle,
  howTitle,
  howDescription,
  discoverBlock,
  compareBlock,
  contactBlock,
  customerBlock,
  providerBlock,
  trustTitle,
  trustDescription,
  privacyBlock,
  localBlock,
  reputationBlock,
  closingTitle,
  closingDescription,
  closingAction,
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
              <div className="mt-8 flex flex-col items-start gap-3 sm:flex-row sm:items-center">
                <a href={discoverUrl} className="market-button">
                  {discoverLinkLabel}
                </a>
                <a href="#vision" className="market-button-secondary">
                  {visionLinkLabel}
                </a>
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
          <div className="market-container py-16 sm:py-24">
            <div className="market-panel grid overflow-hidden lg:grid-cols-[0.72fr_1.28fr]">
              <div className="flex min-h-64 flex-col justify-between bg-accent-soft p-7 sm:p-10 lg:min-h-96">
                <p
                  aria-label="Juntly vision"
                  className="text-5xl font-bold tracking-[-0.075em] text-ink sm:text-7xl"
                >
                  Juntly<span className="text-accent">.</span>
                </p>
                <span
                  aria-hidden="true"
                  className="mt-16 block h-2 w-24 rounded-full bg-accent"
                />
              </div>
              <div className="flex flex-col justify-center p-7 sm:p-10 lg:p-14">
                <h2 className="max-w-3xl text-3xl font-semibold tracking-[-0.045em] text-balance sm:text-5xl">
                  {visionTitle}
                </h2>
                <p className="mt-6 max-w-2xl text-lg leading-8 text-muted">
                  {visionDescription}
                </p>
              </div>
            </div>
          </div>
        </section>

        <section className="border-y border-line bg-surface">
          <div className="market-container py-16 sm:py-24">
            <div className="market-page-header">
              <h2 className="text-3xl font-semibold tracking-[-0.045em] sm:text-5xl">
                {howTitle}
              </h2>
              <p>{howDescription}</p>
            </div>
            <div className="mt-10 grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
              <article className="market-card flex min-h-72 flex-col justify-between bg-accent-soft p-6 sm:p-8 lg:row-span-2">
                <span className="market-chip w-fit border-accent/20 bg-surface text-accent">
                  {discoverLinkLabel}
                </span>
                <div className="mt-16 max-w-xl">
                  <h3 className="text-2xl font-semibold tracking-[-0.035em] sm:text-3xl">
                    {discoverBlock.title}
                  </h3>
                  <p className="mt-4 leading-7 text-muted">
                    {discoverBlock.description}
                  </p>
                </div>
              </article>
              <article className="market-card p-6 sm:p-8">
                <h3 className="text-xl font-semibold tracking-[-0.025em]">
                  {compareBlock.title}
                </h3>
                <p className="mt-3 leading-7 text-muted">
                  {compareBlock.description}
                </p>
              </article>
              <article className="market-card bg-control p-6 sm:p-8">
                <h3 className="text-xl font-semibold tracking-[-0.025em]">
                  {contactBlock.title}
                </h3>
                <p className="mt-3 leading-7 text-muted">
                  {contactBlock.description}
                </p>
              </article>
            </div>
          </div>
        </section>

        <section className="bg-canvas">
          <div className="market-container grid gap-5 py-16 sm:py-24 lg:grid-cols-2">
            <article className="market-panel flex flex-col items-start p-6 sm:p-8 lg:p-10">
              <h2 className="text-3xl font-semibold tracking-[-0.045em]">
                {customerBlock.title}
              </h2>
              <p className="mt-5 max-w-xl leading-7 text-muted">
                {customerBlock.description}
              </p>
              <a className="market-button mt-8" href={discoverUrl}>
                {customerBlock.action}
              </a>
            </article>
            <article className="market-panel grid overflow-hidden bg-ink text-surface sm:grid-cols-[1.05fr_0.95fr]">
              <div className="flex flex-col items-start p-6 sm:p-8 lg:p-10">
                <h2 className="text-3xl font-semibold tracking-[-0.045em]">
                  {providerBlock.title}
                </h2>
                <p className="mt-5 max-w-xl leading-7 text-surface/75">
                  {providerBlock.description}
                </p>
                <a
                  className="market-button-secondary mt-8 border-surface/30 bg-transparent text-surface hover:bg-surface/10"
                  href={accountUrl}
                >
                  {providerBlock.action}
                </a>
              </div>
              <div className="relative min-h-64 sm:min-h-full">
                <Image
                  alt={providerBlock.imageAlt}
                  className="object-cover"
                  fill
                  sizes="(max-width: 640px) 100vw, 38vw"
                  src="/images/local-provider.jpg"
                />
              </div>
            </article>
          </div>
        </section>

        <section className="border-y border-line bg-surface">
          <div className="market-container grid gap-10 py-16 sm:py-24 lg:grid-cols-[0.8fr_1.2fr] lg:gap-16">
            <div className="market-page-header">
              <h2 className="text-3xl font-semibold tracking-[-0.045em] sm:text-5xl">
                {trustTitle}
              </h2>
              <p>{trustDescription}</p>
            </div>
            <div className="grid gap-3">
              {[privacyBlock, localBlock, reputationBlock].map((item) => (
                <article
                  className="border-b border-line py-5 first:pt-0 last:border-b-0"
                  key={item.title}
                >
                  <h3 className="text-lg font-semibold">{item.title}</h3>
                  <p className="mt-2 leading-7 text-muted">
                    {item.description}
                  </p>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className="bg-accent-soft">
          <div className="market-container flex flex-col items-start gap-8 py-16 sm:py-24 lg:flex-row lg:items-end lg:justify-between">
            <div className="max-w-3xl">
              <h2 className="text-4xl font-semibold tracking-[-0.055em] sm:text-6xl">
                {closingTitle}
              </h2>
              <p className="mt-5 max-w-2xl text-lg leading-8 text-muted">
                {closingDescription}
              </p>
            </div>
            <a className="market-button shrink-0" href={signUpUrl}>
              {closingAction}
            </a>
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
