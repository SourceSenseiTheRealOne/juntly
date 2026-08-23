import { describe, expect, it } from "vitest";

import en from "../../messages/en.json";
import es from "../../messages/es.json";
import ptPT from "../../messages/pt-PT.json";

function flattenMessages(
  value: Record<string, unknown>,
  prefix = "",
): Record<string, string> {
  return Object.entries(value).reduce<Record<string, string>>(
    (messages, [key, nestedValue]) => {
      const path = prefix ? `${prefix}.${key}` : key;

      if (typeof nestedValue === "string") {
        messages[path] = nestedValue;
        return messages;
      }

      if (
        typeof nestedValue === "object" &&
        nestedValue !== null &&
        !Array.isArray(nestedValue)
      ) {
        Object.assign(
          messages,
          flattenMessages(nestedValue as Record<string, unknown>, path),
        );
      }

      return messages;
    },
    {},
  );
}

const locales = {
  "pt-PT": flattenMessages(ptPT),
  en: flattenMessages(en),
  es: flattenMessages(es),
};

describe("locale messages", () => {
  it("uses the Portuguese message keys in every supported locale", () => {
    const canonicalKeys = Object.keys(locales["pt-PT"]).sort();

    expect(Object.keys(locales.en).sort()).toEqual(canonicalKeys);
    expect(Object.keys(locales.es).sort()).toEqual(canonicalKeys);
  });

  it("contains only nonblank translated strings", () => {
    for (const messages of Object.values(locales)) {
      for (const message of Object.values(messages)) {
        expect(message.trim()).not.toBe("");
      }
    }
  });

  it("defines the complete account capability copy in every locale", () => {
    const capabilityKeys = [
      "Account.capabilities.customerDescription",
      "Account.capabilities.customerLabel",
      "Account.capabilities.description",
      "Account.capabilities.disabled",
      "Account.capabilities.enabled",
      "Account.capabilities.loadError",
      "Account.capabilities.loading",
      "Account.capabilities.providerDescription",
      "Account.capabilities.providerLabel",
      "Account.capabilities.retry",
      "Account.capabilities.saving",
      "Account.capabilities.title",
    ];

    for (const messages of Object.values(locales)) {
      for (const key of capabilityKeys) {
        expect(messages[key]).toBeTypeOf("string");
        expect(messages[key].trim()).not.toBe("");
      }
    }
  });
});
