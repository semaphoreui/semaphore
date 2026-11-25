// The servers of an mcp inventory as the form edits them, and the checks the server makes on
// them (db.McpServer.Validate), so that a mistake is shown at the field before saving.

// A name goes into the agent's config and environment: letters, digits, '-' and '_' only.
export const MCP_SERVER_NAME_PATTERN = /^[A-Za-z0-9_-]{1,64}$/;

export function isValidMcpServerName(name) {
  return MCP_SERVER_NAME_PATTERN.test(name || '');
}

// An http or https URL with a host and without credentials: a user name or password in the URL
// would be stored and shown in plain text, a token belongs in the token field.
export function isValidMcpServerUrl(value) {
  let url;
  try {
    url = new URL(value);
  } catch (e) {
    return false;
  }
  return ['http:', 'https:'].includes(url.protocol)
    && url.hostname !== ''
    && url.username === ''
    && url.password === '';
}

// A row of the form; `key` tells the rows apart for Vue, the caller numbers them. `token` is a
// new value typed by the user, never a stored one: the API does not return tokens, only whether
// a server has one (`hasToken`).
export function newMcpServerRow(key, server = {}) {
  return {
    key,
    id: server.id ?? null,
    name: server.name ?? '',
    url: server.url ?? '',
    enabled: server.enabled ?? true,
    hasToken: !!server.has_token,
    token: '',
    removeToken: false,
  };
}

// The rows of the servers of a loaded inventory, keyed 1…n.
export function mcpServerRows(servers) {
  return (servers || []).map((server, index) => newMcpServerRow(index + 1, server));
}

// The token change of a row, or null to leave the token as it is. A pasted token often comes
// with a line break: the ends are trimmed, a bearer token has no spaces.
export function mcpTokenChange(row) {
  const secret = (row.token || '').trim();
  if (secret !== '') {
    return { operation: row.hasToken ? 'update' : 'create', secret };
  }
  if (row.hasToken && row.removeToken) {
    return { operation: 'delete' };
  }
  return null;
}

// The `mcp_servers` of a POST or PUT: the whole list, a server left out is deleted with its
// token; a server with an id is updated, one without is created.
export function mcpServersRequest(rows) {
  return rows.map((row) => {
    const server = {
      name: row.name,
      url: row.url,
      enabled: row.enabled,
    };
    if (row.id != null) {
      server.id = row.id;
    }
    const token = mcpTokenChange(row);
    if (token != null) {
      server.token = token;
    }
    return server;
  });
}
