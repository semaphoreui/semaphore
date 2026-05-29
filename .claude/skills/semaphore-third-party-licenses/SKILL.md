---
name: semaphore-third-party-licenses
description: >
  Generate or update the THIRD-PARTY-LICENSES.md file for the Semaphore UI repository. Use this skill whenever the user 
  asks to create, update, regenerate, audit, or refresh third-party license attributions, OSS notices, SBOM-style 
  license inventories, or NOTICE files for Semaphore UI. Trigger on phrases like "third-party licenses", "OSS notices",
  "license attribution file", "SBOM for licenses", "NOTICE file", "audit dependencies",
  "licenses for go.mod / package.json", or any request that mentions compliance with customer agreements requiring an
  OSS components list. Also trigger when the user says "we added a new dependency, regenerate licenses". This skill
  knows the Semaphore UI layout (Go backend + Vue.js frontend), the project's license policy (allowlist/denylist),
  and the canonical output format.
---

# Semaphore UI — Third-Party License Generator

This skill produces `THIRD-PARTY-LICENSES.md` for the Semaphore UI repository. The file lists every open-source
dependency shipped with Semaphore UI by **name**, **version**, **license**, and **source URL**, grouped by component. It
exists to satisfy contractual obligations toward customers (identification of OSS components by name, version,
and license type) and to preserve attribution required by permissive licenses (MIT/BSD/Apache-2.0).

## Repository layout you can expect

Semaphore UI is a self-hosted product. The repository ships:

- **Go backend** — module file at `go.mod`, source under `pkg/`, `services/`, `db/`, etc.
- **Vue.js frontend** — `web/package.json` (or `web2/package.json` in some branches), built into static assets and
  embedded in the Go binary.

External CLI tools (Ansible, Terraform, OpenTofu, Pulumi, Bash) are invoked at runtime as separate operating-system
processes. They are NOT linked into the binary and NOT redistributed with Semaphore UI, so they are intentionally
**not listed** in `THIRD-PARTY-LICENSES.md` — their license terms apply to the user who installs them on the host,
not to Semaphore UI's distribution. The license policy still discusses why this is contractually safe (see
`references/license_policy.md`, "Subprocess exception").

If the layout differs (monorepo restructure, new submodule, etc.), inspect the repo first with
`find . -name "go.mod" -o -name "package.json" -not -path "*/node_modules/*"` and adapt.

## Workflow

Follow these steps in order. Don't skip the policy check — it's the part that actually protects the project from
accidentally shipping a GPL'd dependency.

### 1. Confirm scope with the user

Before generating anything, ask (briefly) which targets to include if it's not obvious from context:

- Backend only? Frontend only? Both?
- Production dependencies only, or also dev dependencies?

**Default**: both backend and frontend, **production-only** (dev dependencies are not distributed to customers and don't
need attribution). Override only if the user explicitly asks for dev deps too.

### 2. Collect raw license data

Run `scripts/collect_licenses.sh` from the repo root. It does the following:

- Sets `GOWORK=off` and runs `go-licenses report ./... --template scripts/go_template.tpl` for the Go backend, so
  workspace siblings (e.g. `pro_impl`) are not pulled in.
- Rolls sub-package rows up to their parent module (uses `go list -m all` for the module → version map) and drops
  any module path that belongs to the root module's namespace (first-party packages, including `replace … => ./pro`
  targets).
- Applies a small `OVERRIDES` table in the script for modules where `go-licenses` can't auto-detect a license
  (e.g. `modernc.org/mathutil` → BSD-3-Clause). Extend the table when new "Unknown" entries appear.
- Runs `license-checker --production --json --excludePrivatePackages` in the frontend directory and strips the
  low-confidence `*` suffix from license strings so `MIT*` collapses into `MIT`.
- Writes raw output to `.licenses-cache/go.tsv` and `.licenses-cache/npm.json`.

If the tools aren't installed, the script prints the install commands. Don't try to install them silently — show the
user what needs to be added so they can decide whether to add them to CI.

### 3. Run the policy check

Run `scripts/check_policy.py .licenses-cache/`. This compares every detected license against the project policy in
`references/license_policy.md`:

