"use client";

import { useEffect, useState } from "react";

export type HealthStatusLabels = {
  checking: string;
  available: string;
  unavailable: string;
};

type HealthStatusIndicatorProps = {
  labels: HealthStatusLabels;
};

type HealthState = "checking" | "available" | "unavailable";

export function HealthStatusIndicator({ labels }: HealthStatusIndicatorProps) {
  const [state, setState] = useState<HealthState>("checking");

  useEffect(() => {
    let ignore = false;

    async function checkHealth() {
      try {
        const response = await fetch("/api/v1/health", {
          cache: "no-store",
          headers: {
            Accept: "application/json",
          },
        });
        const body: unknown = await response.json();

        if (!ignore && response.ok && isAvailableHealth(body)) {
          setState("available");
          return;
        }
      } catch {
        // Keep transport details private; show only localized generic state.
      }

      if (!ignore) {
        setState("unavailable");
      }
    }

    void checkHealth();

    return () => {
      ignore = true;
    };
  }, []);

  return (
    <span
      className="inline-flex min-h-11 items-center rounded-full border border-line bg-surface px-5 text-sm font-medium text-muted shadow-sm"
      role="status"
      aria-live="polite"
    >
      <span
        className={`mr-3 h-2.5 w-2.5 rounded-full ${
          state === "available"
            ? "bg-accent"
            : state === "unavailable"
              ? "bg-earth"
              : "bg-muted"
        }`}
        aria-hidden="true"
      />
      {labels[state]}
    </span>
  );
}

function isAvailableHealth(value: unknown): boolean {
  if (!value || typeof value !== "object") {
    return false;
  }

  const health = value as Record<string, unknown>;

  return health.status === "ok" && health.service === "juntly-api";
}
