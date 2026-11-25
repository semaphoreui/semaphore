// The workflow run panel is addressed by `?w=<workflow id>/<run id>` (`?w=4/10`), as the task
// dialog is by `?t=<task id>`. The two live in the query side by side: opening or closing one
// leaves the other alone.

const RUN_REFERENCE = /^([0-9]+)\/([0-9]+)$/;

// The run the query asks for — `{ workflowId, runId }` — or null when it asks for none (or for
// nonsense).
export function workflowRunFromQuery(query) {
  const raw = (query || {}).w;
  const value = Array.isArray(raw) ? raw[0] : raw;
  const match = RUN_REFERENCE.exec(String(value == null ? '' : value));
  if (!match) {
    return null;
  }
  const workflowId = parseInt(match[1], 10);
  const runId = parseInt(match[2], 10);
  return workflowId > 0 && runId > 0 ? { workflowId, runId } : null;
}

// The current page with the panel open on the run: the target of a link to a run.
export function workflowRunLocation(route, workflowId, runId) {
  return { path: route.path, query: { ...route.query, w: `${workflowId}/${runId}` } };
}

// The query with the panel closed; vue-router drops a parameter that is undefined.
export function queryWithoutWorkflowRun(query) {
  return { ...query, w: undefined };
}
