
# Component: Breadcrumb Bar

Displays the **current navigation path** across panels.

## Responsibilities

- Provide a clear, always-visible indication of:
  - Current repository type.
  - Current path within that type (e.g. selected repository, folder).
  - Current file (if applicable).
- Enable navigation **back up** the panel stack.

## Placement

- Rendered inside the **Header Bar** so navigation is always visible even when panels scroll.

## Behaviour

- Each breadcrumb segment corresponds to **one stack entry** managed by the Concertina Shell.
- Clicking a segment:
  - Requests the Concertina Shell to **PopTo** that segment’s panel.
- Visual representation:
  - Text label per level (e.g. `Container / my-repo / images / v1 / manifest.json`).
  - Styling for the last segment to indicate “current location”.

## Responsiveness

- On narrow screens:
  - Breadcrumbs may:
    - Truncate intermediate segments.
    - Collapse into a shorter form with an overflow indicator (e.g. `Container / … / v1 / manifest.json`).
- Full breadcrumb state should remain available via tooltip or expansion.
