
# Settings Dialog Panel

A modal dialog for **personalisation and basic UI settings**.

## Purpose

- Allow the user to adjust:
  - **Density** (compact vs comfortable).
  - **Theme** (light/dark/system).

## Invocation

- Triggered from the **user menu** in the Header Bar (see `../components/22_ui_component_header_bar.md`).
- Option name: likely “Settings” or “Preferences”.

## Layout

- Modal dialog overlay centered in the viewport.
- Background:
  - The main UI (header, panels, footer) is dimmed and made non-interactive.
- Content sections:
  1. **Density Settings**
     - Radio buttons or segmented control:
       - Compact
       - Comfortable
  2. **Theme Settings**
     - Radio buttons or segmented control:
       - Light
       - Dark
       - System (if supported)
  3. **Future Settings Placeholder**
     - Space where additional options can be added later.

- Footer of the dialog:
  - “Save” or “Apply” button.
  - “Cancel” or “Close” button.
  - For MVP it is acceptable for settings to apply immediately on change (no explicit save), but having an explicit close action is still required.

## Interaction

- Opening:
  - Blocks interaction with everything behind it.
- Closing:
  - Via close icon/button.
  - Via explicit “Close/Done” button.
  - Optional: clicking on the dimmed backdrop.

## States

- **Populated**
  - Primary and only expected state in MVP.
- **Error**
  - Only relevant if persistence of settings fails (e.g. unable to save to storage).
  - If that happens:
    - Inline error in the dialog.
    - Red toast summarising the problem.
