
# Repoxy UI Overview

This document describes the **MVP user interface** for the Repoxy application, based entirely on the interview so far.

Repoxy is a web UI for **browsing repository data cached by the backend** (e.g. Docker registries, Terraform/OpenTofu providers and modules, PyPI indexes, Debian APT repositories).  

For the **MVP**, the UI is **read-only navigation**:

- No creation, deletion, or modification of repositories or files.
- No login/logout flow (single implicit admin user).
- No download or clean-up actions.

Those actions are explicitly **post-MVP** and are documented in `28_ui_post_mvp.md`.

---

## Target Users

- Technical users (SRE / platform engineers / homelab users).
- Comfortable with dense, information-heavy layouts.
- Expect real-time-ish updates when the underlying cache changes.

---

## Technology & Form Factor

- Implemented as a **web application**, intended to be built with **React**.
- **Desktop-first**, but responsive down to tablet and mobile.
- Single-page application behaviour: navigation is handled client-side by the concertina/panel system.

---

## Top-Level Layout

From top to bottom, the screen is structured as:

1. **Header Bar**  
   - Fixed at the top.
   - Left: Application name (e.g. “Repoxy”).
   - Right: Username (`admin` in MVP) and a user menu trigger.

2. **Concertina Shell (Breadcrumb + Panel Region)**  
   - Immediately below the header.
   - Contains:
     - A **breadcrumb bar** for navigation (always visible).
     - A **concertina panel region** that shows **one or two panels at a time**.

3. **Footer Summary Bar**  
   - Fixed at the bottom.
   - Shows a **very brief summary** of the currently selected item (repository, folder, or file) along with its **last updated timestamp** reported by the backend.
   - If the UI loses its live-update connection, the footer suffixes the summary with ` (disconnected)` until connectivity is restored.
   - Distinct from the detailed information shown in the right-hand detail panel.

### Desktop vs Mobile

- **Desktop / Wide Viewports**
  - Header, breadcrumbs, and footer are always visible.
  - Concertina region shows up to **two visible panels** side by side:
    - Left: “context” panel (e.g. list of repo types, or folder tree + file list).
    - Right: “detail” panel (e.g. file details).
  - A **draggable vertical divider** between the two visible panels allows resizing but never fully hides either panel.

- **Mobile / Narrow Viewports**
  - Header, breadcrumbs, and footer still visible.
  - Concertina region shows only **one panel at a time**:
    - The **current (rightmost) panel** is shown.
    - Users move back using the breadcrumb trail (or a built-in back affordance within the concertina shell).
  - Internally the shell still maintains a stack of panels, but only one is rendered.

---

## Core Navigation Flow (MVP)

At a conceptual level, the navigation stack is:

1. **Repository Types Panel**  
   - Tiles representing repository categories that have configured repositories (derived entirely from the backend’s repo catalogue such as `repoxy/conf/repoxy.yaml`, e.g. Docker, Terraform, OpenTofu).  
   - Repo types that have **no configured instances** are hidden; the UI is a direct reflection of configuration.

2. **Repository Browser Panel** (per repository type)  
   - Left side: **Folder Tree View** whose **top-level nodes are repository names** within the chosen type (for the example config: `dockerhub`, `github`, `terraform-hashicorp`, `opentofu-registry`).  
   - The depth and naming of folders under each repo instance are **type specific**, but files only appear in **leaf folders** (no folder contains both files and child folders).
   - Right side: **File List Table** (files in the selected folder).

3. **File Details Panel**  
   - Detailed view for a single selected file.

The root of the app is effectively the **Repository Types** panel. There is no panel “above” it in the stack.

---

## Dynamic Behaviour & Live Updates

The UI assumes a **long-running connection** to the backend (e.g. WebSocket, SSE, or similar) where appropriate:

- When the backend detects changes (files added, removed, or modified) that affect a currently displayed path or folder:
  - The corresponding panels are **refreshed automatically**.
  - This can trigger:
    - Updated file lists.
    - Updated folder trees.
    - Updated file details.

- If the **currently selected object disappears** (e.g. file deleted by another client):
  - The UI automatically navigates **up one or more levels** until it lands on a valid object.
  - The user is not left on a “broken” panel; instead they see the parent folder, repository, or repo type list as appropriate.

- The backend provides some notion of **versioned change notifications** scoped to a particular path.  
  The UI:
  - Treats notifications as authoritative for knowing when to refresh a panel.
  - Does **not** define the exact versioning scheme; that’s a backend concern.

If an error occurs during refresh (e.g. a 500 from the backend), a **toast notification** is displayed (see below).

---

## Toast Notifications

The UI uses a **toast system** for transient user feedback:

- **Placement:** bottom of the viewport.
- **Colour:**
  - **Pastel red** background for errors / failures.
  - **Pastel green** background for success / “all good” messages.
- **Text:** dark text (black or near-black) for maximum legibility.
- **Behaviour:**
  - Toasts auto-dismiss after a few seconds.
  - Multiple toasts are stacked vertically in a **queue**.
  - If an error occurs but a subsequent success/“recovered” state is reached **before** the error toast is displayed:
    - The UI **may skip** the now-irrelevant error toast.
  - This is primarily to avoid spamming users with flickering error/success messages during rapid retries.

Toasts are **informational only** in the MVP; there are no action buttons beyond a possible “dismiss” affordance.

---

## Settings & Personalisation

The MVP includes a **Settings dialog** (accessible from the header user menu) that:

- Appears as a modal dialog, dimming the content behind it.
- Allows the user to control:
  - **Spacing mode:** compact vs comfortable.
  - **Theme mode:** light, dark, and possibly system.
- Additional settings may be added later.

At MVP:

- There is a single implicit user (`admin`).
- Account and logout menu entries are visible but **disabled/greyed** to telegraph future capabilities.

Details of the settings dialog are in `panels/24_ui_panel_settings_dialog.md` and the corresponding component spec.

---

## Documents in This UI Spec

- `20_ui_overview.md` — this file.
- `20_ui_panels.md` — panel catalogue and navigation relationships.
- `panels/24_ui_panel_root.md` — the root state (effectively repository types).
- `panels/24_ui_panel_repository_types.md` — panel showing repository type tiles.
- `panels/24_ui_panel_repository_browser.md` — panel for folder tree + file list per repo type.
- `panels/24_ui_panel_file_details.md` — panel for viewing file metadata.
- `panels/24_ui_panel_settings_dialog.md` — modal settings panel.
- `components/22_ui_component_header_bar.md`
- `components/22_ui_component_concertina_shell.md`
- `components/22_ui_component_breadcrumbs.md`
- `components/22_ui_component_panel_container.md`
- `components/22_ui_component_repo_type_tile.md`
- `components/22_ui_component_tile_grid.md`
- `components/22_ui_component_folder_tree_view.md`
- `components/22_ui_component_file_list_table.md`
- `components/22_ui_component_footer_summary_bar.md`
- `components/22_ui_component_toast.md`
- `components/22_ui_component_toast_queue.md`
- `components/22_ui_component_settings_dialog.md`
- `components/22_ui_component_draggable_divider.md`
- `components/22_ui_component_live_update_subscription.md`
- `28_ui_post_mvp.md` — post-MVP UI roadmap and out-of-scope features.

These markdown files should be sufficient for another LLM or engineer to implement the UI without needing to repeat this interview.
