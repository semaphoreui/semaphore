# License Policy for Semaphore UI

This is the source of truth for which open-source licenses are acceptable in
Semaphore UI dependencies. The policy is enforced by
`scripts/check_policy.py` — if you change anything here, mirror the change
in that script.

The policy exists for two reasons:

1. **Customer obligations.** Our MSA commits us to *not* shipping
   components on terms that would require customers to license or assign
   their IP. That rules out copyleft licenses with strong reciprocity
   (GPL family, AGPL, SSPL).
2. **Project sustainability.** Source-available-but-not-OSS licenses
   (BUSL, Commons Clause, Elastic 2.0) carry use restrictions that conflict
   with our open-source distribution model.

## Allowed (no review needed)

These are permissive licenses with attribution-only obligations. Anything
under one of these can be added without further discussion.

| SPDX ID | Name |
|---------|------|
| MIT, MIT-0 | MIT License |
| Apache-2.0 | Apache License 2.0 |
| BSD-2-Clause | BSD 2-Clause "Simplified" |
| BSD-3-Clause | BSD 3-Clause "New" |
| ISC | ISC License |
| Zlib | zlib License |
| Unlicense | The Unlicense |
| CC0-1.0 | Creative Commons Zero |
| MPL-2.0 | Mozilla Public License 2.0 (file-level copyleft, OK for libraries) |
| 0BSD | BSD Zero Clause |
| BlueOak-1.0.0 | Blue Oak Model License |
| Python-2.0 | Python Software Foundation License |

**Note on MPL-2.0:** MPL is allowed because its copyleft obligation is at
the *file* level, not the project level. We can use MPL-licensed libraries
without making Semaphore UI MPL. If we ever modify an MPL-licensed file,
that file must remain MPL.

## Review required

These licenses are not blanket-forbidden but require a case-by-case
decision. Pause and ask before merging a dependency under one of these.

| SPDX ID | Why review |
|---------|------------|
| LGPL-2.1, LGPL-3.0 (and `-or-later`) | LGPL is fine for **dynamic** linking, but Go statically links by default. An LGPL Go library would force us to provide object files so users can relink. Acceptable only if the library is genuinely irreplaceable and we can document the relinking pathway. |
| EPL-1.0, EPL-2.0 | Eclipse Public License has weak copyleft and an explicit patent clause. Generally workable but worth a legal eyeball. |
| CDDL-1.0, CDDL-1.1 | File-level copyleft like MPL, but with extra requirements. Rare in our ecosystem. |
| EUPL-1.1, EUPL-1.2 | European Union Public License. Compatible with several other licenses; check the specific version. |
| OFL-1.1 | SIL Open Font License. Almost always fine for embedded fonts, but verify the font is being used as intended (not redistributed standalone with a new name). |

## Forbidden

A dependency under one of these licenses **must not be added**. If
`check_policy.py` flags one, the options are: (a) remove it, (b) replace
it with a permissively-licensed alternative, (c) move the functionality
out of process (subprocess invocation breaks the linkage in most cases —
this is why Ansible is acceptable).

| SPDX ID | Why forbidden |
|---------|--------------|
| GPL-2.0, GPL-3.0 (and `-only` / `-or-later`) | Strong copyleft. Linking forces Semaphore UI to be GPL, which conflicts with our distribution model. |
| AGPL-1.0, AGPL-3.0 | Same as GPL plus a network-interaction trigger. Particularly hostile to SaaS/self-hosted server software. |
| SSPL-1.0 | MongoDB's Server Side Public License. Not OSI-approved; imposes obligations on anyone offering the software as a service. |
| BUSL-1.1 | Business Source License. Time-delayed open source with use restrictions during the change window. Not OSS. |
| Commons-Clause | Restricts commercial sale; strips OSS status from any underlying license it's attached to. |
| Elastic-2.0 | Elastic License v2. Use restrictions; not OSS. |
| RSALv2 | Redis Source Available License v2. Same family as BUSL/Elastic. |
| CC-BY-NC-* | Creative Commons "Non-Commercial" variants. Not usable in a commercially-distributable product. |
| Facebook-BSD-Patents (legacy) | The old React patent grant clause. Removed years ago, but flagged here as a defensive measure. |

## Subprocess exception (External Runtime Tools)

Tools that Semaphore UI invokes via `os/exec` (Ansible, Terraform, OpenTofu,
Pulumi, Bash) are NOT linked into our binary and NOT redistributed by us.
Their licenses govern the user's installation of those tools on the host
system, not our distribution.

This is why we can integrate with Ansible (GPL-3.0) and Terraform (BUSL-1.1)
without violating this policy — we never embed their code, we only document
that the user must install them separately.

These tools are intentionally **not listed** in `THIRD-PARTY-LICENSES.md`.
Listing them would wrongly imply that they are distributed with the product
and would invite unnecessary scrutiny from a customer's legal team. Neither
`collect_licenses.sh` nor `generate_md.py` produces or accepts an external-
tools section.

## When the policy itself needs to change

If a dependency we genuinely need is forbidden, the path forward is:

1. Document the technical need (what does it do, what alternatives exist).
2. Get sign-off from a maintainer with the authority to make that call.
3. Update **both** this document and `check_policy.py`, in the same commit.
4. Note the exception in `THIRD-PARTY-LICENSES.md` so customers can see it.

Don't relax the policy silently. The whole point is that customers can
audit us, so the policy must be auditable too.
