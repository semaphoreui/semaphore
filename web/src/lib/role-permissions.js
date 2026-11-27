/**
 * Says whether the signed-in user holds a global permission. `user` is the `/api/user` response:
 * an admin holds everything, anyone else only what `global_permissions` carries.
 */
export function hasGlobalPermission(user, permission) {
  if (!user) {
    return false;
  }
  if (user.admin) {
    return true;
  }
  // eslint-disable-next-line no-bitwise
  return ((user.global_permissions || 0) & permission) === permission;
}

/** Splits a permission mask into the catalog entries it contains. */
export function permissionEntries(catalog, mask) {
  // eslint-disable-next-line no-bitwise
  return (catalog || []).filter((entry) => ((mask || 0) & entry.permission) === entry.permission);
}
