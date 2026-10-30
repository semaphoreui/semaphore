<template>
  <div
    v-if="items != null"
    data-testid="schedule-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >
    <EditDialog
      v-model="editDialog"
      test-id="scheduleDialog"
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
          :type="scheduleType"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      test-id="scheduleRefsDialog"
      object-title="schedule"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      test-id="scheduleDeleteDialog"
      :title="$t('deleteSchedule')"
      :text="$t('askDeleteSchedule')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="schedule-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title class="flex-0-1" data-testid="schedule-title">
        {{ $t('schedule') }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-menu location="bottom">
        <template v-slot:activator="{ props }">
          <v-btn
            class="pr-2"
            v-bind="props"
            color="primary"
            variant="elevated"
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
            data-testid="schedule-new"
          >{{ $t('newSchedule') }}
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </template>

        <v-list data-testid="schedule-newMenu">
          <v-list-item
            link
            @click="editSchedule('new', '');"
            data-testid="schedule-new-cron"
          >
            <template v-slot:prepend>
              <v-icon>mdi-calendar-sync</v-icon>
            </template>
            <v-list-item-title>Cron</v-list-item-title>
          </v-list-item>
          <v-list-item
            link
            @click="editSchedule('new', 'run_at');"
            data-testid="schedule-new-runAt"
          >
            <template v-slot:prepend>
              <v-icon>mdi-clock-time-eight-outline</v-icon>
            </template>
            <v-list-item-title>Run once</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>

    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4 ScheduleTable"
      :items-per-page="-1"
      data-testid="schedule-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
    >
      <template v-slot:item.active="{ item }">
        <!-- min-width: the grid of a Vuetify 3 v-input is 0 px at its narrowest, and the column
             of the switch (beside the 100 % wide «Template») shrank to it: the switch covered
             the names -->
        <v-switch
          :disabled="!can(USER_PERMISSIONS.manageProjectResources)"
          v-model="item.active"
          inset
          style="min-width: max-content"
          @update:model-value="setActive(item.id, $event)"
          data-testid="schedule-rowActive"
        ></v-switch>
      </template>

      <template v-slot:item.name="{ item }">
        <div data-testid="schedule-rowName">{{ item.name || '&mdash;' }}</div>
      </template>

      <template v-slot:item.tpl_name="{ item }">
        <div class="d-flex">
          <router-link
            :to="
            '/project/' + item.project_id +
            '/templates/' + item.template_id"
            data-testid="schedule-rowTemplate"
          >{{ item.tpl_name }}
          </router-link>
        </div>
      </template>
      <template v-slot:item.cron_format="{ item }">
        <div v-if="item.type === 'run_at'" data-testid="schedule-rowRunAt">
          {{ formatRunAt(item) }}
        </div>
        <code v-else data-testid="schedule-rowCron">{{ item.cron_format }}</code>
      </template>

      <template v-slot:item.actions="{ item }">
        <v-btn-group
          density="compact"
          data-testid="schedule-rowActions"
        >
          <v-btn
            @click="askDeleteItem(item.id)"
            data-testid="schedule-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            @click="editSchedule(item.id, item.type)"
            data-testid="schedule-rowEdit"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </v-btn-group>
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
<style lang="scss">
// Not an inline style: the utility `mt-4` of Vuetify 4 has no `!important` and would lose to it.
.ScheduleTable {
  max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px);
  margin: auto;

  // The inset switch of Vuetify 2 (`.v-input--switch--inset`): the control 48 × 24 px and 8 px
  // after it, the track (48 × 28) 4 px out to the left at .32, a white thumb when off. Vuetify 4
  // makes the control 40 px high (a row 16 px higher) and starts the track at the control.
  .v-switch--inset {
    .v-selection-control__wrapper {
      height: 24px;
      margin-inline: -4px 12px;
    }

    .v-switch__track {
      opacity: .32;
    }
  }

  .v-switch--inset.v-theme--light
    .v-selection-control:not(.v-selection-control--dirty) .v-switch__thumb {
    background-color: #fff;
  }

  // Disabled (no right to manage, `.v-input--is-disabled`): an inactive switch is a faint track and
  // a light thumb, an active one is in the primary colour at .6; not the whole control at .38.
  .v-switch--inset .v-selection-control--disabled {
    opacity: 1;

    &.v-selection-control--dirty .v-selection-control__wrapper {
      opacity: .6;
    }

    // Vuetify 4 drops the colour of the track of a disabled switch.
    &.v-selection-control--dirty .v-switch__track {
      background-color: rgb(var(--v-theme-primary));
    }
  }

  .v-switch--inset.v-theme--light
    .v-selection-control--disabled:not(.v-selection-control--dirty) {
    .v-switch__track {
      background-color: rgba(0, 0, 0, .12);
    }

    .v-switch__thumb {
      background-color: #fafafa;
    }
  }

  .v-switch--inset.v-theme--dark
    .v-selection-control--disabled:not(.v-selection-control--dirty) {
    .v-switch__track {
      background-color: rgba(255, 255, 255, .1);
    }

    .v-switch__thumb {
      background-color: #424242;
    }
  }

  // Mobile rows: the value at the right end of the row (`.v-data-table__mobile-row__cell`, a
  // shrink-to-fit flex item); Vuetify 4 stretches it over the second grid column, so the switch
  // and the template link sat in the middle of the row.
  .v-data-table__tr--mobile > td > .v-data-table__td-value {
    justify-self: end;
  }
}
</style>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import ScheduleForm from '@/components/ScheduleForm.vue';
import TaskList from '@/components/TaskList.vue';
import axios from 'axios';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezonePlugin from 'dayjs/plugin/timezone';

dayjs.extend(utc);
dayjs.extend(timezonePlugin);

export default {
  components: { TaskList, ScheduleForm },
  mixins: [ItemListPageBase],
  props: {
    systemInfo: Object,
  },
  data() {
    return {
      openedItems: [],
      scheduleType: null,
    };
  },
  methods: {
    editSchedule(id, type) {
      this.scheduleType = type;
      this.editItem(id);
    },

    formatRunAt(item) {
      if (!item.run_at) {
        return '—';
      }

      const tz = this.systemInfo?.schedule_timezone || 'UTC';
      const parsed = dayjs(item.run_at).tz(tz);

      if (!parsed.isValid()) {
        return '—';
      }

      return `${parsed.format('YYYY-MM-DD HH:mm')} (${tz})`;
    },

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
        title: '',
        key: 'active',
        sortable: false,
      }, {
        title: this.$t('name'),
        key: 'name',
      }, {
        title: this.$t('schedule'),
        key: 'cron_format',
      }, {
        title: this.$t('template'),
        key: 'tpl_name',
        width: '100%',
      }, {
        title: this.$t('actions'),
        key: 'actions',
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
