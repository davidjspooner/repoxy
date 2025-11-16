
# Component: SettingsDialog

Component implementation for the Settings Dialog panel.

## Responsibilities

- Present UI settings controls.
- Apply settings to the application (either immediately or on save).

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

Settings persistence details (e.g. localStorage vs backend) are **out of scope** for this UI spec but must not change the visual behaviour.

