import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import InventoryForm from '@/components/InventoryForm.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountForm(props) {
  return mount(InventoryForm, {
    props: { projectId: 7, features: {}, ...props },
    global: {
      mocks: { $t: (k) => k },
      stubs: { CodeMirror: true, RichEditor: true },
    },
  });
}

describe('InventoryForm.vue managed type', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  it('hides the text editor and the repository for the managed type', async () => {
    http.respond(() => []);
    const wrapper = mountForm({ itemId: 'new' });
    await flush();
    wrapper.vm.item.type = 'static';
    await flush();
    expect(wrapper.find('[data-testid="inventoryForm-inventoryCode"]').exists()).to.equal(true);

    wrapper.vm.item.type = 'managed';
    await flush();

    expect(wrapper.find('[data-testid="inventoryForm-inventoryCode"]').exists()).to.equal(false);
    expect(wrapper.find('[data-testid="inventoryForm-repositoryId"]').exists()).to.equal(false);
  });

  it('disables the type of an existing managed inventory', async () => {
    http.respond((c) => (c.url.endsWith('/inventory/3') ? { id: 3, name: 'prod', type: 'managed' } : []));
    const wrapper = mountForm({ itemId: 3 });
    await flush();

    const type = wrapper.find('[data-testid="inventoryForm-type"]');
    expect(type.classes()).to.include('v-input--disabled');
    expect(wrapper.vm.inventoryTypes.map((t) => t.id)).to.include('managed');
  });

  it('does not offer managed for an existing inventory of another type', async () => {
    http.respond((c) => (c.url.endsWith('/inventory/4') ? { id: 4, name: 'old', type: 'static' } : []));
    const wrapper = mountForm({ itemId: 4 });
    await flush();

    expect(wrapper.vm.inventoryTypes.map((t) => t.id)).to.not.include('managed');
  });
});
