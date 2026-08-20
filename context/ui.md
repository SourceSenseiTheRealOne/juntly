# UI Context

## Confirmed UI system

- Selective React Bits for reviewed visual/interactive treatments.
- shadcn/Radix for accessible forms, dialogs, menus, focus management, and keyboard behavior.
- All third-party/source-distributed components are wrapped behind Juntly-owned primitives and tokens.
- Bootstrap rule: do not install or copy unused UI components.

## Experience principles

Juntly must feel trustworthy, local, clear, and welcoming rather than technical or corporate. Mobile use and slow rural connections are primary constraints. The interface uses plain pt-PT language, short forms, clear confirmations, helpful empty/error states, and visible distinctions between individuals, verified professionals, businesses, organic results, and promotions.

The bootstrap uses neutral, high-contrast foundation tokens only. Final brand visual direction requires a later approved design slice; the shell must not freeze an unreviewed final identity.

## Design-token baseline

- `frontend/src/app/globals.css` is the frontend token authority. Raw color values belong only in its primitive palette; components consume semantic Tailwind utilities mapped to CSS custom properties.
- Semantic light/dark color tokens cover canvas, surfaces, primary/secondary/inverse text, borders, actions, accents, status colors, focus, and elevation shadows.
- Light mode is the deterministic fallback. `prefers-color-scheme` selects dark mode automatically, while `data-theme="light"` and `data-theme="dark"` provide explicit overrides without duplicating component styles.
- Fluid responsive CSS variables use `clamp()` for page gutters, section rhythm, content gaps, container widths, hero sizing, display/body type, touch targets, and decorative scale.
- Structural layout changes continue to use named Tailwind breakpoint variants because CSS custom properties cannot be used as ordinary media-query thresholds. Breakpoint-specific values must still resolve through shared variables rather than raw component literals.
- Minimum interactive targets are 44×44px through the shared touch-target token.
- Restrained radii/shadows and minimal animation use shared shape/elevation tokens.
- Motion never carries essential information and respects `prefers-reduced-motion`.

## Component architecture

- Primitive: Button, Link, Input, Textarea, Select, Checkbox, Dialog, Menu, Field, Badge.
- Composed: search controls, contact reveal, provider/listing cards, proposal comparison, booking timeline.
- Feature components consume typed feature services/hooks, never secrets, database code, or privileged upstream APIs.
- React Bits source is copied only per approved feature, retains attribution/license, and is reviewed as project code.

## Internationalization

Default locale is `pt-PT`; English is supported; Spanish routes/message structure are prepared. All user-facing component strings, metadata, dates, times, currencies, numbers, category names, notifications, and static content are locale-aware. User-generated descriptions are not auto-translated in the MVP.

## Accessibility and verification

Target WCAG 2.2 AA where practical: semantic landmarks/headings, skip/focus behavior, visible focus, keyboard operation, screen-reader labels, associated form errors, no color-only status, strong contrast, zoom/reflow, reduced motion, and descriptive loading/empty/error states. Verify real production output with automated checks plus keyboard and responsive browser review.
