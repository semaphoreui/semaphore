import './setup';
import { expect } from 'chai';
import GlobalRoleAssignments from '@/components/GlobalRoleAssignments.vue';
import mockAxios, { httpError } from './helpers/axiosMock';

const t = (key, params) => (params ? `${key}:${JSON.stringify(params)}` : key);

describe('GlobalRoleAssignments.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  describe('availableRoles', () => {
    const { availableRoles } = GlobalRoleAssignments.computed;

    it('offers only the global roles the user does not hold', () => {
      const roles = [
        { id: 2, slug: 'auditor', permissions: 8 },
        { id: 3, slug: 'ops', permissions: 4 },
        { id: 5, slug: 'empty', permissions: 0 },
      ];
      const assignments = [{ id: 10, role_id: 3 }];
      expect(availableRoles.call({ roles, assignments }).map((r) => r.slug)).to.deep.equal(['auditor', 'empty']);
    });

    it('is empty without roles', () => {
      expect(availableRoles.call({ roles: [], assignments: [] })).to.deep.equal([]);
    });
  });

  describe('provenance', () => {
    const { provenance } = GlobalRoleAssignments.methods;

    it('names the role, the administrator flag or nothing', () => {
      const ctx = { $t: t };
      expect(provenance.call(ctx, { allowed: false })).to.equal('notGranted');
      expect(provenance.call(ctx, { allowed: true, role_id: 3, role_name: 'Ops' })).to.equal('grantedByRole:{"role":"Ops"}');
      expect(provenance.call(ctx, { allowed: true, role_name: 'Administrator' })).to.equal('grantedByAdministrator');
    });
  });

  describe('permissionTitle', () => {
    const { permissionTitle } = GlobalRoleAssignments.methods;
    const catalog = [
      { id: 'global.users.manage', permission: 1, description: 'Manage users' },
      { id: 'global.future', permission: 1024, description: 'Something new' },
    ];

    it('translates known bits and falls back to the catalog description', () => {
      const ctx = { $t: t, catalog };
      expect(permissionTitle.call(ctx, 'global.users.manage')).to.equal('canManageGlobalUsers');
      expect(permissionTitle.call(ctx, 'global.future')).to.equal('Something new');
      expect(permissionTitle.call(ctx, 'global.unknown')).to.equal('global.unknown');
    });
  });

  describe('requests', () => {
    const ctx = () => ({
      userId: 5,
      roles: [],
      catalog: [],
      assignments: [],
      effective: { permissions: 0, decisions: [] },
      newRoleId: 2,
      loading: false,
      loaded: false,
      error: null,
      $t: t,
      ...GlobalRoleAssignments.methods,
    });

    it('loads roles, the catalog and the user in one go', async () => {
      const c = ctx();
      http.respond((config) => {
        switch (config.url) {
          case '/api/global-roles': return [{ id: 2, slug: 'auditor' }];
          case '/api/global-roles/permissions': return [{ id: 'global.audit.read', permission: 8 }];
          case '/api/users/5/global-roles': return [{ id: 10, role_id: 2 }];
          case '/api/users/5/global-permissions': return { permissions: 8, decisions: [] };
          default: throw new Error(`unexpected ${config.url}`);
        }
      });

      await c.load();

      expect(c.roles).to.have.length(1);
      expect(c.catalog).to.have.length(1);
      expect(c.assignments).to.deep.equal([{ id: 10, role_id: 2 }]);
      expect(c.effective.permissions).to.equal(8);
      expect(c.loaded).to.equal(true);
      expect(c.error).to.equal(null);
    });

    it('posts the chosen role and reloads the user', async () => {
      const c = ctx();
      http.respond((config) => (config.method === 'post' ? { id: 11 } : []));

      await c.assign();

      const post = http.requests.find((r) => r.method === 'post');
      expect(post.url).to.equal('/api/users/5/global-roles');
      expect(post.data).to.deep.equal({ role_id: 2 });
      expect(http.requests.map((r) => r.url)).to.include('/api/users/5/global-permissions');
      expect(c.newRoleId).to.equal(null);
    });

    it('deletes an assignment', async () => {
      const c = ctx();
      http.respond(() => ({}));

      await c.remove({ id: 10 });

      const del = http.requests.find((r) => r.method === 'delete');
      expect(del.url).to.equal('/api/users/5/global-roles/10');
    });

    it('shows the server message when the last administrator is protected', async () => {
      const c = ctx();
      http.respond(() => { throw httpError(409, { error: 'the instance must keep at least one administrator' }); });

      await c.remove({ id: 10 });

      expect(c.error).to.contain('at least one administrator');
      expect(c.loading).to.equal(false);
    });
  });
});
