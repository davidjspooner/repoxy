
# Component: Header Bar

The fixed top-level bar visible on all screens.

## Responsibilities

- Host the **breadcrumb bar** so navigation is always visible.
- Provide **access to user-related actions** (settings, account, future logout).
- Never scrolls out of view.

## Layout

- Left-aligned:
  - Breadcrumbs showing the current path (Repository Types → Repository Browser → File Details).
- Right-aligned:
  - Username (MVP: `admin`).
  - Small user/avatar icon or button that opens the **user menu**.

## User Menu Contents (MVP)

- **Settings…** — enabled, opens Settings Dialog.
- **Account…** — visible but **disabled/greyed**, post-MVP.
- **Logout** — visible but **disabled/greyed**, post-MVP.

The visual disabled state communicates future capabilities while making it clear they cannot be used yet.

## Behaviour

- Stays fixed as user scrolls content in panels.
- Works the same on desktop and mobile.
- Responsiveness:
  - Breadcrumbs truncate/collapse on narrow widths but remain interactive.
