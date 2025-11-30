
# Repoxy UI Overview

This document describes the **MVP user interface** for the Repoxy application, based entirely on the interview so far.

Repoxy is a web UI for **browsing repository data cached by the backend** (e.g. Container registries, Terraform/OpenTofu providers and modules, PyPI indexes, Debian APT repositories).  

The webapp used react / nod and matrial UI

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
   - Fixed at the top in a light “paper” colour.
   - Left: **Breadcrumb bar** for navigation (always visible).
   - Right: Username (`admin` in MVP) and a user menu trigger containing Settings (enabled) plus disabled Account/Logout entries.

2. **Concertina Shell (Panel Region)**  
   - Immediately below the header.
   - Contains the **panel containers** (one or two visible at a time) and the draggable divider on wide screens.
   - See `22_ui_panel_flow.md` for the canonical sequence of panels, their slot assignments, and stack transitions.

### Desktop vs Mobile

- **Desktop / Wide Viewports**
  - Header and breadcrumbs are always visible.
  - Concertina region usually shows **two visible panels** side by side:
    - Left: the **Folder Browser** panel when browsing within a repository type.
    - Right: either the **File List** panel or, when a file is selected, the **File Details** panel. (When File Details is visible the File List shifts to the left slot.)
  - Repository Types occupies a single panel; when fewer than two panels exist the shell leaves the unused slot empty.
  - A **draggable vertical divider** between the two visible panels allows resizing but never fully hides either panel.

- **Mobile / Narrow Viewports**
  - Header and breadcrumbs still visible.
 - Concertina region shows only **one panel at a time**:
    - The **current (rightmost) panel** is shown.
    - Users move back using the breadcrumb trail (or a built-in back affordance within the concertina shell).
  - Internally the shell still maintains a stack of panels, but only one is rendered.

### Panel Visual Style

- Every panel renders its content on a **pale pastel cyan background** so the information blocks feel lightweight while staying consistent across panel types.
- The `ScrollableViewPort` that wraps each panel uses a **paper/white background**, stretches to fill (but not exceed) the slot allotted by the concertina shell, and automatically provides scrollbars whenever its single child overflows.
- The Concertina shell itself now sits on the same **paper background with no decorative border**, letting the draggable divider and panel padding define the structure.

---

## Core Navigation Flow (MVP)

At a conceptual level, the navigation stack is:

1. **Repository Types Panel**  
   - Tiles representing repository categories that have configured repositories (derived entirely from the backend’s repo catalogue such as `repoxy/conf/repoxy.yaml`, e.g. Container, Terraform, OpenTofu).  
   - Repo types that have **no configured instances** are hidden; the UI is a direct reflection of configuration.

2. **Folder Browser Panel** (per repository type)  
   - Hosts the **Folder Tree View** whose **top-level nodes are repository names** within the chosen type (for the example config: `dockerhub`, `github`, `terraform-hashicorp`, `opentofu-registry`).  
   - Each repository instance has three **fixed levels** beneath it before files appear:
     1. `host`
     2. `group`
     3. `name`
   - Only the `name` folders contain files; higher levels contain folders only.

3. **File List Panel**  
   - Shows the **File List Table** for the currently selected folder (`name` level).
   - Updates automatically when the user selects a folder in the Folder Browser Panel.

4. **File Details Panel**  
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

If the long-running connection is lost:
- There is **no offline mode**. The UI cannot be used until connectivity returns.
- The Concertina Shell displays a **modal dialog overlay** (blocking interaction) describing the connection loss.
- The dialog shows a countdown to the next automatic retry, following an exponential sequence until it hits a **2-minute ceiling**; after reaching that ceiling it continues retrying every 2 minutes **indefinitely** until the web app reconnects or the browser tab is closed.
- The modal also displays the **elapsed offline time** and a short troubleshooting hint (e.g. “Check homelab backend or network cabling”) so prolonged outages still feel accounted for.
- The countdown updates each second; a prominent **“Retry Now”** button lets the user trigger an immediate attempt (which also resets the backoff sequence).
- Once a retry succeeds the modal is dismissed automatically and the UI resumes normal behaviour.

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
- `components/22_ui_component_toast.md`
- `components/22_ui_component_toast_queue.md`
- `components/22_ui_component_settings_dialog.md`
- `components/22_ui_component_draggable_divider.md`
- `components/22_ui_component_live_update_subscription.md`
- `components/22_ui_component_panel_title.md`
- `28_ui_post_mvp.md` — post-MVP UI roadmap and out-of-scope features.

These markdown files should be sufficient for another LLM or engineer to implement the UI without needing to repeat this interview.
