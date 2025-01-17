<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">

    <EditDialog
      v-model="attachInventoryDialog"
      :save-button-text="$t('Attach')"
      :icon="getAppIcon(itemApp)"
      :icon-color="getAppColor(itemApp)"
      :max-width="450"
      title="Choose workspace to attach"
      @save="attachInventory($event.itemId)"
    >
      <template v-slot:form="{ onSave, needSave, needReset }">
        <TemplateSelectForm
          :app="itemApp"
          :project-id="projectId"
          @save="onSave"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <EditDialog
      expandable
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :icon="getAppIcon(itemApp)"
      :icon-color="getAppColor(itemApp)"
      :title="`${itemId === 'new' ? $t('nnew') : $t('edit')} ${APP_INVENTORY_TITLE[itemApp]}`"
      :max-width="450"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <InventoryForm
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
      object-title="inventory"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="$t('deleteInventory')"
      :text="$t('askDeleteInv')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('inventory') }}</v-toolbar-title>
      <v-spacer></v-spacer>

      <v-menu
        offset-y
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            class="pr-2"
            v-bind="attrs"
            v-on="on"
            color="primary"
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
          >{{ $t('newInventory') }}
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item
            v-for="item in apps"
            :key="item"
            link
            @click="itemApp = item; editItem('new');"
          >
            <v-list-item-icon>
              <v-icon
                :color="getAppColor(item)"
              >{{ getAppIcon(item) }}
              </v-icon>
            </v-list-item-icon>
            <v-list-item-title>{{ APP_INVENTORY_TITLE[item] }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>

    <v-divider />

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.name="{ item }">
        <v-icon class="mr-3" small>
          {{ getAppIcon(getAppByType(item.type)) }}
        </v-icon>

        {{ item.name }}
      </template>

      <template v-slot:item.type="{ item }">
        <code>{{ item.type }}</code>
      </template>
      <template v-slot:item.inventory="{ item }">
        {{ ['file', 'terraform-workspace'].includes(item.type) ? item.inventory : '&mdash;' }}
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn-toggle dense :value-comparator="() => false">
          <v-btn @click="askDeleteItem(item.id)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            @click="itemApp = getAppByType(item.type); editItem(item.id)"
            v-if="!['tofu', 'terraform'].includes(getAppByType(item.type))"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
          <v-btn
            v-else
            @click="
              itemId = item.id;
              itemApp = getAppByType(item.type);
              attachInventoryDialog = true;
            "
          >
            <v-icon>mdi-connection</v-icon>
          </v-btn>
        </v-btn-toggle>
      </template>
    </v-data-table>
  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import InventoryForm from '@/components/InventoryForm.vue';
import { APP_INVENTORY_TITLE } from '@/lib/constants';
import AppsMixin from '@/components/AppsMixin';
import axios from 'axios';
import TemplateSelectForm from '@/components/TemplateSelectForm.vue';

export default {
  computed: {
    APP_INVENTORY_TITLE() {
      return APP_INVENTORY_TITLE;
    },
  },
  mixins: [ItemListPageBase, AppsMixin],
  components: { TemplateSelectForm, InventoryForm },

  data() {
    return {
      apps: ['ansible'],
      itemApp: '',
      attachInventoryDialog: null,
    };
  },

  methods: {
    async attachInventory(templateId) {
      await axios({
        method: 'post',
        url: `/api/project/${this.projectId}/templates/${templateId}/inventory/${this.itemId}/attach`,
      });
      await this.loadItems();
    },

    getAppByType(type) {
      switch (type) {
        case 'tofu-workspace':
          return 'tofu';
        case 'terraform-workspace':
          return 'terraform';
        case '':
        case 'ansible':
          return 'ansible';
        default:
          return 'ansible';
      }
    },

    getHeaders() {
      return [{
        text: this.$i18n.t('name'),
        value: 'name',
        width: '30%',
      },
      {
        text: this.$i18n.t('type'),
        value: 'type',
        width: '20%',
      },
      {
        text: this.$i18n.t('path'),
        value: 'inventory',
        width: '50%',
      },
      {
        value: 'actions',
        sortable: false,
        width: '0%',
      },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/inventory`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/inventory/${this.itemId}`;
    },
    getEventName() {
      return 'i-inventory';
    },
  },
};
</script>
