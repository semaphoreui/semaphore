---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

An extracted value takes a piece of the incoming request, such as the branch of a push event,
and passes it to the task the integration starts. Every value of the integration is taken from
each request that starts a task.

## Name {#integrationValueForm-name}

The name the value is listed under on the page of the integration. It does not change what is
taken or where it goes. The field is required.

## Source of the Value {#integrationValueForm-valueSource}

The part of the request the value is taken from:

- **Header**: the value of a request header.
- **Body**: the body of the request.

## Data Type of Body {#integrationValueForm-bodyDataType}

Shown when **Source of the Value** is **Body**.

How the body is read:

- **JSON**: the body is parsed as JSON, and **Key** is the path to one value in it. The value
  keeps its JSON type: a number, a list or an object reach the task as such.
- **String**: the whole body is taken as text, and **Key** is not used.

## Key {#integrationValueForm-key}

What to take from the request. The field is required.

- For **Header**: the name of the header, for example `X-GitHub-Event`. The case does not
  matter. A header the request lacks gives an empty value.
- For **Body** in **JSON**: the path to the value, its parts joined by dots; an item of an array
  is written as `.[0]`, for example `repository.full_name` or `commits.[0].id`. A path the body
  lacks gives no value, and the task keeps the one set in the integration.
- For **Body** in **String**: not used, but the form needs a value; type anything.

## Variable Usage {#integrationValueForm-variableType}

Where the value goes in a task of a template:

- **Variables**: an extra variable of the task, the way the answer to a survey variable is
  passed. It replaces an answer or a variable of the variable group with the same name.
- **Task Params**: a parameter of the task, the way the options of the task dialog are stored.
  The name `git_branch` sets the branch of the task when the template allows to override the
  branch; a leading `refs/heads/` is removed.

An integration that runs a workflow ignores the usage: every value becomes the answer to the
survey variable of the workflow named in **Variable**.

## Variable {#integrationValueForm-variable}

The name the value is passed under, for example `branch` for an extra variable that the playbook
reads as `{{ branch }}`. The field is required.
