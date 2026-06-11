# feat: search Task Templates by name (#3713)

Adds a name search to the **Task Templates** page so long template lists can be
filtered quickly. Closes [#3713](https://github.com/semaphoreui/semaphore/issues/3713).

## What changed

- A search field sits in a filter bar directly below the view tabs.
- Live, case-insensitive filtering of the templates table by **name**.
- A live result count (`{x} of {y} templates`) appears while searching.
- An empty state (`No templates match your search`) when nothing matches.
- Three new i18n strings in `en.js`.

Files touched:

- `web/src/views/project/Templates.vue`
- `web/src/lang/en.js`

The change is self-contained — it reuses the existing `items` from the
`ItemListPageBase` mixin and Vuetify's built-in `v-data-table` search, so it
works for both the default and per-view template lists with no API change.

## Apply as a branch

From the root of your `semaphore` checkout:

```bash
# 1. create the branch off your current base (e.g. develop)
git checkout -b feat/template-search

# 2. copy this folder's patch into the repo root, then:
git apply template-search.patch

# 3. verify, then commit
git add web/src/views/project/Templates.vue web/src/lang/en.js
git commit -m "feat(web): search task templates by name (#3713)"

# 4. push and open a PR
git push -u origin feat/template-search
```

If `git apply` reports offsets (because your base moved), use:

```bash
git apply --3way template-search.patch
```

## Or copy the full files

The fully-patched files are included in this folder at the same relative paths,
so you can also just overwrite them:

```
branch-feat-template-search/web/src/views/project/Templates.vue  ->  web/src/views/project/Templates.vue
branch-feat-template-search/web/src/lang/en.js                   ->  web/src/lang/en.js
```

## Follow-ups (optional)

- Add the same three keys to the other locale files under `web/src/lang/` if you
  want translated placeholders/empty-state text.
- If template lists grow very large, consider moving filtering server-side via a
  `?filter=` query param on the templates endpoint instead of client-side.
