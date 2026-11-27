import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect } from 'chai';
import { hasHelpDoc, helpDocKeys, loadHelpDoc } from '@/help';

const SRC = path.join(path.dirname(fileURLToPath(import.meta.url)), '../../src');

function source(file) {
  return fs.readFileSync(path.join(SRC, file), 'utf8');
}

const formSource = source('components/InventoryForm.vue');
const viewSource = source('views/project/Inventory.vue');

const FIELD_TAG = /<(v-text-field|v-autocomplete|v-select|v-checkbox|v-switch|v-textarea|CodeMirror)\b/g;

// The inputs of the form, in the order of the template: the `data-testid` of every field
// component. A field added to the form without a section in the document fails this spec.
function formInputs(text) {
  const template = text.slice(0, text.indexOf('<style'));
  return [...template.matchAll(FIELD_TAG)].map((tag) => {
    // The end of the opening tag: the first `>` outside a quoted attribute value (a rule such as
    // `v => !!v` has one inside).
    let quote = null;
    let end = tag.index;
    while (end < template.length && (quote || template[end] !== '>')) {
      const ch = template[end];
      if (quote) {
        quote = ch === quote ? null : quote;
      } else if (ch === '"' || ch === "'") {
        quote = ch;
      }
      end += 1;
    }
    const testId = /\sdata-testid="([^"]+)"/.exec(template.slice(tag.index, end));
    return testId ? testId[1] : `${tag[1]} without data-testid`;
  });
}

describe('help of the Ansible inventory dialog', () => {
  it('is a document of the registry', () => {
    expect(hasHelpDoc('inventory-ansible')).to.equal(true);
    expect(helpDocKeys()).to.include('inventory-ansible');
  });

  it('is not there for a key without a document', async () => {
    expect(hasHelpDoc('inventory-terraform')).to.equal(false);
    expect(hasHelpDoc(null)).to.equal(false);
    expect(hasHelpDoc('toString')).to.equal(false);

    let error = null;
    try {
      await loadHelpDoc('inventory-terraform');
    } catch (err) {
      error = err;
    }
    expect(error).to.be.an('error');
  });

  it('has the title of the dialog and says what the dialog is for', async () => {
    const doc = await loadHelpDoc('inventory-ansible');

    expect(doc.key).to.equal('inventory-ansible');
    expect(doc.title).to.equal('Edit Ansible Inventory');
    expect(doc.intro.filter((block) => block.type === 'p')).to.not.have.length(0);
  });

  it('has one section for every field of the form, in the order of the form', async () => {
    const doc = await loadHelpDoc('inventory-ansible');
    const fields = formInputs(formSource);

    expect(fields).to.deep.equal([
      'inventoryForm-name',
      'inventoryForm-runnerTag',
      'inventoryForm-sshKeyId',
      'inventoryForm-becomeKeyId',
      'inventoryForm-type',
      'inventoryForm-inventoryPath',
      'inventoryForm-repositoryId',
      'inventoryForm-inventoryCode',
    ]);
    expect(doc.sections.map((section) => section.field)).to.deep.equal(fields);
  });

  it('titles the sections as the form labels its fields', async () => {
    const doc = await loadHelpDoc('inventory-ansible');

    expect(doc.sections.map((section) => section.title)).to.deep.equal([
      'Name',
      'Runner tag',
      'User Credentials',
      'Sudo Credentials',
      'Type',
      'Path to Inventory file',
      'Repository',
      'Inventory',
    ]);
    doc.sections.forEach((section) => {
      expect(section.blocks, section.title).to.not.have.length(0);
    });
  });

  it('is opened by the dialog of Ansible inventories only', () => {
    expect(viewSource).to.include(':help-doc="isAnsibleInventory ? \'inventory-ansible\' : null"');
    expect(viewSource).to.match(/isAnsibleInventory\(\) \{\s+return this\.itemApp === 'ansible';/);
  });
});
