
# Component: Concertina Shell

The **core layout/flow controller** for panels.  
Canonical navigation sequences and slot assignments are summarized in `../22_ui_panel_flow.md`; this component doc focuses on mechanics.

## Responsibilities

- Manage a **stack of panels**.
- Render **one or two panels** at a time depending on viewport width.
- Provide **breadcrumbs** for navigation.
- Host `ScrollableViewPort` instances and the draggable divider.
- Display the **Connection Status modal overlay** when instructed (e.g. on live-update disconnects).

## Structure

Within the area below the header and above the bottom edge of the viewport:

1. **Breadcrumb Bar** (top area of the shell).
2. **Visible Panels Region**:
   - One or two **ScrollableViewPorts** representing the top of the stack.

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

- Each visible panel is wrapped in a `ScrollableViewPort` that fills its allotted slot, keeps a white/paper background, and automatically adds horizontal/vertical scrollbars if its lone child exceeds the view.
- The Concertina Shell **does not** handle scrolling directly; it just arranges `ScrollableViewPort`s side by side.
- Shell container flexes to fill the space between the header and the bottom of the viewport with a flat paper background and no additional border treatments; the visual separation is now handled entirely by panel spacing/content.

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
