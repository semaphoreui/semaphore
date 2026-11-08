import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import McpInventoryForm from '@/components/McpInventoryForm.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  for (let i = 0; i < 3; i += 1) {
    // eslint-disable-next-line no-await-in-loop
    await new Promise((resolve) => { setTimeout(resolve, 0); });
  }
}

const inventory = {
  id: 3,
  project_id: 7,
  name: 'Agent MCP',
  type: 'mcp',
  inventory: '',
  mcp_servers: [
    {
      id: 5, inventory_id: 3, name: 'github', url: 'https://a.example.com/mcp', enabled: true, has_token: true,
    },
    {
      id: 6, inventory_id: 3, name: 'docs', url: 'https://b.example.com/mcp', enabled: true, has_token: true,
    },
    {
      id: 7, inventory_id: 3, name: 'search', url: 'https://c.example.com/mcp', enabled: false, has_token: false,
    },
  ],
};

// A fresh copy per response, as axios parses one: the form edits the item it loaded.
const loaded = () => JSON.parse(JSON.stringify(inventory));

function mountForm(itemId) {
  return mount(McpInventoryForm, {
    props: { projectId: 7, itemId },
    global: { mocks: { $t: (k) => k } },
  });
}

const rows = (wrapper) => wrapper.findAll('[data-testid="mcpInventoryForm-server"]');
const field = (row, name) => row.find(`[data-testid="mcpInventoryForm-server${name}"] input`);

describe('McpInventoryForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  it('never shows a stored token, only that a server has one', async () => {
    http.respond(() => loaded());
    const wrapper = mountForm(3);
    await flush();

    expect(rows(wrapper)).to.have.length(3);
    rows(wrapper).forEach((row) => {
      const token = field(row, 'Token');
      expect(token.element.value).to.equal('');
      expect(token.attributes('type')).to.equal('password');
      expect(token.attributes('autocomplete')).to.equal('new-password');
    });
    expect(field(rows(wrapper)[0], 'Token').attributes('placeholder')).to.equal('mcpTokenSet');
    expect(field(rows(wrapper)[2], 'Token').attributes('placeholder')).to.equal('mcpTokenOptional');
    // A server without a token has nothing to remove.
    expect(rows(wrapper)[2].find('[data-testid="mcpInventoryForm-serverTokenRemove"]').exists()).to.equal(false);
  });

  it('sends the servers with only the token changes the user made', async () => {
    http.respond((c) => (c.method === 'get' ? loaded() : ''));
    const wrapper = mountForm(3);
    await flush();

    // docs: remove its token; search: give it one; github: drop the server.
    await rows(wrapper)[1].find('[data-testid="mcpInventoryForm-serverTokenRemove"]').trigger('click');
    expect(rows(wrapper)[1].find('[data-testid="mcpInventoryForm-serverToken"]').attributes('data-state')).to.equal('remove');
    await field(rows(wrapper)[2], 'Token').setValue('tok-new');
    await rows(wrapper)[0].find('[data-testid="mcpInventoryForm-serverRemove"]').trigger('click');

    await wrapper.find('[data-testid="mcpInventoryForm-addServer"]').trigger('click');
    const added = rows(wrapper)[2];
    await field(added, 'Name').setValue('jira');
    await field(added, 'Url').setValue('https://d.example.com/mcp');

    await wrapper.vm.save();

    const put = http.requests.find((r) => r.method === 'put');
    expect(put.url).to.equal('/api/project/7/inventory/3');
    expect(put.data.type).to.equal('mcp');
    expect(put.data.mcp_servers).to.deep.equal([
      {
        id: 6, name: 'docs', url: 'https://b.example.com/mcp', enabled: true, token: { operation: 'delete' },
      },
      {
        id: 7,
        name: 'search',
        url: 'https://c.example.com/mcp',
        enabled: false,
        token: { operation: 'create', secret: 'tok-new' },
      },
      { name: 'jira', url: 'https://d.example.com/mcp', enabled: true },
    ]);
  });

  it('keeps a token whose removal was undone', async () => {
    http.respond((c) => (c.method === 'get' ? loaded() : ''));
    const wrapper = mountForm(3);
    await flush();

    const remove = rows(wrapper)[0].find('[data-testid="mcpInventoryForm-serverTokenRemove"]');
    await remove.trigger('click');
    await remove.trigger('click');
    await wrapper.vm.save();

    const put = http.requests.find((r) => r.method === 'put');
    expect(put.data.mcp_servers[0]).to.deep.equal({
      id: 5, name: 'github', url: 'https://a.example.com/mcp', enabled: true,
    });
  });

  it('creates an mcp inventory from a filled-in first server', async () => {
    http.respond(() => ({ id: 9 }));
    const wrapper = mountForm('new');
    await flush();

    expect(rows(wrapper)).to.have.length(1);
    await wrapper.find('[data-testid="mcpInventoryForm-name"] input').setValue('Agent MCP');
    await field(rows(wrapper)[0], 'Name').setValue('github');
    await field(rows(wrapper)[0], 'Url').setValue('https://a.example.com/mcp');
    await field(rows(wrapper)[0], 'Token').setValue('tok-1');

    await wrapper.vm.save();

    const post = http.requests.find((r) => r.method === 'post');
    expect(post.url).to.equal('/api/project/7/inventory');
    expect(post.data).to.include({ name: 'Agent MCP', type: 'mcp', project_id: 7 });
    expect(post.data.mcp_servers).to.deep.equal([{
      name: 'github', url: 'https://a.example.com/mcp', enabled: true, token: { operation: 'create', secret: 'tok-1' },
    }]);
  });

  [
    ['an invalid name', 'git hub', 'https://a.example.com/mcp'],
    ['a URL with credentials', 'github', 'https://user:pw@a.example.com/mcp'],
    ['a URL of another scheme', 'github', 'file:///etc/passwd'],
  ].forEach(([what, name, url]) => {
    it(`does not save ${what}`, async () => {
      http.respond(() => ({ id: 9 }));
      const wrapper = mountForm('new');
      await flush();

      await wrapper.find('[data-testid="mcpInventoryForm-name"] input').setValue('Agent MCP');
      await field(rows(wrapper)[0], 'Name').setValue(name);
      await field(rows(wrapper)[0], 'Url').setValue(url);
      await wrapper.vm.save();

      expect(http.requests.filter((r) => r.method === 'post')).to.have.length(0);
    });
  });

  it('does not save two servers of one name', async () => {
    http.respond((c) => (c.method === 'get' ? loaded() : ''));
    const wrapper = mountForm(3);
    await flush();

    await field(rows(wrapper)[1], 'Name').setValue('github');
    await wrapper.vm.save();

    expect(http.requests.filter((r) => r.method === 'put')).to.have.length(0);
  });
});
