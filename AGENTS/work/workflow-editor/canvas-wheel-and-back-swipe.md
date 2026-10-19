# Canvas wheel handling and the macOS Back swipe

Facts verified 2026-09-27 in Chromium main sources (`components/input/mouse_wheel_event_queue.cc`,
blink `core/input/mouse_wheel_event_manager.cc`, `platform/widget/input/input_handler_proxy.cc`,
`chrome/browser/renderer_host/chrome_render_widget_host_view_mac_history_swiper.mm`,
`cc/input/input_handler.cc`):

- Blink dispatches DOM wheel events to a latched target chosen at phase Began/MayBegin; the latch
  resets only when the DOM event is cancelled. Cancel every event and targeting follows the cursor.
- MayBegin and Ended/Cancelled phase events never reach the DOM (acked "not consumed", zero delta).
- `send_wheel_events_async_` (rest of gesture non-cancelable) is set only when a kPhaseBegan
  event with non-zero delta was NOT consumed. Momentum events inherit the gesture's mode.
- Compositor delivers a wheel event blocking only if `HasBlockingWheelEventHandlerAt(point)`; a
  `{passive:false}` window listener covers the whole page (verified via CDP: cancelable everywhere).
- History swipe needs: phase Began → kPending, first non-momentum GestureScrollUpdate unconsumed,
  AND `DidOverscroll` with `overscroll_behavior.PropagatesXScroll()`; momentum cannot start it.
  The reported `overscroll_behavior` is the viewport's (html) or `none` when an inner scroller
  with `overscroll-behavior: contain/none` cut the chain.

Outcome: the "Back after scrolling a panel" repro was a Chrome glitch, gone after a browser
restart. The window-level wheel router (`lib/wheelGuard.js`) was removed as over-engineering.
What stays: the canvas's own capture wheel handler plus CSS in `App.vue` — `overscroll-behavior:
none` on html and body while `html.WorkflowEditor-html` is set. Do not reintroduce a window
router without new evidence. Ground truth for the real trackpad must come from the user's
console; testing recipe in `AGENTS/memory/ui-stands.md`.
