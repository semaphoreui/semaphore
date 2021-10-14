<template>
  <v-form
      ref="form"
      lazy-validation
      v-model="formValid"
      v-if="isLoaded()"
  >
    <v-alert
        :value="formError"
        color="error"
        class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-alert
       v-if="commitHash"
    >{{ commitHash.substr(0, 6) }}
    </v-alert>

    <v-text-field
        v-model="item.message"
        label="Message (Optional)"
        :disabled="formSaving"
    />

    <v-select
        v-if="template.type === 'deploy'"
        v-model="item.version"
        label="Build Version"
        :items="buildTasks"
        item-value="version"
        item-text="version"
        :rules="[v => !!v || 'Build Version is required']"
        required
        :disabled="formSaving"
    />

    <v-row no-gutters>
      <v-col>
        <v-checkbox
            v-model="item.debug"
            label="Debug"
        ></v-checkbox>
      </v-col>
      <v-col>
        <v-checkbox
            v-model="item.dry_run"
            label="Dry Run"
        ></v-checkbox>
      </v-col>
    </v-row>

  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],
  props: {
    templateId: Number,
    commitHash: String,
    version: String,
  },
  data() {
    return {
      template: null,
      buildTasks: null,
    };
  },
  watch: {
    needReset(val) {
      if (val) {
        this.item.template_id = this.templateId;
      }
    },

    templateId(val) {
      this.item.template_id = val;
    },

    version(val) {
      this.item.version = val;
    },
  },

  methods: {
    isLoaded() {
      return this.item != null
          && this.template != null
          && this.buildTasks != null;
    },

    async afterLoadData() {
      this.item.template_id = this.templateId;

      this.template = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates/${this.templateId}`,
        responseType: 'json',
      })).data;

      this.buildTasks = (await axios({
        keys: 'get',
        url: `/api/project/${this.projectId}/templates/${this.templateId}/tasks`,
        responseType: 'json',
      })).data.filter((task) => task.version != null);

      if (this.buildTasks.length > 0) {
        this.item.version = this.buildTasks[this.buildTasks.length - 1].version;
      }
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/tasks`;
    },
  },
};
</script>
