<template>
  <v-container fluid class="pb-0 mt-8 TemplateDetails" data-testid="template-details">
    <v-row class="mb-5">
      <v-col>
        <v-list>
          <v-list-item>
            <template v-slot:prepend>
              <v-icon>mdi-book-play</v-icon>
            </template>

              <v-list-item-title>{{ $t('playbook') }}</v-list-item-title>
              <v-list-item-subtitle
                data-testid="template-detailsPlaybook"
              >{{ template.playbook }}</v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list>
          <v-list-item>
            <template v-slot:prepend>
              <v-icon>{{ TEMPLATE_TYPE_ICONS[template.type] }}</v-icon>
            </template>

              <v-list-item-title>{{ $t('type') }}</v-list-item-title>
              <v-list-item-subtitle
                data-testid="template-detailsType"
              >{{ $t(TEMPLATE_TYPE_TITLES[template.type]) }}
              </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list>
          <v-list-item>
            <template v-slot:prepend>
              <v-icon>mdi-monitor</v-icon>
            </template>

              <v-list-item-title>{{ $t('inventory') }}</v-list-item-title>
              <v-list-item-subtitle data-testid="template-detailsInventory">
                {{ (inventory.find((x) => x.id === template.inventory_id) || {name: '—'}).name }}
              </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list>
          <v-list-item>
            <template v-slot:prepend>
              <v-icon>mdi-code-braces</v-icon>
            </template>
              <v-list-item-title>{{ $t('environment') }}</v-list-item-title>
              <v-list-item-subtitle data-testid="template-detailsEnvironment">
                {{ environmentNames }}
              </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-col>
      <v-col>
        <v-list>
          <v-list-item>
            <template v-slot:prepend>
              <v-icon>mdi-git</v-icon>
            </template>
              <v-list-item-title>{{ $t('repository2') }}</v-list-item-title>
              <v-list-item-subtitle data-testid="template-detailsRepository">
                {{ repositories.find((x) => x.id === template.repository_id).name }}
              </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-col>
    </v-row>

    <TaskStats
      :project-id="template.project_id"
      :template-id="template.id"
      data-testid="template-stats"
    />

  </v-container>

</template>
<style lang="scss">
// The property lists of Vuetify 2: `v-list subheader` (no padding on top), an item with the
// padding of its content (12 px above and below, 2 px between title and subtitle: 62 px high) and
// its icon at the top (`.v-list-item__icon`: `align-self: flex-start`, 16 px above),
// and a column that is not narrower than its icon, title and subtitle unless the row is — the row
// wraps instead, as the flex columns of Vuetify 2 did (Vuetify 4 lets the text shrink to 40 px).
// Under the row the 12 px padding of the Vuetify 2 column and `mb-2` — `mb-5` on the row.
.TemplateDetails > .v-row > .v-col {
  min-width: fit-content;
}

.TemplateDetails > .v-row > .v-col > .v-list {
  padding-top: 0;
}

.TemplateDetails > .v-row > .v-col > .v-list > .v-list-item {
  > .v-list-item__prepend {
    align-self: start;
  }

  > .v-list-item__content {
    padding: 12px 0;

    > .v-list-item-title {
      margin-bottom: 2px;
    }
  }
}
</style>
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
    environmentNames() {
      if (!this.template || !Array.isArray(this.template.environment_ids)
          || this.template.environment_ids.length === 0) {
        return '—';
      }
      const names = this.template.environment_ids
        .map((id) => {
          const env = this.environment.find((x) => x.id === id);
          return env ? env.name : null;
        })
        .filter((name) => name != null);
      return names.length > 0 ? names.join(', ') : '—';
    },
  },
};
</script>
