"""Browser helpers of the suite: page setup, lookups by data-testid, states, dialogs, API
interception, masks and the snapshot.

Everything that depends on the DOM of Vuetify lives here (VUETIFY_* selectors, field_message,
FieldInput, the freeze CSS): the selectors hold the Vuetify 2 and the Vuetify 3 form side by side, the
DOM of the page picks the one that matches — the same suite runs before and after the upgrade.
"""

import datetime as dt
import json
import os
import re
import time
from contextlib import contextmanager
from dataclasses import dataclass, field
from pathlib import Path
from urllib.parse import urlsplit

from playwright.sync_api import Locator, Page, Route, expect

SIZES = {
    "mobile": {"width": 390, "height": 844},  # Vuetify xs in 2 and 4
    "tablet": {"width": 768, "height": 1024},  # sm
    "desktop": {"width": 1440, "height": 900},  # lg
}

# ---- Vuetify-dependent selectors: change here, nowhere else -------------------------------------
# Strings with {tid} / {text} are formatted by the helpers below; regexes go to to_have_class().
# Every selector holds the Vuetify 2 form and the Vuetify 3 form side by side (a CSS union / an
# alternation): only one of them can match a given page, so the DOM itself picks the version.
#   Vuetify 2: the mark of a field sits on its <input>; the field root is the .v-input around it.
#   Vuetify 3: the mark (any data-* attribute) goes to the root .v-input, the <input> is inside.


def _field(inner):
    """Selector of <inner> of the field marked {tid}, in Vuetify 2 and in Vuetify 3."""
    return f'.v-input:has([data-testid="{{tid}}"]) {inner}, .v-input[data-testid="{{tid}}"] {inner}'


# the shown messages (validation, hint) of a field: Vuetify 3 keys a message by its text, so a new
# message enters while the old one leaves (2 elements for a frame or two — a strict-mode error); the
# leaving one is left out. Vuetify 2 keys them by index and patches the text in place.
VUETIFY_FIELD_MESSAGE = _field('.v-messages__message:not([class*="-leave-active"])')
VUETIFY_SELECT_SELECTION = (  # the shown choice: 2 — v-select; 3 — v-select, v-autocomplete, v-combobox
    _field(".v-select__selection")
    + ', .v-input[data-testid="{tid}"] .v-autocomplete__selection'
    + ', .v-input[data-testid="{tid}"] .v-combobox__selection')
VUETIFY_FIELD_SLOT = (  # the clickable box of a field: 2 — .v-input__slot, 3 — .v-field
    '.v-input:has([data-testid="{tid}"]) .v-input__slot, .v-input[data-testid="{tid}"] .v-field')
VUETIFY_APPEND_BUTTON = (  # the password eye: 2 — a <button> in .v-input__append-inner; 3 — a clickable
    # icon (role=button) of `append-icon` (.v-input__append, beside the field) or `append-inner-icon`
    '.v-input:has([data-testid="{tid}"]) .v-input__append-inner button'
    ', .v-input[data-testid="{tid}"] .v-input__append [role="button"]'
    ', .v-input[data-testid="{tid}"] .v-field__append-inner [role="button"]:not([class*="__menu-icon"])')
VUETIFY_CLEAR_BUTTON = (  # clearable: 2 — a <button>; 3 — a clickable icon «Clear <label>»
    '.v-input:has([data-testid="{tid}"]) .v-input__icon--clear button'
    ', .v-input[data-testid="{tid}"] .v-field__clearable [role="button"]')
# the open dialog window = its visible box: 2 — .v-dialog--active (also the scroll container);
# 3 — the .v-overlay__content of the active v-dialog (the card inside it scrolls, a fullscreen one —
# the box itself)
VUETIFY_ACTIVE_DIALOG = ".v-dialog--active, .v-overlay--active.v-dialog > .v-overlay__content"
# the open menu of a v-menu / v-select: 2 — .menuable__content__active (tooltips too); 3 — the content
# of the active v-menu overlay (select lists are v-menu too)
VUETIFY_ACTIVE_MENU = ".menuable__content__active, .v-overlay--active.v-menu > .v-overlay__content"
VUETIFY_TOOLTIP = (".v-tooltip__content.menuable__content__active"
                   ", .v-overlay--active.v-tooltip > .v-overlay__content")  # the open tooltip
VUETIFY_CHIP_CLOSE = '.v-chip:has-text("{text}") .v-chip__close'  # the cross of a selection chip (2 and 3)
# the expand button of a show-expand row: 2 — .v-data-table__expand-icon; 3 — the v-btn of the
# data-table-expand cell
VUETIFY_EXPAND_ICON = ".v-data-table__expand-icon, .v-data-table__td--expanded-row button"
VUETIFY_BADGE_DOT = ".v-badge__badge"  # 2 and 3
VUETIFY_DARK_APP = ".v-application.theme--dark, .v-application.v-theme--dark"  # the app in the dark theme
VUETIFY_ACTIVE_ITEM = re.compile(r"\bv-list-item--active\b")  # the current item of a v-list / nav (2 and 3)
VUETIFY_BTN_ACTIVE = re.compile(r"\bv-btn--active\b")  # the chosen button of a v-btn-toggle (2 and 3)
VUETIFY_BTN_ACTIVE_CSS = ".v-btn--active"
VUETIFY_BTN_LOADING = re.compile(r"\bv-btn--loading\b")  # 2 and 3
# a disabled v-tab: 2 — only the class v-tab--disabled; 3 — a v-tab is a v-btn: v-btn--disabled
VUETIFY_TAB_DISABLED = re.compile(r"\bv-(tab|btn)--disabled\b")
# a v-chip coloured primary: 2 — class primary; 3 — bg-primary (flat / elevated) or text-primary (tonal,
# the default variant of v-chip)
VUETIFY_CHIP_ACTIVE = re.compile(r"(^|\s)((bg|text)-)?primary(\s|$)")
# a Vuetify 3 <input> that lets clicks through (v-select: pointer-events none — the field box under it
# takes the click); Vuetify 2 has none
VUETIFY_PASS_THROUGH_JS = "(e) => getComputedStyle(e).pointerEvents === 'none'"

