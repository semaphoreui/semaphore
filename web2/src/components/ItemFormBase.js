import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

/**
 * Most of Semaphore entities (keys, environments, etc) has similar REST API for access and
 * manipulation. This class presents this entity. Something like CRUD. It should be used as mixin
 * in vue.js template. Example: KeyForm.vue.
 */
export default {
  props: {
    itemId: [Number, String],
    projectId: [Number, String],
    needSave: Boolean, // flag which signal about user want to save form
    needReset: Boolean, // flag which signal about user want to reset form
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

    getRequestOptions() {
      return {};
    },

    /**
     * Saves or creates item via API.
     * @returns {Promise<null>} null if validation didn't pass or user data if user saved.
     */
    async save() {
      this.formError = null;

      if (!this.$refs.form.validate()) {
        this.$emit('error', {});
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
          data: {
            ...this.item,
            project_id: this.projectId,
          },
          ...(this.getRequestOptions()),
        })).data;

        this.$emit('save', {
          item: item || this.item,
          action: this.isNew ? 'new' : 'edit',
        });
      } catch (err) {
        this.formError = getErrorMessage(err);
        this.$emit('error', {});
      } finally {
        this.formSaving = false;
      }

      return item || this.item;
    },
  },
};
