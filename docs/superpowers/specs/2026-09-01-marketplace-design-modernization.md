# Juntly Marketplace Design Modernization

**Status:** Approved for autonomous implementation by SourceSensei on 2026-09-01.

## Design read

Juntly is a trust-first local-services marketplace for mobile-first customers and providers in Portugal. The redesign uses a calm editorial marketplace language: off-white canvas, white and cool-mist surfaces, plum-charcoal text, one restrained berry accent, generous hierarchy, and practical responsive behavior.

Design dials: `DESIGN_VARIANCE 6`, `MOTION_INTENSITY 3`, `VISUAL_DENSITY 5`.

## Mode and boundaries

This is a preserve-mode product redesign. Keep every route, navigation label, form field, API call, state transition, authentication boundary, privacy rule, and localization contract unchanged. Change presentation, hierarchy, responsive layout, and accessible interaction feedback only.

## Visual system

- Use Geist as the existing product family with tighter display tracking and calmer body rhythm.
- Keep the approved off-white, cool-mist, white-card, plum-charcoal, and berry system, but increase contrast and reduce the pale-blue startup feel.
- Use a consistent 14-18px soft radius system: controls 14px, cards 18px, large panels 24px. Pills are reserved for compact status metadata.
- Replace generic card stacks with grouped surfaces, asymmetric grids, stronger whitespace, and clear section headings.
- Add system dark-mode tokens while preserving the same hierarchy and single accent.
- Use real craft photography on the landing page. Remove the fake product screenshot, blurred mesh decoration, decorative scroll arrow, and decorative status dot.

## Shared shell

- Keep the desktop navigation under 72px with one-line labels.
- On narrow screens, prioritize brand, discovery, and primary auth/account access without overflow.
- Introduce semantic page-header, toolbar, alert, empty-state, data-list, and form-section utilities.
- Use a subtle canvas texture based on CSS color layers, not gradients that imitate AI mesh art.

## Public marketplace

- Landing: split editorial hero with real local-craft photography, concise CTA, API availability as functional status, and a cleaner statement section.
- Discovery: prominent search toolbar, honest filter affordances, responsive two-column results, stronger provider/location/price hierarchy, clear promoted treatment, and composed empty/error states.
- Listing detail: content-first service summary, visible metadata, stronger provider action rail, and responsive stacking.

## Account and operations

- Account hub: clear overview, capability grouping, and scan-friendly destination links.
- Forms: consistent labels, section grouping, check controls, action rows, and feedback states.
- Messaging: inbox-like split layout on desktop and stacked conversation flow on mobile.
- Bookings, quotations, reviews, subscriptions, moderation, and administration: consistent metric groups, state chips, action hierarchy, readable responsive cards, and horizontal containment.

## Motion and accessibility

- Use CSS-only transform/opacity entrance and hover feedback where it communicates hierarchy or action.
- Content remains visible without JavaScript.
- Disable entrances, hover lifts, and smooth scrolling under `prefers-reduced-motion`.
- Maintain 44px targets, visible focus, strong form contrast, keyboard access, and zero unintended horizontal overflow.

## Verification

Run formatting, lint, component tests, localization tests, typecheck, dependency audit, and production build. Verify landing, discovery, account shell, and representative protected surfaces at desktop and true mobile widths, including focus, reduced motion, overflow, and console errors.
