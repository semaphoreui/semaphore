# research — registry

All research lives in the `research` MCP: its body (write-up, areas, notes, cited
sources) is stored there and nowhere else. This folder holds only this file — a shortlist of the
`RESEARCH@` codes that matter for the project. There are no per-research files or folders here.

## Rules

- Before any research the user asks for: read the shortlist below, then run
  `mcp__research__research_list` / `research_get`.
- Conduct and store new research in the MCP: `research_create` → `area_create` → `body_add` /
  `note_create`, citing kept sources as `SOURCE@`.
- Get a skill with `skills_list` / `skill_get` when its guidance is needed.
- Get the `mermaid` skill before drawing a diagram the user asks for.
- Never copy findings, conclusions or numbers out of the MCP into a repo file — a copy rots and
  diverges from the original silently. Keep only the reason to open one, here in the shortlist.
- After creating or updating a research that matters for the project, add or refresh its row in
  the same pass.

## Shortlist

Curated pointer to the `RESEARCH@` codes that matter for this project specifically — hand
maintained, not a full export of the MCP catalogue.

| Code | Why it matters here |
| --- | --- |
| `RESEARCH@bc356e4b9b` | Kestra YAML deep dive and the agreed `semaphore/v1` Workflow YAML format with its phased plan — start here for any Workflows-as-Code task. |
| `RESEARCH@a97b871b1d` | How CI/CD products pass outputs → inputs between steps (GitHub Actions, GitLab, Tekton, Argo, Kestra, AWX, Spacelift) checked against the workflow outputs plan — read before changing the outputs contract or node input mappings. |
| `RESEARCH@5bf4fc406f` | Workflow canvas libraries benchmark (n8n, Kestra, Airflow, Dify) and why Drawflow stays on Vue 2.7 — read before proposing a canvas change. |
| `RESEARCH@e6d5aa256f` | Wider node-editor market survey (Vue/React libraries, embeddable platforms, licences, canvas UX mechanics). |
| `RESEARCH@e21ac95e0a` | Full audit of the docs submodule (107 pages): per-section AREA@ reports, features documented but absent from develop, P1–P3 plan — start any docs work here. |
| `RESEARCH@c1125073d7` | Docs IA benchmarks (Backstage, Prometheus, Grafana, Argo CD, AWX, Diátaxis) and the agreed target structure for the docs. |
| `RESEARCH@b36f531662` | Earlier GitLab and Terraform docs structure comparison against Semaphore docs. |
| `RESEARCH@24c6ed4f20` | Docs landing-page principles applied to `docs/docs/README.md`. |
| `RESEARCH@e5c7579df7` | Progressive disclosure rules for rare options in the template form (used for Galaxy args). |
| `RESEARCH@702f04c19a` | In-app contextual help panel: offline, build-embedded docs in the Vue app. |
| `RESEARCH@7732ef5556` | How GitLab documents Docker/Kubernetes runner executors — model for the Pro executor docs. |
| `RESEARCH@d8d16e3974` | Per-job secret delivery to remote workers in HA systems — basis for task-bound survey secrets. |
| `RESEARCH@6a7e8ecd1b` | Encryption key separation for secrets at rest (one key vs purpose-specific keys, rotation). |
| `RESEARCH@c81c73d4f9` | How competitors deliver audit logs to SIEM reliably (retries, buffering, integrity). |
| `RESEARCH@8b7b9c2037` | Backstage integration options and what Semaphore lacks for them (API, webhooks, tokens). |
| `RESEARCH@61d07ff4d6` | New Ansible features relevant to Semaphore (ANSIBLE_SECRETS_INPUT_FILES and later). |
| `RESEARCH@84146f310f` | Overview of the separate `integration-tests` Java/Bookwright repo: libraries, layout, pipeline. |
| `RESEARCH@037925c6fb` | Distributing the GPG signing key for release artifacts. |
