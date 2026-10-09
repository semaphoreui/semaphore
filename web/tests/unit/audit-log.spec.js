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
    event_code: 'auth.login',
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

function searched(seq, older = seq) {
  return {
    events: [],
    older,
    newer: null,
    searched_to: { seq, timestamp: `2026-09-${String(1 + (seq % 28)).padStart(2, '0')}T12:00:00.000000Z` },
  };
}

function searchedNewer(seq, newer = seq) {
  return { ...searched(seq, null), newer };
}

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

  // Audit requests get the pages in turn, the last one repeats.
  function respondInTurn(...pages) {
    let next = 0;
    http.respond((config) => {
      if (config.url !== '/api/audit/events') {
        return config.url === '/api/users' ? [] : [{ id: 3, name: 'Infra' }];
      }
      const page = pages[Math.min(next, pages.length - 1)];
      next += 1;
      return page;
    });
  }

  // Audit requests wait until resolved and fail like axios when aborted.
  function respondLater() {
    const pending = [];
    http.respond((config) => {
      if (config.url !== '/api/audit/events') {
        return [];
      }
      return new Promise((resolve, reject) => {
        pending.push(resolve);
        config.signal.addEventListener('abort', () => reject(new axios.CanceledError()));
      });
    });
    return pending;
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

  async function mountPage(
    features = { audit_log_filters: true },
    width = 1280,
    user = { can_read_audit_log: true },
  ) {
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
      propsData: { systemInfo: { features }, user },
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

  it('offers to search older events when the search stopped', async () => {
    respondInTurn(searched(400), PAGE);
    const wrapper = await mountPage();
    expect(auditRequests()).to.have.length(1);
    expect(wrapper.find('[data-testid="audit-searched-to"]').text())
      .to.contain('No more matches among events since');
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    await flush();
    expect(lastQuery()).to.equal('before=400');
  });

  it('keeps searching older until a page has events', async () => {
    respondInTurn(searched(400), searched(300), searched(200), PAGE);
    const wrapper = await mountPage();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    await flush();
    expect(auditRequests().map((r) => r.params.before)).to.deep.equal([undefined, 400, 300, 200]);
    expect(wrapper.findAll('tbody tr')).to.have.length(2);
    expect(wrapper.find('[data-testid="audit-searching"]').exists()).to.equal(false);
    expect(wrapper.vm.loading).to.equal(false);
  });

  it('stops searching older at the end of the log', async () => {
    respondInTurn(searched(400), searched(300, null), PAGE);
    const wrapper = await mountPage();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    await flush();
    expect(auditRequests()).to.have.length(2);
    expect(wrapper.vm.loading).to.equal(false);
  });

  it('keeps searching older with Older', async () => {
    respondInTurn(PAGE, searched(400), searched(300), PAGE);
    const wrapper = await mountPage();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    await flush();
    expect(auditRequests().map((r) => r.params.before)).to.deep.equal([undefined, 79, 400, 300]);
    expect(wrapper.findAll('tbody tr')).to.have.length(2);
  });

  it('keeps searching newer until a page has events', async () => {
    respondInTurn(PAGE, searchedNewer(200), searchedNewer(300), PAGE);
    const wrapper = await mountPage();
    await wrapper.find('[data-testid="audit-newer"]').trigger('click');
    await flush();
    expect(auditRequests().map((r) => r.params.after)).to.deep.equal([undefined, 80, 200, 300]);
    expect(wrapper.findAll('tbody tr')).to.have.length(2);
    expect(wrapper.find('[data-testid="audit-searching"]').exists()).to.equal(false);
    expect(wrapper.vm.loading).to.equal(false);
  });

  it('shows how far a newer search reached and offers to go on after Stop', async () => {
    const wrapper = await mountPage();
    const pending = respondLater();
    await wrapper.find('[data-testid="audit-newer"]').trigger('click');
    pending[0](searchedNewer(310));
    await flush();
    expect(wrapper.find('[data-testid="audit-searching"]').text()).to.contain('Searching newer events');
    const inFlight = auditRequests().pop();
    await wrapper.find('[data-testid="audit-stop"]').trigger('click');
    await flush();
    expect(inFlight.signal.aborted).to.equal(true);
    const line = wrapper.find('[data-testid="audit-searched-to"]').text();
    expect(line).to.contain('No more matches among events until');
    expect(line).to.contain('2026-09-03');
    await wrapper.find('[data-testid="audit-newer"]').trigger('click');
    expect(lastQuery()).to.equal('after=310');
  });

  it('ends a newer search when the filters change', async () => {
    const wrapper = await mountPage();
    const pending = respondLater();
    await wrapper.find('[data-testid="audit-newer"]').trigger('click');
    pending[0](searchedNewer(310));
    await flush();
    const searching = auditRequests().pop();
    wrapper.vm.setFilters({ outcome: 'failure' });
    expect(searching.signal.aborted).to.equal(true);
    pending[2](PAGE);
    await flush();
    expect(auditRequests()).to.have.length(4);
    expect(lastQuery()).to.equal('outcome=failure');
    expect(wrapper.find('[data-testid="audit-searching"]').exists()).to.equal(false);
  });

  it('keeps the stopped line of the shown page until a newer page arrives', async () => {
    respondInTurn({ ...searched(400), newer: 500 });
    const wrapper = await mountPage();
    http.respond(() => { throw new Error('boom'); });
    await wrapper.find('[data-testid="audit-newer"]').trigger('click');
    await flush();
    expect(wrapper.find('[data-testid="audit-searched-to"]').text()).to.contain('since');
    respondInTurn(searchedNewer(600), PAGE);
    await wrapper.find('[data-testid="audit-retry"]').trigger('click');
    await flush();
    expect(auditRequests().map((r) => r.params.after)).to.deep.equal([undefined, 500, 500, 600]);
  });

  it('says no older matches at the end of the log', async () => {
    respondInTurn(searched(400), searched(300, null));
    const wrapper = await mountPage();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    await flush();
    expect(wrapper.text()).to.contain('No older matches.');
    expect(wrapper.text()).to.not.contain('No audit events match');
  });

  it('says no newer matches on an empty newer page', async () => {
    respondInTurn(PAGE, { events: [], older: 81, newer: null });
    const wrapper = await mountPage();
    await wrapper.find('[data-testid="audit-newer"]').trigger('click');
    await flush();
    expect(auditRequests()).to.have.length(2);
    expect(wrapper.text()).to.contain('No newer matches.');
    expect(wrapper.text()).to.not.contain('No audit events match');
  });

  it('says no audit events match on an empty first page', async () => {
    respondWith({ events: [], older: null, newer: null });
    const wrapper = await mountPage();
    expect(wrapper.text()).to.contain('No audit events match.');
  });

  it('shows how far the search reached and stops on Stop', async () => {
    respondInTurn(searched(400));
    const wrapper = await mountPage();
    const pending = respondLater();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    pending[0](searched(310));
    await flush();
    const line = wrapper.find('[data-testid="audit-searching"]');
    expect(line.text()).to.contain('Searching older events');
    expect(line.text()).to.contain('2026-09-03');
    const inFlight = auditRequests().pop();
    await wrapper.find('[data-testid="audit-stop"]').trigger('click');
    await flush();
    expect(inFlight.signal.aborted).to.equal(true);
    expect(auditRequests()).to.have.length(3);
    expect(wrapper.find('[data-testid="audit-searching"]').exists()).to.equal(false);
    expect(wrapper.find('[data-testid="audit-searched-to"]').text()).to.contain('2026-09-03');
    const older = wrapper.find('[data-testid="audit-older"]');
    expect(older.attributes('disabled')).to.equal(undefined);
    expect(wrapper.vm.loading).to.equal(false);
  });

  it('reports no error for a cancelled request', async () => {
    respondInTurn(searched(400));
    const wrapper = await mountPage();
    const pending = respondLater();
    const messages = [];
    const listener = (m) => messages.push(m);
    EventBus.$on('i-snackbar', listener);
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    pending[0](searched(310));
    await flush();
    await wrapper.find('[data-testid="audit-stop"]').trigger('click');
    await flush();
    EventBus.$off('i-snackbar', listener);
    expect(messages).to.have.length(0);
    expect(wrapper.vm.failed).to.equal(false);
  });

  it('ends the search when the filters change', async () => {
    respondInTurn(searched(400));
    const wrapper = await mountPage();
    const pending = respondLater();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    const searching = auditRequests().pop();
    wrapper.vm.setFilters({ outcome: 'failure' });
    expect(searching.signal.aborted).to.equal(true);
    pending[1](PAGE);
    pending[0](searched(300));
    await flush();
    expect(auditRequests()).to.have.length(3);
    expect(lastQuery()).to.equal('outcome=failure');
    expect(wrapper.findAll('tbody tr')).to.have.length(2);
  });

  it('ends the search when leaving the page', async () => {
    respondInTurn(searched(400));
    const wrapper = await mountPage();
    const pending = respondLater();
    await wrapper.find('[data-testid="audit-older"]').trigger('click');
    const searching = auditRequests().pop();
    wrappers.pop().destroy();
    expect(searching.signal.aborted).to.equal(true);
    pending[0](searched(300));
    await flush();
    expect(auditRequests()).to.have.length(2);
  });

  it('hides the search line on an ordinary page', async () => {
    const wrapper = await mountPage();
    expect(wrapper.find('[data-testid="audit-searched-to"]').exists()).to.equal(false);
  });

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

  it('reloads the latest page from the latest page', async () => {
    respondWith({ events: [auditEvent(1)], older: null, newer: null });
    const wrapper = await mountPage();
    wrapper.vm.setFilters({ outcome: 'failure' });
    await flush();
    const latest = wrapper.find('[data-testid="audit-latest"]');
    expect(latest.attributes('disabled')).to.equal(undefined);
    await latest.trigger('click');
    await flush();
    expect(auditRequests()).to.have.length(3);
    expect(lastQuery()).to.equal('outcome=failure');
  });

  it('empties the list when a request for new filters fails', async () => {
    const wrapper = await mountPage();
    http.respond(() => { throw new Error('boom'); });
    wrapper.vm.setFilters({ ip: '10.0.0.1' });
    await flush();
    expect(wrapper.vm.events).to.deep.equal([]);
    expect(wrapper.vm.older).to.equal(null);
    expect(wrapper.vm.newer).to.equal(null);
    expect(wrapper.vm.searchedTo).to.equal(null);
  });

  it('shows an error with Retry instead of no matches when new filters fail', async () => {
    const wrapper = await mountPage();
    http.respond(() => { throw new Error('boom'); });
    wrapper.vm.setFilters({ ip: '10.0.0.1' });
    await flush();
    expect(wrapper.text()).to.not.contain('No audit events match');
    expect(wrapper.find('[data-testid="audit-load-failed"]').exists()).to.equal(true);
    respondWith(PAGE);
    await wrapper.find('[data-testid="audit-retry"]').trigger('click');
    await flush();
    expect(lastQuery()).to.equal('ip=10.0.0.1');
    expect(wrapper.find('[data-testid="audit-load-failed"]').exists()).to.equal(false);
    expect(wrapper.findAll('tbody tr')).to.have.length(2);
  });

  it('shows an error with Retry when the first load fails', async () => {
    http.respond((config) => {
      if (config.url === '/api/audit/events') {
        throw new Error('boom');
      }
      return [];
    });
    const wrapper = await mountPage();
    expect(wrapper.text()).to.not.contain('No audit events match');
    expect(wrapper.find('[data-testid="audit-retry"]').exists()).to.equal(true);
  });

  it('repeats the last page request on Retry', async () => {
    const wrapper = await mountPage();
    http.respond(() => { throw new Error('boom'); });
    wrapper.vm.events = [];
    await wrapper.vm.load({ before: 79 });
    respondWith(PAGE);
    await wrapper.find('[data-testid="audit-retry"]').trigger('click');
    await flush();
    expect(lastQuery()).to.equal('before=79');
  });

  it('keeps the error and Retry above the rows when a page change fails', async () => {
    const wrapper = await mountPage();
    http.respond(() => { throw new Error('boom'); });
    await wrapper.vm.load({ before: 79 });
    await flush();
    expect(wrapper.findAll('tbody tr')).to.have.length(2);
    expect(wrapper.find('[data-testid="audit-load-failed"]').exists()).to.equal(true);
    respondWith(PAGE);
    await wrapper.find('[data-testid="audit-retry"]').trigger('click');
    await flush();
    expect(lastQuery()).to.equal('before=79');
    expect(wrapper.find('[data-testid="audit-load-failed"]').exists()).to.equal(false);
  });

  it('downloads the export instead of opening it', async () => {
    const wrapper = await mountPage();
    await wrapper.find('[data-testid="audit-export"]').trigger('click');
    await flush();
    const links = document.querySelectorAll('a[href^="/api/audit/events/export"]');
    expect(links).to.have.length(2);
    links.forEach((link) => expect(link.hasAttribute('download')).to.equal(true));
  });

  it('sends a user who cannot read the audit log home', async () => {
    let opened = 0;
    const listener = () => { opened += 1; };
    EventBus.$on('i-open-last-project', listener);
    await mountPage({ audit_log_filters: true }, 1280, { can_read_audit_log: false });
    EventBus.$off('i-open-last-project', listener);
    expect(opened).to.equal(1);
    expect(auditRequests()).to.have.length(0);
    expect(http.requests).to.have.length(0);
  });

  it('shows the exact time with milliseconds and the offset in the card', async () => {
    respondWith({ events: [auditEvent(1, { timestamp: '2026-10-06T10:00:00.123456Z' })], older: null, newer: null });
    const wrapper = await mountPage();
    await wrapper.findAll('tbody tr').at(0).trigger('click');
    await flush();
    const card = wrapper.find('[data-testid="audit-card"]').text();
    expect(card).to.contain('2026-10-06 10:00:00.123 +00:00');
    expect(card).to.match(/2026-10-0\d \d\d:\d\d:00\.123 [+-]\d\d:\d\d/);
    expect(wrapper.findAll('tbody tr').at(0).text()).to.not.contain('.123');
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

  it('includes the chosen end minute in the period and its export', async () => {
    const wrapper = await mountPage();
    wrapper.vm.setFilters({ from: '2026-10-01T10:00', to: '2026-10-01T10:00' });
    await flush();
    const end = new Date('2026-10-01T10:01').toISOString();
    const query = new URLSearchParams(lastQuery());
    expect(query.get('from')).to.equal(new Date('2026-10-01T10:00').toISOString());
    expect(query.get('to')).to.equal(end);
    expect(new URLSearchParams(wrapper.vm.exportUrl('csv').split('?')[1]).get('to')).to.equal(end);
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

  it('names a project missing from the list by its ID', async () => {
    const wrapper = await mountPage();
    expect(wrapper.vm.projectName(auditEvent(1, { scope: { project_id: '3' } }))).to.equal('Infra');
    expect(wrapper.vm.projectName(auditEvent(1))).to.equal('Project #9');
    await wrapper.findAll('tbody tr').at(0).trigger('click');
    await flush();
    expect(wrapper.find('[data-testid="audit-card"]').text()).to.not.contain('deleted');
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

  it('names projects when only the user list fails to load', async () => {
    http.respond((config) => {
      if (config.url === '/api/users') {
        throw new Error('boom');
      }
      return config.url === '/api/projects' ? [{ id: 9, name: 'Infra' }] : PAGE;
    });
    const wrapper = await mountPage();
    expect(wrapper.vm.projectName(auditEvent(1))).to.equal('Infra');
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
