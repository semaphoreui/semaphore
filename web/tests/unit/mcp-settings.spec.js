import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import i18n from '@/plugins/i18';
import EventBus from '@/event-bus';
import McpSettings, { MCP_SETTINGS_SECTIONS, MCP_SETTINGS_SWITCHES } from '@/components/McpSettings.vue';
import mockAxios, { httpError } from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

describe('McpSettings.vue', () => {
  let http;
  let snackbars;
  const onSnackbar = (e) => snackbars.push(e);
  const wrappers = [];

  beforeEach(() => {
    http = mockAxios();
    snackbars = [];
    EventBus.$on('i-snackbar', onSnackbar);
  });

  afterEach(() => {
    http.restore();
    EventBus.$off('i-snackbar', onSnackbar);
    wrappers.splice(0).forEach((w) => w.unmount());
  });

  function mountWith(props) {
    const wrapper = mount(McpSettings, {
      props,
      global: { plugins: [i18n] },
    });
    wrappers.push(wrapper);
    return wrapper;
  }

  const off = {
    allow_write: false,
    allow_run: false,
    allow_destroy: false,
    admin_allow_write: false,
    admin_allow_destroy: false,
  };

  it('shows a section per set, in the order the user chose: changes, deleting, running', () => {
    expect(MCP_SETTINGS_SECTIONS.map((s) => [s.key, s.switches.map((w) => w.key)])).to.deep.equal([
      ['project', ['allow_write', 'allow_destroy', 'allow_run']],
      ['admin', ['admin_allow_write', 'admin_allow_destroy']],
    ]);
    expect(MCP_SETTINGS_SWITCHES.map((s) => s.key))
      .to.deep.equal(['allow_write', 'allow_destroy', 'allow_run', 'admin_allow_write', 'admin_allow_destroy']);
    const wrapper = mountWith({ modelValue: off });
    const section = (key) => wrapper.find(`[data-testid="mcp-settings-section-${key}"]`);
    const labels = (key) => section(key).findAll('label').map((l) => l.text());
    expect(section('project').find('[data-testid="mcp-settings-section-title"]').text()).to.equal('Project MCP');
    expect(labels('project'))
      .to.deep.equal(['Allow the change tools', 'Allow the delete tools', 'Allow the task run tools']);
    expect(section('project').text(), 'the hint of the project change tools')
      .to.contain('create and change projects').and.contain('generate SSH keys');
    expect(section('admin').find('[data-testid="mcp-settings-section-title"]').text()).to.equal('Admin MCP');
    expect(labels('admin'))
      .to.deep.equal(['Allow the change tools', 'Allow the delete tools']);
    expect(wrapper.text(), 'no note about REST').to.not.contain('REST API');
    expect(wrapper.find('[data-testid="mcp-settings-scope-note"]').text())
      .to.equal('Global setting of tool availability for every user.');
  });

  it('saves all five values at once and passes on what the server stored', async () => {
    http.respond((config) => config.data);
    const wrapper = mountWith({ modelValue: off });

    wrapper.vm.draft.allow_write = true;
    wrapper.vm.draft.allow_run = true;
    wrapper.vm.draft.admin_allow_destroy = true;
    await wrapper.vm.save();

    const stored = {
      ...off, allow_write: true, allow_run: true, admin_allow_destroy: true,
    };
    const puts = http.requests.filter((r) => r.method === 'put');
    expect(puts.map((r) => [r.url, r.data])).to.deep.equal([['/api/mcp/settings', stored]]);
    expect(wrapper.emitted('update:modelValue')).to.deep.equal([[stored]]);
    expect(snackbars.map((e) => e.color)).to.deep.equal(['success']);
    expect(wrapper.vm.saving).to.equal(false);
  });

  it('saves as soon as a switch is toggled', async () => {
    http.respond((config) => config.data);
    const wrapper = mountWith({ modelValue: off });

    const input = wrapper.find('[data-testid="mcp-settings-allow_destroy"] input');
    input.element.checked = true;
    await input.trigger('input');
    await flush();

    expect(http.requests.filter((r) => r.method === 'put').map((r) => r.data))
      .to.deep.equal([{ ...off, allow_destroy: true }]);
    expect(wrapper.emitted('update:modelValue')).to.deep.equal([[{ ...off, allow_destroy: true }]]);
  });

  it('saves the project change switch apart from the admin one', async () => {
    http.respond((config) => config.data);
    const wrapper = mountWith({ modelValue: off });

    const input = wrapper.find('[data-testid="mcp-settings-allow_write"] input');
    input.element.checked = true;
    await input.trigger('input');
    await flush();

    expect(http.requests.filter((r) => r.method === 'put').map((r) => r.data))
      .to.deep.equal([{ ...off, allow_write: true }]);
  });

  it('puts the switches back when saving fails', async () => {
    http.respond(() => { throw httpError(403, { error: 'Forbidden' }); });
    const wrapper = mountWith({ modelValue: off });

    wrapper.vm.draft.admin_allow_write = true;
    await wrapper.vm.save();

    expect(wrapper.vm.draft).to.deep.equal(off);
    expect(wrapper.emitted('update:modelValue')).to.equal(undefined);
    expect(snackbars).to.deep.equal([{ color: 'error', text: 'Forbidden' }]);
  });

  it('follows the values the page loads', async () => {
    const wrapper = mountWith({ modelValue: null });
    expect(wrapper.findAll('input[type="checkbox"]').length, 'not loaded yet').to.equal(0);

    await wrapper.setProps({ modelValue: { ...off, allow_run: true, admin_allow_write: true } });
    await flush();
    expect(wrapper.vm.draft).to.deep.equal({ ...off, allow_run: true, admin_allow_write: true });
    expect(wrapper.findAll('input[type="checkbox"]').length).to.equal(5);
  });

  it('sends every value even when the server answered without some', async () => {
    http.respond((config) => config.data);
    const wrapper = mountWith({ modelValue: { allow_run: true } });
    expect(wrapper.vm.draft).to.deep.equal({ ...off, allow_run: true });

    await wrapper.vm.save();
    expect(http.requests.filter((r) => r.method === 'put').map((r) => r.data))
      .to.deep.equal([{ ...off, allow_run: true }]);
  });

  it('says why the settings did not load', () => {
    const wrapper = mountWith({ modelValue: null, loadError: 'database is down' });
    expect(wrapper.find('[data-testid="mcp-settings-load-error"]').text()).to.contain('database is down');
  });
});
