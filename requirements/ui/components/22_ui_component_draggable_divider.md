
# Component: DraggableDivider

Vertical divider between two visible panels in the Concertina Shell (desktop only).

## Responsibilities

- Allow the user to resize the relative width of the **left** and **right** ScrollableViewPorts.
- Enforce minimum widths for both panels.

## Behaviour

- Visible only when **two panels** are rendered side by side (desktop/wide layout).
- Grabbable affordance:
  - Clearly visible handle area (e.g. a pill-shaped grip centered inside a wider divider track) so the separator stands out against panel backgrounds.
- Dragging:
  - Horizontal drag adjusts the flex/width of left and right panels.
  - Constraints:
    - Neither panel may shrink below a configured minimum width.
    - Divider movement is clamped accordingly.

- On mobile/narrow:
  - Divider is hidden and inactive; only a single panel is rendered.
