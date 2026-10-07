<template>
  <div class="MarkdownView" :class="{ 'MarkdownView--collapsed': isCollapsed }">
    <!-- eslint-disable vue/no-v-html -- html is sanitized by DOMPurify in renderMarkdown() -->
    <div
      ref="content"
      class="MarkdownView__content"
      :style="isCollapsed ? { maxHeight: `${collapsedHeight}px` } : null"
      v-html="html"
      data-testid="markdown-view"
    ></div>
    <!-- eslint-enable vue/no-v-html -->

    <v-btn
      v-if="collapsible && overflows"
      x-small
      text
      color="primary"
      class="MarkdownView__toggle px-0"
      @click="expanded = !expanded"
    >
      {{ expanded ? $t('markdownShowLess') : $t('markdownShowMore') }}
      <v-icon small>{{ expanded ? 'mdi-chevron-up' : 'mdi-chevron-down' }}</v-icon>
    </v-btn>
  </div>
</template>

<style lang="scss">
.MarkdownView__content {
  line-height: 1.5;
  overflow-wrap: anywhere;
  overflow: hidden;
  position: relative;

  > :first-child {
    margin-top: 0;
  }

  > :last-child {
    margin-bottom: 0;
  }

  h1, h2, h3, h4, h5, h6 {
    font-weight: 500;
    line-height: 1.3;
    margin: 0.8em 0 0.4em;
  }

  h1 { font-size: 1.5em; }
  h2 { font-size: 1.3em; }
  h3 { font-size: 1.15em; }
  h4, h5, h6 { font-size: 1em; }

  p, ul, ol, pre, table, blockquote {
    margin-bottom: 0.6em;
  }

  ul, ol {
    padding-left: 1.5em;
  }

  code {
    font-size: 0.9em;
    padding: 0.1em 0.3em;
    border-radius: 3px;
    background-color: rgba(128, 128, 128, 0.15);
  }

  pre {
    padding: 8px 12px;
    border-radius: 4px;
    overflow-x: auto;
    background-color: rgba(128, 128, 128, 0.15);

    code {
      padding: 0;
      background: none;
    }
  }

  // display: block lets wide tables scroll instead of stretching the page.
  table {
    border-collapse: collapse;
    display: block;
    max-width: 100%;
    overflow-x: auto;
  }

  th, td {
    border: 1px solid rgba(128, 128, 128, 0.4);
    padding: 4px 10px;
    text-align: left;
  }

  blockquote {
    border-left: 4px solid rgba(128, 128, 128, 0.4);
    padding-left: 12px;
    opacity: 0.85;
  }

  img {
    max-width: 100%;
  }

  hr {
    border: 0;
    border-top: 1px solid rgba(128, 128, 128, 0.4);
    margin: 0.8em 0;
  }
}

.MarkdownView--collapsed .MarkdownView__content {
  -webkit-mask-image: linear-gradient(to bottom, #000 calc(100% - 1.2em), transparent);
  mask-image: linear-gradient(to bottom, #000 calc(100% - 1.2em), transparent);
}

.MarkdownView__toggle {
  text-transform: none !important;
  letter-spacing: normal !important;
}
</style>

<script>
import { renderMarkdown } from '@/lib/markdown';

export default {
  props: {
    source: {
      type: String,
      default: '',
    },
    collapsible: {
      type: Boolean,
      default: false,
    },
    // Number of text lines shown while collapsed.
    maxLines: {
      type: Number,
      default: 3,
    },
  },

  data() {
    return {
      expanded: false,
      overflows: false,
      lineHeight: 21,
    };
  },

  computed: {
    html() {
      return renderMarkdown(this.source);
    },

    collapsedHeight() {
      return Math.round(this.lineHeight * this.maxLines);
    },

    isCollapsed() {
      return this.collapsible && this.overflows && !this.expanded;
    },
  },

  watch: {
    html() {
      this.$nextTick(() => this.measure());
    },
  },

  mounted() {
    this.measure();
    if (typeof ResizeObserver !== 'undefined') {
      this.resizeObserver = new ResizeObserver(() => this.measure());
      this.resizeObserver.observe(this.$refs.content);
    }
  },

  beforeDestroy() {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
    }
  },

  methods: {
    measure() {
      const el = this.$refs.content;
      if (!el || !this.collapsible) {
        this.overflows = false;
        return;
      }
      const lineHeight = parseFloat(window.getComputedStyle(el).lineHeight);
      if (lineHeight > 0) {
        this.lineHeight = lineHeight;
      }
      // scrollHeight is the full content height even while max-height is applied.
      // Allow half a line of slack so a description just over the limit isn't collapsed.
      this.overflows = el.scrollHeight > this.collapsedHeight + this.lineHeight / 2;
    },
  },
};
</script>
