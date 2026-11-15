// Parser of the help documents (src/help/dialogs/*.md), run at build time.
//
// A document is written in a strict subset of Markdown and is turned into a tree of plain
// objects; the application draws the tree itself, so no HTML of a document ever reaches the
// browser and nothing has to be sanitised. Everything outside the subset fails the build with
// the file and the line: a construct that is silently shown as text would rot unnoticed.
//
//   ---
//   form: src/components/InventoryForm.vue
//   ---
//
//   # Title of the dialog
//
//   What the dialog is for.
//
//   ## Name {#inventoryForm-name}
//
//   Text of the section: paragraphs, `-` lists, fenced code, `code`, **bold**,
//   [links](https://example.com).
//
// The anchor of a section is the `data-testid` of the element it describes: a field of a form,
// a button of a page. `form` names the file the elements are in, or several files, separated
// by commas, when a page takes its form from another component. The optional `skip` lists, by
// `data-testid` and separated by commas, the fields of those files the document leaves without
// a section on purpose (a field another document describes, a switch for demos).

export class HelpDocError extends Error {
  constructor(file, line, message) {
    super(`${file}:${line}: ${message}`);
    this.name = 'HelpDocError';
    this.file = file;
    this.line = line;
  }
}

const FORM_PATH = /^src\/[\w./-]+\.vue$/;
const ANCHORED_HEADING = /^(.*\S)\s+\{#([A-Za-z][\w-]*)\}$/;
const LINK = /^\[([^\]]*)\]\(([^)]*)\)/;
const HTTPS_URL = /^https:\/\/[^\s"'<>`]+$/;
// Characters that start formatting: text that must stay plain may not hold them.
const FORMATTING = /[`*[<]/;

// Chinese and Japanese write no space between words: a line break next to one of their
// characters (ideographs, kana, full-width punctuation) joins the lines without a space. Korean
// separates words with spaces, so Hangul is not in the set.
const NO_SPACE_SCRIPT = /[\u3000-\u30ff\u3400-\u4dbf\u4e00-\u9fff\uf900-\ufaff\uff00-\uffef]/;

// Joins a line of wrapped text to the text before it.
function joinWrapped(text, line) {
  const next = line.trim();
  const tight = NO_SPACE_SCRIPT.test(text.slice(-1)) || NO_SPACE_SCRIPT.test(next[0]);
  return `${text}${tight ? '' : ' '}${next}`;
}

// Text of a paragraph or a list item ->
// [{ type: 'text' | 'code' | 'strong' | 'link', text, href? }].
function parseInline(text, fail) {
  const nodes = [];
  let plain = '';
  let i = 0;

  const push = (node) => {
    if (plain) {
      nodes.push({ type: 'text', text: plain });
      plain = '';
    }
    if (node) {
      nodes.push(node);
    }
  };

  while (i < text.length) {
    const ch = text[i];

    if (ch === '`') {
      const end = text.indexOf('`', i + 1);
      if (end < 0) {
        fail('`code` is not closed');
      }
      if (end === i + 1) {
        fail('empty `code`; double backticks are not supported');
      }
      push({ type: 'code', text: text.slice(i + 1, end) });
      i = end + 1;
    } else if (text.startsWith('**', i)) {
      const end = text.indexOf('**', i + 2);
      if (end < 0) {
        fail('**bold** is not closed');
      }
      const inner = text.slice(i + 2, end);
      if (!inner.trim() || FORMATTING.test(inner)) {
        fail('**bold** takes plain text only');
      }
      push({ type: 'strong', text: inner });
      i = end + 2;
    } else if (ch === '*') {
      fail('emphasis with a single * is not supported; use **bold**');
    } else if (ch === '<') {
      fail('raw HTML is not allowed; put a literal < in `code`');
    } else if (text.startsWith('![', i)) {
      fail('images are not supported');
    } else if (ch === '[') {
      const link = LINK.exec(text.slice(i));
      if (!link) {
        // A bracket that starts no link, as in "the [web] group".
        plain += ch;
        i += 1;
      } else {
        const [whole, label, href] = link;
        if (!HTTPS_URL.test(href)) {
          fail(`a link must point to an https:// address, got "${href}"`);
        }
        if (!label.trim() || FORMATTING.test(label)) {
          fail('the text of a link takes plain text only');
        }
        push({ type: 'link', text: label, href });
        i += whole.length;
      }
    } else {
      plain += ch;
      i += 1;
    }
  }

  push(null);
  return nodes;
}

// The front matter: `---`, `key: value` lines, `---`. Returns the values and the index of the
// first line after it.
function parseFrontMatter(lines, fail) {
  if (lines[0] !== '---') {
    fail(1, 'the document must start with a front matter block (---)');
  }

  const meta = {};
  let i = 1;
  while (i < lines.length && lines[i] !== '---') {
    const pair = /^([a-z_]+):\s*(\S.*)$/.exec(lines[i]);
    if (!pair) {
      fail(i + 1, 'a front matter line is `key: value`');
    }
    const [, key, value] = pair;
    if (key !== 'form' && key !== 'skip') {
      fail(i + 1, `unknown front matter key "${key}"`);
    }
    if (meta[key] != null) {
      fail(i + 1, `front matter key "${key}" is repeated`);
    }
    meta[key] = value.trim();
    i += 1;
  }

  if (i >= lines.length) {
    fail(1, 'the front matter block is not closed (---)');
  }
  if (meta.form == null) {
    fail(1, 'the front matter must name the form: `form: src/components/<Name>Form.vue`');
  }
  const forms = meta.form.split(',').map((form) => form.trim());
  forms.forEach((form) => {
    if (!FORM_PATH.test(form) || form.includes('..')) {
      fail(1, `"form" must be a path to a .vue file under src/, got "${form}"`);
    }
  });
  if (new Set(forms).size !== forms.length) {
    fail(1, '"form" names a file twice');
  }

  const skip = meta.skip == null ? [] : meta.skip.split(',').map((field) => field.trim());
  skip.forEach((field) => {
    if (!/^[A-Za-z][\w-]*$/.test(field)) {
      fail(1, `"skip" lists the data-testid of fields, got "${field}"`);
    }
  });
  if (new Set(skip).size !== skip.length) {
    fail(1, '"skip" names a field twice');
  }

  return { forms, skip, next: i + 1 };
}

