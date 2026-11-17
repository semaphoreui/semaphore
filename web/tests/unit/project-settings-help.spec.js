import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect } from 'chai';
import { hasHelpDoc, helpDocKeys, loadHelpDoc } from '@/help';

const SRC = path.join(path.dirname(fileURLToPath(import.meta.url)), '../../src');

function source(file) {
  return fs.readFileSync(path.join(SRC, file), 'utf8');
}

const formSource = source('components/ProjectForm.vue');
const pageSource = source('views/project/Settings.vue');

const FIELD_TAG = /<(v-text-field|v-autocomplete|v-select|v-checkbox|v-switch|v-textarea)\b/g;
const BUTTON_TAG = /<v-btn\b/g;

// The `data-testid` of every component of a kind, in the order of the template.
function testIds(text, tags) {
  const template = text.slice(0, text.indexOf('<style') > 0 ? text.indexOf('<style') : text.indexOf('<script'));
  return [...template.matchAll(tags)].map((tag) => {
    // The end of the opening tag: the first `>` outside a quoted attribute value (a rule such as
    // `(v) => !!v` has one inside).
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
    return testId ? testId[1] : `${tag[0]} without data-testid`;
  });
}

describe('help of the project Settings page', () => {
  it('is a document of the registry, next to the documents of dialogs', () => {
    expect(hasHelpDoc('project-settings')).to.equal(true);
    expect(helpDocKeys()).to.include.members(['project-settings', 'inventory-ansible']);
  });

  it('has a title and says what the page is for', async () => {
    const doc = await loadHelpDoc('project-settings');

    expect(doc.key).to.equal('project-settings');
    expect(doc.title).to.equal('Project Settings');
    expect(doc.intro.filter((block) => block.type === 'p')).to.not.have.length(0);
  });

  it('has a section for every field of the form and every action of the page, in their order', async () => {
    const doc = await loadHelpDoc('project-settings');

    // The Demo switch is a part of the New Project dialog only; Save needs no section, nor do the
    // button of the side menu and the help button itself.
    const fields = testIds(formSource, FIELD_TAG).filter((id) => id !== 'newProject-demo');
    const actions = testIds(pageSource, BUTTON_TAG).filter((id) => id !== 'settings-save');

    expect(fields).to.deep.equal([
      'newProject-name',
      'newProject-maxParallelTasks',
      'newProject-tg',
      'newProject-alert',
    ]);
    expect(actions).to.deep.equal([
      'settings-testAlerts',
      'settings-exportProject',
      'settings-clearCache',
      'settings-deleteProject',
    ]);
    expect(doc.sections.map((section) => section.field)).to.deep.equal([...fields, ...actions]);
  });

  it('titles the sections as the page labels its fields and buttons', async () => {
    const doc = await loadHelpDoc('project-settings');

    expect(doc.sections.map((section) => section.title)).to.deep.equal([
      'Project Name',
      'Max number of parallel tasks',
      'Telegram Chat ID',
      'Allow alerts for this project',
      'Test Alerts',
      'Backup Project',
      'Clear cache',
      'Delete Project',
    ]);
    doc.sections.forEach((section) => {
      expect(section.blocks, section.title).to.not.have.length(0);
    });
  });

  it('is opened by the Settings page', () => {
    expect(pageSource).to.include('<HelpPage help-doc="project-settings"');
    expect(pageSource).to.include('data-testid="settings-docs"');
  });
});
