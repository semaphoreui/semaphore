import { Marked } from 'marked';
import DOMPurify from 'dompurify';

/**
 * Only http(s), mailto, fragment and relative URLs are allowed.
 * Everything else (javascript:, vbscript:, data: on links, tel:, ...) is stripped.
 */
const ALLOWED_URI_REGEXP = /^(?:(?:https?|mailto):|[^a-z]|[a-z+.-]+(?:[^a-z+.\-:]|$))/i;

// Allowlist: text formatting and links only. Nothing that loads external
// resources (img, media, embeds, CSS) or restyles the app UI (style, class).
const PURIFY_CONFIG = {
  ALLOWED_TAGS: [
    'a', 'p', 'br', 'hr', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
    'strong', 'b', 'em', 'i', 'del', 's', 'sub', 'sup', 'kbd', 'code', 'pre',
    'blockquote', 'ul', 'ol', 'li', 'table', 'thead', 'tbody', 'tr', 'th', 'td',
    'details', 'summary',
  ],
  ALLOWED_ATTR: ['href', 'title', 'align', 'start', 'colspan', 'rowspan', 'open'],
  ALLOWED_URI_REGEXP,
  ALLOW_DATA_ATTR: false,
  SANITIZE_NAMED_PROPS: true,
};

function hardenLinks(node) {
  if (node.tagName === 'A' && node.hasAttribute('href')) {
    node.setAttribute('target', '_blank');
    node.setAttribute('rel', 'noopener noreferrer nofollow');
  }
}

// Markdown images become plain links, so viewing a description never fetches a remote URL.
const renderer = {
  image(token) {
    return this.link(token.text ? token : { ...token, text: token.href, autolink: true });
  },
};

/**
 * Renders GitHub-flavoured Markdown to sanitized HTML that is safe for v-html.
 *
 * A fresh Marked and DOMPurify instance is created per call, so neither the
 * Markdown options nor the sanitizer hooks leak into other users of the libraries.
 *
 * @param {string} source Raw Markdown.
 * @param {Window} [win] Window used by DOMPurify (defaults to the global window).
 * @returns {string} Sanitized HTML.
 */
export default function renderMarkdown(source, win) {
  if (source == null || source === '') {
    return '';
  }

  const marked = new Marked({
    gfm: true, breaks: true, async: false, renderer,
  });
  const html = marked.parse(String(source));

  const purify = DOMPurify(win || window);
  if (!purify.isSupported) {
    return '';
  }
  purify.addHook('afterSanitizeAttributes', hardenLinks);

  return purify.sanitize(html, PURIFY_CONFIG);
}
