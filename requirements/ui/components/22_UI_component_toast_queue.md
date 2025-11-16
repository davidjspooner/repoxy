
# Component: ToastQueue

Manages the **lifecycle and display order** of Toast components.

## Responsibilities

- Maintain a queue/stack of pending toasts.
- Decide when to show or skip toasts.
- Handle auto-dismiss timing and stacking.

## Behaviour

- New toast requests are added to the queue with metadata (type, message, timestamp, etc.).
- Display:
  - Multiple toasts may be visible at once in a vertical stack.
  - New toasts appear above or below existing ones depending on desired convention.

## Skipping Logic

- If an **error toast** becomes irrelevant (e.g. before it appears, a corresponding retry succeeds):
  - The queue may decide to **skip** that error toast.
- This avoids a UX where users see “error!” immediately followed by “all good!” for transient situations.

The exact mapping between certain “success” events and earlier “error” events can be tuned by implementation, but the behaviour should remain consistent.

