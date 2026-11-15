import { expect } from 'chai';
import { mount, flushPromises } from '@vue/test-utils';
import EditDialog from '@/components/EditDialog.vue';
import EventBus from '@/event-bus';
// Vitest has run the setup ahead of the spec already (`test.setupFiles`).
import createTestVuetify from './setup';

// The help documents the dialog can open: the registry is replaced, so the specs do not depend
// on the documents the application ships.
const { helpDoc } = vi.hoisted(() => ({
  helpDoc: (key) => ({
    key,
    title: `Edit ${key}`,
    intro: [{ type: 'p', inline: [{ type: 'text', text: 'What the dialog is for.' }] }],
    sections: [{
      field: 'demoForm-name',
      title: 'Name',
      blocks: [{ type: 'p', inline: [{ type: 'text', text: 'The name.' }] }],
    }],
  }),
}));

vi.mock('@/help', () => ({
  hasHelpDoc: (key) => ['demo', 'other', 'broken'].includes(key),
  loadHelpDoc: (key) => (key === 'broken'
    ? Promise.reject(new Error('chunk failed to load'))
    : Promise.resolve(helpDoc(key))),
}));

function pressEscape() {
  window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', keyCode: 27, bubbles: true }));
}

// Vuetify 3 updates `globalTop` of the open overlays in a timer after the stack changes.
function overlayStackSettled() {
  return new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountDialog(props = {}, plugins = []) {
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
    global: { mocks: { $t: (k) => k }, plugins },
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

  async function open(props, plugins) {
    const wrapper = mountDialog(props, plugins);
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
});
