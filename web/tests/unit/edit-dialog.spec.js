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
});
