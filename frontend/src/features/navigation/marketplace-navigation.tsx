import { AuthNavigation } from "@/features/auth/auth-navigation";

type MarketplaceNavigationProps = {
  accountLabel: string;
  accountUrl: string;
  discoverLabel: string;
  discoverUrl: string;
  accountNavigationLabel: string;
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
  navigationLabel,
  signInLabel,
  signInUrl,
  signUpLabel,
  signUpUrl,
}: MarketplaceNavigationProps) {
  const localeRoot = discoverUrl.replace(/\/discover$/, "");

  return (
    <header className="sticky top-0 z-20 border-b border-line/80 bg-surface/95 shadow-[0_1px_0_rgba(73,57,65,0.04)] backdrop-blur">
      <div className="mx-auto flex min-h-16 w-full max-w-7xl items-center gap-4 px-4 sm:px-6 lg:px-8">
        <a
          aria-label="Juntly"
          className="shrink-0 text-xl font-bold tracking-[-0.06em] text-ink transition-colors outline-none hover:text-accent focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4"
          href={localeRoot}
        >
          Juntly<span className="text-accent">.</span>
        </a>
        <nav
          aria-label={navigationLabel}
          className="flex min-w-0 flex-1 items-center gap-1"
        >
          <a
            className="inline-flex min-h-11 items-center rounded-xl bg-accent-soft px-3 text-sm font-semibold text-ink transition-colors outline-none hover:bg-accent-soft/70 focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2 sm:px-4"
            href={discoverUrl}
          >
            {discoverLabel}
          </a>
          <a
            className="hidden min-h-11 items-center rounded-xl px-3 text-sm font-semibold text-muted transition-colors outline-none hover:bg-canvas hover:text-ink focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2 sm:inline-flex"
            href={accountUrl}
          >
            {accountLabel}
          </a>
        </nav>
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
