import './setup';
import { expect } from 'chai';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import i18n from '@/plugins/i18';
import EventBus from '@/event-bus';
import mockAxios, { httpError } from './helpers/axiosMock';

describe('Team.vue', () => {
  let Team;
  let http;

  before(() => {
    // eslint-disable-next-line global-require
    Team = require('@/views/project/Team.vue').default;
  });

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  const ctx = (extra) => ({
    projectId: 1,
    roles: [
      {
        slug: 'owner', name: 'Owner', builtin: true, project_id: null,
      },
    ],
    loaded: 0,
    saving: Promise.resolve(),
    pendingSaves: 0,
    async loadItems() { this.loaded += 1; },
    $i18n: { t: (key) => key },
    ...extra,
  });

  it('saves the roles of a member', async () => {
    const c = ctx();
    await Team.methods.updateProjectUser.call(c, { id: 5, roles: ['owner', 'guest'] });
    expect(http.requests).to.have.length(1);
    expect(http.requests[0].method).to.equal('put');
    expect(http.requests[0].url).to.equal('/api/project/1/users/5');
    expect(http.requests[0].data).to.deep.equal({ roles: ['owner', 'guest'] });
    expect(c.loaded).to.equal(1);
  });

  it('does not save an empty list', async () => {
    const shown = [];
    const listener = (e) => shown.push(e.text);
    EventBus.$on('i-snackbar', listener);
    const c = ctx();

    await Team.methods.updateProjectUser.call(c, { id: 5, roles: [] });

    EventBus.$off('i-snackbar', listener);
    expect(http.requests).to.have.length(0);
    expect(shown).to.deep.equal(['memberNeedsRole']);
    expect(c.loaded).to.equal(1);
  });

  it('shows the server refusal', async () => {
    const shown = [];
    const listener = (e) => shown.push(e.text);
    EventBus.$on('i-snackbar', listener);
    http.respond(() => {
      throw httpError(400, { error: 'The project needs at least one Owner' });
    });
    const c = ctx();

    await Team.methods.updateProjectUser.call(c, { id: 5, roles: ['guest'] });

    EventBus.$off('i-snackbar', listener);
    expect(shown).to.deep.equal(['The project needs at least one Owner']);
    expect(c.loaded).to.equal(1);
  });

  it('saves quick picks one after another', async () => {
    const tick = () => new Promise((resolve) => { setTimeout(resolve); });
    let release;
    http.respond(() => (http.requests.length === 1
      ? new Promise((resolve) => { release = resolve; })
      : {}));
    const c = ctx();
    const member = { id: 5, roles: ['owner'] };

    const first = Team.methods.updateProjectUser.call(c, member);
    await tick();
    member.roles = ['owner', 'guest'];
    const second = Team.methods.updateProjectUser.call(c, member);
    await tick();
    expect(http.requests).to.have.length(1);

    release({});
    await Promise.all([first, second]);
    expect(http.requests.map((r) => r.data.roles)).to.deep.equal([['owner'], ['owner', 'guest']]);
    expect(c.loaded).to.equal(1);
  });

  it('keeps a role missing from the list', () => {
    const items = Team.methods.roleItems.call(ctx(), { roles: ['owner', 'deployer'] });
    expect(items.filter((i) => i.slug).map((i) => i.slug)).to.deep.equal(['owner', 'deployer']);
    expect(items.find((i) => i.slug === 'deployer').name).to.equal('deployer');
  });

  it('keeps refusing an empty list without mapped roles', async () => {
    const shown = [];
    const listener = (e) => shown.push(e.text);
    EventBus.$on('i-snackbar', listener);
    const c = ctx();

    await Team.methods.updateProjectUser.call(c, { id: 5, roles: [], mapped_roles: [] });

    EventBus.$off('i-snackbar', listener);
    expect(http.requests).to.have.length(0);
    expect(shown).to.deep.equal(['memberNeedsRole']);
  });

  it('saves an empty list when mapped roles remain', async () => {
    const c = ctx();
    const mapped = [{ role: 'owner', provider: 'okta', group: 'Admins' }];

    await Team.methods.updateProjectUser.call(c, { id: 5, roles: [], mapped_roles: mapped });

    expect(http.requests).to.have.length(1);
    expect(http.requests[0].method).to.equal('put');
    expect(http.requests[0].data).to.deep.equal({ roles: [] });
  });

  it('tells that mapped roles remain after delete', async () => {
    const shown = [];
    const listener = (e) => shown.push(e.text);
    EventBus.$on('i-snackbar', listener);
    const c = {
      ...ctx(),
      items: [],
      getSingleItemUrl() { return `/api/project/1/users/${this.itemId}`; },
      async loadItems() { this.items = [{ id: 5, roles: [], mapped_roles: [{ role: 'owner' }] }]; },
    };

    await Team.methods.deleteItem.call(c, 5);

    EventBus.$off('i-snackbar', listener);
    expect(http.requests[0].method).to.equal('delete');
    expect(http.requests[0].url).to.equal('/api/project/1/users/5');
    expect(shown).to.deep.equal(['memberKeepsMappedRoles']);
  });

  it('shows the refusal to delete a member with only mapped roles', async () => {
    const shown = [];
    const listener = (e) => shown.push(e.text);
    EventBus.$on('i-snackbar', listener);
    http.respond(() => {
      throw httpError(400, { error: 'Access comes from group mappings. An admin can change the mappings.' });
    });
    const c = { ...ctx(), getSingleItemUrl() { return '/api/project/1/users/5'; } };

    await Team.methods.deleteItem.call(c, 5);

    EventBus.$off('i-snackbar', listener);
    expect(shown).to.deep.equal(['Access comes from group mappings. An admin can change the mappings.']);
  });

  it('shows mapped roles as locked chips', async () => {
    http.respond((config) => {
      if (config.url === '/api/project/1/roles/all') {
        return [{
          slug: 'owner', name: 'Owner', builtin: true, project_id: null,
        }];
      }
      return [{
        id: 5, name: 'Alice', username: 'alice', roles: [], mapped_roles: [{ role: 'owner', provider: 'okta', group: 'Admins' }],
      }];
    });
    const wrapper = mount(Team, {
      localVue: createLocalVue(),
      vuetify: new Vuetify(),
      i18n,
      propsData: { projectId: 1, systemInfo: { teams: {} } },
      stubs: {
        EditTeamMemberDialog: true,
        YesNoDialog: true,
        TeamMenu: true,
        'v-tooltip': { template: '<div><slot name="activator" :on="{}" :attrs="{}" /><slot /></div>' },
      },
    });
    await new Promise((resolve) => { setTimeout(resolve, 0); });

    const chip = wrapper.find('.v-chip');
    expect(chip.find('.mdi-lock').exists()).to.equal(true);
    expect(chip.text()).to.equal('Owner');
    expect(wrapper.text()).to.contain('okta: Admins');
    wrapper.destroy();
  });
});
