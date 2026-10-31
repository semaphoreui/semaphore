// How each MCP client connects to the Semaphore MCP server: Streamable HTTP
// with a static "Authorization: Bearer <token>" header. Field names differ
// between clients, and a wrong one is often ignored or switches the client to
// the legacy SSE transport, which the server does not speak. Checked against
// the clients' documentation in October 2026.

export const TOKEN_PLACEHOLDER = '<MCP_TOKEN>';

const json = (value) => JSON.stringify(value, null, 2);

// A ${name} reference for the client to expand, built so it is not read as a JS template.
const ref = (name) => ['$', '{', name, '}'].join('');

const bearer = `Bearer ${TOKEN_PLACEHOLDER}`;

const urlAndHeaders = (url) => ({
  mcpServers: {
    semaphore: {
      url,
      headers: { Authorization: bearer },
    },
  },
});

// Each client: id, name, and snippets — { title, lang, code }; note is optional.
export function mcpClients(url) {
  return [
    {
      id: 'claude-code',
      name: 'Claude Code',
      snippets: [
        {
          title: 'Command',
          lang: 'bash',
          code: `claude mcp add --transport http --scope user semaphore ${url} --header "Authorization: ${bearer}"`,
        },
        {
          title: '.mcp.json in the project root',
          lang: 'json',
          code: json({
            mcpServers: {
              semaphore: {
                type: 'http',
                url,
                headers: { Authorization: `Bearer ${ref('SEMAPHORE_MCP_TOKEN')}` },
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
          code: json(urlAndHeaders(url)),
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
              semaphore: {
                type: 'http',
                url,
                headers: { Authorization: bearer },
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
                id: 'semaphore-token',
                description: 'Semaphore MCP token',
                password: true,
              },
            ],
            servers: {
              semaphore: {
                type: 'http',
                url,
                headers: { Authorization: `Bearer ${ref('input:semaphore-token')}` },
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
          code: `export SEMAPHORE_MCP_TOKEN=${TOKEN_PLACEHOLDER}\ncodex mcp add semaphore --url ${url} --bearer-token-env-var SEMAPHORE_MCP_TOKEN`,
        },
        {
          title: '~/.codex/config.toml',
          lang: 'toml',
          code: `[mcp_servers.semaphore]\nurl = "${url}"\nbearer_token_env_var = "SEMAPHORE_MCP_TOKEN"`,
        },
      ],
      note: 'The Codex CLI, the IDE extension and ChatGPT desktop share ~/.codex/config.toml. Codex has no "headers" key.',
    },
    {
      id: 'gemini',
      name: 'Gemini CLI',
      snippets: [
        {
          title: 'Command',
          lang: 'bash',
          code: `gemini mcp add --transport http --scope user --header "Authorization: ${bearer}" semaphore ${url}`,
        },
        {
          title: '~/.gemini/settings.json',
          lang: 'json',
          code: json({
            mcpServers: {
              semaphore: {
                httpUrl: url,
                headers: { Authorization: bearer },
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
          code: json(urlAndHeaders(url)),
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
          code: json(urlAndHeaders(url)),
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
              semaphore: {
                url,
                headers: { Authorization: bearer },
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
          code: `cline mcp add semaphore ${url} --transport http --header "Authorization: ${bearer}" --yes`,
        },
        {
          title: 'cline_mcp_settings.json (Cline → MCP Servers → Configure)',
          lang: 'json',
          code: json({
            mcpServers: {
              semaphore: {
                type: 'streamableHttp',
                url,
                headers: { Authorization: bearer },
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
          title: '.continue/mcpServers/semaphore.yaml',
          lang: 'yaml',
          code: [
            'name: Semaphore',
            'version: 0.0.1',
            'schema: v1',
            'mcpServers:',
            '  - name: semaphore',
            '    type: streamable-http',
            `    url: ${url}`,
            '    requestOptions:',
            '      headers:',
            `        Authorization: ${bearer}`,
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
              semaphore: {
                command: 'npx',
                args: ['-y', 'mcp-remote', url, '--transport', 'http-only',
                  '--header', `Authorization:${ref('SEMAPHORE_AUTH')}`],
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
          code: `npx @modelcontextprotocol/inspector --cli ${url} --transport http --header "Authorization: ${bearer}" --method tools/list`,
        },
      ],
      note: 'In the Inspector UI choose the Streamable HTTP transport and add the Authorization header.',
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
              semaphore: {
                type: 'streamable-http',
                url,
                headers: { Authorization: bearer },
              },
            },
          }),
        },
      ],
      note: `Any client that supports the Streamable HTTP transport and a custom header works: URL ${url}, header "Authorization: ${bearer}". The server does not support the legacy SSE transport or OAuth.`,
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
