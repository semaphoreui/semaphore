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
    expect(allowed({ allow_write: true, allow_run: true, allow_destroy: true }))
      .to.deep.equal([['write', true], ['destroy', true], ['run', true]]);
    expect(allowed({ allow_write: false, allow_run: true, allow_destroy: false }))
      .to.deep.equal([['write', false], ['destroy', false], ['run', true]]);
    expect(allowed({ allow_write: true, allow_run: false, allow_destroy: false }))
      .to.deep.equal([['write', true], ['destroy', false], ['run', false]]);
    expect(allowed({ admin_allow_write: true }), 'the admin settings')
      .to.deep.equal([['write', false], ['destroy', false], ['run', false]]);
    expect(allowed(null), 'not loaded yet').to.deep.equal([['write', false], ['destroy', false], ['run', false]]);

    const every = { write: true, run: true, destroy: true };
    const on = (settings) => connection(settings, every).switches.map((s) => [s.key, s.on]);
    expect(on({ allow_write: true, allow_run: true, allow_destroy: true }))
      .to.deep.equal([['write', true], ['destroy', true], ['run', true]]);
    expect(on({ allow_run: true, allow_destroy: false }), 'a disabled switch shows off')
      .to.deep.equal([['write', false], ['destroy', false], ['run', true]]);
    expect(connection({}, {}).switches.map((s) => s.label), 'changes, deleting, then running')
      .to.deep.equal(['mcp_settings_admin_allow_write', 'mcp_settings_allow_destroy', 'mcp_settings_allow_run']);
  });

  it('allows the admin groups by the admin settings only', () => {
    const allowed = (settings) => connection(settings, {}, MCP_SETS.admin).switches
      .map((s) => [s.key, s.allowed]);
    const project = { allow_write: true, allow_run: true, allow_destroy: true };
    expect(allowed(project), 'the project settings').to.deep.equal([['write', false], ['destroy', false]]);
    expect(allowed({ admin_allow_write: true, admin_allow_destroy: false }))
      .to.deep.equal([['write', true], ['destroy', false]]);
    expect(allowed({ admin_allow_write: false, admin_allow_destroy: true }))
      .to.deep.equal([['write', false], ['destroy', true]]);
    expect(connection({ admin_allow_write: true }, {}, MCP_SETS.admin).switches.map((s) => s.label))
      .to.deep.equal(['mcp_settings_admin_allow_write', 'mcp_settings_allow_destroy']);
  });

  it('sends a header only for a group switched on and allowed by the server', () => {
    const every = { allow_write: true, allow_run: true, allow_destroy: true };
    expect(connection(every, { write: false, run: false, destroy: false }).headers, 'all start off').to.deep.equal({});
    expect(connection(every, { write: true }).headers).to.deep.equal({ 'X-Semaphore-MCP-Write': 'on' });
    expect(connection(every, { run: true }).headers).to.deep.equal({ 'X-Semaphore-MCP-Run': 'on' });
    expect(connection(every, { run: true, destroy: true }).headers).to.deep.equal({
      'X-Semaphore-MCP-Run': 'on', 'X-Semaphore-MCP-Destroy': 'on',
    });
    expect(Object.keys(connection(every, { destroy: true, run: true, write: true }).headers), 'in the order of the groups')
      .to.deep.equal(['X-Semaphore-MCP-Write', 'X-Semaphore-MCP-Destroy', 'X-Semaphore-MCP-Run']);
    const runAllowedOff = { allow_write: true, allow_run: false, allow_destroy: true };
    const runOff = connection(runAllowedOff, { run: true, destroy: true }).headers;
    expect(runOff, 'run switched off by the administrator')
      .to.deep.equal({ 'X-Semaphore-MCP-Destroy': 'on' });
    const adminWriteOnly = { allow_run: true, admin_allow_write: true };
    const writeOff = connection(adminWriteOnly, { write: true, run: true }).headers;
    expect(writeOff, 'write switched off by the administrator, whatever the admin setting')
      .to.deep.equal({ 'X-Semaphore-MCP-Run': 'on' });
    expect(connection(null, { write: true, run: true, destroy: true }).headers, 'settings not loaded').to.deep.equal({});
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
    expect(admin({ allow_write: true, allow_run: true, allow_destroy: true }, all), 'the project settings')
      .to.deep.equal({});
  });

  it('starts every group of both sets off', () => {
    expect(McpServer.data.call({}).connection).to.deep.equal({
      project: { write: false, destroy: false, run: false },
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
      expect(project({ write: true })).to.deep.equal({ 'X-Semaphore-MCP-Write': 'on' });
      expect(Object.keys(project({ destroy: true, run: true, write: true })), 'in the order of the groups')
        .to.deep.equal(['X-Semaphore-MCP-Write', 'X-Semaphore-MCP-Destroy', 'X-Semaphore-MCP-Run']);
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

  describe('tool list', () => {
    let mcpToolIcon;
    let mcpToolSummary;

    beforeAll(async () => {
      ({ mcpToolIcon, mcpToolSummary } = await import('@/lib/mcpClients'));
    });

    it('picks the icon of a tool by what it does', () => {
      expect(mcpToolIcon('get_task')).to.equal('mdi-eye-outline');
      expect(mcpToolIcon('list_tasks')).to.equal('mdi-format-list-bulleted');
      expect(mcpToolIcon('save_template')).to.equal('mdi-pencil-outline');
      expect(mcpToolIcon('generate_ssh_key')).to.equal('mdi-plus');
      expect(mcpToolIcon('delete_user')).to.equal('mdi-trash-can-outline');
      expect(mcpToolIcon('run_task')).to.equal('mdi-play-outline');
      expect(mcpToolIcon('stop_task')).to.equal('mdi-stop-circle-outline');
      expect(mcpToolIcon('confirm_task')).to.equal('mdi-check-circle-outline');
      expect(mcpToolIcon('reject_task')).to.equal('mdi-close-circle-outline');
      expect(mcpToolIcon('save_user_global_roles'), 'a prefix').to.equal('mdi-pencil-outline');
      expect(mcpToolIcon('run_workflow'), 'only run_task plays').to.equal('mdi-wrench-outline');
      expect(mcpToolIcon('something_new'), 'the fallback').to.equal('mdi-wrench-outline');
      expect(mcpToolIcon(undefined), 'the fallback').to.equal('mdi-wrench-outline');
    });

    it('shortens a description to its first sentence', () => {
      expect(mcpToolSummary('Returns a task. With its stages.')).to.equal('Returns a task.');
      expect(mcpToolSummary('Lists tasks: get_task.output shows more. Then this.'), 'a dot inside a word')
        .to.equal('Lists tasks: get_task.output shows more.');
      expect(mcpToolSummary('Stops it! Now.')).to.equal('Stops it!');
      expect(mcpToolSummary('No full stop')).to.equal('No full stop');
      expect(mcpToolSummary('Ends here.')).to.equal('Ends here.');
      expect(mcpToolSummary('')).to.equal('');
      expect(mcpToolSummary(undefined)).to.equal('');
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

  // tools answers GET /api/mcp/tools: a function of the request config, or what it returns.
  async function mountAt(path, settings, tools = { tools: [] }) {
    http.respond((config) => {
      if (config.url !== '/api/mcp/tools') {
        return settings;
      }
      return typeof tools === 'function' ? tools(config) : tools;
    });
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
      allow_write: true,
      allow_run: true,
      allow_destroy: true,
      admin_allow_write: true,
      admin_allow_destroy: false,
    });
    expect(wrapper.find('[data-testid="mcp-connect-admin"]').exists()).to.equal(true);
    expect(wrapper.find('[data-testid="mcp-intro"]').text()).to.contain('administering the server');
    expect(wrapper.find('[data-testid="mcp-endpoint"]').exists(), 'the address is in the snippets only').to.equal(false);
    expect(codes(wrapper)[0]).to.contain('--scope user semaphore-admin https://semaphore.example.com/api/mcp/admin');

    expect(disabled(wrapper, 'admin', 'write')).to.equal(false);
    expect(disabled(wrapper, 'admin', 'destroy'), 'switched off in Settings').to.equal(true);
    expect(wrapper.text(), 'no header hints').to.not.contain('Adds the header');
    expect(wrapper.text()).to.contain('Switched off by the administrator in Settings');
    expect(switchOf(wrapper, 'admin', 'write').text(), 'the hint of an allowed group')
      .to.contain('add and change users');
    expect(codes(wrapper).join('\n')).to.not.contain('X-Semaphore-MCP');

    const write = switchOf(wrapper, 'admin', 'write').find('input');
    write.element.checked = true;
    await write.trigger('input');
    expect(wrapper.vm.connection.admin.write).to.equal(true);
    expect(codes(wrapper)[0]).to.contain('--header "X-Semaphore-MCP-Write: on"');
    expect(codes(wrapper).join('\n')).to.not.contain('X-Semaphore-MCP-Destroy');
  });

  it('shows the project set at /api/mcp/project without the admin lines', async () => {
    const wrapper = await mountAt('/mcp/project', {
      allow_write: false,
      allow_run: false,
      allow_destroy: false,
      admin_allow_write: true,
      admin_allow_destroy: true,
    });
    expect(wrapper.find('[data-testid="mcp-connect-project"]').exists()).to.equal(true);
    expect(wrapper.find('[data-testid="mcp-intro"]').text()).to.contain('work inside a project');
    expect(codes(wrapper)[0]).to.contain('--scope user semaphore https://semaphore.example.com/api/mcp/project');
    expect(disabled(wrapper, 'project', 'write'), 'the admin write setting does not allow it').to.equal(true);
    expect(disabled(wrapper, 'project', 'run')).to.equal(true);
    expect(disabled(wrapper, 'project', 'destroy')).to.equal(true);
  });

  it('shows the project switches in the order changes, runs, deleting', async () => {
    const wrapper = await mountAt('/mcp/project', {
      allow_write: true,
      allow_run: false,
      allow_destroy: false,
      admin_allow_write: false,
      admin_allow_destroy: false,
    });
    const ids = wrapper.findAll('[data-testid^="mcp-connection-project-"]').map((w) => w.attributes('data-testid'));
    expect(ids).to.deep.equal(['mcp-connection-project-write', 'mcp-connection-project-destroy', 'mcp-connection-project-run']);
    expect(disabled(wrapper, 'project', 'write')).to.equal(false);
    expect(switchOf(wrapper, 'project', 'write').text()).to.contain('Allow the change tools')
      .and.contain('generate SSH keys');

    const write = switchOf(wrapper, 'project', 'write').find('input');
    write.element.checked = true;
    await write.trigger('input');
    expect(wrapper.vm.connection.project.write).to.equal(true);
    expect(codes(wrapper)[0]).to.contain('--header "X-Semaphore-MCP-Write: on"');
    expect(codes(wrapper).join('\n')).to.not.contain('X-Semaphore-MCP-Run');
  });

  it('puts the switches above the configs below md and to their right from md', async () => {
    const wrapper = await mountAt('/mcp/project', {
      allow_write: true,
      allow_run: true,
      allow_destroy: true,
      admin_allow_write: true,
      admin_allow_destroy: true,
    });
    const columns = wrapper.find('[data-testid="mcp-connect-project"] .v-row').findAll(':scope > .v-col');
    expect(columns).to.have.length(2);
    // From md the order of the markup holds: the connection left, the switches right.
    expect(columns[0].find('[data-testid="mcp-connection-title"]').exists()).to.equal(true);
    expect(columns[1].attributes('data-testid')).to.equal('mcp-connection-tools');
    expect(columns[0].classes()).to.include.members(['v-col--cols-12', 'v-col--cols-md-8']);
    expect(columns[1].classes()).to.include.members(['v-col--cols-12', 'v-col--cols-md-4', 'order-first', 'order-md-0']);
    expect(columns[0].classes().some((c) => c.startsWith('order-')), 'the connection keeps its place')
      .to.equal(false);
  });
});

describe('McpServer.vue tool list', () => {
  let McpServer;
  let http;
  let httpError;
  const wrappers = [];
  const all = {
    allow_write: true,
    allow_run: true,
    allow_destroy: true,
    admin_allow_write: true,
    admin_allow_destroy: true,
  };
  const tool = (name, description = `Does ${name}. More words that the page leaves out.`) => ({
    name, description, group: 'read', read_only: true, destructive: false, available: true,
  });
  // A tool GET /api/mcp/tools lists as not available for the reason, and the missing permission.
  const off = (name, reason, permission) => ({
    ...tool(name), available: false, reason, ...(permission ? { permission } : {}),
  });

  beforeAll(async () => {
    McpServer = (await import('@/views/McpServer.vue')).default;
    ({ httpError } = await import('./helpers/axiosMock'));
  });

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
    wrappers.splice(0).forEach((w) => w.unmount());
  });

  const settle = () => new Promise((resolve) => { setTimeout(resolve, 0); });

  async function mountAt(path, tools, settings = all) {
    http.respond((config) => {
      if (config.url === '/api/mcp/settings') {
        return settings;
      }
      return config.url === '/api/mcp/tools' ? tools(config) : [];
    });
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
    await settle();
    return wrapper;
  }

  const toolRequests = () => http.requests.filter((r) => r.url === '/api/mcp/tools');
  const names = (wrapper) => wrapper.findAll('[data-testid="mcp-tool-name"]').map((w) => w.text());
  const count = (wrapper) => String(wrapper.findAll('[data-testid="mcp-tool"]').length);

  async function toggle(wrapper, set, key) {
    const input = wrapper.find(`[data-testid="mcp-connection-${set}-${key}"]`).find('input');
    input.element.checked = !input.element.checked;
    await input.trigger('input');
    await settle();
  }

  it('renders the tools of the answer with their icons and first sentences', async () => {
    const wrapper = await mountAt('/mcp/project', () => ({
      tools: [tool('get_project'), tool('list_projects'), tool('save_project'), tool('run_task'),
        tool('delete_task', 'Deletes a task!  And its output.')],
    }));
    expect(names(wrapper)).to.deep.equal(['get_project', 'list_projects', 'save_project', 'run_task', 'delete_task']);
    expect(count(wrapper)).to.equal('5');
    expect(wrapper.find('[data-testid="mcp-tools-title"]').text()).to.equal('Tools');
    const items = wrapper.findAll('[data-testid="mcp-tool"]');
    expect(items[0].attributes('title'), 'no hover text').to.equal(undefined);
    expect(items[0].text()).to.contain('Does get_project.');
    expect(items[0].text()).to.not.contain('More words');
    expect(items[4].text()).to.contain('Deletes a task!');
    const icons = wrapper.findAll('[data-testid="mcp-tool-icon"]').map((w) => w.classes());
    expect(icons[0]).to.include('mdi-eye-outline');
    expect(icons[1]).to.include('mdi-format-list-bulleted');
    expect(icons[3]).to.include('mdi-play-outline');
    expect(icons[4]).to.include('mdi-trash-can-outline');
    expect(wrapper.find('[data-testid="mcp-tools-loading"]').exists()).to.equal(false);
    expect(wrapper.find('[data-testid="mcp-tools-error"]').exists()).to.equal(false);
  });

  it('asks for the set of the tab and the switched-on flags only, again on every switch', async () => {
    const wrapper = await mountAt('/mcp/project', (config) => ({
      tools: config.params.run ? [tool('list_projects'), tool('run_task')] : [tool('list_projects')],
    }));
    expect(toolRequests().map((r) => r.params)).to.deep.equal([{ set: 'project' }]);
    expect(count(wrapper)).to.equal('1');

    await toggle(wrapper, 'project', 'run');
    expect(toolRequests().map((r) => r.params)).to.deep.equal([{ set: 'project' }, { set: 'project', run: 'on' }]);
    expect(names(wrapper)).to.deep.equal(['list_projects', 'run_task']);
    expect(count(wrapper)).to.equal('2');

    await toggle(wrapper, 'project', 'destroy');
    expect(toolRequests()[2].params).to.deep.equal({ set: 'project', run: 'on', destroy: 'on' });
    await toggle(wrapper, 'project', 'run');
    expect(toolRequests()[3].params).to.deep.equal({ set: 'project', destroy: 'on' });
    expect(count(wrapper)).to.equal('1');
  });

  it('asks for the admin set on the admin tab, and not for a switch the administrator turned off', async () => {
    const wrapper = await mountAt('/mcp/admin', () => ({ tools: [tool('list_users')] }), {
      ...all, admin_allow_destroy: false,
    });
    expect(toolRequests().map((r) => r.params)).to.deep.equal([{ set: 'admin' }]);
    await toggle(wrapper, 'admin', 'write');
    expect(toolRequests()[1].params).to.deep.equal({ set: 'admin', write: 'on' });
    expect(names(wrapper)).to.deep.equal(['list_users']);
  });

  it('asks nothing on the tokens tab', async () => {
    await mountAt('/mcp', () => ({ tools: [] }));
    expect(toolRequests()).to.deep.equal([]);
  });

  it('shows the error of a failed request instead of the list', async () => {
    const wrapper = await mountAt('/mcp/project', () => {
      throw httpError(403, { error: 'The MCP server needs a Pro seat' });
    });
    const alert = wrapper.find('[data-testid="mcp-tools-error"]');
    expect(alert.exists()).to.equal(true);
    expect(alert.text()).to.contain('Could not load the tools of this connection');
    expect(alert.text()).to.contain('The MCP server needs a Pro seat');
    expect(names(wrapper)).to.deep.equal([]);
    expect(wrapper.findAll('[data-testid="mcp-tool"]').length).to.equal(0);
  });

  it('greys out a tool that is not available and names the reason in a badge', async () => {
    const wrapper = await mountAt('/mcp/admin', () => ({
      tools: [tool('list_project_members'), off('save_project_member', 'setting'),
        off('delete_project_member', 'connection'), off('list_users', 'permission', 'global.users.manage'),
        off('list_new_things', 'permission', 'global.new.manage')],
    }));
    const items = wrapper.findAll('[data-testid="mcp-tool"]');
    expect(names(wrapper), 'every tool, in the order of the answer').to.deep.equal([
      'list_project_members', 'save_project_member', 'delete_project_member', 'list_users', 'list_new_things',
    ]);
    expect(items.map((w) => w.attributes('data-available'))).to.deep.equal(['true', 'false', 'false', 'false', 'false']);
    const reason = (w) => w.find('[data-testid="mcp-tool-reason"]');
    expect(reason(items[0]).exists()).to.equal(false);
    expect(reason(items[1]).text()).to.equal('Off');
    expect(reason(items[2]).text()).to.equal('Off');
    expect(reason(items[3]).text(), 'the translated name of the permission').to.equal('Needs: Can manage users');
    expect(reason(items[4]).text(), 'an unknown permission by its ID').to.equal('Needs: global.new.manage');
    items.slice(1).forEach((item) => {
      expect(reason(item).classes()).to.include('v-chip--size-x-small');
      expect(item.find('[data-testid="mcp-tool-name"]').classes()).to.include('text-disabled');
      expect(item.find('[data-testid="mcp-tool-icon"]').classes()).to.include('text-disabled');
      expect(item.find('[data-testid="mcp-tool-icon"]').classes()).to.not.include('text-primary');
      expect(item.find('.McpServer__toolSummary').classes()).to.include('text-disabled');
      expect(item.find('.McpServer__toolSummary').classes()).to.not.include('text-medium-emphasis');
    });
  });

  it('draws an available tool as before: primary icon, medium-emphasis summary, no badge', async () => {
    const wrapper = await mountAt('/mcp/project', () => ({ tools: [tool('list_projects')] }));
    const item = wrapper.find('[data-testid="mcp-tool"]');
    expect(item.find('[data-testid="mcp-tool-reason"]').exists()).to.equal(false);
    expect(item.find('[data-testid="mcp-tool-name"]').classes()).to.not.include('text-disabled');
    expect(item.find('[data-testid="mcp-tool-icon"]').classes()).to.include('text-primary');
    expect(item.find('[data-testid="mcp-tool-icon"]').classes()).to.not.include('text-disabled');
    expect(item.find('.McpServer__toolSummary').classes()).to.include('text-medium-emphasis');
    expect(item.find('.McpServer__toolSummary').classes()).to.not.include('text-disabled');
  });

  it('refetches on a toggle and changes the state of a tool in place', async () => {
    const wrapper = await mountAt('/mcp/project', (config) => ({
      tools: [tool('list_projects'), config.params.run ? tool('run_task') : off('run_task', 'connection')],
    }));
    const runTask = () => wrapper.findAll('[data-testid="mcp-tool"]')[1];
    expect(runTask().attributes('data-available')).to.equal('false');
    expect(runTask().find('[data-testid="mcp-tool-reason"]').text()).to.equal('Off');

    await toggle(wrapper, 'project', 'run');
    expect(toolRequests().map((r) => r.params)).to.deep.equal([{ set: 'project' }, { set: 'project', run: 'on' }]);
    expect(names(wrapper)).to.deep.equal(['list_projects', 'run_task']);
    expect(runTask().attributes('data-available')).to.equal('true');
    expect(runTask().find('[data-testid="mcp-tool-reason"]').exists()).to.equal(false);
    expect(runTask().find('[data-testid="mcp-tool-name"]').classes()).to.not.include('text-disabled');

    await toggle(wrapper, 'project', 'run');
    expect(runTask().attributes('data-available')).to.equal('false');
  });

  it('shows the progress line while the list loads', async () => {
    let answer;
    const wrapper = await mountAt('/mcp/project', () => new Promise((resolve) => { answer = resolve; }));
    expect(wrapper.find('[data-testid="mcp-tools-loading"]').exists()).to.equal(true);
    answer({ tools: [tool('list_projects')] });
    await settle();
    expect(wrapper.find('[data-testid="mcp-tools-loading"]').exists()).to.equal(false);
    expect(names(wrapper)).to.deep.equal(['list_projects']);
  });
});

describe('McpServer.vue tokens and banners', () => {
  let McpServer;
  let http;
  const wrappers = [];
  const settings = {
    allow_write: false,
    allow_run: false,
    allow_destroy: false,
    admin_allow_write: false,
    admin_allow_destroy: false,
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
    expect(has(wrapper, 'mcp-settings-without-mcp'), 'MCP unavailable: no settings, even without a license').to.equal(false);
  });

  it('asks for a Pro seat and shows no settings, even to a system manager', async () => {
    const wrapper = await mountWith({
      features: { task_summary: true, mcp: false },
      subscription_state: 'active',
      mcp: { enabled: true },
    }, []);
    expect(has(wrapper, 'mcp-needs-pro-seat')).to.equal(true);
    expect(has(wrapper, 'mcp-assign-pro-seat')).to.equal(true);
    expect(has(wrapper, 'mcp-revoke-only'), 'no tokens to revoke').to.equal(false);
    expect(has(wrapper, 'mcp-settings-without-mcp'), 'MCP unavailable: no settings').to.equal(false);
  });

  it('says MCP is switched off and shows no settings', async () => {
    const wrapper = await mountWith({ features: { mcp: true }, mcp: { enabled: false } }, [], {
      admin: false,
    });
    expect(has(wrapper, 'mcp-disabled')).to.equal(true);
    expect(has(wrapper, 'mcp-only-pro')).to.equal(false);
    expect(has(wrapper, 'mcp-settings-without-mcp'), 'MCP unavailable: no settings').to.equal(false);
  });
});
