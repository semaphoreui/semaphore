import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
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
    wrappers.splice(0).forEach((w) => w.destroy());
  });

  function mountWith(propsData) {
    const wrapper = mount(McpSettings, {
      localVue: createLocalVue(),
      vuetify: new Vuetify(),
      i18n,
      propsData,
    });
    wrappers.push(wrapper);
    return wrapper;
  }

  const off = {
    allow_run: false, allow_destroy: false, admin_allow_write: false, admin_allow_destroy: false,
  };

  it('shows a section per set, active actions first, then deleting', () => {
    expect(MCP_SETTINGS_SECTIONS.map((s) => [s.key, s.switches.map((w) => w.key)])).to.deep.equal([
      ['project', ['allow_run', 'allow_destroy']],
      ['admin', ['admin_allow_write', 'admin_allow_destroy']],
    ]);
    expect(MCP_SETTINGS_SWITCHES.map((s) => s.key))
      .to.deep.equal(['allow_run', 'allow_destroy', 'admin_allow_write', 'admin_allow_destroy']);
    const wrapper = mountWith({ value: off });
    const section = (key) => wrapper.find(`[data-testid="mcp-settings-section-${key}"]`);
    const labels = (key) => section(key).findAll('label').wrappers.map((l) => l.text());
    expect(section('project').find('.v-subheader').text()).to.equal('Project MCP');
    expect(labels('project')).to.deep.equal(['Run tasks', 'Destructive tools (delete)']);
    expect(section('admin').find('.v-subheader').text()).to.equal('Admin MCP');
    expect(labels('admin'))
      .to.deep.equal(['Changes (add users and other active actions)', 'Destructive tools (delete)']);
    expect(wrapper.find('[data-testid="mcp-settings-rest-note"]').text()).to.contain('REST API');
  });

  it('saves all four values at once and passes on what the server stored', async () => {
    http.respond((config) => config.data);
    const wrapper = mountWith({ value: off });

    wrapper.vm.draft.allow_run = true;
    wrapper.vm.draft.admin_allow_destroy = true;
    await wrapper.vm.save();

    const stored = { ...off, allow_run: true, admin_allow_destroy: true };
    const puts = http.requests.filter((r) => r.method === 'put');
    expect(puts.map((r) => [r.url, r.data])).to.deep.equal([['/api/mcp/settings', stored]]);
    expect(wrapper.emitted('input')).to.deep.equal([[stored]]);
    expect(snackbars.map((e) => e.color)).to.deep.equal(['success']);
    expect(wrapper.vm.saving).to.equal(false);
  });

  it('puts the switches back when saving fails', async () => {
    http.respond(() => { throw httpError(403, { error: 'Forbidden' }); });
    const wrapper = mountWith({ value: off });

    wrapper.vm.draft.admin_allow_write = true;
    await wrapper.vm.save();

    expect(wrapper.vm.draft).to.deep.equal(off);
    expect(wrapper.emitted('input')).to.equal(undefined);
    expect(snackbars).to.deep.equal([{ color: 'error', text: 'Forbidden' }]);
  });

  it('follows the values the page loads', async () => {
    const wrapper = mountWith({ value: null });
    expect(wrapper.findAll('input[type="checkbox"]').length, 'not loaded yet').to.equal(0);

    await wrapper.setProps({ value: { ...off, allow_run: true, admin_allow_write: true } });
    await flush();
    expect(wrapper.vm.draft).to.deep.equal({ ...off, allow_run: true, admin_allow_write: true });
    expect(wrapper.findAll('input[type="checkbox"]').length).to.equal(4);
  });

  it('sends every value even when the server answered without some', async () => {
    http.respond((config) => config.data);
    const wrapper = mountWith({ value: { allow_run: true } });
    expect(wrapper.vm.draft).to.deep.equal({ ...off, allow_run: true });

    await wrapper.vm.save();
    expect(http.requests.filter((r) => r.method === 'put').map((r) => r.data))
      .to.deep.equal([{ ...off, allow_run: true }]);
  });

  it('says why the settings did not load', () => {
    const wrapper = mountWith({ value: null, loadError: 'database is down' });
    expect(wrapper.find('[data-testid="mcp-settings-load-error"]').text()).to.contain('database is down');
  });
});
