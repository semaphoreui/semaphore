<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && templates != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="itemId === 'new' ? $t('NewIntegration') : $t('EditIntegration')"
      :max-width="450"
      :transition="false"
      @save="loadItems()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <IntegrationForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="integration"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('DeleteIntegration')"
      :text="$t('DeleteIntegrationMsg')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('integrations') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        @click="editItem('new')"
      >{{ $t('NewIntegration') }}
      </v-btn>
    </v-toolbar>

    <div
      class="px-4 py-3"
    >
      <div class="mb-3 pl-1" v-if="(aliases || []).length === 0">There is no aliases.</div>

      <div v-else v-for="alias of (aliases || [])" :key="alias.id">
        <code class="mr-2">{{ alias.url }}</code>
        <v-btn icon
               @click="copyToClipboard(alias.url, $t('aliasUrlCopied'))">
          <v-icon>mdi-content-copy</v-icon>
        </v-btn>
        <v-btn icon @click="deleteAlias(alias.id)">
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>

      <v-btn color="primary" @click="addAlias()" :disabled="aliases == null">
        {{ aliases == null ? $t('LoadAlias') : $t('AddAlias') }}
      </v-btn>
    </div>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.name="{ item }">
        <router-link
          :to="`/project/${projectId}/integrations/${item.id}`"
        >{{ item.name }}
        </router-link>
      </template>
      <template v-slot:item.template_id="{ item }">
        <router-link
          :to="`/project/${projectId}/templates/${item.template_id}`"
        >
          {{ (templates.find((t) => t.id === item.template_id) || {name: '—'}).name }}
        </router-link>
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
    </v-data-table>
  </div>
</template>
<script>
import axios from 'axios';

import { USER_PERMISSIONS } from '@/lib/constants';

import ItemListPageBase from '@/components/ItemListPageBase';
import IntegrationForm from '@/components/IntegrationForm.vue';
import IntegrationsBase from '@/views/project/IntegrationsBase';
import copyToClipboard from '@/lib/copyToClipboard';

export default {
  mixins: [ItemListPageBase, IntegrationsBase],
  components: { IntegrationForm },
  data() {
    return {
      templates: null,
    };
  },

  async created() {
    this.templates = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/templates`,
      responseType: 'json',
    })).data;
  },

  methods: {
    copyToClipboard,
    allowActions() {
      return this.can(USER_PERMISSIONS.updateProject);
    },
    getHeaders() {
      return [{
        text: this.$i18n.t('name'),
        value: 'name',
        width: '40%',
        sortable: true,
      },
      {
        text: this.$i18n.t('template'),
        value: 'template_id',
        width: '60%',
        sortable: true,
      },
      {
        value: 'actions',
        sortable: false,
        width: '0%',
      }];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/integrations`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/integrations/${this.itemId}`;
    },
    getEventName() {
      return 'w-integration';
    },
  },
};
</script>
