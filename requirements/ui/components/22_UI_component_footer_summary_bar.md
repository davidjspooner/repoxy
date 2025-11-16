
# Component: Footer Summary Bar

A persistent footer showing a **brief summary** of the currently selected object.

## Responsibilities

- Provide a quick, always-visible summary of:
  - Selected repository type.
  - Selected repository/folder/file (if any).

## Content Examples

Depending on selection, it might display:

- **Repository Types view only:**
  - “Viewing repository types.”

- **Inside a specific repository type and folder:**
  - “Type: Docker | Repo: my-repo | Path: /images/v1”

- **With a file selected:**
  - “File: manifest.json | Size: 12 KB | Modified: 2025-11-16 12:34”

The footer intentionally uses **less detail** than the File Details panel.

## Behaviour

- Fixed at bottom of viewport.
- Updated whenever the **selection context** changes:
  - Changing folder.
  - Selecting a file.
  - Navigating back via breadcrumbs.

