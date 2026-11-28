---
form: src/components/WorkflowRunDialog.vue
---

# Run Workflow

Starts a run of the workflow once you have answered its survey. The dialog opens only for a
workflow with a survey: a workflow without one starts at once. After **Run** the page of the new
run opens; when the server refuses the values, the dialog stays open with your answers.

## Variables {#workflowRunDialog-survey}

One field per variable of the survey of the workflow, in the order of the survey, with its
default value filled in. A field marked `*` is required, an Integer field takes digits only, and
a list field takes one value or, when the variable allows several, any number of them. **Run**
starts nothing until every field is valid.

The values are stored with the run and shown on its page under **Variables**. A task of the
workflow does not get them by itself: a variable of its template receives a value only when its
node binds it to a variable of the workflow. The name alone delivers nothing.

A workflow survey holds no secrets, because the run keeps its values in clear. Put passwords and
tokens in a key or a variable group instead.
