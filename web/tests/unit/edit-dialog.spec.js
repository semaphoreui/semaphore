import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import EditDialog from '@/components/EditDialog.vue';

function pressEscape() {
  window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', keyCode: 27, bubbles: true }));
}

function mountDialog(propsData = {}) {
  // Vuetify detaches dialog content into the [data-app] element. Mount into a
  // child node, because attachTo replaces the target element itself.
  const app = document.createElement('div');
  app.setAttribute('data-app', 'true');
  document.body.appendChild(app);
  const mountPoint = document.createElement('div');
  app.appendChild(mountPoint);

  const localVue = createLocalVue();
  return mount(EditDialog, {
    localVue,
    vuetify: new Vuetify(),
    attachTo: mountPoint,
    propsData: { title: 'Test', ...propsData },
    mocks: { $t: (k) => k },
  });
}

describe('EditDialog.vue', () => {
  const wrappers = [];

  afterEach(() => {
    while (wrappers.length) {
      wrappers.pop().destroy();
    }
    document.querySelectorAll('[data-app]').forEach((el) => el.remove());
  });

  async function open(propsData) {
    const wrapper = mountDialog(propsData);
    wrappers.push(wrapper);
    await wrapper.setProps({ value: true });
    await wrapper.vm.$nextTick();
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

    pressEscape();
    await outer.vm.$nextTick();

    expect(outer.vm.dialog, 'outer dialog closes on the next Escape').to.equal(false);
  });
});
