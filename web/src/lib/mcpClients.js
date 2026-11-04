// How each MCP client connects to the Semaphore MCP server: Streamable HTTP
// with a static "Authorization: Bearer <token>" header. Field names differ
// between clients, and a wrong one is often ignored or switches the client to
// the legacy SSE transport, which the server does not speak. Checked against
// the clients' documentation in October 2026.
//
// The connection may also turn tool groups on with extra headers (see
// mcpConnectionHeaders); every config and command below carries them next to
// Authorization, in the form that client takes. The project and the admin sets
// (MCP_SETS) are two servers for a client and differ only in the URL, the
// server name and the headers; one MCP token signs in to both.

export const TOKEN_PLACEHOLDER = '<MCP_TOKEN>';

const json = (value) => JSON.stringify(value, null, 2);

// A ${name} reference for the client to expand, built so it is not read as a JS template.
const ref = (name) => ['$', '{', name, '}'].join('');

const bearer = `Bearer ${TOKEN_PLACEHOLDER}`;

// The headers that turn a tool group on for one connection. The server reads a
// missing header as off, so a group that is off sends no header at all.
export const MCP_RUN_HEADER = 'X-Semaphore-MCP-Run';
export const MCP_WRITE_HEADER = 'X-Semaphore-MCP-Write';
export const MCP_DESTROY_HEADER = 'X-Semaphore-MCP-Destroy';

// The two sets of MCP tools, each its own MCP server for a client: the path of its endpoint, the
// server name in the client configs, and its tool groups in the order of the page — the header
// that turns a group on for a connection and the server setting that allows it at all.
export const MCP_SETS = {
  project: {
    id: 'project',
    path: 'api/mcp/project',
    serverName: 'semaphore',
    groups: [
      { key: 'run', header: MCP_RUN_HEADER, setting: 'allow_run' },
      { key: 'destroy', header: MCP_DESTROY_HEADER, setting: 'allow_destroy' },
    ],
  },
  admin: {
    id: 'admin',
    path: 'api/mcp/admin',
    serverName: 'semaphore-admin',
    groups: [
      { key: 'write', header: MCP_WRITE_HEADER, setting: 'admin_allow_write' },
      { key: 'destroy', header: MCP_DESTROY_HEADER, setting: 'admin_allow_destroy' },
    ],
  },
};

/**
 * The extra headers of a connection to `set`: `on` — { group key: bool } → { header: 'on' }, in
 * the order of the set's groups.
 */
export function mcpConnectionHeaders(set, on = {}) {
  const headers = {};
  set.groups.forEach((group) => {
    if (on[group.key]) {
      headers[group.header] = 'on';
    }
  });
  return headers;
}

