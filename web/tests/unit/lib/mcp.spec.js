import { expect } from 'chai';
import {
  isValidMcpServerName,
  isValidMcpServerUrl,
  mcpServerRows,
  mcpServersRequest,
  mcpTokenChange,
  newMcpServerRow,
} from '@/lib/mcp';
import {
  AGENT_APPS, APP_FIELDS, APP_INVENTORY_TYPES, MCP_INVENTORY,
} from '@/lib/constants';

describe('lib/mcp', () => {
  describe('isValidMcpServerName', () => {
    [
      ['github', true],
      ['my_docs-2', true],
      ['a'.repeat(64), true],
      ['', false],
      ['a'.repeat(65), false],
      ['with space', false],
      ['dot.name', false],
      ['quote"', false],
      ['имя', false],
      [undefined, false],
    ].forEach(([name, valid]) => {
      it(`${valid ? 'accepts' : 'rejects'} ${JSON.stringify(name)}`, () => {
        expect(isValidMcpServerName(name)).to.equal(valid);
      });
    });
  });

  describe('isValidMcpServerUrl', () => {
    [
      ['https://mcp.example.com/mcp', true],
      ['http://localhost:8765/mcp', true],
      ['https://user:pass@mcp.example.com/', false],
      ['https://token@mcp.example.com/', false],
      ['ftp://mcp.example.com/', false],
      ['file:///etc/passwd', false],
      // eslint-disable-next-line no-script-url
      ['javascript:alert(1)', false],
      ['mcp.example.com', false],
      ['', false],
    ].forEach(([url, valid]) => {
      it(`${valid ? 'accepts' : 'rejects'} ${JSON.stringify(url)}`, () => {
        expect(isValidMcpServerUrl(url)).to.equal(valid);
      });
    });
  });

  it('turns the servers of the API into rows without a token value', () => {
    const rows = mcpServerRows([
      {
        id: 5, inventory_id: 2, name: 'github', url: 'https://a', enabled: false, has_token: true,
      },
      {
        id: 6, inventory_id: 2, name: 'docs', url: 'https://b', enabled: true, has_token: false,
      },
    ]);

    expect(rows).to.deep.equal([
      {
        key: 1, id: 5, name: 'github', url: 'https://a', enabled: false, hasToken: true, token: '', removeToken: false,
      },
      {
        key: 2, id: 6, name: 'docs', url: 'https://b', enabled: true, hasToken: false, token: '', removeToken: false,
      },
    ]);
    expect(mcpServerRows(undefined)).to.deep.equal([]);
  });

  it('starts a new row enabled and empty', () => {
    expect(newMcpServerRow(9)).to.deep.equal({
      key: 9, id: null, name: '', url: '', enabled: true, hasToken: false, token: '', removeToken: false,
    });
  });

  describe('mcpTokenChange', () => {
    it('creates a token for a server without one', () => {
      expect(mcpTokenChange({ hasToken: false, token: 'abc' })).to.deep.equal({ operation: 'create', secret: 'abc' });
    });

    it('replaces the token of a server with one', () => {
      expect(mcpTokenChange({ hasToken: true, token: 'abc' })).to.deep.equal({ operation: 'update', secret: 'abc' });
    });

    it('trims a pasted line break', () => {
      expect(mcpTokenChange({ hasToken: false, token: ' abc\n' })).to.deep.equal({ operation: 'create', secret: 'abc' });
    });

    it('deletes a token marked for removal', () => {
      expect(mcpTokenChange({ hasToken: true, token: '', removeToken: true })).to.deep.equal({ operation: 'delete' });
    });

    it('leaves the token alone otherwise', () => {
      expect(mcpTokenChange({ hasToken: true, token: '' })).to.equal(null);
      expect(mcpTokenChange({ hasToken: false, token: '   ' })).to.equal(null);
      // Nothing to delete on a server without a token.
      expect(mcpTokenChange({ hasToken: false, token: '', removeToken: true })).to.equal(null);
    });
  });

  it('builds the mcp_servers of a request', () => {
    const request = mcpServersRequest([
      {
        id: 5, name: 'github', url: 'https://a', enabled: true, hasToken: true, token: '', removeToken: false,
      },
      {
        id: 6, name: 'docs', url: 'https://b', enabled: false, hasToken: true, token: '', removeToken: true,
      },
      {
        id: null, name: 'new', url: 'https://c', enabled: true, hasToken: false, token: 't0k', removeToken: false,
      },
    ]);

    expect(request).to.deep.equal([
      {
        id: 5, name: 'github', url: 'https://a', enabled: true,
      },
      {
        id: 6, name: 'docs', url: 'https://b', enabled: false, token: { operation: 'delete' },
      },
      {
        name: 'new', url: 'https://c', enabled: true, token: { operation: 'create', secret: 't0k' },
      },
    ]);
  });

  it('gives the agent apps only the mcp inventory (db.TemplateApp.InventoryTypes)', () => {
    expect(AGENT_APPS).to.deep.equal(['claude', 'codex']);
    AGENT_APPS.forEach((app) => {
      expect(APP_INVENTORY_TYPES[app]).to.deep.equal([MCP_INVENTORY.type]);
      expect(APP_FIELDS[app].prompt).to.not.equal(undefined);
      expect(APP_FIELDS[app].allow_override_prompt).to.not.equal(undefined);
      expect(APP_FIELDS[app].inventory.optional).to.equal(true);
      expect(APP_FIELDS[app].playbook.directories).to.equal(true);
      expect(APP_FIELDS[app].vault).to.equal(undefined);
      expect(APP_FIELDS[app].allow_override_inventory).to.equal(undefined);
    });
    expect(APP_INVENTORY_TYPES.ansible).to.not.include(MCP_INVENTORY.type);
  });
});
