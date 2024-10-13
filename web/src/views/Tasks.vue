<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <YesNoDialog
      :title="$t('deleteRunner')"
      :text="$t('askDeleteRunner', item.name)"
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

      <template v-slot:item.max_parallel_tasks="{ item }">
        {{ item.max_parallel_tasks || '∞' }}
      </template>

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
import EventBus from '@/event-bus';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ItemListPageBase from '@/components/ItemListPageBase';
import axios from 'axios';

export default {
  mixins: [ItemListPageBase],

  components: {
    YesNoDialog,
  },

  props: {
  },

  computed: {
  },

  data() {
    return {
      newRunnerTokenDialog: null,
      newRunner: null,
    };
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
      return [{
        text: this.$i18n.t('task_id'),
        value: 'task_id',
        width: '50%',
      }, {
        text: this.$i18n.t('username'),
        value: 'username',
        width: '50%',
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
      return '/api/tasks';
    },

    getSingleItemUrl() {
      return `/api/tasks/${this.itemId}`;
    },

    getEventName() {
      return 'i-task';
    },
  },
};
</script>
