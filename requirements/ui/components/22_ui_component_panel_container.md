
# Component: Panel Container

A generic wrapper that hosts an individual panel’s content.

## Responsibilities

- Provide a **scrollable area** for panel content.
- Handle both **vertical** and optional **horizontal** scrolling.
- Constrain content to its allocated width/height.

## Behaviour

- Receives a fixed rectangle from the Concertina Shell:
  - Width determined by shell layout and draggable divider.
  - Height determined by available space between the header (including breadcrumbs) and the bottom of the viewport.
- Uses the theme's **paper** background with a consistent **2 px border** so every panel reads as a defined card, matching the header surface.
- Internal scrolling:
  - Vertical scroll when content height exceeds container height.
  - Horizontal scroll when content width exceeds container width.
- Panels themselves are not responsible for scrollbars; they just render content.
