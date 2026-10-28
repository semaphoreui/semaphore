<template>
  <div
    v-if="items != null"
    data-testid="inventory-page"
    :data-state="items.length === 0 ? 'empty' : 'ready'"
  >

    <EditDialog
      v-model="attachInventoryDialog"
      test-id="inventoryAttachDialog"
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
      v-model="editDialog"
      test-id="inventoryDialog"
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
          :features="features"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      test-id="inventoryRefsDialog"
      object-title="inventory"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      test-id="inventoryDeleteDialog"
      :title="$t('deleteInventory')"
      :text="$t('askDeleteInv')"
      v-model="deleteItemDialog"
      @yes="deleteItem(itemId)"
    />

    <v-toolbar flat>
      <v-app-bar-nav-icon
        @click="showDrawer()"
        data-testid="inventory-drawerToggle"
      ></v-app-bar-nav-icon>
      <v-toolbar-title
        class="flex-0-1"
        data-testid="inventory-title"
      >{{ $t('inventory') }}</v-toolbar-title>
      <v-spacer></v-spacer>

      <v-menu
        location="bottom"
      >
        <template v-slot:activator="{ props }">
          <v-btn
            class="pr-2"
            v-bind="props"
            color="primary"
            variant="elevated"
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
            data-testid="inventory-new"
          >{{ $t('newInventory') }}
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </template>
        <v-list data-testid="inventory-newMenu">
          <v-list-item
            v-for="item in apps"
            :key="item"
            link
            @click="itemApp = item; editItem('new');"
            :data-testid="`inventory-new-${item}`"
          >
            <template v-slot:prepend>
              <v-icon
                :color="getAppColor(item)"
              >{{ getAppIcon(item) }}
              </v-icon>
            </template>
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
      :items-per-page="-1"
      :mobile-breakpoint="600"
      data-testid="inventory-table"
      :data-state="items.length === 0 ? 'empty' : 'ready'"
      style="max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 200px); margin: auto;"
    >
      <template v-slot:item.name="{ item }">
        <v-icon class="mr-3" size="small" data-testid="inventory-rowIcon">
          {{ getAppIcon(getAppByType(item.type)) }}
        </v-icon>

        {{ item.name }}
      </template>

      <template v-slot:item.type="{ item }">
        <code data-testid="inventory-rowType">{{ item.type }}</code>
      </template>
      <template v-slot:item.inventory="{ item }">
        {{ ['file', 'terraform-workspace'].includes(item.type) ? item.inventory : '&mdash;' }}
      </template>
      <template v-slot:item.actions="{ item }">
        <v-btn-group
          density="compact"
          data-testid="inventory-rowActions"
        >
          <v-btn
            @click="askDeleteItem(item.id)"
            data-testid="inventory-rowDelete"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
          <v-btn
            @click="itemApp = getAppByType(item.type); editItem(item.id)"
            v-if="!['tofu', 'terraform'].includes(getAppByType(item.type))"
            data-testid="inventory-rowEdit"
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
            data-testid="inventory-rowAttach"
          >
            <v-icon>mdi-connection</v-icon>
          </v-btn>
        </v-btn-group>
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

  props: {
    features: Object,
  },

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
        title: this.$t('name'),
        key: 'name',
        width: '30%',
      },
      {
        title: this.$t('type'),
        key: 'type',
        width: '20%',
      },
      {
        title: this.$t('path'),
        key: 'inventory',
        width: '50%',
      },
      {
        key: 'actions',
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
