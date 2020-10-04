<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="item != null">
    <v-toolbar flat color="white">
      <v-toolbar-title>
        {{ isNewItem ? 'New task template' : `Edit task template: ${item.alias}` }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="error"
        @click="goBack()"
        class="mr-2"
      >
        <v-icon left>mdi-close</v-icon>
        Cancel
      </v-btn>
      <v-btn
        color="primary"
        @click="saveItem()"
      >
        <v-icon left>mdi-content-save</v-icon>
        {{ isNewItem ? 'Create' : 'Save' }}
      </v-btn>
    </v-toolbar>
  </div>

</template>
<style lang="scss">

</style>
<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    projectId: Number,
  },
  data() {
    return {
      headers: [
        {
          text: 'Alias',
          value: 'alias',
        },
        {
          text: 'Playbook',
          value: 'playbook',
          sortable: false,
        },
        {
          text: 'SSH key',
          value: 'email',
          sortable: false,
        },
        {
          text: 'Inventory',
          value: 'inventory',
          sortable: false,
        },
        {
          text: 'Environment',
          value: 'environment',
          sortable: false,
        },
        {
          text: 'Repository',
          value: 'repository',
          sortable: false,
        },
        {
          text: 'Actions',
          value: 'actions',
          sortable: false,
        },
      ],
      item: null,
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
      return this.$route.params.templateId;
    },
    isNewItem() {
      return this.itemId === 'new';
    },
  },

  async created() {
    if (this.isNewItem) {
      this.item = {};
    } else {
      await this.loadItem();
    }
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

      if (!this.$refs.itemForm.validate()) {
        return;
      }

      this.itemFormSaving = true;
      try {
        await axios({
          method: this.isNewItem ? 'post' : 'put',
          url: this.isNewItem
            ? `/api/project/${this.projectId}/templates`
            : `/api/project/${this.projectId}/templates/${this.item.id}`,
          responseType: 'json',
          data: this.item,
        });

        if (this.isNewItem) {
          this.items.push(this.item);
        } else {
          const userIndex = this.items.findIndex((item) => this.item.id === item.id);
          if (userIndex !== -1) {
            this.items.splice(userIndex, 1, this.item);
          }
        }
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: this.isNewItem ? `Template "${this.item.name}" created` : `Template "${this.item.name}" changed`,
        });

        this.goBack();
      } catch (err) {
        this.itemFormError = getErrorMessage(err);
      }
    },

    async loadItem() {
      this.item = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/templates/${this.itemId}`,
        responseType: 'json',
      })).data;
    },
  },
};
</script>
