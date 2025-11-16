
# Component: Toast

Represents a single transient notification.

## Responsibilities

- Convey success or error messages briefly and clearly.
- Be visually distinct yet unobtrusive.

## Visual Design

- Background:
  - **Pastel green** for success/info.
  - **Pastel red** for errors.
- Text:
  - Dark (black or near-black) for readability.
- May include:
  - Short title.
  - One-line or short multi-line message.

## Behaviour

- Appears at the **bottom** of the viewport, stacked with other toasts.
- Auto-dismiss after a configurable number of seconds.
- Optional close/dismiss icon for immediate removal.

The Toast component itself does not manage stacking or skipping; that is done by the Toast Queue.

