
# Repository Browser Panel

Displays the **folder tree and files** for a specific repository type (and possibly specific repository instance) in a single panel.

## Purpose

- Allow users to:
  - Navigate the **folder hierarchy** (left).
  - View and sort **files within the selected folder** (right).
  - Select a file to open the File Details panel.

## Parent / Children

- **Parent:** Repository Types Panel.
- **Child:** File Details Panel.

## Layout

Within this **single panel**, we have a **two-column layout** managed by the panel’s internal layout (not the Concertina Shell):

1. **Left Column – Folder Tree (FolderTreeView component)**  
   - Displays **folders only**.
   - No file nodes appear in the tree.
   - Behaves like a typical file system tree:
     - Expand/collapse folders.
     - Click selects a folder.

2. **Right Column – File List (FileListTable component)**  
   - Displays files within the **currently selected folder**.
   - Tabular layout with sortable headers:
     - File name.
     - Last modified date.
     - Size.
   - Includes a simple **filter/search input** (MVP):
     - Text-based filter that narrows the visible rows (e.g. substring match on name).
     - Filtering happens within the UI; the underlying folder contents are not re-requested unless required by backend design.

## Independent Scrolling

- The Repository Browser panel itself is wrapped in a **Panel Container** that manages scrolling.
- Inside the panel:
  - Folder tree column may have its own vertical scrollbar (if long).
  - File list column may have its own vertical scrollbar.
- Horizontal scrolling:
  - Primarily used in the File List if columns overflow.
  - Folder tree horizontals should be rare but allowed if paths are long.

## Interaction Details

- **Selecting a Folder (Tree View)**
  - Highlights that folder.
  - Triggers a refresh of the File List for that folder.
  - May cause a loading state within the File List only (tree remains visible).

- **Selecting a File (File List Table)**
  - Row click opens a **File Details** panel to the right (concertina shell action).
  - The Repository Browser panel remains in place as the left-hand context on desktop.
  - On mobile, Repository Browser is conceptually still on the stack but not visible.

- **Sorting**
  - Clicking a table header toggles sort order for that column (ascending/descending).
  - Only one active sort column at a time in MVP.

- **Filtering**
  - Changes to the filter input narrow down the visible rows in the File List.
  - Filter state is local to this panel instance and resets when navigating to a different repository type (unless preserved intentionally by implementation).

## States

- **Loading**
  - Initial folder tree + file list load:
    - Show skeleton or spinner in both columns.
  - When changing folders:
    - Tree stays visible.
    - File List area may show a smaller loading indicator.

- **Empty**
  - If the repository type has **no repositories or folders**:
    - Show empty state message.
  - If a folder has **no files**:
    - File List shows a “No files in this folder” message.

- **Error**
  - Tree and/or file list show an inline error if loading fails.
  - Also triggers a red toast summarising the failure.

- **Populated**
  - Normal tree + list.

## Live Updates

- When new files are added/removed in the selected folder by any client:
  - The File List is updated automatically.
- When a folder is added/removed:
  - Tree view updates accordingly.
- If the **current folder** is deleted:
  - The panel requests its parent folder.
  - If none exists, the concertina shell navigates upward to the Repository Types panel.

