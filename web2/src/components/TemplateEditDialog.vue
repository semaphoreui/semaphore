<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <v-dialog
    v-model="editDialog"
    max-width="400"
    persistent
    :transition="false"
  >
    <v-card>
      <v-card-title class="headline">{{ isNewItem ? 'New' : 'Edit' }} Template</v-card-title>

      <v-card-text>
        <TemplateEditForm :template-id="templateId" :project-id="projectId" ref="itemForm" />
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>

        <v-btn
          color="blue darken-1"
          text
          @click="editDialog = false"
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

import TemplateEditForm from '@/components/TemplateEditForm.vue';
import { getErrorMessage } from '@/lib/error';

export default {
  components: { TemplateEditForm },
  props: {
    projectId: Number,
    templateId: [Number, String],
    value: Boolean,
  },

  data() {
    return {
      editDialog: false,
      editFormSaving: false,
      editFormValid: false,
      editFormError: null,
    };
  },

  watch: {
    editDialog(val) {
      this.$emit('input', val);
    },

    async value(val) {
      if (!val) {
        this.editDialog = val;
        return;
      }
      this.editFormError = false;
      this.editDialog = val;
    },
  },

  methods: {
    isNewItem() {
      return this.templateId === 'new';
    },

    async save() {
      this.editFormSaving = true;
      try {
        const item = await this.$refs.itemForm.saveItem();
        if (item) {
          this.$emit('saved', {
            item,
            action: this.isNewItem ? 'new' : 'edit',
          });
          this.editDialog = false;
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
