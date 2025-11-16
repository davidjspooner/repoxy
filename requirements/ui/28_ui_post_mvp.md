
# Post-MVP UI Features (Out of Scope for MVP)

This file explicitly lists features that are **known future enhancements** and must **not** be implemented in the MVP.  
UI elements for these features may appear **disabled** to hint at future capabilities.

---

## Authentication & User Management

- Full **login/logout** flows.
- Multiple users with roles/permissions.
- Account/profile pages and editable user settings.

In the MVP:
- Only a single implicit admin user exists.
- “Account” and “Logout” entries are visible in the header user menu but disabled.

---

## Repository & Registry Management

- Add new **repository types**.
- Remove existing repository types.
- Add/remove **upstream registries** or repositories within a type.
- Edit repository configuration via UI.

In the MVP:
- Repository types and underlying registries are considered **pre-configured** via external configuration.

---

## File Operations

- **Download** files from the UI.
- **Delete** files individually.
- Run clean-up flows such as:
  - A **clean-up wizard** to bulk-remove old or unused artifacts.
  - Per-file or per-folder deletion actions.

UI-wise, these would likely surface as:

- Buttons or menus in:
  - File List Table (per row or bulk actions).
  - File Details panel.
- Additional confirmation dialogs or wizards.

Currently, these controls must either:
- Not be present at all, or
- Be present but disabled and clearly marked as future functionality.

---

## Advanced Navigation & Search

- **Global search** across repository types.
- Advanced filtering (by tag, date ranges, size ranges, etc.).
- Saved searches or pinned views.

---

## Additional Visualisation & Metrics

- Dashboards showing:
  - Request counts and hit/miss ratios.
  - Error rates per repository type.
  - Storage consumption and trending.

These would likely introduce new panels and components beyond the MVP’s browsing-oriented UI.

