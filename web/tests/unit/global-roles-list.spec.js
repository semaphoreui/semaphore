import './setup';
import { expect } from 'chai';
import EventBus from '@/event-bus';
import mockAxios, { httpError } from './helpers/axiosMock';

describe('GlobalRolesList.vue', () => {
  let GlobalRolesList;
  let http;

  before(() => {
    // eslint-disable-next-line global-require
    GlobalRolesList = require('@/components/GlobalRolesList.vue').default;
  });

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  const ctx = () => ({
    itemId: null,
    items: [{ id: 4, slug: 'auditor' }],
    roleImpact: null,
    forceDeleteDialog: false,
    $t: (key) => key, // vue-i18n 11 (composition mode): components translate with $t
    async loadItems() { return null; },
    getEventName() { return 'i-global-role'; },
    getSingleItemUrl() { return `/api/global-roles/${this.itemId}`; },
  });

  describe('loadItems', () => {
    it('does not ask the API without the Pro feature and shows an empty list', async () => {
      const c = { features: { global_roles_management: false }, items: null, getItemsUrl: () => '/api/global-roles' };
      await GlobalRolesList.methods.loadItems.call(c);
      expect(c.items).to.deep.equal([]);
      expect(http.requests).to.have.length(0);
    });

    it('loads the roles with the feature', async () => {
      http.respond(() => [{ id: 4, slug: 'auditor' }]);
      const c = { features: { global_roles_management: true }, items: null, getItemsUrl: () => '/api/global-roles' };
      await GlobalRolesList.methods.loadItems.call(c);
      expect(c.items).to.deep.equal([{ id: 4, slug: 'auditor' }]);
      expect(http.requests[0].url).to.equal('/api/global-roles');
    });

    it('treats a 403 from the API as an empty list', async () => {
      http.respond(() => { throw httpError(403, { error: 'Pro only' }); });
      const c = { features: { global_roles_management: true }, items: null, getItemsUrl: () => '/api/global-roles' };
      await GlobalRolesList.methods.loadItems.call(c);
      expect(c.items).to.deep.equal([]);
    });
  });

  it('talks to the global roles API by id', () => {
    const { getItemsUrl, getSingleItemUrl } = GlobalRolesList.methods;
    expect(getItemsUrl.call({})).to.equal('/api/global-roles');
    expect(getSingleItemUrl.call({ itemId: 4 })).to.equal('/api/global-roles/4');
  });

  it('decides the actions column from the global permission alone', () => {
    const { allowActions } = GlobalRolesList.methods;
    expect(allowActions.call({ user: { admin: true } })).to.equal(true);
    expect(allowActions.call({ user: { admin: false, global_permissions: 2 } })).to.equal(true);
    expect(allowActions.call({ user: { admin: false, global_permissions: 1 } })).to.equal(false);
  });

  it('asks again with the number of holders when the role is assigned', async () => {
    const impact = { members: 2, projects: 0, template_grants: 0 };
    http.respond(() => { throw httpError(409, { error: 'assigned', impact }); });
    const c = ctx();

    await GlobalRolesList.methods.deleteItem.call(c, 4);

    expect(c.forceDeleteDialog).to.equal(true);
    expect(c.roleImpact).to.deep.equal(impact);
  });

  it('shows the error of a conflict without a body', async () => {
    const shown = [];
    const listener = (e) => shown.push(e.text);
    EventBus.$on('i-snackbar', listener);
    http.respond(() => { throw httpError(409, undefined, 'Conflict'); });
    const c = ctx();

    await GlobalRolesList.methods.deleteItem.call(c, 4);

    EventBus.$off('i-snackbar', listener);
    expect(c.forceDeleteDialog).to.equal(false);
    expect(shown).to.deep.equal(['Request failed with status code 409']);
  });

  it('deletes with force after the second question', async () => {
    const c = ctx();

    await GlobalRolesList.methods.deleteItem.call(c, 4, true);

    expect(http.requests[0].method).to.equal('delete');
    expect(http.requests[0].url).to.equal('/api/global-roles/4?force=true');
  });
});
