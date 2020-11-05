import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    itemId: [Number, String],
    projectId: [Number, String],
    needSave: Boolean,
    needReset: Boolean,
  },

  data() {
    return {
      item: null,
      formValid: false,
      formError: null,
      formSaving: false,
    };
  },

  async created() {
    await this.loadData();
  },

  computed: {
    isNew() {
      return this.itemId === 'new';
    },
  },

  watch: {
    async needSave(val) {
      if (val) {
        await this.save();
      }
    },
    async needReset(val) {
      if (val) {
        await this.reset();
      }
    },
  },

  methods: {
    async reset() {
      this.item = null;
      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }
      await this.loadData();
    },

    getItemsUrl() {
      throw new Error('Not implemented');
    },

    getSingleItemUrl() {
      throw new Error('Not implemented');
    },

    async loadData() {
      if (this.isNew) {
        this.item = {};
      } else {
        this.item = (await axios({
          method: 'get',
          url: this.getSingleItemUrl(),
          responseType: 'json',
        })).data;
      }
    },

    /**
     * Saves or creates item via API.
     * @returns {Promise<null>} null if validation didn't pass or user data if user saved.
     */
    async save() {
      this.formError = null;

      if (!this.$refs.form.validate()) {
        return null;
      }

      this.formSaving = true;
      let item;

      try {
        item = (await axios({
          method: this.isNew ? 'post' : 'put',
          url: this.isNew
            ? this.getItemsUrl()
            : this.getSingleItemUrl(),
          responseType: 'json',
          data: this.item,
        })).data;

        this.$emit('save', {
          item: item || this.item,
          action: this.isNew ? 'new' : 'edit',
        });
      } catch (err) {
        this.formError = getErrorMessage(err);
      } finally {
        this.formSaving = false;
      }

      return item || this.item;
    },
  },
};
