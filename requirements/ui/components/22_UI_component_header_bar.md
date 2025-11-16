
# Component: Header Bar

The fixed top-level bar visible on all screens.

## Responsibilities

- Provide **global context** (app identity).
- Provide **access to user-related actions** (settings, account, future logout).
- Never scrolls out of view.

## Layout

- Left-aligned:
  - Application name, e.g. **Repoxy**.
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
  - On very narrow screens, text labels may shorten (e.g. using just an icon for the user menu).

