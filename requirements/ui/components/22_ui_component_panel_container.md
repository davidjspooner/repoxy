
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
- Uses a **lilac / pastel purple background** (#f3e5f5) so it visually separates the concertina shell from the panel content. The container stretches to fill the height allocated by the shell but may not expand beyond that height; any overflow is handled via scrollbars.
- Internal scrolling:
  - Vertical scroll when content height exceeds container height.
  - Horizontal scroll when content width exceeds container width.
- Panels themselves are not responsible for scrollbars; they just render content.
