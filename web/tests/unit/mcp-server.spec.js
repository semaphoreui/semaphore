import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import { createRouter, createMemoryHistory } from 'vue-router';
import i18n from '@/plugins/i18';
import mockAxios from './helpers/axiosMock';

describe('McpServer.vue', () => {
  let McpServer;
  let MCP_SETS;

  beforeAll(async () => {
    McpServer = (await import('@/views/McpServer.vue')).default;
    ({ MCP_SETS } = await import('@/lib/mcpClients'));
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

  it('tells a user without a Pro seat on a Pro license apart from a missing license', () => {
    const withoutSeat = (state, features) => McpServer.computed.proLicenseWithoutSeat.call({
      features,
      systemInfo: { subscription_state: state },
    });
    expect(withoutSeat('active', { task_summary: true, mcp: false }), 'Pro, no seat').to.equal(true);
    expect(withoutSeat('active', { task_summary: true, mcp: true }), 'Pro with a seat or Enterprise').to.equal(false);
    expect(withoutSeat('', {}), 'no license').to.equal(false);
    expect(withoutSeat('expired', {}), 'an expired license').to.equal(false);
    expect(withoutSeat('active', {}), 'a plan the server does not know').to.equal(false);
  });

  it('asks for a Pro seat only while MCP is switched on', () => {
    const needsProSeat = (proLicenseWithoutSeat, enabled) => McpServer.computed.needsProSeat.call({
      proLicenseWithoutSeat,
      systemInfo: { mcp: { enabled } },
    });
    expect(needsProSeat(true, true)).to.equal(true);
    expect(needsProSeat(true, false), 'switched off: the page says so instead').to.equal(false);
    expect(needsProSeat(false, true)).to.equal(false);
  });

  it('takes the tab from the address and changes the address with the tab', () => {
    const { get, set } = McpServer.computed.tab;
    const at = (path, canManageSystem = true, canUseAdminMcp = canManageSystem) => get.call({
      $route: { path }, canManageSystem, canUseAdminMcp,
    });
    expect(at('/mcp')).to.equal('tokens');
    expect(at('/mcp/project')).to.equal('project');
    expect(at('/mcp/admin')).to.equal('admin');
    expect(at('/mcp/settings')).to.equal('settings');
    expect(at('/mcp/admin', false), 'no global permission').to.equal('tokens');
    expect(at('/mcp/settings', false), 'a system tab for anyone else').to.equal('tokens');
    expect(at('/mcp/admin', false, true), 'a global permission other than the system').to.equal('admin');
    expect(at('/mcp/settings', false, true), 'the settings need the system').to.equal('tokens');
    expect(at('/mcp/project', false)).to.equal('project');

    const pushed = [];
    const ctx = (path) => ({ $route: { path }, $router: { push: (p) => pushed.push(p) } });
    set.call(ctx('/mcp'), 'project');
    set.call(ctx('/mcp/project'), 'admin');
    set.call(ctx('/mcp/admin'), 'settings');
    set.call(ctx('/mcp/settings'), 'tokens');
    set.call(ctx('/mcp'), 'tokens');
    expect(pushed).to.deep.equal(['/mcp/project', '/mcp/admin', '/mcp/settings', '/mcp']);
  });

  it('keeps the old address of the project tab', async () => {
    const { routes } = (await import('@/router')).default.options;
    expect(routes.find((r) => r.path === '/mcp/connect').redirect).to.equal('/mcp/project');
    ['/mcp', '/mcp/project', '/mcp/admin', '/mcp/settings'].forEach((path) => {
      expect(routes.find((r) => r.path === path).component, path).to.equal(McpServer);
    });
  });

  it('shows the settings without MCP only to a system manager of a licensed instance', () => {
    const shown = (canManageSystem, features, proLicenseWithoutSeat) => McpServer.computed
      .settingsWithoutMcp.call({ canManageSystem, features, proLicenseWithoutSeat });
    expect(shown(true, { mcp: true }, false), 'MCP switched off').to.equal(true);
    expect(shown(true, { mcp: false }, true), 'no Pro seat of their own').to.equal(true);
    expect(shown(true, {}, false), 'no license: the upgrade banner only').to.equal(false);
    expect(shown(false, { mcp: true }, false), 'not a system manager').to.equal(false);
  });

  // The connection computeds of the page for one set; `on` — the switches of the connection.
  const connection = (mcpSettings, on, set = MCP_SETS.project) => {
    const ctx = { mcpSettings, mcpSet: set, connection: { [set.id]: on } };
    ctx.connectionSwitches = McpServer.computed.connectionSwitches.call(ctx);
    return {
      switches: ctx.connectionSwitches,
      headers: McpServer.computed.connectionHeaders.call(ctx),
    };
  };

  it('shows the admin tab to an admin or a user with any global permission', () => {
    const can = (user) => McpServer.computed.canUseAdminMcp.call({ user });
    expect(can({ admin: true })).to.equal(true);
    expect(can({ admin: false, global_permissions: 8 }), 'read audit only').to.equal(true);
    expect(can({ admin: false, global_permissions: 0 })).to.equal(false);
    expect(can({ admin: false }), 'a server that does not report them').to.equal(false);
    expect(can(undefined), 'not loaded yet').to.equal(false);
  });

  it('connects each tab to its own set', () => {
    const set = (tab) => McpServer.computed.mcpSet.call({ tab });
    expect(set('project')).to.equal(MCP_SETS.project);
    expect(set('admin')).to.equal(MCP_SETS.admin);
    expect(set('tokens'), 'a tab without a set').to.equal(MCP_SETS.project);
  });

  it('offers a connection only the groups the administrator allows', () => {
    const allowed = (settings) => connection(settings, {}).switches
      .map((s) => [s.key, s.allowed]);
    expect(allowed({ allow_run: true, allow_destroy: true })).to.deep.equal([['run', true], ['destroy', true]]);
    expect(allowed({ allow_run: true, allow_destroy: false })).to.deep.equal([['run', true], ['destroy', false]]);
    expect(allowed({ allow_run: false, allow_destroy: false })).to.deep.equal([['run', false], ['destroy', false]]);
    expect(allowed(null), 'not loaded yet').to.deep.equal([['run', false], ['destroy', false]]);

    const on = (settings) => connection(settings, { run: true, destroy: true }).switches
      .map((s) => [s.key, s.on]);
    expect(on({ allow_run: true, allow_destroy: true })).to.deep.equal([['run', true], ['destroy', true]]);
    expect(on({ allow_run: true, allow_destroy: false }), 'a disabled switch shows off')
      .to.deep.equal([['run', true], ['destroy', false]]);
  });

  it('allows the admin groups by the admin settings only', () => {
    const allowed = (settings) => connection(settings, {}, MCP_SETS.admin).switches
      .map((s) => [s.key, s.allowed]);
    const project = { allow_run: true, allow_destroy: true };
    expect(allowed(project), 'the project settings').to.deep.equal([['write', false], ['destroy', false]]);
    expect(allowed({ admin_allow_write: true, admin_allow_destroy: false }))
      .to.deep.equal([['write', true], ['destroy', false]]);
    expect(allowed({ admin_allow_write: false, admin_allow_destroy: true }))
      .to.deep.equal([['write', false], ['destroy', true]]);
    expect(connection({ admin_allow_write: true }, {}, MCP_SETS.admin).switches.map((s) => s.label))
      .to.deep.equal(['mcp_settings_admin_allow_write', 'mcp_settings_allow_destroy']);
  });

  it('sends a header only for a group switched on and allowed by the server', () => {
    const both = { allow_run: true, allow_destroy: true };
    expect(connection(both, { run: false, destroy: false }).headers, 'both start off').to.deep.equal({});
    expect(connection(both, { run: true }).headers).to.deep.equal({ 'X-Semaphore-MCP-Run': 'on' });
    expect(connection(both, { run: true, destroy: true }).headers).to.deep.equal({
      'X-Semaphore-MCP-Run': 'on', 'X-Semaphore-MCP-Destroy': 'on',
    });
    const runAllowedOff = { allow_run: false, allow_destroy: true };
    const runOff = connection(runAllowedOff, { run: true, destroy: true }).headers;
    expect(runOff, 'run switched off by the administrator')
      .to.deep.equal({ 'X-Semaphore-MCP-Destroy': 'on' });
    expect(connection(null, { run: true, destroy: true }).headers, 'settings not loaded').to.deep.equal({});
  });

  it('sends the admin headers only when switched on and allowed', () => {
    const admin = (settings, on) => connection(settings, on, MCP_SETS.admin).headers;
    const both = { admin_allow_write: true, admin_allow_destroy: true };
    expect(admin(both, {}), 'both start off').to.deep.equal({});
    expect(admin(both, { write: true })).to.deep.equal({ 'X-Semaphore-MCP-Write': 'on' });
    expect(admin(both, { write: true, destroy: true })).to.deep.equal({
      'X-Semaphore-MCP-Write': 'on', 'X-Semaphore-MCP-Destroy': 'on',
    });
    const all = { write: true, destroy: true };
    expect(admin({ admin_allow_write: false, admin_allow_destroy: true }, all), 'write switched off')
      .to.deep.equal({ 'X-Semaphore-MCP-Destroy': 'on' });
    expect(admin({ allow_run: true, allow_destroy: true }, all), 'the project settings')
      .to.deep.equal({});
  });

  it('starts every group of both sets off', () => {
    expect(McpServer.data.call({}).connection).to.deep.equal({
      project: { run: false, destroy: false },
      admin: { write: false, destroy: false },
    });
  });

  it('gives the client configs the headers and the server name of the set', () => {
    const configs = (mcpSet, connectionHeaders) => McpServer.computed.clients.call({
      endpoint: `https://semaphore.example.com/${mcpSet.path}`, mcpSet, connectionHeaders,
    });
    const project = configs(MCP_SETS.project, { 'X-Semaphore-MCP-Run': 'on' });
    const cursor = JSON.parse(project.find((c) => c.id === 'cursor').snippets[0].code);
    expect(Object.keys(cursor.mcpServers)).to.deep.equal(['semaphore']);
    expect(cursor.mcpServers.semaphore.headers).to.have.property('X-Semaphore-MCP-Run', 'on');

    const admin = configs(MCP_SETS.admin, { 'X-Semaphore-MCP-Write': 'on' });
    const adminCursor = JSON.parse(admin.find((c) => c.id === 'cursor').snippets[0].code);
    expect(Object.keys(adminCursor.mcpServers)).to.deep.equal(['semaphore-admin']);
    expect(adminCursor.mcpServers['semaphore-admin']).to.deep.equal({
      url: 'https://semaphore.example.com/api/mcp/admin',
      headers: { Authorization: 'Bearer <MCP_TOKEN>', 'X-Semaphore-MCP-Write': 'on' },
    });
  });

  const endpointOf = (systemInfo, mcpSet = MCP_SETS.project) => McpServer.computed.endpoint
    .call({ systemInfo, mcpSet });

  it('points the client at the endpoint of the set on the configured web host', () => {
    expect(endpointOf({ web_host: 'https://semaphore.example.com' }))
      .to.equal('https://semaphore.example.com/api/mcp/project');
    expect(endpointOf({ web_host: 'https://semaphore.example.com' }, MCP_SETS.admin))
      .to.equal('https://semaphore.example.com/api/mcp/admin');
    expect(endpointOf({})).to.equal(new URL('api/mcp/project', document.baseURI).href);
    expect(endpointOf({}, MCP_SETS.admin)).to.equal(new URL('api/mcp/admin', document.baseURI).href);
  });

  it('keeps the path of web_host and does not double the slash', () => {
    const endpoint = (webHost, set) => endpointOf({ web_host: webHost }, set);
    expect(endpoint('https://example.com/semaphore')).to.equal('https://example.com/semaphore/api/mcp/project');
    expect(endpoint('https://example.com/semaphore/')).to.equal('https://example.com/semaphore/api/mcp/project');
    expect(endpoint('https://example.com/semaphore/', MCP_SETS.admin))
      .to.equal('https://example.com/semaphore/api/mcp/admin');
    expect(endpoint('https://example.com/')).to.equal('https://example.com/api/mcp/project');
    expect(endpoint('not a url'), 'an unparsable web_host')
      .to.equal(new URL('api/mcp/project', document.baseURI).href);
  });
});

describe('Tokens.vue kinds', () => {
  let Tokens;

  beforeAll(async () => {
    Tokens = (await import('@/views/Tokens.vue')).default;
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

  it('draws nothing for a page that hides an empty list', () => {
    const shown = (items, hideEmpty) => Tokens.computed.shown.call({ items, hideEmpty });
    expect(shown(null, false), 'not loaded yet').to.equal(false);
    expect(shown([], false), 'an empty list on the tokens page').to.equal(true);
    expect(shown([], true), 'no tokens to revoke').to.equal(false);
    expect(shown([{ id: 'abcdefgh' }], true), 'tokens to revoke').to.equal(true);
  });

  it('refuses an unknown kind', () => {
    expect(Tokens.props.kind.validator('api')).to.equal(true);
    expect(Tokens.props.kind.validator('mcp')).to.equal(true);
    expect(Tokens.props.kind.validator('ssh')).to.equal(false);
  });
});

describe('lib/mcpClients', () => {
  let mcpClients;
  let mcpConnectionHeaders;
  let isPlainRemoteHttp;
  let MCP_SETS;

  beforeAll(async () => {
    ({
      mcpClients, mcpConnectionHeaders, isPlainRemoteHttp, MCP_SETS,
    } = await import('@/lib/mcpClients'));
  });

  const url = 'https://semaphore.example.com/api/mcp/project';
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

  describe('connection headers', () => {
    const run = { 'X-Semaphore-MCP-Run': 'on' };
    const both = { 'X-Semaphore-MCP-Run': 'on', 'X-Semaphore-MCP-Destroy': 'on' };
    const all = (headers) => mcpClients(url, { headers });

    it('builds the headers of the switches, off by default', () => {
      const project = (on) => mcpConnectionHeaders(MCP_SETS.project, on);
      expect(project()).to.deep.equal({});
      expect(project({ run: false, destroy: false })).to.deep.equal({});
      expect(project({ run: true })).to.deep.equal(run);
      expect(project({ destroy: true })).to.deep.equal({ 'X-Semaphore-MCP-Destroy': 'on' });
      expect(project({ run: true, destroy: true })).to.deep.equal(both);
      expect(project({ write: true }), 'a group of the other set').to.deep.equal({});
    });

    it('builds the headers of the admin switches', () => {
      const admin = (on) => mcpConnectionHeaders(MCP_SETS.admin, on);
      expect(admin()).to.deep.equal({});
      expect(admin({ write: true })).to.deep.equal({ 'X-Semaphore-MCP-Write': 'on' });
      expect(admin({ destroy: true })).to.deep.equal({ 'X-Semaphore-MCP-Destroy': 'on' });
      expect(Object.keys(admin({ destroy: true, write: true })), 'in the order of the groups')
        .to.deep.equal(['X-Semaphore-MCP-Write', 'X-Semaphore-MCP-Destroy']);
      expect(admin({ run: true }), 'a group of the other set').to.deep.equal({});
    });

    it('adds no header to any client while both switches are off', () => {
      all({}).forEach((c) => {
        c.snippets.forEach((s) => {
          expect(s.code, `${c.id}: ${s.title}`).to.not.contain('X-Semaphore-MCP');
        });
        expect(c.note || '', c.id).to.not.contain('X-Semaphore-MCP');
      });
      expect(all({}), 'the default options').to.deep.equal(mcpClients(url));
    });

    it('puts every header into every client', () => {
      all(both).forEach((c) => {
        const text = c.snippets.map((s) => s.code).join('\n');
        Object.keys(both).forEach((name) => {
          expect(text, `${c.id}: ${name}`).to.contain(name);
        });
        c.snippets.filter((s) => s.lang === 'json').forEach((s) => {
          expect(() => JSON.parse(s.code), `${c.id}: ${s.title}`).to.not.throw();
        });
      });
    });

    it('puts the headers next to Authorization in the JSON configs', () => {
      const headersOf = (id, i, path) => {
        const client = mcpClients(url, { headers: both }).find((c) => c.id === id);
        return path.reduce((o, k) => o[k], JSON.parse(client.snippets[i].code)).headers;
      };
      const server = ['mcpServers', 'semaphore'];
      [['claude-code', 1, server], ['cursor', 0, server], ['vscode', 0, server],
        ['vscode', 1, ['servers', 'semaphore']], ['gemini', 1, server], ['devin', 0, server],
        ['jetbrains', 0, server], ['zed', 0, ['context_servers', 'semaphore']], ['cline', 1, server],
        ['other', 0, server]].forEach(([id, i, path]) => {
        const headers = headersOf(id, i, path);
        expect(Object.keys(headers), `${id} ${i}`).to.deep.equal(['Authorization', ...Object.keys(both)]);
        expect(headers['X-Semaphore-MCP-Run'], id).to.equal('on');
      });
    });

    it('repeats --header in the commands', () => {
      ['claude-code', 'gemini', 'cline', 'inspector'].forEach((id) => {
        const client = mcpClients(url, { headers: both }).find((c) => c.id === id);
        const command = client.snippets[0].code;
        expect(command, id).to.contain('--header "X-Semaphore-MCP-Run: on"');
        expect(command, id).to.contain('--header "X-Semaphore-MCP-Destroy: on"');
      });
    });

    it('writes the headers in the form of Codex, Continue and mcp-remote', () => {
      const clients = mcpClients(url, { headers: run });
      const codex = clients.find((c) => c.id === 'codex');
      expect(codex.snippets[1].code).to.contain('http_headers = { "X-Semaphore-MCP-Run" = "on" }');
      expect(codex.note, 'codex mcp add takes no headers').to.contain('http_headers');
      expect(clients.find((c) => c.id === 'continue').snippets[0].code)
        .to.contain('      headers:\n        Authorization: Bearer <MCP_TOKEN>\n        X-Semaphore-MCP-Run: on');
      const desktop = JSON.parse(clients.find((c) => c.id === 'claude-desktop').snippets[0].code);
      expect(desktop.mcpServers.semaphore.args.slice(-2)).to.deep.equal(['--header', 'X-Semaphore-MCP-Run:on']);
      expect(clients.find((c) => c.id === 'other').note).to.contain('"X-Semaphore-MCP-Run: on"');
    });
  });

  describe('server name', () => {
    const adminUrl = 'https://semaphore.example.com/api/mcp/admin';
    const admin = () => mcpClients(adminUrl, { serverName: MCP_SETS.admin.serverName });

    it('names the server semaphore by default and semaphore-admin for the admin set', () => {
      expect(MCP_SETS.project.serverName).to.equal('semaphore');
      expect(MCP_SETS.admin.serverName).to.equal('semaphore-admin');
      expect(mcpClients(url, { serverName: 'semaphore' }), 'the default name').to.deep.equal(mcpClients(url));
    });

    it('uses the server name of the set everywhere a client names the server', () => {
      // The Inspector command connects once and names no server.
      admin().filter((c) => c.id !== 'inspector').forEach((c) => {
        c.snippets.forEach((s) => {
          expect(s.code, `${c.id}: ${s.title}`).to.contain('semaphore-admin');
          const rest = s.code.split(adminUrl).join('').split('semaphore-admin').join('');
          expect(rest, `${c.id}: ${s.title}`).to.not.match(/semaphore/);
        });
      });
    });

    it('keys the JSON configs by the server name', () => {
      const keys = (id, i, path) => {
        const client = admin().find((c) => c.id === id);
        return Object.keys(path.reduce((o, k) => o[k], JSON.parse(client.snippets[i].code)));
      };
      [['claude-code', 1, 'mcpServers'], ['cursor', 0, 'mcpServers'], ['vscode', 0, 'mcpServers'],
        ['vscode', 1, 'servers'], ['gemini', 1, 'mcpServers'], ['devin', 0, 'mcpServers'],
        ['jetbrains', 0, 'mcpServers'], ['zed', 0, 'context_servers'], ['cline', 1, 'mcpServers'],
        ['claude-desktop', 0, 'mcpServers'], ['other', 0, 'mcpServers']].forEach(([id, i, key]) => {
        expect(keys(id, i, [key]), `${id} ${i}`).to.deep.equal(['semaphore-admin']);
      });
      const codex = admin().find((c) => c.id === 'codex');
      expect(codex.snippets[0].code).to.contain('codex mcp add semaphore-admin --url');
      expect(codex.snippets[1].code).to.contain('[mcp_servers.semaphore-admin]');
      expect(admin().find((c) => c.id === 'claude-code').snippets[0].code)
        .to.contain(`--scope user semaphore-admin ${adminUrl}`);
    });
  });

  it('warns about plain http only for a remote address', () => {
    expect(isPlainRemoteHttp('http://semaphore.internal:3000/api/mcp')).to.equal(true);
    expect(isPlainRemoteHttp('https://semaphore.internal/api/mcp')).to.equal(false);
    expect(isPlainRemoteHttp('http://localhost:3000/api/mcp')).to.equal(false);
    expect(isPlainRemoteHttp('http://127.0.0.1:3000/api/mcp')).to.equal(false);
    expect(isPlainRemoteHttp('not a url')).to.equal(false);
  });
});

describe('McpServer.vue connection tabs', () => {
  let McpServer;
  let http;
  const wrappers = [];

  beforeAll(async () => {
    McpServer = (await import('@/views/McpServer.vue')).default;
  });

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
    wrappers.splice(0).forEach((w) => w.unmount());
  });

  async function mountAt(path, settings) {
    http.respond(() => settings);
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/:pathMatch(.*)*', component: { render: () => null } }],
    });
    await router.push(path);
    const wrapper = mount(McpServer, {
      props: {
        user: { admin: true },
        systemInfo: { features: { mcp: true }, mcp: { enabled: true }, web_host: 'https://semaphore.example.com' },
      },
      global: { plugins: [router, i18n] },
    });
    wrappers.push(wrapper);
    await new Promise((resolve) => { setTimeout(resolve, 0); });
    return wrapper;
  }

  const switchOf = (wrapper, set, key) => wrapper.find(`[data-testid="mcp-connection-${set}-${key}"]`);
  // Vuetify 3+: the test id is on the root of the switch, `disabled` on its checkbox
  const disabled = (wrapper, set, key) => switchOf(wrapper, set, key).find('input').element.disabled;
  const codes = (wrapper) => wrapper.findAll('pre').map((w) => w.text());

  it('shows the admin set with its switches limited by the admin settings', async () => {
    const wrapper = await mountAt('/mcp/admin', {
      allow_run: true, allow_destroy: true, admin_allow_write: true, admin_allow_destroy: false,
    });
    expect(wrapper.find('[data-testid="mcp-connect-admin"]').exists()).to.equal(true);
    expect(codes(wrapper)[0]).to.equal('https://semaphore.example.com/api/mcp/admin');
    expect(codes(wrapper)[1]).to.contain('--scope user semaphore-admin https://semaphore.example.com/api/mcp/admin');

    expect(disabled(wrapper, 'admin', 'write')).to.equal(false);
    expect(disabled(wrapper, 'admin', 'destroy'), 'switched off in Settings').to.equal(true);
    expect(wrapper.text()).to.contain('Adds the header X-Semaphore-MCP-Write: on');
    expect(wrapper.text()).to.contain('Switched off by the administrator in Settings');
    expect(wrapper.find('[data-testid="mcp-admin-groups"]').text()).to.contain('delete_active_task');
    expect(wrapper.find('[data-testid="mcp-admin-hidden-tools"]').exists()).to.equal(true);
    expect(codes(wrapper).join('\n')).to.not.contain('X-Semaphore-MCP');

    const write = switchOf(wrapper, 'admin', 'write').find('input');
    write.element.checked = true;
    await write.trigger('input');
    expect(wrapper.vm.connection.admin.write).to.equal(true);
    expect(codes(wrapper)[1]).to.contain('--header "X-Semaphore-MCP-Write: on"');
    expect(codes(wrapper).join('\n')).to.not.contain('X-Semaphore-MCP-Destroy');
  });

  it('shows the project set at /api/mcp/project without the admin lines', async () => {
    const wrapper = await mountAt('/mcp/project', {
      allow_run: false, allow_destroy: false, admin_allow_write: true, admin_allow_destroy: true,
    });
    expect(wrapper.find('[data-testid="mcp-connect-project"]').exists()).to.equal(true);
    expect(codes(wrapper)[0]).to.equal('https://semaphore.example.com/api/mcp/project');
    expect(codes(wrapper)[1]).to.contain('--scope user semaphore https://');
    expect(disabled(wrapper, 'project', 'run')).to.equal(true);
    expect(disabled(wrapper, 'project', 'destroy')).to.equal(true);
    expect(wrapper.find('[data-testid="mcp-admin-groups"]').exists()).to.equal(false);
  });
});

