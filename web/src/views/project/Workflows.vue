<template>
  <div>
    <v-toolbar flat>
      <v-toolbar-title>{{ $t('workflows') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="createDialog = true"
      >
        {{ $t('new_workflow') }}
      </v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      :loading="loading"
      class="elevation-1"
    >
      <template v-slot:item.name="{ item }">
        <router-link :to="`/project/${projectId}/workflows/${item.id}`">
          {{ item.name }}
        </router-link>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-icon
          small
          @click="deleteItem(item)"
        >
          mdi-delete
        </v-icon>
      </template>
    </v-data-table>

    <v-dialog v-model="createDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">{{ $t('new_workflow') }}</span>
        </v-card-title>

        <v-card-text>
          <v-container>
            <v-row>
              <v-col cols="12">
                <v-text-field
                  v-model="editedItem.name"
                  :label="$t('name')"
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <v-textarea
                  v-model="editedItem.description"
                  :label="$t('description')"
                ></v-textarea>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="close">{{ $t('cancel') }}</v-btn>
          <v-btn color="blue darken-1" text @click="save">{{ $t('save') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
import axios from 'axios';

export default {
  data: () => ({
    loading: false,
    items: [],
    createDialog: false,
    editedItem: {
      name: '',
      description: '',
    },
    projectId: null,
  }),

  computed: {
    headers() {
      return [
        { text: this.$t('name'), value: 'name' },
        { text: this.$t('description'), value: 'description' },
        { text: this.$t('created'), value: 'created_at' },
        { text: this.$t('actions'), value: 'actions', sortable: false },
      ];
    },
  },

  created() {
    this.projectId = this.$route.params.projectId;
    this.load();
  },

  methods: {
    async load() {
      this.loading = true;
      try {
        const response = await axios.get(`/api/project/${this.projectId}/workflows`);
        this.items = response.data;
      } catch (e) {
        // handle error
      } finally {
        this.loading = false;
      }
    },

    close() {
      this.createDialog = false;
      this.editedItem = { name: '', description: '' };
    },

    async save() {
      try {
        await axios.post(`/api/project/${this.projectId}/workflows`, this.editedItem);
        this.load();
        this.close();
      } catch (e) {
        // handle error
      }
    },

    async deleteItem() {
      // if (confirm(this.$t('confirm_delete'))) {
      //   try {
      //     await axios.delete(`/api/project/${this.projectId}/workflows/${item.id}`);
      //     this.load();
      //   } catch (e) {
      //     // handle error
      //   }
      // }
    },
  },
};
</script>