- **Allowed** (green): MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, Zlib, Unlicense, CC0-1.0, MPL-2.0 — proceed.
- **Review** (yellow): LGPL-2.1, LGPL-3.0, EPL-2.0, CDDL-1.0, custom/unknown — pause and ask the user. LGPL is fine for
  dynamic linking but Go statically links, so flag it for review.
- **Forbidden** (red): GPL-2.0, GPL-3.0, AGPL-3.0, SSPL-1.0, BUSL-1.1, Commons Clause, "non-commercial" clauses — STOP.
  Do not proceed with file generation. Report the offending package and ask the user how to handle it (replace, vendor,
  or accept and document the consequence).

**This step is non-negotiable.** A forbidden license slipping into THIRD-PARTY-LICENSES.md isn't just a documentation
problem — it's a contract breach with every customer. If `check_policy.py` returns a  non-zero exit code, surface the 
failures prominently and stop.

### 4. Generate the markdown

Run `scripts/generate_md.py .licenses-cache/ > THIRD-PARTY-LICENSES.md`. This produces the file using the structure
described in `references/template.md`. Key points the generator handles:

- Sorts entries alphabetically within each section.
- Groups by component category (Go backend / npm frontend).
- Deduplicates packages that appear at multiple versions (lists all versions on one line).
- Renders a summary table at the top with license counts.
- Adds a generation timestamp and a note pointing to `scripts/collect_licenses.sh` so the file is reproducible.

### 5. Verify and commit

After generation:

1. Diff against the previous `THIRD-PARTY-LICENSES.md` (if it exists). New rows = new dependencies — confirm with the
   user that they're expected.
2. Verify the file isn't truncated and ends with the closing footer (`<!-- end of generated file -->`).
3. Suggest the commit message: `chore(licenses): regenerate THIRD-PARTY-LICENSES.md for vX.Y.Z`.
4. If a CI pipeline exists (`.github/workflows/`), suggest adding a job that runs `scripts/check_policy.py` on every
   PR — that's how the project prevents regressions, not by re-checking manually each release.

## Notes on edge cases

**Vendored dependencies.** If the repo has a `vendor/` directory, `go-licenses` will read from it. That's fine —
vendored code still ships and still needs attribution.

**Dual-licensed packages.** Some packages are offered under "MIT OR Apache-2.0". Pick MIT (the simpler one) and note the
choice in the entry: `License: MIT (also available under Apache-2.0)`.

**Packages with no detectable license.** `go-licenses` flags these as "Unknown". Don't ship the file with Unknowns —
manually inspect the repo, fill in the SPDX identifier, and if there's truly no license, treat the package as
forbidden (no license = "all rights reserved", which is incompatible with redistribution).

**Ansible / Terraform / OpenTofu / Pulumi / Bash.** Do NOT list them in `THIRD-PARTY-LICENSES.md`. They are invoked
as separate OS processes (`os/exec`), not linked, not redistributed. Their licenses bind the user who installs them
on the host, not Semaphore UI's distribution. Listing them would imply otherwise and is an unnecessary signal to a
reviewing customer's legal team. The "Subprocess exception" section in `references/license_policy.md` explains the
reasoning if a reviewer asks.

**First-party packages with no LICENSE.** The root module and any `replace … => ./<dir>` target (e.g. `./pro`) live
inside this repo and don't need to be self-attributed. The collection script filters them out by the root module
prefix from `go.mod`. If you see "Unknown license" rows for `github.com/<this-org>/<this-repo>/…`, the filter is
working correctly — those rows shouldn't appear in the final cache.

**Embedded license texts.** For packages under MIT/BSD/ISC, the license requires preserving the copyright notice. The
generator embeds the full LICENSE file text from each package's source. For Apache-2.0 packages, it embeds the NOTICE
file if one exists.

## What this skill does NOT do

- It does not generate a CycloneDX or SPDX SBOM. If the user wants a machine-readable SBOM (for compliance scanners,
  customer attestations, etc.), point them at `syft packages dir:. -o cyclonedx-json` as a separate step.
  THIRD-PARTY-LICENSES.md is the human-readable artifact.
- It does not check for CVEs. That's a security concern, not a licensing one — use `govulncheck` and `npm audit` for
  that.
- It does not modify go.mod or package.json. The skill is read-only with respect to the dependency manifests; it only
  generates the attribution document.

See `references/license_policy.md` for the full policy and `references/template.md` for the output format.
