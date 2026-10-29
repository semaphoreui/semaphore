<template>
  <div
    v-if="items != null"
    data-testid="integrationMatcher-section"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="integrationMatcherDialog"
      :save-button-text="itemId === 'new' ? 'Create' : 'Save'"
      :title="`${itemId === 'new' ? 'New' : 'Edit'} Matcher`"
      :max-width="450"
      :transition="false"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <IntegrationMatcherForm
          :project-id="projectId"
          :integration-id="integrationId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      test-id="integrationMatcherRefsDialog"
      object-title="matcher"
      :object-refs="itemRefs"
      :integration-id="integrationId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      test-id="integrationMatcherDeleteDialog"
      title="Delete Integration Matcher"
      text="Are you sure you want to delete this Integration Matcher?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="integrationMatcher-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="integrationMatcher-title"
      >Matcher</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        data-testid="integrationMatcher-new"
        >New Matcher</v-btn
      >
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page="-1"
      data-testid="integrationMatcher-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.name="{ item }">
        {{ item.name }}
      </template>
      <template v-slot:item.match_type="{ item }">
        <code data-testid="integrationMatcher-rowMatchType">{{ item.match_type }}</code>
      </template>
      <template v-slot:item.method="{ item }">
        <code data-testid="integrationMatcher-rowMethod">{{ item.method }}</code>
      </template>
      <template v-slot:item.body_data_type="{ item }">
        {{ item.body_data_type || 'N/A' }}
      </template>
      <template v-slot:item.key="{ item }">
        <code data-testid="integrationMatcher-rowKey">{{ item.key }}</code>
      </template>
      <template v-slot:item.value="{ item }">
        <code data-testid="integrationMatcher-rowValue">{{ item.value }}</code>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="integrationMatcher-rowActions">
          <v-btn
            icon
            class="mr-1"
            @click="askDeleteItem(item.id)"
            data-testid="integrationMatcher-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
            icon
            class="mr-1"
            @click="editItem(item.id)"
            data-testid="integrationMatcher-rowEdit"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';

import IntegrationMatcherForm from '@/components/IntegrationMatcherForm.vue';

export default {
  mixins: [ItemListPageBase],
  components: { IntegrationMatcherForm },
  computed: {
    // projectId() {
    //   if (/^-?\d+$/.test(this.$route.params.projectId)) {
    //     return parseInt(this.$route.params.projectId, 10);
    //   }
    //   return this.$route.params.projectId;
    // },
    integrationId() {
      if (/^-?\d+$/.test(this.$route.params.integrationId)) {
        return parseInt(this.$route.params.integrationId, 10);
      }
      return this.$route.params.integrationId;
    },
  },
  methods: {
    allowActions() {
      return true;
    },
    getHeaders() {
      return [
        {
          title: 'Name',
          key: 'name',
          sortable: true,
        },
        {
          title: 'Match Type',
          key: 'match_type',
          sortable: true,
        },
        {
          title: 'Body Data Type',
          key: 'body_data_type',
          sortable: false,
        },
        {
          title: 'Key',
          key: 'key',
          sortable: true,
        },
        {
          title: 'Method',
          key: 'method',
          sortable: true,
        },
        {
          title: 'Value',
          key: 'value',
          sortable: true,
        },
        {
          title: '',
          key: 'actions',
          sortable: false,
        },
      ];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/matchers`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/matchers/${this.itemId}`;
    },
    getEventName() {
      return 'w-integration-matcher';
    },
  },
};
</script>
