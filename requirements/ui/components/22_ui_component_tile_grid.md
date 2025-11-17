
# Component: Tile Grid

Layout component for arranging tiles such as `RepoTypeTile`.

## Responsibilities

- Arrange child tiles in a responsive grid or list.
- Handle cases where the number of tiles changes (e.g. new repo types appear).

## Behaviour

- Uses **fixed-width tiles** (≈260 px) and allows them to wrap naturally as the viewport changes.
- Does not assume a fixed number of columns; instead tiles flow across each row while maintaining constant width and spacing.

The Tile Grid is a simple layout helper; all semantics and actions belong to the child tiles.
