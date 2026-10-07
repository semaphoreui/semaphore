import { Marked } from 'marked';
import DOMPurify from 'dompurify';

/**
 * Only http(s), mailto, fragment and relative URLs are allowed.
 * Everything else (javascript:, vbscript:, data: on links, tel:, ...) is stripped.
 */
const ALLOWED_URI_REGEXP = /^(?:(?:https?|mailto):|[^a-z]|[a-z+.-]+(?:[^a-z+.\-:]|$))/i;

const PURIFY_CONFIG = {
  USE_PROFILES: { html: true },
  ALLOWED_URI_REGEXP,
  ALLOW_DATA_ATTR: false,
  // Elements that can change page behaviour or escape the description box.
  FORBID_TAGS: [
    'style', 'form', 'input', 'button', 'textarea', 'select', 'option',
    'iframe', 'frame', 'frameset', 'object', 'embed', 'audio', 'video', 'source',
    'track', 'dialog', 'template', 'base', 'link', 'meta',
  ],
  // `style`/`class` would let the description overlay or restyle the app UI.
  FORBID_ATTR: ['style', 'class', 'target', 'srcset', 'formaction'],
  SANITIZE_NAMED_PROPS: true,
};

function hardenLinks(node) {
  if (node.tagName === 'A' && node.hasAttribute('href')) {
    node.setAttribute('target', '_blank');
    node.setAttribute('rel', 'noopener noreferrer nofollow');
  }
  if (node.tagName === 'IMG') {
    node.setAttribute('referrerpolicy', 'no-referrer');
    node.setAttribute('loading', 'lazy');
  }
}

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

  const marked = new Marked({ gfm: true, breaks: true, async: false });
  const html = marked.parse(String(source));

  const purify = DOMPurify(win || window);
  if (!purify.isSupported) {
    return '';
  }
  purify.addHook('afterSanitizeAttributes', hardenLinks);

  return purify.sanitize(html, PURIFY_CONFIG);
}
