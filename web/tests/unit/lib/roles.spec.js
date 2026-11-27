import { expect } from 'chai';
import groupRolesForSelect, { roleScope } from '@/lib/roles';

const labels = { projectRoles: 'Project', sharedRoles: 'Shared', builtInRoles: 'Built-in' };
const t = (key) => labels[key];

describe('groupRolesForSelect', () => {
  it('groups by builtin flag and project', () => {
    const items = groupRolesForSelect([
      {
        slug: 'owner', name: 'Owner', builtin: true, project_id: null,
      },
      {
        slug: 'deployer', name: 'Deployer', builtin: false, project_id: null,
      },
      {
        slug: 'local', name: 'Local', builtin: false, project_id: 1,
      },
    ], t);
    expect(items.filter((i) => i.type === 'subheader').map((i) => i.name))
      .to.deep.equal(['Project', 'Shared', 'Built-in']);
    expect(items.filter((i) => i.type === 'divider')).to.have.length(2);
    expect(items.filter((i) => i.slug).map((i) => i.slug)).to.deep.equal(['local', 'deployer', 'owner']);
  });

  it('orders same names by slug', () => {
    const items = groupRolesForSelect([
      {
        slug: 'b', name: 'Same', builtin: false, project_id: 1,
      },
      {
        slug: 'a', name: 'Same', builtin: false, project_id: 1,
      },
    ], t);
    expect(items.filter((i) => i.slug).map((i) => i.slug)).to.deep.equal(['a', 'b']);
  });
});

describe('roleScope', () => {
  it('names built-in, shared and project roles', () => {
    expect(roleScope({ builtin: true, project_id: null })).to.equal('builtIn');
    expect(roleScope({ builtin: false, project_id: null })).to.equal('shared');
    expect(roleScope({ builtin: false, project_id: 1 })).to.equal('project');
  });
});
