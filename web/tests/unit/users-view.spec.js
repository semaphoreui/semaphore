import './setup';
import { expect } from 'chai';
import EventBus from '@/event-bus';
import Users from '@/views/Users.vue';

describe('Users.vue', () => {
  // ItemListPageBase calls allowActions() from data(), before any computed property exists.
  it('decides the actions column from props alone', () => {
    const { allowActions } = Users.methods;
    expect(allowActions.call({ user: { admin: true } })).to.equal(true);
    expect(allowActions.call({ user: { admin: false, global_permissions: 1 } })).to.equal(true);
    expect(allowActions.call({ user: { admin: false, global_permissions: 2 } })).to.equal(false);
    expect(allowActions.call({ user: null })).to.equal(false);
  });

  it('shows the Global roles tab only in a Pro build', () => {
    const can = (isPro, user) => Users.computed.canManageGlobalRoles.call({ isPro, user });
    expect(can(true, { admin: true })).to.equal(true);
    expect(can(true, { admin: false, global_permissions: 2 })).to.equal(true);
    expect(can(true, { admin: false, global_permissions: 1 })).to.equal(false);
    expect(can(false, { admin: true })).to.equal(false);
  });

  it('offers role assignment only with the permission and the Pro feature', () => {
    const can = (canManageGlobalRoles, features) => Users.computed.canAssignGlobalRoles.call({
      canManageGlobalRoles, features,
    });
    expect(can(true, { global_roles_management: true })).to.equal(true);
    expect(can(true, {})).to.equal(false);
    expect(can(false, { global_roles_management: true })).to.equal(false);
    expect(can(true, undefined)).to.equal(false);
  });

  describe('tab', () => {
    const { get, set } = Users.computed.tab;

    it('is read from the address', () => {
      expect(get.call({ canManageGlobalRoles: true, $route: { path: '/users/roles' } })).to.equal('global-roles');
      expect(get.call({ canManageGlobalRoles: true, $route: { path: '/users' } })).to.equal('users');
      expect(get.call({ canManageGlobalRoles: false, $route: { path: '/users/roles' } })).to.equal('users');
    });

    it('changes the address when switched', () => {
      const pushed = [];
      const c = { $route: { path: '/users' }, $router: { push: (p) => pushed.push(p) } };
      set.call(c, 'global-roles');
      set.call(c, 'users');
      c.$route.path = '/users/roles';
      set.call(c, 'global-roles');
      expect(pushed).to.deep.equal(['/users/roles']);
    });
  });

  describe('onUserSaved', () => {
    const ctx = () => {
      const c = {
        itemId: 'new',
        loaded: 0,
        $i18n: { t: (key) => key },
        async loadItems() { this.loaded += 1; },
      };
      return c;
    };

    it('keeps a created user open for editing and says so', async () => {
      const c = ctx();
      const shown = [];
      const onSnackbar = (e) => shown.push(e);
      EventBus.$on('i-snackbar', onSnackbar);
      try {
        await Users.methods.onUserSaved.call(c, { action: 'new', item: { id: 12 } });
      } finally {
        EventBus.$off('i-snackbar', onSnackbar);
      }
      expect(c.itemId).to.equal(12);
      expect(c.loaded).to.equal(1);
      expect(shown).to.deep.equal([{ color: 'success', text: 'userCreated' }]);
    });

    it('leaves an edited user alone', async () => {
      const c = ctx();
      c.itemId = 7;
      await Users.methods.onUserSaved.call(c, { action: 'edit', item: { id: 7 } });
      expect(c.itemId).to.equal(7);
      expect(c.loaded).to.equal(0);
    });
  });
});
