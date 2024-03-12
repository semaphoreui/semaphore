import axios from 'axios';
import { getErrorMessage } from '@/lib/error';

export default {
  props: {
    integrationId: [Number, String],
  },
  methods: {
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
        await this.beforeSave();

        item = (await axios({
          method: this.isNew ? 'post' : 'put',
          url: this.isNew
            ? this.getItemsUrl()
            : this.getSingleItemUrl(),
          responseType: 'json',
          data: {
            ...this.item,
            project_id: this.$route.params.projectId,
            integration_id: this.integrationId,
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
