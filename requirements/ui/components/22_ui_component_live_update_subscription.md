
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

- On connection loss:
  - Immediately instructs the Concertina Shell to display a **Connection Status modal** that blocks interaction (there is no offline mode).
  - Drives an **exponential backoff retry loop** that grows until it reaches a maximum interval of **2 minutes**, then continues retrying every 2 minutes until success or the tab is closed.
  - Updates the modal every second with the remaining time until the next attempt.
  - Supports a user-triggered **Retry Now** action exposed in the modal; pressing it triggers an immediate retry and restarts the backoff ladder.
  - Once reconnection succeeds, hides the modal and resumes normal notifications (including refreshing any stale panels).

- The **exact protocol** (e.g. WebSocket message format, version numbers) is defined by the backend, but the UI assumes:
  - It can register for and receive scoped change notifications.
  - It can distinguish between “content changed” and “object deleted”.
  - It can detect connection loss/recovery events in order to drive the modal countdown state.
