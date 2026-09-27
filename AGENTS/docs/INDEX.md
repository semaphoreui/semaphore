# docs — portable standard by zone

Approaches and conventions that would serve another project, not this project's own facts. One
folder per zone, each with its own `INDEX.md`. This file routes to a zone; the zone's index routes
to the doc.

## Zone index (example)

The rows below are example data, illustrating the shape a real routing table takes; replace them
once an actual zone exists.

| When you need it | Read |
| --- | --- |
| Example: Writing backend code: models, routes, permissions, entity codes | `backend/INDEX.md` |
| Example: Running or writing backend tests — mandatory before the first command | `backend/INDEX.md` |
| Example: Editing the frontend: app layout, import boundaries, i18n | `frontend/INDEX.md` |
| Example: Setting up a local stack: reverse proxy, HMR behind a proxy, tunnels — the approach, not this project's ports | `infra/INDEX.md` |
| Example: Editing an env file or another human-read file: grouping, order, language, line length | `standards/INDEX.md` |
