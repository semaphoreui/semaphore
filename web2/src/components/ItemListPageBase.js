import axios from 'axios';
import EventBus from '@/event-bus';
import EditDialog from '@/components/EditDialog.vue';
import YesNoDialog from '@/components/YesNoDialog.vue';
import { getErrorMessage } from '@/lib/error';

export default {
  components: {
    YesNoDialog,
    EditDialog,
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
    };
  },

  async created() {
    await this.beforeLoadItems();
    await this.loadItems();
  },

  methods: {
    // eslint-disable-next-line no-empty-function
    async beforeLoadItems() { },

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

    askDeleteItem(itemId) {
      this.itemId = itemId;
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
