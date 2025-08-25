<template>
  <div v-if="inventories != null && states != null">

    <EditDialog
      v-model="editDialog"
      :save-button-text="aliasId === 'new' ? $t('create') : $t('save')"
      :title="`${aliasId === 'new' ? $t('nnew') : $t('edit')} Key`"
      :max-width="450"
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
      @yes="deleteInventory()"
    />

    <EditDialog
      v-model="inventoryDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :icon="getAppIcon(template.app)"
      :icon-color="getAppColor(template.app)"
      :title="`${$t('nnew')} ${APP_INVENTORY_TITLE[template.app]}`"
      :max-width="450"
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
          dense
          v-model="inventoryId"
          v-if="inventories.length > 0"
          mandatory
          style="overflow: auto"
          class="mb-2"
        >
          <v-btn
            v-for="inv in inventories"
            :key="inv.id"
            :value="inv.id"
            style="border-radius: 0"
          >
            {{ inv.inventory }}
            <v-icon
              dark
              v-if="inv.id === template.inventory_id"
              class="ml-1"
              color="success"
            >mdi-check
            </v-icon>
          </v-btn>
        </v-btn-toggle>

        <span v-else>No workspaces.</span>

        <v-menu offset-y>
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              color="primary"
              dark
              fab
              small
              class="ml-3"
              v-bind="attrs"
              v-on="on"
            >
              <v-icon dark>mdi-plus</v-icon>
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="itemId = 'new'; inventoryDialog = true">
              <v-list-item-icon>
                <v-icon>mdi-pencil</v-icon>
              </v-list-item-icon>
              <v-list-item-title>New workspace</v-list-item-title>
            </v-list-item>
            <v-list-item @click="attachInventoryDialog = true">
              <v-list-item-icon>
                <v-icon>mdi-connection</v-icon>
              </v-list-item-icon>
              <v-list-item-title>Attach existing workspace</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </div>

      <v-card
        style="background-color: var(--highlighted-card-bg-color);"
        v-if="inventories.length > 0"
      >

        <v-card-title
          class="d-flex justify-space-between align-center"
          style="font-weight: normal;"
        >

          <span>Workspace: <strong>{{ inventory.inventory }}</strong></span>

          <div>

            <v-btn
              class="mr-4"
              :disabled="inventoryId === template.inventory_id"
              color="success"
              @click="setDefaultInventory()"
            >
              Make default
            </v-btn>

            <v-btn
              class="mr-4"
              color="primary"
              :disabled="inventoryId === template.inventory_id"
              @click="detachInventory()"
            >
              Detach
            </v-btn>

            <v-btn
              icon
              class="mr-2"
              color="error"
              :disabled="inventoryId === template.inventory_id"
              @click="deleteInventoryDialog = true;"
            >
              <v-icon>mdi-delete</v-icon>
            </v-btn>

            <v-btn
              icon
              @click="itemId = inventoryId; inventoryDialog = true;"
            >
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
          </div>

        </v-card-title>

        <v-divider />

        <v-alert
          type="info"
          text
          color="hsl(348deg, 86%, 61%)"
          style="border-radius: 0;"
          v-if="!premiumFeatures.terraform_backend"
        >
            <span class="mr-2">
              Terraform/OpenTofu HTTP backend available only in <b>PRO</b> version.
            </span>
          <v-btn
            color="hsl(348deg, 86%, 61%)"
            href="https://semaphoreui.com/pro#runners"
          >
            Learn more
            <v-icon>mdi-chevron-right</v-icon>
          </v-btn>
        </v-alert>
        <v-card-text>

          <h3>Aliases</h3>
          <div class="mb-6">Unique endpoints (aliases) to access your Terraform HTTP backend.</div>

          <div v-for="alias of (aliases || [])" :key="alias.id">
            <code class="mr-2">{{ alias.url }}</code>
            <CopyClipboardButton
              :text="alias.url"
              :success-message="$t('aliasUrlCopied')"
            />
            <v-btn icon @click="editAlias(alias.id)">
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
            <v-btn icon @click="deleteAlias(alias.id)">
              <v-icon>mdi-delete</v-icon>
            </v-btn>
          </div>

          <v-btn
            color="primary"
            @click="addAlias()"
            :disabled="!premiumFeatures.terraform_backend"
            class="mb-8"
          >
            {{ aliases == null ? $t('LoadAlias') : $t('AddAlias') }}
          </v-btn>

          <v-divider class="mb-4" />

          <h3>State history</h3>
          <div class="mb-6">Chronological history of the workspace state changes.</div>

          <v-data-table
            style="
              background: transparent;
            "
            v-if="premiumFeatures.terraform_backend"
            :headers="headers"
            :items="states"
            :footer-props="{ itemsPerPageOptions: [20] }"
            single-expand
            show-expand
            class="mt-4 TaskListTable TaskListTable TerraformStateTable"
          >
            <template v-slot:item.id="{ item }">
              #{{ item.id }}
            </template>

            <template v-slot:item.task_id="{ item }">
              <TaskLink
                v-if="item.task_id"
                :task-id="item.task_id"
                :label="'#' + item.task_id"
              />
              <div v-else>&mdash;</div>
            </template>

            <template v-slot:item.status="{ item }">
              <TaskStatus :status="item.status" />
            </template>

            <template v-slot:item.created="{ item }">
              {{ item.created | formatDate }}
            </template>

            <template v-slot:expanded-item="{ headers, item }">
              <td
                :colspan="headers.length"
                style="max-width: 400px"
              >
                <TerraformStateView
                  class="mb-1"
                  :project-id="template.project_id"
                  :inventory-id="inventoryId"
                  :state-id="item.id"
                />
              </td>
            </template>
          </v-data-table>

          <v-container v-else>
            <div style="text-align: center; color: grey;">No state available.</div>
          </v-container>

        </v-card-text>
      </v-card>

    </div>

  </div>
</template>
<style lang="scss">
.TerraformStateTable {

  .v-data-table__wrapper {
    padding-left: 0 !important;
    padding-right: 0 !important;
  }

  .v-data-footer {
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
    premiumFeatures: Object,
  },

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
      headers: [
        {
          text: 'ID',
          value: 'id',
          sortable: false,
        },
        {
          text: this.$i18n.t('taskId'),
          value: 'task_id',
          sortable: false,
        },
        {
          text: this.$i18n.t('created'),
          value: 'created',
          sortable: false,
        },
        {
          value: 'actions',
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
