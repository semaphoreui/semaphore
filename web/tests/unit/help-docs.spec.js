import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { expect } from 'chai';
import { HelpDocError, parseHelpDoc } from '../../vite/help/parse.mjs';
import {
  checkFields, findDocRefs, formFields, isDocKey, toModule,
} from '../../vite/help/validate.mjs';
import helpDocs, { checkDocRefs, compileHelpDoc } from '../../vite/help/plugin.mjs';

const FRONT = '---\nform: src/components/DemoForm.vue\n---\n';

// A document with the given body under a valid front matter, title and intro.
function docWith(body) {
  return `${FRONT}\n# Edit Demo\n\nWhat the dialog is for.\n\n${body}\n`;
}

describe('help documents', () => {
  describe('parseHelpDoc', () => {
    it('turns a document into a tree', () => {
      const doc = parseHelpDoc([
        '---',
        'form: src/components/DemoForm.vue',
        '---',
        '',
        '# Edit Demo',
        '',
        'What the dialog',
        'is for.',
        '',
        '## Name {#demoForm-name}',
        '',
        'A `code`, a **bold** and a [link](https://example.com/a?b=1) in the [web] group.',
        '',
        '- first',
        '  continued',
        '- second with `code`',
        '',
        '```ini',
        '[web]',
        '<b>10.0.0.1</b>',
        '```',
        '',
        '## Type {#demoForm-type}',
        '',
        'Shown when the type is File.',
      ].join('\n'), 'demo.md');

      expect(doc).to.deep.equal({
        form: 'src/components/DemoForm.vue',
        title: 'Edit Demo',
        intro: [{ type: 'p', inline: [{ type: 'text', text: 'What the dialog is for.' }] }],
        sections: [{
          field: 'demoForm-name',
          title: 'Name',
          line: 10,
          blocks: [{
            type: 'p',
            inline: [
              { type: 'text', text: 'A ' },
              { type: 'code', text: 'code' },
              { type: 'text', text: ', a ' },
              { type: 'strong', text: 'bold' },
              { type: 'text', text: ' and a ' },
              { type: 'link', text: 'link', href: 'https://example.com/a?b=1' },
              { type: 'text', text: ' in the [web] group.' },
            ],
          }, {
            type: 'ul',
            items: [
              [{ type: 'text', text: 'first continued' }],
              [{ type: 'text', text: 'second with ' }, { type: 'code', text: 'code' }],
            ],
          }, {
            type: 'code',
            text: '[web]\n<b>10.0.0.1</b>',
          }],
        }, {
          field: 'demoForm-type',
          title: 'Type',
          line: 23,
          blocks: [{ type: 'p', inline: [{ type: 'text', text: 'Shown when the type is File.' }] }],
        }],
      });
    });

    it('accepts Windows line endings', () => {
      const doc = parseHelpDoc(docWith('## Name {#demoForm-name}\n\nText.').replace(/\n/g, '\r\n'), 'demo.md');
      expect(doc.sections.map((section) => section.field)).to.deep.equal(['demoForm-name']);
    });

    it('keeps HTML inside code as text', () => {
      const doc = parseHelpDoc(docWith('## Name {#demoForm-name}\n\nUse `<script>` never.'), 'demo.md');
      expect(doc.sections[0].blocks[0].inline[1]).to.deep.equal({ type: 'code', text: '<script>' });
    });

    const rejected = [
      { name: 'no front matter', source: '# Edit Demo\n\nText.\n', error: /demo\.md:1: .*front matter/ },
      { name: 'front matter not closed', source: '---\nform: src/a/B.vue', error: /not closed/ },
      { name: 'text inside the front matter', source: '---\nform: src/a/B.vue\n\n# T\n', error: /demo\.md:3: a front matter line/ },
      { name: 'no form', source: '---\n---\n\n# T\n\nText.\n', error: /must name the form/ },
      { name: 'unknown front matter key', source: '---\ntitle: X\n---\n', error: /demo\.md:2: unknown front matter key "title"/ },
      { name: 'form outside src', source: '---\nform: ../secret/X.vue\n---\n', error: /"form" must be a path/ },
      { name: 'form with a parent segment', source: '---\nform: src/../X.vue\n---\n', error: /"form" must be a path/ },
      { name: 'text before the title', source: `${FRONT}\nText.\n\n# T\n`, error: /demo\.md:5: .*must start with its `# Title`/ },
      { name: 'no title', source: `${FRONT}\n`, error: /has no `# Title`/ },
      { name: 'no intro paragraph', source: `${FRONT}\n# T\n\n## Name {#a}\n\nText.\n`, error: /needs a paragraph/ },
      { name: 'second title', source: docWith('# Another'), error: /one `# Title`/ },
      { name: 'section without an anchor', source: docWith('## Name\n\nText.'), error: /needs the anchor/ },
      { name: 'anchor used twice', source: docWith('## A {#x}\n\nText.\n\n## B {#x}\n\nText.'), error: /"x" is used by two sections/ },
      { name: 'empty section', source: docWith('## A {#x}\n\n## B {#y}\n\nText.'), error: /section "A" is empty/ },
      { name: 'deep heading', source: docWith('### Deep'), error: /deeper than ##/ },
      { name: 'heading without a space', source: docWith('#Name'), error: /a heading is/ },
      { name: 'script element', source: docWith('<script>alert(1)</script>'), error: /raw HTML is not allowed/ },
      { name: 'element with a handler', source: docWith('Text <img src=x onerror=alert(1)> here.'), error: /raw HTML is not allowed/ },
      { name: 'HTML in a list item', source: docWith('- one <b>two</b>'), error: /raw HTML is not allowed/ },
      { name: 'javascript link', source: docWith('[x](javascript:alert(1))'), error: /must point to an https:\/\/ address/ },
      { name: 'http link', source: docWith('[x](http://example.com)'), error: /must point to an https:\/\/ address/ },
      { name: 'relative link', source: docWith('[x](/project/1)'), error: /must point to an https:\/\/ address/ },
      { name: 'link with a space', source: docWith('[x](https://example.com/a b)'), error: /must point to an https:\/\/ address/ },
      { name: 'link with a quote', source: docWith('[x](https://example.com/"onmouseover="x)'), error: /must point to an https:\/\/ address/ },
      { name: 'formatted link text', source: docWith('[`x`](https://example.com)'), error: /text of a link takes plain text/ },
      { name: 'image', source: docWith('![alt](https://example.com/a.png)'), error: /images are not supported/ },
      { name: 'table', source: docWith('| a | b |\n|---|---|'), error: /tables are not supported/ },
      { name: 'block quote', source: docWith('> quote'), error: /block quotes are not supported/ },
      { name: 'ordered list', source: docWith('1. one'), error: /only `-` lists/ },
      { name: 'star list', source: docWith('* one'), error: /only `-` lists/ },
      { name: 'nested list', source: docWith('- one\n  - two'), error: /nested lists/ },
      { name: 'indented text', source: docWith('    code'), error: /indented text/ },
      { name: 'rule', source: docWith('---'), error: /rules and underlined headings/ },
      { name: 'single star emphasis', source: docWith('an *emphasis* here'), error: /single \*/ },
      { name: 'unclosed code', source: docWith('an `unclosed code'), error: /`code` is not closed/ },
      { name: 'unclosed bold', source: docWith('an **unclosed bold'), error: /\*\*bold\*\* is not closed/ },
      { name: 'formatted bold', source: docWith('a **`code` in bold**'), error: /\*\*bold\*\* takes plain text/ },
      { name: 'unclosed code block', source: docWith('```\ncode'), error: /code block is not closed/ },
      { name: 'code fence with attributes', source: docWith('```js {1}\ncode\n```'), error: /language name only/ },
    ];

    rejected.forEach(({ name, source, error }) => {
      it(`rejects: ${name}`, () => {
        expect(() => parseHelpDoc(source, 'demo.md')).to.throw(HelpDocError, error);
      });
    });

    it('reports the line of the offending text', () => {
      const source = docWith('## Name {#a}\n\nFine.\n\nBroken <b>here</b>.');
      expect(() => parseHelpDoc(source, 'dialogs/demo.md'))
        .to.throw(HelpDocError, /^dialogs\/demo\.md:13: raw HTML/);
    });
  });

  describe('formFields', () => {
    it('collects literal data-testid values only', () => {
      const fields = formFields([
        '<v-text-field data-testid="demoForm-name" />',
        // eslint-disable-next-line no-template-curly-in-string
        '<div :data-testid="`demoForm-${id}`"></div>',
        '<RichEditor test-id="demoForm-editor" />',
        '<v-select\n  data-testid="demoForm-type"\n/>',
      ].join('\n'));
      expect([...fields]).to.deep.equal(['demoForm-name', 'demoForm-type']);
    });
  });

  describe('checkFields', () => {
    const doc = parseHelpDoc(docWith('## Name {#demoForm-name}\n\nText.\n\n## Gone {#demoForm-gone}\n\nText.'), 'demo.md');

    it('passes when the form has every field', () => {
      const form = '<a data-testid="demoForm-name"></a><a data-testid="demoForm-gone"></a>';
      expect(() => checkFields(doc, form, 'demo.md')).to.not.throw();
    });

    it('fails on a section of a field the form does not have', () => {
      expect(() => checkFields(doc, '<a data-testid="demoForm-name"></a>', 'demo.md')).to.throw(
        HelpDocError,
        /^demo\.md:13: the section "Gone" is anchored to "demoForm-gone", but src\/components\/DemoForm\.vue has no data-testid="demoForm-gone"$/,
      );
    });
  });

  describe('findDocRefs', () => {
    const tests = [
      { name: 'static attribute', source: '<EditDialog help-doc="inventory-ansible" />', keys: ['inventory-ansible'] },
      {
        name: 'bound expression with literals',
        source: '<EditDialog :help-doc="app === \'ansible\' ? \'inventory-ansible\' : null" />',
        keys: ['ansible', 'inventory-ansible'],
      },
      { name: 'v-bind form', source: '<EditDialog v-bind:help-doc="\'key-one\'" />', keys: ['key-one'] },
      { name: 'bound variable', source: '<EditDialog :help-doc="docKey" />', keys: [] },
      { name: 'empty attribute', source: '<EditDialog help-doc="" />', keys: [] },
      { name: 'another attribute', source: '<EditDialog no-help-doc="x" data-help-doc="y" />', keys: [] },
      { name: 'no attribute', source: '<EditDialog title="x" />', keys: [] },
    ];

    tests.forEach(({ name, source, keys }) => {
      it(name, () => {
        expect(findDocRefs(source).map((ref) => ref.key)).to.deep.equal(keys);
      });
    });

    it('reports the line of the attribute', () => {
      expect(findDocRefs('<template>\n  <EditDialog\n    help-doc="a"\n  />\n</template>'))
        .to.deep.equal([{ key: 'a', line: 3 }]);
    });
  });

  describe('isDocKey', () => {
    const tests = [
      ['inventory-ansible', true], ['key', true], ['a1-b2', true],
      ['Inventory', false], ['a_b', false], ['-a', false], ['a--b', false], ['../a', false], ['', false],
    ];
    tests.forEach(([key, valid]) => {
      it(`${JSON.stringify(key)} -> ${valid}`, () => {
        expect(isDocKey(key)).to.equal(valid);
      });
    });
  });

  describe('toModule', () => {
    it('drops what only the build needs', () => {
      const doc = parseHelpDoc(docWith('## Name {#demoForm-name}\n\nText.'), 'demo.md');
      expect(toModule(doc, 'demo')).to.deep.equal({
        key: 'demo',
        title: 'Edit Demo',
        intro: [{ type: 'p', inline: [{ type: 'text', text: 'What the dialog is for.' }] }],
        sections: [{
          field: 'demoForm-name',
          title: 'Name',
          blocks: [{ type: 'p', inline: [{ type: 'text', text: 'Text.' }] }],
        }],
      });
    });
  });

  describe('plugin', () => {
    let root;

    function write(file, text) {
      const full = path.join(root, file);
      fs.mkdirSync(path.dirname(full), { recursive: true });
      fs.writeFileSync(full, text);
      return full;
    }

    // The hooks of a plugin are called with the context of the bundler.
    function transform(file, code) {
      const watched = [];
      const context = { addWatchFile: (f) => watched.push(f) };
      const result = helpDocs(root).transform.call(context, code, file);
      return { result, watched };
    }

    beforeEach(() => {
      root = fs.mkdtempSync(path.join(os.tmpdir(), 'help-docs-'));
      write('src/components/DemoForm.vue', '<template><input data-testid="demoForm-name"></template>');
    });

    afterEach(() => {
      fs.rmSync(root, { recursive: true, force: true });
    });

    it('compiles a document into a module and watches its form', () => {
      const source = docWith('## Name {#demoForm-name}\n\nText.');
      const file = write('src/help/dialogs/demo.md', source);

      const { result, watched } = transform(file, source);

      expect(watched).to.deep.equal([path.join(root, 'src/components/DemoForm.vue')]);
      expect(result.code).to.match(/^export default \{/);
      const doc = JSON.parse(result.code.replace(/^export default /, '').replace(/;\n$/, ''));
      expect(doc.key).to.equal('demo');
      expect(doc.sections.map((section) => section.field)).to.deep.equal(['demoForm-name']);
      expect(doc).to.not.have.property('form');
    });

    it('fails on a section of a field the form does not have', () => {
      const source = docWith('## Gone {#demoForm-gone}\n\nText.');
      const file = write('src/help/dialogs/demo.md', source);

      expect(() => transform(file, source)).to.throw(
        HelpDocError,
        /^src\/help\/dialogs\/demo\.md:9: .*has no data-testid="demoForm-gone"$/,
      );
    });

    it('fails on a form that does not exist', () => {
      const source = docWith('## Name {#demoForm-name}\n\nText.').replace('DemoForm', 'NoForm');
      const file = write('src/help/dialogs/demo.md', source);

      expect(() => compileHelpDoc(root, file, source))
        .to.throw(HelpDocError, /the form "src\/components\/NoForm\.vue" does not exist/);
    });

    it('fails on a document name that can not be a key', () => {
      const source = docWith('## Name {#demoForm-name}\n\nText.');
      const file = write('src/help/dialogs/Demo_Doc.md', source);

      expect(() => compileHelpDoc(root, file, source)).to.throw(HelpDocError, /name of a document/);
    });

    it('leaves other Markdown files alone', () => {
      const file = write('src/assets/readme.md', '<b>anything</b>');
      expect(transform(file, '<b>anything</b>').result).to.equal(null);
    });

    it('fails on a component that names a missing document', () => {
      const code = '<template>\n  <EditDialog help-doc="missing-doc" />\n</template>';
      const file = write('src/views/Demo.vue', code);

      expect(() => transform(file, code)).to.throw(
        HelpDocError,
        /^src\/views\/Demo\.vue:2: help document "missing-doc" does not exist \(src\/help\/dialogs\/missing-doc\.md\)$/,
      );
    });

    it('passes a component that names an existing document', () => {
      write('src/help/dialogs/demo.md', docWith('## Name {#demoForm-name}\n\nText.'));
      const code = '<template><EditDialog :help-doc="open ? \'demo\' : null" /></template>';
      const file = write('src/views/Demo.vue', code);

      expect(() => checkDocRefs(root, file, code)).to.not.throw();
      expect(transform(file, code).result).to.equal(null);
    });

    it('does not check the blocks the Vue plugin splits a component into', () => {
      const code = '<EditDialog help-doc="missing-doc" />';
      const file = write('src/views/Demo.vue', code);
      expect(transform(`${file}?vue&type=template&lang.js`, code).result).to.equal(null);
    });
  });
});
