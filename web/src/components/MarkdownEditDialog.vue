<template>
  <v-dialog
    v-model="dialog"
    max-width="900"
    persistent
    scrollable
    :fullscreen="$vuetify.breakpoint.xsOnly"
    :transition="false"
  >
    <v-card data-testid="markdownEditDialog">
      <v-card-title class="pr-3">
        {{ title }}
        <v-spacer />
        <v-btn icon @click="close()" :disabled="saving">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <div class="d-flex align-center flex-wrap px-4">
        <v-tabs v-model="tab" class="MarkdownEditDialog__tabs" height="40">
          <v-tab data-testid="markdownEditDialog-write">{{ $t('markdownWrite') }}</v-tab>
          <v-tab data-testid="markdownEditDialog-preview">{{ $t('markdownPreview') }}</v-tab>
        </v-tabs>

        <input
          ref="fileInput"
          type="file"
          accept=".md,.markdown,text/markdown,text/plain"
          class="d-none"
          :aria-label="$t('markdownUploadFile')"
          data-testid="markdownEditDialog-file"
          @change="onFileSelected"
        >
        <v-btn
          text
          small
          class="my-1"
          :disabled="saving"
          @click="$refs.fileInput.click()"
          data-testid="markdownEditDialog-upload"
        >
          <v-icon left small>mdi-upload</v-icon>
          {{ $t('markdownUploadFile') }}
        </v-btn>
      </div>

      <v-divider />

      <v-card-text class="pt-4 MarkdownEditDialog__body">
        <v-alert
          v-if="error"
          type="error"
          dense
          text
          dismissible
          @input="error = null"
        >{{ error }}</v-alert>

        <v-textarea
          v-show="tab === 0"
          v-model="text"
          outlined
          auto-grow
          rows="12"
          class="MarkdownEditDialog__textarea"
          :placeholder="$t('markdownPlaceholder')"
          :hint="$t('markdownHint')"
          persistent-hint
          :disabled="saving"
          data-testid="markdownEditDialog-textarea"
        />

        <div v-if="tab === 1" class="MarkdownEditDialog__preview">
          <MarkdownView v-if="text" :source="text" />
          <div v-else class="grey--text">{{ $t('markdownNothingToPreview') }}</div>
        </div>
      </v-card-text>

      <v-divider />

      <v-card-actions class="px-4 py-3">
        <v-spacer />
        <v-btn text @click="close()" :disabled="saving">{{ $t('cancel') }}</v-btn>
        <v-btn
          color="primary"
          depressed
          :loading="saving"
          @click="save()"
          data-testid="markdownEditDialog-save"
        >{{ $t('save') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style lang="scss">
.MarkdownEditDialog__textarea textarea {
  font-family: 'Roboto Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace !important;
  font-size: 13px !important;
  line-height: 1.5 !important;
}

.MarkdownEditDialog__tabs {
  flex: 1 1 auto;
  width: auto;
}

.MarkdownEditDialog__body {
  min-height: 300px;
}

.MarkdownEditDialog__preview {
  min-height: 280px;
  padding: 8px 4px;
  font-size: 14px;
}
</style>

<script>
import MarkdownView from '@/components/MarkdownView.vue';

function readFileAsText(file) {
  if (typeof file.text === 'function') {
    return file.text();
  }
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = () => reject(reader.error);
    reader.readAsText(file);
  });
}

export default {
  components: { MarkdownView },

  props: {
    value: Boolean,
    source: {
      type: String,
      default: '',
    },
    title: String,
    /**
     * Called with the new text; must return a promise. The dialog stays open
     * (and keeps the text) when it rejects.
     */
    saveHandler: {
      type: Function,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      tab: 0,
      text: '',
      error: null,
      saving: false,
    };
  },

  watch: {
    value(val) {
      this.dialog = val;
      if (val) {
        this.reset();
      }
    },

    dialog(val) {
      this.$emit('input', val);
    },
  },

  created() {
    this.dialog = this.value;
    if (this.value) {
      this.reset();
    }
  },

  methods: {
    reset() {
      this.tab = 0;
      this.text = this.source || '';
      this.error = null;
      this.saving = false;
    },

    close() {
      this.dialog = false;
    },

    async onFileSelected(event) {
      const input = event.target;
      const file = input.files && input.files[0];
      // Allow selecting the same file again.
      input.value = '';

      if (!file) {
        return;
      }

      try {
        const content = await readFileAsText(file);
        this.error = null;
        this.text = content.replace(/\r\n?/g, '\n');
        this.tab = 0;
      } catch (err) {
        this.error = this.$t('markdownFileReadError');
      }
    },

    async save() {
      this.saving = true;
      this.error = null;
      try {
        await this.saveHandler(this.text);
        this.dialog = false;
      } catch (err) {
        // The handler reports the error itself; keep the dialog open with the text.
      } finally {
        this.saving = false;
      }
    },
  },
};
</script>
