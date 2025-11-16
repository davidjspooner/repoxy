
# Component: Concertina Shell

The **core layout/flow controller** for panels.

## Responsibilities

- Manage a **stack of panels**.
- Render **one or two panels** at a time depending on viewport width.
- Provide **breadcrumbs** for navigation.
- Host panel containers and the draggable divider.
- Display the **Connection Status modal overlay** when instructed (e.g. on live-update disconnects).

## Structure

Within the area below the header and above the footer:

1. **Breadcrumb Bar** (top area of the shell).
2. **Visible Panels Region**:
   - One or two **Panel Containers** representing the top of the stack.

## Panel Stack

- Maintains an ordered stack of panel descriptors (e.g. [Repository Types, Repository Browser, File Details]).
- The **rightmost** entry is always the **current panel**.
- Operations:
  - **Push(panel)** — navigate deeper.
  - **PopTo(panelIndex)** — navigate back to a specific ancestor panel (used by breadcrumbs).

## Desktop vs Mobile Rendering

- **Desktop/Wide View**
  - Render **two topmost panels**:
    - Left: `stack[-2]` (if exists).
    - Right: `stack[-1]`.
  - Insert the **Draggable Divider** between them.
  - Apply width constraints:
    - Each panel must have a minimum width; divider cannot collapse a panel completely.

- **Mobile/Narrow View**
  - Render only the **rightmost panel** (`stack[-1]`).
  - Divider is not shown.
  - Previous panels remain in the stack (for breadcrumbs), but are not rendered.

Threshold for determining “wide” vs “narrow” is implementation-specific.

## Breadcrumb Integration

- Breadcrumb items map directly to stack entries.
- Each breadcrumb entry, when clicked, calls **PopTo** with the index of that panel.
- The Concertina Shell is solely responsible for this behaviour; panels themselves don’t need to know about stack indexes.

## Scroll Handling

- Each visible panel is wrapped in a `PanelContainer` with its own vertical (and optional horizontal) scrolling.
- The Concertina Shell **does not** handle scrolling directly; it just arranges panel containers side by side.

## Live Updates & Deletion Handling

- Receives change notifications from the Live Update Subscription system.
- If a notification indicates that the **current panel’s subject no longer exists**:
  - Shell pops the affected panel(s) and re-renders using the remaining stack.
- If updates affect non-current panels (e.g. repository list changes while user is deep in a folder):
  - Shell determines whether to refresh those panels when they are next displayed.
- If the Live Update Subscription reports a lost connection (no offline mode):
  - Shell renders a **modal overlay** (blocking panel interaction) showing the countdown supplied by the subscription logic and a **Retry Now** button.
  - Modal also displays elapsed offline time plus a short troubleshooting hint received from the subscription service.
  - Automatic retries follow the exponential/backoff rules until capped at two minutes, after which they continue at two-minute intervals indefinitely (or until the user closes the tab).
  - The overlay is dismissed automatically once reconnection succeeds; users cannot interact with panels while it is visible.
