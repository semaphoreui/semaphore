<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      save-button-text="Save"
      title="Edit Environment"
      :max-width="500"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <EnvironmentForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <EditDialog
      v-model="itemRefsDialog"
      title="Environment in use"
      :max-width="500"
      hide-buttons
    >
      <template>
        <ObjectRefsView :object-refs="itemRefs" />
      </template>
    </EditDialog>

    <YesNoDialog
      title="Delete environment"
      text="Are you really want to delete this environment?"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat color="white">
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Environment</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >New Environment</v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
    >
      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
            icon
            class="mr-1"
            @click="askDeleteItem(item.id)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
            icon
            class="mr-1"
            @click="editItem(item.id)"
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
import EnvironmentForm from '@/components/EnvironmentForm.vue';

export default {
  components: { EnvironmentForm },
  mixins: [ItemListPageBase],
  methods: {
    getHeaders() {
      return [{
        text: 'Name',
        value: 'name',
        width: '100%',
      },
      {
        text: 'Actions',
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
