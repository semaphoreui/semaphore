---
form: src/components/SurveyVars.vue
---

# Survey Variable

A survey variable is a question the task dialog asks every time you run the template. The
answer reaches the task as a variable: an extra variable for Ansible, a `-var` for Terraform and
OpenTofu, an argument for a script, or an environment variable. The variable is saved with the
template, when you save the template dialog. The same dialog edits the survey of a workflow,
where the type **Secret** is not offered.

## Name {#surveyVarDialog-name}

The name the task gets the answer under: the name of the extra variable, of the Terraform
variable, of the `name=value` argument of a script, or of the environment variable. Spaces at
the ends are dropped. The field is required.

Use a name no other variable of the survey has, and one the tool accepts, for example
`app_version`. An answer replaces a variable of the same name from the variable group of the
template.

## Title {#surveyVarDialog-title}

The label of the field in the task dialog, for example `Version to deploy`. It is also the text
of the chip of the variable in the template dialog. The field is required.

## Description {#surveyVarDialog-description}

Optional. A hint shown under the field of the task dialog while it has the focus: what to
enter, or an example.

## Type {#surveyVarDialog-type}

What kind of answer the task dialog asks for:

- **String**: one line of text. This is the default.
- **Integer**: digits only; the task dialog refuses anything else, a minus sign too. The answer
  still reaches the task as text: convert it in the playbook, for example with `| int`.
- **Secret**: text the task dialog masks. The answer is not saved with the task, so a re-run
  asks for it again, and Ansible gets it in a vault-encrypted file. A secret has no default
  value.
- **Enum**: one value picked from a list you define below.
- **Text**: several lines of text.
- **Select**: several values picked from a list you define below. The task gets them as a list;
  in an argument, a `-var` or an environment variable the list is written as JSON, for example
  `["a","b"]`.

Changing the type clears the default value, and a type without a list clears the values.

## Pass variable as {#surveyVarDialog-target}

How the answer reaches the task:

- **Extra variable**: the usual way of the tool. Ansible gets it as an extra variable,
  Terraform and OpenTofu as `-var name=value`, a Bash, PowerShell or Python script as a
  `name=value` argument after the script.
- **Environment variable**: an environment variable of the process, `name=value`, for every
  kind of template. The variable is then not passed the other way.

## Key {#surveyVarDialog-valueName}

Shown when the type is **Enum** or **Select**.

The text the task dialog shows for this value in the list. Every value needs a key, and the
keys must differ: the dialog refuses to save the variable otherwise.

## Value {#surveyVarDialog-valueValue}

Shown when the type is **Enum** or **Select**.

What the task gets when this value is picked. It may differ from its key: the key `Production`
can pass `prod`.

## Add Value {#surveyVarDialog-addValue}

Shown when the type is **Enum** or **Select**.

Adds an empty row to the list of values. An **Enum** or a **Select** needs at least one value.
The bin icon of a row removes it.

## Default value {#surveyVarDialog-defaultValue}

Shown for every type but **Secret**.

The answer the task dialog fills in when it opens; you can change it there. For an **Enum** pick
one of the values, for a **Select** any number of them. Leave the field empty to start with an
empty answer.

The default is filled in by the task dialog only. A task started another way, through the API
for example, gets no default.

## Required {#surveyVarDialog-required}

The task dialog does not start a task until the variable has an answer. Like the default, the
check is made by the task dialog only, not by the server.
