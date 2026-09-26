// Pure helpers for the workflow graph editor (WorkflowGraph.vue).

/**
 * True when adding the edge source -> dest would close a cycle in the graph
 * described by `edges` ({ source_node_id, destination_node_id }).
 */
export function wouldCreateCycle(edges, source, dest) {
  if (source === dest) {
    return true;
  }

  // Walk forward from `dest` over existing edges; a path back to `source`
  // means the new edge source->dest closes a cycle.
  const adjacency = {};
  (edges || []).forEach((edge) => {
    if (!adjacency[edge.source_node_id]) adjacency[edge.source_node_id] = [];
    adjacency[edge.source_node_id].push(edge.destination_node_id);
  });

  const stack = [dest];
  const seen = new Set();
  while (stack.length) {
    const cur = stack.pop();
    if (cur === source) return true;
    if (!seen.has(cur)) {
      seen.add(cur);
      (adjacency[cur] || []).forEach((n) => stack.push(n));
    }
  }
  return false;
}

/** Returns max(ids) + 1, or 1 for an empty graph. Non-numeric ids count as 0. */
export function nextNodeId(ids) {
  const numeric = (ids || []).map((id) => Number(id) || 0);
  return (numeric.length === 0 ? 0 : Math.max(...numeric)) + 1;
}

/** Formats a delay in milliseconds as "Ns" or "Mm Ns", rounding seconds up. */
export function formatDuration(ms) {
  const totalSeconds = Math.ceil(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes <= 0) return `${seconds}s`;
  return `${minutes}m ${seconds}s`;
}

/** Like formatDuration, but folds full hours: "1h 2m", "3m 4s", "5s". */
export function formatDurationLong(ms) {
  const totalSeconds = Math.max(0, Math.ceil(ms / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes}m ${seconds}s`;
  return `${seconds}s`;
}

/** Escapes a value for safe insertion into element content. */
export function escapeHtml(value) {
  return String(value == null ? '' : value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

/**
 * Collapses the raw status of a task / approval / delay into the handful of
 * visual states the node card knows how to draw.
 */
export function statusKind(status) {
  switch (status) {
    case 'success':
    case 'approved':
    case 'confirmed':
      return 'success';
    case 'error':
    case 'failed':
    case 'stopped':
    case 'rejected':
      return 'error';
    case 'running':
    case 'starting':
    case 'stopping':
      return 'running';
    case 'waiting':
    case 'waiting_confirmation':
      return 'waiting';
    case 'pending':
      return 'pending';
    default:
      return null;
  }
}

export function isFinishedStatus(status) {
  const kind = statusKind(status);
  return kind === 'success' || kind === 'error';
}

/** True when an edge with `condition` is taken after the source ends with `status`. */
export function edgeConditionMet(condition, status) {
  const kind = statusKind(status);
  if (kind !== 'success' && kind !== 'error') return false;
  if (condition === 'always') return true;
  if (condition === 'on_failure') return kind === 'error';
  return kind === 'success';
}

/**
 * Visual state of an edge in a run view, given the run info of the nodes
 * ({ [nodeId]: { status } }):
 *   'active' — the source is done and the destination is in progress;
 *   'passed' — both ends have run;
 *   'dim'    — the destination has not started (or the edge was not taken).
 */
export function edgeRunState(edge, runs) {
  const source = runs && runs[edge.source_node_id];
  const dest = runs && runs[edge.destination_node_id];
  const sourceStatus = source ? source.status : null;
  const destStatus = dest ? dest.status : null;
  const destKind = statusKind(destStatus);

  if (!isFinishedStatus(sourceStatus) || !edgeConditionMet(edge.condition, sourceStatus)) {
    return 'dim';
  }
  if (destKind === 'running' || destKind === 'waiting' || destKind === 'pending') {
    return 'active';
  }
  if (destKind === 'success' || destKind === 'error') {
    return 'passed';
  }
  return 'dim';
}
