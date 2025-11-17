
# Component: RepoTypeTile

Represents a single repository type as a selectable tile.

## Responsibilities

- Display basic information about a repository type.
- Trigger navigation into a **Repository Browser** panel when selected.

## Contents

- Primary label:
  - Human-readable type name, e.g. “Docker”, “PyPI”, “APT”.
- Optional secondary text:
  - Short description or example (optional in MVP).
- Future (post-MVP):
  - Logo/icon for the ecosystem.

## Interaction

- Click/activate:
  - Notify parent (Repository Types panel) of the selection.
  - Parent then instructs Concertina Shell to push the corresponding Repository Browser panel.
- Tiles show a **clear selected state** (e.g. thicker border, stronger elevation) so users can see the currently active repository type.

- Focus/keyboard support:
  - Should be focusable and activatable via keyboard for accessibility.
