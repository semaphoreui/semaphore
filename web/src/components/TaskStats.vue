<template>
  <v-card
    style="background: rgba(133, 133, 133, 0.06)"
    class="mx-4"
    :data-state="stats == null
      ? 'loading'
      : (stats.every((p) => !Object.values(p.count_by_status || {}).some((n) => n > 0))
        ? 'empty'
        : 'ready')"
  >
    <v-card-title
      class="task-stats__title d-flex flex-wrap align-center"
      data-testid="taskStats-title"
    >
      Task Status
      <v-spacer />
      <v-select
        hide-details
        density="compact"
        :items="dateRanges"
        item-title="text"
        class="mr-6"
        style="width: 200px; max-width: 200px"
        :menu-props="menuProps"
        v-model="dateRange"
        data-testid="taskStats-period"
      />

      <v-select
        hide-details
        density="compact"
        :items="users"
        item-title="text"
        style="width: 200px; max-width: 200px"
        :menu-props="menuProps"
        v-model="user"
        data-testid="taskStats-user"
      />
    </v-card-title>
    <v-card-text>
      <LineChart :source-data="stats" data-testid="taskStats-chart"/>
    </v-card-text>
  </v-card>
</template>
<style lang="scss">
// Vuetify 2 look: the lists take the weight of the title (`.v-input` of Vuetify 2 sets none,
// Vuetify 4 sets 400). Their width of 200 px (template) is also their size in the line, as the
// wider content of a Vuetify 2 select was: on a phone each list goes to its own line. A dense
// select of Vuetify 2 takes 30 px of the line (4 px margin and a 26 px field), the compact field of
// Vuetify 4 32 px: 2 px above it go into the line, the underline stays where it was.
.task-stats__title > .v-input {
  font-weight: inherit;
  margin-top: -2px;

  // The arrow in the middle of the 26 px field of Vuetify 2, 2 px above the text bottom.
  .v-field__append-inner {
    padding-top: 6px;
  }
}
</style>
<script>
import axios from 'axios';
import LineChart from '@/components/LineChart.vue';

export default {
  components: { LineChart },

  props: {
    templateId: Number,
    projectId: Number,
  },

  data() {
    return {
      // The menu over the field (default of VSelect, plugins/vuetify.js) 6 px lower, where the
      // menu of the dense Vuetify 2 select opened: the compact field of Vuetify 4 is 32 px high and
      // starts 6 px above the 26 px field of Vuetify 2 (style below), the menu opens from its top.
      menuProps: { location: 'top', origin: 'top', offset: -6 },
      dateRanges: [{
        text: 'Past week',
        value: 'last_week',
      }, {
        text: 'Past month',
        value: 'last_month',
      }, {
        text: 'Past year',
        value: 'last_year',
      }],
      users: [{
        text: 'All users',
        value: null,
      }],
      user: null,
      stats: null,
      dateRange: 'last_week',
    };
  },

  computed: {
    startDate() {
      const date = new Date();

      switch (this.dateRange) {
        case 'last_year':
          date.setFullYear(date.getFullYear() - 1);
          break;
        case 'last_month':
          date.setDate(date.getDate() - 30);
          break;
        case 'last_week':
        default:
          date.setDate(date.getDate() - 7);
          break;
      }

      return date.toISOString().split('T')[0];
    },
  },

  watch: {
    async startDate() {
      await this.refreshData();
    },
    async user() {
      await this.refreshData();
    },
  },

  async created() {
    await this.refreshData();

    this.users = [{
      text: 'All users',
      value: null,
    }, ...(await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/users`,
      responseType: 'json',
    })).data.map((x) => ({
      value: x.id,
      text: x.name,
    }))];
  },

  methods: {
    async refreshData() {
      let url;

      if (this.templateId) {
        url = `/api/project/${this.projectId}/templates/${this.templateId}/stats?start=${this.startDate}`;
      } else {
        url = `/api/project/${this.projectId}/stats?start=${this.startDate}`;
      }

      if (this.user) {
        url += `&user_id=${this.user}`;
      }

      this.stats = (await axios({
        method: 'get',
        url,
        responseType: 'json',
      })).data;

      const firstPoint = this.stats[0];

      if (!firstPoint || firstPoint.date > this.startDate) {
        this.stats.unshift({
          date: this.startDate,
          count_by_status: {
            success: 0,
            failed: 0,
            stopped: 0,
          },
        });
      }

      const lastPoint = this.stats[this.stats.length - 1];

      if (lastPoint.date < new Date().toISOString().split('T')[0]) {
        this.stats.push({
          date: new Date().toISOString().split('T')[0],
          count_by_status: {
            success: 0,
            failed: 0,
            stopped: 0,
          },
        });
      }
    },
  },
};
</script>
