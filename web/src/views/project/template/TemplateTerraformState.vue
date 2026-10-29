<template>
  <div
    v-if="inventories != null && states != null"
    class="mt-8"
    data-testid="templateState-page"
    :data-state="inventories.length === 0 ? 'empty' : 'ready'"
  >

    <EditDialog
      v-model="editDialog"
      :save-button-text="aliasId === 'new' ? $t('create') : $t('save')"
      :title="`${aliasId === 'new' ? $t('nnew') : $t('edit')} Key`"
      :max-width="450"
      test-id="tfAliasDialog"
      @save="loadAliases()"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TerraformAliasForm
          :project-id="template.project_id"
          :item-id="aliasId"
          :inventory-id="inventoryId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <YesNoDialog
      :title="$t('deleteInventory')"
      :text="$t('askDeleteInv')"
      v-model="deleteInventoryDialog"
      test-id="tfWorkspaceDeleteDialog"
      @yes="deleteInventory()"
    />

    <EditDialog
      v-model="inventoryDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :icon="getAppIcon(template.app)"
      :icon-color="getAppColor(template.app)"
      :title="`${$t('nnew')} ${APP_INVENTORY_TITLE[template.app]}`"
      :max-width="450"
      test-id="tfWorkspaceDialog"
      @save="onNewInventory"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TerraformInventoryForm
          :template-id="template.id"
          :project-id="template.project_id"
          :name-prefix="`${template.name} - `"
          :item-id="itemId"
          :type="`${template.app}-workspace`"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <EditDialog
      v-model="attachInventoryDialog"
      :save-button-text="$t('Attach')"
      :icon="getAppIcon(template.app)"
      :icon-color="getAppColor(template.app)"
      :max-width="450"
      title="Choose workspace to attach"
      test-id="tfWorkspaceAttachDialog"
      @save="attachInventory($event.itemId)"
    >
      <template v-slot:form="{ onSave, needSave, needReset }">
        <InventorySelectForm
          :app="template.app"
          :project-id="template.project_id"
          @save="onSave"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <div
      class="px-4 py-3 CenterToScreen"
      style="max-width: 1000px; margin: auto;"
    >
      <div class="mb-6">
        <v-btn-toggle
          density="compact"
          v-model="inventoryId"
          v-if="inventories.length > 0"
          mandatory="force"
          style="overflow: auto"
          class="mb-2"
          data-testid="templateState-workspaces"
        >
          <v-btn
            v-for="inv in inventories"
            :key="inv.id"
            :value="inv.id"
            style="border-radius: 0"
            data-testid="templateState-workspace"
          >
            {{ inv.inventory }}
            <v-icon
              v-if="inv.id === template.inventory_id"
              class="ml-1"
              color="success"
              data-testid="templateState-workspaceDefault"
            >mdi-check
            </v-icon>
          </v-btn>
        </v-btn-toggle>

        <span v-else data-testid="templateState-empty">No workspaces.</span>

        <v-menu location="bottom">
          <template v-slot:activator="{ props }">
            <v-btn
              color="primary"
              icon
              size="small"
              class="ml-3"
              v-bind="props"
              data-testid="templateState-new"
            >
              <v-icon>mdi-plus</v-icon>
            </v-btn>
          </template>
          <v-list>
            <v-list-item
              @click="itemId = 'new'; inventoryDialog = true"
              data-testid="templateState-new-workspace"
            >
              <template v-slot:prepend>
                <v-icon>mdi-pencil</v-icon>
              </template>
              <v-list-item-title>New workspace</v-list-item-title>
            </v-list-item>
            <v-list-item
              @click="attachInventoryDialog = true"
              data-testid="templateState-new-attach"
            >
              <template v-slot:prepend>
                <v-icon>mdi-connection</v-icon>
              </template>
              <v-list-item-title>Attach existing workspace</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </div>

      <v-card
        style="background-color: var(--highlighted-card-bg-color);"
        v-if="inventories.length > 0"
        data-testid="templateState-workspaceCard"
      >

        <!-- flex-wrap, text-wrap: v-card__title of Vuetify 2 wrapped on narrow screens -->
        <v-card-title
          class="d-flex flex-wrap text-wrap justify-space-between align-center"
          style="font-weight: normal;"
        >

          <span data-testid="templateState-workspaceTitle">Workspace: <strong
            data-testid="templateState-workspaceName"
          >{{ inventory.inventory }}</strong></span>

          <div>

            <v-btn
              class="mr-4"
              :disabled="inventoryId === template.inventory_id"
              color="success"
              @click="setDefaultInventory()"
              data-testid="templateState-makeDefault"
            >
              Make default
            </v-btn>

            <v-btn
              class="mr-4"
              color="primary"
              :disabled="inventoryId === template.inventory_id"
              @click="detachInventory()"
              data-testid="templateState-detach"
            >
              Detach
            </v-btn>

            <v-btn
              icon
              class="mr-2"
              color="error"
              :disabled="inventoryId === template.inventory_id"
              @click="deleteInventoryDialog = true;"
              data-testid="templateState-delete"
            >
              <v-icon>mdi-delete</v-icon>
            </v-btn>

            <v-btn
              icon
              @click="itemId = inventoryId; inventoryDialog = true;"
              data-testid="templateState-edit"
            >
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
          </div>

        </v-card-title>

        <v-divider />

        <v-alert
          type="info"
          variant="tonal"
          color="hsl(348deg, 86%, 61%)"
          style="border-radius: 0;"
          v-if="!features.terraform_backend"
          data-testid="templateState-proAlert"
        >
            <span class="mr-2">
              Terraform/OpenTofu HTTP backend available only in <b>PRO</b> version.
            </span>
          <v-btn
            color="hsl(348deg, 86%, 61%)"
            href="https://semaphoreui.com/pro#runners"
            data-testid="templateState-learnMore"
          >
            Learn more
            <v-icon>mdi-chevron-right</v-icon>
          </v-btn>
        </v-alert>
        <v-card-text>

          <h3 data-testid="templateState-aliasesTitle">Aliases</h3>
          <div
            class="mb-6"
            data-testid="templateState-aliasesHint"
          >Unique endpoints (aliases) to access your Terraform HTTP backend.</div>

          <div
            v-for="alias of (aliases || [])"
            :key="alias.id"
            data-testid="templateState-alias"
          >
            <code class="mr-2" data-testid="templateState-aliasUrl">{{ alias.url }}</code>
            <CopyClipboardButton
              :text="alias.url"
              :success-message="$t('aliasUrlCopied')"
              test-id="templateState-aliasCopy"
            />
            <v-btn icon @click="editAlias(alias.id)" data-testid="templateState-aliasEdit">
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
            <v-btn icon @click="deleteAlias(alias.id)" data-testid="templateState-aliasDelete">
              <v-icon>mdi-delete</v-icon>
            </v-btn>
          </div>

          <v-btn
            color="primary"
            @click="addAlias()"
            :disabled="!features.terraform_backend"
            class="mb-8"
            data-testid="templateState-addAlias"
            :data-state="aliases == null ? 'error' : (aliases.length === 0 ? 'empty' : 'ready')"
          >
            {{ aliases == null ? $t('LoadAlias') : $t('AddAlias') }}
          </v-btn>

          <v-divider class="mb-4" />

          <h3 data-testid="templateState-historyTitle">State history</h3>
          <div
            class="mb-6"
            data-testid="templateState-historyHint"
          >Chronological history of the workspace state changes.</div>

          <v-data-table
            style="
              background: transparent;
            "
            v-if="features.terraform_backend"
            :headers="headers"
            :items="states"
            :items-per-page-options="[20]"
            :items-per-page="20"
            show-expand
            :expanded="expanded"
            @update:expanded="expanded = $event.slice(-1)"
            class="mt-4 TaskListTable TaskListTable TerraformStateTable FixedPageSizeTable"
            data-testid="templateState-table"
            :data-state="states.length === 0 ? 'empty' : 'ready'"
          >
            <template v-slot:item.id="{ item }">
              #{{ item.id }}
            </template>

            <template v-slot:item.task_id="{ item }">
              <TaskLink
                v-if="item.task_id"
                :task-id="item.task_id"
                :label="'#' + item.task_id"
                data-testid="templateState-rowTask"
              />
              <div v-else data-testid="templateState-rowNoTask">&mdash;</div>
            </template>

            <template v-slot:item.status="{ item }">
              <TaskStatus :status="item.status" />
            </template>

            <template v-slot:item.created="{ item }">
              {{ $filters.formatDate(item.created) }}
            </template>

            <!-- the slot draws the whole row: Vuetify 3 does not wrap it in a <tr> -->
            <template v-slot:expanded-row="{ columns, item }">
              <tr>
                <td
                  :colspan="columns.length"
                  style="max-width: 400px"
                  data-testid="templateState-rowState"
                >
                  <TerraformStateView
                    class="mb-1"
                    :project-id="template.project_id"
                    :inventory-id="inventoryId"
                    :state-id="item.id"
                  />
                </td>
              </tr>
            </template>
          </v-data-table>

          <v-container v-else data-testid="templateState-noBackend">
            <div style="text-align: center; color: grey;">No state available.</div>
          </v-container>

        </v-card-text>
      </v-card>

    </div>

  </div>
