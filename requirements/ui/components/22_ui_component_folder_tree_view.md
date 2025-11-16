
# Component: FolderTreeView

Displays a **folder-only tree** for a repository’s contents.  
When scoped to a repository type, the **root-level nodes represent repository names** returned by the backend (e.g. entries from `repos` in `repoxy/conf/repoxy.yaml`).  
Folder depth beneath each repo is type-specific, but the tree enforces a convention that files only appear at the leaves (a folder either has child folders or files, not both).

## Responsibilities

- Show hierarchical **folder structure**.
- Allow users to select a folder to view its files.

## Behaviour

- Contains only folder nodes:
  - No files are displayed in the tree to avoid massive node counts.
- Each node:
  - Can expand/collapse to show/hide children.
  - Can be selected to make that folder “active”.

## Interaction

- On folder select:
  - Emits an event to parent (Repository Browser panel).
  - Parent then loads and displays the corresponding files in the File List.

- State preservation:
  - Expanded/collapsed state should remain stable while navigating within the same Repository Browser panel instance.

## Scrolling

- May have its own vertical scrollbar if the tree is long.
- Typically no horizontal scroll; however long folder names may require it.
