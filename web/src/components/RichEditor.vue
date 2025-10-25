<template>
  <div>
    <v-dialog
        v-model="envEditorDialog"
        persistent
        :fullscreen="true"
        :transition="false"
    >
      <div style="position: relative; height: 100%;">
        <codemirror
            v-if="envEditorDialog"
            class="EnvironmentMaximizedEditor"
            :style="{ border: '1px solid lightgray' }"
            v-model="text"
            :options="cmOptions"
            :placeholder="$t('enterExtraVariablesJson')"
        />

        <v-btn
            v-if="validatable"
            dark
            fab
            small
            color="success"
            style="
            position: absolute;
            right: 70px;
            top: 0;
            margin: 10px;
          "
            @click="spellcheck()"
        >
          <v-icon>mdi-spellcheck</v-icon>
        </v-btn>

        <v-btn
            dark
            fab
            small
            color="blue-grey"
            style="
            position: absolute;
            right: 20px;
            top: 0;
            margin: 10px;
          "
            @click="save()"
        >
          <v-icon>mdi-arrow-collapse</v-icon>
        </v-btn>

        <v-alert
            v-model="showAlert"
            :color="errorMessage ? 'error' : 'success'"
            dismissible
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
        dark
        fab
        small
        color="blue-grey"
        @click="envEditorDialog = true"
    >
      <v-icon>mdi-arrow-expand</v-icon>
    </v-btn>

  </div>
</template>

<script>
/* eslint-disable import/no-extraneous-dependencies,import/extensions */
import { codemirror } from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/vue/vue.js';
import 'codemirror/addon/display/placeholder.js';
import { getErrorMessage } from '../lib/error';
// import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    value: String,
    type: String,
  },

  components: {
    codemirror,
  },

  watch: {
    envEditorDialog(val) {
      this.$emit('maximize', {
        maximized: val,
      });
    },

    value() {
      this.text = this.value;
    },
  },

  created() {
    this.text = this.value;
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

    cmOptions() {
      return {
        tabSize: 2,
        mode: 'application/json',
        lineNumbers: true,
        line: true,
        lint: true,
        indentWithTabs: false,
      };
    },

    validatable() {
      return ['json', 'json_array'].includes(this.type);
    },

    validationSuccessMessage() {
      switch (this.type) {
        case 'json':
          return 'Valid JSON format.';
        case 'json_array':
          return 'Valid JSON array format.';
        default:
          return 'Validation passed successfully.';
      }
    },
  },

  methods: {
    cancel() {
      this.errorMessage = null;
      this.text = this.value;
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
        default:
      }
      this.showAlert = true;
    },
    save() {
      // this.errorMessage = null;
      // switch (this.type) {
      //   case 'json':
      //     try {
      //       JSON.parse(this.text);
      //     } catch (e) {
      //       this.errorMessage = getErrorMessage(e);
      //       return;
      //     }
      //     break;
      //   case 'json_array':
      //     try {
      //       const res = JSON.parse(this.text);
      //       if (!Array.isArray(res)) {
      //         throw new Error('Must be JSON array');
      //       }
      //     } catch (e) {
      //       this.errorMessage = getErrorMessage(e);
      //       return;
      //     }
      //     break;
      //   default:
      // }
      if (this.text !== this.value) {
        this.$emit('input', this.text);
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

  .CodeMirror {
    height: 100% !important;
    font-size: 14px;
    border-radius: 0 !important;
  }
}
</style>
