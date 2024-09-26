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

    <v-toolbar flat>
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
      >{{ $t('newRunner') }}
      </v-btn>
    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :footer-props="{ itemsPerPageOptions: [20] }"
    >
      <template v-slot:item.active="{ item }">
        <v-switch
          v-model="item.active"
          inset
          @change="setActive(item.id, item.active)"
        ></v-switch>
      </template>

      <template v-slot:item.name="{ item }">{{ item.name || '&mdash;' }}</template>

      <template v-slot:item.webhook="{ item }">{{ item.webhook || '&mdash;' }}</template>

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
import axios from 'axios';

export default {
  mixins: [ItemListPageBase],

  components: {
    RunnerForm,
    YesNoDialog,
    EditDialog,
  },

  methods: {
    async setActive(runnerId, active) {
      await axios({
        method: 'post',
        url: `/api/runners/${runnerId}/active`,
        responseType: 'json',
        data: {
          active,
        },
      });
    },

    getHeaders() {
      return [
        {
          value: 'active',
        }, {
          text: this.$i18n.t('name'),
          value: 'name',
          width: '50%',
        },
        {
          text: this.$i18n.t('webhook'),
          value: 'webhook',
        },
        {
          text: this.$i18n.t('max_parallel_tasks'),
          value: 'max_parallel_tasks',
        }, {
          text: this.$i18n.t('actions'),
          value: 'actions',
          sortable: false,
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
