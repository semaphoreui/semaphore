import { expect } from 'chai';
import { hasGlobalPermission, permissionEntries } from '@/lib/role-permissions';
import { GLOBAL_PERMISSIONS } from '@/lib/constants';

describe('lib/role-permissions', () => {
  const { manageUsers, manageRoles, readAudit } = GLOBAL_PERMISSIONS;

  describe('hasGlobalPermission', () => {
    it('grants everything to an admin', () => {
      const admin = { admin: true, global_permissions: 0 };
      expect(hasGlobalPermission(admin, manageRoles)).to.equal(true);
    });

    it('reads the mask of everyone else', () => {
      const user = { admin: false, global_permissions: manageUsers | readAudit };
      expect(hasGlobalPermission(user, manageUsers)).to.equal(true);
      expect(hasGlobalPermission(user, readAudit)).to.equal(true);
      expect(hasGlobalPermission(user, manageRoles)).to.equal(false);
      expect(hasGlobalPermission(user, manageUsers | manageRoles)).to.equal(false);
    });

    it('denies without a user or a mask', () => {
      expect(hasGlobalPermission(null, manageUsers)).to.equal(false);
      expect(hasGlobalPermission({}, manageUsers)).to.equal(false);
    });
  });

  it('lists the catalog entries of a mask', () => {
    const catalog = [
      { id: 'a', permission: 1 }, { id: 'b', permission: 2 }, { id: 'c', permission: 4 },
    ];
    expect(permissionEntries(catalog, 5).map((e) => e.id)).to.deep.equal(['a', 'c']);
    expect(permissionEntries(catalog, 0)).to.deep.equal([]);
    expect(permissionEntries(undefined, 7)).to.deep.equal([]);
  });
});
