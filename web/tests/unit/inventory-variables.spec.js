import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import InventoryVariables from '@/views/project/InventoryVariables.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

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
    http.respond((c) => (c.method === 'get'
      ? {
        id: 3, name: 'prod', type: 'managed', ssh_key_id: 1, variables: 'a: 1',
      }
      : {}));
  });

  afterEach(() => {
    http.restore();
  });

  it('saves only the variables', async () => {
    const wrapper = mountVariables();
    await flush();
    wrapper.vm.inventory.variables = 'a: 2';

    await wrapper.find('[data-testid="inventoryVariables-save"]').trigger('click');
    await flush();

    const put = http.requests.find((r) => r.method === 'put');
    expect(put.url).to.equal('/api/project/7/inventory/3/variables');
    expect(put.data).to.deep.equal({ variables: 'a: 2' });
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
