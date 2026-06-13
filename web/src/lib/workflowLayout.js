// Layered topological layout for workflow graphs.
//
// Used to position nodes on the canvas when a workflow has no stored
// coordinates (legacy workflows created before the graphical editor, or runs of
// such workflows). Nodes are placed in columns by their longest distance from a
// root (zero in-degree) node, and stacked vertically within each column.
//
// Returns a map: { [nodeId]: { x, y } }.

const COLUMN_WIDTH = 240;
const ROW_HEIGHT = 130;
const OFFSET_X = 80;
const OFFSET_Y = 60;

// True when every node sits at the origin, i.e. no real layout is stored yet.
export function needsAutoLayout(nodes) {
  return (nodes || []).length > 0
    && nodes.every((n) => !n.position_x && !n.position_y);
}

export function layoutWorkflowNodes(nodes, edges) {
  const incoming = {};
  const adjacency = {};
  nodes.forEach((n) => {
    incoming[n.id] = 0;
    adjacency[n.id] = [];
  });
  (edges || []).forEach((e) => {
    if (incoming[e.destination_node_id] != null) incoming[e.destination_node_id] += 1;
    if (adjacency[e.source_node_id]) adjacency[e.source_node_id].push(e.destination_node_id);
  });

  const depth = {};
  const pending = { ...incoming };
  const queue = nodes.filter((n) => !incoming[n.id]).map((n) => n.id);
  queue.forEach((id) => { depth[id] = 0; });

  while (queue.length) {
    const id = queue.shift();
    (adjacency[id] || []).forEach((next) => {
      depth[next] = Math.max(depth[next] || 0, (depth[id] || 0) + 1);
      pending[next] -= 1;
      if (pending[next] <= 0) queue.push(next);
    });
  }

  const perColumn = {};
  const positions = {};
  nodes.forEach((n) => {
    const col = depth[n.id] || 0;
    const row = perColumn[col] || 0;
    positions[n.id] = {
      x: OFFSET_X + col * COLUMN_WIDTH,
      y: OFFSET_Y + row * ROW_HEIGHT,
    };
    perColumn[col] = row + 1;
  });
  return positions;
}
