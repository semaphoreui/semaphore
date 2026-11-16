import { expect } from 'chai';
import { reactive } from 'vue';
import { mount, flushPromises } from '@vue/test-utils';
import EditDialog from '@/components/EditDialog.vue';
import EventBus from '@/event-bus';
// Vitest has run the setup ahead of the spec already (`test.setupFiles`).
import createTestVuetify from './setup';

// The help documents the dialog can open: the registry is replaced, so the specs do not depend
// on the documents the application ships.
const { helpDoc } = vi.hoisted(() => ({
  helpDoc: (key, locale = 'en') => ({
    key,
    locale,
    title: locale === 'en' ? `Edit ${key}` : `[${locale}] Edit ${key}`,
    intro: [{ type: 'p', inline: [{ type: 'text', text: 'What the dialog is for.' }] }],
    // The last field is one the form of the specs does not show.
    sections: ['demoForm-name', 'demoForm-type', 'demoForm-gone'].map((field) => ({
      field,
      title: field,
      blocks: [{ type: 'p', inline: [{ type: 'text', text: `About ${field}.` }] }],
    })),
  }),
}));

// `demo` is translated into German; the others are English only.
vi.mock('@/help', () => {
  const hasHelpDoc = (key) => ['demo', 'other', 'broken'].includes(key);
  const helpDocLocale = (key, locale) => {
    if (!hasHelpDoc(key)) return null;
    const base = String(locale || '').toLowerCase().split('_')[0];
    return key === 'demo' && base === 'de' ? 'de' : 'en';
  };
  return {
    hasHelpDoc,
    helpDocLocale,
    loadHelpDoc: (key, locale) => (key === 'broken'
      ? Promise.reject(new Error('chunk failed to load'))
      : Promise.resolve(helpDoc(key, helpDocLocale(key, locale)))),
  };
});

function pressEscape() {
  window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', keyCode: 27, bubbles: true }));
}

