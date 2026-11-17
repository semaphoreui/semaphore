---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

A matcher is a condition on the incoming request. Matchers decide which integration a request
to the alias of the project starts, so one URL can serve several integrations: an integration
runs only when every one of its matchers holds. They are used only when **Use project alias** is
on for the integration; with it on, the integration is reached only through the project alias.

An integration found by the project alias with no matchers never runs: add at least one.

## Name {#integrationMatcherForm-name}

The name the matcher is listed under on the page of the integration. It does not change what the
matcher checks. The field is required.

## Match on {#integrationMatcherForm-matchType}

The part of the request the matcher reads:

- **Header**: the value of a request header.
- **Body**: the body of the request.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Shown when **Match on** is **Body**.

How the body is read:

- **JSON**: the body is parsed as JSON, and **Key** is the path to one value in it.
- **String**: the whole body is compared as text, and **Key** is not used.

## Key {#integrationMatcherForm-key}

What to read from the request. The field is required.

- For **Header**: the name of the header, for example `X-GitHub-Event`. The case does not
  matter.
- For **Body** in **JSON**: the path to the value, its parts joined by dots; an item of an array
  is written as `.[0]`, for example `repository.full_name` or `commits.[0].id`.
- For **Body** in **String**: not used, but the form needs a value; type anything.

## Comparison Method {#integrationMatcherForm-method}

How the value read from the request is compared with **Value**:

- **==**: the two are equal.
- **!=**: the two differ. A header or a key the request lacks counts as different.
- **Contains**: the value read from the request contains **Value**.

## Value {#integrationMatcherForm-value}

The text to compare with, for example `push` or `refs/heads/main`. The comparison is exact and
case-sensitive. A number or a boolean of the JSON body is compared in its text form: `42`,
`true`. The field is required.
