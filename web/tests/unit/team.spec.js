import './setup';
import { expect } from 'chai';
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

  it('keeps a role missing from the list', () => {
    const items = Team.methods.roleItems.call(ctx(), { roles: ['owner', 'deployer'] });
    expect(items.filter((i) => i.slug).map((i) => i.slug)).to.deep.equal(['owner', 'deployer']);
    expect(items.find((i) => i.slug === 'deployer').name).to.equal('deployer');
  });
});
