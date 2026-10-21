<template>
  <div v-if="items != null">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('integrations') }}</v-toolbar-title>
    </v-toolbar>

    <v-tabs>
      <v-tab :to="`/project/${projectId}/integrations`" exact>{{ $t('integrations') }}</v-tab>
      <v-tab :to="`/project/${projectId}/integrations/history`">{{ $t('history') }}</v-tab>
    </v-tabs>

    <div class="d-flex align-center px-4 py-3">
      <v-select
        v-model="integrationId"
        :items="integrationOptions"
        :label="$t('integration')"
        class="mr-4"
        dense
        hide-details
        style="max-width: 260px"
      ></v-select>

      <v-text-field
        v-model="alias"
        :label="$t('alias')"
        append-icon="mdi-magnify"
        clearable
        dense
        hide-details
        style="max-width: 260px"
        @keyup.enter="resetAndLoad()"
        @click:append="resetAndLoad()"
        @click:clear="clearAlias()"
      ></v-text-field>
    </div>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      :items-per-page="-1"
      :loading="loading"
      :no-data-text="$t('noIntegrationCalls')"
      hide-default-footer
      class="IntegrationCallsTable"
    >
      <template v-slot:item.created="{ item }">
        {{ item.created | formatDate }}
      </template>

      <template v-slot:item.integration_name="{ item }">
        <router-link
          v-if="item.integration_id"
          :to="`/project/${projectId}/integrations/${item.integration_id}`"
        >{{ item.integration_name }}
        </router-link>
        <span v-else class="text--secondary">{{ item.integration_name }}</span>

        <div style="font-size: 14px">
          <a @click.prevent="filterByAlias(item.alias)" :title="$t('alias')">
            <code>{{ item.alias }}</code>
          </a>
        </div>
      </template>

      <template v-slot:item.outcome="{ item }">
        <v-chip :color="OUTCOME_COLORS[item.outcome]" small>{{ item.outcome }}</v-chip>

        <div v-if="item.reason" class="text--secondary" style="font-size: 14px">
          {{ item.reason }}
        </div>
      </template>

      <template v-slot:item.task_id="{ item }">
        <TaskLink v-if="item.task_id" :label="'#' + item.task_id" :task-id="item.task_id" />
        <span v-else>&mdash;</span>
      </template>

      <template v-slot:item.duration="{ item }"> {{ item.duration }} ms </template>
    </v-data-table>

    <div class="d-flex align-center justify-end mt-2 mr-2 mb-4">
      <span class="text--secondary mr-2">{{ $t('rowsPerPage') }}</span>

      <v-select
        v-model="perPage"
        :disabled="loading"
        :items="perPageOptions"
        class="mr-4"
        dense
        hide-details
        style="max-width: 72px"
      ></v-select>

      <v-btn :disabled="loading || pageIndex === 0" icon @click="goPrev()">
        <v-icon>mdi-chevron-left</v-icon>
      </v-btn>

      <span class="mx-2 text--secondary">{{ rangeStart }} - {{ rangeEnd }}</span>

      <v-btn :disabled="loading || !hasNext" icon @click="goNext()">
        <v-icon>mdi-chevron-right</v-icon>
      </v-btn>
    </div>
  </div>
</template>

<style lang="scss">
.IntegrationCallsTable td {
  height: 60px !important;
}
</style>

<script>
import axios from 'axios';
import ItemListPageBase from '@/components/ItemListPageBase';
import TaskLink from '@/components/TaskLink.vue';

const OUTCOME_COLORS = {
  ok: 'success',
  unauthorized: 'warning',
  skipped: 'grey',
  error: 'error',
};

export default {
  mixins: [ItemListPageBase],

  components: { TaskLink },

  data() {
    return {
      OUTCOME_COLORS,
      integrations: [],
      integrationId: null,
      alias: '',
      loading: false,
      // Keyset pagination: `cursors[i]` is the `before` call id of page i (null
      // for the newest page), `hasNext` comes from the X-Has-Next header.
      cursors: [null],
      pageIndex: 0,
      hasNext: false,
      perPage: 20,
      perPageOptions: [10, 20, 50, 100],
    };
  },

  computed: {
    integrationOptions() {
      return [
        { text: this.$i18n.t('allIntegrations'), value: null },
        ...this.integrations.map((i) => ({ text: i.name, value: i.id })),
      ];
    },

    rangeStart() {
      return this.pageIndex * this.perPage;
    },

    rangeEnd() {
      return this.rangeStart + (this.items ? this.items.length : 0);
    },
  },

  watch: {
    async projectId() {
      await this.resetAndLoad();
    },

    async perPage() {
      await this.resetAndLoad();
    },

    async integrationId() {
      await this.resetAndLoad();
    },
  },

  methods: {
    getHeaders() {
      return [
        { text: this.$i18n.t('time'), value: 'created', sortable: false },
        { text: this.$i18n.t('integration'), value: 'integration_name', sortable: false },
        { text: this.$i18n.t('sourceIp'), value: 'source_ip', sortable: false },
        { text: this.$i18n.t('method'), value: 'method', sortable: false },
        { text: this.$i18n.t('status'), value: 'outcome', sortable: false },
        { text: this.$i18n.t('task2'), value: 'task_id', sortable: false },
        { text: this.$i18n.t('duration'), value: 'duration', sortable: false },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations/calls`;
    },

    async beforeLoadItems() {
      this.integrations = (
        await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/integrations`,
          responseType: 'json',
        })
      ).data;
    },

    async filterByAlias(alias) {
      this.alias = alias;
      await this.resetAndLoad();
    },

    async clearAlias() {
      this.alias = '';
      await this.resetAndLoad();
    },

    // Existing cursors become meaningless when the project, the filters or the
    // page size change, so go back to the newest page.
    async resetAndLoad() {
      this.cursors = [null];
      this.pageIndex = 0;
      await this.loadItems();
    },

    // Overrides ItemListPageBase.loadItems to page with the keyset cursor.
    async loadItems() {
      const params = new URLSearchParams({ count: this.perPage });

      const before = this.cursors[this.pageIndex];
      if (before != null) {
        params.set('before', before);
      }

      if (this.integrationId != null) {
        params.set('integration_id', this.integrationId);
      }

      if (this.alias) {
        params.set('alias', this.alias);
      }

      this.loading = true;
      try {
        const response = await axios({
          method: 'get',
          url: `${this.getItemsUrl()}?${params.toString()}`,
          responseType: 'json',
        });

        this.items = response.data;
        this.hasNext = response.headers['x-has-next'] === 'true';
      } finally {
        this.loading = false;
      }
    },

    async goNext() {
      if (!this.hasNext || this.loading || this.items.length === 0) {
        return;
      }

      // Calls are ordered by id desc, so the last row holds the smallest id on
      // this page — the cursor for the next, older page.
      this.cursors = this.cursors.slice(0, this.pageIndex + 1);
      this.cursors.push(this.items[this.items.length - 1].id);
      this.pageIndex += 1;

      await this.loadItems();
    },

    async goPrev() {
      if (this.pageIndex === 0 || this.loading) {
        return;
      }

      this.pageIndex -= 1;
      await this.loadItems();
    },
  },
};
</script>
