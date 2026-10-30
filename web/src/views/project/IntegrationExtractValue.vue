<template>
  <div
    v-if="items != null"
    data-testid="integrationValue-section"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="integrationValueDialog"
      :save-button-text="itemId === 'new' ? 'Create' : 'Save'"
      :title="`${itemId === 'new' ? 'New' : 'Edit'} Extract Value`"
      :max-width="450"
      :transition="false"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <IntegrationExtractValueForm
          :projectId="projectId"
          :integration-id="integrationId"
          :item-id="itemId"
          :project-id="projectId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      test-id="integrationValueRefsDialog"
      object-title="extractvalue"
      :object-refs="itemRefs"
      :integration-id="integrationId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      test-id="integrationValueDeleteDialog"
      title="Delete Integration ExtractValue"
      text="Are you sure you want to delete this Integration ExtractValue?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="integrationValue-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="integrationValue-title"
      >ExtractValue</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
        color="primary"
        variant="elevated"
        @click="editItem('new')"
        data-testid="integrationValue-new"
        >New Extracted Value</v-btn
      >
    </v-toolbar>
    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page="-1"
      data-testid="integrationValue-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.name="{ item }"> {{ item.name }} {{ item.extractorId }} </template>
      <template v-slot:item.value_source="{ item }">
        <code data-testid="integrationValue-rowValueSource">{{ item.value_source }}</code>
      </template>
      <template v-slot:item.body_data_type="{ item }">
        <code data-testid="integrationValue-rowBodyDataType">{{ item.body_data_type }}</code>
      </template>
      <template v-slot:item.key="{ item }">
        <code data-testid="integrationValue-rowKey">{{ item.key }}</code>
      </template>
      <template v-slot:item.variable="{ item }">
        <code data-testid="integrationValue-rowVariable">{{ item.variable }}</code>
      </template>
      <template v-slot:item.variable_type="{ item }">
        <code data-testid="integrationValue-rowVariableType">{{ item.variable_type }}</code>
      </template>
      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap" data-testid="integrationValue-rowActions">
          <v-btn
            icon
            variant="text"
            class="mr-1"
            @click="askDeleteItem(item.id)"
            data-testid="integrationValue-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
            icon
            variant="text"
            class="mr-1"
            @click="editItem(item.id)"
            data-testid="integrationValue-rowEdit"
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

import IntegrationExtractValueForm from '@/components/IntegrationExtractValueForm.vue';

export default {
  mixins: [ItemListPageBase],
  components: { IntegrationExtractValueForm },

  computed: {
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
          title: 'Value Source',
          key: 'value_source',
          sortable: false,
        },
        // {
        //   title: 'Body Data Type',
        //   key: 'body_data_type',
        //   sortable: false,
        // },
        {
          title: 'Key',
          key: 'key',
          sortable: false,
        },
        {
          title: 'Variable',
          key: 'variable',
          sortable: false,
        },
        {
          title: 'Variable Type',
          key: 'variable_type',
          sortable: false,
        },
        {
          title: 'Actions',
          key: 'actions',
          sortable: false,
        },
      ];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/values`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.integrationId}/values/${this.itemId}`;
    },
    getEventName() {
      return 'w-integration-extract-value';
    },
  },
};
</script>
