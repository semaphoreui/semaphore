<template>
  <div
    v-if="items != null && keys != null"
    data-testid="repositories-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="repositoriesDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="`${itemId === 'new' ? $t('nnew') : $t('edit')} Repository`"
      @save="loadItems()"
      :max-width="450"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <RepositoryForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      test-id="repositoriesRefsDialog"
      object-title="repository"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      test-id="repositoriesDeleteDialog"
      :title="$t('deleteRepository')"
      :text="$t('askDeleteRepo')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="repositories-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="repositories-title"
      >{{ $t('repositories') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        data-testid="repositories-new"
      >{{ $t('newRepository') }}</v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="-1"
      :mobile-breakpoint="600"
      data-testid="repositories-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 100px); margin: auto;"
    >
      <template v-slot:item.git_url="{ item }">
        {{ item.git_url }}
        <code
          v-if="!isLocalRepositoryPath(item.git_url)"
          data-testid="repositories-rowBranch"
        >{{ item.git_branch }}</code>
      </template>

      <template v-slot:item.ssh_key_id="{ item }">
        {{ keys.find((k) => k.id === item.ssh_key_id).name }}
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group
          density="compact"
          data-testid="repositories-rowActions"
        >
          <v-btn
            @click="askDeleteItem(item.id)"
            data-testid="repositories-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            @click="editItem(item.id)"
            data-testid="repositories-rowEdit"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-group>
      </template>
    </v-data-table>
  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import RepositoryForm from '@/components/RepositoryForm.vue';
import { isLocalRepositoryPath } from '@/lib/repositoryUrl';
import axios from 'axios';

export default {
  mixins: [ItemListPageBase],
  components: { RepositoryForm },
  data() {
    return {
      keys: null,
    };
  },

  async created() {
    this.keys = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },

  methods: {
    isLocalRepositoryPath,

    getHeaders() {
      return [{
        title: this.$t('name'),
        key: 'name',
        width: '25%',
      },
      {
        title: this.$t('gitUrl'),
        key: 'git_url',
        width: '50%',
      },
      {
        title: this.$t('sshKey'),
        key: 'ssh_key_id',
        width: '25%',
      },
      {
        title: this.$t('actions'),
        key: 'actions',
        sortable: false,
      }];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/repositories`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/repositories/${this.itemId}`;
    },
    getEventName() {
      return 'i-repositories';
    },
  },
};
</script>
