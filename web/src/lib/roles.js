const SCOPE_GROUPS = [
  { scope: 'project', label: 'projectRoles' },
  { scope: 'global', label: 'globalRoles' },
  { scope: 'builtIn', label: 'builtInRoles' },
];

function compareRoles(left, right) {
  return left.name.localeCompare(right.name) || left.slug.localeCompare(right.slug);
}

export function roleScope(role) {
  if (role.builtin) {
    return 'builtIn';
  }
  return role.project_id == null ? 'global' : 'project';
}

export default function groupRolesForSelect(roles, t) {
  const availableRoles = roles || [];
  return SCOPE_GROUPS.reduce((items, group) => {
    const groupRoles = availableRoles.filter((role) => roleScope(role) === group.scope);
    if (groupRoles.length === 0) {
      return items;
    }
    if (items.length > 0) {
      items.push({ divider: true });
    }
    items.push({ header: t(group.label) });
    items.push(...groupRoles.sort(compareRoles));
    return items;
  }, []);
}
