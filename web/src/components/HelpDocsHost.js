// The built-in help of a dialog or a page: the state of the help panel (HelpPanel.vue) and its
// tie to the elements the document describes. Mixed into the component that hosts the panel,
// EditDialog for a dialog and HelpPage for a page.
//
// A section of a document is anchored to an element by its `data-testid`: a field of a form or
// a button. The host knows which element is the active one (the one in focus, the one whose
// section is in focus, or the one whose section is under the pointer), marks it, tells the panel
// to scroll to the section of the element in focus, and lists the elements that are not there.
//
// The host provides:
//   - `docsKey`: the key of the document (src/help), as a prop or a computed property;
//   - `docsRootElement()`: the element the described elements are looked up in;
//   - `docsCanOpen()`: optional, whether the panel may open once its document is loaded.
// and in its template:
//   - `@focusin="onDocsHostFocusIn($event)"` on the element that holds the described elements;
//   - `<HelpPanel v-if="docsShown" v-bind="docsPanelProps" @close="closeDocs()"
//       @hover-field="docsHoverField = $event" @focus-field="onDocsFocusField($event)" />`.
import EventBus from '@/event-bus';
import { hasHelpDoc, helpDocLocale, loadHelpDoc } from '@/help';

// The attribute of the element the active section of the panel describes (styled in
// HelpPanel.vue). An attribute, not a class: Vue rewrites the class of a Vuetify field on every
// change of its state (focused, dirty), which would drop a class added from outside, while it
// leaves alone an attribute it does not render.
const DOCS_FIELD_ATTR = 'data-help-docs-field';

// Whether an element of the host is on screen: not inside an element hidden with
// `display: none`, as a tab of `v-window` the user has left, which stays in the page. The
// browser answers with `checkVisibility`; where it lacks it (Safari before 17.4, jsdom) the
// ancestors are walked up to the root of the host.
function isShown(el, root) {
  if (typeof el.checkVisibility === 'function') {
    return el.checkVisibility();
  }
  const view = el.ownerDocument.defaultView;
  for (let node = el; node && node !== root.parentElement; node = node.parentElement) {
    if (view.getComputedStyle(node).display === 'none') {
      return false;
    }
  }
  return true;
}

// The width of the help panel (HelpPanel.vue).
export const DOCS_WIDTH = 420;

// The narrowest dialog, and the margin Vuetify gives a dialog on each side.
export const DIALOG_MIN_WIDTH = 400;
const DIALOG_MARGIN = 24;