</template>
<style lang="scss">
.TerraformStateTable {

  .v-table__wrapper {
    padding-left: 0 !important;
    padding-right: 0 !important;
  }

  .v-data-table-footer {
    margin-left: 0 !important;
    margin-right: 0 !important;
  }
}
</style>
<script>

import axios from 'axios';
import TerraformInventoryForm from '@/components/TerraformInventoryForm.vue';
import EditDialog from '@/components/EditDialog.vue';
import { APP_INVENTORY_TITLE } from '@/lib/constants';
import AppsMixin from '@/components/AppsMixin';
import YesNoDialog from '@/components/YesNoDialog.vue';
import InventorySelectForm from '@/components/InventorySelectForm.vue';
import TerraformAliasForm from '@/components/TerraformAliasForm.vue';
import TaskStatus from '@/components/TaskStatus.vue';
import TaskLink from '@/components/TaskLink.vue';
import TerraformStateView from '@/components/TerraformStateView.vue';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';

export default {
  mixins: [AppsMixin],
  computed: {
    APP_INVENTORY_TITLE() {
      return APP_INVENTORY_TITLE;
    },
  },

  components: {
    CopyClipboardButton,
    TerraformStateView,
    TaskLink,
    TaskStatus,
    TerraformAliasForm,
    InventorySelectForm,
    YesNoDialog,
    EditDialog,
    TerraformInventoryForm,
  },

  props: {
    template: Object,
    features: Object,
  },

  emits: ['update-template'],

  watch: {
    async inventoryId() {
      this.inventory = (this.inventories || []).find((inv) => inv.id === this.inventoryId);
      await this.loadAliases();
      await this.loadStates();
    },
  },

  data() {
    return {
      aliases: [],
      states: null,
      inventory: null,
      inventories: null,
      inventoryDialog: null,
      inventoryId: null,
      deleteInventoryDialog: null,
      attachInventoryDialog: null,
      aliasId: null,
      editDialog: null,
      // single-expand of Vuetify 2: only the last expanded row stays open
      expanded: [],
      headers: [
        // the expand column first, as Vuetify 2 drew it (Vuetify 3 adds it last)
        {
          key: 'data-table-expand',
        },
        {
          title: 'ID',
          key: 'id',
          sortable: false,
        },
        {
          title: this.$t('taskId'),
          key: 'task_id',
          sortable: false,
        },
        {
          title: this.$t('created'),
          key: 'created',
          sortable: false,
        },
        {
          key: 'actions',
          sortable: false,
          width: '0%',
        },
      ],
      itemId: null,
    };
  },

  async created() {
    this.inventoryId = this.template.inventory_id;
    await this.loadInventories();
    this.inventory = this.inventories.find((inv) => inv.id === this.inventoryId);
    await this.loadAliases();
    await this.loadStates();
  },

  methods: {

    async setDefaultInventory() {
      await axios({
        method: 'post',
        url: `/api/project/${this.template.project_id}/templates/${this.template.id}/inventory/${this.inventoryId}/set_default`,
      });
      this.$emit('update-template', {});
    },

    async detachInventory() {
      await axios({
        method: 'post',
        url: `/api/project/${this.template.project_id}/templates/${this.template.id}/inventory/${this.inventoryId}/detach`,
      });
      await this.loadInventories();
    },

    async deleteInventory() {
      await axios({
        method: 'delete',
        url: `/api/project/${this.template.project_id}/inventory/${this.inventoryId}`,
      });
      await this.loadInventories();
    },

    async onNewInventory(e) {
      await this.loadInventories();
      this.inventoryId = e.item.id;
    },

    async attachInventory(inventoryId) {
      await axios({
        method: 'post',
        url: `/api/project/${this.template.project_id}/templates/${this.template.id}/inventory/${inventoryId}/attach`,
      });
      await this.loadInventories();
      this.inventoryId = inventoryId;
    },

    async loadStates() {
      if (!this.inventoryId) {
        this.states = [];
        return;
      }
      this.states = (await axios.get(`/api/project/${this.template.project_id}/inventory/${this.inventoryId}/terraform/states`)).data;
    },

    async loadInventories() {
      this.inventories = (await axios({
        url: `/api/project/${this.template.project_id}/inventory?template_id=${this.template.id}`,
        responseType: 'json',
      })).data;

      if (
        (this.inventoryId == null
          || !this.inventories.some((inv) => inv.id === this.inventoryId))
        && this.inventories.length > 0) {
        this.inventoryId = this.inventories[0].id;
      }
    },

    async loadAliases() {
      if (!this.inventoryId) {
        this.aliases = [];
        return;
      }
      try {
        this.aliases = (await axios({
          url: `/api/project/${this.template.project_id}/inventory/${this.inventoryId}/terraform/aliases`,
          responseType: 'json',
        })).data;
      } catch {
        this.aliases = null;
      }
    },

    async deleteAlias(alias) {
      await axios({
        method: 'delete',
        url: `/api/project/${this.template.project_id}/inventory/${this.inventoryId}/terraform/aliases/${alias}`,
      });
      await this.loadAliases();
    },

    editAlias(alias) {
      this.aliasId = alias;
      this.editDialog = true;
    },

    async addAlias() {
      this.aliasId = 'new';
      this.editDialog = true;
    },
  },
};
</script>
