import './setup';
import { expect } from 'chai';
import TeamMemberForm from '@/components/TeamMemberForm.vue';
import { USER_ROLES } from '@/lib/constants';

describe('TeamMemberForm', () => {
  it('offers the built-in roles when invites are on', () => {
    const roles = TeamMemberForm.computed.availableRoles.call({
      invitesEnabled: true,
      roles: USER_ROLES,
    });
    expect(roles.map((r) => r.slug)).to.deep.equal(['owner', 'manager', 'task_runner', 'guest']);
  });

  it('offers all roles otherwise', () => {
    const roles = TeamMemberForm.computed.availableRoles.call({
      invitesEnabled: false,
      roles: [{ slug: 'owner', builtin: true }, { slug: 'deployer', builtin: false }],
    });
    expect(roles.map((r) => r.slug)).to.deep.equal(['owner', 'deployer']);
  });
});