# CodeMirror 5 (vue-codemirror, Vue 2) or CodeMirror 6 (vue-codemirror6, Vue 3)
CM_EDITOR = '[data-testid="{tid}"] .CodeMirror, [data-testid="{tid}"] .cm-content'  # the editable area
# its own blinking cursor (caret="hide" does not hide it): a mask
CM_CURSOR = ".CodeMirror-cursors, .cm-cursorLayer"

# v-alert dismissible: the cross is the last button of the alert (2 and 3)
ALERT_CLOSE = '[data-testid="{tid}"] button'

# Drawflow (workflow canvas)
DRAWFLOW_NODE = ".drawflow-node"
DRAWFLOW_SELECTED = ".drawflow-node.selected"
DRAWFLOW_INPUT = ".input"
DRAWFLOW_OUTPUT = ".output"
DRAWFLOW_PRECANVAS = ".drawflow"
DRAWFLOW_EDGE_PATH = "svg.connection path.main-path"

FREEZE_CSS = """
*, *::before, *::after {
  transition: none !important; transition-duration: 0s !important; transition-delay: 0s !important;
  animation-duration: 0s !important; animation-delay: 0s !important; animation-iteration-count: 1 !important;
  caret-color: transparent !important; scroll-behavior: auto !important;
}
.v-ripple__container { display: none !important; }  /* the ripple of Vuetify 2 and 3 */
/* the error overlay of the dev server (vue-cli-service serve, the hot-reload UI — never in a build): Vue 3
   in a dev build throws the errors of hooks, and the overlay over the page takes every click */
#webpack-dev-server-client-overlay { display: none !important; }
"""

INIT_JS_TEMPLATE = """(() => {
  try { %(lang)s localStorage.removeItem('darkMode'); } catch (e) {}
  const css = %(css)s;
  const add = () => {
    if (document.getElementById('vu-freeze')) return;
    const s = document.createElement('style');
    s.id = 'vu-freeze';
    s.textContent = css;
    (document.head || document.documentElement).appendChild(s);
  };
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', add); else add();
})();"""


def init_js(lang="en"):
    """Init script of every context: freeze CSS, no dark mode; localStorage.lang = <lang> on every
    load, or left alone when <lang> is None (the app picks the language itself, a switch survives a
    reload)."""
    set_lang = f"localStorage.setItem('lang', {json.dumps(lang)});" if lang else ""
    return INIT_JS_TEMPLATE % {"lang": set_lang, "css": json.dumps(FREEZE_CSS)}


INIT_JS = init_js("en")

EXTERNAL = re.compile(r"^(https?|wss?)://(?!(localhost|127\.0\.0\.1)(:\d+)?/)")


@dataclass
class PageInfo:
    """What the suite keeps on every page it opens (page.vu)."""
    size: str
    role: str
    origin: str
    out_dir: Path
    shots: set
    read_only: bool
    console_errors: list = field(default_factory=list)
    http_errors: list = field(default_factory=list)
    writes_blocked: list = field(default_factory=list)


def strip_origin(text, origin):
    return text.replace(origin, "") if origin else text


def attach_recorders(page: Page, info: PageInfo):
    def on_console(msg):
        if msg.type == "error":
            info.console_errors.append(strip_origin(msg.text, info.origin))

    def on_pageerror(err):
        # «<name>: <message>» as console.error prints an error: Vue 2, and Vue 3 in a production build,
        # log an error of a hook to the console; Vue 3 in a dev build (the hot-reload UI) throws it, and it
        # arrives here with the message alone (no «AxiosError»)
        name, text = getattr(err, "name", "") or "", str(err)
        if name and not text.startswith(name):
            text = f"{name}: {text}"
        info.console_errors.append("pageerror: " + strip_origin(text, info.origin))

    def on_response(resp):
        if resp.status >= 400:
            info.http_errors.append(f"{resp.request.method} {strip_origin(resp.url, info.origin)} {resp.status}")

    page.on("console", on_console)
    page.on("pageerror", on_pageerror)
    page.on("response", on_response)


def guard_writes(context, info: PageInfo):
    """Snapshot tests are read-only: a POST/PUT/DELETE to the API that no page.route() of the test
    answered is aborted and fails the test. page.route() handlers take precedence over this one."""
    def handler(route: Route):
        req = route.request
        if req.method in ("GET", "HEAD", "OPTIONS"):
            route.fallback()
            return
        info.writes_blocked.append(f"{req.method} {strip_origin(req.url, info.origin)}")
        route.abort()

    context.route("**/api/**", handler)


# ---- lookups ------------------------------------------------------------------------------------

def tid(scope, test_id) -> Locator:
    """Element by data-testid inside a page or a locator."""
    return scope.get_by_test_id(test_id)


def shown(scope, test_id) -> Locator:
    """The visible element(s) <test_id>: v-dialog and RichEditor keep closed copies in the DOM, so a
    mark of a dialog opened twice exists twice."""
    return tid(scope, test_id).filter(visible=True)


def wait_state(scope, test_id, state, timeout=15000) -> Locator:
    """Wait until [data-testid=<test_id>] is visible with data-state=<state>."""
    loc = tid(scope, test_id)
    expect(loc).to_have_attribute("data-state", state, timeout=timeout)
    return loc


def field_message(scope, test_id) -> Locator:
    """Validation / hint messages of the field marked <test_id> (the mark sits on the input)."""
    return scope.locator(VUETIFY_FIELD_MESSAGE.format(tid=test_id))


class FieldInput(Locator):
    """The <input> of a field (field_input). A plain Locator but for click(): the <input> of a Vuetify 3
    v-select lets clicks through (pointer-events: none) to the field box under it, so Playwright would
    wait for it to «receive events» for good. Such an input is clicked at its place without that check
    once it is visible — the click lands on the field box and opens the list, as a click on the input
    of a Vuetify 2 v-select does. Every other input is clicked as usual."""

    def click(self, **kwargs):
        if not kwargs.get("force") and self.evaluate(VUETIFY_PASS_THROUGH_JS, timeout=kwargs.get("timeout")):
            expect(self).to_be_visible(**({"timeout": kwargs["timeout"]} if "timeout" in kwargs else {}))
            expect(self).to_be_enabled(**({"timeout": kwargs["timeout"]} if "timeout" in kwargs else {}))
            kwargs["force"] = True
        return super().click(**kwargs)


