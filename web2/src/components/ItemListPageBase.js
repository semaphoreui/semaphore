import axios from 'axios';
import EventBus from '@/event-bus';
import EditDialog from '@/components/EditDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import ObjectRefsView from '@/components/ObjectRefsView.vue';

import { getErrorMessage } from '@/lib/error';

export default {
  components: {
    YesNoDialog,
    EditDialog,
    ObjectRefsView,
  },

  props: {
    projectId: Number,
    userId: Number,
  },

  data() {
    return {
      headers: this.getHeaders(),
      items: null,

      itemId: null,
      editDialog: null,
      deleteItemDialog: null,

      itemRefs: null,
      itemRefsDialog: null,
    };
  },

  async created() {
    await this.beforeLoadItems();
    await this.loadItems();
  },

  methods: {
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
