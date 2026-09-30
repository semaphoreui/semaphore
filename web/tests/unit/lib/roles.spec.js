import { expect } from 'chai';
import groupRolesForSelect from '@/lib/roles';

describe('groupRolesForSelect', () => {
  it('groups roles by scope and sorts each group by name and ID', () => {
    const projectID = 7;
    const projectRole2 = {
      id: 5,
      name: 'Deploy',
      project_id: projectID,
      builtin_key: null,
    };
    const projectRole1 = {
      id: 3,
      name: 'Deploy',
      project_id: projectID,
      builtin_key: null,
    };
    const globalRole = {
      id: 4,
      name: 'Auditor',
      project_id: null,
      builtin_key: null,
    };
    const builtinRole = {
      id: 2,
      name: 'Manager',
      project_id: null,
      builtin_key: 'manager',
    };

    const items = groupRolesForSelect(
      [builtinRole, projectRole2, globalRole, projectRole1],
      {
        project: 'Project roles',
        global: 'Global roles',
        builtIn: 'Built-in roles',
      },
    );

    expect(items).to.deep.equal([
      { header: 'Project roles' },
      projectRole1,
      projectRole2,
      { divider: true },
      { header: 'Global roles' },
      globalRole,
      { divider: true },
      { header: 'Built-in roles' },
      builtinRole,
    ]);
  });
});