def field_input(scope, test_id) -> Locator:
    """The <input> or <textarea> of a field marked <test_id>: Vuetify 2 puts the mark on the input
    itself, Vuetify 3 on the root .v-input around it. A Vuetify 2 autocomplete shows its selection here
    (to_have_value); a Vuetify 3 one only while focused (blurred — field_selection). A v-textarea is
    found too. click() on the input of a Vuetify 3 v-select opens its list (FieldInput)."""
    t = f'[data-testid="{test_id}"]'
    loc = scope.locator(f"input{t}, textarea{t}, {t} input, {t} textarea").first
    return FieldInput(loc._impl_obj)


def field_selection(scope, test_id) -> Locator:
    """The shown selection(s) of a v-select marked <test_id>; in Vuetify 3 also of a v-autocomplete /
    v-combobox (its <input> holds the text only while focused). A Vuetify 2 autocomplete shows the
    selection in its input: field_input(...).to_have_value."""
    return scope.locator(VUETIFY_SELECT_SELECTION.format(tid=test_id))


def field_append(scope, test_id) -> Locator:
    """The button inside the field on the right (the password «eye», «<label> appended action»)."""
    return scope.locator(VUETIFY_APPEND_BUTTON.format(tid=test_id))


def field_clear(scope, test_id) -> Locator:
    """The clear button of a clearable select / autocomplete / text field («Clear <label>»)."""
    return scope.locator(VUETIFY_CLEAR_BUTTON.format(tid=test_id))


def toggle(scope, test_id, on=True):
    """Check / uncheck a checkbox or switch marked <test_id> (its hidden <input> — Vuetify 2: the mark is
    on it; Vuetify 3: inside the marked .v-input; hence force)."""
    box = field_input(scope, test_id)
    box.set_checked(on, force=True)
    if on:
        expect(box).to_be_checked()
    else:
        expect(box).not_to_be_checked()
    return box


def open_menu(scope, test_id):
    """Open the list of the select / autocomplete / combobox marked <test_id>: a click on the free right
    end of the field box (Vuetify 2 .v-input__slot, 3 .v-field), left of the arrow. Works when the
    <input> is 0 px high (multiple + chips, nothing chosen) and does not hit a selection chip in the
    middle."""
    slot = scope.locator(VUETIFY_FIELD_SLOT.format(tid=test_id)).first
    expect(slot).to_be_visible()
    box = slot.bounding_box()
    slot.click(position={"x": max(box["width"] - 40, 1), "y": box["height"] / 2})


def option(page: Page, name) -> Locator:
    """A visible option of the open menu by its exact text (menu items have no marks)."""
    return page.get_by_role("option", name=name, exact=True)


def pick(page: Page, scope, test_id, name):
    """Open the select / autocomplete <test_id> inside <scope> and choose the option <name>."""
    open_menu(scope, test_id)
    expect(option(page, name)).to_be_visible()
    option(page, name).click()


def active_menu(page: Page) -> Locator:
    """The open v-menu / v-select list (no mark of its own; Vuetify 3: the content of the active menu
    overlay, teleported to <body>); prefer menu_of() when an item has a mark."""
    return page.locator(VUETIFY_ACTIVE_MENU).last


def menu_of(page: Page, item_test_id) -> Locator:
    """The open menu that holds the item <item_test_id>: Vuetify 2 — the ARIA role=menu; Vuetify 3 gives
    a v-menu no role — the content of the active menu overlay (VUETIFY_ACTIVE_MENU)."""
    menu = page.locator(f'[role="menu"], {VUETIFY_ACTIVE_MENU}').filter(has=tid(page, item_test_id)).last
    expect(menu).to_be_visible()
    return menu


def tooltip(page: Page) -> Locator:
    """The open tooltip."""
    return page.locator(VUETIFY_TOOLTIP)


def chip_close(chip: Locator) -> Locator:
    """The close cross of a chip: a button named «Close», no mark (Vuetify 2 and 3)."""
    return chip.get_by_role("button", name="Close")


def chip_close_by_text(scope, text) -> Locator:
    """The close cross of the selection chip whose text is <text> (chips of a select have no mark)."""
    return scope.locator(VUETIFY_CHIP_CLOSE.format(text=text)).first


def expand_button(row_loc: Locator) -> Locator:
    """The expand button of a v-data-table row with show-expand (first cell on desktop, its own line
    in the mobile rows); it carries no mark."""
    return row_loc.locator(VUETIFY_EXPAND_ICON).first


def badge_dot(scope) -> Locator:
    """The dot / counter of a v-badge inside <scope> (a locator of the marked badge)."""
    return scope.locator(VUETIFY_BADGE_DOT)


def alert_close(scope, test_id) -> Locator:
    """The cross of a dismissible v-alert marked <test_id>."""
    return scope.locator(ALERT_CLOSE.format(tid=test_id)).last


def active_in_group(scope, group_test_id) -> Locator:
    """The chosen button(s) of a v-btn-toggle marked <group_test_id> (Vuetify 2 sets no aria-pressed)."""
    return tid(scope, group_test_id).locator(VUETIFY_BTN_ACTIVE_CSS)


def expect_active(loc: Locator, active=True):
    """A v-list-item / nav item is (not) the current one."""
    (expect(loc).to_have_class if active else expect(loc).not_to_have_class)(VUETIFY_ACTIVE_ITEM)


def expect_btn_active(loc: Locator, active=True):
    """A button of a v-btn-toggle is (not) chosen."""
    (expect(loc).to_have_class if active else expect(loc).not_to_have_class)(VUETIFY_BTN_ACTIVE)


def expect_loading(loc: Locator, loading=True):
    """A v-btn is (not) in its loading state."""
    (expect(loc).to_have_class if loading else expect(loc).not_to_have_class)(VUETIFY_BTN_LOADING)


def expect_tab_disabled(loc: Locator, disabled=True):
    """A v-tab is (not) disabled (Vuetify 2: only a class shows it; Vuetify 3: a v-tab is a v-btn —
    v-btn--disabled)."""
    (expect(loc).to_have_class if disabled else expect(loc).not_to_have_class)(VUETIFY_TAB_DISABLED)


