
# Component: SettingsDialog

Component implementation for the Settings Dialog panel.

## Responsibilities

- Present UI settings controls.
- Apply settings to the application immediately (live) as the user changes controls; there are no Save/Apply buttons.
- Persist selections to **localStorage** immediately so they survive reloads, falling back to a cookie (per browser best practices) if localStorage is unavailable.
- Default state when nothing is stored locally:
  - Theme: Light
  - Density: Comfortable

## Sections

1. **Density**
   - Options:
     - Compact
     - Comfortable
   - Changes may affect:
     - Padding in tables and lists.
     - Vertical spacing between UI elements.

2. **Theme**
   - Options:
     - Light
     - Dark
     - System (if supported).
   - Applying theme should update the overall application styling.

## Behaviour

- Opened from the Header’s user menu.
- Modal overlay; background content is dimmed and non-interactive.
- Close actions:
  - Close button (X).
  - Footer button (“Close” or “Done”).
- Optional: click on dimmed backdrop.
  - These close actions **do not** apply changes; they merely dismiss the dialog because settings already apply live.

Settings persistence is handled in-browser via localStorage/cookie fallback as described above; no backend round-trip is required in MVP.
