import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HealthStatusIndicator } from "./health-status-indicator";

const labels = {
  checking: "A verificar a API.",
  available: "API disponível.",
  unavailable: "API temporariamente indisponível.",
};

describe("HealthStatusIndicator", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("calls the same-origin BFF and renders the available state", async () => {
    const fetchHealth = vi.fn(async () =>
      Response.json(
        {
          status: "ok",
          service: "juntly-api",
          version: "0.1.0",
          checkedAt: "2026-08-20T09:30:00Z",
          requestId: "req_ui_123",
        },
        {
          headers: {
            "X-Request-ID": "req_ui_123",
          },
        },
      ),
    );
    vi.stubGlobal("fetch", fetchHealth);

    render(<HealthStatusIndicator labels={labels} />);

    expect(screen.getByRole("status")).toHaveTextContent(labels.checking);
    await waitFor(() =>
      expect(screen.getByRole("status")).toHaveTextContent(labels.available),
    );
    expect(fetchHealth).toHaveBeenCalledWith("/api/v1/health", {
      cache: "no-store",
      headers: {
        Accept: "application/json",
      },
    });
  });

  it("renders the unavailable state without exposing transport details", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("connect ECONNREFUSED http://go-api:8080");
      }),
    );

    render(<HealthStatusIndicator labels={labels} />);

    await waitFor(() =>
      expect(screen.getByRole("status")).toHaveTextContent(labels.unavailable),
    );
    expect(screen.getByRole("status")).not.toHaveTextContent("go-api");
    expect(screen.getByRole("status")).not.toHaveTextContent("ECONNREFUSED");
  });
});
