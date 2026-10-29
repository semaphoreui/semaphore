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
        @change="applyIntegration($event)"
      ></v-select>

      <v-text-field
        v-model="alias"
        :label="$t('alias')"
        append-icon="mdi-magnify"
        clearable
        dense
        hide-details
        style="max-width: 260px"
        @keyup.enter="applyAlias(alias)"
        @click:append="applyAlias(alias)"
        @click:clear="applyAlias('')"
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
      @click:row="openCall"
    >
      <template v-slot:item.created="{ item }">
        {{ item.created | formatDate }}
      </template>

      <!-- Links inside a row stop the click, or it would also open the drawer. -->
      <template v-slot:item.integration_name="{ item }">
        <span @click.stop>
          <router-link
            v-if="item.integration_id"
            :to="`/project/${projectId}/integrations/${item.integration_id}`"
          >{{ item.integration_name }}
          </router-link>
          <span v-else class="text--secondary">{{ item.integration_name }}</span>
        </span>

        <div style="font-size: 14px">
          <a @click.prevent.stop="applyAlias(item.alias)" :title="$t('alias')">
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
        <span @click.stop>
          <router-link
            v-if="item.run_id && item.workflow_id"
            :to="`/project/${projectId}/workflows/${item.workflow_id}/runs/${item.run_id}`"
          >{{ $t('run') }} #{{ item.run_id }}
          </router-link>
          <span v-else-if="item.run_id">{{ $t('run') }} #{{ item.run_id }}</span>
          <TaskLink v-else-if="item.task_id" :label="'#' + item.task_id" :task-id="item.task_id" />
          <span v-else>&mdash;</span>
        </span>
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

    <v-navigation-drawer
      :value="selected !== null"
      right
      fixed
      temporary
      width="480"
      @input="(open) => { if (!open) selected = null; }"
    >
      <IntegrationCallCard
        v-if="selected"
        :call="selected"
        :project-id="projectId"
        :loading="selectedLoading"
        @close="selected = null"
      />
    </v-navigation-drawer>
  </div>
</template>

<style lang="scss">
.IntegrationCallsTable td {
  height: 60px !important;
}

.IntegrationCallsTable tbody tr {
  cursor: pointer;
}
</style>

<script>
import axios from 'axios';
import ItemListPageBase from '@/components/ItemListPageBase';
import TaskLink from '@/components/TaskLink.vue';
import IntegrationCallCard from '@/components/IntegrationCallCard.vue';

const OUTCOME_COLORS = {
  ok: 'success',
  unauthorized: 'warning',
  skipped: 'grey',
  error: 'error',
};

