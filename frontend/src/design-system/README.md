# UNSOLERO frontend design system

The design system is dependency-free beyond React, Tailwind CSS, React Router, and the project's existing Lucide icon set.

## Foundations

- Tokens live in `tokens.css`: semantic colors, typography, spacing, radius, shadows, motion, breakpoints, and container widths.
- The 320px layout is the base. Named breakpoints cover 375, 390, 430, 768, 1024, 1280, 1440, and 1920px.
- Radius and shadow are intentionally restrained. Whitespace and typography provide most hierarchy.
- Motion uses short opacity, color, and position transitions and respects `prefers-reduced-motion`.

## Components

Generic primitives live under `components/ui` and are exported from its `index.ts`. Product composition lives under `components/product`; shared navigation and brand structure live under `components/layout`.

Form components own label, hint, error, required, and disabled presentation while forwarding native attributes and refs. Business validation remains in feature schemas and hooks.

Modal and drawer components use the native `dialog` element for focus containment and Escape behavior. Tabs implement roving keyboard selection. Toasts require `ToastProvider`, already mounted by the application provider tree.

## Temporary showcase

`/design-system` renders the complete inventory for responsive and interaction review. Product examples are explicitly fictional seed data. Remove the route when a dedicated documentation environment replaces it.
