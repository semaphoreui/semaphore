<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-dialog
    v-model="dialog"
    max-width="400"
    persistent
    :transition="false"
  >
    <v-card>
      <v-card-title class="headline">{{ isNew ? 'New' : 'Edit' }} Template</v-card-title>

      <v-card-text>
        <TemplateForm :template-id="templateId" :project-id="projectId" ref="form" />
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>

        <v-btn
          color="blue darken-1"
          text
          @click="dialog = false"
        >
          Cancel
        </v-btn>

        <v-btn
          color="blue darken-1"
          text
          @click="save()"
        >
          Save
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>

import TemplateForm from '@/components/TemplateForm.vue';
import { getErrorMessage } from '@/lib/error';

export default {
  components: { TemplateForm },
  props: {
    projectId: Number,
    templateId: [Number, String],
    value: Boolean,
  },

  data() {
    return {
      dialog: false,
      editFormSaving: false,
      editFormValid: false,
      editFormError: null,
    };
  },

  computed: {
    isNew() {
      return this.templateId === 'new';
    },
  },

  watch: {
    dialog(val) {
      this.$emit('input', val);
    },

    async value(val) {
      if (!val) {
        this.dialog = val;
        return;
      }
      this.editFormError = false;
      this.dialog = val;
    },
  },

  methods: {
    async save() {
      this.editFormSaving = true;
      try {
        const item = await this.$refs.form.saveItem();
        if (item) {
          this.$emit('saved', {
            item,
            action: this.isNew ? 'new' : 'edit',
          });
          this.dialog = false;
        }
      } catch (err) {
        this.editFormError = getErrorMessage(err);
      } finally {
        this.editFormSaving = false;
      }
    },
  },
};
</script>
