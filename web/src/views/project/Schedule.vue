<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="$t('save')"
      :title="$t('editSchedule')"
      :max-width="500"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <ScheduleForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
          :timezone="systemInfo.schedule_timezone"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="schedule"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('deleteSchedule')"
      :text="$t('askDeleteSchedule')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('schedule') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
        v-if="can(USER_PERMISSIONS.manageProjectResources)"
      >{{ $t('newSchedule') }}
      </v-btn>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="
          max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px);
          margin: auto;
        "
    >
      <template v-slot:item.active="{ item }">
        <v-switch
          :disabled="!can(USER_PERMISSIONS.manageProjectResources)"
          v-model="item.active"
          inset
          @change="setActive(item.id, item.active)"
        ></v-switch>
      </template>

      <template v-slot:item.name="{ item }">
        <div>{{ item.name || '&mdash;' }}</div>
      </template>

      <template v-slot:item.tpl_name="{ item }">
        <div class="d-flex">
          <router-link :to="
            '/project/' + item.project_id +
            '/templates/' + item.template_id"
          >{{ item.tpl_name }}
          </router-link>
        </div>
      </template>
      <template v-slot:item.cron_format="{ item }">
        <code>{{ item.cron_format }}</code>
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

      <template v-slot:expanded-item="{ headers, item }">
        <td
          :colspan="headers.length"
          v-if="openedItems.some((template) => template.id === item.id)"
        >
          <TaskList
            style="border: 1px solid lightgray; border-radius: 6px; margin: 10px 0;"
            :template="item"
            :limit="5"
            :hide-footer="true"
          />
        </td>
      </template>
    </v-data-table>
  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import ScheduleForm from '@/components/ScheduleForm.vue';
import TaskList from '@/components/TaskList.vue';
import axios from 'axios';

export default {
  components: { TaskList, ScheduleForm },
  mixins: [ItemListPageBase],
  props: {
    systemInfo: Object,
  },
  data() {
    return {
      openedItems: [],
    };
  },
  methods: {
    async setActive(scheduleId, active) {
      await axios({
        method: 'put',
        url: `/api/project/${this.projectId}/schedules/${scheduleId}/active`,
        responseType: 'json',
        data: {
          active,
        },
      });
    },

    getHeaders() {
      return [{
        text: '',
        value: 'active',
        sortable: false,
      }, {
        text: this.$i18n.t('name'),
        value: 'name',
      }, {
        text: this.$i18n.t('Cron'),
        value: 'cron_format',
      }, {
        text: this.$i18n.t('template'),
        value: 'tpl_name',
        width: '100%',
      }, {
        text: this.$i18n.t('actions'),
        value: 'actions',
        sortable: false,
      }];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/schedules`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/schedules/${this.itemId}`;
    },
    getEventName() {
      return 'i-schedule';
    },
  },
};
</script>
