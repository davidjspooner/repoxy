
# Root Panel

The **Root Panel** is the conceptual top of the navigation stack.

## Behaviour

- On app load, the Root Panel immediately transitions into the **Repository Types Panel**.
- There is no panel “above” Root; the user cannot navigate further back.
- In most practical implementations, Root may be purely virtual:
  - The UI simply renders Repository Types as the initial panel.

## States

- Root does not have independent visible states; all user-visible behaviour is delegated to the Repository Types Panel.

