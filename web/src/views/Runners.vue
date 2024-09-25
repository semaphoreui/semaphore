<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      save-button-text="Save"
      :title="$t('editUser')"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <RunnerForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :is-admin="true"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteRunner')"
      :text="$t('askDeleteRunner')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat >
      <v-btn
        icon
        class="mr-4"
        @click="returnToProjects()"
      >
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
      <v-toolbar-title>{{ $t('runners') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >{{ $t('newRunner') }}</v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.external="{ item }">
        <v-icon v-if="item.external">mdi-checkbox-marked</v-icon>
        <v-icon v-else>mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.alert="{ item }">
        <v-icon v-if="item.alert">mdi-checkbox-marked</v-icon>
        <v-icon v-else>mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.admin="{ item }">
        <v-icon v-if="item.admin">mdi-checkbox-marked</v-icon>
        <v-icon v-else>mdi-checkbox-blank-outline</v-icon>
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
            icon
            class="mr-1"
            @click="askDeleteItem(item.id)"
            :disabled="item.id === userId"
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
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import EditDialog from '@/components/EditDialog.vue';
import RunnerForm from '@/components/RunnerForm.vue';

export default {
  mixins: [ItemListPageBase],

  components: {
    RunnerForm,
    YesNoDialog,
    EditDialog,
  },

  methods: {
    getHeaders() {
      return [{
        text: this.$i18n.t('name'),
        value: 'name',
        width: '50%',
      },
      {
        text: this.$i18n.t('active'),
        value: 'active',
      }];
    },

    async returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },

    getItemsUrl() {
      return '/api/runners';
    },

    getSingleItemUrl() {
      return `/api/runners/${this.itemId}`;
    },

    getEventName() {
      return 'i-runner';
    },
  },
};
</script>
