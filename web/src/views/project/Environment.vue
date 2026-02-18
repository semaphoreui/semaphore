<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
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
          :support-storages="premiumFeatures.secret_storages"
          @maximize="editNoEscape = $event.maximized"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="environment"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('deleteEnvironment')"
      :text="$t('askDeleteEnv')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('environment') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('newEnvironment') }}
      </v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4 CenterToScreen"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.name="{ item }">
        <a @click="editItem(item.id)">{{ item.name }}</a>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="askDeleteItem(item.id)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn @click="editItem(item.id)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import EnvironmentForm from '@/components/EnvironmentForm.vue';
import PageMixin from '@/components/PageMixin';

export default {
  components: { EnvironmentForm },
  mixins: [ItemListPageBase, PageMixin],
  data() {
    return {
      editNoEscape: false,
    };
  },
  methods: {
    getHeaders() {
      return [{
        text: this.$i18n.t('name'),
        value: 'name',
        width: '100%',
      },
      {
        value: 'actions',
        sortable: false,
      }];
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
