import axios from 'axios';
import EventBus from '@/event-bus';
import EditDialog from '@/components/EditDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ObjectRefsDialog from '@/components/ObjectRefsDialog.vue';

import { getErrorMessage } from '@/lib/error';
import { USER_PERMISSIONS } from '@/lib/constants';
import PermissionsCheck from '@/components/PermissionsCheck';

export default {
  components: {
    YesNoDialog,
    EditDialog,
    ObjectRefsDialog,
  },

  mixins: [PermissionsCheck],

  props: {
    projectId: Number,
    projectType: String,
    userId: Number,
    userRole: String,
    user: Object,
  },

  data() {
    const allowActions = this.allowActions();

    const headers = this.getHeaders().filter((header) => allowActions || header.value !== 'actions');

    return {
      headers,
      items: null,

      itemId: null,
      editDialog: null,
      deleteItemDialog: null,

      itemRefs: null,
      itemRefsDialog: null,

      USER_PERMISSIONS,
    };
  },

  async created() {
    await this.beforeLoadItems();
    await this.loadItems();
  },

  methods: {
    allowActions() {
      return this.can(USER_PERMISSIONS.manageProjectResources);
    },

    // eslint-disable-next-line no-empty-function
    async beforeLoadItems() {
    },

    getSingleItemUrl() {
      throw new Error('Not implemented');
    },

    getHeaders() {
      throw new Error('Not implemented');
    },

    getEventName() {
      throw new Error('Not implemented');
    },

    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async onItemSave() {
      await this.loadItems();
    },

    async askDeleteItem(itemId) {
      this.itemId = itemId;

      try {
        this.itemRefs = (await axios({
          method: 'get',
          url: `${this.getSingleItemUrl()}/refs`,
          responseType: 'json',
        })).data;

        if (this.itemRefs.templates.length > 0
          || this.itemRefs.repositories.length > 0
          || this.itemRefs.inventories.length > 0
          || this.itemRefs.schedules.length > 0) {
          this.itemRefsDialog = true;
          return;
        }
      } catch (e) {
        // Do nothing
      }

      this.deleteItemDialog = true;
    },

    async deleteItem(itemId) {
      try {
        const item = this.items.find((x) => x.id === itemId);

        await axios({
          method: 'delete',
          url: this.getSingleItemUrl(),
          responseType: 'json',
        });

        EventBus.$emit(this.getEventName(), {
          action: 'delete',
          item,
        });

        await this.loadItems();
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },

    editItem(itemId) {
      this.itemId = itemId;
      this.editDialog = true;
    },

    async loadItems() {
      this.items = (await axios({
        method: 'get',
        url: this.getItemsUrl(),
        responseType: 'json',
      })).data;
    },
  },
};