// Each client: id, name, and snippets — { title, lang, code }; note is optional.
// options.headers — the extra headers of the connection (mcpConnectionHeaders);
// options.serverName — the name the client knows the server by, one per set.
export function mcpClients(url, {
  headers: extra = {},
  serverName: name = MCP_SETS.project.serverName,
} = {}) {
  const extraEntries = Object.entries(extra);
  // The headers of a JSON config: Authorization first, then the extra ones.
  const withExtra = (authorization) => ({ Authorization: authorization, ...extra });
  // The --header flags of a command line after the Authorization one.
  const headerFlags = extraEntries.map(([header, value]) => ` --header "${header}: ${value}"`).join('');
  const urlAndHeaders = () => ({
    mcpServers: {
      [name]: {
        url,
        headers: withExtra(bearer),
      },
    },
  });
  const codexHeaders = extraEntries.length === 0 ? ''
    : `\nhttp_headers = { ${extraEntries.map(([header, value]) => `"${header}" = "${value}"`).join(', ')} }`;
  const otherHeaders = [`"Authorization: ${bearer}"`]
    .concat(extraEntries.map(([header, value]) => `"${header}: ${value}"`)).join(', ');

  return [
    {
      id: 'claude-code',
      name: 'Claude Code',
      snippets: [
        {
          title: 'Command',
          lang: 'bash',
          code: `claude mcp add --transport http --scope user ${name} ${url} --header "Authorization: ${bearer}"${headerFlags}`,
        },
        {
          title: '.mcp.json in the project root',
          lang: 'json',
          code: json({
            mcpServers: {
              [name]: {
                type: 'http',
                url,
                headers: withExtra(`Bearer ${ref('SEMAPHORE_MCP_TOKEN')}`),
              },
            },
          }),
        },
      ],
      note: `"type": "http" is required in .mcp.json. ${ref('SEMAPHORE_MCP_TOKEN')} is read from the environment.`,
    },
    {
      id: 'cursor',
      name: 'Cursor',
      snippets: [
        {
          title: '~/.cursor/mcp.json',
          lang: 'json',
          code: json(urlAndHeaders()),
        },
      ],
    },
    {
      id: 'vscode',
      name: 'VS Code (Copilot)',
      snippets: [
        {
          title: '.mcp.json in the project root or ~/.copilot/mcp-config.json',
          lang: 'json',
          code: json({
            mcpServers: {
              [name]: {
                type: 'http',
                url,
                headers: withExtra(bearer),
              },
            },
          }),
        },
        {
          title: '.vscode/mcp.json (asks for the token on start)',
          lang: 'json',
          code: json({
            inputs: [
              {
                type: 'promptString',
                id: `${name}-token`,
                description: `${name} MCP token`,
                password: true,
              },
            ],
            servers: {
              [name]: {
                type: 'http',
                url,
                headers: withExtra(`Bearer ${ref(`input:${name}-token`)}`),
              },
            },
          }),
        },
      ],
      note: `Servers that use ${ref('input:...')} are not passed to Copilot agent sessions; use the first file there.`,
    },
    {
      id: 'codex',
      name: 'Codex (CLI, IDE, ChatGPT desktop)',
      snippets: [
        {
          title: 'Command',
          lang: 'bash',
          code: `export SEMAPHORE_MCP_TOKEN=${TOKEN_PLACEHOLDER}\ncodex mcp add ${name} --url ${url} --bearer-token-env-var SEMAPHORE_MCP_TOKEN`,
        },
        {
          title: '~/.codex/config.toml',
          lang: 'toml',
          code: `[mcp_servers.${name}]\nurl = "${url}"\nbearer_token_env_var = "SEMAPHORE_MCP_TOKEN"${codexHeaders}`,
        },
      ],
      note: extraEntries.length === 0
        ? 'The Codex CLI, the IDE extension and ChatGPT desktop share ~/.codex/config.toml. Codex takes the token from bearer_token_env_var, not from a header.'
        : 'The Codex CLI, the IDE extension and ChatGPT desktop share ~/.codex/config.toml. Codex takes the token from bearer_token_env_var; codex mcp add has no option for other headers, so add the http_headers line to config.toml after the command.',
    },
    {
      id: 'gemini',
      name: 'Gemini CLI',
      snippets: [
        {
          title: 'Command',
          lang: 'bash',
          code: `gemini mcp add --transport http --scope user --header "Authorization: ${bearer}"${headerFlags} ${name} ${url}`,
        },
        {
          title: '~/.gemini/settings.json',
          lang: 'json',
          code: json({
            mcpServers: {
              [name]: {
                httpUrl: url,
                headers: withExtra(bearer),
              },
            },
          }),
        },
      ],
      note: 'Use "httpUrl": Gemini CLI reads a plain "url" as the legacy SSE transport.',
    },
    {
      id: 'devin',
      name: 'Windsurf / Devin Desktop',
      snippets: [
        {
          title: '~/.config/devin/mcp_config.json',
          lang: 'json',
          code: json(urlAndHeaders()),
        },
      ],
      note: 'On Windows the file is %APPDATA%\\devin\\mcp_config.json. If an older Cascade build shows a red status, rename "url" to "serverUrl".',
    },
    {
      id: 'jetbrains',
      name: 'JetBrains (AI Assistant, Junie)',
      snippets: [
        {
          title: 'AI Assistant: Settings → Tools → AI Assistant → Model Context Protocol → Add; Junie: ~/.junie/mcp/mcp.json',
          lang: 'json',
          code: json(urlAndHeaders()),
        },
      ],
      note: 'AI Assistant connects to a remote URL from version 2025.3.',
    },
    {
      id: 'zed',
      name: 'Zed',
      snippets: [
        {
          title: 'settings.json',
          lang: 'json',
          code: json({
            context_servers: {
              [name]: {
                url,
                headers: withExtra(bearer),
              },
            },
          }),
        },
      ],
      note: 'Do not add "type". Environment variables are not expanded in headers.',
    },
    {
      id: 'cline',
      name: 'Cline',
      snippets: [
        {
          title: 'Command',
          lang: 'bash',
          code: `cline mcp add ${name} ${url} --transport http --header "Authorization: ${bearer}"${headerFlags} --yes`,
        },
        {
          title: 'cline_mcp_settings.json (Cline → MCP Servers → Configure)',
          lang: 'json',
          code: json({
            mcpServers: {
              [name]: {
                type: 'streamableHttp',
                url,
                headers: withExtra(bearer),
                disabled: false,
              },
            },
          }),
        },
      ],
      note: 'Keep "type": "streamableHttp": without it Cline uses the legacy SSE transport.',
    },
    {
      id: 'continue',
      name: 'Continue',
      snippets: [
        {
          title: `.continue/mcpServers/${name}.yaml`,
          lang: 'yaml',
          code: [
            `name: ${name}`,
            'version: 0.0.1',
            'schema: v1',
            'mcpServers:',
            `  - name: ${name}`,
            '    type: streamable-http',
            `    url: ${url}`,
            '    requestOptions:',
            '      headers:',
            `        Authorization: ${bearer}`,
            ...extraEntries.map(([header, value]) => `        ${header}: ${value}`),
          ].join('\n'),
        },
      ],
      note: 'For a self-signed certificate add requestOptions.caBundlePath.',
    },
    {
      id: 'claude-desktop',
      name: 'Claude Desktop (chat)',
      snippets: [
        {
          title: 'claude_desktop_config.json',
          lang: 'json',
          code: json({
            mcpServers: {
              [name]: {
                command: 'npx',
                args: ['-y', 'mcp-remote', url, '--transport', 'http-only',
                  '--header', `Authorization:${ref('SEMAPHORE_AUTH')}`,
                  ...extraEntries.flatMap(([header, value]) => ['--header', `${header}:${value}`])],
                env: { SEMAPHORE_AUTH: bearer },
              },
            },
          }),
        },
      ],
      note: 'The chat config does not accept a URL, so the local mcp-remote bridge connects instead. Add "--allow-http" for a plain http address; set NODE_EXTRA_CA_CERTS for a self-signed certificate. Custom connectors in Settings → Connectors connect from Anthropic\'s cloud and reach only a Semaphore open to the internet.',
    },
    {
      id: 'inspector',
      name: 'MCP Inspector',
      snippets: [
        {
          title: 'Command',
          lang: 'bash',
          code: `npx @modelcontextprotocol/inspector --cli ${url} --transport http --header "Authorization: ${bearer}"${headerFlags} --method tools/list`,
        },
      ],
      note: extraEntries.length === 0
        ? 'In the Inspector UI choose the Streamable HTTP transport and add the Authorization header.'
        : 'In the Inspector UI choose the Streamable HTTP transport and add the Authorization header and the headers of the command.',
    },
    {
      id: 'other',
      name: 'Other clients',
      snippets: [
        {
          title: 'Zoo Code (formerly Roo Code): .roo/mcp.json',
          lang: 'json',
          code: json({
            mcpServers: {
              [name]: {
                type: 'streamable-http',
                url,
                headers: withExtra(bearer),
              },
            },
          }),
        },
      ],
      note: `Any client that supports the Streamable HTTP transport and custom headers works: URL ${url}, ${extraEntries.length === 0 ? 'header' : 'headers'} ${otherHeaders}. The server does not support the legacy SSE transport or OAuth.`,
    },
  ];
}

// isPlainRemoteHttp reports whether the token would travel unencrypted over
// the network: an http address that is not the local machine.
export function isPlainRemoteHttp(url) {
  let parsed;
  try {
    parsed = new URL(url);
  } catch (e) {
    return false;
  }
  if (parsed.protocol !== 'http:') {
    return false;
  }
  return !['localhost', '127.0.0.1', '[::1]'].includes(parsed.hostname);
}
