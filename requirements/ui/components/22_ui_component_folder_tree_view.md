
# Component: FolderTreeView

Displays a **folder-only tree** for a repository’s contents.

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

