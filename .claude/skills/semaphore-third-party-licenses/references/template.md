# THIRD-PARTY-LICENSES.md output format

This is the canonical structure that `scripts/generate_md.py` produces.
Reference for humans reviewing the generator output, and for anyone who
needs to understand what each section is for.

## Structure

```
# Third-Party Licenses

[intro paragraph explaining what this file is and why it exists]

_Generated on YYYY-MM-DD HH:MM UTC by scripts/collect_licenses.sh._

[regeneration instructions — always include these so the file is
 reproducible by anyone, not just the original author]

## Summary

[total component count]

| Ecosystem | Components |
|-----------|------------|
| Go (backend)        | NN |
| npm (frontend)      | NN |

### License distribution

| License | Count |
|---------|-------|
| MIT          | NN |
| Apache-2.0   | NN |
| BSD-3-Clause | NN |
| ...          | NN |

## Go Backend Dependencies

[one-line description of what these are and where they come from]

| Component | Version(s) | License | Source |
|-----------|------------|---------|--------|
| `github.com/foo/bar`  | v1.2.3 | MIT | [link](https://...) |
| `github.com/foo/baz`  | v0.4.1, v0.5.0 | Apache-2.0 | [link](https://...) |
| ...

## Frontend Dependencies (npm)

[one-line description]

| Component | Version(s) | License | Source |
|-----------|------------|---------|--------|
| `vue`        | 3.4.21 | MIT | [link](https://...) |
| `vuetify`    | 3.5.8  | MIT | [link](https://...) |
| ...

---

## License texts

[paragraph explaining where to find full license texts and how to report
 missing/incorrect attributions]

<!-- end of generated file -->
```

## Why this structure

**Summary table at the top.** Lets a reader (especially a customer's
compliance team) see the shape of the dependency tree at a glance
without scrolling through hundreds of rows.

**License distribution.** Surfaces concentration risk. If 95% of deps are
MIT/Apache and 1 is MPL-2.0, that 1 deserves attention. It also makes it
easy to spot a forbidden license that somehow slipped through (a single
GPL-3.0 row in the distribution table is a very loud red flag).

**Sorted alphabetically within each section.** Diffs between releases
become useful — a new dependency shows up as an added row in one
predictable place, not scattered across the file.

**Versions collapsed onto one line.** A dependency pinned at multiple
versions (common with transitive Go deps that get MVS-resolved differently
in different parts of the tree) shouldn't take 5 rows. One row, comma-
separated versions.

**External tools deliberately excluded.** Ansible, Terraform, OpenTofu,
Pulumi, and Bash are invoked as subprocesses, not linked, and not
redistributed with Semaphore UI. They are not listed in this file —
including them would wrongly imply distribution. The license policy
document covers them under "Subprocess exception" if a reviewer asks
why they're absent.

**Regeneration instructions in the header.** Means the file is not a
hand-maintained artifact that drifts. Anyone can rebuild it and verify
it matches what's actually in the dependency manifests.

**Footer comment marker.** `<!-- end of generated file -->` lets CI
verify the file wasn't truncated mid-write — useful when generation runs
in a pipeline and might fail partway.

## What NOT to include

- **Full license texts inline.** Don't copy the MIT license text 200
  times into one file. Refer readers to the source URL. The packages
  themselves carry their LICENSE files; the user has them on disk after
  `go mod download` / `npm install`.
- **CVE / vulnerability info.** That belongs in security advisories,
  not the license document. Mixing the two confuses the reader and makes
  the file go stale faster.
- **Internal-only dependencies.** Build tooling (linters, test
  frameworks, code generators) doesn't ship to customers. Production
  dependencies only.
- **Hand-edited rows.** If something needs a special note, add it via
  the generator (extend `generate_md.py`), not by editing the output
  file. Hand edits get destroyed on the next regeneration.
