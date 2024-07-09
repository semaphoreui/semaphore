import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

/**
 * Most of Semaphore entities (keys, environments, etc) have similar REST API for
 * access and manipulation.
 * This class presents mixin for creating editing form for such entities.
 * This class should be used as mixin in Vue-template.
 *
 * Simplest example: KeyForm.vue. It demonstrate all you need to understand how it works.
 *
 * You must provide next required properties to use this mixin:
 *
 * * itemId
 * * projectId
 *
 * Your template must have <v-form ref="form">...</v-form>.
 *
 * You must provide next methods in your template:
 *
 * * getItemsUrl() - returns URL for retrieving collection of entities (GET-method).
 * * getSingleItemUrl() - returns URL for retrieving and manipulation of single entity
 *                        (GET, POST, PUT, DELETE methods).
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
      this.formError = null;
      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }
      await this.afterReset();
      await this.loadData();
    },

    getItemsUrl() {
      throw new Error('Not implemented'); // must me implemented in template
    },

    getSingleItemUrl() {
      throw new Error('Not implemented'); // must me implemented in template
    },

    beforeSave() {

    },

    afterReset() {

    },

    afterSave() {

    },

    beforeLoadData() {

    },

    afterLoadData() {

    },

    getNewItem() {
      return {};
    },

    async loadData() {
      await this.beforeLoadData();

      if (this.isNew) {
        this.item = this.getNewItem();
      } else {
        this.item = (await axios({
          method: 'get',
          url: this.getSingleItemUrl(),
          responseType: 'json',
        })).data;
      }

      await this.afterLoadData();
    },

    /**
     * You add add/override some PUT/POST request options with using this method.
     * For example, you want to change response type, just override this method:
     * ```
     * getRequestOptions() {
     *   return {
     *     responseType: 'text'
     *   }
     * }
     * ```
     *
     * This method works only for create (POST) and update (PUT) requests.
     * @returns {Object}
     */
    getRequestOptions() {
      return {};
    },

    /**
     * Saves or creates item via API.
     * @returns {Promise<null>} null if validation didn't pass or user data if user saved.
     */
    async save(data = {}) {
      this.formError = null;

      if (!this.$refs.form.validate()) {
        this.$emit('error', {});
        return null;
      }

      this.formSaving = true;
      let item;

      try {
        await this.beforeSave();

        item = (await axios({
          method: this.isNew ? 'post' : 'put',
          url: this.isNew
            ? this.getItemsUrl()
            : this.getSingleItemUrl(),
          responseType: 'json',
          data: {
            ...this.item,
            project_id: this.projectId,
            ...data,
          },
          ...(this.getRequestOptions()),
        })).data;

        await this.afterSave(item);

        this.$emit('save', {
          item: item || this.item,
          action: this.isNew ? 'new' : 'edit',
        });
      } catch (err) {
        this.formError = getErrorMessage(err);
        this.$emit('error', {
          message: this.formError,
        });
      } finally {
        this.formSaving = false;
      }

      return item || this.item;
    },
  },
};
