import { expect } from 'chai';
import groupRolesForSelect, { roleScope } from '@/lib/roles';

const labels = { projectRoles: 'Project', globalRoles: 'Global', builtInRoles: 'Built-in' };
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
    expect(items.filter((i) => i.header).map((i) => i.header)).to.deep.equal(['Project', 'Global', 'Built-in']);
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
  it('names built-in, global and project roles', () => {
    expect(roleScope({ builtin: true, project_id: null })).to.equal('builtIn');
    expect(roleScope({ builtin: false, project_id: null })).to.equal('global');
    expect(roleScope({ builtin: false, project_id: 1 })).to.equal('project');
  });
});
