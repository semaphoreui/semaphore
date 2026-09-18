import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import PermissionsCheck from '@/components/PermissionsCheck';
import { USER_PERMISSIONS } from '@/lib/constants';

const Host = {
  mixins: [PermissionsCheck],
  props: { item: Object },
  render: (h) => h('div'),
};

function can(propsData, permission) {
  return shallowMount(Host, { propsData }).vm.can(permission);
}

describe('PermissionsCheck mixin', () => {
  const {
    runProjectTasks, updateProject, manageProjectResources, manageProjectUsers,
  } = USER_PERMISSIONS;

  it('grants everything to admins regardless of permissions', () => {
    expect(can({ isAdmin: true, userPermissions: 0 }, manageProjectUsers)).to.equal(true);
  });

  const tests = [
    {
      name: 'no permissions', userPermissions: 0, permission: runProjectTasks, expected: false,
    },
    {
      name: 'exact bit', userPermissions: runProjectTasks, permission: runProjectTasks, expected: true,
    },
    {
      name: 'other bit only', userPermissions: updateProject, permission: runProjectTasks, expected: false,
    },
    {
      name: 'combined mask contains bit', userPermissions: runProjectTasks | manageProjectResources, permission: manageProjectResources, expected: true,
    },
    {
      name: 'combined mask lacks bit', userPermissions: runProjectTasks | manageProjectResources, permission: manageProjectUsers, expected: false,
    },
    {
      name: 'combined permission fully present', userPermissions: 15, permission: runProjectTasks | updateProject, expected: true,
    },
    {
      name: 'combined permission partially present', userPermissions: runProjectTasks, permission: runProjectTasks | updateProject, expected: false,
    },
    {
      name: 'undefined permissions', userPermissions: undefined, permission: runProjectTasks, expected: false,
    },
  ];

  tests.forEach(({
    name, userPermissions, permission, expected,
  }) => {
    it(`userPermissions: ${name}`, () => {
      expect(can({ userPermissions }, permission)).to.equal(expected);
    });
  });

  it('prefers item.permissions over userPermissions when the item has them', () => {
    const props = { userPermissions: 15, item: { permissions: runProjectTasks } };
    expect(can(props, runProjectTasks)).to.equal(true);
    expect(can(props, manageProjectUsers)).to.equal(false);
  });

  it('falls back to userPermissions when the item has no permissions field', () => {
    const props = { userPermissions: manageProjectUsers, item: { id: 1 } };
    expect(can(props, manageProjectUsers)).to.equal(true);
  });
});
