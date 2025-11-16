
# Component/Service: LiveUpdateSubscription

A non-visual component/service coordinating **change notifications** from the backend with the visible UI.

## Responsibilities

- Subscribe to backend-provided change notifications.
- Route relevant update events to panels that are currently mounted/displayed.
- Coordinate with Concertina Shell to handle deletions of the active object.

## Behaviour

- For each “scope” (e.g. a repository type, folder path, or file path), the UI may:
  - Register interest when the corresponding panel is mounted or becomes visible.
  - Deregister interest when that panel is unmounted or no longer visible.

- On receiving a change event:
  - If it affects visible data:
    - Trigger the appropriate panel to refresh its data.
  - If it indicates the **current object was deleted**:
    - Informs Concertina Shell, which then:
      - Pops affected panel(s).
      - Navigates to a valid ancestor level.

- The **exact protocol** (e.g. WebSocket message format, version numbers) is defined by the backend, but the UI assumes:
  - It can register for and receive scoped change notifications.
  - It can distinguish between “content changed” and “object deleted”.

