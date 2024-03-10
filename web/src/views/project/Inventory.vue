<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :icon="itemApp.icon"
      :icon-color="itemApp.color"
      :title="`${itemId === 'new' ? $t('nnew') : $t('edit')} ${itemApp.title}`"
      :max-width="450"
      @save="loadItems"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <TerraformInventoryForm
          v-if="itemApp.slug === 'terraform'"
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
        <InventoryForm
          v-else
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

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('inventory') }}</v-toolbar-title>
      <v-spacer></v-spacer>

      <v-menu
        open-on-hover
        offset-y
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            v-bind="attrs"
            v-on="on"
            color="primary"
            v-if="can(USER_PERMISSIONS.manageProjectResources)"
          >{{ $t('newInventory') }}</v-btn>
        </template>
        <v-list>
          <v-list-item
            v-for="item in templateApps"
            :key="item.slug"
            link
            @click="itemApp = item; editItem('new');"
          >
            <v-list-item-icon>
              <v-icon :color="item.color">{{ item.icon }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>

    </v-toolbar>

    <v-data-table
      :headers="headers"
      :items="items"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
    >
      <template v-slot:item.type="{ item }">
        <code>{{ item.type }}</code>
      </template>
      <template v-slot:item.inventory="{ item }">
        {{ item.type === 'file' ? item.inventory : '&mdash;' }}
      </template>
      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
            icon
            class="mr-1"
            @click="askDeleteItem(item.id)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>

          <v-btn
            icon
            class="mr-1"
            @click="editItem(item.id)"
          >
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import InventoryForm from '@/components/InventoryForm.vue';
import TerraformInventoryForm from '@/components/TerraformInventoryForm.vue';

export default {
  mixins: [ItemListPageBase],
  components: { TerraformInventoryForm, InventoryForm },

  data() {
    return {
      templateApps: [{
        slug: '',
        title: 'Ansible Inventory',
        icon: 'mdi-ansible',
        color: 'black',
      }, {
        slug: 'terraform',
        title: 'Terraform Workspace',
        icon: 'mdi-terraform',
        color: '#7b42bc',
      }],
      itemApp: {},
    };
  },

  methods: {
    getHeaders() {
      return [{
        text: this.$i18n.t('name'),
        value: 'name',
        width: '33.33%',
      },
      {
        text: this.$i18n.t('type'),
        value: 'type',
        width: '33.33%',
      },
      {
        text: this.$i18n.t('path'),
        value: 'inventory',
        width: '33.33%',
      },
      {
        text: this.$i18n.t('actions'),
        value: 'actions',
        sortable: false,
      }];
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
