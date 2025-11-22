
# Item & Version Browser Panel

Displays the **item list and version list** for a specific repository instance before handing off to the file list and file details panels.

## Purpose

- Allow users to:
  - Browse the list of **items/containers/packages** exposed by the chosen repository (examples: `library/nginx`, `hashicorp/aws`).
  - Inspect the **available versions/tags/releases** for a selected item.
  - Select a version to continue to the File List panel (which then leads to File Details).

## Parent / Children

- **Parent:** Repository Instances Panel (repository selection).
- **Children:** Version List Panel → File List Panel → File Details Panel.

Stack placement, slot usage, and transitions between the Item/Version panel, File List, and File Details panels are detailed in `../22_ui_panel_flow.md`.

## Layout

The panel occupies one concertina slot at a time, but conceptually forms a **two-step list experience**:

1. **Item List**
   - Rendered as a vertically stacked list (not tiles) so users can scan long names quickly.
   - Each row highlights the item label (typically `<group>/<name>`) with optional subtext for the upstream host/path.
   - Selecting an item focuses the list and loads the matching Version List while keeping the Item List visible on desktop.
2. **Version List**
   - Populated after an item is chosen.
   - Shows available versions/tags/releases for the current item, including additional metadata (publish dates, status, etc.).
   - Selecting a version pushes the File List panel onto the stack.

## Independent Scrolling

- The panel lives inside a `ScrollableViewPort`, so whichever list is visible can scroll independently.
- On desktop, the Item List and Version List appear side by side with their own scroll regions; on mobile only the active list is visible at a time.

## Interaction Details

- **Selecting an Item**
  - Marks the item as active.
  - Refreshes the Version List with versions specific to that item.
- **Selecting a Version**
  - Pushes the File List panel to the right (desktop) or replaces the current view (mobile).
  - The Item List remains in the stack for breadcrumb/back navigation.
- **Sorting/Filtering**
  - Minimum viable product does not require complex filtering, but the UI should support alphabetical sorting and optional search for both items and versions when long lists are expected.

## States

- **Loading**
  - Show skeleton placeholders for both lists while fetching items/versions.
- **Empty**
  - If a repository has no items, display a central empty state message.
  - If an item has no versions, Version List shows a friendly “No versions found” message.
- **Error**
  - Both lists should be able to display inline errors if data fetch fails, and also surface a toast.
- **Populated**
  - Normal list presentation with the selected entries highlighted.

## Live Updates

- When new items or versions appear due to backend updates:
  - The Item List and Version List refresh in place.
- If the currently selected item or version disappears:
  - The panel automatically navigates up one level (clearing the version or item selection) per the general lifecycle rules.
