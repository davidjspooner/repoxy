
# Panel Catalogue & Navigation

This document describes each **panel** (screen/major view) and how they connect.  
Panel stacking behaviour and transitions are handled by the **Concertina Shell** (see `components/22_ui_component_concertina_shell.md`) and **not** by individual panels.

---

## Panel List (MVP)

1. **Repository Types Panel** (`panels/24_ui_panel_repository_types.md`)
2. **Repository Browser Panel** (`panels/24_ui_panel_repository_browser.md`)
3. **File Details Panel** (`panels/24_ui_panel_file_details.md`)
4. **Settings Dialog Panel** (`panels/24_ui_panel_settings_dialog.md`)
5. **Root Panel** (`panels/24_ui_panel_root.md`) — conceptual root; in practice this is the same as Repository Types.

---

## High-Level Flow

### Entry

- The application starts with the **Repository Types Panel** as the visible panel.
- There is no “back” navigation beyond this; it is the root.

### From Repository Types → Repository Browser

- User selects a **Repository Type tile** (e.g. Docker, Terraform).
- Action:
  - Concertina Shell:
    - Pushes the **Repository Browser** panel onto the stack.
    - Makes it the right-hand panel.
  - New panel is scoped to the selected repository type.

### From Repository Browser → File Details

- User selects a **file row** in the **File List Table**.
- Action:
  - Concertina Shell:
    - Pushes the **File Details** panel as the right-hand panel.
  - Left-hand panel remains the Repository Browser for context.

### Backwards Navigation

- Achieved via **breadcrumbs** in the Concertina Shell, not within the panels themselves.
- When a breadcrumb representing a parent level is clicked:
  - All panels to its right are popped from the stack.
  - Example:
    - User on File Details → click the breadcrumb for the folder → File Details is removed; Repository Browser becomes rightmost.

### Settings Dialog

- Triggered from the **Header user menu**.
- Appears as a **modal dialog** overlaying the current panel(s).
- Does **not** participate in the concertina panel stack (no breadcrumb entry).
- Dims content behind it.
- Dismissed via close button or clicking outside (exact interaction can be tuned later).

---

## Panel States

Each panel supports the following states:

- **Loading**
  - Data not yet retrieved.
  - Show loading indication inside the panel (spinner, skeleton, etc.).

- **Empty**
  - No data available for the panel’s scope.
  - Example: a folder has no files; a repository type has no repositories.
  - Display a neutral empty state message.

- **Error**
  - Data retrieval failed.
  - Show inline error message (inside the panel).
  - Additionally trigger a **red toast** describing the error.

- **Populated**
  - Primary, normal state with data visible.

The specific treatment of these states is described in each panel’s markdown.
