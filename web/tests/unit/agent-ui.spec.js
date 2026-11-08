import './setup';
import { expect } from 'chai';
import Inventory from '@/views/project/Inventory.vue';
import TemplateForm from '@/components/TemplateForm.vue';
import { MCP_INVENTORY } from '@/lib/constants';

describe('Inventory.vue inventory kinds', () => {
  const kinds = (activeAppIds) => Inventory.computed.kinds.call({ appsMixin: { activeAppIds } });

  it('offers the MCP inventory once an agent app is active', () => {
    expect(kinds(['ansible', 'terraform'])).to.deep.equal(['ansible']);
    expect(kinds(['ansible', 'claude'])).to.deep.equal(['ansible', 'mcp']);
    expect(kinds(['codex'])).to.deep.equal(['ansible', 'mcp']);
  });

  it('edits an mcp inventory as its own kind', () => {
    expect(Inventory.methods.getAppByType('mcp')).to.equal('mcp');
    expect(Inventory.methods.getAppByType('static')).to.equal('ansible');
    expect(Inventory.methods.getAppByType('terraform-workspace')).to.equal('terraform');
  });

  it('names the MCP kind with its own icon and title', () => {
    const page = { getAppIcon: (app) => `icon:${app}`, getAppColor: (app) => `color:${app}` };

    expect(Inventory.methods.kindIcon.call(page, 'mcp')).to.equal(MCP_INVENTORY.icon);
    expect(Inventory.methods.kindColor.call(page, 'mcp')).to.equal(undefined);
    expect(Inventory.methods.kindTitle.call(page, 'mcp')).to.equal('MCP Inventory');
    expect(Inventory.methods.kindIcon.call(page, 'ansible')).to.equal('icon:ansible');
    expect(Inventory.methods.kindTitle.call(page, 'ansible')).to.equal('Ansible Inventory');
  });
});

describe('TemplateForm.vue agent block', () => {
  it('gives the agent apps a block of their own', () => {
    ['claude', 'codex', 'ansible', '', 'terraform', 'tofu'].forEach((app) => {
      expect(TemplateForm.computed.needAppBlock.call({ app }), app).to.equal(true);
    });
    ['bash', 'python', 'powershell', 'terragrunt', 'pulumi'].forEach((app) => {
      expect(TemplateForm.computed.needAppBlock.call({ app }), app).to.equal(false);
    });
  });

  it('keeps the loader height of the apps with a block', () => {
    const height = (app, taskType) => TemplateForm.computed.loaderHeight.call({
      taskType,
      needAppBlock: TemplateForm.computed.needAppBlock.call({ app }),
    });

    expect([height('ansible'), height('bash')]).to.deep.equal([564, 514]);
    expect([height('claude', 'build'), height('codex', 'deploy')]).to.deep.equal([626, 676]);
  });
});
