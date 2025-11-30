
# Repository Types Panel

Displays the list of repository **types** that currently contain cached data.  
These types map directly to the `type` field in the backend configuration (`repos` entries in `repoxy/conf/repoxy.yaml`), and the UI only shows types that have at least one configured repo instance (i.e. unused types remain hidden).

## Purpose

- Provide a **high-level entry point** into different ecosystems:
  - Docker/OCI containers
  - Terraform / OpenTofu providers
  - Terraform / OpenTofu modules
  - PyPI
  - Debian APT
  - (And any others supported by Repoxy)

Only types that actually have **data** are shown in the MVP.

## Parent / Children

- **Parent:** Root (conceptual).
- **Children:** Repository Browser Panel (for the selected repository type).

See `../22_ui_panel_flow.md` for how this panel sits in the concertina stack (single vs paired slots) and which panels follow it.

## Layout & Contents

- Main content is a **grid or list of tiles**, each rendered via the `RepoTypeTile` component (`../components/22_ui_component_repo_type_tile.md`).
- Each tile contains:
  - Repository type name (e.g. "Container", "PyPI").
  - Optionally a short description or example (non-essential for MVP).
  - Future (post-MVP): small logo/icon for the ecosystem.

## Interaction

- **Single click** (or tap) on a tile:
  - Tells the Concertina Shell to push a **Repository Browser** panel for that type.
  - The new panel becomes the **right-hand** panel.
- On desktop:
  - Repository Types may still be visible as the **left-hand** panel while the Repository Browser is on the right if the stack design chooses to keep two levels visible.
  - Alternatively, Repository Types can be replaced entirely; this is up to the Concertina Shell configuration.
- On mobile:
  - Repository Types is replaced by Repository Browser, showing only the new panel.

## States

- **Loading:**  
  - Show a skeleton or spinner in place of tiles.

- **Empty:**  
  - If there are no repository types with configured repositories (e.g. every entry in `repos` is disabled or absent):
    - Show a message like: “No cached repositories available.”
    - MVP: no call-to-action, since adding repositories is post-MVP.

- **Error:**  
  - Inline message (“Failed to retrieve repository types.”).
  - Red toast via Toast system.

- **Populated:**  
  - Normal tile grid.
