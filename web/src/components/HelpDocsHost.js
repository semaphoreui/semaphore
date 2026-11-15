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

// The class of the element the active section of the panel describes (styled in HelpPanel.vue).
const DOCS_FIELD_CLASS = 'help-docs-field';

// The width of the help panel (HelpPanel.vue).
export const DOCS_WIDTH = 420;

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
    // The element marked as the one the active section describes, and the observer of the
    // elements the host shows.
    this.docsMarked = null;
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
      this.docsHoverField = null;
    },

    // The described element another element belongs to: the nearest `data-testid`, on the
    // element or above it, that a section of the document is anchored to. Vuetify puts the
    // attribute on the root of an input, around the element that takes the focus.
    docsFieldOf(el) {
      const fields = new Set(this.docsDoc.sections.map((section) => section.field));
      let node = el.closest('[data-testid]');
      while (node) {
        const id = node.getAttribute('data-testid');
        if (fields.has(id)) {
          return id;
        }
        node = node.parentElement ? node.parentElement.closest('[data-testid]') : null;
      }
      return null;
    },

    // The focus came to an element of the host. A menu of a select is outside the host: the
    // focus in its items is not seen here, and the field stays the active one.
    onDocsHostFocusIn(ev) {
      if (!this.docsShown) {
        return;
      }
      const field = this.docsFieldOf(ev.target);
      this.docsFocusField = field;
      this.docsFocusSource = field ? 'form' : null;
    },

    // The focus came to a section of the panel (its field), or to another part of the panel
    // (null).
    onDocsFocusField(field) {
      this.docsFocusField = field;
      this.docsFocusSource = field ? 'panel' : null;
      if (field) {
        // After the watcher has marked the element.
        this.$nextTick(() => this.revealDocsField());
      }
    },

    docsFieldElement(field) {
      const root = this.docsRootElement();
      if (!root || !field) {
        return null;
      }
      // An anchor is a name ([A-Za-z][\w-]*, checked at build time): safe in a selector.
      return root.querySelector(`[data-testid="${field}"]`);
    },

    // Marks the element the active section describes, and unmarks the one marked before. Of
    // a Vuetify input the frame is marked (or the control, of a checkbox or a switch, which
    // have no frame), not the root, which also holds the messages under the field; anything
    // else, as a button or the code editor, is marked itself.
    markDocsField() {
      if (this.docsMarked) {
        this.docsMarked.classList.remove(DOCS_FIELD_CLASS);
        this.docsMarked = null;
      }
      const el = this.docsShown ? this.docsFieldElement(this.docsActiveField) : null;
      if (el) {
        this.docsMarked = el.querySelector('.v-field') || el.querySelector('.v-input__control') || el;
        this.docsMarked.classList.add(DOCS_FIELD_CLASS);
      }
    },

    // Scrolls the host to the marked element when it is out of view.
    revealDocsField() {
      if (this.docsMarked && typeof this.docsMarked.scrollIntoView === 'function') {
        this.docsMarked.scrollIntoView({ block: 'nearest' });
      }
    },

    // Elements come and go with the state of the host (the path of an inventory with its
    // type): the missing ones are listed for the panel, and the mark follows the element.
    watchDocsFields() {
      const root = this.docsRootElement();
      if (!root || !this.docsShown || this.docsObserver) {
        return;
      }

      const update = () => {
        const missing = this.docsDoc.sections
          .map((section) => section.field)
          .filter((field) => this.docsFieldElement(field) == null);
        if (missing.join() !== this.docsMissingFields.join()) {
          this.docsMissingFields = missing;
        }
        this.markDocsField();
      };

      update();
      this.docsObserver = new MutationObserver(update);
      this.docsObserver.observe(root, { childList: true, subtree: true });
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
