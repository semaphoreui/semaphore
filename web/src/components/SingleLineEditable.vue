<template>
  <div>
    <v-icon size="small" class="pr-1 SingleLineEditable__icon">mdi-message-outline</v-icon>

    <span
      key="placeholder"
      v-if="!editable"
      class="SingleLineEditable__content"
      data-testid="singleLineEditable-text"
      :class="{
        'SingleLineEditable__content--placeholder': !modelValue,
      }"
    >{{ modelValue || placeholder || $t('empty') }}</span>

    <span
      key="content"
      v-else
      ref="editableDiv"
      :contenteditable="true"
      class="SingleLineEditable__editable"
      data-testid="singleLineEditable-input"
      @paste="handlePaste"
      @keypress="handleKeypress"
    >{{ modelValue }}</span>

    <v-btn
      icon
      variant="text" @click="edit()"
      size="small"
      v-if="canEdit && !editable"
      class="ml-1 SingleLineEditable__button"
      data-testid="singleLineEditable-edit"
    >
      <v-icon>mdi-pencil</v-icon>
    </v-btn>

    <v-btn
      icon
      variant="text"
      @click="save()"
      v-if="editable"
      class="ml-2 SingleLineEditable__button"
      size="small"
      data-testid="singleLineEditable-save"
    >
      <v-icon color="green">mdi-check</v-icon>
    </v-btn>

    <v-btn
      icon
      variant="text"
      @click="cancel()"
      v-if="editable"
      class="SingleLineEditable__button"
      size="small"
      data-testid="singleLineEditable-cancel"
    >
      <v-icon color="red">mdi-close</v-icon>
    </v-btn>
  </div>
</template>
<style lang="scss" scoped>
  // The icon of Vuetify 2 (`.theme--light.v-icon`): 16 px of glyph plus its padding (Vuetify 4 fits
  // the padding into the 16 px box), grey .54 in the light theme, white in the dark one.
  .SingleLineEditable__icon {
    box-sizing: content-box;

    &.v-theme--light {
      color: rgba(0, 0, 0, .54);
    }

    &.v-theme--dark {
      color: #fff;
    }
  }

  .SingleLineEditable__editable {
    min-width: 50px;
    outline: none;
    background-color: rgba(128, 128, 128, 0.4);
    border-radius: 4px;
    padding: 2px 5px;
  }

  .SingleLineEditable__content {
    padding: 2px 5px;
  }

  .SingleLineEditable__content--placeholder {
    color: grey;
  }

  .SingleLineEditable__button {
    margin-bottom: -7px;
    margin-top: -10px;
  }
</style>
<script>
export default {
  props: {
    modelValue: String,
    placeholder: String,
    canEdit: Boolean,
  },
  emits: ['update:modelValue', 'save'],
  data() {
    return {
      editable: false,
    };
  },
  watch: {
    modelValue(newVal) {
      if (this.$refs.editableDiv.innerText !== newVal) {
        this.$refs.editableDiv.innerText = newVal || '';
      }
    },
  },
  methods: {

    handlePaste(e) {
      if (!this.editable) {
        return;
      }
      e.preventDefault();
      const text = (e.clipboardData || window.clipboardData).getData('text');
      document.execCommand('insertText', false, text.replace(/\n/g, ' '));
    },

    edit() {
      this.editable = true;
      setTimeout(() => {
        this.$refs.editableDiv.focus({ preventScroll: true });
      }, 100);
    },

    save() {
      this.editable = false;
      this.$emit('update:modelValue', this.$refs.editableDiv.innerText);
      this.$emit('save');
    },

    cancel() {
      this.$refs.editableDiv.innerText = this.modelValue;
      this.editable = false;
    },

    handleKeypress(e) {
      if (!this.editable) {
        return;
      }
      if (e.key === 'Enter') {
        e.preventDefault();
        this.save();
      }
    },
  },
};
</script>
