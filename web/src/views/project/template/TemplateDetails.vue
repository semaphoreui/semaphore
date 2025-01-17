<template>
  <v-container fluid class="pb-0">
    <v-alert
      text
      dense
      class="mb-0 ml-4 mr-4 mb-6 d-inline-block"
      v-if="template.description"
    >{{ template.description }}
    </v-alert>

    <v-row class="mb-2">
      <v-col>
        <v-list subheader>
          <v-list-item>
            <v-list-item-icon>
              <v-icon>mdi-book-play</v-icon>
            </v-list-item-icon>

            <v-list-item-content>
              <v-list-item-title>{{ $t('playbook') }}</v-list-item-title>
              <v-list-item-subtitle>{{ template.playbook }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list subheader>
          <v-list-item>
            <v-list-item-icon>
              <v-icon>{{ TEMPLATE_TYPE_ICONS[template.type] }}</v-icon>
            </v-list-item-icon>

            <v-list-item-content>
              <v-list-item-title>{{ $t('type') }}</v-list-item-title>
              <v-list-item-subtitle>{{ $t(TEMPLATE_TYPE_TITLES[template.type]) }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list subheader>
          <v-list-item>
            <v-list-item-icon>
              <v-icon>mdi-monitor</v-icon>
            </v-list-item-icon>

            <v-list-item-content>
              <v-list-item-title>{{ $t('inventory') }}</v-list-item-title>
              <v-list-item-subtitle>
                {{ (inventory.find((x) => x.id === template.inventory_id) || {name: '—'}).name }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list subheader>
          <v-list-item>
            <v-list-item-icon>
              <v-icon>mdi-code-braces</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>{{ $t('environment') }}</v-list-item-title>
              <v-list-item-subtitle>
                {{ environment.find((x) => x.id === template.environment_id).name }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list subheader>
          <v-list-item>
            <v-list-item-icon>
              <v-icon>mdi-git</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>{{ $t('repository2') }}</v-list-item-title>
              <v-list-item-subtitle>
                {{ repositories.find((x) => x.id === template.repository_id).name }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-col>
    </v-row>

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

  </v-container>

</template>
<script>
import {
  TEMPLATE_TYPE_ACTION_TITLES,
  TEMPLATE_TYPE_ICONS,
  TEMPLATE_TYPE_TITLES,
} from '@/lib/constants';
import axios from 'axios';
import LineChart from '@/components/LineChart.vue';

export default {
  components: { LineChart },

  props: {
    template: Object,
    repositories: Array,
    inventory: Array,
    environment: Array,
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
      TEMPLATE_TYPE_ICONS,
      TEMPLATE_TYPE_TITLES,
      TEMPLATE_TYPE_ACTION_TITLES,
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
      url: `/api/project/${this.template.project_id}/users`,
      responseType: 'json',
    })).data.map((x) => ({
      value: x.id,
      text: x.name,
    }))];
  },

  methods: {
    async refreshData() {
      let url = `/api/project/${this.template.project_id}/templates/${this.template.id}/stats?start=${this.startDate}`;

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
