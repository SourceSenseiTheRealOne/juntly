# Marketplace UI Redesign — Reference Mapping

**Status:** Approved visual direction from SourceSensei on 2026-08-25

## Intent

Translate the supplied browse-screen reference into Juntly's own visual system across public and protected marketplace surfaces. The redesign is presentation-only: locale structure, same-origin BFF calls, authorization, privacy boundaries, contact reveal behavior, and OpenAPI contracts remain unchanged.

## Evidence and mapping

The supplied `freelance-match-browse.png` reference is a 1440×1214 marketplace browse composition with a high-key white/cool-gray surface family and a dark plum-charcoal emphasis family. Juntly adopts the layout principles, not reference branding, content, imagery, or assets.

| Reference cue               | Juntly decision                                                                                                                                                                            |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Strong marketplace top bar  | Shared responsive header with Juntly brand, discovery entry point, account/provider routes, and Clerk account controls.                                                                    |
| Bright browse canvas        | Off-white canvas, cool mist control surfaces, white cards, hairline blue-gray borders, dark plum ink.                                                                                      |
| Search/filter-led discovery | Search field becomes a prominent, full-width toolbar with filter-context chips. Existing fixed-query behavior remains unchanged.                                                           |
| Editorial service cards     | Listing cards use a media-safe decorative category tile, compact metadata, provider/locality information, and a clear detail CTA. No object reference or private media data is introduced. |
| Rounded controls            | Buttons, fields, and chips share rounded-rectangle geometry, clear selected/disabled states, 44px minimum targets, and keyboard focus rings.                                               |
| Dense but calm browsing     | Wider desktop grid; single-column mobile collapse; content-first hierarchy; no required animation.                                                                                         |

## Scope

1. Add semantic tokens and reusable CSS component classes.
2. Add shared locale-aware navigation across the app.
3. Restyle home, discovery, public detail, account, provider, listings, contact-channel, and moderation page containers plus their repeated controls/cards.
4. Preserve plain public data boundaries and all existing accessible landmarks/labels.

## Explicit non-goals

- No Figma/reference logo, copy, images, or assets.
- No API, database, auth, contact reveal, moderation, or persistence changes.
- No faux listing media or sample marketplace data.
- No additional filters whose server contract does not already exist.
- No motion beyond existing reduced-motion-safe microinteraction patterns.

## Acceptance

- Every route has a shared, keyboard-accessible marketplace navigation surface.
- Public discovery and details remain contact-free before explicit authenticated reveal.
- Existing localized strings are retained; new navigation labels are localized in pt-PT, English, and Spanish.
- Desktop and narrow mobile layouts have no horizontal overflow and retain 44px controls/focus visibility.
- Full frontend quality gate passes and runtime proof checks current routes without source/security boundary changes.
