---
name: semaphore-config-schema
description: >
  Generate or update the JSON Schema (in YAML form) that describes the Semaphore configuration file. The schema
  lives at `config.schema.yaml` in the repository root and is derived from the Go struct `util.ConfigType`
  (with nested types in `util/config.go`, `util/config_auth.go`, `util/App.go`, `util/OdbcProvider.go`). Use this
  skill whenever the user asks to "regenerate the config schema", "update config.schema.yaml", "sync the schema
  with ConfigType", "add a new field to the schema", "the schema is out of date", or after a Go struct change
  that affects configuration. Also trigger when the user adds a new top-level field, new nested struct, or
  changes a `json:`, `default:`, `rule:`, or `env:` tag on a config struct field.
---

# Semaphore — Config Schema Generator

This skill produces and maintains `config.schema.yaml` — a **JSON Schema draft 2020-12** document (written in
YAML) describing every valid key in a Semaphore `config.json` / `config.yaml`. The schema is the source of truth
for IDE autocomplete (via `yaml-language-server`) and for any future server-side config validation.

The schema is **derived from Go code**, never hand-edited speculatively. If `ConfigType` doesn't have a field,
the schema must not have it. If a field has a `default:` or `rule:` tag, the schema must reflect it.

## Inputs (source of truth)

All Go types live in the `util` package. Read these files **in full** before writing anything:

| File | What's in it |
|------|--------------|
| `util/config.go` | `ConfigType` (root) + most nested types: `DbConfig`, `TLSConfig`, `RunnerConfig`, `TotpConfig`, `EventLogType`, `TaskLogType`, `ConfigLog`, `SyslogConfig`, `ConfigProcess`, `ConfigAppNamespaces`, `ScheduleConfig`, `DebuggingConfig`, `HAConfig`, `HARedisConfig`, `TeamsConfig`, `ConfigDirs`, `SubscriptionConfig`, `LdapMappings`, `oidcEndpoint` |
| `util/config_auth.go` | `MultifactorAuthConfig`, `EmailAuthConfig`, `TotpConfig`, `RecaptchaConfig` |
| `util/App.go` | `App` |
| `util/OdbcProvider.go` | `OidcProvider` (note: filename is `Odbc...`, typo'd — don't "fix" it) |

If a struct moves or a new file appears, find it with:

```bash
grep -rn "^type .* struct" util/ | grep -v _test
```

## Output

A single file: **`config.schema.yaml`** in the repository root.

Header (top of file) — keep exactly this shape:

```yaml
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "https://semaphoreui.com/schemas/config.schema.json"
title: Semaphore configuration
description: Schema for Semaphore UI config.json / config.yaml. Generated from util.ConfigType.
type: object
```

## Field-mapping rules

These are non-negotiable — apply them mechanically.

### Property name

Use the **`json:` tag's first segment** (before any comma):

```go
Port string `json:"port,omitempty" ...`   // → property name "port"
```

Skip fields where the `json:` tag is `"-"` (e.g. `RunnerConfig.RegistrationToken`, `SubscriptionConfig.Key` —
those are env-only, not file-configurable). Document this in the schema's `description` for the containing
struct if a sensitive token is reachable only via env.

Skip unexported fields (lowercase first letter) — but include their **referenced types** if they appear via
exported fields. Example: `oidcEndpoint` is unexported but referenced from `OidcProvider.Endpoint`, so define
it under `$defs` as `OidcEndpoint`.

### JSON Schema `type`

| Go type | Schema type |
|---------|-------------|
| `string` | `string` |
| `bool` | `boolean` |
| `int`, `int64`, `uint32` | `integer` |
| `float64` | `number` |
| `[]T` | `array`, with `items: { ... }` |
| `map[string]T` | `object`, with `additionalProperties: { ... }` |
| `*T` (pointer) | same as `T`. If `*int`, allow `null`: `type: [integer, "null"]` |
| `*Struct` | `{ $ref: "#/$defs/Struct" }` |

### Constraints from tags

- **`default:"X"`** → emit `default: X` (cast to the right type — `default:"10"` on `int` becomes `default: 10`).
- **`rule:"^...$"`** → emit `pattern: "^...$"`. **Don't** also add a `format:` — the regex is more authoritative.
- **`env:"NAME,sensitive"`** → append `(sensitive).` to the `description`. The schema doesn't have a native
  "sensitive" flag, so this is documentation-only.
- **`env:"NAME"`** without `sensitive` → no change to schema.
- URL-looking string fields (web hooks, redirect URLs) → add `format: uri` **unless** a `pattern` is already
  set from a `rule:` tag.

### `enum` discovery

`enum` is not always in tags — sometimes it's in code:

1. Check for a `rule:"^a|b|c$"` regex — convert that to `enum: [a, b, c]` instead of a pattern (more useful in
   IDE).
