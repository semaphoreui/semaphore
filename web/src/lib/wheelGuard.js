// Wheel events nobody on the page can consume.
//
// The workflow editor page does not scroll (the editor fills the viewport and
// <html> has overflow hidden), so a trackpad gesture that no inner scroller
// takes reaches the browser as overscroll. On macOS a horizontal overscroll is
// Back / Forward navigation. Browsers decide that from the first wheel event
// of a gesture, and Chrome makes the rest of the gesture non-cancelable when
// that first event was not cancelled, so the guard has to run on every wheel
// event, wherever the gesture starts: the canvas, a side panel, the toolbar,
// the navigation drawer.

const SCROLLABLE_OVERFLOW = ['auto', 'scroll'];

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

/**
 * Cancels a wheel event that no element on the page would scroll, so it does
 * not turn into browser overscroll navigation. Register on window in the
 * capture phase with { passive: false } while the page is non-scrollable.
 */
export function cancelUnconsumedWheel(ev) {
  if (!ev.cancelable || ev.defaultPrevented) return false;
  if (wheelConsumer(ev.target, ev.deltaX, ev.deltaY)) return false;
  ev.preventDefault();
  return true;
}