// Vuetify 3 updates `globalTop` of the open overlays in a timer after the stack changes.
function overlayStackSettled() {
  return new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountDialog(props = {}, plugins = [], slots = {}, mocks = {}) {
  // Vuetify 3 teleports the dialog content to its overlay container in <body>. Mount into a
  // child node of a [data-app] element, removed after each test.
  const app = document.createElement('div');
  app.setAttribute('data-app', 'true');
  document.body.appendChild(app);
  const mountPoint = document.createElement('div');
  app.appendChild(mountPoint);

  return mount(EditDialog, {
    attachTo: mountPoint,
    props: { title: 'Test', ...props },
    global: { mocks: { $t: (k) => k, ...mocks }, plugins },
    slots,
  });
}

describe('EditDialog.vue', () => {
  const wrappers = [];

  afterEach(() => {
    while (wrappers.length) {
      wrappers.pop().unmount();
    }
    document.querySelectorAll('[data-app]').forEach((el) => el.remove());
  });

  async function open(props, plugins, slots, mocks) {
    const wrapper = mountDialog(props, plugins, slots, mocks);
    wrappers.push(wrapper);
    await wrapper.setProps({ modelValue: true });
    await wrapper.vm.$nextTick();
    await overlayStackSettled();
    expect(wrapper.vm.dialog).to.equal(true);
    return wrapper;
  }

  it('closes on Escape when it is the only dialog', async () => {
    const wrapper = await open();

    pressEscape();
    await wrapper.vm.$nextTick();

    expect(wrapper.vm.dialog).to.equal(false);
  });

  it('does not close on Escape when noEscape is set', async () => {
    const wrapper = await open({ noEscape: true });

    pressEscape();
    await wrapper.vm.$nextTick();

    expect(wrapper.vm.dialog).to.equal(true);
  });

  it('closes only the topmost dialog when dialogs are stacked', async () => {
    const outer = await open();
    const inner = await open();

    pressEscape();
    await outer.vm.$nextTick();

    expect(inner.vm.dialog, 'inner dialog must close').to.equal(false);
    expect(outer.vm.dialog, 'outer dialog must stay open').to.equal(true);
    await overlayStackSettled();

    pressEscape();
    await outer.vm.$nextTick();

    expect(outer.vm.dialog, 'outer dialog closes on the next Escape').to.equal(false);
  });

  describe('resizable', () => {
    // jsdom lays nothing out: the dialog measures 1000 px in an overlay of 1600 px with the
    // 24 px margins of the Vuetify stylesheet, and its handles capture no pointer.
    function layOut(handle, width = 1000) {
      const content = handle.parentElement;
      content.style.margin = '24px';
      content.getBoundingClientRect = () => ({ width });
      Object.defineProperty(content.parentElement, 'clientWidth', { value: 1600, configurable: true });
      // eslint-disable-next-line no-param-reassign
      handle.setPointerCapture = () => {};
    }

    function handleOf(side) {
      return document.querySelector(`[data-testid="editDialog-resize-${side}"]`);
    }

    async function drag(wrapper, side, from, to) {
      const handle = handleOf(side);
      layOut(handle);
      handle.dispatchEvent(new MouseEvent('pointerdown', { clientX: from, bubbles: true }));
      handle.dispatchEvent(new MouseEvent('pointermove', { clientX: to, bubbles: true }));
      await wrapper.vm.$nextTick();
      return handle;
    }

    function maxWidthOf(handle) {
      return handle.parentElement.style.maxWidth;
    }

    afterEach(() => {
      localStorage.clear();
    });

    it('has no handles unless resizable', async () => {
      await open({ maxWidth: 1000 });

      expect(handleOf('left')).to.equal(null);
      expect(handleOf('right')).to.equal(null);
    });

    it('has no handles in fullscreen', async () => {
      const wrapper = await open({ maxWidth: 1000, resizable: true, expandable: true });
      expect(handleOf('right')).to.not.equal(null);

      wrapper.vm.toggleFullscreen();
      await wrapper.vm.$nextTick();

      expect(handleOf('left')).to.equal(null);
      expect(handleOf('right')).to.equal(null);
    });

    it('keeps the width of maxWidth until it is dragged', async () => {
      await open({ maxWidth: 1000, resizable: true });

      expect(maxWidthOf(handleOf('right'))).to.equal('1000px');
    });

    const drags = [
      {
        name: 'right edge outwards', side: 'right', from: 1300, to: 1350, width: 1100,
      },
      {
        name: 'right edge inwards', side: 'right', from: 1300, to: 1200, width: 800,
      },
      {
        name: 'left edge outwards', side: 'left', from: 300, to: 250, width: 1100,
      },
      {
        name: 'left edge inwards', side: 'left', from: 300, to: 400, width: 800,
      },
      {
        name: 'not past the margins of the window', side: 'right', from: 1300, to: 1900, width: 1552,
      },
      {
        name: 'not below the minimum width', side: 'left', from: 300, to: 790, width: 400,
      },
    ];

    drags.forEach((tt) => {
      // Both edges move by the distance of the pointer, so the centred dialog stays centred.
      it(`changes the width by twice the distance: ${tt.name}`, async () => {
        const wrapper = await open({ maxWidth: 1000, resizable: true });

        const handle = await drag(wrapper, tt.side, tt.from, tt.to);

        expect(wrapper.vm.width).to.equal(tt.width);
        expect(maxWidthOf(handle)).to.equal(`${tt.width}px`);
      });
    });

    it('ignores the pointer when no drag is in progress', async () => {
      const wrapper = await open({ maxWidth: 1000, resizable: true });

      const handle = await drag(wrapper, 'right', 1300, 1350);
      handle.dispatchEvent(new MouseEvent('pointerup', { clientX: 1350, bubbles: true }));
      handle.dispatchEvent(new MouseEvent('pointermove', { clientX: 1500, bubbles: true }));
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.width).to.equal(1100);
    });

    it('remembers the width per dialog name and restores it', async () => {
      const wrapper = await open({ maxWidth: 1000, resizable: true, name: 'Resized' });

      const handle = await drag(wrapper, 'right', 1300, 1350);
      handle.dispatchEvent(new MouseEvent('pointerup', { clientX: 1350, bubbles: true }));

      expect(localStorage.getItem('EditDialog_Resized__width')).to.equal('1100');

      const restored = await open({ maxWidth: 1000, resizable: true, name: 'Resized' });
      expect(restored.vm.width).to.equal(1100);

      const other = await open({ maxWidth: 1000, resizable: true, name: 'Other' });
      expect(other.vm.width).to.equal(null);
    });

    it('returns to maxWidth on a double click of an edge', async () => {
      const wrapper = await open({ maxWidth: 1000, resizable: true, name: 'Resized' });

      const handle = await drag(wrapper, 'right', 1300, 1350);
      handle.dispatchEvent(new MouseEvent('pointerup', { clientX: 1350, bubbles: true }));
      handle.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }));
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.width).to.equal(null);
      expect(maxWidthOf(handle)).to.equal('1000px');
      expect(localStorage.getItem('EditDialog_Resized__width')).to.equal(null);
    });
  });

  describe('help panel', () => {
    const byTestId = (id) => document.querySelector(`[data-testid="${id}"]`);
    const docsButton = () => byTestId('editDialog-docs');
    const panel = () => byTestId('editDialog-docsPanel');

    // The document is loaded before the panel is drawn.
    async function click(wrapper, el) {
      el.click();
      await flushPromises();
      await wrapper.vm.$nextTick();
    }

    function isAhead(first, second) {
      // eslint-disable-next-line no-bitwise
      return (first.compareDocumentPosition(second) & Node.DOCUMENT_POSITION_FOLLOWING) !== 0;
    }

    it('has no help button without a document', async () => {
      await open();

      expect(docsButton()).to.equal(null);
    });

    it('has no help button for a document that does not exist', async () => {
      await open({ helpDoc: 'nope' });

      expect(docsButton()).to.equal(null);
    });

    it('puts the help button ahead of the other buttons of the title', async () => {
      await open({ helpDoc: 'demo', expandable: true });

      expect(docsButton()).to.not.equal(null);
      expect(isAhead(docsButton(), byTestId('editDialog-fullscreen'))).to.equal(true);
      expect(isAhead(docsButton(), byTestId('editDialog-close'))).to.equal(true);
      expect(docsButton().getAttribute('aria-pressed')).to.equal('false');
      expect(panel()).to.equal(null);
    });

    it('opens the panel on a click and closes it on the next one', async () => {
      const wrapper = await open({ helpDoc: 'demo' });

      await click(wrapper, docsButton());

      expect(panel()).to.not.equal(null);
      expect(byTestId('editDialog-docsTitle').textContent).to.equal('Edit demo');
      expect(docsButton().getAttribute('aria-pressed')).to.equal('true');
      expect(wrapper.vm.dialog).to.equal(true);

      await click(wrapper, docsButton());

      expect(panel()).to.equal(null);
      expect(wrapper.vm.dialog, 'the dialog stays open').to.equal(true);
    });

    it('closes the panel with its own button and keeps the dialog', async () => {
      const wrapper = await open({ helpDoc: 'demo' });
      await click(wrapper, docsButton());

      await click(wrapper, byTestId('editDialog-docsClose'));

      expect(panel()).to.equal(null);
      expect(wrapper.vm.dialog).to.equal(true);
    });

    it('draws the panel inside the content of the dialog, next to the card', async () => {
      const wrapper = await open({ helpDoc: 'demo', testId: 'demoDialog' });
      await click(wrapper, docsButton());

      expect(panel().parentElement).to.equal(byTestId('demoDialog').parentElement);
      expect(panel().parentElement.classList.contains('v-overlay__content')).to.equal(true);
    });

    it('lifts the containment of the content while the panel is open', async () => {
      const wrapper = await open({ helpDoc: 'demo' });
      const content = () => byTestId('editDialog-title').closest('.v-overlay__content');
      expect(content().classList.contains('item-dialog--docs')).to.equal(false);

      await click(wrapper, docsButton());

      expect(content().classList.contains('item-dialog--docs')).to.equal(true);
      expect(content().classList.contains('item-dialog--docs-beside')).to.equal(true);

      await click(wrapper, docsButton());

      expect(content().classList.contains('item-dialog--docs')).to.equal(false);
      expect(content().classList.contains('item-dialog--docs-beside')).to.equal(false);
    });

    it('narrows a wide dialog to what the panel leaves of the window', async () => {
      // jsdom: a window of 1024 px. The panel takes 420 px, the margins of the dialog 2 x 24 px.
      const wrapper = await open({ helpDoc: 'demo', maxWidth: 1000 });
      const content = () => byTestId('editDialog-title').closest('.v-overlay__content');
      expect(content().style.maxWidth).to.equal('1000px');

      await click(wrapper, docsButton());
      expect(content().style.maxWidth).to.equal(`${window.innerWidth - 468}px`);

      await click(wrapper, docsButton());
      expect(content().style.maxWidth).to.equal('1000px');
    });

    it('keeps the width of a dialog that fits next to the panel', async () => {
      const wrapper = await open({ helpDoc: 'demo', maxWidth: 450 });
      await click(wrapper, docsButton());

      const content = byTestId('editDialog-title').closest('.v-overlay__content');
      expect(content.style.maxWidth).to.equal('450px');
    });

    it('lays the panel over the dialog in a window too narrow for both', async () => {
      // The display is measured when the Vuetify instance is created.
      const innerWidth = window.innerWidth;
      window.innerWidth = 860;
      const vuetify = createTestVuetify();
      window.innerWidth = innerWidth;

      const wrapper = await open({ helpDoc: 'demo', maxWidth: 450 }, [vuetify]);
      await click(wrapper, docsButton());

      const content = byTestId('editDialog-title').closest('.v-overlay__content');
      expect(panel()).to.not.equal(null);
      expect(content.classList.contains('item-dialog--docs')).to.equal(true);
      expect(content.classList.contains('item-dialog--docs-beside')).to.equal(false);
      expect(content.style.maxWidth).to.equal('450px');
    });

    it('forgets the panel when the dialog closes', async () => {
      const wrapper = await open({ helpDoc: 'demo' });
      await click(wrapper, docsButton());

      wrapper.vm.close();
      await wrapper.vm.$nextTick();
      await wrapper.setProps({ modelValue: false });
      await wrapper.setProps({ modelValue: true });
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.dialog).to.equal(true);
      expect(panel()).to.equal(null);
    });

    it('closes the dialog on Escape with the panel open', async () => {
      const wrapper = await open({ helpDoc: 'demo' });
      await click(wrapper, docsButton());
      await overlayStackSettled();

      pressEscape();
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.dialog).to.equal(false);
      expect(wrapper.vm.docsOpen).to.equal(false);
    });

    it('keeps the button of the inline hints apart', async () => {
      const wrapper = await open({ helpDoc: 'demo', helpButton: true });

      await click(wrapper, byTestId('editDialog-help'));

      expect(wrapper.vm.needHelp).to.equal(true);
      expect(panel()).to.equal(null);

      await click(wrapper, docsButton());

      expect(panel()).to.not.equal(null);
      expect(wrapper.vm.needHelp, 'the panel leaves the inline hints as they were').to.equal(true);
    });

    it('drops the panel when the dialog gets another document', async () => {
      const wrapper = await open({ helpDoc: 'demo' });
      await click(wrapper, docsButton());

      await wrapper.setProps({ helpDoc: 'other' });
      expect(panel()).to.equal(null);

      await click(wrapper, docsButton());
      expect(byTestId('editDialog-docsTitle').textContent).to.equal('Edit other');
    });

    it('hides the button and the panel when the dialog loses its document', async () => {
      const wrapper = await open({ helpDoc: 'demo' });
      await click(wrapper, docsButton());

      await wrapper.setProps({ helpDoc: null });

      expect(docsButton()).to.equal(null);
      expect(panel()).to.equal(null);
    });

    describe('language', () => {
      // `$i18n` of vue-i18n: its locale is reactive and changes without a reload after login.
      function i18n(locale) {
        return reactive({ locale });
      }

      it('shows the document in English without vue-i18n', async () => {
        const wrapper = await open({ helpDoc: 'demo' });
        await click(wrapper, docsButton());

        expect(byTestId('editDialog-docsTitle').textContent).to.equal('Edit demo');
        expect(wrapper.vm.docsDoc.locale).to.equal('en');
      });

      it('shows the document in the language of the interface', async () => {
        const wrapper = await open({ helpDoc: 'demo' }, [], {}, { $i18n: i18n('de') });
        await click(wrapper, docsButton());

        expect(byTestId('editDialog-docsTitle').textContent).to.equal('[de] Edit demo');
      });

      it('takes the language without its region', async () => {
        const wrapper = await open({ helpDoc: 'demo' }, [], {}, { $i18n: i18n('de_de') });
        await click(wrapper, docsButton());

        expect(byTestId('editDialog-docsTitle').textContent).to.equal('[de] Edit demo');
      });

      it('falls back to English for a language the document is not translated into', async () => {
        const wrapper = await open({ helpDoc: 'other' }, [], {}, { $i18n: i18n('de') });
        await click(wrapper, docsButton());

        expect(byTestId('editDialog-docsTitle').textContent).to.equal('Edit other');
      });

      it('changes the document of an open panel when the language changes', async () => {
        const lang = i18n('en');
        const wrapper = await open({ helpDoc: 'demo' }, [], {}, { $i18n: lang });
        await click(wrapper, docsButton());
        expect(byTestId('editDialog-docsTitle').textContent).to.equal('Edit demo');

        lang.locale = 'de';
        await flushPromises();
        await wrapper.vm.$nextTick();

        expect(panel(), 'the panel stays open').to.not.equal(null);
        expect(byTestId('editDialog-docsTitle').textContent).to.equal('[de] Edit demo');
      });

      it('loads the new language on the next open when the panel is closed', async () => {
        const lang = i18n('en');
        const wrapper = await open({ helpDoc: 'demo' }, [], {}, { $i18n: lang });
        await click(wrapper, docsButton());
        await click(wrapper, docsButton());

        lang.locale = 'de';
        await flushPromises();
        expect(panel()).to.equal(null);

        await click(wrapper, docsButton());
        expect(byTestId('editDialog-docsTitle').textContent).to.equal('[de] Edit demo');
      });
    });

    it('reports a document that fails to load and stays closed', async () => {
      const shown = [];
      const onSnackbar = (e) => shown.push(e);
      EventBus.$on('i-snackbar', onSnackbar);

      const wrapper = await open({ helpDoc: 'broken' });
      await click(wrapper, docsButton());
      EventBus.$off('i-snackbar', onSnackbar);

      expect(panel()).to.equal(null);
      expect(wrapper.vm.dialog).to.equal(true);
      expect(shown).to.deep.equal([{ color: 'error', text: 'chunk failed to load' }]);
    });
  });

  describe('help panel and the fields of the form', () => {
    // A form as Vuetify draws one: the `data-testid` of an input is on its root, around the
    // frame (`.v-field`) and the element that takes the focus; a plain element carries its own.
    const FORM = `
      <div class="demo-form">
        <div data-testid="demoForm-name"><div class="v-field"><input class="demo-name"></div></div>
        <textarea data-testid="demoForm-type" class="demo-type"></textarea>
        <button type="button" class="demo-other">other</button>
      </div>`;
    const MARK = 'data-help-docs-field';

    const byTestId = (id) => document.querySelector(`[data-testid="${id}"]`);
    const section = (field) => byTestId(`editDialog-docsSection-${field}`);
    const marked = () => [...document.querySelectorAll(`[${MARK}]`)];
    const nameFrame = () => document.querySelector('[data-testid="demoForm-name"] .v-field');

    async function settled(wrapper) {
      await flushPromises();
      await wrapper.vm.$nextTick();
    }

    async function openWithPanel(props = {}) {
      const wrapper = await open({ helpDoc: 'demo', ...props }, [], { form: FORM });
      byTestId('editDialog-docs').click();
      await settled(wrapper);
      return wrapper;
    }

    async function focusIn(wrapper, el) {
      el.dispatchEvent(new FocusEvent('focusin', { bubbles: true }));
      await settled(wrapper);
    }

    async function pointer(wrapper, el, type) {
      el.dispatchEvent(new MouseEvent(type));
      await settled(wrapper);
    }

    // jsdom lays nothing out and scrolls nothing: the panel is given a view of 100 px with the
    // sections one under another, 300 px each, and a scrollTo that records its argument.
    function layOutPanel() {
      const body = byTestId('editDialog-docsBody');
      const calls = [];
      body.getBoundingClientRect = () => ({ top: 50, bottom: 150 });
      body.scrollTo = (to) => calls.push(to);
      [...body.querySelectorAll('section')].forEach((el, i) => {
        // eslint-disable-next-line no-param-reassign
        el.getBoundingClientRect = () => ({ top: 50 + 300 * i, bottom: 350 + 300 * i });
      });
      return calls;
    }

    describe('focus in a field of the form', () => {
      it('makes the section of the field the active one and scrolls the panel to it', async () => {
        const wrapper = await openWithPanel();
        const scrolls = layOutPanel();

        await focusIn(wrapper, document.querySelector('.demo-type'));

        expect(wrapper.vm.docsActiveField).to.equal('demoForm-type');
        expect(section('demoForm-type').getAttribute('aria-current')).to.equal('true');
        expect(section('demoForm-name').getAttribute('aria-current')).to.equal(null);
        expect(scrolls).to.deep.equal([{ top: 300, behavior: 'smooth' }]);
      });

      it('finds the field of an element inside the root of an input', async () => {
        const wrapper = await openWithPanel();

        await focusIn(wrapper, document.querySelector('.demo-name'));

        expect(wrapper.vm.docsActiveField).to.equal('demoForm-name');
      });

      it('leaves the panel where it is when the section is in view', async () => {
        const wrapper = await openWithPanel();
        const scrolls = layOutPanel();
        byTestId('editDialog-docsBody').getBoundingClientRect = () => ({ top: 0, bottom: 2000 });

        await focusIn(wrapper, document.querySelector('.demo-type'));

        expect(wrapper.vm.docsActiveField).to.equal('demoForm-type');
        expect(scrolls).to.deep.equal([]);
      });

      it('scrolls again when the focus comes back to the field', async () => {
        const wrapper = await openWithPanel();
        const scrolls = layOutPanel();

        await focusIn(wrapper, document.querySelector('.demo-type'));
        await focusIn(wrapper, document.querySelector('.demo-other'));
        await focusIn(wrapper, document.querySelector('.demo-type'));

        expect(scrolls).to.have.length(2);
      });

      it('has no active section while the focus is on an element without one', async () => {
        const wrapper = await openWithPanel();
        const scrolls = layOutPanel();
        await focusIn(wrapper, document.querySelector('.demo-name'));
        scrolls.length = 0;

        await focusIn(wrapper, document.querySelector('.demo-other'));

        expect(wrapper.vm.docsActiveField).to.equal(null);
        expect(section('demoForm-name').getAttribute('aria-current')).to.equal(null);
        expect(marked()).to.deep.equal([]);
        expect(scrolls).to.deep.equal([]);
      });

      it('does nothing while the panel is closed', async () => {
        const wrapper = await open({ helpDoc: 'demo' }, [], { form: FORM });

        await focusIn(wrapper, document.querySelector('.demo-name'));

        expect(wrapper.vm.docsActiveField).to.equal(null);
        expect(marked()).to.deep.equal([]);
      });

      it('opens the panel at its top, whatever field had the focus', async () => {
        const wrapper = await open({ helpDoc: 'demo' }, [], { form: FORM });
        await focusIn(wrapper, document.querySelector('.demo-name'));

        byTestId('editDialog-docs').click();
        await settled(wrapper);

        expect(wrapper.vm.docsActiveField).to.equal(null);
        expect(wrapper.vm.docsFollowField).to.equal(null);
      });

      it('marks the field in focus as the field of the active section', async () => {
        const wrapper = await openWithPanel();

        await focusIn(wrapper, document.querySelector('.demo-name'));

        expect(marked()).to.deep.equal([nameFrame()]);
      });
    });

    describe('active section of the panel', () => {
      it('marks the field of the section under the pointer, and unmarks it after', async () => {
        const wrapper = await openWithPanel();

        await pointer(wrapper, section('demoForm-name'), 'mouseenter');

        expect(wrapper.vm.docsActiveField).to.equal('demoForm-name');
        expect(section('demoForm-name').classList.contains('HelpPanel__section--active')).to.equal(true);
        expect(marked(), 'the frame of the input, not its root').to.deep.equal([nameFrame()]);

        await pointer(wrapper, section('demoForm-name'), 'mouseleave');

        expect(wrapper.vm.docsActiveField).to.equal(null);
        expect(section('demoForm-name').classList.contains('HelpPanel__section--active')).to.equal(false);
        expect(marked()).to.deep.equal([]);
      });

      it('marks an element without a frame itself', async () => {
        const wrapper = await openWithPanel();

        await pointer(wrapper, section('demoForm-type'), 'mouseenter');

        expect(marked()).to.deep.equal([document.querySelector('.demo-type')]);
      });

      it('marks the field of the section in focus and brings it into view', async () => {
        const wrapper = await openWithPanel();
        const scrolls = layOutPanel();
        const revealed = [];
        document.querySelector('.demo-type').scrollIntoView = (options) => revealed.push(options);

        await focusIn(wrapper, section('demoForm-type'));

        expect(wrapper.vm.docsActiveField).to.equal('demoForm-type');
        expect(marked()).to.deep.equal([document.querySelector('.demo-type')]);
        expect(revealed).to.deep.equal([{ block: 'nearest' }]);
        expect(scrolls, 'the panel does not scroll to a section activated in the panel').to.deep.equal([]);
      });

      it('does not scroll the dialog for a section under the pointer', async () => {
        const wrapper = await openWithPanel();
        const revealed = [];
        document.querySelector('.demo-type').scrollIntoView = (options) => revealed.push(options);

        await pointer(wrapper, section('demoForm-type'), 'mouseenter');

        expect(revealed).to.deep.equal([]);
      });

      it('prefers the section under the pointer to the one in focus', async () => {
        const wrapper = await openWithPanel();
        await focusIn(wrapper, section('demoForm-type'));

        await pointer(wrapper, section('demoForm-name'), 'mouseenter');
        expect(marked()).to.deep.equal([nameFrame()]);

        await pointer(wrapper, section('demoForm-name'), 'mouseleave');
        expect(marked()).to.deep.equal([document.querySelector('.demo-type')]);
      });

      it('has no active section while the focus is elsewhere in the panel', async () => {
        const wrapper = await openWithPanel();
        await focusIn(wrapper, section('demoForm-name'));

        await focusIn(wrapper, byTestId('editDialog-docsClose'));

        expect(wrapper.vm.docsActiveField).to.equal(null);
        expect(marked()).to.deep.equal([]);
      });

      it('disables the section of a field the form does not show', async () => {
        const wrapper = await openWithPanel();
        const gone = section('demoForm-gone');

        expect(byTestId('editDialog-docsSectionHidden-demoForm-gone')).to.not.equal(null);
        expect(byTestId('editDialog-docsSectionHidden-demoForm-name')).to.equal(null);
        expect(byTestId('editDialog-docsSectionHidden-demoForm-type')).to.equal(null);
        expect(gone.classList.contains('HelpPanel__section--disabled')).to.equal(true);
        expect(gone.hasAttribute('tabindex'), 'out of the Tab order').to.equal(false);
        expect(section('demoForm-name').classList.contains('HelpPanel__section--disabled'))
          .to.equal(false);

        await pointer(wrapper, gone, 'mouseenter');
        await focusIn(wrapper, gone);

        expect(wrapper.vm.docsActiveField).to.equal(null);
        expect(gone.classList.contains('HelpPanel__section--active')).to.equal(false);
        expect(marked()).to.deep.equal([]);
      });

      it('opens with the sections of missing fields disabled in its first render', async () => {
        const wrapper = await open({ helpDoc: 'demo' }, [], { form: FORM });

        // Resolves when the panel is to open, ahead of its render and of the watcher of the
        // fields: a section drawn enabled first would fade to disabled in the browser.
        await wrapper.vm.toggleDocs();

        expect(wrapper.vm.docsOpen).to.equal(true);
        expect(wrapper.vm.docsMissingFields).to.deep.equal(['demoForm-gone']);
      });

      it('enables the section when its field appears and disables it when the field goes', async () => {
        const wrapper = await openWithPanel();
        const gone = section('demoForm-gone');

        const field = document.createElement('input');
        field.setAttribute('data-testid', 'demoForm-gone');
        document.querySelector('.demo-form').appendChild(field);
        await settled(wrapper);

        expect(byTestId('editDialog-docsSectionHidden-demoForm-gone')).to.equal(null);
        expect(gone.classList.contains('HelpPanel__section--disabled')).to.equal(false);
        expect(gone.getAttribute('tabindex')).to.equal('0');

        await pointer(wrapper, gone, 'mouseenter');

        expect(gone.classList.contains('HelpPanel__section--active')).to.equal(true);
        expect(marked()).to.deep.equal([field]);

        // The field goes while the pointer is still over its section.
        field.remove();
        await settled(wrapper);

        expect(byTestId('editDialog-docsSectionHidden-demoForm-gone')).to.not.equal(null);
        expect(gone.classList.contains('HelpPanel__section--disabled')).to.equal(true);
        expect(gone.classList.contains('HelpPanel__section--active')).to.equal(false);
        expect(marked()).to.deep.equal([]);
      });

      it('unmarks the field when the panel closes', async () => {
        const wrapper = await openWithPanel();
        await focusIn(wrapper, section('demoForm-name'));
        expect(marked()).to.have.length(1);

        byTestId('editDialog-docsClose').click();
        await settled(wrapper);

        expect(marked()).to.deep.equal([]);
        expect(wrapper.vm.docsActiveField).to.equal(null);
      });

      it('unmarks the field when the dialog closes', async () => {
        const wrapper = await openWithPanel();
        await pointer(wrapper, section('demoForm-name'), 'mouseenter');
        expect(marked()).to.have.length(1);

        wrapper.vm.close();
        await settled(wrapper);

        expect(marked()).to.deep.equal([]);
        expect(wrapper.vm.docsActiveField).to.equal(null);
      });

      it('looks for the field in its own dialog only', async () => {
        const outer = await openWithPanel({ testId: 'outerDialog' });
        // A second dialog with the same form and no panel.
        await open({ testId: 'innerDialog' }, [], { form: FORM });

        await pointer(outer, section('demoForm-name'), 'mouseenter');

        expect(marked()).to.have.length(1);
        expect(byTestId('outerDialog').contains(marked()[0])).to.equal(true);
      });
    });

    // A tab of `v-window` the user has left stays in the page, hidden by `display: none` on its
    // item (`v-show`), as does any other part of a form behind `v-show`.
    describe('a field hidden in the page', () => {
      async function hiddenField(wrapper) {
        const tab = document.createElement('div');
        tab.style.display = 'none';
        const field = document.createElement('input');
        field.setAttribute('data-testid', 'demoForm-gone');
        tab.appendChild(field);
        document.querySelector('.demo-form').appendChild(tab);
        await settled(wrapper);
        return { tab, field };
      }

      it('disables the section of a field inside a hidden element', async () => {
        const wrapper = await openWithPanel();

        await hiddenField(wrapper);

        expect(byTestId('editDialog-docsSectionHidden-demoForm-gone')).to.not.equal(null);
        expect(section('demoForm-gone').classList.contains('HelpPanel__section--disabled'))
          .to.equal(true);
      });

      it('enables the section when the hidden element is shown, and disables it again', async () => {
        const wrapper = await openWithPanel();
        const { tab, field } = await hiddenField(wrapper);

        tab.style.display = '';
        await settled(wrapper);

        expect(byTestId('editDialog-docsSectionHidden-demoForm-gone')).to.equal(null);
        await pointer(wrapper, section('demoForm-gone'), 'mouseenter');
        expect(marked()).to.deep.equal([field]);

        // The user goes to another tab while the pointer is over the section.
        tab.style.display = 'none';
        await settled(wrapper);

        expect(byTestId('editDialog-docsSectionHidden-demoForm-gone')).to.not.equal(null);
        expect(marked()).to.deep.equal([]);
      });

      it('marks the shown one of the fields of a section', async () => {
        const wrapper = await openWithPanel();
        const { tab } = await hiddenField(wrapper);
        const shown = document.createElement('input');
        shown.setAttribute('data-testid', 'demoForm-gone');
        document.querySelector('.demo-form').appendChild(shown);
        await settled(wrapper);
        // The hidden one comes first in the page.
        expect(document.querySelector('[data-testid="demoForm-gone"]').parentElement).to.equal(tab);

        await pointer(wrapper, section('demoForm-gone'), 'mouseenter');

        expect(marked()).to.deep.equal([shown]);
      });
    });

    // The rows of a table (the variables of a variable group) repeat the `data-testid`s of
    // their fields: one section describes the field of every row.
    describe('a field repeated in rows', () => {
      async function secondRow(wrapper) {
        const row = document.createElement('div');
        row.innerHTML = '<div data-testid="demoForm-name"><div class="v-field"><input class="demo-name2"></div></div>';
        document.querySelector('.demo-form').appendChild(row);
        await settled(wrapper);
        return row.querySelector('.v-field');
      }

      it('marks the row in focus, not the first one', async () => {
        const wrapper = await openWithPanel();
        const frame = await secondRow(wrapper);

        await focusIn(wrapper, document.querySelector('.demo-name2'));

        expect(wrapper.vm.docsActiveField).to.equal('demoForm-name');
        expect(marked()).to.deep.equal([frame]);
      });

      it('moves the mark with the focus from row to row', async () => {
        const wrapper = await openWithPanel();
        const frame = await secondRow(wrapper);

        await focusIn(wrapper, document.querySelector('.demo-name'));
        expect(marked()).to.deep.equal([nameFrame()]);

        await focusIn(wrapper, document.querySelector('.demo-name2'));
        expect(marked()).to.deep.equal([frame]);

        await focusIn(wrapper, document.querySelector('.demo-name'));
        expect(marked()).to.deep.equal([nameFrame()]);
      });

      it('keeps the row in focus marked under the pointer over its section', async () => {
        const wrapper = await openWithPanel();
        const frame = await secondRow(wrapper);
        await focusIn(wrapper, document.querySelector('.demo-name2'));

        await pointer(wrapper, section('demoForm-name'), 'mouseenter');
        expect(marked()).to.deep.equal([frame]);

        await pointer(wrapper, section('demoForm-type'), 'mouseenter');
        expect(marked()).to.deep.equal([document.querySelector('.demo-type')]);

        await pointer(wrapper, section('demoForm-type'), 'mouseleave');
        expect(marked()).to.deep.equal([frame]);
      });

      it('marks the first row for a section in focus', async () => {
        const wrapper = await openWithPanel();
        await secondRow(wrapper);
        await focusIn(wrapper, document.querySelector('.demo-name2'));

        await focusIn(wrapper, section('demoForm-name'));

        expect(marked()).to.deep.equal([nameFrame()]);
      });
    });
  });
});
