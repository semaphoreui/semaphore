<template>
  <v-card style="background: rgba(133, 133, 133, 0.06)" class="mx-4">
    <v-card-title>
      Task Status
      <v-spacer />
      <v-select
        hide-details
        dense
        :items="dateRanges"
        class="mr-6"
        style="max-width: 200px"
        v-model="dateRange"
      />

      <v-select
        hide-details
        dense
        :items="users"
        style="max-width: 200px"
        v-model="user"
      />
    </v-card-title>
    <v-card-text>
      <LineChart :source-data="stats"/>
    </v-card-text>
  </v-card>
</template>
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