export default {
  data() {
    return {
      // The help panel: whether it is open, and the document it shows (loaded on first open).
      docsOpen: false,
      docsDoc: null,
      // The element the help panel is about: the one in focus (an element of the host, or the
      // section of the panel that describes it), unless the pointer is over a section of the
      // panel. `docsFocusSource` is 'form' or 'panel': where the focus is.
      docsFocusField: null,
      docsFocusSource: null,
      docsHoverField: null,
      // Elements the document describes that the host does not show now.
      docsMissingFields: [],
    };
  },

  computed: {
    hasDocs() {
      return hasHelpDoc(this.docsKey);
    },

    // The language the document is shown in: the one of the interface, as far as the document
    // is translated (src/help). Without vue-i18n (unit tests) the document is English.
    docsLocale() {
      const i18n = this.$i18n;
      return helpDocLocale(this.docsKey, i18n ? i18n.locale : 'en');
    },

    docsShown() {
      return this.docsOpen && this.docsDoc != null;
    },

    docsActiveField() {
      return this.docsHoverField ?? this.docsFocusField;
    },

    // The element the panel scrolls to: only an element of the host in focus moves the panel,
    // not a section of the panel itself.
    docsFollowField() {
      return this.docsFocusSource === 'form' ? this.docsFocusField : null;
    },

    // A dialog host: the narrowest dialog and the panel fit side by side; in a narrower window
    // the panel lies over the dialog.
    docsDialogBeside() {
      return this.docsShown
        && this.$vuetify.display.width >= DIALOG_MIN_WIDTH + 2 * DIALOG_MARGIN + DOCS_WIDTH;
    },

    // A dialog host: the classes of the content element of its v-dialog (HelpPanel.vue styles).
    docsDialogClasses() {
      return [
        this.docsShown ? 'item-dialog--docs' : '',
        this.docsDialogBeside ? 'item-dialog--docs-beside' : '',
      ].filter((name) => name !== '');
    },

    docsPanelProps() {
      return {
        doc: this.docsDoc,
        activeField: this.docsActiveField,
        followField: this.docsFollowField,
        missingFields: this.docsMissingFields,
      };
    },
  },

  watch: {
    docsShown(val) {
      if (val) {
        // The panel and the described elements are in the document after the next render.
        this.$nextTick(() => this.watchDocsFields());
      } else {
        this.unwatchDocsFields();
      }
    },

    docsActiveField() {
      this.markDocsField();
    },

    // Another document: the panel must not keep showing the previous one.
    docsKey() {
      this.closeDocs();
      this.docsDoc = null;
    },

    // The language of the interface changed (it does so without a reload after login): an open
    // panel shows the document in the new language, a closed one loads it on the next open.
    docsLocale() {
      if (this.docsOpen) {
        this.loadDocs();
      } else {
        this.docsDoc = null;
      }
    },
  },

  created() {
    // The element marked as the one the active section describes, the element in focus with
    // the `data-testid` of `docsFocusField` (one of the rows of a table, which share their
    // `data-testid`s), and the observer of the elements the host shows. Not reactive: elements.
    this.docsMarked = null;
    this.docsFocusElement = null;
    this.docsObserver = null;
  },

  beforeUnmount() {
    // Closed first: nothing is marked for a panel that is not shown.
    this.closeDocs();
    this.unwatchDocsFields();
  },

  methods: {
    docsRootElement() {
      return this.$el;
    },

    docsCanOpen() {
      return true;
    },

    async toggleDocs() {
      if (this.docsOpen) {
        this.closeDocs();
        return;
      }
      if (await this.loadDocs()) {
        // Listed before the panel is drawn: it opens with its sections disabled already,
        // rather than drawn enabled and then fading out.
        this.docsMissingFields = this.listDocsMissingFields();
        this.docsOpen = true;
      }
    },

    // Loads the document in the language of the interface, unless it is loaded already.
    // Returns whether the document is there.
    async loadDocs() {
      const key = this.docsKey;
      const locale = this.docsLocale;
      if (this.docsDoc != null && this.docsDoc.key === key && this.docsDoc.locale === locale) {
        return true;
      }

      let doc;
      try {
        doc = await loadHelpDoc(key, locale);
      } catch (err) {
        EventBus.$emit('i-snackbar', { color: 'error', text: err.message });
        return false;
      }
      // The host was closed, or given another document or language, while this one was loading.
      if (!this.docsCanOpen() || this.docsKey !== key || this.docsLocale !== locale) {
        return false;
      }
      this.docsDoc = doc;
      return true;
    },

    closeDocs() {
      this.docsOpen = false;
      this.docsFocusField = null;
      this.docsFocusSource = null;
      this.docsFocusElement = null;
      this.docsHoverField = null;
    },

    // The described element another element belongs to: the nearest element, the element
    // itself or one above it, with a `data-testid` a section of the document is anchored to.
    // Vuetify puts the attribute on the root of an input, around the element that takes the
    // focus.
    docsFieldElementOf(el) {
      const fields = new Set(this.docsDoc.sections.map((section) => section.field));
      let node = el.closest('[data-testid]');
      while (node) {
        if (fields.has(node.getAttribute('data-testid'))) {
          return node;
        }
        node = node.parentElement ? node.parentElement.closest('[data-testid]') : null;
      }
      return null;
    },

    // A dialog host: the width a dialog of its own `own` width may take next to the panel.
    docsDialogMaxWidth(own) {
      if (!this.docsDialogBeside) {
        return own;
      }
      const available = this.$vuetify.display.width - 2 * DIALOG_MARGIN - DOCS_WIDTH;
      const width = Number.parseFloat(own);
      return Number.isNaN(width) ? available : Math.min(width, available);
    },

    // The focus came to an element of the host. A menu of a select is outside the host: the
    // focus in its items is not seen here, and the field stays the active one. The focus in the
    // panel is the panel's own business (HelpPanel reports it), also when the host listens on an
    // element that holds the panel too.
    onDocsHostFocusIn(ev) {
      if (!this.docsShown || ev.target.closest('.HelpPanel')) {
        return;
      }
      const el = this.docsFieldElementOf(ev.target);
      this.docsFocusElement = el;
      this.docsFocusField = el ? el.getAttribute('data-testid') : null;
      this.docsFocusSource = el ? 'form' : null;
      // The focus may go to another row of the same field: the active field stays, the mark
      // moves.
      this.markDocsField();
    },

    // The focus came to a section of the panel (its field), or to another part of the panel
    // (null).
    onDocsFocusField(field) {
      this.docsFocusField = field;
      this.docsFocusSource = field ? 'panel' : null;
      this.docsFocusElement = null;
      this.markDocsField();
      if (field) {
        this.$nextTick(() => this.revealDocsField());
      }
    },

    // The element of the host a section is anchored to: the first one on screen. The rows of
    // a table share the `data-testid`s of their fields; a tab the user has left stays in the
    // page, hidden.
    docsFieldElement(field) {
      const root = this.docsRootElement();
      if (!root || !field) {
        return null;
      }
      // An anchor is a name ([A-Za-z][\w-]*, checked at build time): safe in a selector.
      return [...root.querySelectorAll(`[data-testid="${field}"]`)]
        .find((el) => isShown(el, root)) || null;
    },

    // The element the active section describes: the element in focus when the section is the
    // one of the field in focus, so that of the rows of a table the one in focus is marked;
    // else the first element of the section on screen.
    docsActiveElement() {
      const field = this.docsActiveField;
      const el = this.docsFocusElement;
      const root = this.docsRootElement();
      if (el && root && field === this.docsFocusField && root.contains(el) && isShown(el, root)) {
        return el;
      }
      return this.docsFieldElement(field);
    },

    // Marks the element the active section describes, and unmarks the one marked before. Of
    // a Vuetify input the frame is marked (or the control, of a checkbox or a switch, which
    // have no frame), not the root, which also holds the messages under the field; anything
    // else, as a button or the code editor, is marked itself.
    markDocsField() {
      if (this.docsMarked) {
        this.docsMarked.removeAttribute(DOCS_FIELD_ATTR);
        this.docsMarked = null;
      }
      const el = this.docsShown ? this.docsActiveElement() : null;
      if (el) {
        this.docsMarked = el.querySelector('.v-field') || el.querySelector('.v-input__control') || el;
        this.docsMarked.setAttribute(DOCS_FIELD_ATTR, '');
      }
    },

    // Scrolls the host to the marked element when it is out of view.
    revealDocsField() {
      if (this.docsMarked && typeof this.docsMarked.scrollIntoView === 'function') {
        this.docsMarked.scrollIntoView({ block: 'nearest' });
      }
    },

    // Elements come and go with the state of the host (the path of an inventory with its
    // type), and are hidden and shown without leaving the page (`v-show`: the tabs of
    // `v-window`, whose transition ends with a `style` and a `class` on the item): the missing
    // ones are listed for the panel, and the mark follows the element. The mark itself is an
    // attribute the observer does not watch.
    watchDocsFields() {
      const root = this.docsRootElement();
      if (!root || !this.docsShown || this.docsObserver) {
        return;
      }

      const update = () => {
        const missing = this.listDocsMissingFields();
        if (missing.join() !== this.docsMissingFields.join()) {
          this.docsMissingFields = missing;
        }
        this.markDocsField();
      };

      update();
      this.docsObserver = new MutationObserver(update);
      this.docsObserver.observe(root, {
        childList: true,
        subtree: true,
        attributes: true,
        attributeFilter: ['style', 'class', 'hidden'],
      });
    },

    // The fields of the sections of the document that the host does not show now.
    listDocsMissingFields() {
      return this.docsDoc.sections
        .map((section) => section.field)
        .filter((field) => this.docsFieldElement(field) == null);
    },

    unwatchDocsFields() {
      if (this.docsObserver) {
        this.docsObserver.disconnect();
        this.docsObserver = null;
      }
      this.docsMissingFields = [];
      this.markDocsField();
    },
  },
};