def expect_chip_active(loc: Locator, active=True):
    """A v-chip is (not) coloured primary — the «selected» look of filter chips."""
    (expect(loc).to_have_class if active else expect(loc).not_to_have_class)(VUETIFY_CHIP_ACTIVE)


def editor(scope, test_id) -> Locator:
    """The editable area of the code editor marked <test_id> (CodeMirror 5 on Vue 2, CodeMirror 6 —
    .cm-content — on Vue 3)."""
    return scope.locator(CM_EDITOR.format(tid=test_id)).last


def editor_type(page: Page, test_id, text):
    """Replace the whole text of the code editor <test_id> by typing <text> (neither CodeMirror 5 nor the
    minimal CodeMirror 6 setup of lib/codeEditor.js closes brackets); blur() afterwards hides its own
    cursor."""
    editor(page, test_id).click()
    page.keyboard.press("Control+A")
    page.keyboard.type(text)


def blur(page: Page):
    """Take the focus off the focused element (CodeMirror's cursor, a field's focus ring)."""
    page.evaluate("() => document.activeElement && document.activeElement.blur()")


def press(loc: Locator, event=False):
    """Click <loc>; when its centre lies left or right of the window (a mobile toolbar running past the
    right edge, a tab outside a clipped slide group) or <event> is set, dispatch the click event instead of a
    mouse click."""
    expect(loc).to_be_visible()
    if not event:
        box = loc.bounding_box()
        event = bool(box) and not 0 <= box["x"] + box["width"] / 2 <= loc.page.viewport_size["width"]
    if event:
        loc.dispatch_event("click")
    else:
        loc.click()


def select_tab(tab: Locator):
    """Select a v-tab by the click event (on 390 px the tabs overflow the slide group, and a mouse
    click after Playwright's scroll misses), then wait for aria-selected."""
    tab.dispatch_event("click")
    expect(tab).to_have_attribute("aria-selected", "true")


def dialog(page, test_id, timeout=10000) -> Locator:
    """The card of an open dialog (EditDialog / YesNoDialog test-id), once visible and drawn at its own
    size (the enter transition scales it for a frame: a click then lands on the overlay)."""
    card = tid(page, test_id)
    expect(card).to_be_visible(timeout=timeout)
    wait_unscaled(page, card)
    return card


def wait_unscaled(page: Page, loc: Locator, timeout_ms=2000):
    """Wait until <loc> is drawn at its layout size (no scale transform left); best effort."""
    deadline = time.time() + timeout_ms / 1000
    while time.time() < deadline:
        try:
            if loc.evaluate(UNSCALED_JS):
                return True
        except Exception:  # noqa: BLE001 — detached / re-rendered: look again
            pass
        page.wait_for_timeout(20)
    return False


UNSCALED_JS = """(el) => { const r = el.getBoundingClientRect();
  return r.width > 0 && Math.abs(r.width - el.offsetWidth) < 1 && Math.abs(r.height - el.offsetHeight) < 1; }"""


def active_dialog(page) -> Locator:
    """The open dialog that has no mark of its own (prefer dialog(page, test_id))."""
    return page.locator(VUETIFY_ACTIVE_DIALOG).last


def dialog_frame(page, card_test_id) -> Locator:
    """The window of the open dialog holding the card <card_test_id> — its scroll container. A card
    taller than the window (mobile) shot as an element takes the page around its clipped part; shoot
    the frame instead (snap(element=dialog_frame(...))), scrolling it for the lower parts."""
    frame = page.locator(VUETIFY_ACTIVE_DIALOG).filter(has=tid(page, card_test_id)).last
    expect(frame).to_be_visible()
    return frame


def open_dialog(page, trigger, test_id) -> Locator:
    """Click <trigger> (a test id or a locator) and return the dialog card <test_id>."""
    (tid(page, trigger) if isinstance(trigger, str) else trigger).click()
    return dialog(page, test_id)


def button(card: Locator, test_id) -> Locator:
    """A button inside the dialog card: editDialog-save, yesNoDialog-yes, … are shared by all
    dialogs, so they are looked up only inside the card."""
    return tid(card, test_id)


def row(page, table_test_id, text, exact=False) -> Locator:
    """Table row containing <text> (cells without slots carry no marks — NOTE «ячейки без слота»).
    has_text is a case-insensitive substring ('Empty' matches 'Z15 sync empty'); exact=True keeps rows
    with an element whose whole text is <text>."""
    rows = tid(page, table_test_id).locator("tbody tr")
    if exact:
        return rows.filter(has=page.get_by_text(text, exact=True))
    return rows.filter(has_text=text)


def open_sidebar(page, toggle_test_id) -> Locator:
    """mobile / tablet (< 960 px): the sidebar is temporary and closed — open it with the screen's
    <p>-drawerToggle and wait until it is fully in the window. desktop: it is always open."""
    sidebar = tid(page, "sidebar")
    if page.vu.size != "desktop":
        tid(page, toggle_test_id).click()
    expect(sidebar).to_be_in_viewport(ratio=1)
    return sidebar


# ---- snackbar (App.vue) -------------------------------------------------------------------------
# Vuetify 2 keeps the root of v-snackbar (app-snackbar) in the DOM and «visible» to Playwright even
# when closed, with data-state=ready; Vuetify 3 renders the root (an overlay) only from the first
# opening on and keeps it after closing. Either way open / closed is told by its close button
# app-snackbarClose.

SNACKBAR_OPAQUE_JS = """() => {
  const root = document.querySelector('[data-testid="app-snackbar"]');
  let el = document.querySelector('[data-testid="app-snackbarClose"]');
  if (!root || !el || !root.contains(el)) return false;
  for (; el && el !== root.parentElement; el = el.parentElement)
    if (getComputedStyle(el).opacity !== '1') return false;
  return true;
}"""


def snackbar(page: Page) -> Locator:
    return tid(page, "app-snackbar")


