<template>
  <v-form
      ref="form"
      lazy-validation
      v-model="formValid"
      v-if="item != null"
  >
    <v-alert
        :value="formError"
        color="error"
        class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-text-field
        v-model="item.name"
        :label="$t('name')"
        :disabled="formSaving"
    ></v-text-field>

    <v-text-field
        v-model="item.path"
        :label="$t('Path')"
        :disabled="formSaving"
    ></v-text-field>

    <v-text-field
        type="number"
        v-model.number="item.priority"
        :label="$t('Priority')"
        :disabled="formSaving"
    ></v-text-field>

    <ArgsPicker style="margin-top: -10px;" :vars="item.parsedArgs" @change="setArgs"/>

    <v-checkbox
        v-model="item.active"
        :label="$t('Active')"
    ></v-checkbox>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import ArgsPicker from './ArgsPicker.vue';

export default {
  components: { ArgsPicker },
  mixins: [ItemFormBase],

  props: {
    appId: String,
  },

  methods: {
    setArgs(args) {
      this.item.parsedArgs = args;
    },

    getNewItem() {
      return {
        name: '',
        path: '',
        args: null,
        parsedArgs: [],
        active: true,
        priority: 0,
      };
    },

    afterLoadData() {
      if (this.item) {
        try {
          this.item.parsedArgs = this.item.args ? JSON.parse(this.item.args) : [];
        } catch (e) {
          this.item.parsedArgs = [];
        }
      }
    },

    beforeSave() {
      if (this.item.parsedArgs && this.item.parsedArgs.length > 0) {
        this.item.args = JSON.stringify(this.item.parsedArgs);
      } else {
        this.item.args = null;
      }
    },

    getItemsUrl() {
      return `/api/apps/${this.appId}/versions`;
    },

    getSingleItemUrl() {
      return `/api/apps/${this.appId}/versions/${this.itemId}`;
    },
  },
};
</script>
