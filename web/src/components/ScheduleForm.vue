<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="templates && item != null"
  >

    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}</v-alert>

<!--    <v-text-field-->
<!--      v-model="item.name"-->
<!--      :label="$t('Name')"-->
<!--      :rules="[v => !!v || $t('name_required')]"-->
<!--      required-->
<!--      :disabled="formSaving"-->
<!--      class="mb-4"-->
<!--    ></v-text-field>-->

    <v-select
      v-model="item.template_id"
      :label="$t('Template')"
      :items="templates"
      item-value="id"
      :item-text="(itm) => itm.name"
      :rules="[v => !!v || $t('template_required')]"
      required
      :disabled="formSaving"
    />

    <v-text-field
      v-model="item.cron_format"
      :label="$t('Cron')"
      :rules="[v => !!v || $t('Cron required')]"
      required
      :disabled="formSaving"
      class="mb-4"
    ></v-text-field>

  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      templates: null,
    };
  },

  async created() {
    this.templates = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/templates`,
      responseType: 'json',
    })).data;
  },

  methods: {

    getItemsUrl() {
      return `/api/project/${this.projectId}/schedules`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/schedules/${this.itemId}`;
    },

  },
};
</script>
