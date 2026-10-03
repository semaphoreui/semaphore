// Pure viewport math for the workflow canvas (WorkflowGraph.vue).
//
// A viewport is { x, y, zoom }: the canvas is rendered with
// `transform: translate(x, y) scale(zoom)` and transform-origin 0 0, so a
// canvas point c maps to the screen point x + c * zoom.

export const GRID_SIZE = 20;
export const ZOOM_MIN = 0.25;
export const ZOOM_MAX = 2;
export const ZOOM_STEP = 1.2;
export const FIT_PADDING = 48;
export const DEFAULT_NODE_SIZE = { width: 220, height: 64 };

export function clampZoom(zoom, min = ZOOM_MIN, max = ZOOM_MAX) {
  if (!Number.isFinite(zoom)) return min;
  return Math.min(max, Math.max(min, zoom));
}

/** Rounds a coordinate to the nearest grid line. */
export function snapToGrid(value, grid = GRID_SIZE) {
  return Math.round(value / grid) * grid;
}

/**
 * Bounding box of the nodes ({ id, x, y }) using their rendered sizes
 * ({ [id]: { width, height } }); nodes without a size use DEFAULT_NODE_SIZE.
 * Returns null for an empty graph.
 */
export function graphBounds(nodes, sizes = {}) {
  if (!nodes || nodes.length === 0) return null;
  let minX = Infinity;
  let minY = Infinity;
  let maxX = -Infinity;
  let maxY = -Infinity;
  nodes.forEach((n) => {
    const size = sizes[n.id] || DEFAULT_NODE_SIZE;
    minX = Math.min(minX, n.x);
    minY = Math.min(minY, n.y);
    maxX = Math.max(maxX, n.x + size.width);
    maxY = Math.max(maxY, n.y + size.height);
  });
  return {
    x: minX, y: minY, width: maxX - minX, height: maxY - minY,
  };
}

/**
 * Viewport that shows the whole `bounds` centered in a container of
 * { width, height }. Never zooms in past `maxZoom` (1 by default): a small
 * graph is centered at its natural size, not blown up.
 */
export function fitViewport(bounds, container, options = {}) {
  const { padding = FIT_PADDING, minZoom = ZOOM_MIN, maxZoom = 1 } = options;
  if (!bounds || !container || container.width <= 0 || container.height <= 0) {
    return { x: 0, y: 0, zoom: 1 };
  }
  const availableWidth = Math.max(1, container.width - padding * 2);
  const availableHeight = Math.max(1, container.height - padding * 2);
  const zoom = clampZoom(
    Math.min(
      availableWidth / Math.max(bounds.width, 1),
      availableHeight / Math.max(bounds.height, 1),
    ),
    minZoom,
    maxZoom,
  );
  return {
    x: (container.width - bounds.width * zoom) / 2 - bounds.x * zoom,
    y: (container.height - bounds.height * zoom) / 2 - bounds.y * zoom,
    zoom,
  };
}

/**
 * Changes the zoom so that the canvas point under `point` (container
 * coordinates) stays under it.
 */
export function zoomAt(viewport, point, nextZoom, options = {}) {
  const zoom = clampZoom(nextZoom, options.minZoom, options.maxZoom);
  const canvasX = (point.x - viewport.x) / viewport.zoom;
  const canvasY = (point.y - viewport.y) / viewport.zoom;
  return {
    x: point.x - canvasX * zoom,
    y: point.y - canvasY * zoom,
    zoom,
  };
}

export function panBy(viewport, dx, dy) {
  return { ...viewport, x: viewport.x + dx, y: viewport.y + dy };
}

/** Container coordinates -> canvas coordinates. */
export function screenToCanvas(viewport, point) {
  return {
    x: (point.x - viewport.x) / viewport.zoom,
    y: (point.y - viewport.y) / viewport.zoom,
  };
}

/** Multiplicative zoom factor for a wheel event's deltaY (pixels). */
export function wheelZoomFactor(deltaY) {
  return Math.exp(-deltaY * 0.002);
}
