import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { expect } from 'chai';
import { HelpDocError, parseHelpDoc } from '../../vite/help/parse.mjs';
import {
  checkFields, findDocRefs, formFields, isDocKey, splitDocName, toModule,
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
        forms: ['src/components/DemoForm.vue'],
        skip: [],
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

    it('takes the fields it skips, separated by commas', () => {
      const source = docWith('## Name {#demoForm-name}\n\nText.')
        .replace('form: src/components/DemoForm.vue', 'form: src/components/DemoForm.vue\nskip: demoForm-demo, demoForm-x');
      expect(parseHelpDoc(source, 'demo.md').skip).to.deep.equal(['demoForm-demo', 'demoForm-x']);
    });

    it('takes several forms, separated by commas', () => {
      const source = docWith('## Name {#demoForm-name}\n\nText.')
        .replace('form: src/components/DemoForm.vue', 'form: src/views/Demo.vue,  src/components/DemoForm.vue');
      expect(parseHelpDoc(source, 'demo.md').forms)
        .to.deep.equal(['src/views/Demo.vue', 'src/components/DemoForm.vue']);
    });

    describe('joins wrapped lines', () => {
      const tests = [
        { name: 'with a space in English', lines: ['one', 'two'], text: 'one two' },
        { name: 'without a space between Chinese characters', lines: ['库存告诉', '主机'], text: '库存告诉主机' },
        { name: 'without a space after full-width punctuation', lines: ['运行，', '该模板'], text: '运行，该模板' },
        { name: 'without a space between Latin and Chinese', lines: ['使用 Semaphore', '的任务'], text: '使用 Semaphore的任务' },
        { name: 'without a space in Japanese', lines: ['インベントリは', 'ホスト'], text: 'インベントリはホスト' },
        { name: 'with a space in Korean', lines: ['인벤토리는', '호스트를'], text: '인벤토리는 호스트를' },
      ];

      tests.forEach(({ name, lines, text }) => {
        it(name, () => {
          const doc = parseHelpDoc(docWith(`## Name {#a}\n\n${lines.join('\n')}\n\n- ${lines.join('\n  ')}`), 'demo.md');
          expect(doc.sections[0].blocks[0].inline).to.deep.equal([{ type: 'text', text }]);
          expect(doc.sections[0].blocks[1].items[0]).to.deep.equal([{ type: 'text', text }]);
        });
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
      { name: 'bad path among several forms', source: '---\nform: src/a/A.vue, /etc/passwd\n---\n', error: /"form" must be a path .* got "\/etc\/passwd"/ },
      { name: 'empty path among several forms', source: '---\nform: src/a/A.vue,\n---\n', error: /"form" must be a path/ },
      { name: 'form named twice', source: '---\nform: src/a/A.vue, src/a/A.vue\n---\n', error: /names a file twice/ },
      { name: 'skip of something that is not a test id', source: '---\nform: src/a/A.vue\nskip: a b\n---\n', error: /"skip" lists the data-testid of fields, got "a b"/ },
      { name: 'skip naming a field twice', source: '---\nform: src/a/A.vue\nskip: a, a\n---\n', error: /"skip" names a field twice/ },
      { name: 'a field skipped and described', source: docWith('## Name {#a}\n\nText.').replace('---\n\n#', 'skip: a\n---\n\n#'), error: /the field "a" has a section and is in "skip" as well/ },
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
      expect(() => checkFields(doc, [form], 'demo.md')).to.not.throw();
    });

    it('fails on a section of a field the form does not have', () => {
      expect(() => checkFields(doc, ['<a data-testid="demoForm-name"></a>'], 'demo.md')).to.throw(
        HelpDocError,
        /^demo\.md:13: the section "Gone" is anchored to "demoForm-gone", but src\/components\/DemoForm\.vue has no data-testid="demoForm-gone"$/,
      );
    });

    it('fails on a skipped field the form does not have', () => {
      const skipping = { ...doc, skip: ['demoForm-nope'] };
      const form = '<a data-testid="demoForm-name"></a><a data-testid="demoForm-gone"></a>';
      expect(() => checkFields(skipping, [form], 'demo.md')).to.throw(
        HelpDocError,
        /^demo\.md:1: "skip" names "demoForm-nope", but src\/components\/DemoForm\.vue has no data-testid="demoForm-nope"$/,
      );
    });

    describe('with several forms', () => {
      const pageDoc = { ...doc, forms: ['src/views/Demo.vue', 'src/components/DemoForm.vue'] };

      it('passes when every field is in one of the forms', () => {
        const sources = ['<a data-testid="demoForm-gone"></a>', '<a data-testid="demoForm-name"></a>'];
        expect(() => checkFields(pageDoc, sources, 'demo.md')).to.not.throw();
      });

      it('fails on a field none of the forms has, and names them all', () => {
        const sources = ['<a data-testid="other"></a>', '<a data-testid="demoForm-name"></a>'];
        expect(() => checkFields(pageDoc, sources, 'demo.md')).to.throw(
          HelpDocError,
          /but none of src\/views\/Demo\.vue, src\/components\/DemoForm\.vue has data-testid="demoForm-gone"$/,
        );
      });
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

  describe('splitDocName', () => {
    const tests = [
      ['inventory-ansible', { key: 'inventory-ansible', locale: null }],
      ['inventory-ansible.de', { key: 'inventory-ansible', locale: 'de' }],
      ['inventory-ansible.pt_br', { key: 'inventory-ansible', locale: 'pt_br' }],
      ['inventory-ansible.DE', null],
      ['inventory-ansible.de.ru', null],
      ['inventory-ansible.deu', null],
      ['Inventory', null],
    ];
    tests.forEach(([name, parts]) => {
      it(`${name} -> ${JSON.stringify(parts)}`, () => {
        expect(splitDocName(name)).to.deep.equal(parts);
      });
    });
  });

  describe('toModule', () => {
    it('drops what only the build needs', () => {
      const doc = parseHelpDoc(docWith('## Name {#demoForm-name}\n\nText.'), 'demo.md');
      expect(toModule(doc, 'demo')).to.deep.equal({
        key: 'demo',
        locale: 'en',
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
      expect(doc).to.not.have.property('forms');
    });

    it('compiles the document of a page, anchored to the page and to its form', () => {
      write('src/views/Demo.vue', '<template><button data-testid="demo-save"></button></template>');
      const source = docWith('## Name {#demoForm-name}\n\nText.\n\n## Save {#demo-save}\n\nText.')
        .replace('form: src/components/DemoForm.vue', 'form: src/views/Demo.vue, src/components/DemoForm.vue');
      const file = write('src/help/pages/demo-page.md', source);

      const { result, watched } = transform(file, source);

      expect(watched).to.deep.equal([
        path.join(root, 'src/views/Demo.vue'),
        path.join(root, 'src/components/DemoForm.vue'),
      ]);
      const doc = JSON.parse(result.code.replace(/^export default /, '').replace(/;\n$/, ''));
      expect(doc.key).to.equal('demo-page');
      expect(doc.sections.map((section) => section.field)).to.deep.equal(['demoForm-name', 'demo-save']);
    });

    it('fails on two documents of one name', () => {
      const source = docWith('## Name {#demoForm-name}\n\nText.');
      write('src/help/dialogs/demo.md', source);
      const file = write('src/help/pages/demo.md', source);

      expect(() => transform(file, source)).to.throw(
        HelpDocError,
        /^src\/help\/pages\/demo\.md:1: another document has the same name: src\/help\/dialogs\/demo\.md$/,
      );
    });

    it('fails when one of several forms does not exist', () => {
      const source = docWith('## Name {#demoForm-name}\n\nText.')
        .replace('form: src/components/DemoForm.vue', 'form: src/components/DemoForm.vue, src/views/No.vue');
      const file = write('src/help/pages/demo-page.md', source);

      expect(() => compileHelpDoc(root, file, source))
        .to.throw(HelpDocError, /the form "src\/views\/No\.vue" does not exist/);
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
        /^src\/views\/Demo\.vue:2: help document "missing-doc" does not exist \(src\/help\/dialogs\/missing-doc\.md or src\/help\/pages\/missing-doc\.md\)$/,
      );
    });

    it('passes a component that names an existing document', () => {
      write('src/help/dialogs/demo.md', docWith('## Name {#demoForm-name}\n\nText.'));
      const code = '<template><EditDialog :help-doc="open ? \'demo\' : null" /></template>';
      const file = write('src/views/Demo.vue', code);

      expect(() => checkDocRefs(root, file, code)).to.not.throw();
      expect(transform(file, code).result).to.equal(null);
    });

    it('passes a component that names the document of a page', () => {
      write('src/help/pages/demo-page.md', docWith('## Name {#demoForm-name}\n\nText.'));
      const code = '<template><HelpPage help-doc="demo-page"></HelpPage></template>';
      const file = write('src/views/Demo.vue', code);

      expect(() => checkDocRefs(root, file, code)).to.not.throw();
    });

    describe('translations', () => {
      const SOURCE = docWith([
        '## Name {#demoForm-name}',
        '',
        'See [the guide](https://example.com/guide).',
        '',
        '```ini',
        '[web]',
        '```',
        '',
        '## Type {#demoForm-type}',
        '',
        'Text.',
      ].join('\n'));
      const TRANSLATION = SOURCE
        .replace('# Edit Demo', '# Demo bearbeiten')
        .replace('What the dialog is for.', 'Wofür der Dialog da ist.')
        .replace('## Name {#demoForm-name}', '## Name {#demoForm-name}')
        .replace('See [the guide]', 'Siehe [die Anleitung]')
        .replace('## Type {#demoForm-type}', '## Typ {#demoForm-type}')
        .replace('Text.', 'Text auf Deutsch.');

      beforeEach(() => {
        const form = '<input data-testid="demoForm-name"><input data-testid="demoForm-type">';
        write('src/components/DemoForm.vue', `<template>${form}</template>`);
        write('src/lang/en.js', 'export default {};');
        write('src/lang/de.js', 'export default {};');
        write('src/lang/pt_br.js', 'export default {};');
        write('src/lang/index.js', '');
        write('src/help/dialogs/demo.md', SOURCE);
      });

      function translate(name, text = TRANSLATION) {
        const file = write(`src/help/dialogs/${name}`, text);
        return transform(file, text);
      }

      it('compiles a translation, with its language, and watches its source', () => {
        const { result, watched } = translate('demo.de.md');

        const doc = JSON.parse(result.code.replace(/^export default /, '').replace(/;\n$/, ''));
        expect(doc.key).to.equal('demo');
        expect(doc.locale).to.equal('de');
        expect(doc.title).to.equal('Demo bearbeiten');
        expect(doc.sections.map((s) => s.title)).to.deep.equal(['Name', 'Typ']);
        expect(watched).to.deep.equal([
          path.join(root, 'src/components/DemoForm.vue'),
          path.join(root, 'src/help/dialogs/demo.md'),
        ]);
      });

      it('takes a language with a region', () => {
        const { result } = translate('demo.pt_br.md');
        expect(result.code).to.include('"locale":"pt_br"');
      });

      it('gives the English source the locale en', () => {
        const { result } = transform(path.join(root, 'src/help/dialogs/demo.md'), SOURCE);
        expect(result.code).to.include('"locale":"en"');
      });

      const rejected = [
        { name: 'a language the interface does not have', file: 'demo.xx.md', error: /"xx" is not a language of the interface \(src\/lang\): de, en, pt_br$/ },
        { name: 'an English translation', file: 'demo.en.md', error: /the English document is demo\.md, without a locale/ },
        { name: 'a translation without a source', file: 'gone.de.md', error: /a translation of a document that does not exist: src\/help\/dialogs\/gone\.md/ },
        { name: 'a name with an upper-case language', file: 'demo.DE.md', error: /with "\.<locale>" for a translation, got "demo\.DE\.md"/ },
        {
          name: 'a section missing',
          file: 'demo.de.md',
          text: TRANSLATION.slice(0, TRANSLATION.indexOf('## Typ')),
          error: /^src\/help\/dialogs\/demo\.de\.md:1: the section \{#demoForm-type\} of src\/help\/dialogs\/demo\.md is missing$/,
        },
        {
          name: 'a section more',
          file: 'demo.de.md',
          text: `${TRANSLATION}\n## Extra {#demoForm-name2}\n\nText.\n`,
          error: /the section \{#demoForm-name2\} is not in src\/help\/dialogs\/demo\.md/,
        },
        {
          name: 'sections in another order',
          file: 'demo.de.md',
          text: TRANSLATION.replace('{#demoForm-name}', '{#tmp}').replace('{#demoForm-type}', '{#demoForm-name}').replace('{#tmp}', '{#demoForm-type}'),
          error: /:9: the section \{#demoForm-type\} stands where .* has \{#demoForm-name\}/,
        },
        {
          name: 'translated code',
          file: 'demo.de.md',
          text: TRANSLATION.replace('[web]', '[Web-Server]'),
          error: /the code blocks must be the same as in src\/help\/dialogs\/demo\.md/,
        },
        {
          name: 'another link',
          file: 'demo.de.md',
          text: TRANSLATION.replace('https://example.com/guide', 'https://example.com/de/guide'),
          error: /the links must be the same as in .*: https:\/\/example\.com\/guide; found https:\/\/example\.com\/de\/guide$/,
        },
        {
          name: 'another form',
          file: 'demo.de.md',
          text: TRANSLATION.replace('form: src/components/DemoForm.vue', 'form: src/components/DemoForm.vue, src/views/Demo.vue'),
          error: /"form" must be the same as in src\/help\/dialogs\/demo\.md: src\/components\/DemoForm\.vue$/,
        },
        {
          name: 'another skip',
          file: 'demo.de.md',
          text: TRANSLATION.replace('form: src/components/DemoForm.vue', 'form: src/components/DemoForm.vue\nskip: demoForm-type2'),
          error: /"skip" must be the same as in src\/help\/dialogs\/demo\.md: none$/,
        },
        {
          name: 'HTML in a translation',
          file: 'demo.de.md',
          text: TRANSLATION.replace('Text auf Deutsch.', 'Text <b>auf</b> Deutsch.'),
          error: /raw HTML is not allowed/,
        },
      ];

      rejected.forEach(({
        name, file, text, error,
      }) => {
        it(`rejects ${name}`, () => {
          expect(() => translate(file, text)).to.throw(HelpDocError, error);
        });
      });

      it('fails a translation when its source is broken', () => {
        write('src/help/dialogs/demo.md', SOURCE.replace('{#demoForm-type}', '{#demoForm-gone}'));
        expect(() => translate('demo.de.md')).to.throw(HelpDocError, /^src\/help\/dialogs\/demo\.md:\d+: .*demoForm-gone/);
      });

      it('lets a component name the key, not a translation', () => {
        translate('demo.de.md');
        const ok = '<template><EditDialog help-doc="demo" /></template>';
        const bad = '<template><EditDialog help-doc="demo.de" /></template>';
        const file = write('src/views/Demo.vue', ok);

        expect(() => checkDocRefs(root, file, ok)).to.not.throw();
        expect(() => checkDocRefs(root, file, bad)).to.throw(HelpDocError, /help document "demo\.de" does not exist/);
      });
    });

    it('does not check the blocks the Vue plugin splits a component into', () => {
      const code = '<EditDialog help-doc="missing-doc" />';
      const file = write('src/views/Demo.vue', code);
      expect(transform(`${file}?vue&type=template&lang.js`, code).result).to.equal(null);
    });
  });
});
