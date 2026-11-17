import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import EditDialog from '@/components/EditDialog.vue';

function pressEscape() {
  window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', keyCode: 27, bubbles: true }));
}

// Vuetify 3 updates `globalTop` of the open overlays in a timer after the stack changes.
function overlayStackSettled() {
  return new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountDialog(props = {}) {
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
    global: { mocks: { $t: (k) => k } },
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

  async function open(props) {
    const wrapper = mountDialog(props);
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
});
