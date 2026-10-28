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

        <v-btn
            v-if="validatable"
            icon
            size="small"
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
            size="small"
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
        size="small"
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