def expect_snackbar(page: Page, text=None, state=None, timeout=15000) -> Locator:
    """The snackbar is shown (with <text>, in data-state <state>: ready | error)."""
    bar = snackbar(page)
    expect(tid(page, "app-snackbarClose")).to_be_visible(timeout=timeout)
    if state:
        expect(bar).to_have_attribute("data-state", state)
    if text:
        expect(bar).to_contain_text(text)
    return bar


def expect_no_snackbar(page: Page):
    """The snackbar is closed (to_be_hidden() on its root never passes in Vuetify 2; in Vuetify 3 the root
    may be missing altogether)."""
    expect(tid(page, "app-snackbarClose")).to_be_hidden()


def close_snackbar(page: Page, text=None):
    """Check the shown snackbar (<text>) and close it; fine if its own 3 s timer closed it first."""
    expect_snackbar(page, text)
    try:
        tid(page, "app-snackbarClose").click(timeout=3000)
    except Exception:  # noqa: BLE001 — closed by its timer meanwhile
        pass
    expect_no_snackbar(page)


def hold_snackbar(page: Page, clock, text=None, state=None, masks=()) -> Locator:
    """Wait for the snackbar (<text>, <state>), fully opaque, then stop the page timers
    (stop_timers) so its 3 s timeout cannot close it during the shot. Nothing that needs timers or
    requestAnimationFrame may run afterwards (wait_until(..., settle=False))."""
    bar = expect_snackbar(page, text, state)
    page.wait_for_function(SNACKBAR_OPAQUE_JS)
    stop_timers(page, clock, masks)
    expect(tid(page, "app-snackbarClose")).to_be_visible()
    return bar


def stop_timers(page: Page, clock, masks=(), attempts=20):
    """Pause the page clock (polls, the snackbar timeout, live timers) at <clock> + 1 s — once the page
    shows the same frame twice in a row: paused timers also stop requestAnimationFrame, so a slider or
    a transition caught half-way would stay so. Step on with page.clock.run_for(ms) for a fixed value."""
    page.mouse.move(-1, -1)
    settle_tabs(page)
    opts = dict(animations="disabled", caret="hide", mask=_mask_locators(page, masks))
    prev = page.screenshot(**opts)
    for _ in range(attempts):
        page.wait_for_timeout(150)
        cur = page.screenshot(**opts)
        if cur == prev:
            break
        prev = cur
    else:
        raise AssertionError("the page did not settle before the timers were stopped")
    page.clock.pause_at(clock + dt.timedelta(seconds=1))


# ---- API interception ---------------------------------------------------------------------------
# Every helper takes <method> ('GET', 'POST', … or None for any) and a <pattern>: a glob
# ('**/api/project/*/keys'), a compiled regex (searched in the full URL — use it when the URL has a
# query: '?' of a glob is not a literal) or a predicate(url). Other methods fall through to the next
# handler (the write guard of snapshot tests, then the network). <skip>: the first <skip> matching
# calls go on untouched (answer the 2nd GET only: skip=1).
# Handlers never call page methods (page.evaluate, page.wait_*): a handler runs inside the page's
# event loop, and a page call from it blocks the sync API for good (the test hangs at 100 % CPU).

class Hits(list):
    """Requests a helper answered (playwright Request objects, in order)."""

    @property
    def urls(self):
        return [r.url for r in self]

    @property
    def bodies(self):
        """The JSON bodies sent (None for a request without one)."""
        out = []
        for r in self:
            try:
                out.append(r.post_data_json)
            except Exception:  # noqa: BLE001 — not JSON
                out.append(r.post_data)
        return out


def _closed(e):
    text = str(e)
    return any(w in text for w in ("closed", "disposed", "Target page", "Target crashed"))


def _route(page: Page, method, pattern, action, skip=0):
    """page.route(<pattern>) with the method filter and <skip>; action(route) answers. An error from a
    closed page (the test ended while a request was on the way) is swallowed. Returns the handler."""
    seen = [0]

    def handler(route: Route):
        req = route.request
        if method and req.method != method.upper():
            route.fallback()
            return
        seen[0] += 1
        if seen[0] <= skip:
            route.fallback()
            return
        try:
            action(route)
        except Exception as e:  # noqa: BLE001
            if not _closed(e):
                raise

    page.route(pattern, handler)
    return handler


def _json_body(body):
    return "" if body is None else json.dumps(body)


def answer_api(page: Page, method, pattern, status=200, body=None, skip=0) -> Hits:
    """Answer <method> <pattern> with <status> and the JSON <body> (None — an empty body, as a 204 of
    the server; a callable(request) — computed per request, e.g. to change it during the test). The
    request never reaches the server. Returns the hits (hits.bodies — what the page sent)."""
    hits = Hits()

    def action(route: Route):
        hits.append(route.request)
        data = body(route.request) if callable(body) else body
        route.fulfill(status=status, content_type="application/json", body=_json_body(data))

    _route(page, method, pattern, action, skip)
    return hits


def fail_api(page: Page, method, path_glob, status=500, body=None, skip=0):
    """Answer <method> requests to <path_glob> (e.g. '**/api/project/*/host_configs') with an error.
    <body> None → {"error": "vu <status> <method>"}; an empty dict {} is sent as is (the page then shows
    axios' own «Request failed with status code N»). Other methods go on. Returns the list of hit URLs
    (answer_api returns the requests themselves)."""
    hits = []
    payload = json.dumps(body if body is not None else {"error": f"vu {status} {method}"})

    def action(route: Route):
        hits.append(route.request.url)
        route.fulfill(status=status, content_type="application/json", body=payload)

    _route(page, method, path_glob, action, skip)
    return hits


def abort_api(page: Page, method, pattern, skip=0) -> Hits:
    """Drop <method> <pattern> as a network error (axios: «Network Error»). Returns the hits."""
    hits = Hits()

    def action(route: Route):
        hits.append(route.request)
        route.abort()

    _route(page, method, pattern, action, skip)
    return hits


