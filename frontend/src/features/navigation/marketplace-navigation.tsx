import { AuthNavigation } from "@/features/auth/auth-navigation";

type MarketplaceNavigationProps = {
  accountLabel: string;
  accountUrl: string;
  discoverLabel: string;
  discoverUrl: string;
  accountNavigationLabel: string;
  createServiceLabel: string;
  createServiceUrl: string;
  navigationLabel: string;
  signInLabel: string;
  signInUrl: string;
  signUpLabel: string;
  signUpUrl: string;
};

export function MarketplaceNavigation({
  accountLabel,
  accountUrl,
  discoverLabel,
  discoverUrl,
  accountNavigationLabel,
  createServiceLabel,
  createServiceUrl,
  navigationLabel,
  signInLabel,
  signInUrl,
  signUpLabel,
  signUpUrl,
}: MarketplaceNavigationProps) {
  const localeRoot = discoverUrl.replace(/\/discover$/, "");

  return (
    <header className="sticky top-0 z-20 border-b border-line bg-surface/95 backdrop-blur-xl">
      <div className="mx-auto flex min-h-[4.5rem] w-full max-w-[80rem] items-center gap-2 px-3 sm:gap-5 sm:px-6 lg:px-8">
        <a
          aria-label="Juntly"
          className="shrink-0 rounded-lg text-xl font-bold tracking-[-0.065em] text-ink transition-colors outline-none hover:text-accent focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2 sm:text-2xl"
          href={localeRoot}
        >
          Juntly<span className="text-accent">.</span>
        </a>
        <nav
          aria-label={navigationLabel}
          className="flex min-w-0 flex-1 items-center gap-1 sm:gap-2"
        >
          <a
            className="hidden min-h-11 items-center rounded-xl px-3 text-sm font-semibold text-ink transition-colors outline-none hover:bg-control focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2 sm:inline-flex sm:px-4"
            href={discoverUrl}
          >
            {discoverLabel}
          </a>
          <a
            className="hidden min-h-11 items-center rounded-xl px-3 text-sm font-semibold text-muted transition-colors outline-none hover:bg-control hover:text-ink focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2 md:inline-flex"
            href={accountUrl}
          >
            {accountLabel}
          </a>
        </nav>
        <a
          className="market-button shrink-0 px-3 text-sm sm:px-4"
          href={createServiceUrl}
        >
          {createServiceLabel}
        </a>
        <AuthNavigation
          navigationLabel={accountNavigationLabel}
          signInLabel={signInLabel}
          signInUrl={signInUrl}
          signUpLabel={signUpLabel}
          signUpUrl={signUpUrl}
        />
      </div>
    </header>
  );
}
