<template>
  <div v-if="!isLoaded">
    <v-progress-linear
      indeterminate
      color="primary darken-2"
    ></v-progress-linear>
  </div>
  <div v-else>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ $t('workflows') }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        class="mr-1"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        @click="editItem('new')"
      >
        {{ $t('newWorkflow') }}
      </v-btn>
    </v-toolbar>

    <v-dialog
      v-model="editDialog"
      :max-width="600"
      persistent
    >
      <v-card>
        <v-card-title>
          {{ itemId === 'new' ? $t('newWorkflow') : $t('editWorkflow') }}
          <v-spacer></v-spacer>
          <v-btn icon @click="closeEditDialog()">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="editItemData.name"
            :label="$t('name')"
            required
          ></v-text-field>
          <v-textarea
            v-model="editItemData.description"
            :label="$t('description')"
            rows="3"
          ></v-textarea>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="closeEditDialog()">
            {{ $t('cancel') }}
          </v-btn>
          <v-btn color="primary" @click="saveItem()">
            {{ $t('save') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-data-table
      :headers="headers"
      :items="items"
      :items-per-page="15"
      class="elevation-1"
    >
      <template v-slot:item.name="{ item }">
        <router-link :to="`/project/${projectId}/workflows/${item.id}/editor`">
          {{ item.name }}
        </router-link>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn
          icon
          small
          @click="editItem(item.id)"
          v-if="can(USER_PERMISSIONS.manageProjectResources)"
        >
          <v-icon small>mdi-pencil</v-icon>
        </v-btn>
        <v-btn
          icon
          small
          @click="deleteItem(item)"
          v-if="can(USER_PERMISSIONS.manageProjectResources)"
        >
          <v-icon small>mdi-delete</v-icon>
        </v-btn>
      </template>
    </v-data-table>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';
import api from '@/lib/api';
import { USER_PERMISSIONS } from '@/lib/constants';

export default {
  name: 'Workflows',
  mixins: [
    require('@/components/mixins/ProjectPermissionsMixin').default,
    require('@/components/mixins/DrawerMixin').default,
  ],
  data() {
    return {
      USER_PERMISSIONS,
      isLoaded: false,
      items: [],
      editDialog: false,
      itemId: null,
      editItemData: {
        name: '',
        description: '',
      },
      headers: [
        { text: this.$t('name'), value: 'name' },
        { text: this.$t('description'), value: 'description' },
        { text: this.$t('created'), value: 'created' },
        { text: this.$t('actions'), value: 'actions', sortable: false },
      ],
    };
  },
  computed: {
    ...mapGetters(['projectId']),
  },
  mounted() {
    this.loadItems();
  },
  methods: {
    async loadItems() {
      try {
        this.isLoaded = false;
        const { data } = await api.get(`/project/${this.projectId}/workflows`);
        this.items = data;
      } catch (err) {
        this.$store.dispatch('setError', err);
      } finally {
        this.isLoaded = true;
      }
    },
    editItem(id) {
      this.itemId = id;
      if (id === 'new') {
        this.editItemData = {
          name: '',
          description: '',
        };
      } else {
        const item = this.items.find((i) => i.id === id);
        this.editItemData = {
          name: item.name,
          description: item.description || '',
        };
      }
      this.editDialog = true;
    },
    async saveItem() {
      try {
        if (this.itemId === 'new') {
          await api.post(`/project/${this.projectId}/workflows`, this.editItemData);
        } else {
          await api.put(`/project/${this.projectId}/workflows/${this.itemId}`, this.editItemData);
        }
        this.closeEditDialog();
        this.loadItems();
      } catch (err) {
        this.$store.dispatch('setError', err);
      }
    },
    async deleteItem(item) {
      if (confirm(this.$t('confirmDelete', { name: item.name }))) {
        try {
          await api.delete(`/project/${this.projectId}/workflows/${item.id}`);
          this.loadItems();
        } catch (err) {
          this.$store.dispatch('setError', err);
        }
      }
    },
    closeEditDialog() {
      this.editDialog = false;
      this.itemId = null;
    },
  },
};
</script>
