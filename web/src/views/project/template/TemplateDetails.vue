<template>
  <v-container fluid class="pb-0 mt-8">
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
                {{ primaryEnvironmentName }}
              </v-list-item-subtitle>
              <v-list-item-subtitle
                v-if="additionalEnvironmentNames"
                class="text--secondary"
              >
                + {{ additionalEnvironmentNames }}
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

    <TaskStats :project-id="template.project_id" :template-id="template.id" />

  </v-container>

</template>
<script>
import {
  TEMPLATE_TYPE_ACTION_TITLES,
  TEMPLATE_TYPE_ICONS,
  TEMPLATE_TYPE_TITLES,
} from '@/lib/constants';
import TaskStats from '@/components/TaskStats.vue';

export default {
  components: { TaskStats },

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
      TEMPLATE_TYPE_ICONS,
      TEMPLATE_TYPE_TITLES,
      TEMPLATE_TYPE_ACTION_TITLES,
    };
  },

  computed: {
    primaryEnvironmentName() {
      if (!this.template || this.template.environment_id == null) {
        return '—';
      }
      const env = this.environment.find((x) => x.id === this.template.environment_id);
      return env ? env.name : '—';
    },

    additionalEnvironmentNames() {
      if (!this.template || !Array.isArray(this.template.environment_ids)
          || this.template.environment_ids.length === 0) {
        return '';
      }
      const primaryId = this.template.environment_id;
      return this.template.environment_ids
        .filter((id) => id !== primaryId)
        .map((id) => {
          const env = this.environment.find((x) => x.id === id);
          return env ? env.name : null;
        })
        .filter((name) => name != null)
        .join(', ');
    },
  },
};
</script>
