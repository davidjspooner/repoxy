
# File Details Panel

Shows detailed information about a single file.

## Purpose

- Provide **read-only metadata** about a selected file:
  - For verification (e.g. size, checksum).
  - For debugging (e.g. number of downloads).

## Parent / Children

- **Parent:** Repository Browser Panel.
- **Children:** None in MVP (no further drill-down).

## Layout & Contents

The File Details panel is a **single-column details layout**. Typical sections:

1. **Header Section**
   - File name (full, possibly with path).
   - Repository type and repository identifier (if relevant to display).
   - Folder path.

2. **Metadata Section**
   - File size.
   - Last modified timestamp.
   - Content type (if applicable).
   - Checksums (MD5, SHA-256 etc. depending on backend).

3. **Usage Section**
   - Number of times downloaded or requested (if the backend tracks this).
   - Last accessed timestamp (if available).

4. **Future Actions (Post-MVP)**
   - Download button.
   - Delete / clean-up actions.
   - These are **not active in MVP** and, if shown, must be disabled.

## Interaction

- Opened when:
  - A file row is selected in the File List Table.
- Closed when:
  - User clicks a higher-level breadcrumb, or
  - Concertina Shell pops this panel in response to navigation.

The File Details panel itself does **not** provide an explicit “Back” button; the shell controls navigation via breadcrumbs.

## States

- **Loading**
  - When the panel first opens and metadata is being fetched.
  - Display skeletons or placeholders.

- **Error**
  - Inline error message if metadata cannot be retrieved.
  - Red toast for the error.

- **Populated**
  - Full metadata visible.

## Live Updates

- If metadata changes while this panel is open (e.g. download count incremented by other clients):
  - Panel updates fields as new data arrives.
- If the **file is deleted** while this panel is visible:
  - Concertina Shell automatically:
    - Pops File Details.
    - Shows the Repository Browser panel as the active panel.
  - A toast may be used to indicate that the file no longer exists.

