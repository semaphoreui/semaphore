import './setup';
import { expect } from 'chai';

describe('McpServer.vue', () => {
  let McpServer;

  before(() => {
    // eslint-disable-next-line global-require
    McpServer = require('@/views/McpServer.vue').default;
  });

  const available = (features, mcp) => McpServer.computed.available.call({
    features,
    systemInfo: { mcp },
  });

  it('manages MCP tokens only when the user may use MCP and it is switched on', () => {
    expect(available({ mcp: true }, { enabled: true })).to.equal(true);
    expect(available({ mcp: false }, { enabled: true }), 'no license or Pro seat').to.equal(false);
    expect(available({}, { enabled: true }), 'a build without MCP').to.equal(false);
    expect(available({ mcp: true }, { enabled: false }), 'switched off').to.equal(false);
    expect(available({ mcp: true }, undefined), 'a server that does not report MCP').to.equal(false);
  });

  it('takes the tab from the address and changes the address with the tab', () => {
    const { get, set } = McpServer.computed.tab;
    expect(get.call({ $route: { path: '/mcp' } })).to.equal('tokens');
    expect(get.call({ $route: { path: '/mcp/connect' } })).to.equal('connect');

    const pushed = [];
    const ctx = (path) => ({ $route: { path }, $router: { push: (p) => pushed.push(p) } });
    set.call(ctx('/mcp'), 'connect');
    set.call(ctx('/mcp/connect'), 'tokens');
    set.call(ctx('/mcp'), 'tokens');
    expect(pushed).to.deep.equal(['/mcp/connect', '/mcp']);
  });

  it('points the client at /api/mcp of the configured web host', () => {
    const endpoint = (systemInfo) => McpServer.computed.endpoint.call({ systemInfo });
    expect(endpoint({ web_host: 'https://semaphore.example.com' }))
      .to.equal('https://semaphore.example.com/api/mcp');
    expect(endpoint({})).to.equal(new URL('api/mcp', document.baseURI).href);
  });

  it('keeps the path of web_host and does not double the slash', () => {
    const endpoint = (webHost) => McpServer.computed.endpoint
      .call({ systemInfo: { web_host: webHost } });
    expect(endpoint('https://example.com/semaphore')).to.equal('https://example.com/semaphore/api/mcp');
    expect(endpoint('https://example.com/semaphore/')).to.equal('https://example.com/semaphore/api/mcp');
    expect(endpoint('https://example.com/')).to.equal('https://example.com/api/mcp');
    expect(endpoint('not a url'), 'an unparsable web_host').to.equal(new URL('api/mcp', document.baseURI).href);
  });
});

describe('Tokens.vue kinds', () => {
  let Tokens;

  before(() => {
    // eslint-disable-next-line global-require
    Tokens = require('@/views/Tokens.vue').default;
  });

  const urls = (kind) => {
    const ctx = { kind, itemId: 'abcdefgh' };
    ctx.getItemsUrl = () => Tokens.methods.getItemsUrl.call(ctx);
    return {
      items: Tokens.methods.getItemsUrl.call(ctx),
      single: Tokens.methods.getSingleItemUrl.call(ctx),
      event: Tokens.methods.getEventName.call(ctx),
    };
  };

  it('manages REST API tokens by default', () => {
    expect(Tokens.props.kind.default).to.equal('api');
    expect(urls('api')).to.deep.equal({
      items: '/api/user/tokens',
      single: '/api/user/tokens/abcdefgh',
      event: 'i-token',
    });
  });

  it('manages MCP tokens on their own routes', () => {
    expect(urls('mcp')).to.deep.equal({
      items: '/api/user/mcp-tokens',
      single: '/api/user/mcp-tokens/abcdefgh',
      event: 'i-mcp-token',
    });
  });

  it('lets every user revoke their own tokens, whatever the project role', () => {
    expect(Tokens.methods.allowActions.call({ can: () => false, isAdmin: false })).to.equal(true);
  });

  it('refuses an unknown kind', () => {
    expect(Tokens.props.kind.validator('api')).to.equal(true);
    expect(Tokens.props.kind.validator('mcp')).to.equal(true);
    expect(Tokens.props.kind.validator('ssh')).to.equal(false);
  });
});

describe('lib/mcpClients', () => {
  let mcpClients;
  let isPlainRemoteHttp;

  before(() => {
    // eslint-disable-next-line global-require
    ({ mcpClients, isPlainRemoteHttp } = require('@/lib/mcpClients'));
  });

  const url = 'https://semaphore.example.com/api/mcp';
  const byId = (id) => mcpClients(url).find((c) => c.id === id);
  const parsed = (id, i = 0) => JSON.parse(byId(id).snippets[i].code);

  it('gives every client a unique id and the endpoint in its snippets', () => {
    const clients = mcpClients(url);
    expect(new Set(clients.map((c) => c.id)).size).to.equal(clients.length);
    clients.forEach((c) => {
      expect(c.snippets.length, c.id).to.be.above(0);
      expect(c.snippets.some((s) => s.code.includes(url)), c.id).to.equal(true);
      c.snippets.filter((s) => s.lang === 'json').forEach((s) => {
        expect(() => JSON.parse(s.code), `${c.id}: ${s.title}`).to.not.throw();
      });
    });
  });

  it('uses the field names each client expects for Streamable HTTP', () => {
    expect(parsed('claude-code', 1).mcpServers.semaphore.type).to.equal('http');
    expect(parsed('gemini', 1).mcpServers.semaphore).to.have.property('httpUrl', url)
      .and.not.have.property('url');
    expect(parsed('cline', 1).mcpServers.semaphore.type).to.equal('streamableHttp');
    expect(parsed('zed').context_servers.semaphore).to.not.have.property('type');
    expect(parsed('vscode', 1).servers.semaphore.type).to.equal('http');
    expect(parsed('cursor').mcpServers.semaphore).to.have.property('url', url);
  });

  it('configures Codex in TOML without a headers key', () => {
    const toml = byId('codex').snippets[1].code;
    expect(toml).to.contain('[mcp_servers.semaphore]');
    expect(toml).to.contain('bearer_token_env_var');
    expect(toml).to.not.contain('headers');
  });

  it('connects Claude Desktop through the mcp-remote bridge', () => {
    const server = parsed('claude-desktop').mcpServers.semaphore;
    expect(server.command).to.equal('npx');
    expect(server.args).to.include('mcp-remote').and.include(url);
    expect(server).to.not.have.property('url');
  });

  it('warns about plain http only for a remote address', () => {
    expect(isPlainRemoteHttp('http://semaphore.internal:3000/api/mcp')).to.equal(true);
    expect(isPlainRemoteHttp('https://semaphore.internal/api/mcp')).to.equal(false);
    expect(isPlainRemoteHttp('http://localhost:3000/api/mcp')).to.equal(false);
    expect(isPlainRemoteHttp('http://127.0.0.1:3000/api/mcp')).to.equal(false);
    expect(isPlainRemoteHttp('not a url')).to.equal(false);
  });
});
