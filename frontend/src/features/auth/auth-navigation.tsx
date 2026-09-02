"use client";

import { Show, UserButton } from "@clerk/nextjs";

type AuthNavigationProps = {
  navigationLabel: string;
  signInLabel: string;
  signInUrl: string;
  signUpLabel: string;
  signUpUrl: string;
};

export function AuthNavigation({
  navigationLabel,
  signInLabel,
  signInUrl,
  signUpLabel,
  signUpUrl,
}: AuthNavigationProps) {
  return (
    <nav aria-label={navigationLabel}>
      <Show when="signed-out">
        <div className="flex items-center gap-1.5 sm:gap-2">
          <a
            href={signInUrl}
            className="market-button-secondary hidden min-h-11 px-4 sm:inline-flex"
          >
            {signInLabel}
          </a>
          <a href={signUpUrl} className="market-button min-h-11 px-3 sm:px-4">
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
