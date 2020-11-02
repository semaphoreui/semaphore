<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-toolbar flat color="white">
      <v-toolbar-title>
        {{ isNewItem ? 'New task template' : `Edit task template` }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="error"
        @click="goBack()"
        class="mr-2"
        :disabled="itemFormSaving"
      >
        <v-icon left>mdi-close</v-icon>
        Cancel
      </v-btn>
      <v-btn
        color="primary"
        @click="saveItem()"
        :disabled="itemFormSaving"
      >
        <v-icon left>mdi-content-save</v-icon>
        {{ isNewItem ? 'Create' : 'Save' }}
      </v-btn>
    </v-toolbar>

    <div style="max-width: 400px; margin: 80px auto auto;">
      <TemplateEditForm :template-id="itemId" :project-id="projectId" ref="itemForm" />
    </div>
  </div>

</template>
<style lang="scss">

</style>
<script>
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import TemplateEditForm from '@/components/TemplateEditForm.vue';

export default {
  components: { TemplateEditForm },
  props: {
    projectId: Number,
  },
  data() {
    return {
      itemFormValid: false,
      itemFormError: null,
      itemFormSaving: false,
    };
  },

  computed: {
    cancelPath() {
      let prevItemId;
      if (this.isNewItem) {
        if (this.$route.query.id) {
          prevItemId = this.$route.query.id;
        } else {
          prevItemId = '';
        }
      } else {
        prevItemId = this.item.id;
      }
      return `/project/${this.projectId}/templates/${prevItemId}`;
    },
    itemId() {
      return this.$route.params.templateId === 'new' ? 'new' : parseInt(this.$route.params.templateId, 10);
    },
    isNewItem() {
      return this.itemId === 'new';
    },
  },

  methods: {
    async goBack() {
      // TODO: Determine how page has been opened: by router or by address bar.
      const pageOpenedDirectly = false;
      if (pageOpenedDirectly) {
        await this.$router.replace({
          path: this.cancelPath,
        });
      } else {
        // eslint-disable-next-line no-restricted-globals
        history.go(-1);
      }
    },

    async saveItem() {
      this.itemFormError = null;
      try {
        const item = await this.$refs.itemForm.saveItem();
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.isNewItem ? `Template "${item.alias}" created` : `Template "${item.alias}" saved`,
        });

        await this.goBack();
      } catch (err) {
        this.itemFormError = getErrorMessage(err);
      }
    },
  },
};
</script>
