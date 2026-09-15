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

/** Escapes a value for safe insertion into element content. */
export function escapeHtml(value) {
  return String(value == null ? '' : value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}
