"use client";

import { Show, UserButton } from "@clerk/nextjs";

type AuthNavigationProps = {
  signInLabel: string;
  signInUrl: string;
  signUpLabel: string;
  signUpUrl: string;
};

export function AuthNavigation({
  signInLabel,
  signInUrl,
  signUpLabel,
  signUpUrl,
}: AuthNavigationProps) {
  return (
    <nav aria-label="Account">
      <Show when="signed-out">
        <div className="flex items-center gap-2">
          <a
            href={signInUrl}
            className="inline-flex min-h-11 items-center justify-center rounded-full px-4 text-sm font-semibold text-ink transition-colors outline-none hover:bg-surface focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas"
          >
            {signInLabel}
          </a>
          <a
            href={signUpUrl}
            className="inline-flex min-h-11 items-center justify-center rounded-full bg-ink px-4 text-sm font-semibold text-canvas transition-transform outline-none hover:-translate-y-0.5 focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas motion-reduce:transform-none"
          >
            {signUpLabel}
          </a>
        </div>
      </Show>
      <Show when="signed-in">
        <UserButton />
      </Show>
    </nav>
  );
}
