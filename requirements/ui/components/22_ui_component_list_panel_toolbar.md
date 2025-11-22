# Component: ListPanelToolbar

Reusable toolbar shared by list-based panels (e.g. repo types, repo instances, item list, version list).

## Purpose

- Provide a consistent control surface for switching between **list** and **tile** display modes.
- Offer a quick **filter/search** input so users can narrow long lists without leaving the panel.

## Behaviour

- Toolbar aligns to the **left** edge of the panel content for consistency.
- **Mode toggle**:
  - Two-state toggle (List / Tiles).
  - Emits a callback whenever the selected mode changes.
  - Does not persist mode globally; the owning panel decides whether to remember the selection.
- **Filter input**:
  - Standard text field with a magnifier icon when empty.
  - When text is entered, the trailing icon switches to a **clear (×)** button that empties the field on click.
  - Input value is pushed to the parent via callback so filtering can happen outside the toolbar.
  - Minimum width ~220 px, but can flex based on parent layout.
- The toolbar itself does not apply filtering or mode changes; it only surfaces user intent.

## States

- **Default**: list mode selected, filter empty, magnifier shown.
- **Filtered**: filter text updates the view immediately; toolbar shows clear icon until emptied.
- **Disabled**: parent may disable the toggle or input (e.g., when no items). Disabled visuals follow standard MUI styling.
