import { expect } from 'chai';
import renderMarkdown from '@/lib/markdown';

function toDom(html) {
  const div = document.createElement('div');
  div.innerHTML = html;
  return div;
}

describe('lib/markdown', () => {
  describe('renderMarkdown', () => {
    it('returns an empty string for empty input', () => {
      expect(renderMarkdown('')).to.equal('');
      expect(renderMarkdown(null)).to.equal('');
      expect(renderMarkdown(undefined)).to.equal('');
    });

    it('renders GFM: headings, emphasis, lists, code and tables', () => {
      const dom = toDom(renderMarkdown([
        '# Title',
        '',
        'Some **bold** and *italic* and `code`.',
        '',
        '- one',
        '- two',
        '',
        '```',
        'echo hi',
        '```',
        '',
        '| a | b |',
        '| - | - |',
        '| 1 | 2 |',
      ].join('\n')));

      expect(dom.querySelector('h1').textContent).to.equal('Title');
      expect(dom.querySelector('strong').textContent).to.equal('bold');
      expect(dom.querySelector('em').textContent).to.equal('italic');
      expect(dom.querySelectorAll('ul > li')).to.have.length(2);
      expect(dom.querySelector('pre > code').textContent).to.include('echo hi');
      expect(dom.querySelectorAll('table td')).to.have.length(2);
    });

    it('removes <script> elements', () => {
      const html = renderMarkdown('hello <script>alert(1)</script>');
      expect(html).to.not.include('<script');
      expect(html).to.not.include('alert(1)');
    });

    it('removes event handler attributes', () => {
      const dom = toDom(renderMarkdown('<img src="x" onerror="alert(1)"> <b onclick="alert(2)">b</b>'));
      expect(dom.querySelector('[onerror]')).to.equal(null);
      expect(dom.querySelector('[onclick]')).to.equal(null);
    });

    it('removes javascript: URLs from markdown and raw HTML links', () => {
      const dom = toDom(renderMarkdown('[x](javascript:alert(1)) <a href="JaVaScRiPt:alert(2)">y</a>'));
      dom.querySelectorAll('a').forEach((a) => {
        expect(a.getAttribute('href') || '').to.not.match(/javascript:/i);
      });
    });

    it('removes iframes, forms and style', () => {
      const html = renderMarkdown([
        '<iframe src="https://example.com"></iframe>',
        '<form action="/x"><input name="a"></form>',
        '<style>body{display:none}</style>',
        '<p style="position:fixed" class="v-overlay">p</p>',
      ].join('\n'));
      expect(html).to.not.match(/<iframe|<form|<input|<style|style=|class=/i);
    });

    it('opens links in a new tab without opener or referrer', () => {
      const dom = toDom(renderMarkdown('[site](https://example.com) <a href="https://e.org" target="_self" rel="opener">r</a>'));
      const links = dom.querySelectorAll('a');
      expect(links).to.have.length(2);
      links.forEach((a) => {
        expect(a.getAttribute('target')).to.equal('_blank');
        expect(a.getAttribute('rel')).to.equal('noopener noreferrer nofollow');
      });
      expect(links[0].getAttribute('href')).to.equal('https://example.com');
    });

    it('does not leak sanitizer hooks between calls', () => {
      renderMarkdown('[a](https://example.com)');
      // eslint-disable-next-line global-require
      const DOMPurify = require('dompurify').default || require('dompurify');
      const html = DOMPurify.sanitize('<a href="https://example.com">a</a>');
      expect(html).to.not.include('target=');
    });
  });
});