export default {
  mixins: [ItemListPageBase],

  components: { TaskLink, IntegrationCallCard },

  data() {
    return {
      OUTCOME_COLORS,
      integrations: [],
      // What the controls show. Separate from the applied values below: a
      // filter the server has not answered yet must not change what the rows
      // on screen mean.
      integrationId: null,
      alias: '',
      // What the rows on screen are actually filtered and paged by. `load`
      // commits these together with the rows, and only once a response has
      // arrived, so the table, the filters it was built from and the cursor can
      // never disagree — and a failed or superseded request needs no rollback
      // because it changed nothing.
      appliedIntegrationId: null,
      appliedAlias: '',
      appliedPerPage: 20,
      cursors: [null],
      pageIndex: 0,
      loading: false,
      // Changing a filter or the project starts a load without waiting for the
      // one in flight, so each request carries a sequence number and only the
      // newest applies. The two lists are fetched independently and so count
      // independently.
      requestSeq: 0,
      integrationsSeq: 0,
      // What the request in flight asks for, so a filter changed before it
      // answers builds on it instead of on the older applied state.
      pending: null,
      // `cursors[i]` is the `before` call id of page i, null for the newest
      // page; `hasNext` comes from the X-Has-Next header.
      hasNext: false,
      perPage: 20,
      perPageOptions: [10, 20, 50, 100],
      // The call open in the drawer.
      selected: null,
      selectedLoading: false,
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
      return this.pageIndex * this.appliedPerPage;
    },

    rangeEnd() {
      return this.rangeStart + (this.items ? this.items.length : 0);
    },
  },

  watch: {
    // The component is reused across projects, so created() does not run again:
    // both filters and the integration options still belong to the project that
    // was open before and have to be cleared here.
    async projectId() {
      this.integrations = [];
      this.selected = null;
      this.integrationId = null;
      this.alias = '';
      this.pending = null;
      this.appliedIntegrationId = null;
      this.appliedAlias = '';
      this.cursors = [null];
      this.pageIndex = 0;
      // Not null: that unmounts the page through the v-if and flashes it blank.
      this.items = [];
      this.hasNext = false;

      // Independent, so a failure to list the integrations cannot leave the
      // previous project's calls on screen.
      await Promise.all([this.load(), this.beforeLoadItems()]);
    },

    async perPage() {
      await this.load({ appliedPerPage: this.perPage, cursors: [null], pageIndex: 0 });
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
      this.integrationsSeq += 1;
      const seq = this.integrationsSeq;

      const integrations = await this.loadProjectResources('integrations');

      // Switching A -> B -> A gives overlapping requests with the same project
      // id, so the sequence number decides, not the id.
      if (seq === this.integrationsSeq) {
        this.integrations = integrations;
      }
    },

    // Opens with the row from the list at once; its body comes with the single
    // call. A call is immutable, so its id alone says whether a response still
    // belongs to the call on screen when the user clicks through rows quickly.
    async openCall(row) {
      this.selected = row;
      this.selectedLoading = true;

      try {
        const call = await this.loadProjectEndpoint(`/integrations/calls/${row.id}`);
        if (this.selected && this.selected.id === row.id) {
          this.selected = call;
        }
      } finally {
        if (this.selected && this.selected.id === row.id) {
          this.selectedLoading = false;
        }
      }
    },

    async applyAlias(alias) {
      this.alias = alias;
      await this.load({ appliedAlias: alias, cursors: [null], pageIndex: 0 });
    },

    async applyIntegration(integrationId) {
      await this.load({ appliedIntegrationId: integrationId, cursors: [null], pageIndex: 0 });
    },

    // Overrides ItemListPageBase.loadItems; ItemListPageBase.created calls it.
    async loadItems() {
      await this.load();
    },

    // Loads the page described by `target` — the applied filters and the cursor
    // it should be built from, defaulting to the ones on screen. They are
    // committed with the rows and only on success, so a request that fails or
    // is overtaken by a newer one leaves the page exactly as it was.
    async load(target = {}) {
      const next = {
        appliedIntegrationId: this.appliedIntegrationId,
        appliedAlias: this.appliedAlias,
        appliedPerPage: this.appliedPerPage,
        cursors: this.cursors,
        pageIndex: this.pageIndex,
        ...this.pending,
        ...target,
      };

      const params = new URLSearchParams({ count: next.appliedPerPage });

      const before = next.cursors[next.pageIndex];
      if (before != null) {
        params.set('before', before);
      }

      if (next.appliedIntegrationId != null) {
        params.set('integration_id', next.appliedIntegrationId);
      }

      if (next.appliedAlias) {
        params.set('alias', next.appliedAlias);
      }

      this.requestSeq += 1;
      const seq = this.requestSeq;

      this.pending = next;
      this.loading = true;
      try {
        const response = await axios({
          method: 'get',
          url: `${this.getItemsUrl()}?${params.toString()}`,
          responseType: 'json',
        });

        if (seq !== this.requestSeq) {
          return;
        }

        Object.assign(this, next);
        this.items = response.data;
        this.hasNext = response.headers['x-has-next'] === 'true';
      } finally {
        if (seq === this.requestSeq) {
          this.pending = null;
          this.loading = false;
        }
      }
    },

    async goNext() {
      if (!this.hasNext || this.loading || this.items.length === 0) {
        return;
      }

      // Calls are ordered by id desc, so the last row holds the smallest id on
      // this page — the cursor for the next, older page.
      const kept = this.cursors.slice(0, this.pageIndex + 1);

      await this.load({
        cursors: [...kept, this.items[this.items.length - 1].id],
        pageIndex: this.pageIndex + 1,
      });
    },

    async goPrev() {
      if (this.pageIndex === 0 || this.loading) {
        return;
      }

      await this.load({ pageIndex: this.pageIndex - 1 });
    },
  },
};
</script>
