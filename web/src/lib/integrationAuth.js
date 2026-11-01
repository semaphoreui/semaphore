// Webhook authorization methods of endpoints and integration groups, as the API names them.

/** Methods whose secret is read from a request header named by `auth_header`. */
export const HEADER_AUTH_METHODS = ['token', 'hmac', 'hmac-sha512'];

/** Select items; the empty id is "None" and takes its title from `t`, a translate function. */
export function authMethods(t) {
  return [
    { id: '', title: t('none') },
    { id: 'github', title: 'GitHub Webhooks' },
    { id: 'bitbucket', title: 'Bitbucket Webhooks' },
    { id: 'token', title: 'Token' },
    { id: 'hmac', title: 'HMAC' },
    { id: 'hmac-sha512', title: 'HMAC (SHA-512)' },
    { id: 'basic', title: 'BasicAuth' },
  ];
}

export function authMethodTitle(method, t) {
  const found = authMethods(t).find((m) => m.id === (method || ''));
  return found ? found.title : method;
}

// Token and HMAC-style methods verify one secret string, so a String key or
// a login/password key fits; BasicAuth also needs a login, so only the latter.
export function allowedSecretKeyTypes(method) {
  return method === 'basic' ? ['login_password'] : ['login_password', 'string'];
}
