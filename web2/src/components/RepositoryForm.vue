<template>
  <v-form
      ref="form"
      lazy-validation
      v-model="formValid"
      v-if="item != null && keys != null"
  >
    <v-alert
        :value="formError"
        color="error"
        class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-text-field
        v-model="item.name"
        label="Name"
        :rules="[v => !!v || 'Name is required']"
        required
        :disabled="formSaving"
    ></v-text-field>

    <v-text-field
        v-model="item.git_url"
        label="Git URL"
        append-outer-icon="mdi-help-circle"
        :rules="[v => !!v || 'Repository is required']"
        required
        :disabled="formSaving"
        @click:append-outer="showGitUrlHelp()"
    ></v-text-field>

    <v-select
        v-model="item.ssh_key_id"
        label="Access Key"
        append-outer-icon="mdi-help-circle"
        :items="keys"
        item-value="id"
        item-text="name"
        :rules="[v => !!v || 'Key is required']"
        required
        :disabled="formSaving"
        @click:append-outer="showKeyHelp()"
    ></v-select>

    <v-dialog
        v-model="gitUrlHelpDialog"
        hide-overlay
        width="300"
    >
      <v-alert
          border="top"
          colored-border
          type="info"
          elevation="2"
          class="mb-0"
      >
        <p><b>Git URL</b> can be SSH (git@***) or HTTPS (https://***) URL.</p>
        <p>If you use SSH URL you should specify <b>Access Key</b> with type <code>SSH</code>.</p>
        <p>If you use HTTPS URL you should specify <b>Access Key</b> with type
          <code>None</code>.</p>
      </v-alert>
    </v-dialog>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      gitUrlHelpDialog: false,
      keyHelpDialog: false,

      keys: null,
      inventoryTypes: [{
        id: 'static',
        name: 'Static',
      }, {
        id: 'file',
        name: 'File',
      }],
    };
  },
  async created() {
    this.keys = (await axios({
      keys: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
  },
  methods: {
    showGitUrlHelp() {
      this.gitUrlHelpDialog = true;
    },

    showKeyHelp() {
      this.gitUrlHelpDialog = true;
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/repositories`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/repositories/${this.itemId}`;
    },
  },
};
</script>
