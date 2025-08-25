<template>
  <div>
    <v-dialog
      v-model="envEditorDialog"
      max-width="800"
      persistent
      :transition="false"
    >
      <div style="position: relative;">
        <codemirror
          v-if="envEditorDialog"
          class="EnvironmentMaximizedEditor"
          :style="{ border: '1px solid lightgray' }"
          v-model="text"
          :options="cmOptions"
          :placeholder="$t('enterExtraVariablesJson')"
        />

        <v-btn
          dark
          fab
          small
          color="success"
          style="
            position: absolute;
            right: 50px;
            top: 0;
            margin: 10px;
          "
          @click="save()"
        >
          <v-icon>mdi-check</v-icon>
        </v-btn>

        <v-btn
          dark
          fab
          small
          color="error"
          style="
            position: absolute;
            right: 0;
            top: 0;
            margin: 10px;
          "
          @click="cancel()"
        >
          <v-icon>mdi-close</v-icon>
        </v-btn>

        <v-alert
          v-model="hasError"
          dismissible
          style="
            position: absolute;
            bottom: 0;
            left: 50%;
            transform: translateX(-50%);
          "
        >{{ errorMessage }}</v-alert>
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
import { getErrorMessage } from '@/lib/error';

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
    };
  },

  computed: {
    hasError: {
      get() {
        return this.errorMessage != null;
      },
      set(value) {
        if (!value) {
          this.errorMessage = null;
        }
      },
    },

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
  },

  methods: {
    cancel() {
      this.errorMessage = null;
      this.text = this.value;
      this.envEditorDialog = false;
    },
    save() {
      this.errorMessage = null;
      switch (this.type) {
        case 'json':
          try {
            JSON.parse(this.text);
          } catch (e) {
            this.errorMessage = getErrorMessage(e);
            return;
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
            return;
          }
          break;
        default:
      }
      if (this.text !== this.value) {
        this.$emit('input', this.text);
      }
      this.envEditorDialog = false;
    },
  },
};
</script>
<style lang="scss">
.EnvironmentMaximizedEditor {
  .CodeMirror {
    font-size: 14px;
    height: 600px !important;
  }
}
</style>
