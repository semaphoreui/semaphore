import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import Hosts from '@/views/Hosts.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const rows = [
  {
    id: 1,
    name: 'web1',
    last_status: 'ok',
    templates: [],
    projects: [],
    inventories: [{ id: 3, name: 'prod' }, { id: 4, name: 'stage' }],
  },
];

describe('Hosts.vue', () => {
  let http;
  let wrapper;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => (c.url.endsWith('/hosts') ? rows : []));
  });

  afterEach(() => {
    if (wrapper) wrapper.unmount();
    http.restore();
  });

  function mountHosts(props) {
    wrapper = mount(Hosts, {
      // The table does not take rows that are still null
      data: () => ({ rows: [] }),
      props: { systemInfo: { features: { hosts: true } }, user: { admin: true }, ...props },
      global: {
        mocks: { $t: (k) => k, $route: { query: {} }, $router: { replace: () => {} } },
        stubs: {
          RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
          YesNoDialog: true,
        },
      },
    });
  }

  it('links the inventories of a host in a project', async () => {
    mountHosts({ projectId: 7 });
    await flush();

    const links = wrapper.get('[data-testid="hosts-rowInventories"]').findAll('a');
    expect(links.map((a) => a.attributes('href'))).to.deep.equal([
      '/project/7/inventory/3/hosts',
      '/project/7/inventory/4/hosts',
    ]);
  });

  it('has no inventories column on the page of all projects', async () => {
    mountHosts({});
    await flush();

    expect(wrapper.find('[data-testid="hosts-rowInventories"]').exists()).to.equal(false);
  });
});