def patch_api(page: Page, pattern, change, method="GET", skip=0) -> Hits:
    """Answer <method> <pattern> with the server's real JSON answer passed through <change>: a
    callable(data) returning the new data (or None after changing it in place), or a dict merged into
    an object answer. A non-2xx or non-JSON answer goes through unchanged. Use it for GETs — a write
    would reach the server. Returns the hits."""
    hits = Hits()

    def action(route: Route):
        hits.append(route.request)
        resp = route.fetch()
        if not 200 <= resp.status < 300:
            route.fulfill(response=resp)
            return
        try:
            data = resp.json()
        except Exception:  # noqa: BLE001 — not JSON
            route.fulfill(response=resp)
            return
        if callable(change):
            new = change(data)
            data = data if new is None else new
        else:
            data = {**data, **change}
        route.fulfill(response=resp, json=data)

    _route(page, method, pattern, action, skip)
    return hits


def only_names(page: Page, pattern, names, key="name") -> Hits:
    """GET <pattern> (a list) cut down to the items whose <key> is in <names> — e.g. the base set
    (seed.base.BASE_NAMES) plus the zone's own objects, whatever other zones seeded."""
    keep = set(names)
    return patch_api(page, pattern, lambda items: [i for i in items if i.get(key) in keep])


def patch_info(page: Page, **features) -> Hits:
    """GET /api/info with systemInfo.features changed (e.g. secret_storages=False, ha=True)."""
    def change(data):
        data.setdefault("features", {}).update(features)

    return patch_api(page, "**/api/info", change)


class Held(list):
    """Routes kept without an answer by hold_api (the loading / saving state stands still)."""

    def __init__(self, page: Page, pattern):
        super().__init__()
        self.page, self.pattern, self.handler, self.done = page, pattern, None, False

    def wait(self, count=1, timeout=10000):
        """Wait until <count> requests are held (the page really sent them)."""
        wait_until(self.page, lambda: len(self) >= count, f"{count} held request(s) {self.pattern}",
                   timeout=timeout, settle=False)
        return self

    def _finish(self, act):
        self.done = True  # requests arriving from now on fall through (guard / network), never held
        for route in list(self):
            try:
                act(route)
            except Exception:  # noqa: BLE001 — the page may have dropped the request already
                pass
        self.clear()
        try:
            self.page.unroute(self.pattern, self.handler)
        except Exception:  # noqa: BLE001 — the page is closed
            pass

    def release(self, status=500, body=None):
        """Answer every held request with <status> / <body> (default {"error": "vu released"}), then
        drop the handler. Answer first: unroute() of a still held request sends it to the server — a
        held POST would create the object."""
        payload = json.dumps(body if body is not None else {"error": "vu released"})
        self._finish(lambda r: r.fulfill(status=status, content_type="application/json", body=payload))

    def proceed(self):
        """Let every held request go on (to the next handler, then the server), drop the handler. In a
        snapshot test a held write then meets the write guard and fails the test."""
        self._finish(lambda r: r.fallback())

    def drop(self):
        """Abort every held request (network error), drop the handler."""
        self._finish(lambda r: r.abort())


def hold_api(page: Page, method, pattern, skip=0) -> Held:
    """Keep <method> <pattern> unanswered until held.release(...) / .proceed() / .drop(). Returns the
    held routes (a list: `wait_until(page, lambda: held, …)` or held.wait())."""
    held = Held(page, pattern)

    def action(route: Route):
        if held.done:
            route.fallback()
        else:
            held.append(route)

    held.handler = _route(page, method, pattern, action, skip)
    return held


def _matcher(match):
    if match is None:
        return lambda url: True
    if isinstance(match, re.Pattern):
        return lambda url: bool(match.search(url))
    if callable(match):
        return match
    return lambda url: url.split("?", 1)[0].endswith(match)


def requests_of(page: Page, method=None, match=None) -> Hits:
    """Requests the page sends from now on (also those answered by a route): <method> (None — any);
    <match>: a string — the end of the URL path (query ignored), a regex — searched in the URL, or a
    predicate(url). len() of it checks that a request was (not) sent."""
    out = Hits()
    ok = _matcher(match)
    page.on("request", lambda r: out.append(r) if (method is None or r.method == method.upper()) and ok(r.url)
            else None)
    return out


def responses_of(page: Page, method=None, match=None) -> list:
    """Responses from now on, like requests_of (r.status, r.request.method)."""
    out = []
    ok = _matcher(match)
    page.on("response", lambda r: out.append(r)
            if (method is None or r.request.method == method.upper()) and ok(r.url) else None)
    return out


def wait_console(page: Page, text, count=1, timeout=10000):
    """Wait for <count> console errors with <text>: an unhandled axios rejection is logged a moment
    after the failed response, and the .json of a shot must not depend on that moment."""
    wait_until(page, lambda: sum(text in e for e in page.vu.console_errors) >= count,
               f"{count} console error(s) «{text}»", timeout=timeout)


def mock_socket(page: Page) -> list:
    """The app's websocket /api/ws answered by the test, not the server (no update / log messages of
    other tasks); returns the WebSocketRoute objects once open (ws.send(json.dumps(...)))."""
    sockets = []
    page.route_web_socket(re.compile(r"/api/ws$"), lambda ws: sockets.append(ws))
    return sockets


# ---- navigation ---------------------------------------------------------------------------------

