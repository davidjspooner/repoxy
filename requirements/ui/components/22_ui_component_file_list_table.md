
# Component: FileListTable

A tabular view of files in the currently selected folder.

## Responsibilities

- Display files and basic metadata.
- Provide sorting within the current folder.
- Allow selecting a file to open File Details.
- Assume folder contents are **small enough** to render in one table for MVP; pagination and filtering are deferred to post-MVP.

## Columns (MVP)

- File Name
- Last Modified Date/Time
- Size

Additional columns can be added later, but these are sufficient for MVP.

## Sorting

- Clicking a column header:
  - Sorts by that column.
  - Toggles between ascending/descending.
- Only one sort key active at a time.

## Selection

- Row click:
  - Marks the file as selected.
  - Notifies parent (Repository Browser panel).
  - Parent then instructs Concertina Shell to open a File Details panel.

## Scrolling

- Vertical scroll for large file lists.
- Horizontal scroll when combined column widths exceed available space.