/**
 * Parses a help document.
 *
 * @param {string} source text of the document
 * @param {string} file name of the document, for error messages
 * @returns {{ forms: string[], skip: string[], title: string, intro: object[],
 *   sections: { field: string, title: string, line: number, blocks: object[] }[] }}
 * @throws {HelpDocError} on anything outside the subset
 */
export function parseHelpDoc(source, file) {
  const lines = source.replace(/\r\n?/g, '\n').split('\n');
  const fail = (line, message) => {
    throw new HelpDocError(file, line, message);
  };

  const { forms, skip, next } = parseFrontMatter(lines, fail);
  const doc = {
    forms,
    skip,
    title: null,
    intro: [],
    sections: [],
  };

  // Where blocks go: the intro after the title, then the section in hand. Null before the title.
  let blocks = null;
  // The paragraph and the list being collected: { line, text } and { items: [{ line, text }] }.
  let paragraph = null;
  let list = null;

  const inline = ({ line, text }) => parseInline(text, (message) => fail(line, message));

  const flush = () => {
    if (paragraph) {
      blocks.push({ type: 'p', inline: inline(paragraph) });
      paragraph = null;
    }
    if (list) {
      blocks.push({ type: 'ul', items: list.items.map(inline) });
      list = null;
    }
  };

  const target = (n) => {
    if (blocks == null) {
      fail(n, 'the document must start with its `# Title`');
    }
    return blocks;
  };

  for (let i = next; i < lines.length; i += 1) {
    const line = lines[i];
    const n = i + 1;

    if (line.startsWith('```')) {
      flush();
      if (!/^```[\w-]*$/.test(line)) {
        fail(n, 'a code fence takes a language name only');
      }
      const body = [];
      i += 1;
      while (i < lines.length && lines[i] !== '```') {
        body.push(lines[i]);
        i += 1;
      }
      if (i >= lines.length) {
        fail(n, 'the code block is not closed (```)');
      }
      target(n).push({ type: 'code', text: body.join('\n') });
    } else if (line.trim() === '') {
      flush();
    } else if (line.startsWith('#')) {
      flush();
      const heading = /^(#+)\s+(\S.*)$/.exec(line);
      if (!heading) {
        fail(n, 'a heading is `# Title` or `## Title {#field}`');
      }
      const level = heading[1].length;
      const text = heading[2].trim();

      if (level === 1) {
        if (doc.title != null) {
          fail(n, 'a document has one `# Title`');
        }
        if (/[`*[<{]/.test(text)) {
          fail(n, 'the title takes plain text only');
        }
        doc.title = text;
        blocks = doc.intro;
      } else if (level === 2) {
        target(n);
        const anchored = ANCHORED_HEADING.exec(text);
        if (!anchored) {
          fail(n, 'a section needs the anchor of its field: `## Title {#data-testid}`');
        }
        const [, title, field] = anchored;
        if (FORMATTING.test(title)) {
          fail(n, 'the title of a section takes plain text only');
        }
        if (doc.sections.some((section) => section.field === field)) {
          fail(n, `the anchor "${field}" is used by two sections`);
        }
        const section = {
          field,
          title,
          line: n,
          blocks: [],
        };
        doc.sections.push(section);
        blocks = section.blocks;
      } else {
        fail(n, 'headings deeper than ## are not supported');
      }
    } else if (/^- \S/.test(line)) {
      if (paragraph) {
        flush();
      }
      target(n);
      if (!list) {
        list = { items: [] };
      }
      list.items.push({ line: n, text: line.slice(2).trim() });
    } else if (/^\s/.test(line)) {
      // An indented line continues the list item above it and nothing else.
      if (!list) {
        fail(n, 'indented text is not supported');
      }
      if (/^\s+[-*+] /.test(line)) {
        fail(n, 'nested lists are not supported');
      }
      const item = list.items[list.items.length - 1];
      item.text = joinWrapped(item.text, line);
    } else if (/^([*+]|\d+[.)])\s/.test(line)) {
      fail(n, 'only `-` lists are supported');
    } else if (line.startsWith('>')) {
      fail(n, 'block quotes are not supported');
    } else if (line.startsWith('|')) {
      fail(n, 'tables are not supported');
    } else if (/^(-{3,}|\*{3,}|_{3,}|={3,})\s*$/.test(line)) {
      fail(n, 'rules and underlined headings are not supported');
    } else {
      if (list) {
        flush();
      }
      target(n);
      if (paragraph) {
        paragraph.text = joinWrapped(paragraph.text, line);
      } else {
        paragraph = { line: n, text: line.trim() };
      }
    }
  }
  flush();

  if (doc.title == null) {
    fail(next, 'the document has no `# Title`');
  }
  if (!doc.intro.some((block) => block.type === 'p')) {
    fail(next, 'the document needs a paragraph on what the dialog is for, under its title');
  }
  doc.sections.forEach((section) => {
    if (doc.skip.includes(section.field)) {
      fail(section.line, `the field "${section.field}" has a section and is in "skip" as well`);
    }
    if (section.blocks.length === 0) {
      fail(section.line, `the section "${section.title}" is empty`);
    }
  });

  return doc;
}
