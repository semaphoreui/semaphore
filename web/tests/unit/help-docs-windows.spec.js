import { expect } from 'chai';
import { HelpDocError } from '../../vite/help/parse.mjs';
import helpDocs from '../../vite/help/plugin.mjs';

// The help plugin as a Windows build runs it: `node:path` is `path.win32`, the root of the
// frontend comes from `fileURLToPath` (backslashes, a trailing one), while Vite hands the plugin
// module ids with forward slashes. The files live in a map, as no Windows disk is at hand.
const { files } = vi.hoisted(() => ({ files: new Map() }));

vi.mock('node:path', async () => {
  const { win32 } = await vi.importActual('node:path');
  return { default: win32, ...win32 };
});

vi.mock('node:fs', async () => {
  const { win32 } = await vi.importActual('node:path');
  // Windows paths: either separator, any letter case.
  const key = (file) => win32.resolve(file).toLowerCase();
  const fs = {
    existsSync: (file) => files.has(key(file)),
    readFileSync: (file) => {
      if (!files.has(key(file))) throw new Error(`ENOENT: ${file}`);
      return files.get(key(file)).text;
    },
    readdirSync: (dir) => [...files.values()]
      .filter((f) => key(win32.dirname(f.path)) === key(dir))
      .map((f) => win32.basename(f.path)),
  };
  return { default: fs, ...fs };
});

const ROOT = 'C:\\repo\\web\\';

const SOURCE = [
  '---',
  'form: src/components/DemoForm.vue',
  '---',
  '',
  '# Edit Demo',
  '',
  'What the dialog is for.',
  '',
  '## Name {#demoForm-name}',
  '',
  'Text.',
  '',
].join('\n');

function write(file, text) {
  const full = `${ROOT}${file.split('/').join('\\')}`;
  files.set(full.toLowerCase(), { path: full, text });
}

// The id Vite gives a module: the absolute path with forward slashes.
const viteId = (file) => `C:/repo/web/${file}`;

function transform(file, code) {
  return helpDocs(ROOT).transform.call({ addWatchFile: () => {} }, code, viteId(file));
}

const parse = (result) => JSON.parse(result.code.replace(/^export default /, '').replace(/;\n$/, ''));

describe('help documents in a Windows build', () => {
  beforeEach(() => {
    files.clear();
    write('src/components/DemoForm.vue', '<template><input data-testid="demoForm-name"></template>');
    write('src/lang/en.js', 'export default {};');
    write('src/lang/de.js', 'export default {};');
  });

  it('compiles a document Vite names with forward slashes', () => {
    write('src/help/dialogs/demo.md', SOURCE);

    const doc = parse(transform('src/help/dialogs/demo.md', SOURCE));

    expect(doc.key).to.equal('demo');
    expect(doc.sections.map((section) => section.field)).to.deep.equal(['demoForm-name']);
  });

  it('compiles the document of a page', () => {
    write('src/help/pages/demo-page.md', SOURCE);

    expect(parse(transform('src/help/pages/demo-page.md', SOURCE)).key).to.equal('demo-page');
  });

  it('compiles a translation', () => {
    write('src/help/dialogs/demo.md', SOURCE);
    const text = SOURCE.replace('# Edit Demo', '# Demo bearbeiten');
    write('src/help/dialogs/demo.de.md', text);

    const doc = parse(transform('src/help/dialogs/demo.de.md', text));

    expect(doc.locale).to.equal('de');
    expect(doc.title).to.equal('Demo bearbeiten');
  });

  it('still fails on two documents of one name, with the path of the other in slashes', () => {
    write('src/help/dialogs/demo.md', SOURCE);
    write('src/help/pages/demo.md', SOURCE);

    expect(() => transform('src/help/pages/demo.md', SOURCE)).to.throw(
      HelpDocError,
      /^src\/help\/pages\/demo\.md:1: another document has the same name: src\/help\/dialogs\/demo\.md$/,
    );
  });
});
