---
form: src/components/EditViewsForm.vue
---

# Edit Views

Views are the tabs above the list of task templates. A view shows the templates you put into it,
or every template of the project. The dialog saves every change at once: there is no Save button
for the whole dialog. Drag a view by its handle to change the order of the tabs.

## Title {#viewsForm-title}

The name of the tab. Type it and confirm it with the check mark next to the field; the cross
puts back the title that is saved. A new view is saved once it has a title and you confirm it
or change one of its settings. Closing the dialog does not save a title you have not confirmed.

The gear opens the settings of the view below its title.

## Delete {#viewsForm-viewDelete}

The bin icon deletes the view at once, without asking. The templates of the view are not
deleted: they stay in the project and are in no view any more.

## Type {#viewsForm-type}

Which templates the tab shows:

- **Custom**: the templates whose **View** field, in the template dialog, names this view. This
  is the default.
- **All**: every template of the project, whatever its view.

A member sees only the templates the member may see, in any view.

## Sort by {#viewsForm-sortColumn}

The column the templates of the tab are sorted by; **Name** is the only one. A tab without it
is sorted by name as well, so the field matters together with **Reverse**.

## Reverse {#viewsForm-sortReverse}

Sorts the tab in descending order. It works only when **Sort by** is set.

## Hidden {#viewsForm-hidden}

Hides the tab from members who can not manage the resources of the project. Those who can still
see it, struck through. Hiding a tab does not hide its templates: a view of the type **All**
still lists them.

## Show task counts {#viewsForm-showTemplateStats}

Shows on the tab how many of its templates are in each state, as one chip of coloured segments:
green for a last task that succeeded, red for a failed one, grey for a stopped one or a template
that was never launched, orange for a task that waits for confirmation, blue for a queued or running
one. It is off by default, and the counts are computed only for the views that have it on.

## Add view {#viewsForm-add}

Adds a view at the end of the list, with an empty title to type in.
