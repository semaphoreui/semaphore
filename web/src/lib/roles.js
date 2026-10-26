const SCOPE_GROUPS = [
  { scope: 'project', label: 'projectRoles' },
  { scope: 'shared', label: 'sharedRoles' },
  { scope: 'builtIn', label: 'builtInRoles' },
];

function compareRoles(left, right) {
  return left.name.localeCompare(right.name) || left.slug.localeCompare(right.slug);
}

/**
 * Where a project role is defined: in one project, shared by every project, or built in.
 * Global roles (instance-wide permissions) are a different entity and never come here.
 */
export function roleScope(role) {
  if (role.builtin) {
    return 'builtIn';
  }
  return role.project_id == null ? 'shared' : 'project';
}

const SLUG_RE = /^[A-Za-z0-9_-]{1,100}$/;

export function isValidSlug(slug) {
  return SLUG_RE.test(slug);
}

/** Derives a slug from a name: lower case, one hyphen between words, at most 100 characters. */
export function slugFromName(name) {
  return (name || '')
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 100);
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
