function compareRoles(left, right) {
  return left.name.localeCompare(right.name) || left.id - right.id;
}

export default function groupRolesForSelect(roles, labels) {
  const availableRoles = roles || [];
  const groups = [
    {
      label: labels.project,
      roles: availableRoles.filter((role) => role.builtin_key == null && role.project_id != null),
    },
    {
      label: labels.global,
      roles: availableRoles.filter((role) => role.builtin_key == null && role.project_id == null),
    },
    {
      label: labels.builtIn,
      roles: availableRoles.filter((role) => role.builtin_key != null),
    },
  ].filter((group) => group.roles.length > 0);

  return groups.reduce((items, group, index) => {
    if (index > 0) {
      items.push({ divider: true });
    }
    items.push({ header: group.label });
    items.push(...group.roles.sort(compareRoles));
    return items;
  }, []);
}
