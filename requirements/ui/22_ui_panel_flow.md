# UI Panel Flow

This document centralizes how Repoxy panels flow from one to another and how each panel occupies the concertina shell in both mobile (single-column) and desktop (two-panel) states. It supplements the component docs by keeping navigation rules and layout roles in one place.

---

## Panel Roster

| Panel ID | Title in UI | Purpose |
| --- | --- | --- |
| `repository-types` | Repoxy | Root dashboard showing configured repository-type tiles. |
| `repo-instances` | `<Repo Type> Repositories` | Tile list of repositories belonging to the selected type; user must pick one to continue. |
| `artifact-list` | `<Repo Name> Items` | List of logical items/containers/packages (e.g., `hashicorp/aws`, `library/nginx`). |
| `version-list` | `<Item Name> Versions` | Lists available versions/tags/releases for the selected item. |
| `file-list` | `<Version> Files` | Tabular listing of files inside the currently selected version. |
| `file-details` | `<File Name>` | Detail view for a single file (metadata + usage). |

> **Notes**
> - Settings dialog, toast queue, and modal overlays are not part of the panel stack, so they are omitted here.
> - Panels later in the list depend on selections made in preceding panels; the stack never skips a prerequisite.

---

## Flow Summary

### Desktop / Wide (Two Visible Panels)

```
Empty Left Slot + Repository Types
          │ select type
          ▼
Repository Tiles (only visible panel)
          │ select repository name
          ▼
Item List (left) + Version List (right, prompts until item picked)
          │ select version
          ▼
Version List (left) + File List (right, prompts until file picked)
          │ select file
          ▼
File List (left) + File Details (right)
```

- When the viewport supports two columns, the Concertina Shell always renders the top two panels in the stack.
- **Initial state**: Repository Types is the only panel, occupying the right slot (left slot unused).
- **Selecting a type** pushes the Repo Instances (repository “name”) panel. Only the Repo Instances panel becomes visible; Repository Types stays in the stack for breadcrumb/back navigation.
- **Selecting a repository name** pushes the Item List and Version List pair. Repo Instances remains in the stack but scrolls offscreen; the visible pair becomes `Items | Versions`.
- **Selecting an item** refreshes the Version List content. Slot assignments remain `Items | Versions`.
- **Selecting a version** pushes the File List. With only two slots visible, the layout becomes `Versions | Files`.
- **Selecting a file** pushes File Details. The visible pair becomes `Files | File Details`.
- **Going back** via breadcrumbs pops panels from the right. Removing File Details redisplays the Folder Browser + File List pairing; clearing folder selection removes File List, leaving Folder Browser + empty right slot, and clearing the repo type pops back to Repository Types.
- The draggable divider is shown whenever both visible slots contain panels.

### Mobile / Narrow (Single Visible Panel)

```
Repository Types
      │ select type
      ▼
Repository List
      │ select repository name
      ▼
Item List
      │ select item
      ▼
Version List
      │ select version
      ▼
File List
      │ select file
      ▼
   File Details
```

- **Only one panel is rendered** at a time; the stack order matches desktop but visibility is restricted to the top entry.
- **Selecting a type** pushes the Repo Instances panel and makes it visible full-screen.
- **Selecting a repository name** pushes the Item List while showing it full-screen. Versions are prepared in the background.
- **Selecting an item** keeps the Item List visible or transitions to the Version List (implementation choice) but promotes Version List to the top of the stack so it becomes visible.
- **Selecting a version** promotes the File List panel.
- **Selecting a file** pushes File Details, replacing File List on screen. The underlying stack order is `[Repo Types, Repo Instances, Item List, Version List, File List, File Details]`, enabling backwards navigation through breadcrumbs.
- **Back navigation** pops panels in reverse order: File Details → File List → Version List → Item List → Repo Instances → Repository Types.
- No draggable divider is shown because there is never a second visible panel.
- Only the top (rightmost) panel in the stack is shown; earlier panels remain in the stack purely for breadcrumbs/back navigation.
- **Selecting a repo type** replaces Repository Types with Folder Browser; File List sits behind it in the stack ready to become visible.
- **Selecting a folder** keeps Folder Browser visible; File List only becomes visible after the user navigates forward (e.g., via breadcrumb or implicit transition) or selects a file? Wait ensure message: On mobile when folder selected, File List should be visible. so flows as diagram above. Continue patch accordingly. Need to adjust text describing transitions. Let's continue patch to include bullet list describing mobile transitions.