describe('McpServer.vue tokens and banners', () => {
  let McpServer;
  let http;
  const wrappers = [];
  const settings = {
    allow_run: false, allow_destroy: false, admin_allow_write: false, admin_allow_destroy: false,
  };

  beforeAll(async () => {
    McpServer = (await import('@/views/McpServer.vue')).default;
  });

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
    wrappers.splice(0).forEach((w) => w.unmount());
  });

  async function mountWith(systemInfo, tokens, user = { admin: true }) {
    http.respond((config) => (config.url === '/api/mcp/settings' ? settings : tokens));
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/:pathMatch(.*)*', component: { render: () => null } }],
    });
    await router.push('/mcp');
    const wrapper = mount(McpServer, {
      props: { user, systemInfo },
      global: { plugins: [router, i18n] },
    });
    wrappers.push(wrapper);
    await new Promise((resolve) => { setTimeout(resolve, 0); });
    return wrapper;
  }

  const has = (wrapper, id) => wrapper.find(`[data-testid="${id}"]`).exists();
  const token = { id: 'abcdefgh', name: 'agent', created: '2026-11-01T10:00:00Z' };

  it('lists the MCP tokens under the page toolbar of its own', async () => {
    const wrapper = await mountWith({ features: { mcp: true }, mcp: { enabled: true } }, [token]);
    expect(http.requests.map((r) => r.url)).to.include('/api/user/mcp-tokens');
    expect(has(wrapper, 'mcp-tab-tokens')).to.equal(true);
    expect(has(wrapper, 'mcp-new-token')).to.equal(true);
    expect(has(wrapper, 'tokens-table')).to.equal(true);
    expect(has(wrapper, 'tokens-title'), 'the toolbar of the API tokens page').to.equal(false);
  });

  it('offers the upgrade without a license and lists the tokens left for revoking', async () => {
    const wrapper = await mountWith({ features: {}, mcp: { enabled: true } }, [token]);
    expect(has(wrapper, 'mcp-only-pro')).to.equal(true);
    expect(has(wrapper, 'mcp-tabs')).to.equal(false);
    expect(has(wrapper, 'mcp-new-token')).to.equal(false);
    expect(has(wrapper, 'mcp-revoke-only')).to.equal(true);
    expect(has(wrapper, 'tokens-rowDelete')).to.equal(true);
    expect(has(wrapper, 'mcp-settings-without-mcp'), 'no license: no settings').to.equal(false);
  });

  it('asks for a Pro seat and still shows a system manager the settings', async () => {
    const wrapper = await mountWith({
      features: { task_summary: true, mcp: false },
      subscription_state: 'active',
      mcp: { enabled: true },
    }, []);
    expect(has(wrapper, 'mcp-needs-pro-seat')).to.equal(true);
    expect(has(wrapper, 'mcp-assign-pro-seat')).to.equal(true);
    expect(has(wrapper, 'mcp-revoke-only'), 'no tokens to revoke').to.equal(false);
    expect(has(wrapper, 'mcp-settings-without-mcp')).to.equal(true);
  });

  it('says MCP is switched off', async () => {
    const wrapper = await mountWith({ features: { mcp: true }, mcp: { enabled: false } }, [], {
      admin: false,
    });
    expect(has(wrapper, 'mcp-disabled')).to.equal(true);
    expect(has(wrapper, 'mcp-only-pro')).to.equal(false);
    expect(has(wrapper, 'mcp-settings-without-mcp'), 'not a system manager').to.equal(false);
  });
});
