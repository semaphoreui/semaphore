// Wheel routing for a page that does not scroll itself.
//
// The workflow editor fills the viewport and <html> has overflow hidden, so a
// trackpad gesture that no inner scroller takes reaches the browser as
// overscroll, which macOS turns into Back / Forward navigation. Two Chrome
// behaviours make a plain per-element wheel handler insufficient:
//
// - Target latching: every wheel event of a gesture, including the inertia
//   tail, is dispatched to the element under the cursor when the gesture
//   started. Moving from the navigation drawer onto the canvas mid-gesture
//   keeps sending events to the drawer, so the canvas never sees them.
// - Async wheel: when the first wheel event of a gesture is not cancelled, the
//   rest of the gesture is dispatched non-cancelable and preventDefault() no
//   longer stops the browser from scrolling (and overscrolling).
//
// So while the editor is open a single window-level capture listener cancels
// every wheel event and dispatches it by the *current* pointer position:
// over the graph it pans / zooms, over a scrollable element it scrolls that
// element, anywhere else it is dropped.

const SCROLLABLE_OVERFLOW = ['auto', 'scroll'];
const LINE_HEIGHT = 16;

function canScroll(el, axis, delta) {
  const style = window.getComputedStyle(el);
  const overflow = axis === 'x' ? style.overflowX : style.overflowY;
  if (!SCROLLABLE_OVERFLOW.includes(overflow)) return false;
  const size = axis === 'x' ? el.clientWidth : el.clientHeight;
  const scrollSize = axis === 'x' ? el.scrollWidth : el.scrollHeight;
  const pos = axis === 'x' ? el.scrollLeft : el.scrollTop;
  if (scrollSize <= size) return false;
  return delta < 0 ? pos > 0 : pos + size < scrollSize - 1;
}

/**
 * Returns the closest ancestor of `target` (inclusive) that can scroll in the
 * dominant direction of the wheel delta, or null when nothing can.
 */
export function wheelConsumer(target, deltaX, deltaY) {
  const horizontal = Math.abs(deltaX) > Math.abs(deltaY);
  const axis = horizontal ? 'x' : 'y';
  const delta = horizontal ? deltaX : deltaY;
  if (delta === 0) return null;
  let el = target instanceof Element ? target : null;
  while (el && el !== document.documentElement) {
    if (canScroll(el, axis, delta)) return el;
    el = el.parentElement;
  }
  return null;
}

// Wheel delta in CSS pixels (Firefox reports lines for a mouse wheel).
export function wheelPixels(ev, axis) {
  const delta = axis === 'x' ? ev.deltaX : ev.deltaY;
  switch (ev.deltaMode) {
    case 1: return delta * LINE_HEIGHT;
    case 2: return delta * (axis === 'x' ? window.innerWidth : window.innerHeight);
    default: return delta;
  }
}

export function scrollByWheel(el, ev) {
  el.scrollBy(wheelPixels(ev, 'x'), wheelPixels(ev, 'y'));
}

/**
 * Handles one wheel event for a non-scrolling page. `isGraph(el)` tells
 * whether the element under the pointer belongs to the graph, `onGraph(ev)`
 * pans / zooms it. Returns what was done: 'ignored' (non-cancelable event,
 * the gesture started before the router was installed), 'graph', 'scrolled'
 * or 'dropped'.
 */
export function routeWheel(ev, { isGraph, onGraph }) {
  if (!ev.cancelable) return 'ignored';
  ev.preventDefault();
  const el = document.elementFromPoint(ev.clientX, ev.clientY);
  if (el && isGraph(el)) {
    // The graph's own listener must not handle the same event again.
    ev.stopImmediatePropagation();
    onGraph(ev);
    return 'graph';
  }
  const scroller = wheelConsumer(el, ev.deltaX, ev.deltaY);
  if (scroller) {
    scrollByWheel(scroller, ev);
    return 'scrolled';
  }
  return 'dropped';
}
