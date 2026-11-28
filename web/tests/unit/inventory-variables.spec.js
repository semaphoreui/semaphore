import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import EventBus from '@/event-bus';
import InventoryVariables from '@/views/project/InventoryVariables.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const sources = [
  { id: 5, name: 'Edge devices', variables: 'region: edge' },
  { id: 2, name: 'Static servers', variables: 'ntp: pool.ntp.org' },
  { id: 4, name: 'Empty', variables: '' },
];

function mountVariables() {
  return mount(InventoryVariables, {
    props: { projectId: 7, inventoryId: 3, isAdmin: true },
    global: {
      mocks: { $t: (k) => k },
      stubs: { CodeMirror: true, YesNoDialog: true },
    },
  });
}

describe('InventoryVariables.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => {
      if (c.method !== 'get') return {};
      if (c.url.endsWith('/sources')) return sources;
      return {
        id: 3, name: 'prod', type: 'managed', ssh_key_id: 1, variables: 'a: 1',
      };
    });
  });

  afterEach(() => {
    http.restore();
  });

  it('saves only the variables', async () => {
    const wrapper = mountVariables();
    await flush();
    wrapper.vm.inventory.variables = 'a: 2';
    const saved = [];
    const onSaved = (e) => saved.push(e);
    EventBus.$on('i-inventory-variables', onSaved);

    await wrapper.find('[data-testid="inventoryVariables-save"]').trigger('click');
    await flush();

    const put = http.requests.find((r) => r.method === 'put');
    expect(put.url).to.equal('/api/project/7/inventory/3/variables');
    EventBus.$off('i-inventory-variables', onSaved);
    expect(put.data).to.deep.equal({ variables: 'a: 2' });
    expect(saved).to.deep.equal([{ projectId: 7, inventoryId: 3, variables: 'a: 2' }]);
  });

  it('shows the variables of each source read-only below the editor in source order', async () => {
    const wrapper = mountVariables();
    await flush();

    const blocks = wrapper.findAll('[data-testid^="inventoryVariables-source-"]');
    expect(blocks.map((b) => b.attributes('data-testid'))).to.deep.equal([
      'inventoryVariables-source-2', 'inventoryVariables-source-5',
    ]);
    expect(blocks[0].text()).to.contain('inventory_variables_from_source');
    const editor = blocks[0].findComponent({ name: 'CodeMirror' });
    expect(editor.props('modelValue')).to.equal('ntp: pool.ntp.org');
    expect(editor.props('readonly')).to.equal(true);
    expect(editor.classes()).to.include('InventoryVariables__source');
    expect(blocks[0].find('.mdi-lock').exists()).to.equal(true);
    expect(wrapper.find('[data-testid="inventoryVariables-editor"]').classes())
      .to.not.include('InventoryVariables__source');
  });

  it('asks before leaving with unsaved variables', async () => {
    const wrapper = mountVariables();
    await flush();
    const leave = (next) => InventoryVariables.beforeRouteLeave.call(wrapper.vm, {}, {}, next);
    const answers = [];

    leave((v) => answers.push(v));
    wrapper.vm.inventory.variables = 'a: 2';
    leave((v) => answers.push(v));
    await flush();
    expect(wrapper.vm.leaveDialog).to.equal(true);
    wrapper.vm.leaveDialog = false;
    await flush();
    leave((v) => answers.push(v));
    wrapper.vm.confirmLeave();

    expect(answers).to.deep.equal([undefined, false, undefined]);
  });
});
