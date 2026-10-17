import './setup';
import { expect } from 'chai';
import EventBus from '@/event-bus';
import mockAxios, { httpError } from './helpers/axiosMock';

describe('Roles.vue', () => {
  let Roles;
  let http;

  before(() => {
    // eslint-disable-next-line global-require
    Roles = require('@/views/Roles.vue').default;
  });

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  const ctx = () => ({
    projectId: null,
    tab: 'project',
    isGlobalTab: false,
    itemId: null,
    items: [{ slug: 'deployer' }],
    IDFieldName: 'slug',
    roleImpact: null,
    forceDeleteDialog: false,
    can() { return true; },
    $i18n: { t: (key) => key },
    async loadItems() { return null; },
    getEventName() { return 'i-role'; },
    getSingleItemUrl() { return `/api/roles/${this.itemId}`; },
  });

  // ItemListPageBase calls allowActions() from data(), before any computed property exists.
  it('decides the actions column from props alone', () => {
    const { allowActions } = Roles.methods;
    const can = () => true;
    const global = (user) => allowActions.call({ projectId: null, user, can });
    expect(global({ admin: true })).to.equal(true);
    expect(global({ admin: false, global_permissions: 2 })).to.equal(true);
    expect(global({ admin: false, global_permissions: 1 })).to.equal(false);
    const project = (canProject) => allowActions.call({
      projectId: 7, user: { admin: false }, can: () => canProject,
    });
    expect(project(false)).to.equal(false);
    expect(project(true)).to.equal(true);
  });

  it('asks again with the impact when the role is assigned', async () => {
    const impact = { members: 2, projects: 1, template_grants: 3 };
    http.respond(() => { throw httpError(409, { error: 'assigned', impact }); });
    const c = ctx();

    await Roles.methods.deleteItem.call(c, 'deployer');

    expect(c.forceDeleteDialog).to.equal(true);
    expect(c.roleImpact).to.deep.equal(impact);
  });

  it('shows the error of a conflict without a body', async () => {
    const shown = [];
    const listener = (e) => shown.push(e.text);
    EventBus.$on('i-snackbar', listener);
    http.respond(() => { throw httpError(409, undefined, 'Conflict'); });
    const c = ctx();

    await Roles.methods.deleteItem.call(c, 'deployer');

    EventBus.$off('i-snackbar', listener);
    expect(c.forceDeleteDialog).to.equal(false);
    expect(shown).to.deep.equal(['Request failed with status code 409']);
  });

  it('tells a member who cannot manage users why the role stays', async () => {
    const shown = [];
    const listener = (e) => shown.push(e.text);
    EventBus.$on('i-snackbar', listener);
    http.respond(() => { throw httpError(409, { error: 'assigned', impact: { members: 1, projects: 1, template_grants: 0 } }); });
    const c = { ...ctx(), projectId: 1, can: () => false };

    await Roles.methods.deleteItem.call(c, 'deployer');

    EventBus.$off('i-snackbar', listener);
    expect(c.forceDeleteDialog).to.equal(false);
    expect(shown).to.deep.equal(['roleInUseNeedsManageUsers']);
  });

  it('addresses project roles by slug and global roles by id', () => {
    const { getItemsUrl, getSingleItemUrl } = Roles.methods;
    const { IDFieldName } = Roles.computed;
    const project = { projectId: null, isGlobalTab: false, itemId: 'deployer' };
    expect(IDFieldName.call(project)).to.equal('slug');
    expect(getItemsUrl.call(project)).to.equal('/api/roles');
    expect(getSingleItemUrl.call(project)).to.equal('/api/roles/deployer');
    const global = { projectId: null, isGlobalTab: true, itemId: 4 };
    expect(IDFieldName.call(global)).to.equal('id');
    expect(getItemsUrl.call(global)).to.equal('/api/global-roles');
    expect(getSingleItemUrl.call(global)).to.equal('/api/global-roles/4');
    const inProject = { projectId: 7, isGlobalTab: false, itemId: 'deployer' };
    expect(getItemsUrl.call(inProject)).to.equal('/api/project/7/roles');
    expect(getSingleItemUrl.call(inProject)).to.equal('/api/project/7/roles/deployer');
  });

  it('has the global tab only on the instance page', () => {
    const { isGlobalTab } = Roles.computed;
    expect(isGlobalTab.call({ projectId: null, tab: 'global' })).to.equal(true);
    expect(isGlobalTab.call({ projectId: null, tab: 'project' })).to.equal(false);
    expect(isGlobalTab.call({ projectId: 7, tab: 'global' })).to.equal(false);
  });

  it('deletes with force after the second question', async () => {
    const c = ctx();

    await Roles.methods.deleteItem.call(c, 'deployer', true);

    expect(http.requests[0].method).to.equal('delete');
    expect(http.requests[0].url).to.equal('/api/roles/deployer?force=true');
  });
});
