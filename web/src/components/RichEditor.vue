<template>
  <div>
    <v-dialog
        v-model="envEditorDialog"
        persistent
        :fullscreen="true"
        :transition="false"
    >
      <div style="position: relative; height: 100%;" :data-testid="`${testId}Dialog`">
        <CodeMirror
            v-if="envEditorDialog"
            class="EnvironmentMaximizedEditor"
            :style="{ border: '1px solid lightgray' }"
            v-model="text"
            minimal
            :tab-size="2"
            :extensions="cmExtensions"
            :placeholder="placeholderText"
            :data-testid="`${testId}-code`"
        />

        <!-- `fab small` of Vuetify 2: a round button of 40 px (shadow: .rich-editor-fab) -->
        <v-btn
            v-if="validatable"
            icon
            size="40"
            class="rich-editor-fab"
            color="success"
            style="
            position: absolute;
            right: 70px;
            top: 0;
            margin: 10px;
          "
            @click="spellcheck()"
            :data-testid="`${testId}-validate`"
        >
          <v-icon>mdi-spellcheck</v-icon>
        </v-btn>

        <v-btn
            icon
            size="40"
            class="rich-editor-fab"
            color="blue-grey"
            style="
            position: absolute;
            right: 20px;
            top: 0;
            margin: 10px;
          "
            @click="save()"
            :data-testid="`${testId}-collapse`"
        >
          <v-icon>mdi-arrow-collapse</v-icon>
        </v-btn>

        <v-alert
            v-model="showAlert"
            :color="errorMessage ? 'error' : 'success'"
            closable
            class="rich-editor-result"
            :data-testid="`${testId}-result`"
            :data-state="errorMessage ? 'error' : 'ready'"
            style="
            position: absolute;
            bottom: 0;
            left: 50%;
            transform: translateX(-50%);
          "
        >{{ errorMessage || validationSuccessMessage }}
        </v-alert>
      </div>
    </v-dialog>

    <v-btn
        icon
        size="40"
        class="rich-editor-fab"
        color="blue-grey"
        @click="envEditorDialog = true"
        :data-testid="`${testId}-expand`"
    >
      <v-icon>mdi-arrow-expand</v-icon>
    </v-btn>

  </div>
</template>

<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */
import CodeMirror from 'vue-codemirror6';
import { load as loadYaml } from 'js-yaml';
import codeEditorExtensions from '@/lib/codeEditor';
import { getErrorMessage } from '@/lib/error';
// import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    modelValue: String,
    type: String,
    testId: {
      type: String,
      default: 'richEditor',
    },
  },

  components: {
    CodeMirror,
  },

  emits: ['maximize', 'update:modelValue'],

  watch: {
    envEditorDialog(val) {
      this.$emit('maximize', {
        maximized: val,
      });
    },

    modelValue() {
      this.text = this.modelValue;
    },
  },

  created() {
    this.text = this.modelValue;
  },

  data() {
    return {
      text: null,
      envEditorDialog: false,
      errorMessage: null,
      showAlert: false,
    };
  },

  computed: {

    cmExtensions() {
      return codeEditorExtensions({ language: this.type === 'yaml' ? 'yaml' : 'json' });
    },

    validatable() {
      return ['json', 'json_array', 'yaml'].includes(this.type);
    },

    placeholderText() {
      return this.type === 'yaml'
        ? this.$t('enterExtraVariablesYaml')
        : this.$t('enterExtraVariablesJson');
    },

    validationSuccessMessage() {
      switch (this.type) {
        case 'json':
          return 'Valid JSON format.';
        case 'json_array':
          return 'Valid JSON array format.';
        case 'yaml':
          return 'Valid YAML format.';
        default:
          return 'Validation passed successfully.';
      }
    },
  },

  methods: {
    cancel() {
      this.errorMessage = null;
      this.text = this.modelValue;
      this.envEditorDialog = false;
    },
    spellcheck() {
      this.errorMessage = null;
      switch (this.type) {
        case 'json':
          try {
            JSON.parse(this.text);
          } catch (e) {
            this.errorMessage = getErrorMessage(e);
          }
          break;
        case 'json_array':
          try {
            const res = JSON.parse(this.text);
            if (!Array.isArray(res)) {
              throw new Error('Must be JSON array');
            }
          } catch (e) {
            this.errorMessage = getErrorMessage(e);
          }
          break;
        case 'yaml':
          try {
            loadYaml(this.text);
          } catch (e) {
            this.errorMessage = getErrorMessage(e);
          }
          break;
        default:
      }
      this.showAlert = true;
    },
    save() {
      if (this.text !== this.modelValue) {
        this.$emit('update:modelValue', this.text);
      }
      this.envEditorDialog = false;
    },
  },
};
</script>
<style lang="scss">
// The shadow of a Vuetify 2 fab: elevation 6 (`.v-btn--fab`); Vuetify 4 has elevation classes up
// to 5 only, and the elevated variant draws level 2.
// A focused button (after a click, too) is lightened by its overlay at .24
// (`.v-btn:focus::before`); Vuetify 4 shows the overlay on `:focus-visible` only.
.v-btn.rich-editor-fab {
  box-shadow: 0 3px 5px -1px rgba(0, 0, 0, .2), 0 6px 10px 0 rgba(0, 0, 0, .14),
    0 1px 18px 0 rgba(0, 0, 0, .12);

  &:focus > .v-btn__overlay {
    opacity: .24;
  }
}

// The verdict of Vuetify 2 (`v-alert` with a colour and no type is in the light theme): the text
// in the primary text colour, the close button a grey icon button pulled 8 px into the padding
// (`.v-alert__dismissible`: margin -16px -8px -16px 8px). Vuetify 4 writes on a coloured alert in
// its `on-` colour (white) and puts the close button 16 px from the text.
.rich-editor-result {
  .v-alert__content {
    color: rgba(var(--v-theme-on-surface), var(--v-high-emphasis-opacity));
  }

  .v-alert__close {
    margin-inline: 8px -8px;
  }

  &.v-theme--light .v-alert__close > .v-btn {
    color: rgba(0, 0, 0, .54);
  }
}

.vue-codemirror.EnvironmentMaximizedEditor {
  height: 100% !important;
  border-radius: 0 !important;

  // CodeMirror 6 has no colours of its own; CodeMirror 5 drew the editor white with black
  // text, and it covers the page under the full-screen dialog.
  .cm-editor {
    height: 100% !important;
    font-size: 14px;
    border-radius: 0 !important;
    background: white;
    color: black;
  }
}
</style>