def goto(page: Page, path, ready_test_id=None, state=None):
    """Open <path> of the stand and wait for the app and, if given, a screen root in <state>; then for
    the web fonts and one window resize event, so the v-tabs slider and slide group are measured with
    the real font whatever the timing of the font load."""
    page.goto(path)
    expect(page.get_by_test_id("app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    if ready_test_id:
        loc = tid(page, ready_test_id)
        if state:
            expect(loc).to_have_attribute("data-state", state, timeout=15000)
        else:
            expect(loc).to_be_visible(timeout=15000)
    settle_tabs(page)


def settle_tabs(page: Page, widen=False):
    """Web fonts loaded, then a window resize event: v-tabs measures its slider (and v-slide-group its
    scroll offset) at mount, with the fallback font when Roboto is late, and again only on a resize.
    widen=True also widens the window within the same breakpoint and back (the offset of tabs that
    overflow on mobile is then computed from 0 — z03 / Settings)."""
    page.evaluate("() => document.fonts.ready.then(() => { window.dispatchEvent(new Event('resize')); })")
    if widen:
        size = page.viewport_size
        page.set_viewport_size({"width": WIDEN_TO[page.vu.size], "height": size["height"]})
        page.wait_for_timeout(100)
        page.set_viewport_size(size)
        page.wait_for_timeout(100)


WIDEN_TO = {"mobile": 580, "tablet": 940, "desktop": 1441}  # same Vuetify breakpoint as SIZES


APPS_RE = re.compile(r"/api/apps$")
TASKS_LAST_RE = re.compile(r"/api/project/\d+/tasks/last(\?.*)?$")


def goto_after_apps(page: Page, path, pattern, ready_test_id=None, state=None, change=None):
    """goto() with GET <pattern> held until every GET /api/apps has finished. Screens that draw app
    icons (History, Templates, Inventory, terraform state) render rows with getAppIcon, which throws
    while the apps are on the way (JOURNAL@236D70B238): a TypeError in the console of some runs only.
    <change>: also rewrite the held answer (patch_api); later calls go on as usual."""
    started, finished = [], []
    page.on("request", lambda r: started.append(r) if APPS_RE.search(r.url) else None)
    page.on("requestfinished", lambda r: finished.append(r) if APPS_RE.search(r.url) else None)
    page.on("requestfailed", lambda r: finished.append(r) if APPS_RE.search(r.url) else None)
    if change is not None:
        patch_api(page, pattern, change)  # registered first → asked after the hold: proceed() falls to it
    held = hold_api(page, "GET", pattern)
    page.goto(path)
    wait_until(page, lambda: held and started and len(finished) >= len(started),
               "GET /api/apps before the held list", timeout=20000, settle=False)
    held.proceed()
    expect(page.get_by_test_id("app-page")).to_have_attribute("data-state", "ready", timeout=20000)
    if ready_test_id:
        loc = tid(page, ready_test_id)
        if state:
            expect(loc).to_have_attribute("data-state", state, timeout=15000)
        else:
            expect(loc).to_be_visible(timeout=15000)
    settle_tabs(page)


# Counts the GET /api/apps the document sends and finishes (XHR — axios). Counted in the page, not by
# Playwright events: with routes on, a «request» event may reach Python after the screen shows «ready».
APPS_COUNT_JS = r"""(() => {
  if (window.__vuApps) return;
  const s = window.__vuApps = {started: 0, done: 0};
  const open = XMLHttpRequest.prototype.open, send = XMLHttpRequest.prototype.send;
  XMLHttpRequest.prototype.open = function (method, url, ...rest) {
    this.__vuApps = String(method).toUpperCase() === 'GET' && /\/api\/apps$/.test(new URL(url, location.href).pathname);
    return open.call(this, method, url, ...rest);
  };
  XMLHttpRequest.prototype.send = function (...args) {
    if (this.__vuApps) { s.started++; this.addEventListener('loadend', () => { s.done++; }); }
    return send.apply(this, args);
  };
})();"""
APPS_DONE_JS = "() => !!window.__vuApps && window.__vuApps.started > 0 && window.__vuApps.done === window.__vuApps.started"


def goto_apps_loaded(page: Page, path, ready_test_id=None, state=None):
    """goto(), then wait until every GET /api/apps the screen sent has finished (ok or not). For screens
    whose controls need the apps: the template screen renders EditTemplateDialog only once its own GET
    /api/apps is in, and «Edit» / «Copy» pressed before that open nothing for good (EditDialog takes
    v-model only on a change) — a missing dialog in some loaded runs. The screen must send at least one
    GET /api/apps once it is <state>."""
    page.add_init_script(APPS_COUNT_JS)  # the guard in the script makes a second call harmless
    goto(page, path, ready_test_id, state)
    page.wait_for_function(APPS_DONE_JS, timeout=20000)


def open_history(page: Page, project_id, state="ready", query="", last_task=None):
    """/project/<id>/history with the task list after /api/apps (goto_after_apps). <last_task>: cut the
    list to tasks with id ≤ it (seed.base.BASE_LAST_TASK — the base tasks #1–#5 of Seed Demo)."""
    change = None if last_task is None else (lambda items: [t for t in items if t.get("id", 0) <= last_task])
    goto_after_apps(page, f"/project/{project_id}/history{query}", TASKS_LAST_RE, "history-page", state, change)


@contextmanager
def full_load(page: Page):
    """Expect a full page load (document.location = …), not a router change. Yields a list that holds
    the path (with query) of the loaded document after the block."""
    out = []
    with page.expect_request(lambda r: r.is_navigation_request() and r.frame == page.main_frame) as req:
        yield out
    u = urlsplit(req.value.url)
    out.append(u.path + (f"?{u.query}" if u.query else ""))


def created_before(items, clock):
    """Items created up to the fixture <clock> (meta['clock']): drops what functional tests added on a
    shared debug stand. On a restored fixture it keeps everything."""
    limit = clock.strftime("%Y-%m-%dT%H:%M:%S") if hasattr(clock, "strftime") else str(clock)[:19]
    return [i for i in items if (i.get("created") or "")[:19] <= limit]


# ---- snapshot -----------------------------------------------------------------------------------

def norm(part):
    """camelCase → kebab, lower case, latin letters, digits and single hyphens."""
    s = re.sub(r"([a-z0-9])([A-Z])", r"\1-\2", str(part))
    s = re.sub(r"[^a-z0-9]+", "-", s.lower()).strip("-")
    if not s:
        raise ValueError(f"empty name part from {part!r}")
    return s


ZONE_CODE = re.compile(r"^(z\d{2}|b\d+|base)$")


def shot_name(zone, screen, state, size):
    """<zone>__<screen>__<state>__<size>; <zone> is the zone code (z19, b1; base for the suite's own
    checks) — the first segment pairs the shots of before and after."""
    z = norm(zone)
    if not ZONE_CODE.match(z):
        raise ValueError(f"shot zone {zone!r}: use the zone code (z19, b1) or 'base', not a name")
    return f"{z}__{norm(screen)}__{norm(state)}__{size}"


def _check_zone_of_test(zone):
    """A test in tests/zones/test_<code>.py shoots as <code>; tests/common/ as base."""
    current = os.environ.get("PYTEST_CURRENT_TEST", "")
    m = re.search(r"tests/zones/test_([a-z0-9]+)\.py", current)
    expected = m.group(1) if m else ("base" if "tests/common/" in current else None)
    if expected and norm(zone) != expected:
        raise ValueError(f"shot zone {zone!r} in {current.split('::')[0]}: expected {expected!r}")


def _mask_locators(page, masks):
    out = []
    for m in masks or ():
        if isinstance(m, Locator):
            out.append(m)
        elif re.fullmatch(r"[A-Za-z][\w-]*", m):  # a bare word is a data-testid
            out.append(page.get_by_test_id(m))
        else:  # anything else is a Playwright / CSS selector
            out.append(page.locator(m))
    return out


BUNDLE = re.compile(r"(/js/[\w~-]+?)\.[0-9a-f]{6,}(\.js)(:\d+:\d+)?")
STACK_LINE = re.compile(r"\n\s+at .*")


def normalize_console(text):
    """A console line without what changes with every build: stack frames and bundle hashes and
    positions (/js/app.3e316d57.js:1:529784 → /js/app.js)."""
    return BUNDLE.sub(r"\1\2", STACK_LINE.sub("", text))


INVENTORY_JS = """(root) => {
  const visible = (el) => !!(el.offsetWidth || el.offsetHeight || el.getClientRects().length)
    && getComputedStyle(el).visibility !== 'hidden';
  const ids = {};
  const states = {};
  for (const el of root.querySelectorAll('[data-testid]')) {
    if (!visible(el)) continue;
    const t = el.getAttribute('data-testid');
    ids[t] = (ids[t] || 0) + 1;
    const s = el.getAttribute('data-state');
    if (s !== null) (states[t] = states[t] || []).push(s);
  }
  const tables = [];
  for (const t of root.querySelectorAll('table')) {
    if (!visible(t)) continue;
    const owner = t.closest('[data-testid]');
    tables.push({ testid: owner ? owner.getAttribute('data-testid') : null,
                  rows: t.querySelectorAll('tbody tr').length });
  }
  return { testids: ids, states, tables };
}"""


def snap(page: Page, zone, screen, state, element=None, masks=(), full_page=True, attempts=10,
         hover=False, pad=0):
    """Screenshot <zone>__<screen>__<state>__<size>.png into runtime/visual/<label>/, with
    .aria.yml (accessibility tree) and .json (visible test ids and states, table rows, console
    errors, 4xx/5xx answers). <zone> — the zone code (z19, b1). <element>: a test id or a locator —
    shoot only it (a dialog card; a card taller than its window — dialog_frame()). <hover>: keep the
    mouse where it is (tooltips, hover states). <pad>: px around <element> (a corner sticking out).
    Stable means two identical frames in a row; otherwise the test fails."""
    info: PageInfo = page.vu
    _check_zone_of_test(zone)
    name = shot_name(zone, screen, state, info.size)
    if name in info.shots:
        raise AssertionError(f"shot {name} taken twice in this run — give the state another name")
    info.shots.add(name)

    target = tid(page, element) if isinstance(element, str) else element
    if target is not None:
        expect(target).to_be_visible(timeout=10000)
    if not hover:
        page.mouse.move(-1, -1)
    page.evaluate("document.fonts.ready.then(() => true)")
    opts = dict(animations="disabled", caret="hide", mask=_mask_locators(page, masks), mask_color="#FF00FF")
    if target is not None and pad:
        target.scroll_into_view_if_needed()
        box = target.bounding_box()
        sx, sy = page.evaluate("() => [window.scrollX, window.scrollY]")
        x, y = max(box["x"] + sx - pad, 0), max(box["y"] + sy - pad, 0)
        opts["clip"] = {"x": x, "y": y, "width": box["x"] + sx + box["width"] + pad - x,
                        "height": box["y"] + sy + box["height"] + pad - y}
        opts["full_page"] = True  # clip in page coordinates, past the window too
        shoot = lambda: page.screenshot(**opts)  # noqa: E731
    elif target is not None:
        shoot = lambda: target.screenshot(**opts)  # noqa: E731
    else:
        opts["full_page"] = full_page
        shoot = lambda: page.screenshot(**opts)  # noqa: E731

    prev = shoot()
    for _ in range(attempts):
        page.wait_for_timeout(150)
        cur = shoot()
        if cur == prev:
            break
        prev = cur
    else:
        raise AssertionError(f"shot {name}: no two identical frames in {attempts + 1} tries — "
                             "wait for a data-state, mask the changing part or stop the timer")

    out = info.out_dir
    (out / f"{name}.png").write_bytes(cur)
    scope = target if target is not None else page.locator("body")
    (out / f"{name}.aria.yml").write_text(strip_origin(scope.aria_snapshot(), info.origin) + "\n")
    inv = scope.evaluate(INVENTORY_JS)
    inv.update({
        "shot": name, "role": info.role, "size": info.size,
        "path": strip_origin(page.url, info.origin),
        "element": element if isinstance(element, str) else (None if element is None else "locator"),
        "masks": [m if isinstance(m, str) else "locator" for m in masks or ()],
        "console_errors": [normalize_console(e) for e in info.console_errors],
        "http_errors": list(info.http_errors),
    })
    if pad:
        inv["pad"] = pad
    (out / f"{name}.json").write_text(json.dumps(inv, indent=2, sort_keys=True) + "\n")
    return out / f"{name}.png"


def wait_until(page: Page, check, what, timeout=10000, settle=True):
    """Poll a Python condition (e.g. the hits list of fail_api) while the page keeps running.
    settle=False after page.clock.pause_at (settle needs requestAnimationFrame, paused with the clock)."""
    deadline = time.time() + timeout / 1000
    while not check():
        if time.time() > deadline:
            raise AssertionError(f"timed out waiting for {what}")
        page.wait_for_timeout(50)
    if settle:
        _two_frames(page)


def _two_frames(page: Page):
    page.evaluate("() => new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)))")


def settle(page: Page, ms=0):
    """Let the page run its pending work (timers, rendering) without depending on wall time.
    Prefer wait_state(); this is for a step that has no observable state."""
    _two_frames(page)
    if ms:
        time.sleep(ms / 1000)
