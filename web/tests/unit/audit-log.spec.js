import './setup';
import { expect } from 'chai';
import axios from 'axios';
import { mount, createLocalVue } from '@vue/test-utils';
import VueRouter from 'vue-router';
import Vuetify from 'vuetify';
import i18n from '@/plugins/i18';
import EventBus from '@/event-bus';
import AuditLog from '@/views/AuditLog.vue';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';
import mockAxios from './helpers/axiosMock';

function auditEvent(seq, fields = {}) {
  return {
    event_id: `e${seq}`,
    seq,
    timestamp: '2026-10-06T10:00:00.000Z',
    category: 'auth',
    event_code: 'login',
    action: 'authenticate',
    outcome: 'success',
    reason: '',
    actor: {
      type: 'user', id: '2', name: 'alice', auth: 'session',
    },
    source: { ip: '203.0.113.7', user_agent: 'curl/8' },
    scope: { project_id: '9' },
    instance_id: 'semaphore-1',
    metadata: {},
    ...fields,
  };
}

const PAGE = { events: [auditEvent(80), auditEvent(79)], older: 79, newer: 80 };

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

describe('AuditLog.vue', () => {
  let http;
  const wrappers = [];

  function respondWith(page) {
    http.respond((config) => {
      if (config.url === '/api/users') {
        return [{ id: 2, name: 'Alice', username: 'alice' }];
      }
      if (config.url === '/api/projects') {
        return [{ id: 3, name: 'Infra' }];
      }
      return page;
    });
  }

  beforeEach(() => {
    http = mockAxios();
    respondWith(PAGE);
  });

  afterEach(() => {
    http.restore();
    while (wrappers.length) {
      wrappers.pop().destroy();
    }
    document.querySelectorAll('[data-app]').forEach((el) => el.remove());
  });

  async function mountPage(features = { audit_log_filters: true }, width = 1280) {
    const app = document.createElement('div');
    app.setAttribute('data-app', 'true');
    document.body.appendChild(app);
    const mountPoint = document.createElement('div');
    app.appendChild(mountPoint);
    const innerWidth = window.innerWidth;
    window.innerWidth = width;
    const vuetify = new Vuetify();
    const localVue = createLocalVue();
    localVue.use(VueRouter);
    const wrapper = mount(AuditLog, {
      localVue,
      router: new VueRouter(),
      vuetify,
      i18n,
      attachTo: mountPoint,
      propsData: { systemInfo: { features } },
    });
    wrappers.push(wrapper);
    await flush();
    window.innerWidth = innerWidth;
    return wrapper;
  }

  function auditRequests() {
    return http.requests.filter((r) => r.url === '/api/audit/events');
  }

  function lastQuery() {
    const request = auditRequests().pop();
    return axios.getUri(request).split('?')[1] || '';
  }

  it('opens on the latest page', async () => {
    await mountPage();
    expect(auditRequests()).to.have.length(1);
    expect(lastQuery()).to.equal('');
  });

  it('pages older, newer and back to the latest', async () => {
    const wrapper = await mountPage();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    await flush();
    expect(lastQuery()).to.equal('before=79');
    await wrapper.find('[data-testid="audit-newer"]').trigger('click');
    await flush();
    expect(lastQuery()).to.equal('after=80');
    await wrapper.find('[data-testid="audit-latest"]').trigger('click');
    expect(lastQuery()).to.equal('');
  });

  it('disables paging past either end', async () => {
    respondWith({ events: [auditEvent(1)], older: null, newer: null });
    const wrapper = await mountPage();
    expect(wrapper.find('[data-testid="audit-older"]').attributes('disabled')).to.equal('disabled');
    expect(wrapper.find('[data-testid="audit-newer"]').attributes('disabled')).to.equal('disabled');
  });

  it('sends several kinds as repeated parameters', async () => {
    const wrapper = await mountPage();
    wrapper.vm.setFilters({ kind: ['auth.login/authenticate', 'iam.role/delete'] });
    await flush();
    expect(decodeURIComponent(lastQuery())).to.equal('kind=auth.login/authenticate&kind=iam.role/delete');
  });

  it('keeps filters while paging', async () => {
    const wrapper = await mountPage();
    wrapper.vm.setFilters({ outcome: 'failure' });
    await flush();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    expect(lastQuery()).to.equal('outcome=failure&before=79');
  });

  it('sends the period in UTC', async () => {
    const wrapper = await mountPage();
    wrapper.vm.setFilters({ from: '2026-10-01T00:00' });
    await flush();
    const from = new URLSearchParams(lastQuery()).get('from');
    expect(from).to.equal(new Date('2026-10-01T00:00').toISOString());
  });

  it('filters by the IP clicked in the card', async () => {
    const wrapper = await mountPage();
    await wrapper.findAll('tbody tr').at(0).trigger('click');
    await flush();
    await wrapper.find('[data-testid="audit-pivot-ip"]').trigger('click');
    await flush();
    expect(lastQuery()).to.equal('ip=203.0.113.7');
  });

  it('filters by the project clicked in the card', async () => {
    const wrapper = await mountPage();
    await wrapper.findAll('tbody tr').at(0).trigger('click');
    await flush();
    await wrapper.find('[data-testid="audit-pivot-project"]').trigger('click');
    await flush();
    expect(lastQuery()).to.equal('project=9');
  });

  it('offers the event JSON for copying in the card', async () => {
    const wrapper = await mountPage();
    await wrapper.findAll('tbody tr').at(0).trigger('click');
    await flush();
    const copy = wrapper.findComponent(CopyClipboardButton);
    expect(copy.exists()).to.equal(true);
    expect(JSON.parse(copy.props('text')).event_id).to.equal('e80');
  });

  it('shows the Pro notice and disables filters, export and pivots in Community', async () => {
    const wrapper = await mountPage({});
    expect(wrapper.find('[data-testid="audit-pro-notice"]').exists()).to.equal(true);
    expect(wrapper.find('[data-testid="audit-export"]').attributes('disabled')).to.equal('disabled');
    const inputs = wrapper.findAll('[data-testid="audit-filters"] input:not([type="hidden"])');
    expect(inputs.length).to.be.greaterThan(0);
    inputs.wrappers.forEach((input) => expect(input.attributes('disabled')).to.equal('disabled'));
    await wrapper.findAll('tbody tr').at(0).trigger('click');
    await flush();
    expect(wrapper.find('[data-testid="audit-pivot-ip"]').exists()).to.equal(false);
  });

  it('shows a second paging bar only on a narrow screen', async () => {
    const wide = await mountPage();
    expect(wide.find('[data-testid="audit-older-top"]').exists()).to.equal(false);
    const tablet = await mountPage({ audit_log_filters: true }, 800);
    expect(tablet.find('[data-testid="audit-older-top"]').exists()).to.equal(false);
    expect(tablet.classes()).to.include('AuditLog--fixed');
    const narrow = await mountPage({ audit_log_filters: true }, 500);
    await narrow.find('[data-testid="audit-older-top"]').trigger('click');
    expect(lastQuery()).to.equal('before=79');
  });

  it('hides the Pro notice in Pro', async () => {
    const wrapper = await mountPage();
    expect(wrapper.find('[data-testid="audit-pro-notice"]').exists()).to.equal(false);
  });

  it('names a deleted project by its ID', async () => {
    const wrapper = await mountPage();
    expect(wrapper.vm.projectName(auditEvent(1, { scope: { project_id: '3' } }))).to.equal('Infra');
    expect(wrapper.vm.projectName(auditEvent(1))).to.equal('Project #9 (deleted)');
    expect(wrapper.vm.projectName(auditEvent(1, { scope: undefined }))).to.equal('');
  });

  it('shows no project name until the names have loaded', async () => {
    http.respond(() => new Promise(() => {}));
    const wrapper = await mountPage();
    expect(wrapper.vm.projectName(auditEvent(1))).to.equal('');
  });

  it('builds the export link from the filters', async () => {
    const wrapper = await mountPage();
    wrapper.vm.setFilters({ user: 2, kind: ['iam.role/delete'] });
    const url = axios.getUri({
      url: '/api/audit/events/export',
      params: { user: 2, kind: ['iam.role/delete'], format: 'csv' },
      paramsSerializer: { indexes: null },
    });
    expect(wrapper.vm.exportUrl('csv')).to.equal(url);
    expect(decodeURIComponent(url)).to.equal('/api/audit/events/export?user=2&kind=iam.role/delete&format=csv');
  });

  it('keeps the newest response when an older one arrives late', async () => {
    const wrapper = await mountPage();
    const pending = [];
    http.respond((config) => {
      if (config.url !== '/api/audit/events') {
        return [];
      }
      return new Promise((resolve) => { pending.push(resolve); });
    });
    wrapper.vm.load({});
    wrapper.vm.load({});
    pending[1]({ events: [auditEvent(2)], older: null, newer: null });
    await flush();
    pending[0]({ events: [auditEvent(1)], older: null, newer: null });
    await flush();
    expect(wrapper.vm.events.map((e) => e.seq)).to.deep.equal([2]);
    expect(wrapper.vm.loading).to.equal(false);
  });

  it('disables paging while loading', async () => {
    const wrapper = await mountPage();
    http.respond(() => new Promise(() => {}));
    wrapper.vm.load({});
    await flush();
    expect(wrapper.find('[data-testid="audit-older"]').attributes('disabled')).to.equal('disabled');
  });

  it('shows the project ID when the project list fails to load', async () => {
    http.respond((config) => {
      if (config.url === '/api/projects') {
        throw new Error('boom');
      }
      return config.url === '/api/users' ? [] : PAGE;
    });
    const wrapper = await mountPage();
    expect(wrapper.vm.projectName(auditEvent(1))).to.equal('Project #9');
  });

  it('reports a failed events request and keeps the list', async () => {
    const wrapper = await mountPage();
    const messages = [];
    const listener = (m) => messages.push(m);
    EventBus.$on('i-snackbar', listener);
    http.respond(() => { throw new Error('boom'); });
    await wrapper.vm.load({});
    EventBus.$off('i-snackbar', listener);
    expect(messages).to.have.length(1);
    expect(messages[0].color).to.equal('error');
    expect(wrapper.vm.events).to.have.length(2);
  });
});
