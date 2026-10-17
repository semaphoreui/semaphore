import './setup';
import { expect } from 'chai';
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

  it('offers the Global roles tab only with the permission and the Pro feature', () => {
    const can = (canManageGlobalRoles, features) => Users.computed.canAssignGlobalRoles.call({
      canManageGlobalRoles, features,
    });
    expect(can(true, { custom_roles_management: true })).to.equal(true);
    expect(can(true, {})).to.equal(false);
    expect(can(false, { custom_roles_management: true })).to.equal(false);
    expect(can(true, undefined)).to.equal(false);
  });
});
