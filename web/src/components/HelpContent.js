// Draws the blocks of a help document (a tree built at build time, web/vite/help). Every element
// is created here from a fixed set, and the text of a document goes into text nodes: nothing of a
// document is ever parsed as HTML.
import { h } from 'vue';

function inline(node, key) {
  switch (node.type) {
    case 'code':
      return h('code', { key }, node.text);
    case 'strong':
      return h('strong', { key }, node.text);
    case 'link':
      // The build accepts https links only; a link of any other kind is shown as its text.
      if (typeof node.href !== 'string' || !node.href.startsWith('https://')) {
        return node.text;
      }
      return h('a', {
        key,
        href: node.href,
        target: '_blank',
        rel: 'noopener noreferrer',
      }, node.text);
    default:
      return node.text;
  }
}

function block(node, key) {
  switch (node.type) {
    case 'ul':
      return h('ul', { key }, node.items.map((item, i) => h('li', { key: i }, item.map(inline))));
    case 'code':
      return h('pre', { key }, [h('code', node.text)]);
    default:
      return h('p', { key }, node.inline.map(inline));
  }
}

export default {
  name: 'HelpContent',

  props: {
    blocks: {
      type: Array,
      required: true,
    },
  },

  render() {
    return h('div', { class: 'HelpContent' }, this.blocks.map(block));
  },
};
