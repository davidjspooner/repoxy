# Component: ListPanel

Shared wrapper used by list-centric panels (types, repo instances, items, versions) to present selectable entries with consistent controls.

## Purpose

- Centralize the **list/tile rendering logic** so every panel offers the same experience.
- Ensure each entry always renders with a **bold title** and **italic detail** text beneath, regardless of mode.
- Embed the `ListPanelToolbar` for mode switching and filtering so individual panels stay lean.

## Behaviour

- Accepts an array of items shaped as `{ id, title, detail }`, plus optional `selectedId`, `onSelect`, `initialMode`, and `emptyMessage`.
- Maintains internal state for:
  - **Display mode** (`list` or `tiles`).
  - **Filter text** (case-insensitive search across title and detail).
- Applies client-side filtering; when no items match, a neutral empty-state message is shown.
- **List mode**:
  - Vertical stack of Paper elements filling the width.
  - Selected item shows stronger outline/border.
- **Tile mode**:
  - Wrap layout with fixed-width tiles to compare entries at a glance.
  - Selected tile uses thicker primary border.
- Clicking an entry fires `onSelect(item)`; ListPanel does not mutate external state beyond that callback.
- Empty-state message is customizable via prop and used both when zero items exist and when filters remove every entry.

## States

- **Default**: initial mode (list by default) with empty filter; magnifier icon shown.
- **Filtered**: filter text updates the rendered list immediately; toolbar displays the clear (×) button.
- **Selected**: active item highlighted visually in all modes.
- **Empty**: centered copy explains that no entries are available (due to data or filter).
- Loading/error handling remains the responsibility of the parent panel (ListPanel only renders the toolbar and item list).