---

## Display Modes by Panel

| Panel | Mobile / Narrow | Desktop / Wide slot logic |
| --- | --- | --- |
| Repository Types | Only panel rendered at app start. | Occupies the right slot (or single slot) until Repo Instances is added. |
| Repo Instances | Only panel rendered immediately after a type is chosen. | Becomes the sole visible panel even on wide screens; Repository Types stays in the stack but is not shown. |
| Item List | Only panel rendered when it’s top-of-stack (repository selected, no item yet). | Occupies **left slot** while Version List sits on the right. Presented as a scrolling list, not tiles, to aid scanability. |
| Version List | Becomes visible after an item is selected; otherwise stays prepared off-screen. | Occupies **right slot** when Item List is visible; shifts to **left slot** when File List is also present. |
| File List | Only panel rendered when it’s the newest panel. | Occupies **right slot** when Version List is visible; shifts to **left slot** when File Details is also present. |
| File Details | Only panel rendered (mobile) whenever a file is selected. | Occupies **right slot** while File List remains on the left. |

Additional rules:
- The draggable divider is only shown when both left and right slots are populated.
- When the stack has fewer than two panels, the unused slot remains blank but keeps the shell layout stable.

---

## Panel-Specific Flow Details

### 1. Repository Types Panel
- **Entry**: root of the stack; always the first panel.
- **Actions**: selecting a repository type pushes the Repo Instances panel (repository names). Switching types replaces any downstream panels.
- **Exit**: selecting the same type again does nothing; selecting a different type replaces downstream panels entirely.
- **Empty / No data**: if no repo types exist, the panel explains that configuration is empty; no further panels are added.

### 2. Repo Instances Panel
- **Entry**: appears immediately after a type is chosen and takes over the visible area; the type panel becomes accessible only via breadcrumbs/back navigation.
- **Actions**: user selects a repository tile to push the Folder Browser / File List pair for that specific repository.
- **Exit**: selecting a different type removes this panel and anything after it; selecting a different repository replaces the browser panels with the new scope.
- **Empty / No data**: if a type has zero repositories, the panel shows an empty message and the flow stops here.

### 3. Item List Panel
- **Entry**: appears after choosing a repository name.
- **Actions**: displays each logical item/package/container (e.g., `library/nginx`, `hashicorp/aws`) as a vertically scrolling list (not tiles) so long names remain readable. Selecting an item pushes the Version List.
- **Exit**: changing repository or type removes this panel and any downstream panels.
- **Dependency**: requires a repository selection.

### 4. Version List Panel
- **Entry**: appears after an item is selected.
- **Actions**: shows versions/tags/releases for the item. Selecting a version pushes the File List.
- **Exit**: choosing a different item replaces this panel (and anything after it) with the new item’s versions.
- **Dependency**: requires an item selection.

### 5. File List Panel
- **Entry**: created whenever a version is selected.
- **Actions**:
  - Selecting a file row pushes File Details.
  - Switching versions refreshes the list and, if File Details was present, clears it until a new file is chosen.
- **Exit**: deselecting the current version removes the File List panel and its dependents.
- **Empty state**: Shows a contextual message when no version is selected or the version has no files.

### 6. File Details Panel
- **Entry**: appears only when a specific file is selected from the File List panel.
- **Actions**: read-only display; any navigation (e.g. selecting another file or using breadcrumbs) pops/replaces it.
- **Exit**:
  - Selecting a different file swaps content in place.
  - Changing folders removes the panel entirely until another file is picked.
- **Dependency**: requires an active File List panel; never shown without it.

---

## Lifecycle Rules & Edge Cases

- **Automatic pruning**: When the backend notifies that a selected file/version item no longer exists, the UI pops panels until it lands on a valid ancestor (e.g., from File Details back to File List, Version List, Item List, etc.).
- **Responsive resizing**: Moving between mobile and desktop widths does not alter the stack order; only the visible count changes (1 vs 2 panels).
- **Modal overlays**: Connection loss overlays block interaction but do not modify the panel stack.

This draft should serve as the single reference for panel transitions and slot usage. Future panels can extend the roster table and add subsections with their own flow rules.