2. Check for typed constants near the struct (e.g. `SyslogFormat` has `SyslogDefault = ""` and `SyslogRFC5424
   = "rfc5424"` → `enum: ["", rfc5424]`).
3. `TeamInviteType` constants (`TeamInviteEmail`, `TeamInviteUsername`, `TeamInviteBoth`) → enum.

### Reusable types → `$defs`

Every named struct that appears as a field type gets a `$defs/<TypeName>` entry. Reference via `$ref:
"#/$defs/<TypeName>"`. Top-level `properties` should reference, not inline.

Exception: `lumberjack.Logger` is external — define a minimal stub `$defs/LumberjackLogger` with the common
fields (`filename`, `maxsize`, `maxage`, `maxbackups`, `localtime`, `compress`). Do not try to fully model it.

### `additionalProperties`

Leave it **unset** (defaults to `true`). The schema is informational, not strict — making it strict would
break existing configs that have stray keys (legacy fields, typos used as comments, etc.).

## Workflow

### 1. Read all four Go files

Use `Read` on each file. Do not skim. Note every exported field on every struct.

### 2. Build a mental map: struct → properties

For `ConfigType` (the root), list every exported field with its `json:` name, Go type, and any `default:` /
`rule:` / `env:` tags. Then for each referenced struct, do the same. Stop when no new structs appear.

### 3. Diff against existing `config.schema.yaml`

Before writing, compare what Go has vs. what the schema has:

```bash
# List all json tag names in Go
grep -hoE 'json:"[^,"]+' util/*.go | sed 's/json:"//' | sort -u > /tmp/go-keys.txt

# List all property keys in the schema
yq '.properties | keys | .[]' config.schema.yaml | sort -u > /tmp/schema-keys.txt

diff /tmp/go-keys.txt /tmp/schema-keys.txt
```

Identify three categories:
- **In Go, missing from schema** → add.
- **In schema, missing from Go** → remove (the Go struct is the source of truth).
- **In both** → check that type / default / pattern still match.

Show the user the diff before applying changes if it's larger than ~5 lines.

### 4. Write `config.schema.yaml`

Use the **YAML formatting conventions** already established in the file:

- Section comments split top-level `properties` into visual groups:
  `# === Database ===`, `# === HTTP server ===`, `# === Authentication ===`, etc.
- Inline-object syntax `{ type: string, default: foo }` for one-property entries.
- Multi-line block form when there are 3+ properties or a `description`.
- Quote string values that contain `:`, `/`, `+`, `=`, or start with a digit. Don't quote bare identifiers
  like `string`, `boolean`, `integer`.
- `$ref` values are quoted: `{ $ref: "#/$defs/DbConfig" }`.

### 5. Validate

After writing, run two checks:

```bash
python3 - <<'PY'
import yaml, jsonschema
with open("config.schema.yaml") as f:
    schema = yaml.safe_load(f)
jsonschema.Draft202012Validator.check_schema(schema)  # meta-validation
print("schema is valid")

# Confirm an existing config.yaml still passes
try:
    with open("config.yaml") as f:
        cfg = yaml.safe_load(f)
    errors = list(jsonschema.Draft202012Validator(schema).iter_errors(cfg))
    print(f"config.yaml: {len(errors)} error(s)")
    for e in errors:
        print(f"  {list(e.path)}: {e.message}")
except FileNotFoundError:
    pass
PY
```

Both must pass. If `config.yaml` doesn't exist locally, skip the second check — the meta-validation alone is
sufficient.

### 6. Don't touch `config.yaml`

The user's `config.yaml` may contain stray legacy keys or comments. Do not "clean it up" as part of this
skill — the schema generation is a one-way operation Go → schema. If the user wants a cleaner config, that's a
separate task.

## Common pitfalls

- **`json:"-"` fields**: easy to accidentally include. Always check the tag, not the field name.
- **Pointer vs. value**: `*int` is nullable, plain `int` is not. Reflect that in `type`.
- **Typo'd filenames**: `util/OdbcProvider.go` (should be `Oidc`) — preserve it; renaming files is out of scope.
- **Tag inheritance illusion**: nested structs don't inherit tags from their parent field. The `default:` on
  `ConfigType.MaxParallelTasks` is on the field, not on `int`.
- **`sensitive` is not standard**: JSON Schema has no built-in sensitive marker. Encode it in `description`
  text only.
- **Existing keys with apostrophes** in `config.yaml` (e.g. `"'email_sender"`) are user-side comment hacks —
  ignore them, they're not in any Go struct.

## When to refuse

If the user asks to add a property to the schema that **doesn't exist in `ConfigType`** ("just add a `foo`
field so my config validates"), refuse and explain: the schema is generated from Go. The fix is to add the
field to the Go struct first, then regenerate.

If `ConfigType` has been heavily refactored and a previous schema entry no longer maps, propose the deletion
explicitly before doing it — never silently drop entries that the user might have been relying on.
