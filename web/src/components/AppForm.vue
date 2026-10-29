<template>
  <v-form
      ref="form"
      v-model="formValid"
      v-if="item != null"
      data-testid="appForm-form"
      :data-state="formSaving ? 'loading' : (formError ? 'error' : 'ready')"
  >
    <v-alert
        :model-value="!!formError"
        color="error"
        class="pb-2 mb-4"
        data-testid="appForm-error"
    >{{ formError }}
    </v-alert>

    <v-text-field
        v-model="id"
        :label="$t('ID')"
        :rules="[v => !!v || $t('id_required')]"
        required
        :disabled="formSaving"
        data-testid="appForm-id"
    ></v-text-field>

    <v-text-field
        v-model="item.icon"
        :label="$t('Icon')"
        :disabled="formSaving"
        data-testid="appForm-icon"
    ></v-text-field>

    <v-text-field
        v-model="item.title"
        :label="$t('name')"
        :disabled="formSaving"
        data-testid="appForm-title"
    ></v-text-field>

    <v-text-field
        v-model="item.path"
        :label="$t('Path')"
        :disabled="formSaving"
        data-testid="appForm-path"
    ></v-text-field>

    <v-text-field
      type="number"
      v-model.number="item.priority"
      :label="$t('Priority')"
      :disabled="formSaving"
      data-testid="appForm-priority"
    ></v-text-field>

    <ArgsPicker
      style="margin-top: -10px;"
      :vars="item.args"
      @change="setArgs"
      test-id="appForm-args"
    />

    <v-checkbox
        v-model="item.active"
        :label="$t('Active')"
        data-testid="appForm-active"
    ></v-checkbox>

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import ArgsPicker from './ArgsPicker.vue';

export default {
  components: { ArgsPicker },
  mixins: [ItemFormBase],

  computed: {
    isNew() {
      return this.itemId === '';
    },
  },

  data() {
    return {
      id: null,
    };
  },

  watch: {
    itemId() {
      this.id = this.itemId;
    },
  },

  methods: {
    setArgs(args) {
      this.item.args = args;
    },

    beforeLoadData() {
      if (!this.isNew) {
        this.id = this.itemId;
      }
    },

    afterReset() {
      this.id = null;
    },

    getItemsUrl() {
      return `/api/apps/${this.id}`;
    },

    getSingleItemUrl() {
      return `/api/apps/${this.id}`;
    },
  },
};
</script>
