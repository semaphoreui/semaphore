<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div
    v-if="items != null"
    data-testid="varGroup-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      :save-button-text="$t('save')"
      :title="$t('editEnvironment')"
      :max-width="700"
      @save="loadItems"
      :help-button="true"
      :no-escape="editNoEscape"
      test-id="varGroupDialog"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset, needHelp }">
        <EnvironmentForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :need-help="needHelp"
          :support-storages="features.secret_storages"
          @maximize="editNoEscape = $event.maximized"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      test-id="varGroupRefsDialog"
      object-title="environment"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      test-id="varGroupDeleteDialog"
      :title="$t('deleteEnvironment')"
      :text="$t('askDeleteEnv')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="varGroup-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title class="flex-0-1" data-testid="varGroup-title">
        {{ $t('environment') }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        data-testid="varGroup-new"
        >{{ $t('newEnvironment') }}
      </v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4 CenterToScreen EnvironmentTable"
      :items-per-page="-1"
      data-testid="varGroup-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.name="{ item }">
        <a @click="editItem(item.id)" data-testid="varGroup-rowName">{{ item.name }}</a>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn-group
          density="compact"
          data-testid="varGroup-rowActions"
        >
          <v-btn
            v-if="item.sync_enabled"
            @click="syncItem(item.id)"
            :disabled="!(item.sync_paths && item.sync_paths.length > 0)"
            data-testid="varGroup-rowSync"
          >
            <v-icon>mdi-sync</v-icon>
          </v-btn>
          <v-btn @click="askDeleteItem(item.id)" data-testid="varGroup-rowDelete">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn @click="editItem(item.id)" data-testid="varGroup-rowEdit">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-group>
      </template>
    </v-data-table>
  </div>
</template>
<style lang="scss">
// Not an inline style: the utility `mt-4` of Vuetify 4 has no `!important` and would lose to it.
.EnvironmentTable {
  max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width) - 200px);
  margin: auto;
}
</style>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import EnvironmentForm from '@/components/EnvironmentForm.vue';
import PageMixin from '@/components/PageMixin';
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  components: { EnvironmentForm },
  mixins: [ItemListPageBase, PageMixin],
  data() {
    return {
      editNoEscape: false,
    };
  },
  methods: {
    async syncItem(itemId) {
      try {
        await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/environment/${itemId}/sync`,
          responseType: 'json',
        });
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Variable group synced successfully',
        });
        await this.loadItems();
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
    getHeaders() {
      return [
        {
          title: this.$t('name'),
          key: 'name',
          width: '100%',
        },
        {
          key: 'actions',
          sortable: false,
          align: 'end',
        },
      ];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/environment`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/environment/${this.itemId}`;
    },
    getEventName() {
      return 'i-environment';
    },
  },
};
</script>
