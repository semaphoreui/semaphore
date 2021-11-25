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
        label="Repository URL"
        :rules="[v => !!v || 'Repository is required']"
        required
        :disabled="formSaving"
        append-outer-icon="mdi-help-circle"
        @click:append-outer="showHelpDialog('url')"
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
        @click:append-outer="showHelpDialog('key')"
    ></v-select>

    <v-dialog
        v-model="helpDialog"
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
        <div v-if="helpKey === 'url'">
          <p>
            Address of the repository with your Ansible playbooks. It can be:
          </p>
          <ul>
            <li>Git URL <code>git://</code></li>
            <li>SSH URL <code>ssh://</code></li>
            <li>HTTPS URL <code>https://</code></li>
            <li>file URL <code>file://</code></li>
          </ul>
        </div>
        <div v-else-if="helpKey === 'key'">
          <p>Credentials to access to the Git repository. It should be:</p>
          <ul>
            <li><code>SSH</code> if you use Git or SSH URL.</li>
            <li><code>None</code> if you use HTTPS or file URL.</li>
          </ul>
        </div>
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
      helpDialog: null,
      helpKey: null,

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
    showHelpDialog(key) {
      this.helpKey = key;
      this.helpDialog = true;
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
