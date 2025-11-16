
# Component: Footer Summary Bar

A persistent footer showing a **brief summary** of the currently selected object.

## Responsibilities

- Provide a quick, always-visible summary of:
  - Selected repository type.
  - Selected repository/folder/file (if any).
- Surface the **last updated timestamp** for the currently selected object (based on metadata returned by the backend or live-update events).
- Indicate connectivity issues by appending ` (disconnected)` when the LiveUpdateSubscription reports loss of connection.

## Content Examples

Depending on selection, it might display:

- **Repository Types view only:**
  - “Viewing repository types.”

- **Inside a specific repository type and folder:**
  - “Type: Docker | Repo: my-repo | Path: /images/v1 | Updated: 2025-05-18 14:03”

- **With a file selected:**
  - “File: manifest.json | Size: 12 KB | Modified: 2025-11-16 12:34 | Updated: 2025-05-18 14:03”

- **Disconnected state (suffix only):**
  - “… | Updated: 2025-05-18 14:03 (disconnected)”

The footer intentionally uses **less detail** than the File Details panel.

## Behaviour

- Fixed at bottom of viewport.
- Updated whenever the **selection context** or **last updated timestamp** changes:
  - Changing folder.
  - Selecting a file.
  - Navigating back via breadcrumbs.
  - Receiving a live-update notification with a newer modified time.
- Appends ` (disconnected)` to whatever text is already present when the LiveUpdateSubscription reports loss of connectivity; removes the suffix automatically once reconnected.
